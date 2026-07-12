# Vyntrio OS

Documentation-first monorepo for the Vyntrio OS platform.

**Status:** Phase 0 — foundation hardened; core product docs pending author input.

## Repository layout

```text
vyntrio-os/
├── backend/          # Go services (module initialized, no app code yet)
├── frontend/         # Web UI workspace (toolchain TBD)
├── docs/             # Canonical documentation — start at docs/README.md
├── adr/              # Architecture decision records
├── scripts/          # Repo automation
├── .github/workflows/# CI
├── Makefile          # Dev commands
└── CONTRIBUTING.md
```

## Quick start

```bash
git clone https://github.com/crankurbex2025-source/vyntrio-os.git
cd vyntrio-os
cp .env.example .env    # optional for Phase 0
make bootstrap
make verify
make test
```

Requires: Git, Go 1.22+, Node.js 20+, Make.

Full setup: **[docs/21_CURSOR_REPO_SETUP.md](docs/21_CURSOR_REPO_SETUP.md)**

## Documentation

| Document | Description |
|----------|-------------|
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/AUDIT_FOUNDATION.md](docs/AUDIT_FOUNDATION.md) | Phase 0 audit report |
| [docs/20_TASKS.md](docs/20_TASKS.md) | Active phase and tasks |
| [docs/00_PROJECT.md](docs/00_PROJECT.md) | Project vision (**draft — TODO**) |

## Development commands

```bash
make help          # List commands
make bootstrap     # Install/check dependencies
make verify        # Validate monorepo layout + docs
make docs-check    # Required documentation present
make test          # Backend + frontend tests
```

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) and the active phase in [docs/20_TASKS.md](docs/20_TASKS.md) before opening a PR.

## License

**Not finalized.** See [LICENSE](LICENSE) — legal choice pending (TODO).

## Links

- GitHub: https://github.com/crankurbex2025-source/vyntrio-os
- Server path: `/opt/vyntrio-os`
