# Install-media runtime boot + first-boot dashboard path

Operator and engineering contract for **Block 9 / Slice 9.14** — runtime boot
verification of the firmware image and the first-boot dashboard wiring.

## Purpose

Two concerns, one slice:

1. **Runtime boot verification** — prove the Slice 9.13 firmware image actually
   boots, when a VM/boot harness exists. Honest skip (with a concrete reason)
   when it does not.
2. **First-boot dashboard path** — the smallest viable wiring so a booted live
   system launches `vyntrio-api` and serves the dashboard at
   `http://<host>:8080`.

This is minimal first-boot progress, not broad appliance features. It does **not**
implement the USB creator, UEFI boot, target-disk install, or a live-rootfs
userland.

## Command

```bash
make install-media-runtime        # boot-test (if VM) + wire first-boot path
make test-install-media-runtime   # verify runtime_boot_tested honesty + wiring
```

Depends on `make install-media-wrap` (Slice 9.13 firmware image).

## Runtime boot harness

- Detects `qemu-system-x86_64` / `qemu-system-i386` (and `/dev/kvm`).
- If present: boots the raw image headless
  (`-drive file=…,format=raw -m 512 -no-reboot -nographic -serial file:…`),
  bounded by a timeout, then scans the serial log for GRUB and kernel markers
  (`GRUB`, `Linux version`, `Decompressing Linux`, `Booting`).
- Records `runtime_boot_tested: true` + `runtime_boot_result` +
  `boot_markers` + the serial log path.
- If absent: `runtime_boot_tested: false` with a concrete `runtime_boot_reason`.

**On this host:** `runtime_boot_tested: false`,
`runtime_boot_reason: no_vm_harness: qemu-system-x86_64/i386 not installed and /dev/kvm absent`.
No boot is faked; the harness will produce a real result on a VM-capable host.

## First-boot dashboard path (wiring)

| File | Role |
|------|------|
| `live_root/usr/lib/vyntrio/firstboot.sh` | Brings up loopback, launches `vyntrio-api --config /etc/vyntrio/config.toml`, prints the dashboard URL |
| `live_root/etc/systemd/system/vyntrio-firstboot.service` | Runs the entrypoint on first boot (`WantedBy=multi-user.target`, `ConditionPathExists=/usr/bin/vyntrio-api`) |
| `live_root/usr/lib/vyntrio/live-init.sh` | Invokes the first-boot entrypoint as its dashboard step |

The dashboard is `vyntrio-api`'s embedded UI on `127.0.0.1:8080` (from
`config.toml`).

## Honesty model

`dashboard_reachable: false` today. The entrypoint **no-ops** when
`/usr/bin/vyntrio-api` is absent and says so. Reaching the dashboard on first boot
still needs:

1. A **live rootfs userland** in the image — `vyntrio-api` and `busybox` are
   **dynamically linked**, so a real userland (libc + loader + binaries) must be
   composed into the initrd/rootfs.
2. A **boot environment** (VM or real hardware) to run it.
3. **Runtime boot verification** reaching the dashboard (needs 1 + 2 + a VM
   harness).

The current firmware image boots **kernel + initrd only**; the first-boot wiring
is the contract for when the live userland lands.

## Provenance (`build/RUNTIME_BOOT.txt`)

Key fields: `firmware_bootable`, `image`, `runtime_boot_tested`,
`runtime_boot_result`, `runtime_boot_reason` (when skipped), `vm_harness` +
`boot_markers` + `serial_log` (when run), and the `first_boot_path` block
(`entrypoint`, `service`, `dashboard_url`, `dashboard_reachable`,
`reachable_requires`).

## Missing for USB creator / first boot

1. Live rootfs userland (so first boot actually reaches the dashboard)
2. Runtime boot verification reaching the dashboard (needs a VM harness here)
3. Secure Boot signing (UEFI dual-mode packaging is the baseline; unsigned)
4. Host USB writer tool

## References

- `distro/install-media/bootability-contract.md`
- `distro/install-media/bootability-manifest.yaml`
- `scripts/verify-runtime-boot.sh`
- `docs/ops/install-media-wrapper.md` (Slice 9.13 firmware image)
- `docs/ops/install-media-image.md` (Slice 9.12 boot chain)
