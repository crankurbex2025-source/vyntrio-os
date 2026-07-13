# frontend

React + TypeScript web dashboard for Vyntrio OS administration.

## Toolchain foundation

This slice initializes a minimal React + TypeScript + Vite frontend foundation.
It currently renders a static page only and does not call backend APIs.

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
- No browser storage/cookie/session-token handling in this foundation slice.
- No product dashboard modules in this slice.
