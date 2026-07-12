# ADR-0003: SQLite with golang-migrate and sqlc

- **Status:** Accepted
- **Date:** 2026-07-12
- **Deciders:** Vyntrio OS engineering (Slice 2 approval; Slice 2b.1 sqlc adoption)

## Context

Slice 1 provides HTTP health endpoints with process-only readiness. Roadmap **0.3 Core Platform** and `docs/10_DATABASE.md` require SQLite as the initial embedded database. Slice 2 added database bootstrap, versioned migrations, and DB-aware `/readyz` — without implementing settings API, auth, or full entity models.

`docs/04_TECH_STACK.md` prefers sqlc or a GORM-free query layer. Slice 2 deferred sqlc because there were no read/write queries beyond ping.

## Decision

1. **Database:** SQLite file at `{VYNTRIO_DATA_DIR}/vyntrio.db`
2. **Driver:** [`modernc.org/sqlite`](https://modernc.org/sqlite) v1.46.1 (pure Go, no CGO; pinned for Go 1.24)
3. **Migrations:** [`github.com/golang-migrate/migrate/v4`](https://github.com/golang-migrate/migrate) with SQL files in `migrations/`
4. **Queries (Slice 2):** `database/sql` ping only
5. **Queries (Slice 2b.1+):** [sqlc](https://sqlc.dev/) for typed settings queries; generated code in `internal/infrastructure/persistence/sqlite/sqlcgen/`
6. **Settings (Slice 2b.1):** Internal repository only — no public HTTP settings API
7. **Settings (Slice 2c):** Internal read-only consumer in `cmd/api` — load and validate at startup, log only
8. **Settings (Slice 2d):** Propagate validated settings to structured logging (`hostname`; `timezone` at debug only)
9. **Settings (Slice 2e):** One-time read-only persistence verification against startup snapshot

## Rationale

| Option | Verdict |
|--------|---------|
| modernc.org/sqlite | Pure Go — simpler CI and cross-compilation for appliance builds |
| mattn/go-sqlite3 | Mature but CGO — deferred |
| golang-migrate | Versioned up/down SQL; matches `docs/10_DATABASE.md` migration rules |
| goose | Viable alternative — not chosen to avoid split tooling conventions |
| GORM | Rejected — contradicts GORM-free preference |
| sqlc in Slice 2 | Rejected — no queries to generate yet |
| sqlc in Slice 2b.1 | Accepted — settings get/upsert/list justify typed query generation |
| PostgreSQL now | Rejected — SQLite first per architecture docs |

## Schema

### Slice 2

Single bootstrap table `schema_meta` (key/value).

### Slice 2b.1

`settings` table with composite primary key `(namespace, key)`:

- **Namespace (v1):** `system` only
- **Keys (v1):** `timezone`, `hostname`
- **Seeds:** `system.timezone=UTC`, `system.hostname=vyntrio`

Full entities (Users, Roles, …) remain deferred.

## Consequences

### Positive

- `/readyz` reflects real DB connectivity
- Migration discipline established early
- Clean Architecture preserved: application ports, infrastructure adapters
- sqlc provides type-safe settings queries without ORM overhead
- Public settings API can be added later behind auth without reversing persistence approach

### Negative / trade-offs

- New dependencies (migrate, modernc sqlite, sqlc-generated code to maintain)
- golang-migrate + modernc requires `WithInstance` wiring (no CGO sqlite3 driver)
- sqlc `schema.sql` must stay aligned with migration SQL
- PostgreSQL adapter still to be designed later

### Follow-up

- [x] Slice 2b.1: sqlc + settings schema (internal repository)
- [x] Slice 2c: internal read-only settings consumer in `cmd/api`
- [x] Slice 2d: internal settings propagation to structured logging
- [x] Slice 2e: internal startup persistence verification (read-only)
- [ ] Public settings API (after auth/RBAC stub)
- [x] Graceful shutdown with DB close (Slice 3.1 — `cmd/api`)
- [ ] PostgreSQL ADR when Enterprise path is scheduled

## References

- `docs/10_DATABASE.md`
- `docs/04_TECH_STACK.md`
- `docs/ADR/0001-use-clean-architecture.md`
- `docs/API_CONVENTIONS.md`
