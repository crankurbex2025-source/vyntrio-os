# internal/application

Use cases and application services. Orchestrates domain logic, enforces policies, publishes events.

- Defines **ports** (interfaces) that infrastructure implements.
- No direct calls to Docker, libvirt, or SQL drivers.

See `docs/02_ARCHITECTURE.md` and `docs/07_BACKEND.md`.
