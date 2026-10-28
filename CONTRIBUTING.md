# Contributing to capsuleers.settings

Thanks for helping! Keep changes small and focused on copying EVE client settings safely.

## Dev setup

Go 1.26+, Node.js LTS, Wails v3 CLI (`go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0`).
Linux builds need GTK3 + webkit2gtk-4.1 development headers and the `gtk3` build tag.

```bash
cd frontend && npm install && npm run build && cd ..
go vet -tags gtk3 ./... && go test -tags gtk3 ./...
```

## Conventions

- Safety first: no write without an explicit Apply; never outside the chosen profile folder; always back up first; refuse while the client runs. Add a test for any change near these rules.
- Outbound `User-Agent` and `X-Compatibility-Date` come only from `internal/esi/esi.go`.
- UI follows the Capsuleers.App design system: dark only, square corners, one orange accent, hover darker. No CDN assets.
- Every visible string goes in `frontend/src/i18n.ts` in both English and Italian.
- Icons come from the Capsuleers brand generator; do not edit them by hand.
