package domain

// Repository memuat & mempersist seluruh konfigurasi aplikasi.
//
// Ini adalah port keluar yang dipakai lapisan service; implementasinya
// (mis. store JSON di disk) tinggal di lapisan infra. Dengan begitu service
// tidak terikat ke detail penyimpanan dan mudah diuji dengan repo palsu.
type Repository interface {
	Load() (*Config, error)
	Save(cfg *Config) error
}

// EventEmitter meng-emit event bernama ke frontend.
//
// Port ini mengabstraksi mekanisme event Wails (application.EventManager)
// sehingga SessionService tidak bergantung langsung pada framework dan bisa
// diuji tanpa membangkitkan aplikasi penuh.
type EventEmitter interface {
	Emit(name string, data ...any) bool
}
