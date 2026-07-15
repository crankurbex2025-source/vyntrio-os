# Installer scaffold (Block 10)

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
| Target-disk install execution | **Not started** — future slice |
| Bootstrap handoff | **Not started** — future slice |

## Preflight (Slice 10.3)

`make installer-preflight` runs `scripts/installer-preflight.sh`, which validates:

- Install-media envelope context (`media_role: install`)
- Manifest payload presence and inventory (no extra files)
- Media exclusions (no DB, backups, license, bootstrap tokens, recovery tooling)
- Recovery separation from install media

**Does not:** partition disks, copy payloads to target, enable services, or invoke bootstrap.

Local development uses `distro/install-media/envelope/` as media-context stand-in
after `make install-media-envelope`.

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

## Related

- `docs/ADR/0007-appliance-installer-contract.md`
- `distro/install-media/manifest.yaml`
- `scripts/installer-preflight.sh`
- `scripts/validate-installer-layout-plan.sh`
- `scripts/installer-mutation-stub.sh`
- `scripts/installer-mutate-directories.sh`
- `cmd/installer/main.go` — stub entrypoint
