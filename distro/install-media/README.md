# Install media scaffold (Block 9, Slice 9.3)

Declarative scaffold for **greenfield install USB/ISO** payloads. This directory
describes what install media **is for** and which artifacts it may carry. It does
**not** build bootable images, partition disks, or run an installer.

**Authority:** `docs/ADR/0006-appliance-install-recovery-media.md` (install media
sections B, D, E). Runtime paths and ownership remain `docs/ADR/0005-appliance-runtime-operations.md`.

## Status

| Item | State |
|------|--------|
| Layout manifest (`manifest.yaml`) | **Scaffold** — declarative only |
| Build contract (`build-contract.md`) | **Contract** — install-image I/O boundary (Slice 9.5) |
| Payload staging (`make install-media-stage`) | **Implemented (Slice 9.6)** — local `staging/payload/` only |
| Live/boot envelope (`envelope-manifest.yaml`) | **Scaffold (Slice 9.7)** — declarative layer contract |
| Envelope assembly (`make install-media-envelope`) | **Implemented (Slice 9.8)** — local `envelope/` only |
| Bootability contract (`bootability-manifest.yaml`) | **Contract + status through Stage 2 / Slice 10.1** (authoritative) |
| Config template (`config.toml.template`) | **Scaffold** — operator-edited at install |
| Boot/live layer population | **Stub (9.11) → real (Slice 9.12)** — `make install-media-image` (kernel/initrd/GRUB) |
| Image stub emission | **Superseded** — real image chain from Slice 9.12 |
| Firmware-bootable image | **Implemented (Slice 9.13)** — raw MBR/BIOS `.img` via `grub-bios-setup`, structurally verified |
| Runtime boot + first-boot path | **Implemented (Slice 9.14)** — `make install-media-runtime` (qemu-gated; honest skip w/o VM) |
| Live-rootfs userland | **Implemented (Slice 9.15)** — `make install-media-live-rootfs` (chroot-verified dashboard) |
| Live-initramfs hardware enablement | **Implemented (Slice 9.16)** — `make install-media-hardware` (storage/net modules + DHCP) |
| Image boots Vyntrio live initramfs | **Implemented (Slice 9.17)** — `make install-media-initrd-swap` (sha256-proven from image) |
| Runtime boot verification (Stage 2) | **Implemented, blocked here (Slice 10.1)** — `make install-media-runtime-verify` (read-only; fails closed w/o qemu) |
| Local dashboard stability (Stage 2) | **Implemented (Slice 10.2)** — dashboard-first; busybox-sh supervised `firstboot.sh`; `/readyz` local proof; LAN bind blocked on TLS |
| First-boot local onboarding clarity (Stage 2) | **Implemented (Slice 10.3)** — clear boot→local dashboard→sign-in→overview prompt; landing page `#first-boot-setup` mirrors the same flow |
| Dashboard-on-boot proof (real HTTP from boot) | **Blocked** — needs a VM/hardware harness (no qemu/`/dev/kvm` here) |
| Host USB creator | **Not started** — deferred |
| `vyntrio-installer` (Block 10) | **Secondary internal infra** — not primary delivery |
| Recovery media | **Separate deliverable** — not under `install-media/` |

## What install media is for

1. Boot a minimal live environment (future).
2. Lay down approved on-disk layout (future slice).
3. Copy **immutable payloads** from `manifest.yaml` onto the target disk.
4. Provision `vyntrio` account and directory layout per `distro/systemd/README.md`.
5. Install `config.toml` from the shipped **template** (no secrets, no Owner).
6. Enable `vyntrio-api.service` and hand off to ADR-0004 bootstrap on the
   **running** installed service.

## What install media is not for

- Persistent appliance state (`/var/lib/vyntrio/` SQLite, backups) — **target disk only**
- Owner credentials, bootstrap tokens, or license files
- In-place restore from backup archives (recovery media + Block 7 — separate)
- Product API, dashboard, or install-progress UI
- Storing operator backup artifacts

## Directory layout (repository scaffold)

```
distro/
├── install-media/          ← this scaffold (payload + envelope contract)
│   ├── README.md
│   ├── manifest.yaml       ← declarative payload list
│   ├── envelope-manifest.yaml  ← live/boot envelope layers (Slice 9.7)
│   ├── envelope-contract.md    ← envelope boundary contract (Slice 9.7)
│   ├── bootability-manifest.yaml ← bootable init layers (Slice 9.9)
│   ├── bootability-contract.md   ← bootable init boundary (9.9) + foundation (9.11)
│   ├── build-contract.md   ← install-image build I/O (Slice 9.5)
│   ├── config.toml.template
│   ├── staging/            ← local disposable payload staging (Slice 9.6, gitignored)
│   ├── envelope/           ← local disposable envelope (Slices 9.8/9.11, gitignored)
│   └── build/              ← image stub emission (Slice 9.11, gitignored)
├── systemd/                ← runtime deployment artifacts (Block 7.3)
│   └── …
└── README.md
```

Future build slices replace stubs with real bootloader/kernel/rootfs. Slice 9.11
emits a non-bootable stub via `make install-media-bootability`.

## Payload sources

| Manifest role | Repository source | Installed path (target disk) |
|---------------|-------------------|------------------------------|
| API binary | `bin/vyntrio-api` (from `make build`) | `/usr/bin/vyntrio-api` |
| Backup binary | `bin/vyntrio-backup` (from `make build`) | `/usr/bin/vyntrio-backup` |
| systemd unit | `distro/systemd/vyntrio-api.service` | `/etc/systemd/system/vyntrio-api.service` |
| sysusers | `distro/systemd/vyntrio.sysusers` | `/usr/lib/sysusers.d/vyntrio.conf` |
| tmpfiles | `distro/systemd/vyntrio.tmpfiles.conf` | `/etc/tmpfiles.d/vyntrio.conf` |
| config template | `distro/install-media/config.toml.template` | `/etc/vyntrio/config.toml` |

Installation steps for systemd artifacts remain documented in
`distro/systemd/README.md` until an automated installer exists.

## Target disk paths (created at install, not stored on USB)

| Path | Purpose |
|------|---------|
| `/var/lib/vyntrio/` | Persistent state (`StateDirectory`; SQLite) |
| `/var/lib/vyntrio/backups/` | Root-only backup destination (future writes by `vyntrio-backup`) |
| `/etc/vyntrio/` | Runtime configuration directory |
| `/etc/vyntrio/config.toml` | From template; administrator-edited |

## Recovery media

Recovery boot USB/ISO is a **separate deliverable** per ADR-0006 section C,
scaffolded under `distro/recovery-media/` (Block 9, Slice 9.4). It is not part
of this install-media tree.

## Related

- `docs/ADR/0006-appliance-install-recovery-media.md`
- `docs/ADR/0005-appliance-runtime-operations.md` — sections D, E
- `docs/ADR/0004-identity-and-access.md` — post-install bootstrap
- `docs/ops/restore-safety-contract.md` — in-place restore (not install)
- `distro/recovery-media/README.md` — recovery media (separate deliverable)
- `distro/install-media/build-contract.md` — install-image build I/O contract
- `distro/install-media/envelope-contract.md` — live/boot envelope contract (Slice 9.7)
- `distro/install-media/envelope-manifest.yaml` — envelope layer inventory (Slice 9.7)
- `distro/install-media/bootability-contract.md` — bootable initialization contract (Slice 9.9)
- `distro/install-media/bootability-manifest.yaml` — bootability layer inventory (Slice 9.9)
- `scripts/stage-install-media.sh` — local payload staging (Slice 9.6)
- `scripts/assemble-install-media-envelope.sh` — local envelope assembly (Slice 9.8)
- `scripts/initialize-install-media-bootability.sh` — bootability foundation (Slice 9.11)
- `docs/ops/install-media-bootability.md` — Slice 9.11 operator contract
- `distro/systemd/README.md` — interim developer/lab install path
