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
| GET | `/api/v1/storage/disks` | Session; Permission `storage:read` (Owner, Operator, Read-only) | Read-only Blockgeräte-Inventar mit Eligibility (Slice 12.1) |
| GET | `/api/v1/storage/pools` | Session; Permission `storage:read` | Declared pools (Slice 12.3); `disk_format_applied: false` |
| GET | `/api/v1/storage/shares` | Session; Permission `storage:read` | Planned shares (Slice 12.3); SMB/NFS noch nicht verfügbar |
| POST | `/api/v1/storage/pools` | Session; `storage:write`; CSRF | Declare pool from eligible disks (no disk format) |
| POST | `/api/v1/storage/pools/{poolID}/datasets` | Session; `storage:write`; CSRF | Prepare dataset plan under a declared pool |
| POST | `/api/v1/storage/shares` | Session; `storage:write`; CSRF | Prepare share plan (protocol remains planned) |

Weitere Probes ohne Auth: `/healthz`, `/readyz`, `/api/v1/version`, `/api/v1/public/install-media`.

Optional (wenn `VYNTRIO_RELEASE_STAGING_DIR` gesetzt): `/release/vyntrio-install-media.img`, `/release/release-manifest.json`.

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
  - `host`: CPU-, Speicher- und State-Filesystem-Metriken (Slice 8.3; siehe unten)
  - `backup`: lokaler Backup-Status (Slice 8.5; siehe unten)
  - `network`: begrenzte Netzwerk-Präsenz (Slice 8.7; siehe unten)
  - `software`: read-only Software-/Release-Metadaten (Slice 8.9; siehe unten)
  - `runtime`: read-only Runtime-Readiness-Label (Slice 8.10; siehe unten)
  - `health`: read-only Health-Summary (Slice 8.11; siehe unten)
  - `storage`: read-only Storage-Zusammenfassung aus Block-Inventar (Slice 12.2; siehe unten)
  - `collected_at`: UTC-Zeitstempel (RFC3339Nano)
- **Readiness in 200:** Datenbankfehler liefern **200** mit `readiness.status = "not_ready"` und `readiness.database = "error"` — kein **503**, keine Behauptung voller Appliance-Gesundheit.
- **Fehler:** **401** fehlende Session; **403** fehlende Permission; **500** interner Fehler — kanonisches JSON-Fehler-Envelope.
- Keine Secrets, Pfade, Config-Inhalte, Backup-Details, Session-Werte oder Rohfehler in der Antwort.

**Zusätzliches Feld `host` (Slice 8.3):**

```json
"host": {
  "cpu": { "status": "ok", "logical_cores": 4, "load_1m": 0.42 },
  "memory": { "status": "ok", "total_bytes": 8589934592, "available_bytes": 4294967296, "used_bytes": 4294967296 },
  "filesystems": [{
    "id": "state",
    "status": "ok",
    "total_bytes": 107374182400,
    "available_bytes": 53687091200,
    "used_bytes": 53687091200,
    "fs_type": "ext4"
  }]
}
```

- **`host.cpu`:** `logical_cores` (positive Integer), `load_1m` (1-Minuten-Load, nicht CPU-%); bei Fehler nur `"status": "unavailable"`.
- **`host.memory`:** `total_bytes`, `available_bytes`, `used_bytes` (used = total − available); bei Fehler nur `"status": "unavailable"`.
- **`host.filesystems`:** genau ein Eintrag mit `"id": "state"` (validierter `state_dir`); optional `fs_type` aus fester Vokabularliste (`ext4`, `xfs`, `btrfs`, `tmpfs`, `other`); nie Pfad, Device oder Mount-Quelle.
- Host-Metrik-Fehler degradieren **nur** die betroffene Sektion; Overview bleibt **200**. Host-Metriken behaupten keine Appliance-Gesundheit.

**Zusätzliches Feld `backup` (Slice 8.5):**

```json
"backup": { "status": "never_run", "ever_succeeded": false }
```

```json
"backup": {
  "status": "succeeded",
  "completed_at": "2026-07-14T11:30:00.000000000Z",
  "ever_succeeded": true
}
```

```json
"backup": {
  "status": "failed",
  "completed_at": "2026-07-14T11:30:00.000000000Z",
  "ever_succeeded": true,
  "failure": "restart"
}
```

```json
"backup": { "status": "unavailable" }
```

- **`status`:** `never_run` (kein Status-Sidecar), `succeeded`, `failed`, `unavailable` (nicht vertrauenswürdig lesbar).
- **`completed_at`:** nur bei `succeeded`/`failed`; UTC RFC3339Nano.
- **`ever_succeeded`:** bei `never_run` immer `false`; bei `succeeded` immer `true`; bei `failed` sticky über spätere Fehler.
- **`failure`:** nur bei `failed`; grobe Klasse (`artifact`, `restart`, `health`, `readiness`, `internal`) — keine Pfade, Artefaktnamen, Hashes oder Rohfehler.
- Status beschreibt den **letzten aufgezeichneten** Abschluss; er beweist nicht, dass ein Artefakt noch existiert oder gültig ist.
- Backup-Status-Fehler degradieren **nur** `backup`; Overview bleibt **200**. Kein Backup-Trigger, keine Artefakt-Enumeration.

**Zusätzliches Feld `network` (Slice 8.7):**

```json
"network": { "status": "available" }
```

```json
"network": { "status": "unknown" }
```

```json
"network": { "status": "unavailable" }
```

- **`network`:** genau ein Schlüssel `status`.
- **`available`:** Enumeration gelang; mindestens eine berechtigte, nicht-loopback Schnittstelle erscheint administrativ up (aus Sicht dieses API-Prozesses).
- **`unknown`:** Enumeration gelang; keine berechtigte Schnittstelle beobachtet.
- **`unavailable`:** Enumeration-, Collector-, Plattform- oder Integritätsfehler.
- Kein Zeitstempel, Count, Grund, Fehler, Identifier oder Diagnosefeld.
- Netzwerk-Sammlungsfehler degradieren **nur** `network`; Overview bleibt **200**, sofern die Kern-Assembly gelingt.
- **Keine Inferenzen:** kein Internet, DNS, Routing, DHCP, LAN-/Public-Reachability, Link-Carrier oder Netzwerkkonfiguration. Keine Schnittstellen-, MAC-, IP- oder Namens-Exposure.

**Zusätzliches Feld `software` (Slice 8.9):**

```json
"software": {
  "status": "ok",
  "version": "0.2.0-dev",
  "commit": "abc123",
  "channel": "development"
}
```

```json
"software": { "status": "unavailable" }
```

- **`software`:** read-only Metadaten aus dem laufenden API-Prozess (bereits eingebettete Version/Commit und abgeleiteter Kanal).
- **`status`:** `ok` oder `unavailable`.
- **`version`:** nur bei `ok`; eingebettete Anwendungsversion.
- **`commit`:** optional bei `ok`; Build-/Revisionskennung, wenn materialisiert (kann `unknown` sein).
- **`channel`:** nur bei `ok`; `development`, `production` oder `unknown` — abgeleitet aus dem bestehenden `api.environment`-Modell, kein Update-Kanal aus dem Netz.
- Kein Update-Check, kein Paketmanager, keine OS-/Kernel-Inventory, keine Registry-/CI-Metadaten.
- Fehlende Version → nur `status: unavailable`; Overview bleibt **200**, sofern die Kern-Assembly gelingt.

**Zusätzliches Feld `runtime` (Slice 8.10):**

```json
"runtime": { "status": "ready" }
```

```json
"runtime": { "status": "degraded", "note": "database" }
```

```json
"runtime": { "status": "unknown" }
```

- **`runtime`:** grobes Runtime-Readiness-Label, **abgeleitet** aus bestehendem `readiness` und `service` — keine neuen Probes.
- **`status`:** `ready`, `degraded` oder `unknown`.
- **`note`:** optional nur bei `degraded`; derzeit nur `database` (wenn `readiness` Datenbankfehler meldet).
- Kein Zeitstempel in `runtime`; nutze `collected_at` auf Overview-Ebene.
- Keine Host-, Netzwerk-, Paket- oder Update-Claims; kein Ersatz für `/readyz`-Betrieb.

**Zusätzliches Feld `health` (Slice 8.11):**

```json
"health": { "status": "healthy" }
```

```json
"health": { "status": "warning", "note": "database" }
```

```json
"health": { "status": "warning", "note": "backup" }
```

```json
"health": { "status": "unknown" }
```

- **`health`:** grobe Health-Summary, **abgeleitet** aus bestehendem `runtime` und `backup` — keine neuen Probes.
- **`status`:** `healthy`, `warning` oder `unknown`.
- **`note`:** optional nur bei `warning`; `database` (Runtime degradiert) oder `backup` (letzter Backup-Versuch `failed` bei sonst ready Runtime).
- Kein Zeitstempel in `health`; nutze `collected_at` auf Overview-Ebene.
- Keine Voll-Appliance-Wellness-, Host- oder Netzwerk-Claims.

### PATCH `/api/v1/settings/instance`

- **Request:** `Content-Type: application/json`, Body `{ "display_name": "<name>" }` (max. 4 KiB).
- **Validierung:** Pflichtfeld `display_name`; Unicode-Trim; max. 80 Zeichen; keine Steuerzeichen; keine unbekannten JSON-Felder.
- **Erfolg:** **200** `{ "display_name": "<persistierter Name>" }`; kein Cookie-Change.
- **No-Op:** Identischer normalisierter Name → **200** ohne DB-Update, Timestamp-Update oder Audit.
- **Fehler:** **400** ungültige Anfrage; **401** fehlende Session; **403** fehlende Berechtigung oder CSRF; **405** andere Methoden; **500** Persistenzfehler.
- Persistiert in `system.hostname` (kanonischer Speicherort für `instance.name` in GET `/api/v1/settings`).

### GET `/api/v1/storage/disks`

- **Auth:** Session erforderlich; Permission `storage:read` (Owner, Operator, Read-only).
- **CSRF:** nicht erforderlich (read-only GET).
- **Cache:** `Cache-Control: no-store`.
- **Erfolg:** **200** mit read-only Inventar:
  - `collected_at`: UTC-Zeitstempel (RFC3339Nano)
  - `status`: `ok` (Discovery gelang) oder `unavailable` (Plattform-/Collector-Fehler)
  - `disks`: Array klassifizierter Blockgeräte (bei `unavailable` typischerweise `null`)
- **Disk-Eintrag:**
  - `id`: stabile opaque ID (`disk-<sha256-prefix>`) — **kein** `/dev/sdX`, kein Kernel-Name
  - `status`: `ok` oder `unavailable` (pro Gerät bei Integritätsfehler)
  - `size_bytes`, `rotational`, `removable`: optional, nur wenn bekannt
  - `eligibility`: `eligible`, `excluded` oder `unknown`
  - `reasons`: Ausschlussgründe (`root_disk`, `state_filesystem`, `removable`, `read_only`, `mounted_in_use`, `unsupported_filesystem`, `install_media`, `virtual_device`, `ambiguous_identity`)
- **Fehler:** **401** fehlende Session; **403** fehlende Permission; **500** interner Fehler.
- **Keine Mutation:** kein Format, kein Wipe, kein Auto-Select; Discovery-Fehler liefern `status: unavailable` ohne Rohfehler oder Pfade.
- Details und Nicht-Ziele: `docs/05_STORAGE.md`.

```json
{
  "collected_at": "2026-07-15T12:00:00.000000000Z",
  "status": "ok",
  "disks": [
    {
      "id": "disk-a1b2c3d4e5f6",
      "status": "ok",
      "size_bytes": 2000000000000,
      "rotational": true,
      "removable": false,
      "eligibility": "excluded",
      "reasons": ["root_disk"]
    },
    {
      "id": "disk-f6e5d4c3b2a1",
      "status": "ok",
      "size_bytes": 4000000000000,
      "rotational": false,
      "removable": false,
      "eligibility": "eligible"
    }
  ]
}
```

### GET `/api/v1/storage/pools`

- **Auth:** Session; Permission `storage:read`.
- **Erfolg:** **200** mit Layout:
  - `pools`: declared pools from appliance state (`status: declared`, `disk_format_state: pending`)
  - `pool_management`: `"declared_pools"`
  - `mutation_available`: `true`
  - `disk_format_applied`: `false` (honest — disks are not formatted yet)
- Details: `docs/05_STORAGE.md` (Slice 12.3).

### POST `/api/v1/storage/pools`

- **Auth:** Session; Permission `storage:write`; `X-CSRF-Token`.
- **Body:** `{ "name": "tank", "disk_ids": ["disk-..."], "confirm": true }`
- **Erfolg:** **201** declared pool. Does **not** wipe or format disks.
- **Fehler:** `CONFIRM_REQUIRED`, `DISK_NOT_ELIGIBLE`, `DISK_IN_USE`, `INVALID_NAME`, `NO_DISKS`.

### POST `/api/v1/storage/pools/{poolID}/datasets`

- Prepare a dataset plan under a declared pool (`path_intent` only; no filesystem creation).

### GET `/api/v1/storage/shares`

- **Auth:** Session; Permission `storage:read`.
- **Erfolg:** **200** mit planned shares:
  - `share_management`: `"planned_shares"`
  - `protocol_support`: `"not_available"`
  - `mutation_available`: `true`

### POST `/api/v1/storage/shares`

- Prepare a share plan. Protocol is stored as `planned` until SMB/NFS ships.

**Overview-Feld `storage` (Slice 12.3):**

```json
"storage": {
  "status": "ok",
  "disk_count": 2,
  "eligible_count": 1,
  "excluded_count": 1,
  "unknown_count": 0,
  "pool_count": 0,
  "share_count": 0,
  "mutation_available": true
}
```

### GET `/api/v1/public/install-media`

- **Auth:** nicht erforderlich.
- **Cache:** `Cache-Control: no-store`.
- **Erfolg:** **200** mit Install-Media-Metadaten:
  - `publication_status`: `not_built`, `local_staging`, oder `unavailable`
  - `primary_artifact`: Name, Format, Firmware-Modus (`bios+uefi`), `bios_support` / `uefi_support` / `dual_mode`, optional Größe/SHA-256, `download_available`
  - `build_target`, `stage_target`, `limitations`
- **Download:** Binärdateien werden **nicht** über diesen Endpoint geliefert. Bei gesetztem `VYNTRIO_RELEASE_STAGING_DIR` liefert `/release/vyntrio-install-media.img` die gestagte Datei.
- Details: `docs/24_INSTALL_MEDIA.md`.

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
