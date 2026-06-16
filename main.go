package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"teer/internal/infra/config"
	"teer/internal/service"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	store, err := config.NewStore()
	if err != nil {
		log.Fatalf("gagal inisialisasi store: %v", err)
	}

	app := application.New(application.Options{
		Name:        "teer",
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

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "teer",
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
