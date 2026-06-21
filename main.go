package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"teer/internal/infra/config"
	"teer/internal/service"
)

// Version di-set saat build: -ldflags "-X main.Version=v1.0.0"
var Version = "dev"

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	store, err := config.NewStore()
	if err != nil {
		log.Fatalf("gagal inisialisasi store: %v", err)
	}

	scrollback, err := config.NewScrollbackStore()
	if err != nil {
		log.Fatalf("gagal inisialisasi scrollback store: %v", err)
	}

	// Bersihkan snapshot scrollback yatim (session yang sudah dihapus dari
	// config) — best-effort, jangan halt startup bila gagal.
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
