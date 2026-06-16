package domain

type Repository interface {
	Load() (*Config, error)
	Save(cfg *Config) error
}

type EventEmitter interface {
	Emit(name string, data ...any) bool
}
