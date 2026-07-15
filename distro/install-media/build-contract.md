# Install-image build contract

- **Status:** Build contract only — **no image builder is implemented**
- **Block:** 9
- **Slice:** 9.5 (documentation)
- **Authority:** `docs/ADR/0006-appliance-install-recovery-media.md` (install media
  sections B, D, E). Payload inventory: `distro/install-media/manifest.yaml`.
  Runtime paths: `docs/ADR/0005-appliance-runtime-operations.md`.

## 1. Explicit status

**Slice 9.6 (implemented):** `make install-media-stage` runs
`scripts/stage-install-media.sh`, which validates required inputs and copies
manifest-listed payloads into the local disposable tree
`distro/install-media/staging/payload/`. This does **not** produce ISO/USB
images, boot layers, or target-disk writes.

**Slice 9.7 (scaffold):** `distro/install-media/envelope-contract.md` and
`envelope-manifest.yaml` define the live/boot envelope layers and how staged
`payload/` is wrapped into a future bootable image. This is declarative only.

There is **no** install-image envelope builder, ISO generator, USB writer, or CI
image publication job yet.

Recovery-image build is a **separate future contract** under
`distro/recovery-media/` — not in scope for this document.

## 2. Build purpose

The future **install-image build** assembles a bootable deliverable (ISO or
raw USB image) that carries:

1. A minimal live/boot layer (envelope scaffold in Slice 9.7; build deferred).
2. A **`payload/`** tree matching `distro/install-media/manifest.yaml` entries
   ready to be copied onto a **target disk** during greenfield install.
3. No appliance persistent state, secrets, or Owner credentials.

The build **does not** install to disk, partition, or start services.

## 3. Inputs (contract)

| Input | Source | Required when |
|-------|--------|----------------|
| Install payload manifest | `distro/install-media/manifest.yaml` | Always (declarative authority) |
| Config template | `distro/install-media/config.toml.template` | Always |
| API binary | `bin/vyntrio-api` from `make build` | Always |
| Backup binary | `bin/vyntrio-backup` from `make build` | Always |
| systemd unit | `distro/systemd/vyntrio-api.service` | Always |
| sysusers drop-in | `distro/systemd/vyntrio.sysusers` | Always |
| tmpfiles drop-in | `distro/systemd/vyntrio.tmpfiles.conf` | Always |
| Embedded UI | Inside `vyntrio-api` via `go:embed` | Satisfied by API binary build |
| Live/boot root filesystem | **Deferred** | Future slice before bootable ISO |

**Not inputs** to install-image build:

- `distro/recovery-media/manifest.yaml` or recovery tooling
- SQLite database, backup tarballs, or `backup-last-run.json`
- Owner/bootstrap credentials or license files
- Running appliance `/etc/vyntrio/config.toml` from a deployed host
- Source checkout at runtime on the target appliance

## 4. Outputs (contract)

Future install-image build produces **ephemeral distribution artifacts** only:

| Output | Description |
|--------|-------------|
| `payload/` tree | Staged copies of manifest `payloads.*` sources with paths suitable for install-time copy |
| Image envelope | Bootable ISO or USB image file (format deferred) |
| Build record | Version, commit, and manifest `schema_version` metadata (future) |

Conceptual on-media layout (aligns with `manifest.yaml` `media_layout`):

```
<install-image>/
├── boot/           # deferred — bootloader/live kernel
├── live_root/      # deferred — minimal runtime for installer/live env
└── payload/        # manifest payloads staged for target-disk copy
    ├── bin/
    ├── etc/
    └── usr/        # paths mirror install_path targets as relative staging
```

**Not outputs:**

- Populated `/var/lib/vyntrio/` on generic media
- Pre-seeded `vyntrio.db` or bootstrap state
- Combined install+recovery image
- Published artifacts to production hosts (CI publication is a later slice)

Persistent state is created on the **target disk** at install time only.

## 5. Included and excluded artifacts

### 5.1 Included (from manifest)

Only entries listed under `payloads` in `distro/install-media/manifest.yaml`:

- `vyntrio-api`, `vyntrio-backup`
- `distro/systemd/*` deployment files referenced there
- `config.toml.template` as install-time template (not live secrets)

### 5.2 Explicitly excluded (never staged into install image)

Matches `excluded_from_media` in `manifest.yaml` and ADR-0006 section B:

- `appliance_sqlite_state`
- `owner_credentials`
- `bootstrap_tokens`
- `license_files`
- `backup_artifacts`
- `instance_secrets`

Additionally excluded from **install-image** build inputs:

- Recovery-media tooling (`distro/recovery-media/`)
- `vyntrio-worker`, `vyntrio-installer`, `vyntrio-update-agent` (not in install manifest)
- Repository `frontend/dist` as a separate tree (UI is embedded in `vyntrio-api`)

## 6. Relationship to scaffolds

| Scaffold | Role in build flow |
|----------|-------------------|
| `distro/install-media/manifest.yaml` | **Payload authority** for install image `payload/` staging |
| `distro/install-media/envelope-manifest.yaml` | **Envelope layer authority** (boot/live_root/payload) — Slice 9.7 |
| `distro/install-media/envelope-contract.md` | Envelope boundary contract — Slice 9.7 |
| `distro/install-media/config.toml.template` | Shipped as template; installed to `/etc/vyntrio/config.toml` on target |
| `distro/install-media/README.md` | Human overview; not consumed by tooling |
| `distro/recovery-media/` | **Out of scope** — separate image build contract (future slice) |
| `distro/systemd/` | Source files referenced by manifest; installed path on target disk |

Until an executable builder lands, operators continue to use the manual path in
`distro/systemd/README.md`.

## 7. Install vs recovery build separation

| | Install-image build | Recovery-image build |
|---|---------------------|----------------------|
| Contract doc | This file | Future `distro/recovery-media/build-contract.md` |
| Manifest | `distro/install-media/manifest.yaml` | `distro/recovery-media/manifest.yaml` |
| Payload model | Copy-to-disk greenfield payloads | Live-environment tooling |
| Combined image | **Forbidden** per ADR-0006 |

## 8. Post-build / post-install handoff (unchanged)

Install-image build ends when a bootable artifact exists. **Install execution**
(deferred) copies `payload/` to target disk. **Bootstrap** (ADR-0004) runs only
after `vyntrio-api.service` starts on the installed host with an empty database.

## 9. Out of scope for Slice 9.5 and the first executable build slice

**Slice 9.5 (this document):** contract only — no builder.

**Deferred to later executable slices:**

- ISO/USB image envelope generation (`xorriso`, `grub`, `dd`, etc.)
- Live root filesystem composition (Debian/minimal base)
- Boot loader and UEFI/BIOS configuration
- Disk partitioning and filesystem creation
- `vyntrio-installer` execution
- Manifest YAML parser consumed by production tooling (optional in first builder)
- Recovery-image build flow
- Restore CLI packaging
- API/dashboard install status

## 10. Related documents

- `docs/ADR/0006-appliance-install-recovery-media.md`
- `docs/ADR/0005-appliance-runtime-operations.md`
- `docs/ADR/0004-identity-and-access.md`
- `docs/ops/restore-safety-contract.md`
- `distro/install-media/README.md`
- `distro/install-media/envelope-contract.md`
- `distro/install-media/envelope-manifest.yaml`
- `distro/recovery-media/README.md`
- `docs/19_RELEASE.md` — `make build` / binary deliverable
