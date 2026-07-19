# cmd — Service Entrypoints

Go `main` packages for Vyntrio OS binaries. Each subdirectory is one deployable process.

## Binaries

| Directory | Binary | Purpose | Phase |
|-----------|--------|---------|-------|
| `api/` | `vyntrio-api` | REST control plane + embedded SPA (`go:embed`) | Block 8+ (shipped) |
| `worker/` | `vyntrio-worker` | Background jobs and long-running tasks | Phase 2 |
| `installer/` | `vyntrio-installer` | Block 10 internal infra (secondary to USB-first): preflight, sandbox install (dev/lab), apply, postflight, workflow | Block 10 |
| `update-agent/` | `vyntrio-update-agent` | Signed update application | Later |
| `backup/` | `vyntrio-backup` | Root-only local appliance backup (Slice 7.9) | Block 7 (shipped) |
| `restore/` | `vyntrio-restore` | Root-only offline restore from local backups (Slice 7.12) | Block 7 (shipped) |
| `verify-artifact/` | `vyntrio-verify-artifact` | Local release manifest + SHA-256 verification (Slice 9.10) | Block 9 (shipped) |

## Rules

- **Only** wiring, config loading, and server/process startup belong here.
- Business logic belongs in `internal/application/` and `internal/domain/`.
- HTTP handlers belong in `internal/interfaces/`, not in `cmd/`.
- No business logic in `cmd/` — only wiring and process startup.

## Build

```bash
make build
# or
go build -o bin/vyntrio-api ./cmd/api
go build -o bin/vyntrio-backup ./cmd/backup
go build -o bin/vyntrio-restore ./cmd/restore
go build -o bin/vyntrio-verify-artifact ./cmd/verify-artifact
```

See `docs/07_BACKEND.md` and `docs/02_ARCHITECTURE.md`.
