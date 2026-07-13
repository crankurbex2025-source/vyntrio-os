# Security

## Pflichtanforderungen
- HTTPS everywhere
- TLS-Härtung
- Firewall standardmäßig aktiv
- sichere Passwörter
- Audit Logs
- signierte Updates
- sichere API

## Sicherheitsprogramm
- Threat Modeling je Kernmodul
- Secret Handling Policy
- Dependency Scanning
- SAST/DAST in CI
- Security Incident Runbooks
- Responsible Disclosure

## Besondere Risiken
- Remote Admin Surface
- Container Escape
- Unsichere Plugin-Erweiterungen
- Manipulierte Update-Pipelines
- Lizenz-Bypass und Token-Leaks

## Identity & Sessions (v1)

Browser-Auth nutzt serverseitige Sessions (`vyntrio_session`, HttpOnly) und CSRF via `X-CSRF-Token` (geliefert nur im Login-JSON). Details: `docs/ADR/0004-identity-and-access.md`, `docs/11_AUTH.md`.

## UI-HTML-Security-Header und Cache-Policy (v1, implementiert)

Die eingebettete UI (`index.html`-Entry und erlaubter SPA-Fallback) wird mit
genau diesen Headern ausgeliefert:

```text
Content-Type: text/html; charset=utf-8
Cache-Control: no-cache
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
```

Diese HTML-Policy gilt **nur** für UI-HTML-Antworten; sie wird nicht auf
JSON-API-Antworten übertragen.

Cache-Trennung:

- Erfolgreiche content-gehashte `/assets/*`-Dateien:
  `Cache-Control: public, max-age=31536000, immutable` plus
  `X-Content-Type-Options: nosniff`.
- `index.html` / SPA-Fallback-HTML: `Cache-Control: no-cache` — HTML wird bei
  jedem Laden revalidiert, weil es die aktuellen Asset-Hashes referenziert;
  die gehashten Assets selbst ändern nie ihren Inhalt unter demselben Namen
  und dürfen deshalb unveränderlich gecacht werden.
- Fehlende oder ungültige Asset-Pfade (inkl. Traversal-Varianten): Nicht-HTML-
  Fehler ohne `immutable`, niemals SPA-Fallback.
- API-, Health-, Readiness- und JSON-Fehlerantworten behalten ihr bestehendes
  Route-/Handler-Verhalten; sie werden nicht in HTML transformiert und
  erhalten keine UI-Cache-Policy.

Die Same-Origin-CSP ergänzt die bestehenden Kontrollen (HttpOnly-Session-
Cookie, CSRF-Header-Pflicht für mutierende Requests, serverseitige
Autorisierung/RBAC, Input-Validierung), ersetzt sie aber nicht: Autorisierung
und CSRF-Prüfung bleiben ausschließlich serverseitig durchgesetzt.

## Appliance-Betriebsmodell (Block 7)

**Implementiert (Slice 7.2):** Laufzeitkonfiguration aus `/etc/vyntrio/config.toml`
(TOML, Fail-Closed, kein Live-Reload). Persistenter State unter
`/var/lib/vyntrio/`; startup-time Validierung von `state_dir` (inkl.
Symlink-Ablehnung) und bekannter SQLite-Dateinamen (`vyntrio.db`,
`vyntrio.db-journal`, `vyntrio.db-wal`, `vyntrio.db-shm`). Keine Garantie für
race-freie Dateisystem-Eindämmung gegen lokale Schreiber im State-Verzeichnis.
Legacy-`VYNTRIO_*`-Umgebungsvariablen beeinflussen den API-Server nicht mehr.

**Implementiert (Slice 7.3):** statisches `vyntrio`-Servicekonto
(`distro/systemd/vyntrio.sysusers`), systemd-Unit `vyntrio-api.service`
(non-root, `StateDirectory=vyntrio`, konservative Sandbox), tmpfiles-Layout für
`/etc/vyntrio` (`0750 root:vyntrio`). Konfigurationsdatei
`/etc/vyntrio/config.toml` bleibt root-administriert (`0640 root:vyntrio`
empfohlen). Systemd/Ownership operationalisiert die Trusted-Admin-Annahme;
pathname-basiertes SQLite bleibt nicht race-frei.

**Beschlossen (Slice 7.8), noch nicht implementiert:** lokale
Backup/Restore-Architektur für Root-Operatoren — Service-Stop vor
SQLite-Dateikopie, Manifest mit SHA-256, Ziel unter
`/var/lib/vyntrio/backups/` (root-only), Offline-Restore mit
Pre-Restore-Preserve und Readiness-Nachweis über Loopback. Kein API/UI-Zugriff,
keine Verschlüsselung/Retention/Cloud in v1. Konfigurationskopie als separater
Manifest-Member; zukünftige Geheimnisse in `config.toml` erfordern expliziten
Operator-Schutz. Autoritativ: `docs/ADR/0005-appliance-runtime-operations.md`
(Abschnitte G, H).
