# Install-media bootability foundation

Operator and engineering contract for **Block 9 / Slice 9.11** — first executable
USB-first media foundation.

## Purpose

Turn the Slice 9.8 envelope into a **bootability foundation**:

1. Replace deferred `boot/` and `live_root/` placeholders with structured stubs
2. Validate envelope layout (boot + live_root + allowlisted payload)
3. Emit a **tangible but non-bootable** image stub under `distro/install-media/build/`

This advances the **primary product path** (USB creator → bootable medium → boot).
It does **not** implement the USB creator or a firmware-bootable ISO yet.

## Command

```bash
make install-media-bootability
make test-install-media-bootability
```

Depends on `make install-media-envelope` (which depends on staging + build).

## Outputs

| Path | Meaning |
|------|---------|
| `envelope/boot/BOOT_LAYER.txt` | Boot stub provenance |
| `envelope/boot/vmlinuz.stub` | Kernel placeholder (not a real kernel) |
| `envelope/boot/initrd.img.stub` | Initrd placeholder |
| `envelope/boot/loader/entries/vyntrio-install.conf.stub` | Bootloader entry stub |
| `envelope/live_root/LIVE_ROOT.txt` | Live-root stub provenance |
| `envelope/live_root/usr/lib/vyntrio/live-init.sh` | Stub live init script |
| `envelope/BOOTABILITY.txt` | Foundation record (`firmware_bootable: false`) |
| `build/vyntrio-install-media-NOT-BOOTABLE.stub.tar` | Image stub archive |
| `build/IMAGE_STUB.txt` | Emission provenance |

## Explicit non-goals (this slice)

- Firmware-bootable ISO or raw USB image
- Real kernel / initrd / GRUB / systemd-boot
- Complete Linux rootfs
- Host USB creator / `dd` flashing tool
- Target-disk install / Block 10 expansion
- Recovery media bootability

## Missing for USB creator

1. Real boot chain binaries
2. Complete live root
3. Firmware-bootable image emission
4. Host USB writer

## References

- `distro/install-media/bootability-contract.md`
- `distro/install-media/bootability-manifest.yaml`
- `scripts/initialize-install-media-bootability.sh`
- `docs/00_PROJECT.md`
