# Architecture Decision Records

Significant, hard-to-reverse technical decisions for Vyntrio OS.

Governance: see `docs/01_MASTERPLAN.md` — ADR required for irreversible technology choices.

## Index

| ADR | Status | Title |
|-----|--------|-------|
| [0001](0001-use-clean-architecture.md) | **Accepted** | Use Clean Architecture |
| [0002](0002-use-chi-router.md) | **Accepted** | Use chi for HTTP routing |
| [0003](0003-sqlite-migrations.md) | **Accepted** | SQLite with golang-migrate and sqlc |
| [Template](ADR_TEMPLATE.md) | Template | ADR template |

## Process

1. Copy `ADR_TEMPLATE.md` to `NNNN-short-title.md`
2. Set status to **Proposed**, discuss in PR
3. On acceptance, update index and link from relevant docs
4. Supersede rather than delete — mark old ADR deprecated

## Related

- Root-level `adr/` from Phase 0 was consolidated here in Phase 1.
