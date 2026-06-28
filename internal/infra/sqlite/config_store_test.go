package sqlite

import (
	"testing"
	"time"

	"teer/internal/domain"
)

func newTestDB(t *testing.T) *ConfigStore {
	t.Helper()
	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewConfigStore(db)
}

func TestConfigStoreEmptyLoad(t *testing.T) {
	cs := newTestDB(t)

	cfg, err := cs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Workspaces) != 0 {
		t.Fatalf("ekspektasi 0 workspace, dapat %d", len(cfg.Workspaces))
	}
}

func TestConfigStoreRoundTrip(t *testing.T) {
	cs := newTestDB(t)

	now := time.Now().UTC().Truncate(time.Second)
	cfg := &domain.Config{
		Workspaces: []*domain.Workspace{
			{
				ID:             "ws-1",
				Name:           "Backend",
				Color:          "#ff0000",
				DefaultCwd:     "/home/user/backend",
				StartupCommand: "echo start",
				Env:            map[string]string{"GO_ENV": "dev"},
				CreatedAt:      now,
				UpdatedAt:      now,
				Sessions: []*domain.SessionDef{
					{
						ID:          "sess-1",
						WorkspaceID: "ws-1",
						Name:        "Server",
						Shell:       "/bin/bash",
						Cwd:         "/home/user/backend",
						AutoStart:   true,
						Env:         map[string]string{"PORT": "8080"},
					},
				},
			},
		},
	}

	if err := cs.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := cs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(got.Workspaces) != 1 {
		t.Fatalf("ekspektasi 1 workspace, dapat %d", len(got.Workspaces))
	}

	ws := got.Workspaces[0]
	if ws.ID != "ws-1" || ws.Name != "Backend" || ws.Color != "#ff0000" {
		t.Fatalf("workspace tidak sesuai: %+v", ws)
	}
	if ws.Env["GO_ENV"] != "dev" {
		t.Fatalf("env workspace salah: %v", ws.Env)
	}

	if len(ws.Sessions) != 1 {
		t.Fatalf("ekspektasi 1 session, dapat %d", len(ws.Sessions))
	}
	sd := ws.Sessions[0]
	if sd.ID != "sess-1" || sd.Name != "Server" || !sd.AutoStart {
		t.Fatalf("session tidak sesuai: %+v", sd)
	}
	if sd.Env["PORT"] != "8080" {
		t.Fatalf("env session salah: %v", sd.Env)
	}
}

func TestConfigStoreSortOrder(t *testing.T) {
	cs := newTestDB(t)

	cfg := &domain.Config{
		Workspaces: []*domain.Workspace{
			{ID: "ws-a", Name: "A", Sessions: []*domain.SessionDef{}},
			{ID: "ws-b", Name: "B", Sessions: []*domain.SessionDef{}},
			{ID: "ws-c", Name: "C", Sessions: []*domain.SessionDef{}},
		},
	}
	if err := cs.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Simpan ulang dengan urutan berbeda.
	cfg.Workspaces = []*domain.Workspace{
		cfg.Workspaces[2],
		cfg.Workspaces[0],
		cfg.Workspaces[1],
	}
	if err := cs.Save(cfg); err != nil {
		t.Fatalf("Save reorder: %v", err)
	}

	got, err := cs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	wantOrder := []string{"ws-c", "ws-a", "ws-b"}
	for i, ws := range got.Workspaces {
		if ws.ID != wantOrder[i] {
			t.Fatalf("urutan[%d] = %s, mau %s", i, ws.ID, wantOrder[i])
		}
	}
}

func TestConfigStoreDeleteWorkspace(t *testing.T) {
	cs := newTestDB(t)

	cfg := &domain.Config{
		Workspaces: []*domain.Workspace{
			{ID: "ws-1", Name: "A", Sessions: []*domain.SessionDef{}},
			{ID: "ws-2", Name: "B", Sessions: []*domain.SessionDef{}},
		},
	}
	if err := cs.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfg.Workspaces = cfg.Workspaces[:1]
	if err := cs.Save(cfg); err != nil {
		t.Fatalf("Save setelah hapus: %v", err)
	}

	got, err := cs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Workspaces) != 1 || got.Workspaces[0].ID != "ws-1" {
		t.Fatalf("workspace tidak sesuai setelah hapus: %+v", got.Workspaces)
	}
}

func TestConfigStoreSessionsCascadeDelete(t *testing.T) {
	cs := newTestDB(t)

	cfg := &domain.Config{
		Workspaces: []*domain.Workspace{
			{
				ID:   "ws-1",
				Name: "A",
				Sessions: []*domain.SessionDef{
					{ID: "s-1", Name: "S1", WorkspaceID: "ws-1"},
					{ID: "s-2", Name: "S2", WorkspaceID: "ws-1"},
				},
			},
		},
	}
	if err := cs.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Hapus workspace → session-nya harus ikut terhapus (CASCADE).
	cfg.Workspaces = []*domain.Workspace{}
	if err := cs.Save(cfg); err != nil {
		t.Fatalf("Save kosong: %v", err)
	}

	// Cek langsung di DB bahwa session_defs kosong.
	var count int
	_ = cs.db.QueryRow(`SELECT COUNT(*) FROM session_defs`).Scan(&count)
	if count != 0 {
		t.Fatalf("session_defs harus kosong setelah cascade, dapat %d", count)
	}
}

func TestConfigStoreNilEnv(t *testing.T) {
	cs := newTestDB(t)

	cfg := &domain.Config{
		Workspaces: []*domain.Workspace{
			{
				ID:   "ws-nil",
				Name: "NilEnv",
				Env:  nil, // env nil harus ditangani
				Sessions: []*domain.SessionDef{
					{ID: "s-nil", Name: "S", WorkspaceID: "ws-nil", Env: nil},
				},
			},
		},
	}
	if err := cs.Save(cfg); err != nil {
		t.Fatalf("Save env nil: %v", err)
	}

	got, err := cs.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Env tidak boleh nil setelah load (domain convention).
	if got.Workspaces[0].Env == nil {
		t.Fatal("workspace.Env tidak boleh nil setelah load")
	}
	if got.Workspaces[0].Sessions[0].Env == nil {
		t.Fatal("session.Env tidak boleh nil setelah load")
	}
}
