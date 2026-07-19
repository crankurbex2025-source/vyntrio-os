# Masterplan

## Gesamtstrategie
Die Entwicklung von Vyntrio OS erfolgt in klar abgegrenzten Produkt- und Plattformphasen. Jede Phase liefert ein testbares, dokumentiertes und releasefähiges Zwischenziel.

## Phasen
1. Foundation: Repository, Build-System, CI/CD, Dokumentations-Backbone.
2. Linux-Basis: Debian-Image, Installer, Bootkette, Hardware-Erkennung.
3. Platform Core: API, Auth, Settings, Jobs, Events, Telemetrie-Basis.
4. Dashboard: UX-Framework, Navigation, Widgets, Live-Status.
5. Storage: Disks, Pools, Shares, Snapshots, Health.
6. Container: Docker-Engine, Compose, Templates, Logs, Backups.
7. Virtualisierung: KVM/QEMU/libvirt, VM-Lifecycle, Media, Passthrough.
8. Hub & Plugins: Marketplace, Signaturen, SDK, Rechte.
9. Commercial Core: Lizenzierung, Aktivierung, Geräte-/Server-Tiers, Update-Server.
10. Beta & GA: Hardening, QA, Migrations, Release Engineering.

## Lieferprinzip
Jede Phase endet mit: aktualisierter Doku, Tests, Architektur-Review, Sicherheits-Review, Demo-Szenario, Risiko-Update und Backlog-Repriorisierung.

## Governance
- RFC-Pflicht für Architekturänderungen.
- ADR-Pflicht für irreversible Technologieentscheidungen.
- Security Review vor jedem externen Release.
- Definition of Done umfasst Code, Tests, Doku, Monitoring und Upgrade-Pfad.
