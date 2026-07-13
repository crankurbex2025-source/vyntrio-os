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

**Implementiert (v1):** Das produktive Frontend wird als statisches Vite-Bundle gebaut, per `go:embed` in das API-Binary eingebettet und same-origin vom selben Binary ausgeliefert (`/assets/*` plus `index.html`-SPA-Fallback für erlaubte GET/HEAD-Pfade). API-, Health- und Readiness-Routen behalten Priorität. Details: `docs/09_API.md`, `docs/17_SECURITY.md`, `docs/19_RELEASE.md`.

**Beschlossener Block-7-Vertrag (noch nicht implementiert):** Vyntrio v1 ist ein Single-Node-Linux-Appliance-Service unter systemd mit dediziertem `vyntrio`-Servicekonto, Host-Admin-Laufzeitkonfiguration unter `/etc/vyntrio/` (TOML, Fail-Closed, Neustart statt Live-Reload), persistentem State unter `/var/lib/vyntrio/` (inkl. SQLite und Sidecar-Dateien) und strikter Trennung von datenbankgestützten Produkt-Settings (Owner/CSRF/Audit über die API). Backup/Restore sind zukünftige lokale Admin-CLI-Operationen. Autoritativ: `docs/ADR/0005-appliance-runtime-operations.md`. Docker/OCI, ISO, Kubernetes, Clustering und Multi-Node bleiben ausdrücklich zurückgestellt.

## Architekturregeln
- Keine Domain-Abhängigkeit auf Infrastrukturpakete.
- Jede externe Integration bekommt ein Interface und mindestens einen Adapter.
- Long-Running Operations laufen als Jobs mit Statusmodell.
- Breaking Changes benötigen ADR und Migrationskonzept.
