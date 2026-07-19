# internal — Private Application Code

All Go code that is not a `main` package lives here. This tree follows Clean Architecture (see `docs/ADR/0001-use-clean-architecture.md`).

## Layers

| Directory | Layer | Contains |
|-----------|-------|----------|
| `domain/` | Domain | Entities, value objects, domain services — **no external deps** |
| `application/` | Application | Use cases, orchestration, policies, ports |
| `infrastructure/` | Infrastructure | DB, Docker, libvirt, systemd, file system adapters |
| `interfaces/` | Interfaces | HTTP handlers, WebSocket, CLI adapters |

## Dependency rule

Dependencies point **inward only**:

```text
interfaces → application → domain
infrastructure → application (via interfaces/ports)
```

## What must NOT go here

- `main()` functions (use `cmd/`)
- React/frontend code (use `frontend/`)
- Shared UI libraries (use `packages/`)
- Test fixtures (use `tests/fixtures/`)

## Status

Populated — domain, application, infrastructure, and HTTP interfaces are in active use (`vyntrio-api`, `vyntrio-backup`). See `docs/07_BACKEND.md` and `docs/02_ARCHITECTURE.md`.
