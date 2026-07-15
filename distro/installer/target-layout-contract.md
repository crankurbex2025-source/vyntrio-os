# Installer target-disk layout contract

- **Status:** Layout scaffold only — **no disk mutation is implemented**
- **Block:** 10
- **Slice:** 10.4 (documentation)
- **Authority:** `docs/ADR/0007-appliance-installer-contract.md` (section E).
  Runtime paths: `docs/ADR/0005-appliance-runtime-operations.md` (section D).
  Layout inventory: `distro/installer/target-layout-manifest.yaml`.

## 1. Explicit status

**Slice 10.4 (this document):** defines the **target-disk layout planning
boundary** — what paths and directories a future installer may create on the
target system. This is **declarative only**. No partitioning, formatting, disk
writes, payload copy, service enablement, or bootstrap runs in this slice.

**Slice 10.3 (implemented):** `make installer-preflight` validates install-media
context and payload inventory only — no target-disk access.

Layout planning **does not** execute install steps. It documents intent for
future executable slices.

## 2. Layout purpose

Greenfield install requires an approved on-disk layout before manifest payloads
can be copied to the target system. This contract specifies:

| Category | Intent |
|----------|--------|
| **Partition layout** | Conceptual OS root mount — details **deferred** |
| **Directories** | Empty `/etc/vyntrio/`, `/var/lib/vyntrio/`, backups subdir |
| **Payload paths** | Six manifest files → target paths per ADR-0007 §E |
| **Provisioning** | systemd/sysusers/tmpfiles — application **deferred** |

Persistent state is created on the **target disk** at install time only. The
installer must **not** pre-seed SQLite, Owner credentials, or backup archives.

## 3. Relationship to preflight and media creation

```
Block 9 (media)          Slice 10.3 (preflight)       Slice 10.4 (this contract)
───────────────          ──────────────────────       ───────────────────────────
envelope/payload/   →    validates media context  →   defines TARGET path plan
                         (read-only)                  (read-only, no disk I/O)
                                                      ↓
                                              future: partition + copy (deferred)
```

| Phase | Mutates target disk? |
|-------|----------------------|
| Block 9 media creation | **No** |
| Slice 10.3 preflight | **No** |
| Slice 10.4 layout plan | **No** |
| Future partition/copy slice | **Yes** (allowlist only) |

## 4. Intended target paths (allowlist)

Matches ADR-0007 section E and `target-layout-manifest.yaml`:

| Target path | Action (future) | Initial state |
|-------------|-----------------|---------------|
| `/usr/bin/vyntrio-api` | Install from payload | Binary |
| `/usr/bin/vyntrio-backup` | Install from payload | Binary |
| `/etc/systemd/system/vyntrio-api.service` | Install unit | Unit file |
| `/usr/lib/sysusers.d/vyntrio.conf` | Install drop-in | Config |
| `/etc/tmpfiles.d/vyntrio.conf` | Install drop-in | Config |
| `/etc/vyntrio/config.toml` | Install template | No secrets |
| `/etc/vyntrio/` | Create directory | Empty except template |
| `/var/lib/vyntrio/` | Create directory | **Empty** — no DB |
| `/var/lib/vyntrio/backups/` | Create directory | Empty |

**Forbidden:** paths outside this allowlist, pre-seeded `vyntrio.db`, bootstrap
tokens, license files, recovery tooling, non-Vyntrio package changes.

## 5. Partition layout (deferred)

Partition table design is **explicitly deferred**:

- UEFI/BIOS boot partition sizing
- Filesystem types (ext4, xfs, etc.)
- LUKS/disk encryption
- Separate `/var` or `/home` volumes

Conceptual requirement: a mount at `/` sufficient to host the allowlisted paths.
Hardware-specific partition contracts require a future implementation slice and
may need ADR amendment.

## 6. Install vs recovery vs restore

| Concern | Target layout (this contract) | Recovery / restore |
|---------|------------------------------|-------------------|
| Use case | Greenfield install | Offline repair on existing system |
| Disk model | Create layout + copy payloads | No greenfield layout redefinition |
| State on disk | Empty until bootstrap | Restore contract governs mutation |
| Authority | ADR-0007, this manifest | restore-safety contract |

## 7. Deferred executable slices

- Partition table creation and filesystem formatting
- Target block-device accessibility checks (beyond planning)
- Payload copy from install media to target paths
- sysusers/tmpfiles application on target
- `vyntrio-api.service` enablement
- ADR-0004 bootstrap handoff

## 8. Out of scope (Slice 10.4)

- Partitioning, formatting, or any disk writes
- Payload copy to target disk
- Service enablement or config mutation beyond planning
- Recovery-media or restore execution
- Bootloader or Block 9 media build logic
- Host paths, services, secrets, or backup artifacts at planning time

## 9. Related documents

- `distro/installer/target-layout-manifest.yaml` — declarative layout inventory
- `distro/installer/preflight-manifest.yaml` — preflight checks (Slice 10.3)
- `docs/ADR/0007-appliance-installer-contract.md`
- `docs/ADR/0005-appliance-runtime-operations.md`
- `docs/ADR/0006-appliance-install-recovery-media.md`
- `distro/install-media/manifest.yaml`
- `docs/ops/restore-safety-contract.md`
