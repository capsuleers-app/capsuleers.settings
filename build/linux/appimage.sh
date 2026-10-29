#!/usr/bin/env bash
# Package a built Linux binary as an AppImage with Wails v3's own bundler
# (`wails3 generate appimage`: linuxdeploy + GTK plugin, bundles GTK3 + WebKitGTK 4.1).
# Usage: build/linux/appimage.sh <binary named capsuleers.settings> <output .AppImage>
# Needs wails3 in PATH (go install -tags gtk3 github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0) and network
# (it downloads linuxdeploy). Uses the extract-and-run mode, so FUSE is not required.
set -euo pipefail
bin=$(realpath "$1"); out=$(realpath -m "$2")
here=$(cd "$(dirname "$0")" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir "$work/in" "$work/build" "$work/out"
cp "$bin" "$work/in/capsuleers.settings"
cp "$here/../appicon.png" "$work/in/app.capsuleers.settings.png"
export APPIMAGE_EXTRACT_AND_RUN=1
wails3 generate appimage \
  -binary "$work/in/capsuleers.settings" \
  -icon "$work/in/app.capsuleers.settings.png" \
  -desktopfile "$here/capsuleers.settings.desktop" \
  -builddir "$work/build" -outputdir "$work/out"
mkdir -p "$(dirname "$out")"
mv "$work"/out/*.AppImage "$out"
