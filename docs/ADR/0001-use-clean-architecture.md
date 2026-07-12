# ADR-0001: Use Clean Architecture

- **Status:** Accepted
- **Date:** 2026-07-12
- **Deciders:** Vyntrio OS architecture (documented in `docs/02_ARCHITECTURE.md`)

## Context

Vyntrio OS is a long-lived appliance platform integrating storage, containers, virtualization, licensing, and updates. The codebase must remain testable, modular, and safe to extend over many release cycles without turning into an unbounded monolith.

Alternatives considered:

- **Layered MVC in a single Go package** — fast initially, but coupling increases quickly across domains.
- **Microservices from day one** — operational overhead inappropriate for an appliance control plane on a single host.
- **Plugin-only architecture without core layers** — insufficient governance for security-critical core paths.

## Decision

Adopt **Clean Architecture** with four primary Go layers under `internal/`:

1. **Domain** — business rules, no external dependencies
2. **Application** — use cases, ports, orchestration
3. **Infrastructure** — adapters (DB, Docker, libvirt, systemd, …)
4. **Interfaces** — HTTP/WebSocket/CLI delivery

Entrypoints in `cmd/` wire configuration and start processes only.

## Consequences

### Positive

- Domain logic can be unit-tested without Docker, DB, or HTTP.
- Infrastructure can be swapped (SQLite → PostgreSQL) behind ports.
- Clear boundaries reduce accidental coupling between product modules.
- Aligns with documented architecture in `docs/02_ARCHITECTURE.md`.

### Negative / trade-offs

- More packages and boilerplate early in the project.
- Requires discipline in code review to enforce dependency direction.
- Mapping between layers adds small overhead.

### Follow-up

- [ ] Phase 2: Create first domain packages (identity, system inventory)
- [ ] Phase 2: Define port interfaces before adapters
- [ ] Ongoing: Architecture review for any layer violation in PRs

## References

- `docs/02_ARCHITECTURE.md`
- `docs/07_BACKEND.md`
- Robert C. Martin — Clean Architecture
