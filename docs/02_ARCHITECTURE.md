# Architektur

## Architekturziel
Vyntrio OS verwendet eine modulare Clean Architecture mit klarer Trennung zwischen Domain, Application, Infrastructure und Interfaces. Das System wird als lokales Appliance-OS mit integriertem Control Plane konzipiert.

## Schichten
### Domain
Enthält die Geschäftslogik für Benutzer, Storage, Container, VMs, Netzwerk, Updates, Lizenzen und Plugins. Keine Framework-Abhängigkeiten.

### Application
Orchestriert Use Cases, Policies, Jobs, Validierung, Event-Publishing und Berechtigungsprüfungen.

### Infrastructure
Adapter für SQLite/PostgreSQL, Docker, libvirt, systemd, nftables, SMART, Update-Signing, Dateisystem-Tools und Lizenzsignaturen.

### Interfaces
REST-API, WebSocket-Streams, CLI-Service-Wrappers, Installer-UI und React-Frontend.

## Kernmodule
- Identity & Access
- System Inventory
- Storage Service
- Network Service
- Container Service
- VM Service
- Backup Service
- Notification Service
- License Service
- Update Service
- Marketplace/Plugin Service

## Kommunikationsmodell
- Synchronous: REST für CRUD, Konfiguration und Admin-Aktionen.
- Real-time: WebSocket für Status, Logs, Progress, Alerts.
- Internal: Event Bus für Domain Events und Long-Running Jobs.

## Persistenz
Start mit SQLite für Appliance-Einfachheit; PostgreSQL als spätere Option für größere Installationen und Enterprise-Szenarien.

## Sicherheitsarchitektur
- TLS-gesicherte Oberfläche.
- Argon2 für Passwort-Hashes.
- Signierte Updates und Plugin-Pakete.
- RBAC mit auditierbaren Admin-Aktionen.
- Secrets niemals im Frontend.

## Deployment-Modell
Control Plane läuft lokal auf dem Host. Systemnahe Services werden über systemd verwaltet. API und Worker laufen als getrennte Go-Binaries oder Prozesse.

**Implementiert (v1):** Das produktive Frontend wird als statisches Vite-Bundle gebaut, per `go:embed` in das API-Binary eingebettet und same-origin vom selben Binary ausgeliefert (`/assets/*` plus `index.html`-SPA-Fallback für erlaubte GET/HEAD-Pfade). API-, Health- und Readiness-Routen behalten Priorität. Der API-Server lädt Laufzeitkonfiguration aus `/etc/vyntrio/config.toml` (TOML, Fail-Closed, Neustart statt Live-Reload); persistenter State unter `/var/lib/vyntrio/` mit startup-time Pfad-/Symlink-Validierung (ohne race-freie SQLite-Eindämmung). **Slice 7.3:** systemd-Unit, statisches `vyntrio`-Konto und Ownership-Modell (`distro/systemd/`). Details: `docs/09_API.md`, `docs/17_SECURITY.md`, `docs/19_RELEASE.md`, `docs/ADR/0005-appliance-runtime-operations.md`.

**Beschlossener Block-7-Vertrag (teilweise implementiert):** Backup-Befehl
`vyntrio-backup` (Slice 7.9); Restore-Sicherheitsvertrag
(`docs/ops/restore-safety-contract.md`, Slice 7.11, noch ohne Restore-CLI);
Packaging bleibt zukünftige Slice. Docker/OCI, ISO, Kubernetes, Clustering und
Multi-Node bleiben ausdrücklich zurückgestellt.

**Block 8 / Slice 8.1 (implementiert):** Authentifizierte read-only Appliance-Overview
(`GET /api/v1/overview`, Permission `system:health`) als Control-Center-Landing;
Application-Layer-DTO in `internal/application/overview`; React-Overview-Shell mit
Session-Bootstrap über Overview statt Owner-only Settings. Owner-Settings bleiben
separat unter `settings:admin:read`/`write`. Keine Host-Metriken, Polling oder
WebSockets in 8.1.

**Block 8 / Slice 8.3 (implementiert):** Read-only Host-Metriken im Overview-DTO
(`host.cpu`, `host.memory`, `host.filesystems[id=state]`) via
`internal/platform/hostmetrics` — direkte in-process `/proc`- und statfs-Sammlung
nur auf dem validierten `state_dir`; per-Metrik `unavailable`-Degradation in HTTP 200.
Kein CPU-%, Netzwerk, Backup-Artefakt-Exposure, Mount-Inventory oder Privilege-Expansion.

**Block 8 / Slice 8.5 (implementiert):** Read-only lokaler Backup-Status im Overview-DTO
(`backup.status`, optional `completed_at`, `ever_succeeded`, `failure`) via
`internal/platform/backupstatus` — API liest nur ein root-geschriebenes Status-Sidecar
unter dem validierten `state_dir`; `vyntrio-backup` ist alleiniger Schreiber.
Per-Status `unavailable`-Degradation in HTTP 200. Kein Backup-Trigger, keine Artefakt-Enumeration.

**Block 8 / Slice 8.7 (implementiert):** Read-only Netzwerk-Präsenz im Overview-DTO
(`network.status` ∈ `available` | `unknown` | `unavailable`) via
`internal/platform/netpresence` — in-process `net.Interfaces()` (Linux) nur Flags und
HardwareAddr-Präsenz; keine Schnittstellen-/MAC-/IP-Exposure. Per-Sektion
`unavailable`-Degradation in HTTP 200. Kein Internet-, DNS- oder Reachability-Nachweis.

**Block 8 / Slice 8.9 (implementiert):** Read-only Software-/Release-Status im Overview-DTO
(`software.status`, optional `version`, `commit`, `channel`) — materialisiert nur aus
bereits lokaler Loader-Metadaten (eingebettete Version/Commit, Kanal aus `api.environment`);
kein Update-Check, keine Paket- oder Host-Inventory.

**Block 8 / Slice 8.10 (implementiert):** Read-only Runtime-Readiness im Overview-DTO
(`runtime.status` ∈ `ready` | `degraded` | `unknown`, optional `note: database`) —
abgeleitet aus bestehendem `readiness` und `service` ohne neue Probes; UI nutzt
`collected_at` für Zeitbezug.

**Block 8 / Slice 8.11 (implementiert):** Read-only Health-Summary im Overview-DTO
(`health.status` ∈ `healthy` | `warning` | `unknown`, optional `note: database|backup`) —
abgeleitet aus bestehendem `runtime` und `backup` ohne neue Probes.

**Block 9 / Slice 9.2 (Vertrag, dokumentiert):** Install- und Recovery-Medien
(ADR-0006) — getrennte Boot-USB/ISO-Rollen für Greenfield-Installation vs.
offline Recovery-Umgebung; persistenter State nur auf Zielplatte
(`/var/lib/vyntrio/`, `/etc/vyntrio/` per ADR-0005); kein Owner-Seeding auf
generischem Medium; License-USB (`docs/15_LICENSE.md`) ist separat. Noch keine
ISO-/Installer-Implementierung.

## Architekturregeln
- Keine Domain-Abhängigkeit auf Infrastrukturpakete.
- Jede externe Integration bekommt ein Interface und mindestens einen Adapter.
- Long-Running Operations laufen als Jobs mit Statusmodell.
- Breaking Changes benötigen ADR und Migrationskonzept.
