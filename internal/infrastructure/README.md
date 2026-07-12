# internal/infrastructure

Adapters for external systems: SQLite/PostgreSQL, Docker, libvirt, systemd, nftables, SMART, signing.

- Implements ports defined in `internal/application/`.
- May import third-party libraries.
- Integration tests live alongside or under `tests/integration/`.

See `docs/04_TECH_STACK.md`.
