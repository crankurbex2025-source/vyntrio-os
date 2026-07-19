# Install-image bootability initialization contract

- **Status:** Firmware-bootable BIOS image that now boots the hardware-enabled **Vyntrio live initramfs** (Slice 9.17, sha256-proven) — **runtime boot not yet verified (no VM); dashboard not yet proven reachable on boot**
- **Block:** 9
- **Slices:** 9.9 (declarative contract), 9.11 (executable foundation), 9.12 (real boot chain), 9.13 (firmware wrapper), 9.14 (runtime boot + first-boot path), 9.15 (live-rootfs userland), 9.16 (live-initramfs hardware enablement), 9.17 (image initrd swapped to live initramfs)
- **Authority:** `docs/ADR/0006-appliance-install-recovery-media.md` (install media
  sections B, D). Layer inventory: `distro/install-media/bootability-manifest.yaml`.
  Envelope contract: `distro/install-media/envelope-contract.md`.
  Envelope assembly: Slice 9.8 (`make install-media-envelope`).
  Executable: `make install-media-bootability` (9.11), `make install-media-image` (9.12),
  `make install-media-wrap` (9.13), `make install-media-runtime` (9.14),
  `make install-media-live-rootfs` (9.15), `make install-media-hardware` (9.16),
  `make install-media-initrd-swap` (9.17).

## 1. Explicit status

**Slice 9.9:** defines the **bootable initialization boundary** (declarative).

**Slice 9.11 (implemented):** `make install-media-bootability` replaces deferred
`boot/LAYER.txt` and `live_root/LAYER.txt` with **honest stubs**, validates the
envelope layout, writes `BOOTABILITY.txt`, and emits a **non-bootable** tar stub
under `distro/install-media/build/`.

**Slice 9.12 (implemented):** `make install-media-image` replaces the `boot/*.stub`
placeholders with a **real boot chain** — a real Linux kernel, a real initrd, a
real GRUB core image (`grub-mkimage`), and a real `grub.cfg` — verifies them
(`grub-file --is-x86-linux`, initrd magic), expands `live_root/` toward a minimal
runtime (`os-release`, `fstab`, structured `live-init.sh`), and emits the closest
image the host tooling allows. See `docs/ops/install-media-image.md`.

| Capability | Status |
|------------|--------|
| Real kernel + initrd in boot layer | **Implemented (verified)** |
| Real GRUB core + `grub.cfg` | **Implemented (9.12)** |
| Live root minimal runtime additions | **Implemented (9.12)** |
| Stub-vs-real validation + honest fallback | **Implemented (9.12)** |
| Firmware-bootable ISO (when writer present) | **Conditional — emitted only with `xorriso`/`genisoimage`** |
| Host USB creator | **Deferred** |

**Slice 9.13 (implemented):** `make install-media-wrap` wraps the real boot chain
into a firmware-bootable image. It tries, in order: (A) an ISO via `grub-mkrescue`
+ an ISO9660 writer; (B) a real **MBR/BIOS raw disk image** built with
`mke2fs` + `debugfs` + `sfdisk` + `grub-bios-setup` (the same method
`grub-install` uses); or (C) an honest, fully documented fallback. See
`docs/ops/install-media-wrapper.md`.

| Capability | Status |
|------------|--------|
| Raw MBR/BIOS bootable image (`grub-bios-setup`) | **Implemented** as dual-mode hybrid fallback component (EF02 + i386-pc) |
| Dual-mode GPT hybrid (BIOS + UEFI) | **Required product baseline** — `vyntrio-install-media.img` |
| Structural verification (MBR 0x55AA, GRUB embed, bootable ext2) | **Implemented (9.13)** |
| Three-axis validation (boot_chain / firmware_bootable / wrapper) | **Implemented (9.13)** |
| Honest fallback with per-strategy blockers | **Implemented (9.13)** |
| Firmware-bootable ISO on this host | **Blocked — no `xorriso`/`genisoimage`** |
| UEFI / `x86_64-efi` target | **Required baseline** — dual-mode hybrid `raw_gpt_hybrid_disk` with ESP/`BOOTX64.EFI` |
| BIOS-only raw MBR | **Incomplete** — fallback only; must not be staged as product primary |
| Runtime boot verification (actually booting the image) | **Best-effort** — qemu/OVMF when present; not claimed without markers |

**What "firmware_bootable: true" means here:** a real, standards-compliant
BIOS-bootable image was **constructed** by the standard GRUB toolchain and passed
**structural** verification. `runtime_boot_tested: false` is recorded explicitly —
the image has not been booted in a VM (no qemu on the build host). This is not a
claim that it was booted; it is not faked.

**Raw-image prerequisites:** `grub-bios-setup` requires a **non-tmpfs** build
directory (it must canonicalize a disk-backed device). On a tmpfs checkout the
wrapper falls back honestly and records the exact blocker.

**Slice 9.14 (implemented):** `make install-media-runtime` adds a **gated
runtime-boot harness** and wires the **first-boot dashboard path**. When a VM
(`qemu-system-x86_64`/`i386`) is present it boots the raw image headless, captures
serial output, and scans for GRUB/kernel markers. It also lays down the smallest
first-boot wiring into `live_root`: an entrypoint (`usr/lib/vyntrio/firstboot.sh`)
that launches `vyntrio-api` (the dashboard, `http://<host>:8080`) and a systemd
unit (`etc/systemd/system/vyntrio-firstboot.service`). See
`docs/ops/install-media-runtime-boot.md`.

| Capability | Status |
|------------|--------|
| Runtime-boot harness (qemu when present) | **Implemented (9.14)** |
| First-boot dashboard entrypoint + service | **Implemented (9.14, wiring)** |
| `runtime_boot_tested` honesty in provenance | **Implemented (9.14)** |
| Runtime boot on this host | **Skipped — no qemu/VM, `/dev/kvm` absent** |
| Dashboard reachable on first boot | **Deferred — needs live userland + boot env** |

**Why runtime boot is not exercised here:** no `qemu`/`kvm`/`firecracker` and no
`/dev/kvm`. The harness runs and produces a real result where a VM exists;
`RUNTIME_BOOT.txt` records `runtime_boot_tested: false` with the concrete reason.
`vyntrio-api` and `busybox` are **dynamically linked**; Slice 9.15 composes that
userland. No boot is faked; the current image boots kernel+initrd only.

**Slice 9.15 (implemented):** `make install-media-live-rootfs` composes the
smallest userland that can actually start the first-boot path — the dynamic loader
and full `ldd` closure, `busybox` (shell + core utils), `vyntrio-api` (the live
dashboard runtime), and an `/init` that hands off to `firstboot.sh` — and emits a
live initramfs (`vyntrio-live-initramfs.cpio.gz`). See
`docs/ops/install-media-live-rootfs.md`.

| Capability | Status |
|------------|--------|
| Loader + libc closure in `live_root` | **Implemented (9.15)** |
| busybox shell + core utilities | **Implemented (9.15)** |
| `vyntrio-api` as live dashboard runtime | **Implemented (9.15)** |
| `/init` -> `firstboot.sh` handoff | **Implemented (9.15)** |
| Live initramfs artifact (cpio.gz, `/init` at root) | **Implemented (9.15)** |
| Userland executes (chroot) | **Verified — busybox sh + vyntrio-api run** |
| Dashboard serves from this userland | **Verified — HTTP 200 via chroot probe** |
| Image initrd swapped to live initramfs | **Deferred — no kernel modules, no VM to verify** |
| Dashboard reachable on boot | **Deferred — see below** |

**What the 9.15 verification does and does not prove.** It proves the composed
userland **runs** and that `vyntrio-api` **serves the embedded dashboard (HTTP
200)** from it, using `chroot` execution plus an HTTP probe. That is **not a
boot**. The image still boots kernel+initrd only; `dashboard_reachable_on_boot`
stays **false**.

**Live-root exclusion boundary (refined in 9.15).** `preinstalled_target_disk_payloads`
forbids staging `payload/` onto a target disk from `live_root`. It does **not**
forbid the live session's own dashboard runtime: `live_root/usr/bin/vyntrio-api`
is the **live** runtime, distinct in role from `payload/usr/bin/vyntrio-api` (the
target-disk artifact). Appliance state, credentials, tokens, licenses and secrets
remain forbidden in `live_root` and are enforced fail-closed. The dashboard HTTP
probe runs against a **throwaway chroot copy** precisely so the database
`vyntrio-api` creates never lands on shipped media.

**Slice 9.16 (implemented):** `make install-media-hardware` adds the smallest
viable hardware bring-up to the live initramfs so a real boot can reach the live
runtime more reliably. It stages a dep-resolved **storage + network kernel-module
set** (matched to the boot chain's host kernel) into `live_root/lib/modules/<kver>/`,
regenerates `modules.dep` with `depmod`, writes an ordered `modules.load`, wires a
module-load + **DHCP** bring-up (`hw-init.sh` + busybox `udhcpc`) into `/init`
ahead of the first-boot path, and **re-emits** the live initramfs carrying the
modules. See `docs/ops/install-media-hardware-enable.md`.

| Capability | Status |
|------------|--------|
| Storage/network kernel modules staged (dep-resolved) | **Implemented (9.16)** |
| `modules.dep` regenerated in `live_root` (`depmod -b`) | **Implemented (9.16)** |
| Ordered `modules.load` referencing staged `.ko` | **Implemented (9.16)** |
| Networking beyond loopback (busybox `udhcpc`) | **Implemented (9.16, wiring)** |
| `/init` runs `hw-init.sh` before `firstboot.sh` | **Implemented (9.16)** |
| Live initramfs re-emitted with modules | **Implemented + verified (contains `modules.dep`)** |
| Sample module validity (`modinfo`) | **Verified (9.16)** |
| Modules loaded on build host | **No — staged only; `modprobe` would target the host kernel** |
| Image initrd swapped to live initramfs | **Deferred — blocked on runtime verification** |
| Dashboard reachable on boot | **Deferred — no VM/boot** |

**What 9.16 does and does not prove.** Modules are **staged**, dependency-resolved,
metadata-regenerated, and `modinfo`-validated, and the re-emitted initramfs is
verified to contain them. Modules are **not loaded** here: running `modprobe` on
the build host would load into the build-host kernel, which is out of scope and
unsafe. At 9.16 `wired_into_image_initrd`, `boot_verified`, and
`dashboard_reachable_on_boot` all stay **false**. This slice supersedes the 9.15
gap `kernel_modules_for_storage_and_network_in_initramfs`.

**Slice 9.17 (implemented):** `make install-media-initrd-swap` closes the last
Stage-1 boot/media gap: it replaces `boot/initrd.img` (the host Debian initrd
staged at 9.12) with the hardware-enabled **Vyntrio live initramfs** and rebuilds
the firmware image via the existing 9.13 wrapper, so GRUB now boots the Vyntrio
live runtime and its `/init`. See `docs/ops/install-media-initrd-swap.md`.

| Capability | Status |
|------------|--------|
| Image boots the Vyntrio live initramfs (not host Debian initrd) | **Implemented (9.17)** |
| Swap proven by sha256 of initrd extracted back out of the image | **Verified (`debugfs dump`)** |
| Extracted initrd carries `/init` + busybox + kernel modules | **Verified (9.17)** |
| Firmware image structure preserved (MBR/GRUB/partition) | **Verified — structural pass** |
| Fail-closed restore of host initrd on any failure | **Implemented (9.17)** |
| Idempotent re-run | **Implemented (9.17)** |
| Runtime boot of the swapped image | **Deferred — no qemu/VM here** |
| Dashboard reachable on boot | **Deferred — needs a real boot** |

**What 9.17 does and does not prove.** It proves — by sha256 of the initrd
extracted back out of the image — that the firmware image now carries the Vyntrio
live initramfs and **not** the host Debian initrd, and that the image still passes
structural verification. It is **not a runtime boot**: no qemu/VM exists here, so
`runtime_boot_tested` and `dashboard_reachable_on_boot` stay **false**. If the
swapped image cannot be built and proven, the host initrd is **restored** and the
known-good image rebuilt (no regression). This closes the deferred
`swap_image_initrd_to_live_initramfs` gap and **completes Stage 1**.

**Stage 2 / Slice 10.3 (implemented — first-boot local onboarding clarity):**
The first-boot path now prints a clear local setup prompt to the operator: boot
-> local browser dashboard at `http://<host>:8080` -> sign in or create owner
-> read-only overview; storage, services, and remote access are explicitly
deferred. The same flow is recorded in `FIRST_BOOT.txt` and `RUNTIME_BOOT.txt`
(`first_boot_onboarding_clarity: clear`, `local_onboarding_steps`) and mirrored
on the customer-facing landing page (`#first-boot-setup`). No new product areas
are started; the landing page change is tied directly to the product-side
onboarding truth. See `docs/ops/install-media-dashboard-stability.md`.

**Stage 2 / Slice 10.2 (implemented — local dashboard stability):**
The local browser dashboard (WebGUI) is the **primary** management surface, so
the first-boot path now keeps it running like an appliance service. Three honest,
in-scope changes: (1) `firstboot.sh` runs under `/bin/sh` (busybox) — the prior
`#!/usr/bin/env bash` could not exec in the busybox-only live runtime, so the
dashboard never started on boot; (2) `vyntrio-api` is **supervised** via a bounded
respawn loop so the WebGUI stays up; (3) the build-time live-rootfs probe now
proves `/readyz` (backend/DB ready), not just `/` (static UI). Honest limits:
LAN exposure is **blocked on TLS** (`cookie_secure` requires TLS for a
non-loopback bind), so the dashboard binds loopback and `dashboard_reachable`
stays `false` until a booted-guest HTTP proof exists (no qemu here). No installer/
target writes; the boot/image chain is unchanged (first-boot wiring only). See
`docs/ops/install-media-dashboard-stability.md`.

**Stage 2 / Slice 10.1 (implemented; blocked on this host):**
`make install-media-runtime-verify` is a **read-only** verifier that boots the
initrd-swapped image on a VM/hardware harness and proves the dashboard answers
over **HTTP from the booted guest** (qemu user-net hostfwd) — or **fails closed**
with exact blockers when no harness exists. It never modifies the image, initrd,
bootloader, or live-rootfs. On this build host there is no `qemu`/`/dev/kvm`, so
it records `status: blocked` with exact blockers and `chain_modified: false`. A
dashboard-on-boot proof requires an actual HTTP 200 from a booted guest; chroot/
structural checks (Slices 9.15/9.16) do **not** count. This is the first Stage-2
slice and is unrelated to the staged-installer Block 10. See
`docs/ops/install-media-runtime-verify.md`.

**Bootable initialization is not installation.** Populating boot/live layers and
emitting an image does not copy `payload/` to a target disk, partition,
enable services, or run ADR-0004 bootstrap.

**USB-first product path:** this slice is the first executable step toward
primary delivery. It does **not** complete USB creator or first boot.

Recovery-image bootability remains a **separate** future contract under
`distro/recovery-media/`.

## 2. What bootable initialization adds

| Step | Beyond 9.8 envelope | Status |
|------|---------------------|--------|
| `boot/` population | Bootloader config stub, kernel stub, initrd stub | **Stub (9.11)** |
| `live_root/` population | Directory skeleton + `live-init.sh` stub | **Stub (9.11)** |
| Layout validation | Verify boot/live_root/payload before emission | **Stub (9.11)** |
| Image emission | Tar provenance stub (`*-NOT-BOOTABLE.stub.tar`) | **Stub (9.11)** |
| Real bootable ISO/USB | `xorriso` / GRUB / real rootfs | **Deferred** |

The `payload/` layer from Slice 9.6/9.8 is **carried forward unchanged**.
Bootability does not add secrets, persistent DB, or target-disk state.

## 3. Relationship to staging and envelope assembly

```
Slice 9.6              Slice 9.8                 Slice 9.11
─────────              ─────────                 ──────────
staging/payload/  →    envelope/payload/    →    carried forward
                       envelope/boot/ (LAYER) →  boot stubs
                       envelope/live_root/    →  live_root stubs
                                                 → build/*-NOT-BOOTABLE.stub.tar
```

| Artifact | Location | Committed | Role |
|----------|----------|-----------|------|
| Staged payloads | `distro/install-media/staging/payload/` | **No** | Slice 9.6 disposable input |
| Assembled envelope | `distro/install-media/envelope/` | **No** | Slice 9.8/9.11 disposable tree |
| Bootability manifest | `distro/install-media/bootability-manifest.yaml` | **Yes** | Declarative + status |
| Image stub | `distro/install-media/build/` | **No** | Non-bootable emission artifact |

## 4. Bootability vs installer execution

| Concern | Bootable initialization | Installer (Block 10, secondary) |
|---------|-------------------------|----------------------------------|
| Purpose | Prepare media layers / image stub | Copy payloads to target disk |
| Operates on | Image envelope layers | Target disk partitions/paths |
| Primary product path | **Yes (USB-first foundation)** | Secondary / live-session infra |

## 5. Still missing for USB creator

- Real kernel and initrd binaries
- UEFI/BIOS bootloader configuration that firmware can load
- Complete live root filesystem
- Firmware-bootable ISO or raw disk image
- Host-side USB writer tool (Linux-first per ADR-0007 §J)

## 6. Related documents

- `docs/ops/install-media-bootability.md` — operator contract for Slice 9.11
- `docs/ops/install-media-live-rootfs.md` — operator contract for Slice 9.15
- `docs/ops/install-media-hardware-enable.md` — operator contract for Slice 9.16
- `docs/ops/install-media-initrd-swap.md` — operator contract for Slice 9.17
- `docs/ops/install-media-runtime-verify.md` — operator contract for Stage 2 / Slice 10.1
- `distro/install-media/bootability-manifest.yaml`
- `scripts/initialize-install-media-bootability.sh`
- `scripts/enable-live-initramfs-hardware.sh`
- `scripts/swap-live-initramfs-into-image.sh`
- `scripts/verify-live-boot-runtime.sh`
- `docs/00_PROJECT.md` — USB-first primary path
- `docs/ADR/0006-appliance-install-recovery-media.md`
- `docs/ADR/0007-appliance-installer-contract.md`
