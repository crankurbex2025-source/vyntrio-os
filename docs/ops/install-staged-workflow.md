# Installer staged workflow

**Secondary to USB-first delivery.** See `docs/00_PROJECT.md` (Primärer Auslieferungsweg)
and `docs/ADR/0006-appliance-install-recovery-media.md`. This workflow documents Block 10
**internal infrastructure** for install execution after bootable media exists.

Operator guide for the bounded Vyntrio installer pipeline (Block 10).

**Canonical policy:** `docs/ops/installer-policy.md` (command boundaries, preconditions,
sequence). Print CLI summary: `vyntrio-installer workflow`.

## Stages

| Stage | Command | Mutates? |
|-------|---------|----------|
| 1. Artifact verification | `vyntrio-verify-artifact` or release checks inside preflight/install | No |
| 2. Target/media preflight | `vyntrio-installer preflight` | No |
| 3. Sandbox payload install | `vyntrio-installer install --force` | Sandbox files only |
| 3b. Guarded target apply | `vyntrio-installer install --force --apply-target` | Delegates to partition apply |
| 3c. Partition apply | `vyntrio-installer apply --force` | Existing partition payloads only |
| 4. Postflight handover | Emitted by install; replay with `vyntrio-installer postflight` | No |

## Postflight handover (Slice 10.12)

After `install` (success or failure with a resolved sandbox path), the installer writes
`HANDOVER_RECORD.txt` beside `INSTALL_RECORD.txt` under the sandbox target tree.

### Record fields

- `overall_status`: `succeeded`, `failed`, or `preflight_only`
- `preflight_status`, `target_preflight`, `media_envelope`, `release_integrity`
- `install_write`: `not_run`, `succeeded`, `failed`, or `partial`
- `sandbox_target_path`, `payloads_copied`, `allowlist_complete`
- `mutation_scope`: sandbox-only, or guarded target apply when `--apply-target` used
- `apply_target`, `host_block_device_mutated`, `target_mutation_record` when applicable
- `deferred_items`: service enablement, bootstrap, partitioning, etc.
- `failure_stage` / `failure_reason` when applicable
- `next_steps`: operator guidance

### Replay handover

```bash
vyntrio-installer postflight --target-root distro/installer/target-sandbox/<disk-id>/
```

Exit 0 only when `overall_status` is not `failed`.

## Honest limits

- **Not** a full OS or hardware installation
- Sandbox copy (default) does **not** mutate block devices
- `--apply-target` writes six allowlisted payloads to one mounted partition only
- **Not** partitioning, formatting, or bootloader installation
- **Not** service start or bootstrap

## References

- `docs/ops/installer-policy.md`
- `docs/ops/install-target-preflight.md`
- `docs/ops/install-payload-write.md`
- `docs/ops/install-target-mutation.md`
- `docs/ops/install-partition-apply.md`
- `docs/ops/release-artifact-verification.md`
- `internal/platform/installpostflight/`
