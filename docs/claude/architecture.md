# Architettura

- `main.go`: app Wails v3, finestra frameless 1240x820, assets embed `frontend/dist`.
- `service.go`: unico servizio esposto al frontend (bindings in `frontend/bindings`, generati). Metodi: Installs, Profile, AddFolder/RemoveFolder, CachedNames/ResolveNames, ClientStatus, Apply, Backups, Restore, DeleteBackup, OpenBackupsFolder. Apply/Restore tornano `Result{ok, code}` (codici: client_running, bad_request, bad_backup, io), mai eccezioni per casi attesi.
- `internal/eve`: scoperta radici (Windows/macOS/Proton + `libraryfolders.vdf` + cartelle manuali), server -> `settings_*` -> `core_char_<digits>.dat`/`core_user_<digits>.dat`; rilevamento client (`/proc/*/cmdline` con `exefile.exe`, tasklist su Windows, ps su macOS).
- `internal/esi`: POST `universe/names` a lotti (404 sul lotto => split binario), cache `names.json`.
- `internal/backup`: zip + `manifest.json` in `<UserConfigDir>/capsuleers.settings/backups`, apply, restore (valida che ogni percorso stia nella `profileDir` del manifest e abbia nome `core_*_<n>.dat`), tiene gli ultimi 20.
- `internal/config`: `config.json` con le cartelle aggiunte a mano.
- Frontend: Vue 3 + TS; stato globale in `src/store.ts` (reactive), 4 viste in `src/views/`, stringhe in `src/i18n.ts`.
- `core_user`: opzione disabilitata nel prototipo; la mappatura account arriva dopo.
- Nota: su KDE/Wayland la finestra frameless mostra comunque la barra di sistema (da indagare).
