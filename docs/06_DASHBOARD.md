# Dashboard

## UX-Ziel
Das Dashboard ist der zentrale Einstiegspunkt für Systemzustand, Administration und operative Aktionen.

## Implementiert (Block 8, Slice 8.1)

Die Control-Center-Overview ist die autorisierte Standard-Landing-View für alle
Session-Rollen mit `system:health` (Owner, Operator, Read-only).

**Enthalten in 8.1:**
- Session-Bootstrap über `GET /api/v1/overview` (nicht mehr über Settings)
- Read-only Overview-Shell mit Instance-Identität (`name`, `version`, `commit`)
- API-Umgebung (`environment`)
- In-Process-Service-Status (`running` während der Handler läuft)
- Readiness-Karten (`ready`/`not_ready`, Datenbank `ok`/`error`) — truthful in HTTP 200
- `collected_at`-Zeitstempel
- Owner-only Einstieg zu Instance Settings über separaten `GET /api/v1/settings`-Flow
- UI-Zustände: loading, 401 login, 403 forbidden, 5xx unavailable, ready, not_ready
- Dark, premium Appliance-Dashboard-Stil (responsive)

**Explizit zukünftig (nicht in 8.1):**
- CPU, RAM, Temperatur, Disk, Netzwerk-Metriken
- Backup-Status und Restore-Aktionen
- Warnungen, Alerts, Management-Aktionen
- WebSocket-Live-Streams und Polling
- Modul-Navigation / Router-Framework
- Storage-, Container-, VM-, User-Management-UI

## Bereiche (Produkt-Roadmap)
- Overview
- Storage
- Containers
- Virtual Machines
- Network
- Users & Roles
- Updates
- License
- Logs & Notifications

## Anforderungen (Langfrist)
- Responsive Layout mit Desktop-First für Admin-Nutzung und funktionaler Mobile-Version.
- Live-Kacheln für CPU, RAM, Temperatur, Speicher, Netzwerk und Servicezustände.
- Globales Such- und Command-Palette-Konzept.
- Kontextbezogene Actions, keine versteckten kritischen Aktionen.
- Dark Mode und langfristig Light Mode.

## Technik (Langfrist)
- Modulbasierte React-Routen.
- WebSocket Live-Streams für Metriken und Logs.
- Fehler-, Lade- und Empty States für jedes Modul.
