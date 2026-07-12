# Testing

## Testpyramide
- Unit Tests für Domain und Application
- Integration Tests für Adapter und Datenbank
- Contract Tests für API
- E2E Tests für Dashboard und Kernworkflows
- Installer- und Upgrade-Tests für OS-Lifecycle

## Pflicht-Szenarien
- Frische Installation
- Upgrade von Vorversion
- Disk-Ausfall und Health-Alarm
- Container Deployment
- VM-Erstellung
- Lizenzaktivierung online/offline
- Rollback nach fehlgeschlagenem Update

## Qualitätsgates
Kein Release ohne grüne CI, definierte Mindestabdeckung, Security Checks und dokumentierte manuelle Prüfungen.
