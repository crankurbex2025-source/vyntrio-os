# Dashboard

## UX-Ziel
Das Dashboard ist der zentrale Einstiegspunkt für Systemzustand, Administration und operative Aktionen.

## Implementiert (Block 8)

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

**Enthalten in 8.3:**
- Read-only Host-Metrik-Karten: CPU-Kerne + 1-Minuten-Load, Speicher (total/verfügbar/genutzt),
  State-Filesystem-Kapazität (`id=state` only)
- Explizite `Unavailable`-Darstellung pro Metrik; kein Polling, kein Refresh-Control, keine Charts
- Load average ist nicht CPU-Auslastung; keine Prozentanzeige

**Enthalten in 8.5:**
- Read-only lokaler Backup-Status: `never_run`, `succeeded`, `failed`, `unavailable`
- Keine Backup-/Restore-Aktionen, keine Artefaktdetails, kein Verlauf

**Enthalten in 8.7:**
- Read-only Network-Panel mit `available`, `unknown`, `unavailable`
- `available`: „Local network interface present“ — mindestens eine berechtigte nicht-loopback Schnittstelle erscheint up (Prozess-Sicht); kein Internet-/DNS-/Reachability-Nachweis
- `unknown`: „Local network presence unclear“ — keine berechtigte Schnittstelle beobachtet; kein Beweis für fehlende Hardware
- `unavailable`: „Network presence could not be determined.“
- Keine Aktionen, kein Polling, keine Schnittstellen-/IP-/MAC-Anzeige, keine technischen Rohdetails

**Explizit zukünftig (nicht in 8.1/8.3/8.5/8.7):**
- CPU-Auslastung in Prozent, 5m/15m-Load, Temperatur
- Netzwerk-Metriken (Durchsatz, Pakete, Carrier, Wi-Fi, IP-Adressen)
- Zusätzliche Filesystem-/Mount-IDs, Backup-Aktionen und Restore-Aktionen
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
