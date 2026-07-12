# API

## Stil
REST für Konfiguration und Ressourcen, WebSocket für Live-Events. JSON-only. Versionierung über /api/v1.

## Resource Families
- /auth
- /users
- /roles
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
