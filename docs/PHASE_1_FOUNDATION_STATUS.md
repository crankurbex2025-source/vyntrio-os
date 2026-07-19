# Phase 1 — Foundation Hardening Status

**Phase:** 1 — Foundation Hardening  
**Status:** Complete (pending human actions P1-10, P1-11)  
**Date:** 2026-07-12

---

## Objective

Turn the repository into a clean, professional, implementation-ready monorepo foundation with predictable developer onboarding for GitHub + SSH + Cursor — **without** implementing product features (storage, Docker, VM, marketplace, licensing).

---

## What was improved

### Repository structure

- Merged canonical project layout from source bundle into repo root
- **Root-level `go.mod`** (module: `github.com/crankurbex2025-source/vyntrio-os`)
- Removed conflicting Phase 0 `backend/` layout
- Removed committed `vyntrio-os-repository.zip` (added `*.zip` to `.gitignore`)
- Created directory tree: `cmd/`, `internal/`, `frontend/`, `packages/`, `tests/`, `build/`, `distro/`

### Code boundaries

- `cmd/*/main.go` — entrypoints; `vyntrio-api` and `vyntrio-backup` are functional (Blocks 7–8); others remain stubs or sandbox scope
- Layer READMEs document what belongs where and dependency rules
- ADR-0001 documents Clean Architecture decision

### Developer experience

- `scripts/bootstrap.sh` — toolchain check, npm install, verify + docs-check
- `scripts/verify-layout.sh` — required files and directories
- `scripts/docs-check.sh` — required documentation list
- `Makefile` — bootstrap, verify, fmt, lint, test, build, docs-check
- `.env.example` (root) + `frontend/.env.example`
- Improved `docs/21_CURSOR_REPO_SETUP.md` (HTTPS + SSH + Cursor + server path)

### Engineering foundation

- `CONTRIBUTING.md` — phase-gated contribution rules
- `docs/README.md` — full documentation index
- `docs/ADR/` — ADR template + ADR-0001
- `.github/PULL_REQUEST_TEMPLATE.md`
- `.github/ISSUE_TEMPLATE/bug_report.md`, `feature_request.md`

### CI/CD

- **`ci.yml`** — docs check, layout verify, gofmt, go vet/test, `make build`, frontend scripts
- **`security.yml`** — secret pattern scan, govulncheck (non-blocking until deps grow)
- **`release.yml`** — manual dispatch with validation only (no fake ISO publish)

### Documentation

- Real product docs restored (`docs/00`–`21` from canonical source)
- Phase 0 audit preserved: `docs/AUDIT_FOUNDATION.md`
- Phase 1 tasks added to top of `docs/20_TASKS.md`

---

## What is still missing

| Item | Owner | Blocks |
|------|-------|--------|
| LICENSE finalization | Legal / owner | External contribution |
| GitHub branch protection | Owner | Merge safety |
| `golangci-lint` config | Phase 2+ | Strict lint gate |
| Concrete T0001–T1000 task definitions | Planning | Backlog hygiene |
| ISO/installer (`distro/`, `build/`) | Phase 0.2 | OS delivery — contracts/sandbox only today |
| PWA manifest / installability (11R.7) | — | **Done** (web shortcut only; no SW) |
| Appliance/login token convergence (11R.6) | — | **Done** |
| Full appliance dashboard (beyond overview/settings) | Block 11+ | Product UI |

---

## Post-Phase-1 progress (audit 2026-07-15)

Phase 1 exit criteria are met for repository foundation. The following shipped **after** Phase 1 and are **not** retroactive changes to the Phase 1 scope:

| Area | Status |
|------|--------|
| `vyntrio-api` | Real HTTP server with auth, settings, overview (Block 8); embedded SPA delivery |
| Frontend toolchain | Vite + React + TypeScript in production (`frontend/`, `package-lock.json`, CI) |
| `go.sum` | Present; Go dependencies in use |
| Public website | Live on `vyntrio.xyz`: `/`, `/download`, `/docs` (Block 11R.10–11R.12) |
| Appliance UI | `/login`, `/app` — session probe, read-only overview, Owner settings (not full dashboard) |
| Installer/media (Blocks 9–10) | Makefile sandbox scripts and contracts only — no real hardware install |

For current product surface status see `docs/22_WEBSITE.md`, `docs/08_FRONTEND.md`, and `docs/23_WEBSITE_DESIGN_DIRECTION.md`.

---

## Verification commands

```bash
make bootstrap
make verify
make test
make build
ls -la bin/    # vyntrio-api, vyntrio-worker, vyntrio-installer, vyntrio-update-agent
```

---

## Phase 2 recommendation — Platform Core

**Title:** Platform Core skeleton (auth, settings, jobs, health — minimal vertical slice)

**Prerequisites:** Phase 1 exit criteria met; LICENSE decided.

**Note (2026-07-15):** Several Phase 2 items are **partially or fully implemented** (health/version endpoints, real `vyntrio-api`, auth/settings, embedded frontend, public website). Treat the list below as historical planning guidance; verify against `docs/20_TASKS.md` and the audit-aligned docs before starting work.

### Scope (in order)

1. **ADR-0002:** HTTP router choice (chi vs gin per `docs/04_TECH_STACK.md`)
2. **ADR-0003:** SQLite schema/migration approach
3. **`internal/domain/`** — first entities (system health, settings)
4. **`internal/application/`** — health use case
5. **`internal/interfaces/http/`** — `GET /healthz`, `GET /readyz`, `GET /api/v1/version`
6. **`cmd/api/`** — real server with structured logging and env config
7. **Frontend** — extend beyond overview/settings; full dashboard widgets remain out of scope
8. **`make dev`** — run API + frontend together
9. **Tests** — unit tests for domain; httptest for health endpoints
10. **CI** — enable golangci-lint; fail on lint errors

### Explicitly out of scope for Phase 2

- Storage, Docker, VM, marketplace, licensing
- Full RBAC beyond current Owner role
- ISO/installer builds (real hardware install)
- Production TLS deployment automation

### Phase 2 exit criteria

- [x] `vyntrio-api` serves health/version and Block 8 contract locally
- [x] Frontend toolchain builds and ships via `go:embed`
- [ ] CI green with golangci-lint + expanded test gates
- [ ] ADR-0002 and ADR-0003 accepted

---

## Related documents

- [20_TASKS.md](20_TASKS.md)
- [03_ROADMAP.md](03_ROADMAP.md) — Phase 0.3 Core Platform
- [AUDIT_FOUNDATION.md](AUDIT_FOUNDATION.md)
- [ADR/0001-use-clean-architecture.md](ADR/0001-use-clean-architecture.md)
