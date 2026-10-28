package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
)

func init() { installDesktopEntry = writeDesktopEntry }

// writeDesktopEntry installs ~/.local/share/applications/<appID>.desktop and the hicolor icon, so
// Wayland compositors (KDE, GNOME) show the app icon in the taskbar instead of a generic one.
// It points Exec at the running binary and rewrites only what changed; failures are logged, never fatal.
func writeDesktopEntry() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		data = filepath.Join(home, ".local", "share")
	}
	entry := "[Desktop Entry]\n" +
		"Type=Application\n" +
		"Name=Capsuleers Settings\n" +
		"Comment=Copy EVE Online client settings from one character to many\n" +
		"Exec=\"" + exe + "\"\n" +
		"Icon=" + appID + "\n" +
		"Terminal=false\n" +
		"Categories=Game;Utility;\n" +
		"StartupWMClass=" + appID + "\n"
	writeIfChanged(filepath.Join(data, "icons", "hicolor", "256x256", "apps", appID+".png"), appIcon)
	writeIfChanged(filepath.Join(data, "applications", appID+".desktop"), []byte(entry))
}

func writeIfChanged(path string, b []byte) {
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, b) {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Printf("desktop entry: %v", err)
		return
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		log.Printf("desktop entry: %v", err)
	}
}
