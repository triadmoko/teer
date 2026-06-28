package sqlite

import (
	"os"
	"testing"
	"time"

	"github.com/adrg/xdg"

	"teer/internal/domain"
	"teer/internal/infra/config"
)

func setupLegacyStores(t *testing.T) (*config.Store, *config.ScrollbackStore) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	xdg.Reload()

	oldStore, err := config.NewStore()
	if err != nil {
		t.Fatalf("NewStore lama: %v", err)
	}
	oldScrollback, err := config.NewScrollbackStore()
	if err != nil {
		t.Fatalf("NewScrollbackStore lama: %v", err)
	}
	return oldStore, oldScrollback
}

func TestMigrateIfNeeded(t *testing.T) {
	oldStore, oldScrollback := setupLegacyStores(t)

	now := time.Now().UTC().Truncate(time.Second)
	cfg := &domain.Config{
		Workspaces: []*domain.Workspace{
			{
				ID:        "ws-migrate",
				Name:      "MigrasiWS",
				CreatedAt: now,
				UpdatedAt: now,
				Env:       map[string]string{"KEY": "val"},
				Sessions: []*domain.SessionDef{
					{ID: "s-migrate", WorkspaceID: "ws-migrate", Name: "SesiMigrasi", AutoStart: true},
				},
			},
		},
	}
	if err := oldStore.Save(cfg); err != nil {
		t.Fatalf("Save lama: %v", err)
	}
	if err := oldScrollback.Save("s-migrate", "output terminal"); err != nil {
		t.Fatalf("Save scrollback lama: %v", err)
	}

	// Jalankan migrasi ke DB baru.
	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	defer db.Close()

	if err := MigrateIfNeeded(db, oldStore, oldScrollback); err != nil {
		t.Fatalf("MigrateIfNeeded: %v", err)
	}

	// Verifikasi workspace & session tersalin.
	cs := NewConfigStore(db)
	got, err := cs.Load()
	if err != nil {
		t.Fatalf("Load setelah migrasi: %v", err)
	}
	if len(got.Workspaces) != 1 {
		t.Fatalf("ekspektasi 1 workspace, dapat %d", len(got.Workspaces))
	}
	if got.Workspaces[0].ID != "ws-migrate" {
		t.Fatalf("workspace ID salah: %s", got.Workspaces[0].ID)
	}
	if len(got.Workspaces[0].Sessions) != 1 {
		t.Fatalf("ekspektasi 1 session, dapat %d", len(got.Workspaces[0].Sessions))
	}

	// Verifikasi scrollback tersalin.
	ss := NewScrollbackStore(db)
	scrollData, err := ss.Load("s-migrate")
	if err != nil {
		t.Fatalf("Load scrollback setelah migrasi: %v", err)
	}
	if scrollData != "output terminal" {
		t.Fatalf("scrollback salah: %q", scrollData)
	}
}

func TestMigrateIfNeededSkipsWhenNotEmpty(t *testing.T) {
	oldStore, oldScrollback := setupLegacyStores(t)

	now := time.Now().UTC()
	cfg := &domain.Config{
		Workspaces: []*domain.Workspace{
			{ID: "ws-lama", Name: "Lama", CreatedAt: now, UpdatedAt: now, Sessions: []*domain.SessionDef{}},
		},
	}
	_ = oldStore.Save(cfg)

	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	defer db.Close()

	// Isi DB SQLite terlebih dahulu.
	cs := NewConfigStore(db)
	existing := &domain.Config{
		Workspaces: []*domain.Workspace{
			{ID: "ws-baru", Name: "Baru", CreatedAt: now, UpdatedAt: now, Sessions: []*domain.SessionDef{}},
		},
	}
	if err := cs.Save(existing); err != nil {
		t.Fatalf("Save existing: %v", err)
	}

	// Migrasi tidak boleh menimpa data yang sudah ada.
	if err := MigrateIfNeeded(db, oldStore, oldScrollback); err != nil {
		t.Fatalf("MigrateIfNeeded: %v", err)
	}

	got, _ := cs.Load()
	if len(got.Workspaces) != 1 || got.Workspaces[0].ID != "ws-baru" {
		t.Fatal("migrasi seharusnya tidak menimpa data yang sudah ada")
	}
}

func TestMigrateIfNeededNoLegacyData(t *testing.T) {
	oldStore, oldScrollback := setupLegacyStores(t)
	// Tidak ada data lama — migration harus jalan tanpa error.

	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	defer db.Close()

	if err := MigrateIfNeeded(db, oldStore, oldScrollback); err != nil {
		t.Fatalf("MigrateIfNeeded tanpa data lama: %v", err)
	}

	cs := NewConfigStore(db)
	got, err := cs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Workspaces) != 0 {
		t.Fatalf("ekspektasi 0 workspace, dapat %d", len(got.Workspaces))
	}
}

func TestMigrateArchivesOldFiles(t *testing.T) {
	oldStore, oldScrollback := setupLegacyStores(t)

	now := time.Now().UTC()
	cfg := &domain.Config{
		Workspaces: []*domain.Workspace{
			{ID: "ws-arsip", Name: "Arsip", CreatedAt: now, UpdatedAt: now, Sessions: []*domain.SessionDef{}},
		},
	}
	_ = oldStore.Save(cfg)

	// Verifikasi config.json lama ada sebelum migrasi.
	jsonPath, _ := xdg.ConfigFile("teer/config.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("config.json lama seharusnya ada: %v", err)
	}

	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	defer db.Close()

	if err := MigrateIfNeeded(db, oldStore, oldScrollback); err != nil {
		t.Fatalf("MigrateIfNeeded: %v", err)
	}

	// config.json lama harus diarsipkan.
	if _, err := os.Stat(jsonPath + ".migrated"); err != nil {
		t.Fatalf("config.json.migrated seharusnya ada setelah migrasi: %v", err)
	}
}
