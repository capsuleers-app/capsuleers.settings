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

# TLS for the webview. The bundled libgio looks for its TLS backend in the build host's compiled-in
# path (/usr/lib/x86_64-linux-gnu/gio/modules), absent on Fedora/Arch/...: without it WebKit cannot
# load any https resource (portraits from images.evetech.net fall back to initials). Bundle the
# build host's glib-networking module (same glib as the bundled libgio), point GIO_MODULE_DIR at it
# and repack with the linuxdeploy the bundler left in the build dir.
gnutls=$(pkg-config --variable=giomoduledir gio-2.0 2>/dev/null || echo /usr/lib/x86_64-linux-gnu/gio/modules)/libgiognutls.so
[ -f "$gnutls" ] || { echo "missing $gnutls: install glib-networking" >&2; exit 1; }
appdir=$(echo "$work"/build/*.AppDir)
mkdir -p "$appdir/usr/lib/gio/modules"
cp "$gnutls" "$appdir/usr/lib/gio/modules/"
grep -q GIO_MODULE_DIR "$appdir/AppRun" ||
  sed -i 's|^cd "\$usr/lib"$|export GIO_MODULE_DIR="$usr/lib/gio/modules"\n&|' "$appdir/AppRun"
grep -q GIO_MODULE_DIR "$appdir/AppRun" || { echo "could not patch AppRun" >&2; exit 1; }
rm -f "$work"/out/*.AppImage "$work"/build/*.AppImage.tmp
(cd "$work/build" && OUTPUT=repacked.AppImage NO_STRIP=1 ./linuxdeploy-*.AppImage --appimage-extract-and-run \
  --appdir "$appdir" --deploy-deps-only "$appdir/usr/lib/gio/modules" --output appimage)
mv "$work/build/repacked.AppImage" "$work/out/"

mkdir -p "$(dirname "$out")"
mv "$work"/out/*.AppImage "$out"
