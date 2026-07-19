# Install-media runtime boot verification (Stage 2)

Operator and engineering contract for **Stage 2 / Slice 10.1** — proving the
structurally-bootable, initrd-swapped image (Stage 1 / Slice 9.17) actually
**boots** on a real harness and reaches the first-boot dashboard over HTTP, or
failing closed with exact blockers when no harness exists.

> "Slice 10.1" is the first Stage-2 (first-boot local dashboard) slice. It is
> unrelated to the staged-installer **Block 10**, which stays secondary internal
> infrastructure.

## Purpose

Stage 1 produced a firmware-bootable image that carries the Vyntrio live
initramfs (proven by sha256 extraction, but **not booted**). Stage 2 begins by
turning that into a **real runtime proof**: boot the image on a VM/hardware
harness, capture serial output, and get an **actual HTTP response** from the
booted guest's dashboard. If the environment cannot boot it, the slice records
exact blockers and fails closed — it does **not** compensate by changing anything.

## Command

```bash
make install-media-runtime-verify        # boot on a harness for real proof; fails closed w/o one
make test-install-media-runtime-verify   # validate record honesty + read-only (no chain mutation)
```

`install-media-runtime-verify` **fails closed (non-zero)** when no harness exists,
so the `test-` target intentionally depends on `install-media-initrd-swap` (the
build), not on the verify target succeeding.

Overrides: `VYNTRIO_RUNTIME_HTTP_PORT` (host-forwarded port, default `8088`),
`VYNTRIO_RUNTIME_BOOT_TIMEOUT` (seconds, default `120`).

## Read-only contract (hard rule)

This verifier is **read-only** with respect to the boot/image chain and the live
rootfs. It never writes `envelope/boot`, `envelope/live_root`, or the image
artifacts. **If no harness exists it changes nothing** — it records blockers and
exits non-zero. The test enforces this by fingerprinting (sha256) the kernel,
initrd, `grub.cfg`, the raw image, and the entire `live_root` tree **before and
after** running the verifier and asserting they are unchanged.

## What counts as a dashboard-on-boot proof

Only an **actual HTTP 200 from the booted guest**. The verifier boots the raw
image under qemu with user-mode networking and a host port-forward
(`hostfwd=tcp:127.0.0.1:<port>-:8080`), then polls `http://127.0.0.1:<port>/`
for HTTP 200 while the guest boots. A chroot probe or a structural inference (as
used in Slices 9.15/9.16 to verify the *userland*) is **not** a dashboard-on-boot
proof and is never recorded as one.

## Outcomes (`status` in `build/RUNTIME_VERIFY.txt`)

| status | meaning | exit |
|--------|---------|------|
| `dashboard_reachable` | booted guest answered HTTP 200 on :8080 | 0 |
| `boot_without_dashboard` | booted (serial markers) but no HTTP 200 in time | non-zero |
| `no_boot_markers` | harness ran but no GRUB/kernel/live-init markers | non-zero |
| `blocked` | no VM/hardware harness — exact blockers recorded | non-zero |
| `stage1_regression` | image no longer boots the live initramfs | non-zero |

`dashboard_reachable_on_boot: true` is only ever set by a real HTTP 200.

## Current environment result

No `qemu-system-x86_64/i386` and no `/dev/kvm` on this build host →
`status: blocked`, `runtime_boot_tested: false`,
`dashboard_reachable_on_boot: false`, and exact blockers. The image, initrd,
bootloader, and live-rootfs were **not** modified.

## What remains missing for dashboard-on-boot proof

1. A **VM or hardware boot harness** (e.g. `qemu-system-x86` on the build host, a
   CI runner with nested virt, or a bare-metal boot+serial rig).
2. A **booted-guest HTTP 200** from `vyntrio-api` on :8080 through the harness
   network path (requires the live boot to bring up NIC + DHCP + firstboot).

## What remains for USB creator (later, gated)

1. Secure Boot signing (UEFI dual-mode packaging is the baseline; unsigned)
2. Host USB writer tool.

## Safety and scope

- No installer/apply behavior, no target-disk writes.
- The boot/image chain is not modified by this slice (read-only verification).
- Stage 2+ storage/remote/licensing work is **not** started here.

## References

- `distro/install-media/bootability-contract.md`
- `distro/install-media/bootability-manifest.yaml`
- `scripts/verify-live-boot-runtime.sh`
- `docs/ops/install-media-initrd-swap.md` (Stage 1 / Slice 9.17)
- `docs/ops/install-media-runtime-boot.md` (Slice 9.14 first-boot wiring)
- `docs/03_ROADMAP.md` (single active development line)
