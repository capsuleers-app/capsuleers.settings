# Architettura

- `main.go`: app Wails v3, finestra frameless 1240x820, assets embed `frontend/dist`.
- `service.go`: unico servizio esposto al frontend (bindings in `frontend/bindings`, generati). Metodi: Installs, Profile, AddFolder/RemoveFolder, CachedNames/ResolveNames, ClientStatus, Apply, Backups, Restore, DeleteBackup, OpenBackupsFolder. Apply/Restore tornano `Result{ok, code}` (codici: client_running, bad_request, bad_backup, io), mai eccezioni per casi attesi.
- `internal/eve`: scoperta radici (Windows/macOS/Proton + `libraryfolders.vdf` + cartelle manuali), server -> `settings_*` -> `core_char_<digits>.dat`/`core_user_<digits>.dat`; rilevamento client (`/proc/*/cmdline` con `exefile.exe`, tasklist su Windows, ps su macOS).
- `internal/esi`: POST `universe/names` a lotti (404 sul lotto => split binario), cache `names.json`.
- `internal/backup`: zip + `manifest.json` in `<UserConfigDir>/capsuleers.settings/backups`, apply, restore (valida che ogni percorso stia nella `profileDir` del manifest e abbia nome `core_*_<n>.dat`), tiene gli ultimi 20.
- `internal/config`: `config.json` con le cartelle aggiunte a mano.
- Frontend: Vue 3 + TS; stato globale in `src/store.ts` (reactive), 4 viste in `src/views/`, stringhe in `src/i18n.ts`.
- `core_user`: opzione disabilitata nel prototipo; la mappatura account arriva dopo.
- Frameless su Linux/Wayland: GTK3 con `decorated=FALSE` non annuncia CSD e KWin aggiunge la sua barra (doppia barra). `frameless_gtk3.go` (cgo, tag gtk3) crea la finestra `Hidden`, a runtime pronto imposta un titlebar GTK vuoto e nascosto (= CSD annunciato) con ombra/bordo/raggio azzerati e poi `Show()`; su X11 non fa nulla. Verificato su KDE 6.7 Wayland: frame == client. Drag/min/max/close sono quelli Wails invariati (click reali non iniettati nel test).
- Rilevamento client: conta solo processi con un argomento il cui nome file e' `exefile.exe` (non una semplice menzione in cmdline: shell, grep).
