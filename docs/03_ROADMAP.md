# 03 — Roadmap

> **Status:** DRAFT — roadmap skeleton. Dates and scope require owner input.

## Roadmap principles

- Documentation and foundation before features
- Each phase has explicit entry/exit criteria
- No phase skipping without ADR

## Phases (proposed)

### Phase 0 — Repository audit and foundation hardening

**Status:** In progress

- Monorepo skeleton
- CI foundation
- Documentation scaffolds
- Audit report

**Exit criteria:** See [AUDIT_FOUNDATION.md](AUDIT_FOUNDATION.md)

### Phase 1 — Core platform slice (TBD)

**Status:** Planned — see [20_TASKS.md](20_TASKS.md) and [AUDIT_FOUNDATION.md](AUDIT_FOUNDATION.md)

TODO: Define exact scope after Phase 0 docs are completed.

Proposed themes (subject to change):

- Authoritative project docs (`00`–`04`, `17`, `18`)
- Backend health/service skeleton
- Frontend toolchain selection (ADR required)
- Local dev bootstrap end-to-end

### Phase 2 — TODO

TODO: First user-visible capability.

### Phase 3 — TODO

TODO: Hardening, security review, expanded test coverage.

## Milestones

| Milestone | Target | Depends on |
|-----------|--------|------------|
| M0: Foundation audit complete | Phase 0 | — |
| M1: Docs baseline accepted | Phase 1 entry | M0 |
| M2: First runnable vertical slice | Phase 1 exit | M1 |
| TODO | | |

## Related documents

- [20_TASKS.md](20_TASKS.md)
- [AUDIT_FOUNDATION.md](AUDIT_FOUNDATION.md)
