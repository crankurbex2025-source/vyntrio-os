# Backend (Go)

Go services and libraries for Vyntrio OS.

## Status

Foundation only — no application packages yet. See `docs/20_TASKS.md` for the implementation sequence.

## Layout (planned)

```
backend/
├── cmd/          # service entrypoints (Phase 1+)
├── internal/     # private application code
├── pkg/          # reusable libraries (if any)
└── go.mod
```

## Commands

From repository root:

```bash
make test-backend
cd backend && go test ./...
```

See `docs/04_TECH_STACK.md` and `docs/18_TESTING.md`.
