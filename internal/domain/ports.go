package domain

type Repository interface {
	Load() (*Config, error)
	Save(cfg *Config) error
}

type Scrollback interface {
	Save(id, data string) error
	Load(id string) (string, error)
	Delete(id string) error
	Prune(validIDs []string) error
}

type EventEmitter interface {
	Emit(name string, data ...any) bool
}
