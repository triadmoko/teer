package config

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adrg/xdg"
)

// defaultMaxSnapshotSize adalah batas atas byte teks scrollback (sebelum gzip).
// Frontend sudah membatasi buffer ke scrollbackLines saat serialize, jadi nilai
// ini hanya berfungsi sebagai guard agar file snapshot tidak meledak bila ada
// output abnormal.
const defaultMaxSnapshotSize = 8 << 20 // 8 MiB

const sessionsRelDir = "teer/sessions"

const scrollbackExt = ".scrollback"

// ScrollbackStore menyimpan snapshot scrollback per-session sebagai file gzip di
// ~/.config/teer/sessions/<id>.scrollback. Snapshot bersifat best-effort:
// kegagalan I/O tidak boleh menggagalkan operasi utama.
type ScrollbackStore struct {
	dir     string // ~/.config/teer/sessions
	maxSize int    // batas byte teks per snapshot sebelum gzip
	mu      sync.Mutex
}

// NewScrollbackStore membangun store dan memastikan direktori sessions ada.
func NewScrollbackStore() (*ScrollbackStore, error) {
	// xdg.ConfigFile memastikan direktori induk dibuat; argumen file dummy hanya
	// dipakai untuk mengambil path direktorinya.
	p, err := xdg.ConfigFile(sessionsRelDir + "/.keep")
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &ScrollbackStore{dir: dir, maxSize: defaultMaxSnapshotSize}, nil
}

// safeID menolak id yang bisa dipakai untuk path-traversal — id menjadi nama
// file, jadi tidak boleh mengandung separator atau "..".
func safeID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	if strings.ContainsAny(id, `/\`) {
		return false
	}
	return !strings.Contains(id, "..")
}

func (s *ScrollbackStore) pathFor(id string) string {
	return filepath.Join(s.dir, id+scrollbackExt)
}

// Save menyimpan snapshot scrollback secara atomic (tmp+rename), gzip, 0600.
// Bila melebihi maxSize, teks dipotong dari awal (buang baris terlama) tanpa
// menghasilkan error. id tidak aman diabaikan diam-diam.
func (s *ScrollbackStore) Save(id, data string) error {
	if !safeID(id) {
		return nil
	}

	if len(data) > s.maxSize {
		data = data[len(data)-s.maxSize:]
		// Buang potongan baris pertama yang tidak utuh agar replay rapi.
		if i := strings.IndexByte(data, '\n'); i >= 0 && i+1 <= len(data) {
			data = data[i+1:]
		}
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(data)); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}

	path := s.pathFor(id)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load mengembalikan teks scrollback. Bila snapshot tidak ada, mengembalikan
// ("", nil).
func (s *ScrollbackStore) Load(id string) (string, error) {
	if !safeID(id) {
		return "", nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.pathFor(id))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	data, err := io.ReadAll(gz)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Delete menghapus snapshot. Tidak error bila file memang tidak ada.
func (s *ScrollbackStore) Delete(id string) error {
	if !safeID(id) {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.pathFor(id))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Prune menghapus snapshot yatim — file yang id-nya tidak ada di validIDs.
func (s *ScrollbackStore) Prune(validIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	valid := make(map[string]struct{}, len(validIDs))
	for _, id := range validIDs {
		valid[id+scrollbackExt] = struct{}{}
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, scrollbackExt) {
			continue
		}
		if _, ok := valid[name]; ok {
			continue
		}
		_ = os.Remove(filepath.Join(s.dir, name))
	}
	return nil
}
