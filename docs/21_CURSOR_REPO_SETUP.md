# 21 — Cursor / GitHub Repository Setup

Bootstrap instructions for developers using Cursor IDE and GitHub.

## Prerequisites

| Tool | Minimum version | Verify |
|------|-----------------|--------|
| Git | 2.30+ | `git --version` |
| Go | 1.22+ | `go version` |
| Node.js | 20+ | `node --version` |
| npm | 10+ | `npm --version` |
| Make | any | `make --version` |
| GitHub CLI (optional) | 2.40+ | `gh --version` |

## Clone the repository

```bash
# HTTPS (recommended for first setup)
git clone https://github.com/crankurbex2025-source/vyntrio-os.git
cd vyntrio-os

# SSH (after SSH key is configured on GitHub)
# git clone git@github.com:crankurbex2025-source/vyntrio-os.git
```

**Common mistake:** Do not mix SSH and HTTPS syntax:

```bash
# WRONG
git clone git@https://github.com/crankurbex2025-source/vyntrio-os

# CORRECT
git clone https://github.com/crankurbex2025-source/vyntrio-os.git
git clone git@github.com:crankurbex2025-source/vyntrio-os.git
```

## Server deployment path

On the project server, the repository lives at:

```text
/opt/vyntrio-os
```

## Open in Cursor

1. **File → Open Folder**
2. Select the repository root (`vyntrio-os/`, not `backend/` alone)
3. Confirm the root contains `docs/`, `backend/`, `frontend/`, and `Makefile`

## Initial bootstrap

```bash
cp .env.example .env   # edit as needed; never commit .env
make bootstrap
make verify
make docs-check
make test
```

Expected Phase 0 result:

- All doc checks pass
- Go tests pass (no packages yet)
- Frontend test script runs (placeholder message)

## GitHub authentication

For HTTPS push/pull, use one of:

- `gh auth login` then `gh auth setup-git`
- Personal Access Token with `repo` scope (classic) or **Contents: Read and write** (fine-grained)

**Never paste tokens into chat or commit them.** Revoke any exposed token immediately.

## Branch workflow

```bash
git checkout main
git pull origin main
git checkout -b feature/short-description
# ... work ...
make verify && make test
git push -u origin feature/short-description
# Open PR on GitHub
```

## Cursor agent guidelines

When using Cursor Agent on this repo:

1. Read `docs/20_TASKS.md` for the active phase
2. Do not build unplanned features or demo code
3. Update docs when changing structure or behavior
4. Run `make verify` before finishing

## CI

Pull requests to `main` trigger `.github/workflows/ci.yml`:

- Required docs presence
- Monorepo layout verification
- Go vet/test
- Frontend npm test

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `fatal: not a git repository` | `cd` to repo root; ensure `.git` exists |
| `Permission denied (403)` on push | Token lacks write scope; use classic `repo` or fine-grained Contents write |
| `go: command not found` | Install Go 1.22+ |
| `docs-check failed` | Missing doc file; see `scripts/docs-check.sh` list |
| Wrong directory opened in Cursor | Open repo root, not subdirectory |

## Related documents

- [../README.md](../README.md)
- [20_TASKS.md](20_TASKS.md)
- [17_SECURITY.md](17_SECURITY.md)
