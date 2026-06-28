package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"teer/internal/domain"
)

// ConfigStore mengimplementasikan domain.Repository dengan SQLite backend.
type ConfigStore struct {
	db *sql.DB
}

var _ domain.Repository = (*ConfigStore)(nil)

func NewConfigStore(db *sql.DB) *ConfigStore {
	return &ConfigStore{db: db}
}

func (s *ConfigStore) Load() (*domain.Config, error) {
	rows, err := s.db.Query(
		`SELECT id, name, color, default_cwd, startup_command, env, sort_order, created_at, updated_at
		 FROM workspaces ORDER BY sort_order`,
	)
	if err != nil {
		return nil, err
	}

	// Kumpulkan semua workspace dulu sebelum menutup rows, agar koneksi bebas
	// untuk query sessions (SQLite MaxOpenConns=1 tidak boleh nested query).
	var workspaces []*domain.Workspace
	for rows.Next() {
		ws, err := scanWorkspace(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for _, ws := range workspaces {
		sessions, err := s.loadSessions(ws.ID)
		if err != nil {
			return nil, err
		}
		ws.Sessions = sessions
	}

	if workspaces == nil {
		workspaces = []*domain.Workspace{}
	}
	return &domain.Config{Workspaces: workspaces}, nil
}

func (s *ConfigStore) loadSessions(workspaceID string) ([]*domain.SessionDef, error) {
	rows, err := s.db.Query(
		`SELECT id, workspace_id, name, shell, cwd, startup_command, env, auto_start, layout, sort_order
		 FROM session_defs WHERE workspace_id = ? ORDER BY sort_order`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.SessionDef
	for rows.Next() {
		sd, err := scanSessionDef(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, sd)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if sessions == nil {
		sessions = []*domain.SessionDef{}
	}
	return sessions, nil
}

func (s *ConfigStore) Save(cfg *domain.Config) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Kumpulkan ID yang masih ada untuk menghapus yang sudah tidak ada.
	keepWorkspaceIDs := make([]string, 0, len(cfg.Workspaces))
	for _, ws := range cfg.Workspaces {
		keepWorkspaceIDs = append(keepWorkspaceIDs, ws.ID)
	}

	if err := deleteAbsent(tx, "workspaces", keepWorkspaceIDs); err != nil {
		return err
	}

	now := time.Now().UTC()

	for i, ws := range cfg.Workspaces {
		if ws.CreatedAt.IsZero() {
			ws.CreatedAt = now
		}
		ws.UpdatedAt = now

		envJSON, err := marshalJSON(ws.Env)
		if err != nil {
			return fmt.Errorf("workspace %s env: %w", ws.ID, err)
		}

		_, err = tx.Exec(
			`INSERT INTO workspaces (id, name, color, default_cwd, startup_command, env, sort_order, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
			   name=excluded.name, color=excluded.color, default_cwd=excluded.default_cwd,
			   startup_command=excluded.startup_command, env=excluded.env,
			   sort_order=excluded.sort_order, updated_at=excluded.updated_at`,
			ws.ID, ws.Name, ws.Color, ws.DefaultCwd, ws.StartupCommand,
			envJSON, i, ws.CreatedAt.Format(time.RFC3339Nano), ws.UpdatedAt.Format(time.RFC3339Nano),
		)
		if err != nil {
			return fmt.Errorf("upsert workspace %s: %w", ws.ID, err)
		}

		keepSessionIDs := make([]string, 0, len(ws.Sessions))
		for _, sd := range ws.Sessions {
			keepSessionIDs = append(keepSessionIDs, sd.ID)
		}
		if err := deleteAbsentInWorkspace(tx, ws.ID, keepSessionIDs); err != nil {
			return err
		}

		for j, sd := range ws.Sessions {
			sd.WorkspaceID = ws.ID
			envJSON, err := marshalJSON(sd.Env)
			if err != nil {
				return fmt.Errorf("session %s env: %w", sd.ID, err)
			}
			layoutJSON, err := marshalNullableJSON(sd.Layout)
			if err != nil {
				return fmt.Errorf("session %s layout: %w", sd.ID, err)
			}

			autoStart := 0
			if sd.AutoStart {
				autoStart = 1
			}

			_, err = tx.Exec(
				`INSERT INTO session_defs
				   (id, workspace_id, name, shell, cwd, startup_command, env, auto_start, layout, sort_order)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				 ON CONFLICT(id) DO UPDATE SET
				   workspace_id=excluded.workspace_id, name=excluded.name, shell=excluded.shell,
				   cwd=excluded.cwd, startup_command=excluded.startup_command, env=excluded.env,
				   auto_start=excluded.auto_start, layout=excluded.layout, sort_order=excluded.sort_order`,
				sd.ID, ws.ID, sd.Name, sd.Shell, sd.Cwd, sd.StartupCommand,
				envJSON, autoStart, layoutJSON, j,
			)
			if err != nil {
				return fmt.Errorf("upsert session %s: %w", sd.ID, err)
			}
		}
	}

	return tx.Commit()
}

// deleteAbsent menghapus baris dari tabel yang ID-nya tidak ada di keepIDs.
func deleteAbsent(tx *sql.Tx, table string, keepIDs []string) error {
	if len(keepIDs) == 0 {
		_, err := tx.Exec("DELETE FROM " + table)
		return err
	}
	placeholders := strings.Repeat("?,", len(keepIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(keepIDs))
	for i, id := range keepIDs {
		args[i] = id
	}
	_, err := tx.Exec("DELETE FROM "+table+" WHERE id NOT IN ("+placeholders+")", args...)
	return err
}

func deleteAbsentInWorkspace(tx *sql.Tx, workspaceID string, keepIDs []string) error {
	if len(keepIDs) == 0 {
		_, err := tx.Exec("DELETE FROM session_defs WHERE workspace_id = ?", workspaceID)
		return err
	}
	placeholders := strings.Repeat("?,", len(keepIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, 0, len(keepIDs)+1)
	args = append(args, workspaceID)
	for _, id := range keepIDs {
		args = append(args, id)
	}
	_, err := tx.Exec(
		"DELETE FROM session_defs WHERE workspace_id = ? AND id NOT IN ("+placeholders+")",
		args...,
	)
	return err
}

func scanWorkspace(rows *sql.Rows) (*domain.Workspace, error) {
	var ws domain.Workspace
	var envJSON string
	var sortOrder int
	var createdAt, updatedAt string

	err := rows.Scan(
		&ws.ID, &ws.Name, &ws.Color, &ws.DefaultCwd, &ws.StartupCommand,
		&envJSON, &sortOrder, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(envJSON), &ws.Env); err != nil {
		ws.Env = map[string]string{}
	}

	ws.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	ws.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)

	return &ws, nil
}

func scanSessionDef(rows *sql.Rows) (*domain.SessionDef, error) {
	var sd domain.SessionDef
	var envJSON string
	var layoutJSON sql.NullString
	var autoStart int
	var sortOrder int

	err := rows.Scan(
		&sd.ID, &sd.WorkspaceID, &sd.Name, &sd.Shell, &sd.Cwd, &sd.StartupCommand,
		&envJSON, &autoStart, &layoutJSON, &sortOrder,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(envJSON), &sd.Env); err != nil {
		sd.Env = map[string]string{}
	}

	sd.AutoStart = autoStart != 0

	if layoutJSON.Valid && layoutJSON.String != "" && layoutJSON.String != "null" {
		_ = json.Unmarshal([]byte(layoutJSON.String), &sd.Layout)
	}

	return &sd, nil
}

func marshalJSON(v any) (string, error) {
	if v == nil {
		return "{}", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	// json.Marshal pada nil map menghasilkan "null"; ganti dengan "{}" agar
	// unmarshal menghasilkan empty map bukan nil.
	if string(b) == "null" {
		return "{}", nil
	}
	return string(b), nil
}

func marshalNullableJSON(v any) (sql.NullString, error) {
	if v == nil {
		return sql.NullString{}, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}
