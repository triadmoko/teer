package service

import (
	"testing"

	"teer/internal/domain"
)

type memRepo struct {
	cfg *domain.Config
}

func newMemRepo() *memRepo {
	return &memRepo{cfg: &domain.Config{Workspaces: []*domain.Workspace{}}}
}

func (m *memRepo) Load() (*domain.Config, error) { return m.cfg, nil }
func (m *memRepo) Save(cfg *domain.Config) error { m.cfg = cfg; return nil }

func TestWorkspaceCRUD(t *testing.T) {
	svc := NewWorkspaceService(newMemRepo())

	list, err := svc.ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("harusnya kosong, dapat %d", len(list))
	}

	ws, err := svc.CreateWorkspace("Proyek A", "#60a5fa", "/tmp")
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if ws == nil || ws.ID == "" {
		t.Fatalf("workspace nil/tanpa id")
	}

	list2, err := svc.ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces#2: %v", err)
	}
	if len(list2) != 1 || list2[0].Name != "Proyek A" {
		t.Fatalf("persistensi gagal: %+v", list2)
	}

	sd, err := svc.AddSession(ws.ID, "term 1", "", "")
	if err != nil {
		t.Fatalf("AddSession: %v", err)
	}
	if sd.ID == "" {
		t.Fatalf("sesi tanpa id")
	}
	if sd.Cwd != "/tmp" {
		t.Fatalf("cwd harus mewarisi defaultCwd, dapat %q", sd.Cwd)
	}
}
