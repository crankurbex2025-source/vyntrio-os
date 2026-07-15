# frontend

React + TypeScript web UI for Vyntrio OS — public product surface and appliance control center.

## Surfaces (Block 11)

| Route | Surface | Description |
|-------|---------|-------------|
| `/` | Public | Static landing page (no API) |
| `/download` | Public | Honest placeholder — install media not yet linked |
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

Open `http://localhost:5173/` for the shipped v2 landing (Block 11R.10). Slice 11.1 rollback: `/design-preview/landing-legacy`.
Design preview (Block 11R.1): `http://localhost:5173/design-preview/landing`
Appliance flow: `/app` or `/login`.

## Rules

- Public surfaces must not import `lib/api` or show fake operational data.
- CSRF tokens stay in memory only for mutating requests; overview GET sends no CSRF header.
- Appliance UI shows only data from real API endpoints (Block 8 honesty contract).
