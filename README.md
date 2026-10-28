# capsuleers.settings

Copy your EVE Online client settings (window layout, chat channels, shortcuts) from one character
to many, safely. Part of the [Capsuleers](https://capsuleers.app) family. Unofficial fan-made tool.

- Finds your settings automatically: Windows, macOS, and Linux (Steam / Proton, Flatpak Steam, extra Steam libraries). Add any folder by hand.
- Pick a profile, one source character, and the targets. Review, then apply.
- Every target is zipped to a backup first; copies are atomic; restore any of the last 20 backups.
- Refuses to write while an EVE client is running (it would overwrite your change on exit).
- Character names from ESI (no login), cached; portraits from images.evetech.net. Works offline with IDs.
- English and Italian from the system locale. Account settings (`core_user`) are planned, disabled for now.

Nothing is written until you press Apply, and never outside the chosen `settings_*` folder.

## Build

Go 1.26+, Node.js, [Wails v3](https://v3.wails.io). On Linux you need GTK3 and webkit2gtk-4.1 headers.

```bash
cd frontend && npm install && npm run build && cd ..
go build -tags "gtk3 production" -o bin/capsuleers.settings .                      # Linux
GOOS=windows CGO_ENABLED=0 go build -tags production -ldflags "-H windowsgui" -o bin/capsuleers.settings.exe .
go test ./internal/...
```

Regenerate bindings after editing `service.go`: `wails3 generate bindings -ts -f "-tags gtk3"`.

## License

MIT, see [LICENSE](LICENSE) and [THIRD_PARTY.md](THIRD_PARTY.md).
