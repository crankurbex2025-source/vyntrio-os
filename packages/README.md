# packages

Shared libraries used by the frontend and optionally by tooling.

## Planned packages

| Package | Purpose | Phase |
|---------|---------|-------|
| `ui/` | Shared React components, design tokens | Phase 0.4 |
| `sdk/` | TypeScript API client generated from OpenAPI | Phase 2 |
| `config/` | Shared ESLint/TS/Tailwind presets | Phase 2 |

## Rules

- No Go code here — Go shared code uses standard module paths under `internal/` or future `pkg/`.
- Packages must be versioned and tested independently where published.
- Do not put application routes or pages here.

## Status

Empty — created in Phase 2+.
