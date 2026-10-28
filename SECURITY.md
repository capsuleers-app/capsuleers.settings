# Security Policy

Only the latest GitHub Release is supported.

## Reporting a vulnerability

**Do not open a public issue for security problems.** Use GitHub Security Advisories
([Report a vulnerability](https://github.com/WilliamFalci/capsuleers.settings/security/advisories/new))
or email [info@capsuleers.app](mailto:info@capsuleers.app).

Include the impact, app version and OS, reproduction steps, and scrubbed logs. We aim to
acknowledge within 72 hours and give a fix timeline within 7 days.

## Scope notes

The app writes only `core_char_*.dat` files inside a user-chosen `settings_*` folder and its own
backups under the user config directory. Restoring validates every path from the backup manifest.
The only network calls are ESI `universe/names` and portraits from images.evetech.net.
