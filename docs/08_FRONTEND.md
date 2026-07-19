# Frontend

## Architektur
React + TypeScript mit Feature-Oriented Structure.

**Block 11 / 11R (implementiert):** Öffentliche v2-Oberflächen auf `/`, `/download`,
`/docs` (DE/EN, lazy-loaded, GSAP motion, ehrliche statische Inhalte). Preview-
Spiegel unter `/design-preview/*`; Slice-11.1-Rollback unter `/design-preview/*-legacy`.
Cutover-Slices 11R.10–11R.12 abgeschlossen. Details: `docs/23_WEBSITE_DESIGN_DIRECTION.md`.

**11R.7 (implementiert):** Web-App-Manifest (`site.webmanifest`), Install-Icons (192/512), Theme-Meta — Shortcut-Install für die Website, **kein** Offline-Stack und **kein** Hardware-/OS-Installer.

**11R.6 (implementiert):** Appliance/Login nutzen dieselben `--vyn-*`-Tokens wie die öffentlichen v2-Oberflächen (`vyntrio.tokens.css`); Legacy `--dashboard-*`-Aliase entfernt.

## Produktions-Auslieferung (implementiert)

Der produktive Vite-Build (`frontend/dist`) wird über `make ui-stage` nach
`internal/interfaces/http/ui/dist/` gestaget (generiert, nicht committet) und
per `go:embed` in das API-Binary eingebettet. Die UI läuft same-origin gegen
die reale API — live auf `vyntrio.xyz` und auf jeder Appliance mit `vyntrio-api`.
Kein Vite-Dev-Proxy und keine CORS-Lockerung in Produktion.
Der CSRF-Token lebt ausschließlich im Frontend-Speicher (kein Browser-Storage).
Details: `docs/09_API.md`, `docs/17_SECURITY.md`, `docs/19_RELEASE.md`.

## Kernprinzipien
- UI-Komponenten strikt von Domain-Logik trennen.
- API-Zugriffe zentral kapseln.
- Live-Daten als Streams behandeln (Appliance; Public bleibt statisch).
- Accessibility und Fehlerzustände als Pflichtbestandteil.

## Strukturen (aktuell)

- `app/` — Router und Layouts
- `surfaces/public/` — Shipped public pages (`landing/`, `download/`, `docs/`) + shared views
- `surfaces/public/components/` — wiederverwendbare Public-Primitives
- `surfaces/public/preview/` — Design-Preview-Spiegel und Motion
- `surfaces/auth/` — Login-Einstieg
- `surfaces/appliance/` — Session-Probe, Overview, Settings
- `features/` — Feature-Module (auth, overview, settings)
- `shared/` — Theme-Tokens (`vyntrio.tokens.css`) und UI-Primitives
- `lib/api/` — API-Client (nur Appliance-Surface)

Langfristig geplant (noch nicht vollständig): `pages/`, `widgets/`, `entities/`.

## UI-Funktionen

**Implementiert:**
- Public landing, release, and docs surfaces (Block 11R)
- Appliance overview (read-only, Block 8) and Owner settings
- Login / logout with CSRF

**Geplant (nicht implementiert):**
- Dashboard widgets beyond overview panels
- Storage, container, and VM wizards
- Realtime log viewer and WebSocket streams
- License and update center UI
- Full appliance visual refresh beyond token parity (Block 11.6)
