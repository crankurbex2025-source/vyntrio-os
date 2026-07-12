# internal/domain

Pure business logic for Vyntrio OS: users, storage, containers, VMs, licenses, updates, etc.

- No imports from `internal/infrastructure` or `internal/interfaces`.
- No framework dependencies (HTTP, SQL, Docker SDK).
- Unit-tested without mocks where possible.

See `docs/02_ARCHITECTURE.md`.
