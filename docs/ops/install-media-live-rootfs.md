# Install-media live-rootfs userland

Operator and engineering contract for **Block 9 / Slice 9.15** — the minimal
live-rootfs userland that can actually start the first-boot path.

## Purpose

Slice 9.14 wired the first-boot path but nothing could run it: `vyntrio-api` and
`busybox` are **dynamically linked**, so the medium had no userland. This slice
composes the smallest userland that closes that gap:

1. Dynamic loader + full `ldd` library closure
2. `busybox` (shell + core utilities)
3. `vyntrio-api` — the **live dashboard runtime**
4. `/init` (PID 1) that hands off to `firstboot.sh`
5. A real live initramfs artifact (`cpio.gz`)

Minimal increment only — no broad appliance features, no installer/apply work.

## Command

```bash
make install-media-live-rootfs        # compose userland + emit live initramfs
make test-install-media-live-rootfs   # verify it actually executes + record honesty
```

Depends on `make install-media-runtime` (Slice 9.14).

## Composition

| Path | Role |
|------|------|
| `live_root/lib64/ld-linux-x86-64.so.2` | Dynamic loader (**dereferenced**, not a symlink) |
| `live_root/lib/x86_64-linux-gnu/` | `ldd` closure of busybox + vyntrio-api (`libc`, `libresolv`) |
| `live_root/bin/busybox` + applet symlinks | `sh`, `mount`, `ip`, `ifconfig`, … |
| `live_root/usr/bin/vyntrio-api` | Live dashboard runtime |
| `live_root/etc/vyntrio/config.toml` | Template config (no secrets, no state) |
| `live_root/init` | PID 1 → mounts pseudo-fs, brings up `lo`, execs `firstboot.sh` |
| `build/vyntrio-live-initramfs.cpio.gz` | Live initramfs (`/init` at archive root) |

Overrides: `VYNTRIO_BUSYBOX`, `VYNTRIO_API_BIN`, `VYNTRIO_LIVE_PROBE_PORT`.

## Verification — what it proves, and what it does not

Two real checks, neither of which requires a VM:

1. **chroot exec** — `busybox sh` and `vyntrio-api` must actually run under the
   composed loader/libc. This is what catches an incomplete closure (e.g. a
   dangling loader symlink).
2. **dashboard HTTP probe** — `vyntrio-api` is started inside a chroot and
   probed. Result here: **HTTP 200**, serving the real embedded UI.

**This proves the userland runs and serves. It is NOT a boot.** The image still
boots kernel+initrd only, so `dashboard_reachable_on_boot: false`.

The probe runs against a **throwaway chroot copy** on a non-default port
(`18080`), because `vyntrio-api` creates a database at runtime and the shipped
`live_root` must stay state-free. The test asserts no database leaked onto media
and that a running production service is never disturbed.

## Live-root exclusion boundary (refined here)

`preinstalled_target_disk_payloads` forbids staging `payload/` onto a target disk
from `live_root`. It does **not** forbid the live session's own dashboard
runtime. `live_root/usr/bin/vyntrio-api` is the **live** runtime; it is distinct
in role from `payload/usr/bin/vyntrio-api`, the target-disk artifact. Appliance
state, credentials, tokens, licenses and secrets remain forbidden and are
enforced fail-closed.

## Why the image initrd is not swapped yet

The live initramfs is emitted but **not** wired in as the image's initrd
(`wired_into_image_initrd: false`). Two honest reasons:

1. This userland carries **no kernel modules** (storage/network drivers), so a
   real boot would likely lack devices and networking.
2. There is **no VM here** to verify the resulting boot — swapping the initrd of
   a verified-bootable image without being able to test it would risk regressing
   bootability on a claim we cannot check.

## Missing for dashboard-reachable boot

1. Kernel modules (storage + network) in the live initramfs
2. Image initrd swapped to the live initramfs
3. Runtime boot verification (VM or real hardware)
4. Network bring-up beyond loopback (so a browser on another machine can reach it)

## Missing for USB creator

1. Host USB writer tool
2. Secure Boot signing (UEFI dual-mode packaging is the baseline; unsigned)

## References

- `distro/install-media/bootability-contract.md`
- `distro/install-media/bootability-manifest.yaml`
- `scripts/compose-live-rootfs.sh`
- `docs/ops/install-media-runtime-boot.md` (Slice 9.14 first-boot wiring)
- `docs/ops/install-media-wrapper.md` (Slice 9.13 firmware image)
