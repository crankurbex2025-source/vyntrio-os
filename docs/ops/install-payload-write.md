# Install payload write

Operator contract for `vyntrio-installer install` (Block 10 / Slice 10.11).

## Purpose

Copy the **six supported install-media payload files** to a **sandbox target tree**
after target preflight and artifact verification succeed. This is the first bounded
mutation step in the installer pipeline — not a full OS install.

## Command

```bash
vyntrio-installer install \
  --target-disk-id <opaque-id> \
  --force \
  [--envelope-root <path>] \
  [--release-manifest <path>] \
  [--artifact-base-dir <path>] \
  [--sandbox-root <path>] \
  [--target-root <path>] \
  [--state-dir <path>]
```

### Required

| Flag | Description |
|------|-------------|
| `--target-disk-id` | Opaque disk ID from storage inventory |
| `--force` | Explicit operator authorization for writes |
| `--envelope-root` **or** `--release-manifest` | Artifact source (one or both) |

### Defaults

| Flag | Default |
|------|---------|
| `--sandbox-root` | `distro/installer/target-sandbox` |
| `--target-root` | `<sandbox-root>/<disk-id>/` |
| `--state-dir` | `/var/lib/vyntrio` |

## Pipeline

1. **Confirmation gate** — fails without `--force`
2. **Target preflight** — same rules as `vyntrio-installer preflight`
3. **Media verification** — envelope inventory and/or release manifest SHA-256
4. **Sandbox path resolution** — writes only under `--sandbox-root`
5. **State directory creation** — `etc/vyntrio`, `var/lib/vyntrio`, `var/lib/vyntrio/backups`
6. **Payload copy** — six allowlisted files only
7. **Post-write verification** — presence, optional SHA-256/size from manifest, no forbidden DB files
8. **Record** — `INSTALL_RECORD.txt` in target tree

## Supported payloads

Copied paths (relative to target root):

- `usr/bin/vyntrio-api`
- `usr/bin/vyntrio-backup`
- `etc/systemd/system/vyntrio-api.service`
- `usr/lib/sysusers.d/vyntrio.conf`
- `etc/tmpfiles.d/vyntrio.conf`
- `etc/vyntrio/config.toml`

## Success output

```
installer install succeeded: disk_id=... target_root=... payloads=6 release=...
```

Stderr notes: sandbox-only writes; host block devices not mutated; service enablement and bootstrap deferred.

## Failure reasons

| Reason | Cause |
|--------|-------|
| `force_required` | Missing `--force` |
| `artifact_source_required` | No envelope or release manifest |
| `unsafe_target_root` | Target path escapes sandbox |
| `preflight_failed` | Target or media preflight failed |
| `post_verify_failed` | Copied tree failed verification |

## Non-goals

- Partitioning, formatting, or raw block device writes
- Service enablement, sysusers/tmpfiles application, bootstrap
- Full appliance/OS install completeness claim
- Auto target selection
- ISO/USB flashing

Default `install` without `--apply-target` remains **sandbox-only**. For guarded
real-target mutation see `docs/ops/install-target-mutation.md`. For the dedicated
partition apply subcommand see `docs/ops/install-partition-apply.md`.

## References

- `internal/platform/installwrite/`
- `docs/ops/install-target-preflight.md`
- `docs/ops/install-target-mutation.md`
- `docs/ops/release-artifact-verification.md`
- `docs/ADR/0007-appliance-installer-contract.md`
