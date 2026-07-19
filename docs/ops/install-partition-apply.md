# Install partition apply (existing filesystem)

Operator contract for `vyntrio-installer apply` (Block 10 / Slice 10.14).

## Purpose

Apply the **six supported install-media payload files** to **one existing, supported,
unmounted partition** on an explicitly selected eligible disk. This is the dedicated
install-application step after target guards — **not** a full OS deployment.

`install --apply-target` delegates to this same flow. Default `install` without
`--apply-target` remains **sandbox-only** and never touches real partitions.

## Command

```bash
vyntrio-installer apply \
  --target-disk-id <opaque-id> \
  --force \
  [--envelope-root <path>] \
  [--release-manifest <path>] \
  [--artifact-base-dir <path>] \
  [--mount-root <path>] \
  [--state-dir <path>]
```

### Required gates (all)

| Gate | Flag / check |
|------|----------------|
| Explicit target | `--target-disk-id` |
| Operator confirmation | `--force` |
| Target preflight | Same rules as `vyntrio-installer preflight` |
| Verified artifact | `--envelope-root` and/or `--release-manifest` |
| Existing partition | Exactly one supported unmounted partition (`ext4`, `xfs`, `btrfs`) |

There is **no** sandbox fallback for `apply`.

## Pipeline

1. Confirmation gate (`--force`)
2. Target + media preflight
3. Target validate (eligible, unmounted, non-root, non-ambiguous)
4. Partition resolve (single supported existing filesystem)
5. Mount at `<mount-root>/<disk-id>/`
6. Copy six allowlisted payloads + state directories
7. Post-write verification
8. `INSTALL_RECORD.txt` on target (`host_block_device_mutated: true`)
9. Unmount
10. `PARTITION_APPLY_RECORD.txt` at `<state-dir>/install-apply/<disk-id>/`
11. `TARGET_MUTATION_RECORD.txt` at `<state-dir>/install-target/<disk-id>/`
12. `HANDOVER_RECORD.txt` beside apply record

## Rollback

On failure after mount:

- Remove copied payload files from mount point
- Unmount temporary mount
- Write `rolled_back` or `failed` to apply and mutation records
- Partial copy counts preserved in records

## Supported scope

- Existing partition with supported filesystem only
- Six allowlisted payload paths
- State dirs: `etc/vyntrio`, `var/lib/vyntrio`, `var/lib/vyntrio/backups`

## Non-goals

- Partition creation or formatting
- Blank disk handling
- Bootloader / full OS install claims
- Auto target selection
- Raw block writes
- Service enablement or bootstrap

## References

- `internal/platform/installapply/`
- `internal/platform/installtarget/`
- `docs/ops/install-target-mutation.md`
- `docs/ops/install-payload-write.md`
