package sqlite

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"io"
	"strings"
	"time"

	"teer/internal/domain"
)

// defaultMaxScrollbackSize adalah batas atas byte teks sebelum gzip.
const defaultMaxScrollbackSize = 8 << 20 // 8 MiB

// ScrollbackStore mengimplementasikan domain.Scrollback dengan SQLite backend.
type ScrollbackStore struct {
	db      *sql.DB
	maxSize int
}

var _ domain.Scrollback = (*ScrollbackStore)(nil)

func NewScrollbackStore(db *sql.DB) *ScrollbackStore {
	return &ScrollbackStore{db: db, maxSize: defaultMaxScrollbackSize}
}

// safeID menolak id yang berpotensi berbahaya.
func safeID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	if strings.ContainsAny(id, `/\`) {
		return false
	}
	return !strings.Contains(id, "..")
}

func (s *ScrollbackStore) Save(id, data string) error {
	if !safeID(id) {
		return nil
	}

	if len(data) > s.maxSize {
		data = data[len(data)-s.maxSize:]
		if i := strings.IndexByte(data, '\n'); i >= 0 && i+1 <= len(data) {
			data = data[i+1:]
		}
	}

	compressed, err := gzipCompress(data)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO scrollbacks (session_id, data, updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT(session_id) DO UPDATE SET data=excluded.data, updated_at=excluded.updated_at`,
		id, compressed, time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *ScrollbackStore) Load(id string) (string, error) {
	if !safeID(id) {
		return "", nil
	}

	var data []byte
	err := s.db.QueryRow(
		`SELECT data FROM scrollbacks WHERE session_id = ?`, id,
	).Scan(&data)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return gzipDecompress(data)
}

func (s *ScrollbackStore) Delete(id string) error {
	if !safeID(id) {
		return nil
	}
	_, err := s.db.Exec(`DELETE FROM scrollbacks WHERE session_id = ?`, id)
	return err
}

func (s *ScrollbackStore) Prune(validIDs []string) error {
	if len(validIDs) == 0 {
		_, err := s.db.Exec(`DELETE FROM scrollbacks`)
		return err
	}

	placeholders := strings.Repeat("?,", len(validIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(validIDs))
	for i, id := range validIDs {
		args[i] = id
	}
	_, err := s.db.Exec(
		"DELETE FROM scrollbacks WHERE session_id NOT IN ("+placeholders+")",
		args...,
	)
	return err
}

func gzipCompress(data string) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(data)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gzipDecompress(data []byte) (string, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer gz.Close()
	out, err := io.ReadAll(gz)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
