// capsuleers.settings copies EVE Online client settings from one character to many.
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	svc, err := NewService()
	if err != nil {
		log.Fatal(err)
	}
	app := application.New(application.Options{
		Name:        "capsuleers.settings",
		Description: "Copy EVE Online client settings from one character to many",
		Services:    []application.Service{application.NewService(svc)},
		Assets:      application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Mac:         application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "CAPSULEERS.SETTINGS",
		Width:            1240,
		Height:           820,
		MinWidth:         900,
		MinHeight:        620,
		Frameless:        true,
		BackgroundColour: application.NewRGB(0x14, 0x14, 0x17),
		URL:              "/",
	})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
