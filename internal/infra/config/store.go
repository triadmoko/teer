// Package config menyediakan implementasi domain.Repository berbasis file JSON
// di direktori konfig OS (spesifikasi XDG, PRD §13.1).
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"

	"teer/internal/domain"
)

// Store membaca/menulis konfigurasi ke file JSON di direktori konfig OS.
// Penyimpanan v1 sengaja JSON sederhana (PRD §13.1). File ditulis dengan
// permission 0600 (PRD NFR-5). Store memenuhi domain.Repository.
type Store struct {
	path string
	mu   sync.Mutex
}

// Pastikan Store memenuhi kontrak port domain.Repository pada waktu kompilasi.
var _ domain.Repository = (*Store)(nil)

const configRelPath = "teer/config.json"

// NewStore menentukan lokasi file konfig via spesifikasi XDG
// (mis. ~/.config/teer/config.json di Linux).
func NewStore() (*Store, error) {
	// xdg.ConfigFile membuat direktori induk bila perlu dan mengembalikan
	// path lengkap untuk file relatif yang diberikan.
	path, err := xdg.ConfigFile(configRelPath)
	if err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}

// Load membaca konfig dari disk. Bila file belum ada, kembalikan config kosong.
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

// Save menulis konfig ke disk secara atomik (tulis ke file temp lalu rename)
// dengan permission ketat 0600.
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
