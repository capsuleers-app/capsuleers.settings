// capsuleers.settings copies EVE Online client settings from one character to many.
package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/capsuleers-app/capsuleers.settings/internal/esi"
	"github.com/capsuleers-app/capsuleers.settings/internal/update"

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
	startUpdater(app, win)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// startUpdater wires the silent self-update (internal/update): release builds only, in the
// background; errors are only logged. A downloaded update is applied on "Restart & install" or,
// if postponed, when the app quits.
func startUpdater(app *application.App, win *application.WebviewWindow) {
	if _, ok := update.ParseVersion(esi.AppVersion); !ok {
		return // dev build
	}
	exe, _ := os.Executable()
	if _, file, ok := update.Target(runtime.GOOS, os.Getenv("APPIMAGE"), exe); ok {
		update.Cleanup(file)
	} else {
		return
	}
	var (
		mu      sync.Mutex
		pending *update.Pending
	)
	// applyPending swaps the downloaded file in (once).
	applyPending := func() bool {
		mu.Lock()
		defer mu.Unlock()
		if pending == nil {
			return false
		}
		p := pending
		pending = nil
		if err := p.Apply(); err != nil {
			log.Println("update: apply:", err)
			p.Discard()
			return false
		}
		return true
	}
	app.OnShutdown(func() { applyPending() })
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), update.CheckTimeout)
		defer cancel()
		u := &update.Updater{Current: esi.AppVersion, UserAgent: esi.UserAgent, Client: &http.Client{}}
		p, err := u.Check(ctx, runtime.GOOS, os.Getenv("APPIMAGE"), exe)
		if err != nil {
			log.Println("update:", err)
			return
		}
		if p == nil {
			return
		}
		mu.Lock()
		pending = p
		mu.Unlock()
		t := update.Texts(update.Lang(update.SystemLocale()), p.Version)
		d := app.Dialog.Question().SetTitle(t.Title).SetMessage(t.Message + "\n\n" + t.Detail).AttachToWindow(win)
		d.AddButton(t.Later).SetAsCancel()
		d.AddButton(t.Restart).SetAsDefault().OnClick(func() {
			if !applyPending() {
				return
			}
			if err := exec.Command(p.File).Start(); err != nil {
				log.Println("update: relaunch:", err)
				return
			}
			app.Quit()
		})
		d.Show()
	}()
}
