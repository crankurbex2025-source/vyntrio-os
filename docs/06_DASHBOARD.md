# Dashboard

## UX-Ziel
Das Dashboard ist der zentrale Einstiegspunkt für Systemzustand, Administration und operative Aktionen.

## Bereiche
- Overview
- Storage
- Containers
- Virtual Machines
- Network
- Users & Roles
- Updates
- License
- Logs & Notifications

## Anforderungen
- Responsive Layout mit Desktop-First für Admin-Nutzung und funktionaler Mobile-Version.
- Live-Kacheln für CPU, RAM, Temperatur, Speicher, Netzwerk und Servicezustände.
- Globales Such- und Command-Palette-Konzept.
- Kontextbezogene Actions, keine versteckten kritischen Aktionen.
- Dark Mode und langfristig Light Mode.

## Technik
- Modulbasierte React-Routen.
- WebSocket Live-Streams für Metriken und Logs.
- Fehler-, Lade- und Empty States für jedes Modul.
