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
Control Plane läuft lokal auf dem Host. Systemnahe Services werden über systemd verwaltet. API und Worker laufen als getrennte Go-Binaries oder Prozesse, das Frontend wird als statisches Web-Bundle ausgeliefert.

## Architekturregeln
- Keine Domain-Abhängigkeit auf Infrastrukturpakete.
- Jede externe Integration bekommt ein Interface und mindestens einen Adapter.
- Long-Running Operations laufen als Jobs mit Statusmodell.
- Breaking Changes benötigen ADR und Migrationskonzept.
