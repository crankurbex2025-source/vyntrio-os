# Install target mutation (guarded apply)

Operator contract for `vyntrio-installer install --apply-target` (Block 10 / Slice 10.13).

## Purpose

Mount a **single eligible, unmounted partition** on an **explicitly selected** disk and copy the
**six supported install-media payload files** onto that filesystem. This is the first installer
path that may mutate a real block device — still **not** a full OS install.

Default `install` without `--apply-target` remains **sandbox-only**. The dedicated
`vyntrio-installer apply` subcommand (Slice 10.14) implements the same payload
application flow; see `docs/ops/install-partition-apply.md`.

## Command

```bash
vyntrio-installer install \
  --target-disk-id <opaque-id> \
  --force \
  --apply-target \
  [--envelope-root <path>] \
  [--release-manifest <path>] \
  [--artifact-base-dir <path>] \
  [--mount-root <path>] \
  [--state-dir <path>]
```

### Required gates (all)

| Gate | Flag / check |
|------|----------------|
| Explicit target | `--target-disk-id` (opaque ID from storage inventory) |
| Operator confirmation | `--force` |
| Real-target mode | `--apply-target` |
| Target preflight | Same rules as `vyntrio-installer preflight` |
| Verified artifact | `--envelope-root` and/or `--release-manifest` |

### Defaults

| Flag | Default |
|------|---------|
| `--mount-root` | `/run/vyntrio-install/mnt` |
| `--state-dir` | `/var/lib/vyntrio` |

## Pipeline

1. **Confirmation gate** — fails without `--force`
2. **Target preflight** — eligible disk, verified media
3. **Target resolve** — opaque ID → internal device (kernel names never exposed)
4. **Target validate** — fail closed on root, mounted, removable, virtual, ambiguous, ineligible
5. **Layout resolve** — exactly one supported unmounted partition/filesystem (`ext4`, `xfs`, `btrfs`)
6. **Mount** — temporary mount under `<mount-root>/<disk-id>/`
7. **Payload copy** — same six allowlisted files as sandbox install
8. **Post-write verification** — presence, optional SHA-256/size, inventory bounds
9. **INSTALL_RECORD.txt** on target filesystem (`host_block_device_mutated: true`)
10. **Unmount**
11. **TARGET_MUTATION_RECORD.txt** under `<state-dir>/install-target/<disk-id>/`
12. **HANDOVER_RECORD.txt** beside mutation record (operator audit trail)

## Rollback

On failure after mount:

- Copied payload files removed from mount point
- Temporary mount unmounted
- `TARGET_MUTATION_RECORD.txt` status `rolled_back` or `failed`
- Partial success is never hidden

No silent retries on destructive steps.

## Supported mutation scope

- One mounted partition with a supported existing filesystem
- Six allowlisted payload paths under the mount root
- State directories: `etc/vyntrio`, `var/lib/vyntrio`, `var/lib/vyntrio/backups`

## Failure reasons

| Reason | Cause |
|--------|-------|
| `force_required` | Missing `--force` |
| `target_mounted` | Target disk or candidate partition mounted |
| `target_not_eligible` | Root, removable, virtual, in-use, or excluded |
| `ambiguous_target_layout` | Zero or multiple supported mount candidates |
| `unsupported_target_state` | No filesystem, blank disk, discovery failure |
| `mount_failed` | `mount(8)` failed |
| `rollback_failed` | Unmount after failed apply failed |
| `preflight_failed` | Target or media preflight failed |
| `post_verify_failed` | Copied tree failed verification |

## Non-goals

- Partitioning, formatting, or bootloader installation
- Blank-disk install (no filesystem → fail closed)
- Raw block writes without mount
- Auto target selection
- Changing block-device ownership as a safety shortcut
- Full appliance/OS install completeness claim
- Service enablement or bootstrap

## References

- `internal/platform/installtarget/`
- `internal/platform/installwrite/`
- `docs/ops/install-payload-write.md`
- `docs/ops/install-target-preflight.md`
- `docs/ops/install-staged-workflow.md`
