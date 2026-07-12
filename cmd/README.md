# cmd — Service Entrypoints

Go `main` packages for Vyntrio OS binaries. Each subdirectory is one deployable process.

## Binaries

| Directory | Binary | Purpose | Phase |
|-----------|--------|---------|-------|
| `api/` | `vyntrio-api` | REST/WebSocket control plane | Phase 2 |
| `worker/` | `vyntrio-worker` | Background jobs and long-running tasks | Phase 2 |
| `installer/` | `vyntrio-installer` | OS installation wizard / first boot | Phase 0.2 |
| `update-agent/` | `vyntrio-update-agent` | Signed update application | Later |

## Rules

- **Only** wiring, config loading, and server/process startup belong here.
- Business logic belongs in `internal/application/` and `internal/domain/`.
- HTTP handlers belong in `internal/interfaces/`, not in `cmd/`.
- No feature implementation during Phase 1 — stubs compile only.

## Build

```bash
make build
# or
go build -o bin/vyntrio-api ./cmd/api
```

See `docs/07_BACKEND.md` and `docs/02_ARCHITECTURE.md`.
