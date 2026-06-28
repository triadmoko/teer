package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"teer/internal/infra/config"
	"teer/internal/infra/sqlite"
	"teer/internal/service"
)

// Version di-set saat build: -ldflags "-X main.Version=v1.0.0"
var Version = "dev"

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	db, err := sqlite.Open()
	if err != nil {
		log.Fatalf("gagal inisialisasi database: %v", err)
	}

	store := sqlite.NewConfigStore(db)
	scrollback := sqlite.NewScrollbackStore(db)

	// Migrasi data lama (JSON + scrollback files) ke SQLite, bila ada.
	if oldStore, storeErr := config.NewStore(); storeErr == nil {
		if oldScrollback, sbErr := config.NewScrollbackStore(); sbErr == nil {
			if err := sqlite.MigrateIfNeeded(db, oldStore, oldScrollback); err != nil {
				log.Printf("migrasi data lama: %v", err)
			}
		}
	}

	// Bersihkan scrollback yatim (session yang sudah dihapus dari config).
	if cfg, err := store.Load(); err == nil {
		var ids []string
		for _, ws := range cfg.Workspaces {
			for _, sd := range ws.Sessions {
				ids = append(ids, sd.ID)
			}
		}
		if err := scrollback.Prune(ids); err != nil {
			log.Printf("prune scrollback: %v", err)
		}
	}

	app := application.New(application.Options{
		Name:        "Teer",
		Description: "Terminal Workspace Manager",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.RegisterService(application.NewService(service.NewSessionService(app.Event, scrollback)))
	app.RegisterService(application.NewService(service.NewWorkspaceService(store)))
	app.RegisterService(application.NewService(service.NewDialogService()))
	app.RegisterService(application.NewService(service.NewUpdaterService(Version, app.Event)))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Teer",
		Width:  1200,
		Height: 760,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		Linux: application.LinuxWindow{
			Icon: appIcon,
		},
		BackgroundColour: application.NewRGB(18, 18, 20),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
