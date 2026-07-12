# ADR-0003: SQLite with golang-migrate (sqlc deferred)

- **Status:** Accepted
- **Date:** 2026-07-12
- **Deciders:** Vyntrio OS engineering (Slice 2 approval)

## Context

Slice 1 provides HTTP health endpoints with process-only readiness. Roadmap **0.3 Core Platform** and `docs/10_DATABASE.md` require SQLite as the initial embedded database. Slice 2 must add database bootstrap, versioned migrations, and DB-aware `/readyz` — without implementing settings API, auth, or full entity models.

`docs/04_TECH_STACK.md` prefers sqlc or a GORM-free query layer. Introducing sqlc before any read/write queries adds tooling overhead without value.

## Decision

1. **Database:** SQLite file at `{VYNTRIO_DATA_DIR}/vyntrio.db`
2. **Driver:** [`modernc.org/sqlite`](https://modernc.org/sqlite) (pure Go, no CGO)
3. **Migrations:** [`github.com/golang-migrate/migrate/v4`](https://github.com/golang-migrate/migrate) with SQL files in `migrations/`
4. **Queries (Slice 2):** `database/sql` ping only — **no sqlc yet**
5. **sqlc:** Deferred to Slice 2b when settings/users queries are implemented

## Rationale

| Option | Verdict |
|--------|---------|
| modernc.org/sqlite | Pure Go — simpler CI and cross-compilation for appliance builds |
| mattn/go-sqlite3 | Mature but CGO — deferred |
| golang-migrate | Versioned up/down SQL; matches `docs/10_DATABASE.md` migration rules |
| goose | Viable alternative — not chosen to avoid split tooling conventions |
| GORM | Rejected — contradicts GORM-free preference |
| sqlc in Slice 2 | Rejected — no queries to generate yet |
| PostgreSQL now | Rejected — SQLite first per architecture docs |

## Schema (Slice 2 only)

Single bootstrap table `schema_meta` (key/value). Full entities (Users, Settings, …) deferred.

## Consequences

### Positive

- `/readyz` reflects real DB connectivity
- Migration discipline established early
- Clean Architecture preserved: `application/health` port, `infrastructure/persistence` adapter
- sqlc can be added without reversing migration approach

### Negative / trade-offs

- New dependencies (migrate, modernc sqlite)
- golang-migrate + modernc requires `WithInstance` wiring (no CGO sqlite3 driver)
- PostgreSQL adapter still to be designed later

### Follow-up

- [ ] Slice 2b: sqlc + settings schema
- [ ] Graceful shutdown with DB close
- [ ] PostgreSQL ADR when Enterprise path is scheduled

## References

- `docs/10_DATABASE.md`
- `docs/04_TECH_STACK.md`
- `docs/ADR/0001-use-clean-architecture.md`
- `docs/API_CONVENTIONS.md`
