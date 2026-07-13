# API

## Stil
REST für Konfiguration und Ressourcen, WebSocket für Live-Events. JSON-only. Versionierung über /api/v1.

## Identity (v1, implementiert)

| Methode | Pfad | Auth | Beschreibung |
|---------|------|------|--------------|
| POST | `/api/v1/identity/bootstrap` | Loopback | Ersten Owner anlegen (nur wenn keine aktiven Benutzer) |
| POST | `/api/v1/identity/login` | — | Anmeldung; bei Erfolg **200** mit `{ "csrf_token": "..." }` und `vyntrio_session`-Cookie |
| POST | `/api/v1/identity/logout` | Session + `X-CSRF-Token` | Session widerrufen und Cookie löschen |
| GET | `/api/v1/settings` | Session; Permission `settings:admin:read` (Owner) | Sichere Admin-Settings-Ansicht (read-only) |

Weitere Probes ohne Auth: `/healthz`, `/readyz`, `/api/v1/version`.

Kein JWT, kein Refresh-Token in v1. Session ist ein opaques serverseitiges Cookie (`vyntrio_session`, HttpOnly).

## Resource Families (geplant)
- /users
- /systems
- /disks
- /pools
- /shares
- /containers
- /vms
- /networks
- /licenses
- /updates
- /logs
- /notifications
- /plugins
- /apps

## API-Konventionen
- Request IDs.
- Standardisierte Fehlerantworten.
- Pagination/Filter/Sortierung.
- OpenAPI-Spezifikation als Pflicht.
- WebSocket Event Namespaces pro Modul.
