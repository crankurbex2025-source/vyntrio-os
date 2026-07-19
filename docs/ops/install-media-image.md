# Install-media boot chain + image emission

Operator and engineering contract for **Block 9 / Slice 9.12** — real boot chain
and best-available image emission, the second executable USB-first step.

## Purpose

Move the Slice 9.11 bootability *foundation* (text stubs) toward a real,
firmware-bootable medium:

1. Replace `boot/*.stub` placeholders with a **real boot chain** — a real Linux
   kernel, a real initrd, a real GRUB core image, and a real `grub.cfg`.
2. Expand `live_root/` toward a minimal runnable environment (`os-release`,
   `fstab`, structured `live-init.sh`).
3. Emit the **closest-to-bootable** artifact the build host can produce.
4. Stay honest: the image is **not yet firmware-bootable** in this environment.

This advances the **primary product path** (USB creator → bootable medium → boot).
It does **not** implement the USB creator, target-disk install, or Block 10 work.

## Command

```bash
make install-media-image        # build real boot chain + emit image
make test-install-media-image   # verify stub-vs-real + emission honesty
```

Depends on `make install-media-bootability` (Slice 9.11), which depends on the
envelope + staging + build chain.

## Inputs (build host)

| Input | Source | Override |
|-------|--------|----------|
| Linux kernel | newest `/boot/vmlinuz-*` | `VYNTRIO_HOST_KERNEL` |
| initrd | matching/newest `/boot/initrd.img-*` | `VYNTRIO_HOST_INITRD` |
| GRUB BIOS modules | `/usr/lib/grub/i386-pc` | `VYNTRIO_GRUB_I386_PC_DIR` |

Kernel is verified with `grub-file --is-x86-linux`; initrd is magic-checked
(gzip/xz/zstd/lz4/lzop/bzip2/newc-cpio). If any input is missing, the build
falls back to `boot_chain: stub` and carries the 9.11 placeholders forward — it
does not fabricate a fake kernel.

## Outputs

| Path | Meaning |
|------|---------|
| `envelope/boot/vmlinuz` | Real Linux kernel (real branch) |
| `envelope/boot/initrd.img` | Real initramfs (real branch) |
| `envelope/boot/grub/grub.cfg` | Real GRUB config referencing kernel + initrd |
| `envelope/boot/grub/i386-pc/core.img` | Real GRUB core (via `grub-mkimage`) |
| `envelope/boot/grub/i386-pc/{boot.img,cdboot.img,*.mod}` | Real BIOS boot blocks |
| `build/vyntrio-install-media.iso` | Firmware-bootable ISO — **only** when an ISO9660 writer exists |
| `build/vyntrio-install-media-REAL-BOOTCHAIN-NO-ISO.tar` | Real boot chain, not yet wrapped (this environment) |
| `build/IMAGE.txt` | Provenance: `boot_chain`, `firmware_bootable`, verified checks, blockers |

All `build/` and `envelope/` outputs are gitignored/disposable — real kernel and
initrd binaries are **never committed**.

## Firmware-bootability blockers (this environment)

A firmware-bootable image is **not** produced here. Concrete blockers, recorded
in `IMAGE.txt`:

1. `iso9660_or_eltorito_writer_missing` — no `xorriso` / `genisoimage`, so
   `grub-mkrescue` cannot emit a bootable ISO (and it cannot be fetched offline).
2. `raw_image_filesystem_tooling_missing` — no `mkfs.vfat` / `mtools` and no
   loop-mount privileges, so a formatted raw USB image cannot be built.
3. `efi_grub_target_missing` — only the `i386-pc` (BIOS) GRUB target is installed;
   no `x86_64-efi` modules for a UEFI image.

When run on a host that has `grub-mkrescue` + `xorriso`, the same target emits a
real bootable `.iso` and sets `firmware_bootable: true` automatically.

## Validation (stub vs real)

`test-install-media-image` reads `IMAGE.txt` and asserts content matches:

- **real**: `boot/vmlinuz` passes `grub-file --is-x86-linux`; initrd magic valid;
  `grub.cfg` has real `linux`/`initrd` directives and no `.stub`; `core.img`
  present and non-trivial; 9.11 stub files removed.
- **stub**: 9.11 `*.stub` placeholders still present; `firmware_bootable: false`.
- **firmware_bootable: true** is allowed only alongside an `.iso` artifact and no
  blockers; otherwise the non-bootable artifact must name concrete blockers and
  the tar must carry `NOT_BOOTABLE.txt`.

It also re-checks 9.11 invariants (payload allowlist, no secrets in `live_root`,
fail-closed on a missing foundation) so the new slice does not regress them.

## Explicit non-goals (this slice)

- Fetching/installing an ISO writer or building a raw USB image
- UEFI/`x86_64-efi` boot support
- Complete live rootfs userland (busybox/systemd) or a running dashboard
- Host USB creator / `dd` flashing tool
- Target-disk install / Block 10 expansion
- Recovery media bootability

## Missing for USB creator / first boot

1. Firmware-bootable image container (ISO9660/El Torito or raw image)
2. Complete live-root userland inside the initrd
3. Live session that runs `vyntrio-api` from media and serves the wizard UI
4. Host USB writer tool

## References

- `distro/install-media/bootability-contract.md`
- `distro/install-media/bootability-manifest.yaml`
- `scripts/build-install-media-image.sh`
- `docs/ops/install-media-bootability.md` (Slice 9.11 foundation)
- `docs/00_PROJECT.md`
