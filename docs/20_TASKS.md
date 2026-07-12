# 20 — Tasks

Phase-gated task list. **Only work on tasks in the active phase** unless explicitly approved.

## Active phase: Phase 0

### Phase 0 — Repository audit and foundation hardening

| ID | Task | Status | Owner |
|----|------|--------|-------|
| P0-1 | Audit repository against expected monorepo layout | Done | Agent |
| P0-2 | Create foundation files (Makefile, CI, .gitignore, etc.) | Done | Agent |
| P0-3 | Create documentation scaffolds and index | Done | Agent |
| P0-4 | Write `docs/AUDIT_FOUNDATION.md` | Done | Agent |
| P0-5 | **Human:** Author `docs/00_PROJECT.md` (replace TODOs) | **Blocked** | Project owner |
| P0-6 | **Human:** Author `docs/01_MASTERPLAN.md` | **Blocked** | Project owner |
| P0-7 | **Human:** Decide LICENSE (see `LICENSE`) | **Blocked** | Legal / owner |
| P0-8 | Mark Phase 0 complete in roadmap | Pending | After P0-5–P0-7 |

**Phase 0 exit criteria:**

- [x] Monorepo skeleton exists
- [x] CI runs foundation checks
- [x] Audit report published
- [ ] Core project docs authored (not just scaffolds)
- [ ] License decision recorded

---

## Next phase: Phase 1 (planned — do not start until exit criteria met)

See [AUDIT_FOUNDATION.md](AUDIT_FOUNDATION.md) for the recommended first implementation slice.

### Phase 1 — Documentation baseline + platform skeleton

| ID | Task | Status | Depends on |
|----|------|--------|------------|
| P1-1 | Finalize `02_ARCHITECTURE.md` | Pending | P0-5, P0-6 |
| P1-2 | ADR: frontend framework + toolchain | Pending | P1-1 |
| P1-3 | ADR: persistence layer (if needed) | Pending | P1-1 |
| P1-4 | Backend: `cmd/` health service (`/healthz`, `/readyz`) | Pending | P1-1 |
| P1-5 | Frontend: initialize chosen toolchain (no feature UI) | Pending | P1-2 |
| P1-6 | Wire local dev: `make dev` (backend + frontend) | Pending | P1-4, P1-5 |
| P1-7 | Add `package-lock.json`, golangci-lint config | Pending | P1-4, P1-5 |
| P1-8 | Expand CI: lint + coverage reporting | Pending | P1-7 |
| P1-9 | Update README with runnable bootstrap | Pending | P1-6 |

---

## Backlog (unscheduled)

- Dependabot configuration
- CODEOWNERS
- Container images
- Staging environment
- Full threat model in `17_SECURITY.md`

## Related documents

- [03_ROADMAP.md](03_ROADMAP.md)
- [AUDIT_FOUNDATION.md](AUDIT_FOUNDATION.md)
- [21_CURSOR_REPO_SETUP.md](21_CURSOR_REPO_SETUP.md)
