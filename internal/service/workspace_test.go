package service

import (
	"testing"

	"teer/internal/domain"
)

// memRepo adalah domain.Repository in-memory untuk menguji use-case tanpa disk.
// Inilah keuntungan inversi dependensi: WorkspaceService bisa diuji murni
// terhadap port-nya, lepas dari XDG/filesystem/Wails.
type memRepo struct {
	cfg *domain.Config
}

func newMemRepo() *memRepo {
	return &memRepo{cfg: &domain.Config{Workspaces: []*domain.Workspace{}}}
}

func (m *memRepo) Load() (*domain.Config, error) { return m.cfg, nil }
func (m *memRepo) Save(cfg *domain.Config) error { m.cfg = cfg; return nil }

// TestWorkspaceCRUD memverifikasi alur Create/List/AddSession lewat port repo.
func TestWorkspaceCRUD(t *testing.T) {
	svc := NewWorkspaceService(newMemRepo())

	// Awalnya kosong.
	list, err := svc.ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("harusnya kosong, dapat %d", len(list))
	}

	// Buat.
	ws, err := svc.CreateWorkspace("Proyek A", "#60a5fa", "/tmp")
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if ws == nil || ws.ID == "" {
		t.Fatalf("workspace nil/tanpa id")
	}

	// Persisten lewat repo: list kedua harus melihatnya.
	list2, err := svc.ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces#2: %v", err)
	}
	if len(list2) != 1 || list2[0].Name != "Proyek A" {
		t.Fatalf("persistensi gagal: %+v", list2)
	}

	// Tambah sesi; cwd kosong harus mewarisi DefaultCwd workspace.
	sd, err := svc.AddSession(ws.ID, "term 1", "", "", "")
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
