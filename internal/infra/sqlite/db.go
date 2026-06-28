package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	_ "modernc.org/sqlite"
)

const dbRelPath = "teer/teer.db"

const schema = `
CREATE TABLE IF NOT EXISTS workspaces (
    id              TEXT    PRIMARY KEY,
    name            TEXT    NOT NULL,
    color           TEXT    NOT NULL DEFAULT '',
    default_cwd     TEXT    NOT NULL DEFAULT '',
    startup_command TEXT    NOT NULL DEFAULT '',
    env             TEXT    NOT NULL DEFAULT '{}',
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS session_defs (
    id              TEXT    PRIMARY KEY,
    workspace_id    TEXT    NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name            TEXT    NOT NULL,
    shell           TEXT    NOT NULL DEFAULT '',
    cwd             TEXT    NOT NULL DEFAULT '',
    startup_command TEXT    NOT NULL DEFAULT '',
    env             TEXT    NOT NULL DEFAULT '{}',
    auto_start      INTEGER NOT NULL DEFAULT 0,
    layout          TEXT,
    sort_order      INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS scrollbacks (
    session_id  TEXT    PRIMARY KEY,
    data        BLOB    NOT NULL,
    updated_at  DATETIME NOT NULL
);
`

// Open membuka (atau membuat) database SQLite di XDG config path dan
// menjalankan schema creation. DB yang dikembalikan di-share ke semua store.
func Open() (*sql.DB, error) {
	p, err := xdg.ConfigFile(dbRelPath)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", p)
	if err != nil {
		return nil, err
	}

	// SQLite tidak mendukung concurrent writers — satu koneksi sudah cukup.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
