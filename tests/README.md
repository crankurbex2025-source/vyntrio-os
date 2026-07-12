# tests

Cross-cutting test suites that span multiple packages or require external dependencies.

## Layout

| Directory | Purpose |
|-----------|---------|
| `integration/` | Go integration tests (DB, adapters) |
| `e2e/` | Playwright end-to-end dashboard tests |
| `fixtures/` | Shared test data, configs, certificates |

## Rules

- Unit tests live next to source (`*_test.go` in `internal/`, Vitest in `frontend/`).
- **Do not** put production code here.
- E2E tests require a running stack — documented in `docs/18_TESTING.md`.

## Status

Empty — populated from Phase 2 onward.

See `docs/18_TESTING.md`.
