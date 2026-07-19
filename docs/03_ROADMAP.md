# Roadmap

## Active Development Line (frozen scope — one track only)

Vyntrio ships along a **single ordered product line**, not multiple parallel
tracks. Exactly **one** stage is active at a time, and each stage is gated on the
previous one being **stable**. This is the only sanctioned sequence of work.

| # | Stage | Gate to start | Status |
|---|-------|---------------|--------|
| 1 | **Boot/media chain completion** — boot the live runtime to a shell/first-boot on real firmware | — | **Complete** |
| 2 | **First-boot local dashboard stability** — `vyntrio-api` reachable in a local browser after boot | Stage 1 boot-to-runtime stable | **Active** |
| 3 | **Storage/NAS UI foundation** (after boot) | Stage 2 dashboard stable on boot | **Active (foundation)** |
| 4 | **Remote Connect foundation** | Stage 3 local dashboard solid | Gated |
| 5 | **Licensing foundation** | Stage 4 remote/local path stable | Gated |

**Current position:** Stage 2 (active). **Stage 1 is complete**: real boot chain,
firmware-bootable BIOS image, first-boot wiring, chroot-verified live userland,
live-initramfs hardware enablement (Slice 9.16), and the image-initrd swap to the
Vyntrio live initramfs (Slice 9.17, sha256-proven). Stage 2 opened with **Slice
10.1 — runtime boot verification** (`make install-media-runtime-verify`): a
**read-only** verifier that boots the image on a VM/hardware harness and proves an
**actual HTTP 200 from the booted guest** dashboard, or fails closed with exact
blockers. On this build host there is no `qemu`/`/dev/kvm`, so it records
`status: blocked` **without modifying** the image/initrd/bootloader/live-rootfs.

**Slice 10.2 — local dashboard stability** made the local WebGUI a dashboard-first,
**supervised** appliance service: `firstboot.sh` now runs under busybox `/bin/sh`
(the prior bash shebang could not exec in the busybox-only live runtime, so the
dashboard never started on boot), the dashboard is kept up via a bounded respawn
loop, and the build-time live-rootfs probe proves `/readyz` (DB-ready), not just
the static UI. `dashboard_reachable` stays false without a boot harness.

**Slice 10.3 — first-boot local onboarding clarity** adds a clear boot → local
dashboard → sign-in / create owner → read-only overview prompt, both in the live
runtime (`firstboot.sh` banner) and on the customer-facing landing page
(`#first-boot-setup`). Storage, services, and remote access are explicitly
deferred in the onboarding text.

**Slice 10.4 — TLS / LAN dashboard bind readiness** removes the TLS **config**
blocker: `vyntrio-api` serves HTTPS when `tls_cert_file` / `tls_key_file` are set;
non-loopback bind still requires `cookie_secure=true`. Runtime TLS prep
(`prepare-live-dashboard-tls.sh`) generates ephemeral certs when the operator sets
`VYNTRIO_LAN_BIND_IP`; loopback HTTP remains the secure default. Records now show
`tls_readiness: ready`. Remote Connect and licensing remain gated.

The remaining Stage-2 need is a boot harness to produce the real
**dashboard-reachable-on-boot** proof.

**Slice 11.1 — Storage/NAS UI foundation (Stage 3, active)** adds a read-only
**Storage inventory** view in the signed-in appliance dashboard, wired to
`GET /api/v1/storage/disks`. Pools, shares, and storage mutations remain deferred.

**Rules (enforced):**
- Only **one** slice in progress, and only within the **current stage**.
- Do **not** start Stage N+1 work until Stage N is stable.
- **No** parallel work in unrelated product areas (storage/remote/license) before
  the boot-to-dashboard path is stable.
- **No** installer refinements and **no** staged-installer feature expansion —
  Block 10 `vyntrio-installer` is frozen internal infrastructure, consumed later,
  not a development track.

The `0.x` phases below are the **gated backlog map** for this single line. They
are sequenced by the table above; they are **not** independent parallel tracks
and must not be worked simultaneously.

## 0.1 Foundation
Repository, Coding Standards, CI, Build, lokale Dev-Umgebung, Basis-Doku, API-Konventionen.

## 0.2 Installer & Base OS  — _Active line Stage 1 (boot/media) → Stage 2 (first-boot dashboard)_

**Priority:** primary product path — bootable live image, boot-to-runtime, first
boot, local dashboard. Host USB creator follows once boot-to-dashboard is stable.
Block 10 staged installer is **frozen** internal infrastructure consumed **after**
boot; it is not a development track and sandbox `vyntrio-installer` paths are not
the operator journey.

Boot chain, firmware-bootable image, live rootfs/initramfs (incl. hardware
enablement), first-boot dashboard path. UEFI/BIOS, disk-layout and unattended
install are later, not parallel.

## 0.3 Core Platform
Auth, RBAC, Settings, Notifications, Logs, Job-System, Health Endpoints.

## 0.4 Dashboard  — _Active line Stage 2 (first-boot local dashboard stability)_
Navigation, Widgets, Systemübersicht, Live-Status, Theme, Layout-System. Gated on
Stage 1 boot-to-runtime being stable.

## 0.5 Storage  — _Active line Stage 3 (storage/NAS UI foundation, after boot)_
Disks, SMART, Pools, Shares, Snapshots, Scrub, Restore, Backup Policies. Gated on
Stage 2 local dashboard being stable on boot; not started before then.

## 0.6 Containers
Docker Host Integration, Compose Stacks, Volumes, Logs, Templates, App Deployments.

## 0.7 Virtualization
VM Creation Wizard, ISO Library, Snapshots, Console, Passthrough-Basis.

## 0.8 Vyntrio Hub
Marketplace, Kategorien, Paketsignaturen, Ratings, Plugin Lifecycle.

## 0.9 Commercial  — _Active line Stage 5 (licensing foundation, last)_
Lizenzportal, Aktivierungen, Offline-Files, **Geräte-/Server-Tier-Bindung** (nicht
fragmentierte Per-Feature-Gates), Billing Hooks. Gated on Stage 4 remote/local
path being stable; Remote Connect foundation (Stage 4) precedes this.

## 1.0 Release
Stabilitätsphase, Beta-Programm, Security Hardening, Upgrade-Tests, Launch-Doku.
