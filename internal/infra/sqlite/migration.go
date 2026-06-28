package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"

	"teer/internal/infra/config"
)

// MigrateIfNeeded memeriksa apakah ada data lama (JSON config + scrollback files)
// dan memindahkannya ke SQLite bila tabel workspaces masih kosong.
// Operasi ini best-effort: error tidak menghentikan startup.
func MigrateIfNeeded(db *sql.DB, oldStore *config.Store, oldScrollback *config.ScrollbackStore) error {
	// Cek apakah tabel workspaces sudah berisi data.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM workspaces`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // sudah ada data, tidak perlu migrasi
	}

	// Muat config lama.
	cfg, err := oldStore.Load()
	if err != nil || len(cfg.Workspaces) == 0 {
		return err
	}

	// Simpan ke SQLite via ConfigStore.
	cs := NewConfigStore(db)
	if err := cs.Save(cfg); err != nil {
		return err
	}

	// Migrasi scrollback files.
	if oldScrollback != nil {
		for _, ws := range cfg.Workspaces {
			for _, sd := range ws.Sessions {
				data, err := oldScrollback.Load(sd.ID)
				if err != nil || data == "" {
					continue
				}
				ss := NewScrollbackStore(db)
				_ = ss.Save(sd.ID, data) // best-effort
			}
		}
	}

	// Arsipkan config lama agar tidak dipakai lagi.
	if p, err := xdg.ConfigFile("teer/config.json"); err == nil {
		_ = os.Rename(p, p+".migrated")
	}

	// Arsipkan direktori sessions lama.
	if p, err := xdg.ConfigFile("teer/sessions/.keep"); err == nil {
		dir := filepath.Dir(p)
		if _, err := os.Stat(dir); err == nil {
			_ = archiveScrollbackDir(dir)
		}
	}

	return nil
}

// archiveScrollbackDir mengganti nama direktori sessions menjadi sessions.migrated.
// File individual tidak dihapus sehingga user bisa memulihkan bila perlu.
func archiveScrollbackDir(dir string) error {
	dest := strings.TrimSuffix(dir, "/") + ".migrated"
	if _, err := os.Stat(dest); err == nil {
		return nil // sudah ada, skip
	}
	return os.Rename(dir, dest)
}
