# Frontend

## Architektur
React + TypeScript mit Feature-Oriented Structure.

## Produktions-Auslieferung (v1, implementiert)

Der produktive Vite-Build (`frontend/dist`) wird über `make ui-stage` nach
`internal/interfaces/http/ui/dist/` gestaget (generiert, nicht committet) und
per `go:embed` in das API-Binary eingebettet. Die UI läuft same-origin gegen
die reale API — ohne Vite-Dev-Proxy und ohne CORS-Lockerung in Produktion.
Der CSRF-Token lebt ausschließlich im Frontend-Speicher (kein Browser-Storage).
Details: `docs/09_API.md`, `docs/17_SECURITY.md`, `docs/19_RELEASE.md`.

## Kernprinzipien
- UI-Komponenten strikt von Domain-Logik trennen.
- API-Zugriffe zentral kapseln.
- Live-Daten als Streams behandeln.
- Accessibility und Fehlerzustände als Pflichtbestandteil.

## Strukturen
- app/
- pages/
- widgets/
- features/
- entities/
- shared/

## UI-Funktionen
- Dashboard Widgets
- Tabellen und Detailansichten
- Wizards für Storage, App-Deployment und VM-Erstellung
- Realtime-Log-Viewer
- Lizenz- und Update-Center
