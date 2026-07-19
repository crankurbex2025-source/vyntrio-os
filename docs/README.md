# Vyntrio OS — Documentation Index

Read in this order for onboarding.

## Product & strategy

| Doc | Topic |
|-----|-------|
| [00_PROJECT.md](00_PROJECT.md) | Vision, goals, positioning, **primary USB-first delivery path** |
| [01_MASTERPLAN.md](01_MASTERPLAN.md) | Phases, governance, delivery |
| [02_ARCHITECTURE.md](02_ARCHITECTURE.md) | Clean Architecture, modules |
| [03_ROADMAP.md](03_ROADMAP.md) | Version roadmap (0.1 → 1.0) |
| [04_TECH_STACK.md](04_TECH_STACK.md) | Technology choices |

## Domain specifications (future phases — read before implementing)

| Doc | Topic | Phase |
|-----|-------|-------|
| [05_STORAGE.md](05_STORAGE.md) | Storage | 0.5 |
| [06_DASHBOARD.md](06_DASHBOARD.md) | Dashboard UX | 0.4 |
| [07_BACKEND.md](07_BACKEND.md) | Backend services | 0.3 |
| [08_FRONTEND.md](08_FRONTEND.md) | Frontend | 0.4 |
| [09_API.md](09_API.md) | REST/WebSocket API | 0.3 |
| [10_DATABASE.md](10_DATABASE.md) | Persistence | 0.3 |
| [11_AUTH.md](11_AUTH.md) | Authentication | 0.3 |
| [12_NETWORK.md](12_NETWORK.md) | Networking | Later |
| [13_DOCKER.md](13_DOCKER.md) | Containers | 0.6 |
| [14_VM.md](14_VM.md) | Virtualization | 0.7 |
| [15_LICENSE.md](15_LICENSE.md) | Licensing | 0.9 |
| [16_UPDATE.md](16_UPDATE.md) | Updates | Later |
| [17_SECURITY.md](17_SECURITY.md) | Security program |
| [18_TESTING.md](18_TESTING.md) | Test strategy |
| [19_RELEASE.md](19_RELEASE.md) | Release engineering |
| [24_INSTALL_MEDIA.md](24_INSTALL_MEDIA.md) | Download & install media (BIOS raw image) |

## Engineering & process

| Doc | Topic |
|-----|-------|
| [20_TASKS.md](20_TASKS.md) | Development backlog |
| [21_CURSOR_REPO_SETUP.md](21_CURSOR_REPO_SETUP.md) | GitHub, SSH, Cursor setup |
| [API_CONVENTIONS.md](API_CONVENTIONS.md) | HTTP errors, env vars, logging |
| [ADR/](ADR/) | Architecture decision records |
| [AUDIT_FOUNDATION.md](AUDIT_FOUNDATION.md) | Phase 0 audit |
| [PHASE_1_FOUNDATION_STATUS.md](PHASE_1_FOUNDATION_STATUS.md) | Phase 1 status |

## Root-level references

- [../README.md](../README.md)
- [../REPOSITORY_STRUCTURE.md](../REPOSITORY_STRUCTURE.md)
- [../CONTRIBUTING.md](../CONTRIBUTING.md)
- [../cursor-prompts/](../cursor-prompts/) — phase prompts for Cursor agents

## Conventions

- Numbered docs (`00_`–`21_`) are canonical
- Do not implement features outside the active phase
- ADR required for irreversible tech decisions (`docs/ADR/`)
