# Install-media firmware wrapper

Operator and engineering contract for wrapping the real boot chain into a
firmware-bootable image, or falling back honestly when no strategy is available.

## Product baseline (hard requirement)

**Dual-mode (BIOS + UEFI) is required.** BIOS-only media is **incomplete** and
must not be staged as the product install artifact or described as modern-hardware
ready.

| Capability | Baseline |
|------------|----------|
| UEFI (`BOOTX64.EFI` on ESP) | Required |
| BIOS/legacy GRUB | Required fallback on hybrid media |
| BIOS-only raw MBR | Incomplete — engineering fallback only |

## Purpose

Make the real boot chain (kernel + initrd + GRUB + `grub.cfg`) consumable by
firmware:

1. **Detect** whether firmware-bootable emission is possible on this host.
2. **Emit** a real bootable image preferring dual-mode hybrid, then ISO, then BIOS-only.
3. **Fall back honestly** with blockers when UEFI tools are missing.
4. **Validate** structural axes including ESP/`BOOTX64.EFI` for hybrid media.

## Command

```bash
make install-media-wrap        # wrap into a firmware-bootable image or fall back
make test-install-media-wrap   # verify axes + no regression
./scripts/verify-uefi-boot.sh  # ESP/BOOTX64 structural (+ optional qemu/OVMF)
```

Depends on `make install-media-image` (real boot chain).

## Strategies (priority order)

| # | Strategy | Tools | Result |
|---|----------|-------|--------|
| A | Raw GPT hybrid (BIOS + UEFI) | `sgdisk` + `mkfs.vfat` + `mtools` + `grub-mkimage` (`i386-pc` + `x86_64-efi`) + `grub-bios-setup` | `vyntrio-install-media.img`, `firmware_boot_mode: bios+uefi`, `product_baseline_complete: true` |
| B | ISO9660 / El Torito | `grub-mkrescue` + `xorriso` | `vyntrio-install-media.iso` (when preferred/available) |
| C | Raw MBR/BIOS only | `mke2fs` + `debugfs` + `sfdisk` + `grub-bios-setup` | `vyntrio-install-media-bios.img`, **incomplete** (`uefi_support: false`) |
| D | Honest fallback | — | real boot-chain tar + `WRAPPER.txt` blockers |

Strategy **A** is the product baseline on hosts with `grub-efi-amd64-bin`,
`dosfstools`, `mtools`, and `gdisk`/`sgdisk`.

Hybrid layout:

1. GPT partition 1: BIOS boot (EF02, ~1 MiB) for `grub-bios-setup`
2. GPT partition 2: ESP FAT32 (EF00) with `EFI/BOOT/BOOTX64.EFI`, `grub.cfg`, `/boot/{vmlinuz,initrd.img}`

## Honesty model

`firmware_bootable: true` means a real bootable-format image was **constructed and
structurally verified**. For hybrid media that includes:

- GPT with EF02 + EF00 partitions
- `BOOTX64.EFI` present as a PE (MZ header)
- shared kernel/initrd reachable from GRUB config

It does **not** claim the image was booted on hardware. `runtime_boot_tested: false`
is recorded until a VM/hardware harness reports success.

`product_baseline_complete: true` requires `uefi_support: true` and `dual_mode: true`.
Release staging (`make release-install-media-stage`) **refuses** BIOS-only images.

## Provenance (`build/WRAPPER.txt`)

Key fields: `boot_chain`, `firmware_bootable`, `image_wrapper`, `artifact`,
`artifact_format`, `boot_method`, `firmware_boot_mode`, `bios_support`,
`uefi_support`, `dual_mode`, `product_baseline_complete`, `structural_verification`,
`runtime_boot_tested`, and either `blockers: none` or a per-strategy blockers list.

## Blockers (recorded when relevant)

- **Hybrid/UEFI:** missing `grub-mkimage` `x86_64-efi` — apt: `grub-efi-amd64-bin`
- **Hybrid tools:** missing `mkfs.vfat` / `mcopy` / `sgdisk` — apt: `dosfstools`, `mtools`, `gdisk`
- **ISO strategy:** no `xorriso`/`xorrisofs`/`genisoimage` — apt: `xorriso`
- **Runtime boot:** no `qemu`/OVMF/KVM to actually boot-test the image

## Missing beyond UEFI packaging

1. Complete live-root userland inside the initrd (when not swapped)
2. Live session that runs `vyntrio-api` from media and serves the wizard UI
3. Secure Boot signing
4. Runtime boot verification harness success on CI hosts without KVM

## References

- `distro/install-media/bootability-contract.md`
- `scripts/wrap-install-media-image.sh`
- `scripts/verify-uefi-boot.sh`
- `docs/24_INSTALL_MEDIA.md`
- `docs/ops/install-media-image.md`
- `docs/ops/install-media-bootability.md`
