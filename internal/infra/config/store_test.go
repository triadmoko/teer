package config

import (
	"testing"

	"github.com/adrg/xdg"

	"teer/internal/domain"
)

// TestStorePersistence memverifikasi Save lalu Load lewat file JSON nyata di
// direktori temp. xdg menghitung path-nya saat init, jadi setelah mengubah
// XDG_CONFIG_HOME kita panggil xdg.Reload() agar path mengarah ke temp dir.
func TestStorePersistence(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	xdg.Reload()

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// Awalnya file belum ada → config kosong.
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load awal: %v", err)
	}
	if len(cfg.Workspaces) != 0 {
		t.Fatalf("harusnya kosong, dapat %d", len(cfg.Workspaces))
	}

	// Simpan satu workspace.
	cfg.Workspaces = append(cfg.Workspaces, &domain.Workspace{ID: "w1", Name: "Proyek A"})
	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Store baru harus membaca data yang sama dari disk.
	store2, _ := NewStore()
	got, err := store2.Load()
	if err != nil {
		t.Fatalf("Load#2: %v", err)
	}
	if len(got.Workspaces) != 1 || got.Workspaces[0].Name != "Proyek A" {
		t.Fatalf("persistensi gagal: %+v", got)
	}
}
