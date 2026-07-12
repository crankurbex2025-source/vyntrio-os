# frontend

React + TypeScript web dashboard for Vyntrio OS administration.

## Planned stack

See `docs/04_TECH_STACK.md` and `docs/08_FRONTEND.md`:

- React, TypeScript, TailwindCSS
- TanStack Query, Zustand (or Redux Toolkit)
- Vite (toolchain selection in Phase 2)

## Status

Foundation only — no application UI in Phase 1. Toolchain initialization is Phase 2.

## Local setup (after Phase 2)

```bash
cd frontend
cp .env.example .env.local   # when applicable
npm install
npm run dev
```

## Rules

- No secrets in frontend bundles.
- API calls go through typed client modules (Phase 2+).
- No product dashboards until Phase 0.4 (Dashboard).
