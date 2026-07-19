# Install-media initrd swap (Vyntrio live initramfs)

Operator and engineering contract for **Block 9 / Slice 9.17** — swapping the
firmware image's initrd to the hardware-enabled Vyntrio live initramfs so GRUB
boots the **Vyntrio live runtime** instead of the host Debian initrd.

## Purpose

Through Slice 9.16 the firmware-bootable BIOS image booted the **host Debian
initrd** (staged by Slice 9.12) — not the Vyntrio live runtime. That was the last
open gap in the Stage-1 boot/media chain. This slice replaces `boot/initrd.img`
with the live initramfs (`build/vyntrio-live-initramfs.cpio.gz`, produced by
Slices 9.15/9.16) and rebuilds the firmware image so it boots the Vyntrio
initramfs and its `/init` (PID 1).

Minimal increment only — just the initrd swap and image rebuild. No installer/
apply changes, no target-disk writes, no new Stage-2+ work.

## Command

```bash
make install-media-initrd-swap        # swap initrd + rebuild image + prove the swap
make test-install-media-initrd-swap   # independently re-prove + honesty + fail-closed
```

Depends on `make install-media-hardware` (Slice 9.16). Runs the existing Slice
9.13 wrapper to rebuild the image, so it reuses that tested build + structural
verification path.

## What it does

1. **Verifies** the live initramfs really is the Vyntrio live runtime: `/init` at
   root, `busybox`, and kernel modules (`lib/modules/<kver>/modules.dep`).
2. **Backs up** the current host initrd to `build/initrd.host-backup.img` (records
   its size + sha256 for provenance).
3. **Swaps**: `boot/initrd.img` ← `build/vyntrio-live-initramfs.cpio.gz`.
4. **Rebuilds** the firmware image via `scripts/wrap-install-media-image.sh`.
5. **Proves the swap** without booting: extracts `/boot/initrd.img` back out of the
   raw image (`debugfs dump` on the carved partition) and matches its **sha256** to
   the live initramfs — and confirms it differs from the host initrd.
6. **Records** `build/INITRD_SWAP.txt`.

## Verification — what it proves, and what it does not

Proven on the build host, no VM required:

- The image's `/boot/initrd.img` is **byte-identical** (sha256) to the Vyntrio live
  initramfs and **not** the host Debian initrd.
- The extracted initrd contains `/init` + `busybox` + kernel modules — i.e. it is
  the Vyntrio live runtime, not a Debian initrd.
- The rebuilt image still passes the wrapper's **structural verification** (MBR
  `0x55AA`, GRUB embed, bootable ext2 partition).

**This is not a runtime boot.** There is no qemu/VM here, so
`runtime_boot_tested: false` and `dashboard_reachable_on_boot: false` remain.
The `boot=live` token in `grub.cfg` is a Debian live-boot hint that the Vyntrio
`/init` ignores; `/init` performs the Slice 9.16 hardware bring-up and the Slice
9.14 first-boot path itself.

## Fail-closed / no regression

If the image cannot be rebuilt or the swap cannot be proven (not firmware-
bootable, structural failure, or initrd sha mismatch), the script **restores the
host initrd** from the backup, rebuilds the known-good BIOS image, records
`swap_status: failed` with a concrete blocker, and exits non-zero. The current
BIOS-bootable raw image path is never left regressed.

The swap is **idempotent**: a re-run detects an already-swapped initrd
(`already_swapped_on_entry: true`) and preserves the host backup.

## Boundary and safety

- The live initramfs is the **live runtime**; it stays distinct from `payload/`
  (the target-disk artifact). No target-disk writes, no installer/apply behavior.
- Appliance state/credentials/tokens/licenses/secrets remain forbidden in
  `live_root` and are enforced fail-closed upstream (9.15/9.16).

## What remains to close Stage 1

1. **Runtime boot verification** on a VM or real hardware (no qemu here) — boot the
   image and confirm it reaches the Vyntrio live runtime.
2. **Dashboard-reachable-on-boot proof** — confirm `vyntrio-api` answers in a local
   browser after a real boot (Stage 2 begins once this is stable).

## What remains for USB creator

1. Secure Boot signing (UEFI dual-mode packaging is the baseline; unsigned)
2. Host USB writer tool.

## References

- `distro/install-media/bootability-contract.md`
- `distro/install-media/bootability-manifest.yaml`
- `scripts/swap-live-initramfs-into-image.sh`
- `scripts/wrap-install-media-image.sh` (Slice 9.13 image build reused here)
- `docs/ops/install-media-hardware-enable.md` (Slice 9.16 live initramfs)
- `docs/ops/install-media-live-rootfs.md` (Slice 9.15 userland)
