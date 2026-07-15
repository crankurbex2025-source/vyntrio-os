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

## Related

- `docs/ADR/0007-appliance-installer-contract.md`
- `distro/install-media/manifest.yaml`
- `scripts/installer-preflight.sh`
- `cmd/installer/main.go` — stub entrypoint
