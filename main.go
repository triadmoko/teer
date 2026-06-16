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

func main() {
	store, err := config.NewStore()
	if err != nil {
		log.Fatalf("gagal inisialisasi store: %v", err)
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

	app.RegisterService(application.NewService(service.NewSessionService(app.Event)))
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
		BackgroundColour: application.NewRGB(18, 18, 20),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
