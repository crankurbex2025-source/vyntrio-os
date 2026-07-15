# ADR-0006: Appliance install and recovery media model

- **Status:** Accepted
- **Date:** 2026-07-15
- **Deciders:** Vyntrio OS engineering (Architecture Block 9; Slice 9.2 acceptance)

## Context

Block 8 is complete: the authenticated overview is read-only and closed. Block 7
established the **runtime and state contract** (ADR-0005) and a **fail-closed
in-place restore safety contract** (`docs/ops/restore-safety-contract.md`) —
restore is not implemented. ADR-0005 explicitly **defers ISO/installer
packaging** (section A) and lists installer integration as a deferred restore
topic (section H.7).

Block 9, Slice 9.1 audited the missing **delivery-media** boundary: how a bare
machine becomes a Vyntrio appliance and how operators may recover it offline,
without expanding the running API, dashboard, or in-service restore.

Current repository truth:

- `cmd/installer/main.go` is an **empty stub** (Phase 0.2 placeholder).
- `distro/` contains **Block 7.3 systemd deployment artifacts** only
  (`distro/systemd/`); no ISO, initramfs, or image-build pipeline.
- Production install today is **manual** per `distro/systemd/README.md`
  (administrator installs binary, sysusers, tmpfiles, config, unit).
- Persistent application state, runtime configuration, and backup layout are
  **already fixed** by ADR-0005 section D; this ADR must not redefine them.

Without an explicit install/recovery media contract, future ISO, USB, and
installer work risks colliding with ADR-0004 bootstrap rules, ADR-0005 path
ownership, Block 7 restore semantics, and commercial license USB features
(`docs/15_LICENSE.md`).

## Decision

### A. Delivery-media scope (Block 9)

Block 9 governs **how Vyntrio is delivered to bare metal** and **how operators
may boot an offline recovery environment** — not how the appliance runs day to
day (ADR-0005) and not how in-place restore mutates live state
(`docs/ops/restore-safety-contract.md`).

Delivery media are **ephemeral carriers** of versioned, immutable install or
recovery payloads. They must **never** be the system of record for appliance
state, secrets, or operator credentials.

### B. Install USB / ISO — purpose

**Install media** (bootable USB image or ISO written to USB) exists for
**greenfield deployment** only:

1. Boot a minimal live environment on target hardware.
2. Lay down the **approved on-disk layout** (future slice; partition details
   deferred).
3. Install immutable artifacts: `vyntrio-api`, future `vyntrio-backup`,
   `distro/systemd/*` deployment files, and other approved binaries.
4. Provision the static `vyntrio` service account and directory ownership per
   ADR-0005 section D and `distro/systemd/README.md`.
5. Create `/etc/vyntrio/` and an administrator-editable `config.toml`
   **template** (not secrets, not Owner credentials).
6. Enable `vyntrio-api.service` and hand off to **first-boot product setup**.

Install media **does not**:

- create or seed the Owner account or SQLite application database;
- embed instance-specific secrets, bootstrap tokens, or license files;
- store persistent backup artifacts;
- expose install progress via the product API or dashboard;
- perform in-place restore from backup archives.

**First Owner creation** remains governed by ADR-0004: bootstrap endpoints
with loopback (or equivalent local-proof) guard after the installed service is
running. Installer integration with bootstrap transport is a **future slice**;
this ADR binds the separation only.

### C. Recovery media — purpose

**Recovery media** (optional, **separate** deliverable from install media) exists
for **offline operator procedures** when the installed system cannot safely
execute them — for example a future root-only restore CLI governed by
`docs/ops/restore-safety-contract.md`.

Recovery media may provide:

- a minimal boot environment with approved Vyntrio tooling versions;
- read-only access to attach and validate backup artifacts;
- a controlled path to invoke future restore commands **offline**.

Recovery media **does not**:

- replace ADR-0005 runtime operation or ADR-0004 session/bootstrap policy on a
  healthy running system;
- implement restore by itself (Block 7 restore contract governs mutation rules);
- auto-run restore, schedule jobs, or expose web/API/UI surfaces;
- carry appliance SQLite state or `/etc/vyntrio/config.toml` as authoritative
  copies (only operator-selected backup artifacts per restore contract).

Install and recovery media are **separate deliverables** in v1. A single
combined "wizard" image that mixes greenfield install with restore-from-backup
is **not permitted** without a future ADR amendment.

### D. Install vs recovery — explicit separation

| Dimension | **Install media** | **Recovery media** |
|-----------|-------------------|---------------------|
| Primary use | Greenfield / intentional reprovision | Offline repair on existing deployment |
| Target disk | Creates layout and installs artifacts | Does not redefine install layout contract |
| Service state | Brings system to first-boot / enabled unit | Assumes quiescent or broken installed system |
| Product DB | Empty until ADR-0004 bootstrap | Restored only via future Block 7 procedure |
| API / UI | No product API during install | No product API on recovery media |
| Block owner | Block 9 | Block 9 (media) + Block 7 (restore procedure) |

### E. Persistent state — location (unchanged)

ADR-0005 section D remains authoritative. Delivery media must respect:

| Path | Purpose | On install media? |
|------|---------|-------------------|
| `/var/lib/vyntrio/` | SQLite, backup sidecar, backup artifacts | **No** — created on target disk only |
| `/etc/vyntrio/config.toml` | Host-admin runtime config | **No** — template on disk at install; not on generic USB |
| `/usr/bin/vyntrio-api` | Immutable binary | Payload on media; installed to disk |
| Generic USB/ISO | Boot + payloads | **Ephemeral** — no persistent appliance state |

Install must **create** state directories and permissions but must **not**
pre-populate application database rows or Owner credentials.

### F. Relationship to ADR-0004 (identity and bootstrap)

- ADR-0004 governs **who** may create the first Owner and **how** bootstrap is
  network-bounded.
- Install media ends when the API service can start with valid
  `/etc/vyntrio/config.toml` and an empty application database.
- Bootstrap completes **after install**, via the running service — not embedded
  in the USB image as a baked-in account.
- Exact installer→bootstrap handoff mechanics (console URL, local tunnel
  instructions) are deferred to a future implementation slice.

### G. Relationship to ADR-0005 (runtime operations)

- ADR-0005 defines **installed appliance** behavior: config, state paths,
  systemd identity, backup destination, overview boundaries.
- This ADR defines **how artifacts reach the disk**; it does not alter runtime
  probes, overview slices, backup CLI semantics, or path validation rules.
- Future packaging/install slices must install binaries and units compatible with
  ADR-0005 layout without introducing alternate state roots or config sources.

### H. Relationship to restore-safety contract (Block 7)

- `docs/ops/restore-safety-contract.md` governs **in-place restore procedure**
  (root-only, offline, manifest-first, allowlisted targets) — **not implemented**.
- Block 7 restore v1 explicitly excludes boot media and installer integration
  from the **restore implementation** scope.
- ADR-0006 defines what **recovery boot media** is for; it does not implement
  restore or supersede the restore contract.
- When restore is implemented, recovery media may **host** the approved offline
  tooling; mutation rules remain those of the restore safety contract.

### I. License USB — separate concern

`docs/15_LICENSE.md` describes a commercial **USB license option** (hardware-bound
activation artifact). This is **not** install media and **not** recovery media.

| Media type | Block | Purpose |
|------------|-------|---------|
| Install USB/ISO | Block 9 | Greenfield deployment |
| Recovery USB/ISO | Block 9 | Offline operator recovery environment |
| License USB | Commercial / Phase 0.9 | Edition activation — separate ADR/track |

No Block 9 deliverable may require a license dongle for install or recovery.

### J. Explicit out of scope (Block 9 until dedicated slices)

The following remain **out of scope** for Block 9 Slice 9.2 and require future
approved slices:

- ISO/USB **build pipeline** execution and CI image publication
- Partitioning, UEFI/BIOS boot chain, disk encryption (LUKS), hardware detection
- `vyntrio-installer` implementation beyond the current stub
- First-boot graphical installer UI
- PXE/network install, ARM targets, multi-node imaging
- Restore CLI implementation (Block 7)
- Upgrade/rollback packaging (ADR-0005 section J)
- Dashboard/overview/API install or recovery status surfaces
- Automatic or unattended install/recovery
- Embedding secrets, license files, or Owner credentials in generic media

## Consequences

### Positive

- Install, recovery, runtime, and restore concerns have non-overlapping owners.
- Persistent state stays on the target disk per ADR-0005; USB remains ephemeral.
- License USB cannot be confused with boot media when both are documented.
- Future ISO/installer slices implement against one reviewed contract.

### Negative / trade-offs

- Separate install and recovery images increase packaging surface area.
- Manual install (`distro/systemd/README.md`) remains the only supported path
  until later Block 9 implementation slices.
- Bootstrap handoff after install is documented at boundary level only.

### Follow-up

- [x] Block 9 Slice 9.3: `distro/install-media/` scaffold (layout manifest, no execution)
- [x] Block 9 Slice 9.4: `distro/recovery-media/` scaffold (tooling manifest, no execution)
- [x] Block 9 Slice 9.5: install-image build contract (`distro/install-media/build-contract.md`)
- [x] Block 9 Slice 9.6: local install-media payload staging (`make install-media-stage`)
- [x] Block 9 Slice 9.7: install-image live/boot envelope scaffold (`envelope-contract.md`, `envelope-manifest.yaml`)
- [x] Block 9 Slice 9.8: local install-media envelope assembly (`make install-media-envelope`)
- [x] Block 9 Slice 9.9: install-image bootability initialization scaffold (`bootability-contract.md`, `bootability-manifest.yaml`)
- [ ] ISO/live-USB bootable image build slice (executable builder)
- [ ] `vyntrio-installer` implementation slice
- [ ] Install→bootstrap handoff slice (ADR-0004 integration)
- [ ] Recovery media image slice (tooling host for Block 7 restore when implemented)

## Alternatives considered

| Option | Why not chosen |
|--------|----------------|
| Extend ADR-0005 with install media sections | Mixes runtime operation with delivery; ADR-0005 already long and accepted |
| Single combined install+recovery USB image | Violates install/recovery separation; increases operator error risk |
| Install media seeds Owner account | Violates ADR-0004 bootstrap boundary and ships secrets on generic media |
| Store backups on install USB | Conflicts with ADR-0005 backup destination model |
| API-driven install from dashboard | Expands remote attack surface; deferred indefinitely |
| Define partition/LUKS layout in Slice 9.2 | Premature without hardware targets; deferred to implementation slice |

## References

- `docs/ADR/0004-identity-and-access.md` — bootstrap and loopback guard
- `docs/ADR/0005-appliance-runtime-operations.md` — runtime, state, backup layout
- `docs/ops/restore-safety-contract.md` — in-place restore procedure (not implemented)
- `docs/02_ARCHITECTURE.md` — Block 9 contract summary
- `distro/README.md` — `distro/` scope vs delivery media
- `distro/install-media/README.md` — install media scaffold
- `distro/install-media/build-contract.md` — install-image build I/O contract
- `distro/install-media/envelope-contract.md` — live/boot envelope contract (Slice 9.7)
- `distro/install-media/envelope-manifest.yaml` — envelope layer inventory (Slice 9.7)
- `distro/install-media/bootability-contract.md` — bootable initialization contract (Slice 9.9)
- `distro/install-media/bootability-manifest.yaml` — bootability layer inventory (Slice 9.9)
- `distro/recovery-media/README.md` — recovery media scaffold
- `distro/systemd/README.md` — current manual install path
- `docs/15_LICENSE.md` — license USB (separate from Block 9)
- `cmd/installer/main.go` — stub entrypoint (Phase 0.2 placeholder)
- Block 9, Slice 9.1 audit (conversation record)
