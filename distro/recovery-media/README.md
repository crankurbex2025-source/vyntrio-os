# Recovery media scaffold (Block 9, Slice 9.4)

Declarative scaffold for **offline recovery USB/ISO** payloads. This directory
describes what recovery media **is for** and which tooling it may carry in a
future live boot environment. It does **not** build bootable images, implement
restore, or mutate appliance state.

**Authority:** `docs/ADR/0006-appliance-install-recovery-media.md` (recovery media
sections C, D, H). Restore mutation rules remain
`docs/ops/restore-safety-contract.md` (procedure **not implemented**).

## Status

| Item | State |
|------|--------|
| Layout manifest (`manifest.yaml`) | **Scaffold** — declarative only |
| Boot/live recovery environment | **Not started** — future slice |
| ISO/USB build pipeline | **Not started** — future slice |
| `vyntrio-restore` CLI | **Implemented** — `cmd/restore/` |
| Install media | **Separate deliverable** — `distro/install-media/` |

## What recovery media is for

1. Boot a minimal **offline** live environment (future).
2. Provide approved Vyntrio **recovery tooling** versions (future `vyntrio-restore`).
3. Allow a **root operator** to read and validate backup artifacts on the **target
   disk** (`/var/lib/vyntrio/backups/`) per the restore safety contract.
4. Offer a controlled path to invoke future restore commands — **offline only**.

Recovery media hosts **tools**, not a fresh appliance install.

## What recovery media is not for

- Greenfield installation (use `distro/install-media/` — separate image)
- Shipping appliance SQLite state, Owner credentials, or license files
- Storing backup archives as the system of record (backups remain on **target disk**)
- Authoritative copies of `/etc/vyntrio/config.toml` from generic media
- Automatic or unattended restore
- Product API, dashboard, or restore UI
- Replacing ADR-0005 runtime operation on a healthy running system

## Install vs recovery (explicit separation)

| | **Install media** (`distro/install-media/`) | **Recovery media** (this tree) |
|---|---------------------------------------------|--------------------------------|
| Primary use | Greenfield deploy | Offline repair / restore host |
| Payload model | Copy artifacts **to** target disk | Run tooling **from** live env |
| Typical binaries | `vyntrio-api`, `vyntrio-backup`, systemd units | Future `vyntrio-restore` only |
| Config template | Yes (`config.toml.template`) | **No** — config restored only via future Block 7 procedure |
| Product DB | Empty until bootstrap | Restored only via future restore CLI |

A single combined install+recovery image is **not permitted** per ADR-0006.

## Directory layout (repository scaffold)

```
distro/
├── install-media/          ← greenfield install payload contract (Slice 9.3)
├── recovery-media/         ← this scaffold (offline recovery tooling contract)
│   ├── README.md
│   └── manifest.yaml
├── systemd/                ← runtime deployment artifacts on installed disk (Block 7.3)
└── README.md
```

## Target disk interaction (reference only)

Recovery tooling may **read** completed backup artifacts from the installed
appliance backup destination when a future restore CLI exists:

| Path | Recovery interaction |
|------|----------------------|
| `/var/lib/vyntrio/backups/` | Read-only artifact selection (`vyntrio-backup-v1_<UTC>.tar`) |
| `/var/lib/vyntrio/` | Mutable only when future restore procedure runs (Block 7 contract) |
| `/etc/vyntrio/config.toml` | Mutable only when policy-controlled restore includes config |

Persistent state stays on the **target disk**. Generic recovery USB remains
ephemeral and must not embed live database files or backup tarballs.

## Related

- `docs/ADR/0006-appliance-install-recovery-media.md`
- `docs/ops/restore-safety-contract.md` — restore procedure (not implemented)
- `docs/ADR/0005-appliance-runtime-operations.md` — backup destination layout
- `distro/install-media/README.md` — install media (separate deliverable)
- `distro/install-media/README.md` — install media (separate deliverable)
- `distro/systemd/README.md` — installed runtime artifacts (not recovery live root)
