# API

## Stil
REST für Konfiguration und Ressourcen, WebSocket für Live-Events. JSON-only. Versionierung über /api/v1.

## Identity (v1, implementiert)

| Methode | Pfad | Auth | Beschreibung |
|---------|------|------|--------------|
| POST | `/api/v1/identity/bootstrap` | Loopback | Ersten Owner anlegen (nur wenn keine aktiven Benutzer) |
| POST | `/api/v1/identity/login` | — | Anmeldung; bei Erfolg **200** mit `{ "csrf_token": "..." }` und `vyntrio_session`-Cookie |
| POST | `/api/v1/identity/logout` | Session + `X-CSRF-Token` | Session widerrufen und Cookie löschen |
| GET | `/api/v1/overview` | Session; Permission `system:health` (Owner, Operator, Read-only) | Sichere Appliance-Overview (read-only) |
| GET | `/api/v1/settings` | Session; Permission `settings:admin:read` (Owner) | Sichere Admin-Settings-Ansicht (read-only) |
| PATCH | `/api/v1/settings/instance` | Session; Permission `settings:admin:write` (Owner); `X-CSRF-Token` | Instanz-Anzeigename (`display_name`) ändern |

Weitere Probes ohne Auth: `/healthz`, `/readyz`, `/api/v1/version`.

## UI-Serving und Routen-Priorität (v1, implementiert)

Das API-Binary liefert die eingebettete produktive UI same-origin aus:

```text
Vyntrio Go-Binary
  /api/v1/*   bestehende JSON-API (unverändert)
  /healthz    operationale JSON-Probe (unverändert)
  /readyz     operationale JSON-Probe (unverändert)
  /assets/*   eingebettete, content-gehashte Vite-Assets
  GET/HEAD nicht-reservierte UI-Pfade  eingebettetes index.html (SPA-Fallback)
```

Regeln:

- API-, Health- und Readiness-Routen werden **vor** dem UI-Fallback
  registriert und behalten Priorität.
- Unbekannte API-Pfade (z. B. `GET /api/v1/does-not-exist`) liefern weiterhin
  den kanonischen JSON-404 bzw. JSON-405 — niemals `index.html`.
- Nur `GET` und `HEAD` erhalten UI-HTML; `POST`/`PUT`/`PATCH`/`DELETE` auf
  UI-Pfade bleiben kanonische Nicht-HTML-Fehler.
- Reservierte Präfixe (`/api`, `/assets`, `/healthz`, `/readyz`) sind nie
  SPA-Fallback.
- Für diese Same-Origin-Topologie sind **kein Vite-Dev-Proxy** und **keine
  CORS-Lockerung** in Produktion erforderlich.

Cache- und Security-Header-Policy: siehe `docs/17_SECURITY.md`;
Build/Embedding: siehe `docs/19_RELEASE.md`.

Kein JWT, kein Refresh-Token in v1. Session ist ein opaques serverseitiges Cookie (`vyntrio_session`, HttpOnly).

### GET `/api/v1/overview`

- **Auth:** Session erforderlich; Permission `system:health` (alle Rollen mit Health-Zugriff).
- **CSRF:** nicht erforderlich (read-only GET).
- **Cache:** `Cache-Control: no-store`.
- **Erfolg:** **200** mit strikt begrenztem DTO:
  - `instance`: `{ "name", "version", "commit" }`
  - `api`: `{ "environment" }`
  - `service`: `{ "status" }` — `"running"` nur während der Handler ausgeführt wird
  - `readiness`: `{ "status", "database" }` — gleiche Semantik wie `/readyz` (`ready`/`not_ready`, `ok`/`error`)
  - `collected_at`: UTC-Zeitstempel (RFC3339Nano)
- **Readiness in 200:** Datenbankfehler liefern **200** mit `readiness.status = "not_ready"` und `readiness.database = "error"` — kein **503**, keine Behauptung voller Appliance-Gesundheit.
- **Fehler:** **401** fehlende Session; **403** fehlende Permission; **500** interner Fehler — kanonisches JSON-Fehler-Envelope.
- Keine Secrets, Pfade, Config-Inhalte, Backup-Details, Session-Werte oder Rohfehler in der Antwort.

### PATCH `/api/v1/settings/instance`

- **Request:** `Content-Type: application/json`, Body `{ "display_name": "<name>" }` (max. 4 KiB).
- **Validierung:** Pflichtfeld `display_name`; Unicode-Trim; max. 80 Zeichen; keine Steuerzeichen; keine unbekannten JSON-Felder.
- **Erfolg:** **200** `{ "display_name": "<persistierter Name>" }`; kein Cookie-Change.
- **No-Op:** Identischer normalisierter Name → **200** ohne DB-Update, Timestamp-Update oder Audit.
- **Fehler:** **400** ungültige Anfrage; **401** fehlende Session; **403** fehlende Berechtigung oder CSRF; **405** andere Methoden; **500** Persistenzfehler.
- Persistiert in `system.hostname` (kanonischer Speicherort für `instance.name` in GET `/api/v1/settings`).

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
