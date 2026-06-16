// Package domain memuat entitas bisnis inti aplikasi beserta port (interface)
// yang harus dipenuhi oleh lapisan luar (infra & delivery).
//
// Lapisan ini tidak bergantung pada framework apa pun (Wails, XDG, dll) — hanya
// pustaka standar. Arah dependensi selalu menuju ke dalam (clean architecture).
package domain

import "time"

// Workspace adalah kumpulan sesi terminal yang dikelompokkan per-proyek.
// Struktur ini dipersist ke file konfig (lihat lapisan infra/config).
type Workspace struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Color          string            `json:"color"`
	DefaultCwd     string            `json:"defaultCwd"`
	StartupCommand string            `json:"startupCommand"`
	Env            map[string]string `json:"env"`
	Sessions       []*SessionDef     `json:"sessions"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// SessionDef adalah definisi sebuah sesi terminal yang dipersist.
// Status runtime (running/exited, PID) TIDAK disimpan di sini — itu dikelola
// oleh SessionService secara in-memory (PRD §5.2).
type SessionDef struct {
	ID          string            `json:"id"`
	WorkspaceID string            `json:"workspaceId"`
	Name        string            `json:"name"`
	Shell       string            `json:"shell"`
	Cwd         string            `json:"cwd"`
	Env         map[string]string `json:"env"`
	AutoStart   bool              `json:"autoStart"`
	Layout      map[string]any    `json:"layout,omitempty"`
}

// Config adalah seluruh state yang dipersist ke disk.
type Config struct {
	Workspaces []*Workspace `json:"workspaces"`
}
