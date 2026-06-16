// Package service memuat use-case aplikasi yang di-bind ke frontend lewat Wails.
//
// Service di sini hanya bergantung pada port domain (domain.Repository,
// domain.EventEmitter) dan abstraksi terminal — bukan pada detail infra.
package service

import (
	"errors"
	"maps"
	"time"

	"github.com/google/uuid"

	"teer/internal/domain"
)

// WorkspaceService mengelola CRUD workspace & definisi sesi serta
// persistensinya. Di-bind ke frontend.
type WorkspaceService struct {
	repo domain.Repository
}

// NewWorkspaceService membuat service dengan repository terinjeksi.
func NewWorkspaceService(repo domain.Repository) *WorkspaceService {
	return &WorkspaceService{repo: repo}
}

// ---- Workspace ----

// ListWorkspaces mengembalikan seluruh workspace yang tersimpan.
func (w *WorkspaceService) ListWorkspaces() ([]*domain.Workspace, error) {
	cfg, err := w.repo.Load()
	if err != nil {
		return nil, err
	}
	return cfg.Workspaces, nil
}

// CreateWorkspace membuat workspace baru dan mengembalikannya.
func (w *WorkspaceService) CreateWorkspace(name, color, defaultCwd string) (*domain.Workspace, error) {
	if name == "" {
		return nil, errors.New("workspace: nama wajib diisi")
	}
	cfg, err := w.repo.Load()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ws := &domain.Workspace{
		ID:         uuid.NewString(),
		Name:       name,
		Color:      color,
		DefaultCwd: defaultCwd,
		Env:        map[string]string{},
		Sessions:   []*domain.SessionDef{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	cfg.Workspaces = append(cfg.Workspaces, ws)
	if err := w.repo.Save(cfg); err != nil {
		return nil, err
	}
	return ws, nil
}

// UpdateWorkspace menyimpan perubahan metadata workspace (nama, warna, cwd, env).
func (w *WorkspaceService) UpdateWorkspace(updated *domain.Workspace) error {
	cfg, err := w.repo.Load()
	if err != nil {
		return err
	}
	ws := findWorkspace(cfg, updated.ID)
	if ws == nil {
		return errors.New("workspace: tidak ditemukan")
	}
	ws.Name = updated.Name
	ws.Color = updated.Color
	ws.DefaultCwd = updated.DefaultCwd
	if updated.Env != nil {
		ws.Env = updated.Env
	}
	ws.UpdatedAt = time.Now()
	return w.repo.Save(cfg)
}

// DeleteWorkspace menghapus workspace beserta definisi sesinya.
func (w *WorkspaceService) DeleteWorkspace(id string) error {
	cfg, err := w.repo.Load()
	if err != nil {
		return err
	}
	out := cfg.Workspaces[:0]
	found := false
	for _, ws := range cfg.Workspaces {
		if ws.ID == id {
			found = true
			continue
		}
		out = append(out, ws)
	}
	if !found {
		return errors.New("workspace: tidak ditemukan")
	}
	cfg.Workspaces = out
	return w.repo.Save(cfg)
}

// ReorderWorkspaces mengatur ulang urutan workspace sesuai slice ids (FR-6).
func (w *WorkspaceService) ReorderWorkspaces(ids []string) error {
	cfg, err := w.repo.Load()
	if err != nil {
		return err
	}
	byID := make(map[string]*domain.Workspace, len(cfg.Workspaces))
	for _, ws := range cfg.Workspaces {
		byID[ws.ID] = ws
	}
	out := make([]*domain.Workspace, 0, len(cfg.Workspaces))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if ws, ok := byID[id]; ok {
			out = append(out, ws)
			seen[id] = true
		}
	}
	// Workspace yang tidak ada dalam ids tetap ditambahkan di akhir.
	for _, ws := range cfg.Workspaces {
		if !seen[ws.ID] {
			out = append(out, ws)
		}
	}
	cfg.Workspaces = out
	return w.repo.Save(cfg)
}

// DuplicateWorkspace menyalin workspace beserta definisi sesinya (FR-5).
// Id baru di-generate untuk workspace dan tiap sesi; tidak ada PTY yang dibuat.
func (w *WorkspaceService) DuplicateWorkspace(id string) (*domain.Workspace, error) {
	cfg, err := w.repo.Load()
	if err != nil {
		return nil, err
	}
	src := findWorkspace(cfg, id)
	if src == nil {
		return nil, errors.New("workspace: tidak ditemukan")
	}

	now := time.Now()
	dup := &domain.Workspace{
		ID:         uuid.NewString(),
		Name:       src.Name + " (copy)",
		Color:      src.Color,
		DefaultCwd: src.DefaultCwd,
		Env:        cloneStringMap(src.Env),
		Sessions:   []*domain.SessionDef{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	for _, sd := range src.Sessions {
		nd := *sd
		nd.ID = uuid.NewString()
		nd.WorkspaceID = dup.ID
		nd.Env = cloneStringMap(sd.Env)
		dup.Sessions = append(dup.Sessions, &nd)
	}

	cfg.Workspaces = append(cfg.Workspaces, dup)
	if err := w.repo.Save(cfg); err != nil {
		return nil, err
	}
	return dup, nil
}

// ---- Session definition ----

// AddSession menambahkan definisi sesi baru ke sebuah workspace.
func (w *WorkspaceService) AddSession(workspaceID, name, shell, cwd, startupCommand string) (*domain.SessionDef, error) {
	cfg, err := w.repo.Load()
	if err != nil {
		return nil, err
	}
	ws := findWorkspace(cfg, workspaceID)
	if ws == nil {
		return nil, errors.New("workspace: tidak ditemukan")
	}
	if cwd == "" {
		cwd = ws.DefaultCwd
	}
	sd := &domain.SessionDef{
		ID:             uuid.NewString(),
		WorkspaceID:    workspaceID,
		Name:           orDefault(name, "terminal"),
		Shell:          shell,
		StartupCommand: startupCommand,
		Cwd:            cwd,
		Env:            map[string]string{},
		AutoStart:      false,
	}
	ws.Sessions = append(ws.Sessions, sd)
	ws.UpdatedAt = time.Now()
	if err := w.repo.Save(cfg); err != nil {
		return nil, err
	}
	return sd, nil
}

// UpdateSession menyimpan perubahan definisi sesi (rename, env, autoStart, dll).
func (w *WorkspaceService) UpdateSession(updated *domain.SessionDef) error {
	cfg, err := w.repo.Load()
	if err != nil {
		return err
	}
	ws := findWorkspace(cfg, updated.WorkspaceID)
	if ws == nil {
		return errors.New("workspace: tidak ditemukan")
	}
	for i, sd := range ws.Sessions {
		if sd.ID == updated.ID {
			ws.Sessions[i] = updated
			ws.UpdatedAt = time.Now()
			return w.repo.Save(cfg)
		}
	}
	return errors.New("session: tidak ditemukan")
}

// DeleteSession menghapus definisi sesi dari workspace.
func (w *WorkspaceService) DeleteSession(workspaceID, sessionID string) error {
	cfg, err := w.repo.Load()
	if err != nil {
		return err
	}
	ws := findWorkspace(cfg, workspaceID)
	if ws == nil {
		return errors.New("workspace: tidak ditemukan")
	}
	out := ws.Sessions[:0]
	found := false
	for _, sd := range ws.Sessions {
		if sd.ID == sessionID {
			found = true
			continue
		}
		out = append(out, sd)
	}
	if !found {
		return errors.New("session: tidak ditemukan")
	}
	ws.Sessions = out
	ws.UpdatedAt = time.Now()
	return w.repo.Save(cfg)
}

// ---- Helpers ----

func findWorkspace(cfg *domain.Config, id string) *domain.Workspace {
	for _, ws := range cfg.Workspaces {
		if ws.ID == id {
			return ws
		}
	}
	return nil
}

func cloneStringMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	maps.Copy(out, m)
	return out
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
