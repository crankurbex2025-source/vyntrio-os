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

## Related

- `docs/ADR/0007-appliance-installer-contract.md`
- `distro/install-media/manifest.yaml`
- `scripts/installer-preflight.sh`
- `scripts/validate-installer-layout-plan.sh`
- `cmd/installer/main.go` — stub entrypoint
