# Appliance product gap audit — capability matrix (2026-07-16)

Functional benchmark: Unraid and TrueNAS as **appliance NAS OS** models (bootable
media, local WebGUI, storage, shares, apps/containers, VMs, operations). No
proprietary text, code, or branding is copied from those products.

Legend for **Vyntrio state**:

| Label | Meaning |
|-------|---------|
| **implemented** | Real product capability in repo, wired to UI or runtime, with tests |
| **partial** | Some real behavior exists; critical path incomplete or unproven |
| **stub-demo-test-only** | Structural API/UI exists but always empty or non-functional |
| **blocked** | Design/build exists; verification blocked by environment |
| **not implemented** | No product code path |

---

## Capability matrix

| # | Category | Unraid (functional) | TrueNAS (functional) | Vyntrio evidence | Vyntrio state |
|---|----------|---------------------|----------------------|------------------|---------------|
| 1 | Bootable install/distribution artifact | USB/bootable image; installer flow | ISO/USB image; installer | `Makefile` `install-media-wrap`; `distro/install-media/build/vyntrio-install-media-bios.img` (~136 MB); `WRAPPER.txt` `firmware_bootable: true`; `tests/installmedia/*`; `PublicDownloadView.tsx` shows not published | **partial** — internal BIOS image builds; no public download; no UEFI; installer service not started |
| 2 | First boot into real runtime | Boots to array/storage-ready OS | Boots to pool-ready OS | `compose-live-rootfs.sh`, `firstboot.sh`, `RUNTIME_BOOT.txt` (`runtime_boot_tested: false`, `dashboard_reachable: false`, `no_vm_harness`) | **partial** — live rootfs composed; boot not proven on build host |
| 3 | Local WebGUI/dashboard | Primary management UI on appliance | Web UI administration | `router.go` + embedded SPA (`ui/handler.go`); `/app` → `ApplianceApp`, `OverviewShell`, `SettingsShell`, `StorageShell`; `docs/06_DASHBOARD.md` | **partial** — real authenticated UI; containers/VMs/network admin absent; not boot-proven |
| 4 | First-boot onboarding/setup | Initial array/user setup wizard | Setup wizard / first admin | `firstboot.sh` banner; `POST /api/v1/identity/bootstrap` (loopback); landing `#first-boot-setup`; no setup wizard UI | **partial** — API + docs; no first-owner UI; not boot-proven |
| 5 | Safe LAN access to local WebGUI | HTTPS / trusted LAN access | HTTPS / certificate options | `tls.go`, `prepare-live-dashboard-tls.sh`, `VYNTRIO_LAN_BIND_IP`; default loopback HTTP; `tests/installmedia/tls_readiness_test.sh` | **partial** — TLS prep exists; operator action required; LAN not boot-proven |
| 6 | Storage inventory | Disk listing in WebGUI | Disk/pool device inventory | `storageinventory` collector; `GET /api/v1/storage/disks`; `StorageShell` block list; Go + frontend tests | **implemented** — read-only discovery + UI (host dev; not boot-guest proven) |
| 7 | Storage pool management | Create/manage storage pools | ZFS pool create/expand | `GET /api/v1/storage/pools` → always `pools: []`, `pool_management: not_available`; UI layout section; `pools_test.go` | **stub-demo-test-only** — honest empty read model; no create/expand/delete |
| 8 | Share management | User shares from WebGUI | Datasets/shares | `GET /api/v1/storage/shares` → always `shares: []`, `share_management: not_available`; UI layout section | **stub-demo-test-only** — no share CRUD |
| 9 | Network storage protocols/services | SMB/NFS/AFP etc. | SMB/NFS/iSCSI | `docs/12_NETWORK.md` planned only; `shares.go` `protocol_support: not_available`; no Samba/NFS daemons | **not implemented** |
| 10 | App/container management | Docker app catalog | Apps / k3s / charts | `docs/13_DOCKER.md` spec; RBAC permissions only; no handlers or UI | **not implemented** |
| 11 | VM management | VM create/manage in WebGUI | VM / bhyve / KVM UI | `docs/14_VM.md` spec; RBAC permissions only; no libvirt/QEMU code | **not implemented** |
| 12 | System settings / appliance operations | System pages, updates, power | Services, updates, alerts | `GET/PATCH /api/v1/settings`, overview host metrics; installer CLI preflight (`cmd/installer`); no reboot/network/update UI | **partial** |
| 13 | Remote management/connect | Remote access / cloud link | TrueNAS cloud / VPN options | Stage 4 gated in `docs/03_ROADMAP.md`; `appliance-baseline-status.md` Not started | **not implemented** |
| 14 | Licensing / edition gating | License key / edition features | Enterprise tiers | `docs/15_LICENSE.md`; Stage 5 gated; no license service or UI | **not implemented** |
| 15 | Customer-facing website / product clarity | Marketing + download | Marketing + docs | `/`, `/download`, `/docs`; `docs/22_WEBSITE.md`; live `https://vyntrio.xyz/` | **implemented** — honest about gaps |

---

## Prioritization (Phase B)

Ranked missing capabilities by impact for a serious Unraid/TrueNAS-style appliance:

1. **Boot-proven runtime + dashboard on install media** — blocked (`no_vm_harness`)
2. **Public install/download path** — artifact exists; publication missing
3. **Storage pool mutation (create from eligible disks)** — highest-value unblocked product gap
4. **Share + protocol foundation (SMB/NFS)** — depends on pools
5. **First-boot owner setup UI** — completes onboarding
6. **App/container management** — major pillar; not started
7. **VM management** — major pillar; not started
8. **Operational settings (network, power, updates)** — partial today
9. **Remote Connect** — gated Stage 4
10. **Licensing** — gated Stage 5

**Chosen executable block (this run):** Storage layout foundation beyond read-only
inventory — `GET /api/v1/storage/pools`, `GET /api/v1/storage/shares`, overview
`storage` summary, and Storage/Overview UI sections with **honest empty state** and
`mutation_available: false`. This is the highest-value work that can ship honestly
without inventing pool/share mutation or repeating landing-only work.

**Not chosen:** Landing polish (already strong), Remote Connect, licensing, pool
creation (requires destructive ops design), public download (publication pipeline
outside this slice).
