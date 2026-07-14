# frontend

React + TypeScript web dashboard for Vyntrio OS administration.

## Toolchain foundation

This slice provides the authenticated Control-Center frontend. On boot it probes
`GET /api/v1/overview` for session authorization, renders the read-only overview
for roles with `system:health`, and loads Owner-only instance settings through a
separate `GET /api/v1/settings` flow.

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

## Rules

- No secrets in frontend bundles.
- CSRF tokens stay in memory only for mutating requests; overview GET sends no CSRF header.
- No product dashboard modules beyond Slice 8.1 overview and existing Owner settings.
