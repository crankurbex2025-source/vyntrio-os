# frontend

React + TypeScript web UI for Vyntrio OS — public product surface and appliance control center.

## Surfaces (Block 11)

| Route | Surface | Description |
|-------|---------|-------------|
| `/` | Public | v2 landing (DE/EN, lazy-loaded) |
| `/download` | Public | v2 release surface — honest artifact status, no binaries linked |
| `/docs` | Public | v2 documentation surface — static topics, no live search |
| `/login` | Auth | Sign in to your appliance |
| `/app` | Appliance | Session probe, read-only overview, Owner settings |

See `docs/22_WEBSITE.md` for the Block 11 track contract.

## Local commands

```bash
cd frontend
npm ci
npm run dev
npm run typecheck
npm run lint
npm run test:run
npm run build
```

Open `http://localhost:5173/` for the shipped v2 landing (Block 11R.10). Production download/docs: `/download`, `/docs`. Slice 11.1 rollback: `/design-preview/landing-legacy`, `/design-preview/download-legacy`.
Design preview (Block 11R.1): `http://localhost:5173/design-preview/landing`
Appliance flow: `/app` or `/login`.

## Web installability (Block 11R.7)

- Manifest: `/site.webmanifest` — `start_url` `/` (production, not preview routes)
- Icons: `/icons/icon-192.png`, `/icons/icon-512.png`
- **Honest boundary:** browser install adds a website shortcut only — not hardware/OS installation and not offline support (no service worker)

## Rules

- Public surfaces must not import `lib/api` or show fake operational data.
- CSRF tokens stay in memory only for mutating requests; overview GET sends no CSRF header.
- Appliance UI shows only data from real API endpoints (Block 8 honesty contract).
