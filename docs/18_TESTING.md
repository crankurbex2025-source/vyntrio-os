# 18 — Testing

> **Status:** DRAFT — testing strategy to be finalized before Phase 1 code.

## Testing principles

1. Test behavior, not implementation details
2. Prefer fast unit tests; use integration tests at boundaries
3. No merging broken CI
4. Document manual test steps when automation is not yet available

## Test pyramid (target)

```text
        ┌─────────────┐
        │  E2E (few)  │
        ├─────────────┤
        │ Integration │
        ├─────────────┤
        │ Unit (many) │
        └─────────────┘
```

## Backend (Go)

**Current:** `go test ./...` runs in CI; no packages yet.

Planned:

| Level | Tooling | Location |
|-------|---------|----------|
| Unit | `testing`, testify (TBD) | `backend/...*_test.go` |
| Integration | `testing` + testcontainers (TBD) | `backend/...` |
| API | httptest / contract tests (TBD) | TBD |

Commands:

```bash
make test-backend
cd backend && go test -race ./...
```

## Frontend

**Current:** placeholder npm script; no framework yet.

Planned (after ADR):

| Level | Tooling | Status |
|-------|---------|--------|
| Unit | TBD (Vitest/Jest) | Pending |
| Component | TBD | Pending |
| E2E | TBD (Playwright/Cypress) | Pending |

Commands:

```bash
make test-frontend
```

## CI

See `.github/workflows/ci.yml`:

- Foundation doc checks
- `go vet` + `go test`
- Frontend npm test script

TODO for Phase 1+:

- Coverage thresholds
- Lint gates (golangci-lint, eslint)
- E2E on PR (optional/nightly)

## Test data

TODO:

- Fixtures location
- Seed data strategy
- No production data in tests

## Related documents

- [04_TECH_STACK.md](04_TECH_STACK.md)
- [20_TASKS.md](20_TASKS.md)
