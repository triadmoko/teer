package domain

import "time"

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

type SessionDef struct {
	ID             string            `json:"id"`
	WorkspaceID    string            `json:"workspaceId"`
	Name           string            `json:"name"`
	Shell          string            `json:"shell"`
	Cwd            string            `json:"cwd"`
	Env            map[string]string `json:"env"`
	StartupCommand string            `json:"startupCommand"`
	AutoStart      bool              `json:"autoStart"`
	Layout         map[string]any    `json:"layout,omitempty"`
}

type Config struct {
	Workspaces []*Workspace `json:"workspaces"`
}
