// capsuleers.settings copies EVE Online client settings from one character to many.
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

// appIcon is the brand icon (docs/brand, from the site's generator): window and taskbar icon on
// Linux. The Windows .exe icon comes from rsrc_windows_amd64.syso (build/windows/icon.ico).
//
//go:embed build/appicon.png
var appIcon []byte

// appID is the Linux application id: Wayland compositors match it to a .desktop file of the same
// name to show the taskbar icon (see desktop_linux.go).
const appID = "app.capsuleers.settings"

// installDesktopEntry is set on Linux: it keeps the user's .desktop entry and icon current.
var installDesktopEntry func()

// framelessFix is set on Linux/GTK3 builds: it removes the compositor title bar that Wayland
// adds around a window GTK3 declared undecorated. It owns showing the window.
var framelessFix func(*application.WebviewWindow)

func main() {
	svc, err := NewService()
	if err != nil {
		log.Fatal(err)
	}
	if installDesktopEntry != nil {
		installDesktopEntry()
	}
	app := application.New(application.Options{
		Name:        "capsuleers.settings",
		Description: "Copy EVE Online client settings from one character to many",
		Icon:        appIcon,
		Services:    []application.Service{application.NewService(svc)},
		Assets:      application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		Linux:       application.LinuxOptions{ApplicationID: appID, ProgramName: appID},
		Mac:         application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "CAPSULEERS.SETTINGS",
		Width:            1240,
		Height:           820,
		MinWidth:         900,
		MinHeight:        620,
		Frameless:        true,
		BackgroundColour: application.NewRGB(0x14, 0x14, 0x17),
		URL:              "/",
		Linux:            application.LinuxWindow{Icon: appIcon},
		Hidden:           framelessFix != nil,
	})
	if framelessFix != nil {
		framelessFix(win)
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
