# capsuleers.settings

App desktop Go + Wails v3 (Windows, Linux, macOS) che copia le impostazioni del client EVE Online
(`core_char_<id>.dat`) da un personaggio a molti, con backup zip prima di ogni scrittura. Open source
(MIT), nessun server né DB; unica rete: ESI `universe/names` e ritratti da images.evetech.net.
Percorsi relativi a questa cartella; per regole di orchestrazione vedi `../CLAUDE.md`.

## Comandi
Host Bazzite senza header GTK: build/vet/test Go nel container `evebuild` (`podman exec -w <repo> evebuild ...`,
`toolbox run` non funziona: container troppo vecchio). Linux usa GTK3: sempre `-tags gtk3` (Wails v3 default GTK4).
```bash
cd frontend && npm install && npm run build         # Vue+Vite -> frontend/dist (embedded)
wails3 generate bindings -ts -f "-tags gtk3"        # dopo ogni modifica a service.go (nel container)
go vet -tags gtk3 ./... && go test -tags gtk3 ./... # nel container; sull'host: go test ./internal/...
go build -buildvcs=false -tags "gtk3 production" -ldflags "-s -w" -o bin/capsuleers.settings .
GOOS=windows CGO_ENABLED=0 go build -tags production -ldflags "-s -w -H windowsgui" -o bin/capsuleers.settings.exe .
```
File creati nel container dal suo utente risultano illeggibili sull'host: `podman unshare chown -R 0:0 <path>`.

## Vincoli
- Mai scrivere senza Apply esplicito; mai fuori dalla cartella `settings_*` scelta; backup zip prima, poi copia atomica (temp + rename).
- Apply e Restore rifiutano se il client gira (`eve.RunningCount`).
- `User-Agent` e `X-Compatibility-Date` solo da `internal/esi/esi.go`; la versione la imprime la CI (`-X ...esi.AppVersion`).
- CI (`.github/workflows/build.yml`, come intel): ogni push su `main` = vet/test + build Linux/Windows + Release `v<major.minor>.<run>`. Dependabot in `.github/dependabot.yml`.
- UI: design system Capsuleers.App (solo dark, radius 0, UN arancione `#fb923c`, hover piu' scuro, `bundle.css` verbatim). Font e asset locali, nessuna CDN; CSP in `frontend/index.html`.
- Bilingue EN/IT dal locale di sistema (`frontend/src/i18n.ts`), nessuno switch.
- Icone: non modificarle qui (vedi brand.md).

## Aree (leggi solo se il task le tocca)
| area | file | leggi se |
|---|---|---|
| architettura | docs/claude/architecture.md | tocchi `internal/`, `service.go`, store/viste del frontend |
| marchio | ../docs/claude/brand.md | tocchi icone o `docs/brand/` |
| rapporti tra repo | ../docs/claude/cross-repo.md | improbabile: il repo e' autonomo |
