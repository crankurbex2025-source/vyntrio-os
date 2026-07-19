# Installer policy (staged pipeline)

**Secondary to USB-first delivery.** The operator product journey is: USB creator →
bootable install medium → boot on target hardware → local browser dashboard → install
from live session. This document governs the **Block 10 internal installer
infrastructure** (`vyntrio-installer`, `vyntrio-verify-artifact`) used **after** boot
and inside the live install environment — not the primary delivery path.

Canonical operator policy for the Vyntrio installer/OS-delivery pipeline (Block 10 /
Slice 10.15). **Policy version:** `vyntrio-installer-policy-v1`.

The installer is a **staged pipeline**, not one monolithic command. Each stage has
explicit boundaries, preconditions, and honest non-goals.

## Command map

| Order | Command | Binary | Mutates? | Purpose |
|-------|---------|--------|----------|---------|
| 1 (optional) | `verify-artifact` | `vyntrio-verify-artifact` | No | Manifest SHA-256 / size integrity |
| 2 (recommended) | `preflight` | `vyntrio-installer` | No | Target eligibility + media checks |
| 3a (default write) | `install` | `vyntrio-installer` | Sandbox only | Six payloads → `target-sandbox/` (**dev/lab; not primary product path**) |
| 3b (explicit real target) | `apply` | `vyntrio-installer` | Partition only | Six payloads → existing partition (**live-session infra; not USB delivery**) |
| 4 (replay) | `postflight` | `vyntrio-installer` | No | Display `HANDOVER_RECORD.txt` |

Print the CLI policy summary:

```bash
vyntrio-installer workflow
```

## Sequence

```mermaid
sequenceDiagram
  participant Op as Operator
  participant VA as vyntrio-verify-artifact
  participant PF as installer preflight
  participant IN as installer install
  participant AP as installer apply
  participant PO as installer postflight

  Op->>VA: manifest.json (optional)
  VA-->>Op: integrity ok (read-only)
  Op->>PF: --target-disk-id (required)
  PF-->>Op: target eligible (read-only)
  Note over Op: Primary path is USB creator → boot → dashboard (Block 9+); diagram is Block 10 infra only
  alt Sandbox staging (dev/lab only)
    Op->>IN: install --force
    IN-->>Op: HANDOVER_RECORD in sandbox
    Op->>PO: postflight --target-root sandbox/...
  else Partition apply (live-session/lab infra)
    Op->>AP: apply --force
    AP-->>Op: HANDOVER_RECORD in install-apply/
    Op->>PO: postflight --target-root install-apply/...
  end
```

## Per-command policy

### 1. `vyntrio-verify-artifact`

**Preconditions:** local manifest path; artifact files present (under `--base-dir` or
manifest directory).

**Does:** structure parse, presence, size, SHA-256.

**Does not:** write disks, mount, install payloads, select targets, replace preflight.

**Not verification of:** Ed25519 signatures (reserved, not implemented).

### 2. `vyntrio-installer preflight`

**Preconditions:** `--target-disk-id` (opaque, from `GET /api/v1/storage/disks`).

**Does:** read-only target eligibility, optional envelope/manifest checks.

**Does not:** write, mount, auto-select target, partition, format.

**Fail closed:** root, mounted, removable, virtual, ambiguous, or ineligible targets.

### 3. `vyntrio-installer install` (default)

**Preconditions:** `--force`, `--target-disk-id`, artifact source
(`--envelope-root` and/or `--release-manifest`). Preflight strongly recommended.

**Does (default):** copy six allowlisted payloads to
`distro/installer/target-sandbox/<disk-id>/`; write `INSTALL_RECORD.txt` and
`HANDOVER_RECORD.txt`.

**Does not (default):** mutate block devices, enable services, bootstrap, claim full
OS install.

**`--apply-target`:** delegates to partition apply (same as `apply` subcommand). Prefer
`apply` for clarity when mutating a real partition.

### 4. `vyntrio-installer apply`

**Preconditions:** same gates as install writes; target must have exactly one existing
supported unmounted partition.

**Does:** mount → copy six payloads → verify → unmount; audit records under
`install-apply/` and `install-target/`.

**Does not:** sandbox fallback, partition creation, formatting, bootloader, full OS
install.

### 5. `vyntrio-installer postflight`

**Preconditions:** `--target-root` containing `HANDOVER_RECORD.txt` from a prior run.

**Does:** replay handover summary to stdout/stderr.

**Does not:** mutate disks, copy payloads, complete deployment.

## Infrastructure guidance (secondary to USB-first)

1. Prefer the USB-first product path for end-user installs (Block 9; not yet shipped).
2. Run `preflight` before any Block 10 write.
3. Use `install` **without** `--apply-target` for **dev/lab sandbox staging only**.
4. Use `apply` only for guarded partition writes (future live session or lab).
5. Always pass explicit `--force` and `--target-disk-id` for writes.
6. Never pass kernel names (`/dev/sdX`) to the CLI.

## Boundaries operators must not confuse

| Concept A | Concept B | Distinction |
|-----------|-----------|-------------|
| `verify-artifact` | `install` / `apply` | Verification is read-only; install commands write |
| `preflight` | `install` / `apply` | Preflight checks only; writes happen in later stages |
| `install` (default) | `apply` | Sandbox files vs mounted partition mutation |
| `apply` | full OS install | Six payloads only; services/bootstrap deferred |
| `postflight` | completion | Replay record only; not deployment success |

## Honest limits (all stages)

- Not a monolithic installer
- Not full OS / appliance deployment
- No partition table creation or formatting
- No bootloader installation
- No service enablement or bootstrap (deferred)
- Six allowlisted payload files only

## References

- `internal/platform/installpolicy/`
- `docs/ops/install-staged-workflow.md`
- `docs/ops/release-artifact-verification.md`
- `docs/ops/install-target-preflight.md`
- `docs/ops/install-payload-write.md`
- `docs/ops/install-partition-apply.md`
