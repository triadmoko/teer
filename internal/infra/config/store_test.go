package config

import (
	"testing"

	"github.com/adrg/xdg"

	"teer/internal/domain"
)

func TestStorePersistence(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	xdg.Reload()

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load awal: %v", err)
	}
	if len(cfg.Workspaces) != 0 {
		t.Fatalf("harusnya kosong, dapat %d", len(cfg.Workspaces))
	}

	cfg.Workspaces = append(cfg.Workspaces, &domain.Workspace{ID: "w1", Name: "Proyek A"})
	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	store2, _ := NewStore()
	got, err := store2.Load()
	if err != nil {
		t.Fatalf("Load#2: %v", err)
	}
	if len(got.Workspaces) != 1 || got.Workspaces[0].Name != "Proyek A" {
		t.Fatalf("persistensi gagal: %+v", got)
	}
}
