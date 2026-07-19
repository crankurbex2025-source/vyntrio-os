# Install media audit (2026-07-17)

Factual inventory of Vyntrio install/boot artifacts and the recommended primary path.

## Product baseline

| Rule | Meaning |
|------|---------|
| Dual-mode required | `bios+uefi` GPT hybrid with ESP/`BOOTX64.EFI` |
| BIOS-only | **Incomplete** — must not be staged as primary |
| Modern hardware claim | Requires verified UEFI packaging (structural); runtime boot separate |

## Artifacts that exist today

| Artifact | Path | Size (this host) | Type | Boot |
|----------|------|------------------|------|------|
| **Primary install image** | `distro/install-media/build/vyntrio-install-media.img` | ~196 MB | Raw GPT hybrid (`raw_gpt_hybrid_disk`) | `firmware_boot_mode: bios+uefi`, `uefi_support: true`, `dual_mode: true` |
| Legacy BIOS-only (incomplete) | `…/vyntrio-install-media-bios.img` | when hybrid tools missing | Raw MBR BIOS | **Not product baseline** |
| Live initramfs (intermediate) | `distro/install-media/build/vyntrio-live-initramfs.cpio.gz` | varies | cpio.gz initrd | swapped into image by Slice 9.17 |
| Boot chain sources | `distro/install-media/envelope/boot/` | — | kernel + initrd + GRUB | real boot chain |
| Honest fallback tar | `distro/install-media/build/vyntrio-install-media-REAL-BOOTCHAIN-NO-ISO.tar` | when wrap fails | tar | not firmware-bootable |

Hybrid ESP offset: `distro/install-media/build/HYBRID_ESP_OFFSET.txt`.

## Build chain (Makefile)

Recommended one-shot build:

```bash
make install-media          # alias for install-media-initrd-swap
./scripts/verify-uefi-boot.sh
make release-install-media-stage   # refuses BIOS-only
make test-release-install-media-stage
```

Full dependency chain:

`install-media-stage` → `install-media-envelope` → `install-media-bootability` →
`install-media-image` → `install-media-wrap` → `install-media-runtime` →
`install-media-live-rootfs` → `install-media-hardware` → **`install-media-initrd-swap`**

Producer scripts: `scripts/wrap-install-media-image.sh`, `scripts/swap-live-initramfs-into-image.sh`, `scripts/stage-release-install-media.sh`.

## Recommended primary install path

| Step | Action |
|------|--------|
| 1 | Build: `make install-media` |
| 2 | Stage for download: `make release-install-media-stage` → `distro/release/staging/` |
| 3 | Verify: `vyntrio-verify-artifact --base-dir distro/release/staging distro/release/staging/release-manifest.json` |
| 4 | Write: `vyntrio-media-creator` / `vyntrio-write-media` **or** `./scripts/write-install-media-usb.sh` **or** VM attach |
| 5 | Boot: **UEFI preferred**, BIOS/legacy fallback |
| 6 | First boot: live initramfs → `firstboot.sh` → local dashboard on loopback |

## Supported scenarios (honest)

| Scenario | Supported |
|----------|-----------|
| UEFI x86_64 bare metal | **Yes** (structurally verified ESP/`BOOTX64.EFI`; runtime best-effort) |
| BIOS/legacy x86_64 bare metal | **Yes** (dual-mode hybrid) |
| UEFI VM (QEMU + OVMF) | **Yes** when qemu/OVMF present (`./scripts/run-install-media-vm.sh --firmware uefi`) |
| BIOS/legacy VM | **Yes** (`--firmware bios`) |
| Secure Boot signed | **No** |
| Public CDN download | **No** — local staging + `/release/*` on API host only |
| Target-disk full install | **No** — live RAM runtime only |
| Boot-proven dashboard | **No** until harness reports dashboard reachability |

## API / website surfacing

| Surface | Path |
|---------|------|
| Public metadata | `GET /api/v1/public/install-media` (includes `bios_support`, `uefi_support`, `dual_mode`) |
| Local download | `GET /release/vyntrio-install-media.img` when `VYNTRIO_RELEASE_STAGING_DIR` set |
| Download page | `/download` — BIOS / UEFI / dual-mode rows + writer section |
| Writer | `vyntrio-media-creator` (local web GUI) + `vyntrio-write-media` CLI |
| Operator doc | `docs/24_INSTALL_MEDIA.md` |

## Not implemented

- Production artifact hosting / CDN
- Secure Boot signing
- In-browser USB writer
- Ed25519 signature verification (manifest integrity only)
- Guaranteed runtime boot/dashboard proof on every CI host
