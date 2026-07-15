# ADR-0007: Appliance installer contract

- **Status:** Accepted
- **Date:** 2026-07-15
- **Deciders:** Vyntrio OS engineering (Architecture Block 10; Slice 10.2 acceptance)

## Context

Block 9 is complete through bootability initialization: install-media payload
staging, envelope assembly, and bootable-init contracts exist. Block 9 governs
**media creation** — how immutable payloads are packaged into a future bootable
install image. It does **not** define install **execution** on target hardware.

Block 10, Slice 10.1 audited the installer boundary. The separation is already
implied by ADR-0004 (bootstrap), ADR-0005 (runtime paths and persistent state),
ADR-0006 (install vs recovery media), the restore-safety contract, and Block 9
media contracts — but no formal **installer contract** existed.

Current repository truth:

- `cmd/installer/main.go` is an **empty stub** (Phase 0.2 placeholder).
- Supported production install today is **manual** per `distro/systemd/README.md`.
- Install payloads are declared in `distro/install-media/manifest.yaml`; media
  build prep exists locally (`make install-media-stage`, `make install-media-envelope`).
- No installer preflight, disk layout, or target-disk mutation code exists.

Without an explicit installer contract, future `vyntrio-installer` work risks
mixing media creation (Block 9), recovery procedures (Block 7), bootstrap
(ADR-0004), and host-side USB/ISO preparation into one undifferentiated flow.

## Decision

### A. Installer scope (Block 10)

Block 10 governs **greenfield install execution** on target hardware: what
`vyntrio-installer` (future) may do **after** bootable install media is available
and the operator has booted the install environment.

The installer operates on the **target disk** only. It does **not** build,
stage, or write install images to removable media.

### B. Installer responsibilities (future implementation)

When implemented, the installer is responsible for:

1. **Preflight** — verify install context before any target-disk mutation
   (section D).
2. **Target-disk layout** — create approved on-disk layout (partition and
   filesystem details deferred to a later implementation slice).
3. **Payload installation** — copy manifest-listed artifacts from install-media
   `payload/` to target-disk paths per `distro/install-media/manifest.yaml`.
4. **Account and directory provisioning** — apply `distro/systemd/vyntrio.sysusers`
   and `distro/systemd/vyntrio.tmpfiles.conf` effects on the target system.
5. **Configuration template** — install `config.toml` from the shipped template
   to `/etc/vyntrio/config.toml` (administrator-editable; no secrets).
6. **Service enablement** — install and enable `vyntrio-api.service` on the
   target system.
7. **State directory creation** — create empty `/var/lib/vyntrio/` layout per
   ADR-0005 section D; **no** pre-populated SQLite database or Owner rows.
8. **Post-install handoff** — end with a runnable, empty-database appliance ready
   for ADR-0004 bootstrap (section F).

### C. Installer non-responsibilities (explicit exclusions)

The installer **does not**:

| Excluded concern | Owner / authority |
|------------------|-------------------|
| Payload staging, envelope assembly, bootability, ISO/USB emission | **Block 9** — `distro/install-media/*` |
| Bootloader, kernel, initrd, live-root build | Block 9 bootability slices |
| Writing install images from operator workstations | **Future host media creator** (macOS/Windows — section J) |
| Recovery, in-place restore, backup mutation | **Block 7** — `docs/ops/restore-safety-contract.md` |
| Recovery-media tooling or combined install+recovery flows | ADR-0006 section C, D |
| Owner account creation or bootstrap token issuance | **ADR-0004** — post-install on running service |
| Embedding secrets, license files, or SQLite state on generic media | ADR-0006 section B, E |
| Product API, dashboard, or install-progress UI | ADR-0006 — deferred indefinitely |
| Network configuration, backup jobs, or runtime service operation | ADR-0005 |

### D. Required preflight checks (contract)

Before any target-disk mutation, a future installer **must** verify:

| Check | Requirement |
|-------|-------------|
| Media role | Booted from **install media** — not recovery media, not license USB |
| Payload presence | All manifest-listed files present under install-media `payload/` |
| Payload inventory | No extra files beyond `distro/install-media/manifest.yaml` entries |
| Media exclusions | No appliance SQLite, Owner credentials, bootstrap tokens, license files, backup archives, or instance secrets on media |
| Target disk | Target block device accessible; operator intent confirmed (greenfield or explicit reprovision — mechanism deferred) |
| Recovery separation | Installer flow is not invoked from recovery-media boot context |
| Bootstrap boundary | Installer does not create Owner account or call bootstrap endpoints |

Preflight failure **must** abort before target-disk mutation (fail-closed).

### E. Target-disk mutation allowlist

Future installer may mutate **only** these target paths (plus layout creation
deferred to a partitioning slice):

| Target path | Action | Authority |
|-------------|--------|-----------|
| `/usr/bin/vyntrio-api` | Install binary from payload | `manifest.yaml` payloads.binaries |
| `/usr/bin/vyntrio-backup` | Install binary from payload | `manifest.yaml` payloads.binaries |
| `/etc/systemd/system/vyntrio-api.service` | Install unit file | `manifest.yaml` payloads.systemd |
| `/usr/lib/sysusers.d/vyntrio.conf` | Install sysusers drop-in | `manifest.yaml` payloads.systemd |
| `/etc/tmpfiles.d/vyntrio.conf` | Install tmpfiles drop-in | `manifest.yaml` payloads.systemd |
| `/etc/vyntrio/config.toml` | Install from template (not secrets) | `manifest.yaml` payloads.config_templates |
| `/etc/vyntrio/` | Create directory layout (empty) | ADR-0005 section D |
| `/var/lib/vyntrio/` | Create empty state root (no DB rows) | ADR-0005 section D |
| Partition/filesystem layout | Create approved layout | Deferred implementation slice |

**Forbidden mutations:**

- Paths outside the allowlist without a future ADR amendment
- Pre-seeding `vyntrio.db` or bootstrap settings
- Writing Owner credentials or license material
- Modifying non-Vyntrio system packages, firewall, TLS, or DNS (installer scope is Vyntrio artifacts only)

### F. Post-install handoff to ADR-0004 bootstrap

Install **ends** when:

- `vyntrio-api.service` is enabled and can start on the target system
- `/etc/vyntrio/config.toml` exists from the template (administrator may edit before first start)
- `/var/lib/vyntrio/` exists with **empty** application database (zero active users)

**Bootstrap begins after install**, on the **running installed service**:

- First Owner creation follows ADR-0004 bootstrap rules (loopback or equivalent local-proof guard)
- Installer **must not** embed bootstrap tokens, pre-created Owner accounts, or LAN-accessible bootstrap shortcuts on generic media
- Exact operator handoff mechanics (console URL, SSH tunnel instructions) are deferred to a future implementation slice

### G. What must stay off installer media

Matches `excluded_from_media` in `distro/install-media/manifest.yaml` and
ADR-0006 section E:

- `appliance_sqlite_state`
- `owner_credentials`
- `bootstrap_tokens`
- `license_files`
- `backup_artifacts`
- `instance_secrets`

Additionally excluded from generic install media:

- Authoritative `/etc/vyntrio/config.toml` copied from a deployed host
- Recovery-media tooling (`distro/recovery-media/`)
- Combined install+recovery payloads

Persistent state is created on the **target disk** at install time only.

### H. Relationship to Block 9 media creation

| Phase | Block | Delivers |
|-------|-------|----------|
| Payload staging | 9.6 | `distro/install-media/staging/payload/` (local build prep) |
| Envelope assembly | 9.8 | `distro/install-media/envelope/` (local structure) |
| Bootability init | 9.9+ | Future bootable ISO/USB image |
| **Install execution** | **10** | Target-disk mutation from booted install media |

```
Block 9 (media)                         Block 10 (installer)
────────────────                        ────────────────────
manifest → staging → envelope      →    reads payload/ from booted install media
→ bootability → bootable image     →    writes allowlisted paths on TARGET DISK
                                        enables service → ADR-0004 bootstrap
```

Media creation **precedes** installer execution. The installer **consumes** install
media; it does not produce it.

### I. Relationship to recovery media and restore

- **Recovery media** (ADR-0006 section C) is a **separate deliverable** for
  offline repair — not an install path.
- **Restore procedure** (`docs/ops/restore-safety-contract.md`) is root-only,
  offline, in-place mutation on an **existing** deployment — not greenfield install.
- The installer **must not** invoke restore commands or package recovery tooling.
- Recovery media **must not** be used as install media.

### J. Future host media creator (macOS/Windows)

Operators on macOS or Windows will eventually use a **separate host-side media
creator** (not `vyntrio-installer`) to:

- obtain or build an install image artifact
- write it to USB or attach it for VM boot

That tool is a **prerequisite flow** only. It does **not** partition target
disks, install payloads, enable services, or perform bootstrap. It is out of
scope for Block 10 installer implementation.

### K. Explicit out of scope (Block 10 until dedicated slices)

- `vyntrio-installer` implementation beyond the current stub
- Partitioning, UEFI/BIOS boot chain, disk encryption (LUKS)
- Graphical or API-driven install UI
- Install progress reporting via product API or dashboard
- Host media creator for macOS/Windows
- Recovery-media installer integration
- Install→bootstrap transport mechanics (deferred handoff slice)

## Consequences

### Positive

- Clear separation between media creation (Block 9) and install execution (Block 10)
- Preflight and allowlist give future implementers fail-closed guardrails
- Bootstrap boundary preserved per ADR-0004
- Recovery and install paths cannot be conflated without an ADR amendment

### Negative / trade-offs

- Manual install (`distro/systemd/README.md`) remains the only supported path until implementation slices land
- Partition/LUKS details still deferred — installer contract does not fix hardware targets
- Host media creator is documented but not specified — separate future work

### Follow-up

- [x] Block 10 Slice 10.1: read-only installer contract audit
- [x] Block 10 Slice 10.2: formal installer contract (this ADR)
- [x] Block 10 Slice 10.3: installer preflight scaffold (`make installer-preflight`)
- [x] Block 10 Slice 10.4: installer target-layout scaffold (`target-layout-manifest.yaml`)
- [x] Block 10 Slice 10.5: preflight-gated mutation dry-run stub (`make installer-mutation-stub`)
- [x] Block 10 Slice 10.6: first target mutation — empty state directories in `target-sandbox/`
- [x] Block 10 Slice 10.7: manifest payload copy to `target-sandbox/`
- [ ] Partition/filesystem executable slice
- [ ] `vyntrio-installer` target-disk mutation slice
- [ ] Install→bootstrap handoff slice (ADR-0004 integration)
- [ ] Host media creator specification (macOS/Windows — separate block or slice)

## Alternatives considered

| Option | Why not chosen |
|--------|----------------|
| Extend ADR-0006 with installer sections | Mixes media delivery with target-disk execution; ADR-0006 is Block 9 scoped |
| Fold installer into Block 9 media build | Violates media vs install separation audited in Slice 10.1 |
| Installer creates Owner during install | Violates ADR-0004 bootstrap boundary and ships secrets risk |
| Single tool for USB writing and disk install | Conflates host media prep with appliance install execution |
| Defer formal contract until implementation | Risks inconsistent preflight and allowlist across slices |

## References

- `docs/ADR/0004-identity-and-access.md` — bootstrap and loopback guard
- `docs/ADR/0005-appliance-runtime-operations.md` — runtime paths and state layout
- `docs/ADR/0006-appliance-install-recovery-media.md` — install and recovery media model
- `docs/ops/restore-safety-contract.md` — in-place restore (not installer)
- `distro/install-media/manifest.yaml` — payload inventory and target paths
- `distro/install-media/build-contract.md` — install-image build I/O (Block 9)
- `distro/install-media/bootability-contract.md` — bootability vs installation boundary
- `distro/recovery-media/README.md` — separate recovery deliverable
- `distro/systemd/README.md` — current manual install path
- `docs/15_LICENSE.md` — license USB (separate from install)
- `cmd/installer/main.go` — stub entrypoint
- `distro/installer/preflight-manifest.yaml` — preflight check inventory (Slice 10.3)
- `distro/installer/target-layout-manifest.yaml` — target layout inventory (Slice 10.4)
- `distro/installer/target-layout-contract.md` — layout planning boundary (Slice 10.4)
- `scripts/installer-preflight.sh` — read-only preflight (Slice 10.3)
- `scripts/validate-installer-layout-plan.sh` — read-only layout plan validation (Slice 10.4)
- `scripts/installer-mutation-stub.sh` — preflight-gated mutation dry-run stub (Slice 10.5)
- `scripts/installer-mutate-directories.sh` — empty state directory mutation in target-sandbox (Slice 10.6)
- `scripts/installer-copy-payloads.sh` — manifest payload copy to target-sandbox (Slice 10.7)
- Block 10, Slice 10.1 audit (conversation record)
