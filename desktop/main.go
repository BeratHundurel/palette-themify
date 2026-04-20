package main

import (
	"embed"
	"log"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	themeExportService := &ThemeExportService{}

	app := application.New(application.Options{
		Name:        "themesmith",
		Description: "Desktop app for ThemeSmith",
		Services: []application.Service{
			application.NewService(themeExportService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	isMac := runtime.GOOS == "darwin"

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:      "ThemeSmith",
		Frameless:  !isMac,
		StartState: application.WindowStateMaximised,
		Width:      1440,
		Height:     900,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
