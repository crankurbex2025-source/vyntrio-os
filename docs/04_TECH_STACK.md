# 04 — Tech Stack

> **Status:** DRAFT — proposed baseline from Phase 0 foundation.
>
> Final selections require ADR acceptance before Phase 1 implementation.

## Repository

| Area | Choice | Status |
|------|--------|--------|
| VCS / hosting | Git + GitHub | **Active** |
| Monorepo layout | `backend/` + `frontend/` + `docs/` | **Active (foundation)** |
| CI | GitHub Actions (`.github/workflows/ci.yml`) | **Active (foundation)** |
| Editor defaults | `.editorconfig` | **Active** |

## Backend (proposed)

| Area | Proposed | Status | Notes |
|------|----------|--------|-------|
| Language | Go 1.22+ | Proposed | `backend/go.mod` initialized |
| Module path | `github.com/crankurbex2025-source/vyntrio-os/backend` | Active | |
| HTTP framework | TBD | **Requires ADR** | e.g. stdlib `net/http`, chi, echo |
| Config | Environment variables + `.env.example` | Proposed | |
| Database | TBD | **Requires ADR** | Document in `02_ARCHITECTURE.md` |
| Migrations | TBD | Pending | |
| Logging | TBD | Pending | Structured logging preferred |

## Frontend (proposed)

| Area | Proposed | Status | Notes |
|------|----------|--------|-------|
| Runtime | Node.js 20+ | Proposed | `engines` in `package.json` |
| Package manager | npm (lockfile TBD) | Proposed | Add `package-lock.json` in Phase 1 |
| Framework | TBD | **Requires ADR** | React, Vue, Svelte, etc. |
| Build tool | TBD | Pending | Vite, etc. |
| Testing | TBD | Pending | See [18_TESTING.md](18_TESTING.md) |

## Infrastructure (future)

| Area | Proposed | Status |
|------|----------|--------|
| Containers | Docker (TBD) | Not started |
| Orchestration | TBD | Not started |
| Reverse proxy | TBD | Not started |

## Development tooling

| Tool | Purpose | Command |
|------|---------|---------|
| Make | Unified dev commands | `make help` |
| `scripts/docs-check.sh` | Required docs validation | `make docs-check` |
| Go toolchain | Backend build/test | `make test-backend` |
| npm | Frontend scripts | `make test-frontend` |

## Version policy

TODO:

- Go version support window
- Node LTS policy
- Dependency update cadence

## ADR triggers

Create an ADR before committing to:

- Frontend framework and build pipeline
- Database and persistence layer
- Authentication/authorization model
- Deployment platform

## Related documents

- [02_ARCHITECTURE.md](02_ARCHITECTURE.md)
- [18_TESTING.md](18_TESTING.md)
- [../adr/README.md](../adr/README.md)
