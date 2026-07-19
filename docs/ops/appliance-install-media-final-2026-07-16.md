# Install media / download path — audit addendum (2026-07-16)

Supplements `docs/ops/appliance-product-gap-final-2026-07-16.md` after Slice 13.1.

## Public install media / Download & Install

### What artifact exists

- **Primary:** `vyntrio-install-media-bios.img` — raw MBR BIOS disk (~136 MB)
- **Build path:** `make install-media` → `distro/install-media/build/`
- **Staging path:** `make release-install-media-stage` → `distro/release/staging/`
- **Provenance:** `WRAPPER.txt` records `firmware_bootable: true`, `firmware_boot_mode: bios_legacy`

### How a user obtains it

1. **Build from source:** `make install-media` (requires full build toolchain + install-media deps)
2. **Stage locally:** `make release-install-media-stage`
3. **Verify:** `vyntrio-verify-artifact --base-dir distro/release/staging distro/release/staging/release-manifest.json`
4. **Download (same host as API):** set `VYNTRIO_RELEASE_STAGING_DIR` to staging dir, restart API, fetch `/release/vyntrio-install-media-bios.img`

There is **no production CDN** or always-on public download URL.

### How a user installs/boots it

1. **Guided flow:** `/download` → **Create bootable install media** (checksum, USB steps, VM steps)
2. **USB (bare metal):** `./scripts/write-install-media-usb.sh --device /dev/sdX` (or manual `dd`)
3. **VM (non-destructive):** `./scripts/run-install-media-vm.sh` or attach `.img` in BIOS/legacy mode
4. Boot → live initramfs → supervised `vyntrio-api` → local dashboard on loopback

### Hardware supported

- x86_64 legacy BIOS bare metal
- x86_64 BIOS-mode VMs

### What is incomplete

- UEFI boot
- In-browser USB writer (CLI helpers + `/download` guide exist)
- Production artifact hosting
- Boot-proven dashboard on build host (`dashboard_reachable_on_boot: false`)
- Target-disk persistent install (installer service not started)
- Ed25519 release signatures

## Revised capability classification

| # | Category | Previous | Now |
|---|----------|----------|-----|
| 1 | Boot media | partial | **partial** — buildable + stageable + local HTTP; no CDN |
| 2 | First boot | partial | **partial** — live path wired; not boot-proven |
| 15 | Website/product clarity | implemented | **implemented** — `/download` shows live API status + build path |

## Remaining gaps for user-friendly install

1. Production download hosting (CDN/object storage)
2. UEFI image variant
3. Boot-proven dashboard on build host (`dashboard_reachable_on_boot: false`)
4. Target-disk persistent install (installer service not started)
5. Signed releases (Ed25519)
6. Optional graphical USB writer (beyond CLI helpers)

## Recommended next block

Runtime boot verification when qemu/KVM is available, or target-disk installer preflight — not CDN work until boot path is proven.
