# Cursor / GitHub Repository Setup

Setup for GitHub, SSH, Cursor IDE, and the project server.

## Prerequisites

| Tool | Version | Verify |
|------|---------|--------|
| Git | 2.30+ | `git --version` |
| Go | 1.24+ | `go version` |
| Node.js | 20+ | `node --version` |
| npm | 10+ | `npm --version` |
| Make | any | `make --version` |

Optional: [GitHub CLI](https://cli.github.com/) (`gh`), SSH key for GitHub.

## Clone (HTTPS)

```bash
git clone https://github.com/crankurbex2025-source/vyntrio-os.git
cd vyntrio-os
make bootstrap
```

## Clone (SSH — recommended for daily development)

### 1. Generate SSH key (if needed)

```bash
ssh-keygen -t ed25519 -C "your@email.com" -f ~/.ssh/id_ed25519_github
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519_github
```

Add the public key (`~/.ssh/id_ed25519_github.pub`) to GitHub → **Settings → SSH and GPG keys**.

### 2. Test SSH

```bash
ssh -T git@github.com
```

### 3. Clone

```bash
git clone git@github.com:crankurbex2025-source/vyntrio-os.git
cd vyntrio-os
```

**Common mistake — do not mix SSH and HTTPS:**

```bash
# WRONG
git clone git@https://github.com/crankurbex2025-source/vyntrio-os

# CORRECT
git clone git@github.com:crankurbex2025-source/vyntrio-os.git
```

## Server deployment

Production/development server path:

```text
/opt/vyntrio-os
```

After clone or pull on the server:

```bash
cd /opt/vyntrio-os
make bootstrap
make verify
```

## Open in Cursor

1. **File → Open Folder**
2. Select repository **root** (`vyntrio-os/`)
3. Confirm presence of `docs/`, `cmd/`, `internal/`, `Makefile`

### Cursor agent rules

- Read `docs/20_TASKS.md` for the active phase
- Read `docs/PHASE_1_FOUNDATION_STATUS.md` for current foundation state
- Use prompts in `cursor-prompts/` for phase-specific work
- No demo features, no phase skipping

## Local configuration

```bash
cp .env.example .env              # root — optional for Phase 1
cp frontend/.env.example frontend/.env.local   # when frontend is active
```

Never commit `.env` files. See `docs/17_SECURITY.md`.

## GitHub authentication (HTTPS)

If not using SSH:

```bash
gh auth login
gh auth setup-git
```

Or use a Personal Access Token with **repo** scope (classic) or **Contents: Read and write** (fine-grained).

**Never paste tokens in chat or commit them.**

## Branch workflow

```bash
git checkout main
git pull origin main
git checkout -b feature/short-description
make verify && make test
git push -u origin feature/short-description
```

Enable branch protection on `main` (GitHub → Settings → Branches):

- Require PR before merge
- Require status checks: `CI / Foundation validation`, `CI / Go build & test`

## CI overview

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | push/PR to `main` | Docs, layout, Go build/test, frontend scripts |
| `security.yml` | weekly + manual | Secret scan, govulncheck |
| `release.yml` | manual | Pre-release validation (artifacts in later phases) |

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `fatal: not a git repository` | `cd` to repo root |
| `Permission denied (403)` | Token lacks write scope; use SSH or new token |
| `go: command not found` | Install Go 1.24+ |
| `docs-check failed` | Missing doc — see `scripts/docs-check.sh` |
| Opened subfolder in Cursor | Re-open repo root |

## Related

- [../README.md](../README.md)
- [20_TASKS.md](20_TASKS.md)
- [17_SECURITY.md](17_SECURITY.md)
