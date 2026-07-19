# Install target and media preflight

Operator contract for `vyntrio-installer preflight` (Block 10 / Slice 10.10).

## Purpose

Read-only validation that a **explicitly selected** install target disk and optional
install media context are suitable **before** any partition, format, write, or install
execution exists.

This answers: “Is this specific target + media combination acceptable for a future
install flow?” — not “Install succeeded” or “System is installable end-to-end.”

## Command

```bash
vyntrio-installer preflight --target-disk-id <opaque-id> [options]
```

### Required

| Flag | Description |
|------|-------------|
| `--target-disk-id` | Opaque disk ID (`disk-<sha256-prefix>`) from storage inventory |

Kernel names (`/dev/sdX`, `sda`, `nvme0n1`) are **not** accepted. Use
`GET /api/v1/storage/disks` or future listing tooling to obtain IDs.

### Optional

| Flag | Default | Description |
|------|---------|-------------|
| `--state-dir` | `/var/lib/vyntrio` | Discovery context for root/state disk exclusion |
| `--min-size-bytes` | `8589934592` (8 GiB) | Minimum target size |
| `--envelope-root` | — | Install-media envelope directory (`ENVELOPE.txt`, `payload/`) |
| `--release-manifest` | — | `vyntrio-release-manifest-v1` path for artifact integrity |
| `--artifact-base-dir` | manifest directory | Base path for release manifest artifacts |

### Exit codes

| Code | Meaning |
|------|---------|
| 0 | Selected checks passed |
| 1 | Target or media preflight failed |
| 2 | Usage error |

Stderr always notes: **no disk writes performed; install execution not authorized**.

## Target checks

Uses `internal/platform/storageinventory` discovery + classification, then
install-specific rules:

| Target status | Meaning |
|---------------|---------|
| `eligible` | Target passed all implemented checks |
| `excluded` | Known unsuitable (root disk, mounted, too small, etc.) |
| `unknown` | Ambiguous identity or discovery failure |
| `not_found` | No device with the given opaque ID |
| `ambiguous` | Multiple inventory entries for same ID (fail-closed) |

### Target rejection reasons

Includes storage inventory reasons (`root_disk`, `state_filesystem`, `mounted_in_use`,
`removable`, `read_only`, `install_media`, `virtual_device`, `unsupported_filesystem`,
`ambiguous_identity`) plus:

| Reason | Cause |
|--------|-------|
| `target_selection_required` | `--target-disk-id` missing |
| `target_not_found` | ID not in current inventory |
| `target_ambiguous` | Duplicate ID match |
| `insufficient_size` | Below `--min-size-bytes` |
| `discovery_unavailable` | Block device discovery failed |

**No auto-selection.** Omitting `--target-disk-id` always fails.

## Media checks (optional)

When `--envelope-root` is set, validates (mirrors `make installer-preflight`):

- Envelope directory exists
- `ENVELOPE.txt` with `media_role: install`
- `payload/` contains exactly six manifest files (no extras)
- No excluded patterns (`.db`, `.sqlite`, archives, license, bootstrap tokens, `vyntrio-restore`)

When `--release-manifest` is set, delegates to `vyntrio-verify-artifact` logic
(SHA-256 + size integrity). Authenticity is **not** verified unless signing is implemented.

## Non-goals

- Partitioning, formatting, flashing, payload copy to target
- Service enablement or bootstrap
- Operator reprovision confirmation UI
- Claiming full install readiness or hardware compatibility beyond listed checks
- Mount/unmount side effects

## References

- `internal/platform/installpreflight/`
- `docs/ADR/0007-appliance-installer-contract.md`
- `distro/installer/preflight-manifest.yaml`
- `docs/ops/release-artifact-verification.md`
