# TrueNAS / Unraid product-class parity audit (2026-07-17)

Functional benchmark only. No proprietary branding, assets, code, or wording copied.
Classifications are strict: docs/tests/placeholders are never “real”.

| Label | Meaning |
|-------|---------|
| **real** | Implemented, user-visible or downloadable, verified |
| **partial** | Meaningful usable subset; major class behavior missing |
| **stub/demo/test-only** | Scaffold without product capability |
| **blocked** | Path exists but cannot complete on this host |
| **absent** | No product implementation |

---

## Capability matrix (27)

| # | Capability | Benchmark class provides | Vyntrio today | Class | Evidence | Gap |
|---|------------|--------------------------|---------------|-------|----------|-----|
| 1 | Public download page | Version, size, checksum, status, writers | Live `/download` + `GET /api/v1/public/install-media` | **real** | `PublicDownloadView.tsx`, `installmediapublic` | No CDN |
| 2 | Install artifact availability | Bootable ISO/USB image | Dual-mode GPT hybrid `vyntrio-install-media.img` (BIOS+UEFI) when staged | **partial** | `distro/release/staging/`, `/release/*`, `WRAPPER.txt` `dual_mode: true` | No ISO/persistent install; runtime boot not always proven |
| 3 | Install artifact metadata | Version, size, SHA-256, status | Live metadata + sidecars | **partial** | `release-manifest.json`, public API | Unsigned |
| 4 | GUI media creator Windows | Downloadable USB creator .exe | **local web GUI** `.exe` | **partial** | `mediacreator`, `vyntrio-media-creator-windows-amd64.exe` | Not Win32/Qt native; unsigned |
| 5 | GUI media creator macOS | .dmg / app | Terminal **helper binaries** only (`.app.zip` withdrawn) | **partial** | packaging script | No .dmg/notarization |
| 6 | GUI media creator Linux | AppImage/.deb/GUI | **`.tar.gz`** with binary + .desktop | **partial** | packaging script | No AppImage (no appimagetool) |
| 7 | First boot working runtime | Boot → admin reachable | Wiring present; dashboard unproven | **blocked** | `RUNTIME_BOOT.txt` `dashboard_reachable: false`, `no_vm_harness` | No qemu/KVM |
| 8 | First-login / setup | Owner wizard | Bootstrap + login/logout/CSRF | **real** | `identity/*` | No multi-step wizard |
| 9 | Web admin shell | Full NAS nav | Overview / Settings / Storage | **partial** | `ApplianceApp.tsx` | Narrow vs class |
| 10 | Disk inventory | Disks, SMART, replace | Classified inventory + eligibility | **partial** | `storageinventory` | No SMART/wipe/replace |
| 11 | Pool create/manage | Format/create pools | Declared pools only (`disk_format_applied: false`) | **partial** | `storagepool`, Storage UI | No on-disk format |
| 12 | Dataset/volume layer | Datasets/zvols | Dataset **plans** only | **partial** | POST datasets | No filesystem |
| 13 | Share create/manage | Live shares | Share **plans** (`protocol: planned`) | **partial** | POST shares | No publish |
| 14 | SMB | smbd | None | **absent** | — | Missing |
| 15 | NFS | nfsd | None | **absent** | — | Missing |
| 16 | iSCSI | Targets | None | **absent** | — | Missing |
| 17 | ACL / permissions | FS ACL editors | API RBAC only | **absent** (FS) | RBAC middleware | No FS ACL |
| 18 | Users/groups | Multi-user directory concepts | Owner + roles; no CRUD UI | **partial** | identity | No users UI |
| 19 | Snapshots | Snapshot UI/tasks | None | **absent** | — | Missing |
| 20 | Replication/backup | Dataset replication | Appliance **state** backup CLI only | **partial** | `cmd/backup` | Not NAS data backup |
| 21 | Apps/containers | Catalog | None | **absent** | docs aspirational | Missing |
| 22 | VMs | Hypervisor UI | None | **absent** | — | Missing |
| 23 | Settings/network/ops | Broad config | Display name + net presence read | **partial** / network config **stub** | settings, netpresence | No network config |
| 24 | Update/release flow | In-product updates | Staging/verify; empty update-agent | **stub** (agent) / **partial** (staging) | `cmd/update-agent` | No update API |
| 25 | Website/product IA | Hero, pillars, use cases, download | Landing + download IA present; honesty updated | **partial** | public surfaces | Product depth still thin |
| 26 | Remote/cloud | Remote access | Deferred | **absent** | roadmap | Correctly absent |
| 27 | Licensing/edition | Edition gates | Docs only | **absent** | `docs/15_LICENSE.md` | No product licensing |

---

## Scoring

Strict row tally: **real 2** (#1, #8) · **partial ~14** · **blocked 1** (#7) · **stub ~2** · **absent ~8**.

**Weighted product-class parity: ~14–20%.**

NAS core (pools/datasets/protocols/ACL) remains largely non-operational. Distribution tooling improved this run (GUI packages) but is still honest-partial vs Unraid-class signed native creators.

### Top 10 missing
1. Boot-proven dashboard (`no_vm_harness`)
2. Persistent target-disk install
3. On-disk pool format/create
4. Real datasets/filesystems
5. SMB
6. NFS
7. FS ACL / multi-user NAS identity
8. UEFI image
9. Apps/containers
10. VMs

### Top 5 misleading signals
1. `mutation_available: true` without disk format/protocols
2. Declared pools/share plans that look like storage management
3. Unused RBAC for containers/VMs/licensing
4. Roadmap docs for Docker/VM/network without code
5. Empty `update-agent`

### Top 5 strongest real capabilities
1. Live download metadata + staged BIOS image path
2. Cross-platform media creator GUI packages (local web wizard) + CLI
3. Session auth bootstrap/login/logout + CSRF + RBAC
4. Block inventory with fail-closed eligibility
5. Declared pool/dataset/share plan persistence (foundation only)

---

## Chosen implementation block (this run)

**#1 GUI media creator downloads (Windows / macOS / Linux)** — feasible via embedded local web GUI in the existing Go writer, cross-compiled from this Linux host.

### Why not boot harness / pool format first
Priority order in the brief puts GUI media creator first when feasible. Feasibility proof:
- Same host can cross-compile Go for windows/darwin/linux (`CGO_ENABLED=0`)
- Electron/Tauri/native Cocoa/Qt **not** required for a real GUI flow
- Blockers for “full Unraid-class” packages are explicit: no `.dmg`/notarization (no macOS/Apple), no AppImage (`appimagetool` absent)

### What shipped
- Local loopback web wizard: image load, device list, confirm, progress, verify, boot next-step
- Packages: Windows `.exe`, macOS `.app.zip`, Linux `.tar.gz`
- Download page lists GUI packages first with live size/SHA-256
- Honest: `native_gui: false`, `gui_kind: local_web_loopback`
