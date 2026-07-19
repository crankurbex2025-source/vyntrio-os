# Vyntrio USB appliance image

Product path for a **USB-boot appliance** (behavior modeled on public Unraid
docs only — no Unraid binaries, rootfs, or assets).

## What the image is

| Layer | Location |
|-------|----------|
| GPT hybrid disk | `vyntrio-install-media.img` |
| BIOS boot | Partition 1 EF02 |
| UEFI ESP | Partition 2 EF00 — `EFI/BOOT/BOOTX64.EFI`, `boot/vmlinuz`, Vyntrio early `initrd.img` |
| Appliance root | Partition 3 `VYNTRIO_SYS` — `/vyntrio/system.squashfs` + `config/` + `state/` |

**Secure Boot:** unsupported (explicit).

**Not this image:** host Debian initrd stub, installer-first target-disk media.

## Build

```bash
make install-media-appliance          # or: make install-media
make test-install-media-appliance
make release-install-media-stage
```

Records: `distro/install-media/build/APPLIANCE.txt`, `WRAPPER.txt`, `INITRD_SWAP.txt`.

## Creator version feed

Staging writes `install-media-public.json` with `image_versions[]` and
`image-versions.json`. The Media Creator lists those real entries (version,
build id, size, SHA-256, UEFI/BIOS, Secure Boot status).

## Honesty

- Runtime boot to dashboard is **wired** in the image (`firstboot.sh` → `vyntrio-api`).
- Runtime boot is **not** claimed proven unless a VM/hardware harness recorded it.
- Storage pools / Docker / VMs are **not** claimed complete in this image.
