# Install-media live-initramfs hardware enablement

Operator and engineering contract for **Block 9 / Slice 9.16** — the smallest
viable hardware bring-up added to the live initramfs so an actual boot can reach
the live runtime more reliably.

## Purpose

Slice 9.15 composed a live-rootfs userland that runs and serves the dashboard in
chroot, but it carried **no kernel modules** and only brought up loopback. That
was the explicit gap blocking a real boot from finding disks and networks. This
slice closes the first half of that gap:

1. A curated, dependency-resolved set of **storage + network kernel modules**,
   matched to the same host kernel the boot chain ships (Slice 9.12).
2. A **module-load + networking (DHCP)** bring-up path invoked from `/init`
   before the first-boot dashboard step.
3. A **re-emitted live initramfs** that now carries those modules.

Minimal increment only — no installer/apply work, no target-disk writes, no
broad distro assembly.

## Command

```bash
make install-media-hardware        # stage modules + wire bring-up + re-emit initramfs
make test-install-media-hardware   # verify staging/dep/bring-up + record honesty
```

Depends on `make install-media-live-rootfs` (Slice 9.15). It augments the
existing `live_root/`; it does **not** recompose the 9.15 userland.

Overrides: `VYNTRIO_LIVE_KVER` (target kernel version), `VYNTRIO_MODULES_DIR`
(module source tree), plus the shared `VYNTRIO_INSTALL_ENVELOPE_ROOT` /
`VYNTRIO_INSTALL_MEDIA_BUILD_ROOT`.

## What it adds

| Path | Role |
|------|------|
| `live_root/lib/modules/<kver>/` | Dep-resolved storage/network `.ko` set + regenerated `modules.dep` (via `depmod`) |
| `live_root/etc/vyntrio/modules.load` | Ordered module load list consumed at live boot |
| `live_root/usr/lib/vyntrio/hw-init.sh` | Loads modules, then DHCP on each non-loopback interface |
| `live_root/usr/share/udhcpc/default.script` | busybox `udhcpc` lease handler (configures IP/route/DNS) |
| `live_root/bin/{modprobe,insmod,rmmod,lsmod,depmod,udhcpc,route}` | busybox applet symlinks for bring-up |
| `live_root/init` | Now runs `hw-init.sh` **before** `firstboot.sh` |
| `build/vyntrio-live-initramfs.cpio.gz` | Re-emitted, now carrying the modules |

Module selection (top-level, before dependency resolution):

- **Storage:** `ahci libahci sd_mod sr_mod nvme virtio_blk virtio_scsi virtio_pci ata_piix ata_generic xhci_pci ehci_pci uhci_hcd usb_storage`
- **Filesystem:** `ext4 vfat isofs overlay squashfs`
- **Network:** `virtio_net e1000 e1000e r8169 igb tg3`

Modules absent from a given host kernel are recorded under
`not_present_on_host_kernel` and are non-fatal.

## /init order

```
mount proc/sys/dev/run  ->  hw-init.sh (modules + DHCP)  ->  firstboot.sh (dashboard)  ->  shell
```

## Verification — what it proves, and what it does not

Verified on the build host without a VM:

1. **Modules staged + dep-resolved** — the requested set (plus transitive deps
   from `modules.dep`) is copied into `live_root`, and `depmod -b` regenerates a
   valid `modules.dep` there.
2. **`modinfo` validity** — a sample staged module parses correctly.
3. **Load list integrity** — every name in `modules.load` maps to a staged `.ko`.
4. **`/init` integration** — `hw-init.sh` runs ahead of `firstboot.sh`.
5. **Userland still executes** — `busybox sh` runs under the composed loader.
6. **Initramfs carries modules** — the re-emitted `cpio.gz` contains
   `lib/modules/<kver>/modules.dep` and still has `/init` + busybox at root.

**Modules are staged, not loaded here.** Running `modprobe` on the build host
would load into the *build-host* kernel, which is out of scope and unsafe. Actual
module load happens only inside a real live boot. This is recorded explicitly.

## Honest limits (unchanged from 9.15)

The live initramfs is still **not** wired in as the image's initrd
(`wired_into_image_initrd: false`), the image is still **not** booted
(`boot_verified: false` — no qemu/VM harness), and the dashboard is still **not**
reachable on boot (`dashboard_reachable_on_boot: false`).

This slice supersedes the specific 9.15 gap
`kernel_modules_for_storage_and_network_in_initramfs`.

## Still missing for true dashboard-on-boot

1. **Image initrd swapped** to the live initramfs (blocked on runtime verification)
2. **Runtime boot verification** on a VM or real hardware (no qemu here)
3. **Module autoload coverage** for arbitrary hardware (this is a curated set +
   DHCP, not full udev/coldplug hardware discovery)

## Still missing for USB creator

1. Host USB writer tool
2. Secure Boot signing (UEFI dual-mode packaging is the baseline; unsigned)

## Boundary and safety

- Modules land in `live_root/lib/modules` (the **live runtime**) and never in
  `payload/` (the target-disk artifact). The test asserts no
  `payload/lib/modules` and the 6-file payload allowlist is unchanged.
- Appliance state, credentials, tokens, licenses and secrets remain forbidden in
  `live_root` and are enforced fail-closed.
- No installer/apply behavior, no target-disk writes, no firmware-image change.

## References

- `distro/install-media/bootability-contract.md`
- `distro/install-media/bootability-manifest.yaml`
- `scripts/enable-live-initramfs-hardware.sh`
- `docs/ops/install-media-live-rootfs.md` (Slice 9.15 userland)
- `docs/ops/install-media-runtime-boot.md` (Slice 9.14 first-boot wiring)
