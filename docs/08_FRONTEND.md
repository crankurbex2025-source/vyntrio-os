# Frontend

## Architektur
React + TypeScript mit Feature-Oriented Structure.

**Block 11 / Slice 11R.1–11R.2 (implementiert):** `vyntrio.tokens.css`, i18n
(`de` default), Preview `/design-preview/landing`; 11R.2 extrahiert Public-
Primitives unter `surfaces/public/components/`. Route `/` unverändert (11.1).
Details: `docs/23_WEBSITE_DESIGN_DIRECTION.md`.

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

## Strukturen (aktuell, Slice 11.1 / 11R.2)
- `app/` — Router und Layouts
- `surfaces/public/components/` — wiederverwendbare Public-Primitives (11R.2)
- `surfaces/public/preview/` — Design-Preview-Routen
- `surfaces/auth/` — Login-Einstieg
- `surfaces/appliance/` — Session-Probe, Overview, Settings
- `features/` — bestehende Feature-Module (auth, overview, settings)
- `shared/` — Theme-Tokens und minimale UI-Primitives
- `lib/api/` — API-Client (nur Appliance-Surface)

Langfristig geplant (noch nicht vollständig): `pages/`, `widgets/`, `entities/`.

## UI-Funktionen
- Dashboard Widgets
- Tabellen und Detailansichten
- Wizards für Storage, App-Deployment und VM-Erstellung
- Realtime-Log-Viewer
- Lizenz- und Update-Center
