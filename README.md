# Vyntrio OS

Commercial Linux-based home-server and private-cloud operating system.

Vyntrio OS combines storage, containers, virtualization, networking, backup, monitoring, automation, app marketplace, and commercial licensing in a unified appliance platform with a modern web dashboard.

**Repository status:** Phase 1 — Foundation Hardening (implementation-ready monorepo skeleton).

## Quick start

```bash
git clone git@github.com:crankurbex2025-source/vyntrio-os.git
cd vyntrio-os
make bootstrap
make verify
make test
make build
```

Requires: Git, Make, Go 1.24+, Node.js 20+ (for frontend phase).

Setup details: **[docs/21_CURSOR_REPO_SETUP.md](docs/21_CURSOR_REPO_SETUP.md)**

## Repository layout

```text
vyntrio-os/
├── cmd/              # Service entrypoints (api, worker, installer, update-agent)
├── internal/         # Clean Architecture layers (domain, application, …)
├── frontend/         # React dashboard (toolchain: Phase 2)
├── packages/         # Shared TS libraries (ui, sdk, config)
├── build/            # Build scripts (Phase 0.2+)
├── distro/           # ISO / image construction (Phase 0.2+)
├── docs/             # Canonical product & engineering documentation
├── tests/            # Integration and E2E tests
├── scripts/          # Bootstrap and validation
├── cursor-prompts/   # Phase-gated Cursor agent prompts
├── ops/              # Release and operations checklists
└── .github/          # CI/CD and contribution templates
```

See also: [REPOSITORY_STRUCTURE.md](REPOSITORY_STRUCTURE.md)

## Documentation

| Document | Description |
|----------|-------------|
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/00_PROJECT.md](docs/00_PROJECT.md) | Product vision and goals |
| [docs/02_ARCHITECTURE.md](docs/02_ARCHITECTURE.md) | System architecture |
| [docs/03_ROADMAP.md](docs/03_ROADMAP.md) | Delivery roadmap |
| [docs/20_TASKS.md](docs/20_TASKS.md) | Development backlog |
| [docs/PHASE_1_FOUNDATION_STATUS.md](docs/PHASE_1_FOUNDATION_STATUS.md) | Current foundation status |

## Development commands

```bash
make help          # List commands
make bootstrap     # Toolchain check + workspace prep
make verify        # Layout + docs validation
make test          # Go + frontend tests
make build         # Compile Go binaries (stubs)
make docs-check    # Required documentation present
```

## Principles

1. Read documentation before coding
2. Modular Clean Architecture
3. Production quality — no demo logic
4. Security and tests as release gates
5. Phase-gated delivery (see `docs/03_ROADMAP.md`)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Follow the active phase in `docs/20_TASKS.md`.

## License

Not finalized — see [LICENSE](LICENSE).

## Links

- GitHub: https://github.com/crankurbex2025-source/vyntrio-os
- Server deployment path: `/opt/vyntrio-os`
