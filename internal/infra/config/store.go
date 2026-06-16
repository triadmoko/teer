package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"

	"teer/internal/domain"
)

type Store struct {
	path string
	mu   sync.Mutex
}

var _ domain.Repository = (*Store)(nil)

const configRelPath = "teer/config.json"

func NewStore() (*Store, error) {
	path, err := xdg.ConfigFile(configRelPath)
	if err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}

func (s *Store) Load() (*domain.Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &domain.Config{Workspaces: []*domain.Workspace{}}, nil
		}
		return nil, err
	}

	var cfg domain.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Workspaces == nil {
		cfg.Workspaces = []*domain.Workspace{}
	}
	return &cfg, nil
}

func (s *Store) Save(cfg *domain.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
