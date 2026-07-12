# Contributing to Vyntrio OS

Vyntrio OS is a commercial Linux-based home-server platform. Contributions must follow documented phases and production-quality standards.

## Before you start

1. Read `docs/00_PROJECT.md` and `docs/01_MASTERPLAN.md`
2. Read `docs/02_ARCHITECTURE.md` and `docs/ADR/0001-use-clean-architecture.md`
3. Check the **active phase** in `docs/20_TASKS.md` and `docs/PHASE_1_FOUNDATION_STATUS.md`
4. Read `docs/17_SECURITY.md` — never commit secrets

## Development setup

```bash
git clone git@github.com:crankurbex2025-source/vyntrio-os.git
cd vyntrio-os
cp .env.example .env          # optional locally; never commit .env
make bootstrap
make verify
make test
```

Full instructions: `docs/21_CURSOR_REPO_SETUP.md`

## Branch workflow

```bash
git checkout main
git pull origin main
git checkout -b feature/short-description
# ... changes ...
make verify && make test
git push -u origin feature/short-description
```

Open a PR using the template. Link docs/tasks and include a test plan.

## Code organization

| Path | Purpose |
|------|---------|
| `cmd/` | Binary entrypoints only |
| `internal/domain/` | Business logic, no external deps |
| `internal/application/` | Use cases and ports |
| `internal/infrastructure/` | Adapters |
| `internal/interfaces/` | HTTP/WebSocket handlers |
| `frontend/` | React dashboard |
| `packages/` | Shared TS libraries |
| `tests/` | Cross-cutting integration/E2E |

## Architecture decisions

Irreversible technology choices require an ADR in `docs/ADR/`. Use `ADR_TEMPLATE.md`.

## What we reject

- Features outside the active roadmap phase
- Demo/mock services presented as production code
- Secrets, tokens, or `.env` files in git
- Domain logic in `cmd/` or infrastructure imported by domain
- Skipping docs updates when structure changes

## Style

- Go: `gofmt`, `go vet`; tabs per `.editorconfig`
- Frontend: conventions in `docs/08_FRONTEND.md` (Phase 2+)
- Commits: clear, focused; reference phase/task when helpful

## License

See `LICENSE` — legal choice pending. Do not assume open-source redistribution until finalized.
