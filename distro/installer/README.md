# Installer scaffold (Block 10)

**Secondary to USB-first delivery.** Block 9 (bootable install medium + USB creator) is
the primary operator path. Block 10 tooling here is **reusable internal infrastructure**
for install execution from a **live boot session** — not the product install journey by
itself.

Declarative and read-only installer tooling for greenfield install **execution**
on target hardware. Media creation remains **Block 9** (`distro/install-media/`).

**Authority:** `docs/ADR/0007-appliance-installer-contract.md`

## Status

| Item | State |
|------|--------|
| Installer contract | **Accepted** — ADR-0007 |
| Preflight manifest (`preflight-manifest.yaml`) | **Scaffold** — declarative only |
| Preflight check (`make installer-preflight`) | **Implemented (Slice 10.3)** — read-only |
| Target layout (`target-layout-manifest.yaml`) | **Scaffold (Slice 10.4)** — declarative only |
| Layout plan validation (`make installer-layout-plan`) | **Implemented (Slice 10.4)** — read-only |
| Mutation stub (`make installer-mutation-stub`) | **Implemented (Slice 10.5)** — preflight-gated dry-run |
| Directory mutation (`make installer-mutate-directories`) | **Implemented (Slice 10.6)** — empty state dirs in target-sandbox |
| Payload copy (`make installer-copy-payloads`) | **Implemented (Slice 10.7)** — manifest payloads to target-sandbox |
| Service prep (`make installer-prepare-service`) | **Implemented (Slice 10.8)** — enablement prep in target-sandbox |
| Service enable (`make installer-enable-service`) | **Implemented (Slice 10.9)** — controlled enable in target-sandbox |
| Target/media preflight CLI (`vyntrio-installer preflight`) | **Implemented (Slice 10.10)** — read-only target + optional media |
| Payload install write (`vyntrio-installer install`) | **Implemented (Slice 10.11)** — sandbox payload copy after gates |
| Postflight handover (`HANDOVER_RECORD.txt`, `postflight` subcommand) | **Implemented (Slice 10.12)** |
| Target mutation (`install --apply-target`) | **Implemented (Slice 10.13)** |
| Partition apply (`vyntrio-installer apply`) | **Implemented (Slice 10.14)** |
| Staged pipeline policy (`vyntrio-installer workflow`) | **Implemented (Slice 10.15)** — `docs/ops/installer-policy.md` |
| Service start / bootstrap | **Not started** — future slice |

## Preflight (Slice 10.3)

`make installer-preflight` runs `scripts/installer-preflight.sh`, which validates:

- Install-media envelope context (`media_role: install`)
- Manifest payload presence and inventory (no extra files)
- Media exclusions (no DB, backups, license, bootstrap tokens, recovery tooling)
- Recovery separation from install media

**Does not:** partition disks, copy payloads to target, enable services, or invoke bootstrap.

Local development uses `distro/install-media/envelope/` as media-context stand-in
after `make install-media-envelope`.

## Target and media preflight CLI (Slice 10.10)

`vyntrio-installer preflight` performs read-only validation of an **explicitly
selected** install target disk (opaque `--target-disk-id`) and optional install
media (`--envelope-root`, `--release-manifest`). See `docs/ops/install-target-preflight.md`.

**Does not:** auto-select a target, partition, format, write, flash, or install.

## Payload install write (Slice 10.11)

`vyntrio-installer install` copies the six manifest payloads to
`target-sandbox/<disk-id>/` after `--force`, preflight, and artifact verification.
See `docs/ops/install-payload-write.md`.

**Does not:** write raw block devices, enable services, or run bootstrap.

## Postflight handover (Slice 10.12)

`vyntrio-installer install` writes `HANDOVER_RECORD.txt` with verification results,
deferred items, and mutation scope. Replay with `vyntrio-installer postflight
--target-root <sandbox-path>`. See `docs/ops/install-staged-workflow.md`.

## Staged pipeline policy (Slice 10.15)

The Go installer is a **staged pipeline**, not one monolithic install command.
Run `vyntrio-installer workflow` for operator policy, or see
`docs/ops/installer-policy.md` for command boundaries between
`vyntrio-verify-artifact`, `preflight`, `install`, `apply`, and `postflight`.

## Target layout plan (Slice 10.4)

`make installer-layout-plan` runs `scripts/validate-installer-layout-plan.sh`,
which read-only validates `target-layout-manifest.yaml`:

- ADR-0007 payload target allowlist (6 paths)
- Empty state directories (`/etc/vyntrio/`, `/var/lib/vyntrio/`, backups subdir)
- Partition/filesystem layout marked **deferred**

**Does not:** partition, format, or write to any disk.

See `distro/installer/target-layout-contract.md` for the layout boundary.

## Mutation stub (Slice 10.5)

`make installer-mutation-stub` runs `scripts/installer-mutation-stub.sh`, which:

1. **Requires** successful `installer-preflight` (fail-closed gate)
2. **Consumes** `target-layout-manifest.yaml` via layout plan validation
3. **Writes** a local dry-run record to `distro/installer/dry-run/MUTATION_STUB.txt`
4. **Does not** partition, format, copy payloads to target, enable services, or invoke bootstrap

The dry-run record lists planned directories and payload targets with
`executed: false` for every action.

## Directory mutation (Slice 10.6)

`make installer-mutate-directories` runs `scripts/installer-mutate-directories.sh`,
the **first real mutation step**. It:

1. **Requires** successful preflight and valid layout plan (fail-closed)
2. **Creates** only three **empty** state directories under `distro/installer/target-sandbox/`:
   - `etc/vyntrio/`
   - `var/lib/vyntrio/`
   - `var/lib/vyntrio/backups/`
3. **Refuses** any target root outside `target-sandbox/` (no host `/etc` or `/var/lib` writes)
4. **Does not** copy payloads, install config, enable services, or partition disks

Production install will map the same relative layout onto the real target filesystem
in a future slice.

## Payload copy (Slice 10.7)

`make installer-copy-payloads` runs `scripts/installer-copy-payloads.sh`, which:

1. **Requires** successful preflight and valid layout plan (fail-closed)
2. **Depends on** Slice 10.6 directory layout in `target-sandbox/`
3. **Copies** only the six manifest-listed files from `distro/install-media/envelope/payload/`
4. **Refuses** any target root outside `target-sandbox/`
5. **Does not** enable services, invoke bootstrap, or partition disks

Writes `target-sandbox/PAYLOAD_COPY.txt` with `bootstrap_handoff: deferred`.

## Service enablement preparation (Slice 10.8)

`make installer-prepare-service` runs `scripts/installer-prepare-service.sh`, which:

1. **Requires** successful preflight, layout plan, and payload copy gates (fail-closed)
2. **Verifies** `vyntrio-api.service` is present in the target sandbox from Slice 10.7
3. **Writes** enablement prep marker `etc/systemd/system/vyntrio-api.service.enable-prep`
4. **Writes** `target-sandbox/SERVICE_PREP.txt` with `enablement_status: prepared_not_enabled`
5. **Does not** invoke systemctl, start services, or run bootstrap

## Controlled service enablement (Slice 10.9)

`make installer-enable-service` runs `scripts/installer-enable-service.sh`, which:

1. **Requires** successful preflight, layout plan, payload copy, and service-prep gates (fail-closed)
2. **Creates** systemd wants symlink `etc/systemd/system/multi-user.target.wants/vyntrio-api.service`
3. **Updates** enablement marker to `enabled: true`, `started: false`
4. **Writes** `target-sandbox/SERVICE_ENABLE.txt` with `enablement_status: enabled_not_started`
5. **Does not** invoke systemctl, start services, or run bootstrap

## Related

- `docs/ADR/0007-appliance-installer-contract.md`
- `distro/install-media/manifest.yaml`
- `scripts/installer-preflight.sh`
- `scripts/validate-installer-layout-plan.sh`
- `scripts/installer-mutation-stub.sh`
- `scripts/installer-mutate-directories.sh`
- `scripts/installer-copy-payloads.sh`
- `scripts/installer-prepare-service.sh`
- `scripts/installer-enable-service.sh`
- `cmd/installer/main.go` — `preflight`, `install`, `apply`, `postflight`, `workflow`
