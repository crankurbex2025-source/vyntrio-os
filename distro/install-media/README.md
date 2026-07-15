# Install media scaffold (Block 9, Slice 9.3)

Declarative scaffold for **greenfield install USB/ISO** payloads. This directory
describes what install media **is for** and which artifacts it may carry. It does
**not** build bootable images, partition disks, or run an installer.

**Authority:** `docs/ADR/0006-appliance-install-recovery-media.md` (install media
sections B, D, E). Runtime paths and ownership remain `docs/ADR/0005-appliance-runtime-operations.md`.

## Status

| Item | State |
|------|--------|
| Layout manifest (`manifest.yaml`) | **Scaffold** — declarative only |
| Build contract (`build-contract.md`) | **Contract** — install-image I/O boundary (Slice 9.5); no builder |
| Config template (`config.toml.template`) | **Scaffold** — operator-edited at install |
| Boot/live environment | **Not started** — future slice |
| ISO/USB build pipeline | **Not started** — future slice |
| `vyntrio-installer` behavior | **Not started** — `cmd/installer` is a stub |
| Recovery media | **Separate deliverable** — not under `install-media/` |

## What install media is for

1. Boot a minimal live environment (future).
2. Lay down approved on-disk layout (future slice).
3. Copy **immutable payloads** from `manifest.yaml` onto the target disk.
4. Provision `vyntrio` account and directory layout per `distro/systemd/README.md`.
5. Install `config.toml` from the shipped **template** (no secrets, no Owner).
6. Enable `vyntrio-api.service` and hand off to ADR-0004 bootstrap on the
   **running** installed service.

## What install media is not for

- Persistent appliance state (`/var/lib/vyntrio/` SQLite, backups) — **target disk only**
- Owner credentials, bootstrap tokens, or license files
- In-place restore from backup archives (recovery media + Block 7 — separate)
- Product API, dashboard, or install-progress UI
- Storing operator backup artifacts

## Directory layout (repository scaffold)

```
distro/
├── install-media/          ← this scaffold (payload contract)
│   ├── README.md
│   ├── manifest.yaml       ← declarative payload list
│   └── config.toml.template
├── systemd/                ← runtime deployment artifacts (Block 7.3)
│   └── …
└── README.md
```

Future build slices may add `distro/install-media/boot/` or external build roots;
nothing in this slice executes or packages them.

## Payload sources

| Manifest role | Repository source | Installed path (target disk) |
|---------------|-------------------|------------------------------|
| API binary | `bin/vyntrio-api` (from `make build`) | `/usr/bin/vyntrio-api` |
| Backup binary | `bin/vyntrio-backup` (from `make build`) | `/usr/bin/vyntrio-backup` |
| systemd unit | `distro/systemd/vyntrio-api.service` | `/etc/systemd/system/vyntrio-api.service` |
| sysusers | `distro/systemd/vyntrio.sysusers` | `/usr/lib/sysusers.d/vyntrio.conf` |
| tmpfiles | `distro/systemd/vyntrio.tmpfiles.conf` | `/etc/tmpfiles.d/vyntrio.conf` |
| config template | `distro/install-media/config.toml.template` | `/etc/vyntrio/config.toml` |

Installation steps for systemd artifacts remain documented in
`distro/systemd/README.md` until an automated installer exists.

## Target disk paths (created at install, not stored on USB)

| Path | Purpose |
|------|---------|
| `/var/lib/vyntrio/` | Persistent state (`StateDirectory`; SQLite) |
| `/var/lib/vyntrio/backups/` | Root-only backup destination (future writes by `vyntrio-backup`) |
| `/etc/vyntrio/` | Runtime configuration directory |
| `/etc/vyntrio/config.toml` | From template; administrator-edited |

## Recovery media

Recovery boot USB/ISO is a **separate deliverable** per ADR-0006 section C,
scaffolded under `distro/recovery-media/` (Block 9, Slice 9.4). It is not part
of this install-media tree.

## Related

- `docs/ADR/0006-appliance-install-recovery-media.md`
- `docs/ADR/0005-appliance-runtime-operations.md` — sections D, E
- `docs/ADR/0004-identity-and-access.md` — post-install bootstrap
- `docs/ops/restore-safety-contract.md` — in-place restore (not install)
- `distro/recovery-media/README.md` — recovery media (separate deliverable)
- `distro/install-media/build-contract.md` — install-image build I/O contract
- `distro/systemd/README.md` — current manual install path
