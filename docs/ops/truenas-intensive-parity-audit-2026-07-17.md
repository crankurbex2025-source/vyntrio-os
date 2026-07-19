# TrueNAS intensive parity audit (2026-07-17)

Functional benchmark only. No TrueNAS text, branding, assets, or code is used.
Classifications are strict:

| Label | Meaning |
|-------|---------|
| **real** | Usable in product UI/runtime or as a real downloadable artifact; verified in repo |
| **partial** | User can exercise a meaningful subset; major TrueNAS-class behavior still missing |
| **stub/demo/test-only** | Endpoint/UI/scaffold exists but does not deliver product capability |
| **blocked** | Path exists but cannot be proven/completed on this host (missing harness/deps) |
| **absent** | No product implementation |

---

## Phase A — Capability matrix

| # | Capability area | TrueNAS (functional) | Vyntrio today | Class | Exact evidence | Gap | Recommended next action |
|---|-----------------|----------------------|---------------|-------|----------------|-----|-------------------------|
| 1 | Download page / release clarity | Clear download + checksums + notes | `/download` loads live `GET /api/v1/public/install-media`; shows version, filename, type, size, SHA-256, generated_at, support_status, writer platforms; honest not-staged states | **real** | `PublicDownloadView.tsx`, `PublicInstallMediaSection.tsx`, `PublicInstallMediaWriterSection.tsx`, `installmediapublic/metadata.go`, `router.go` | No production CDN; engineering-only channel | Keep honest; do not polish marketing over product |
| 2 | Install artifact availability | ISO/USB installer image published | Dual-mode GPT hybrid `vyntrio-install-media.img` buildable + stageable + downloadable when `VYNTRIO_RELEASE_STAGING_DIR` set; BIOS-only refused at stage | **partial** | `distro/install-media/build/vyntrio-install-media.img`, `make install-media`, `make release-install-media-stage`, `/release/vyntrio-install-media.img` | No ISO, no persistent target-disk OS install, no CDN; Secure Boot unsigned | Boot-proven live path then persistent installer |
| 3 | Install artifact verification | Published SHA-256 (+ often signatures) | Real SHA-256 in `release-manifest.json` + public API + writer verify; unsigned | **partial** | `release-manifest.json`, `cmd/verify-artifact`, writer `verify-image`/`verify-device`, `/download` verify commands | No Ed25519/GPG/codesign | Add signed manifests after boot path proven |
| 4 | Media creator / writer downloads | Desktop tools to write boot media | Cross-platform CLI `vyntrio-write-media`: Windows `.exe`, macOS/Linux binaries served at `/release/writer/*` with size+SHA-256 | **partial** | `cmd/write-media`, `scripts/package-write-media.sh`, `distro/release/staging/writer/*`, `handlers/release_files.go` `ServeWriter` | CLI only (not GUI); no macOS `.dmg`/notarization; unsigned | Optional GUI wrapper later; not blocking vs on-disk storage |
| 5 | First boot into working runtime | Install → reboot → working system | Live initramfs + `firstboot.sh` wired; structural `firmware_bootable: true`; **dashboard never proven on this host** | **blocked** / **partial** wiring | `WRAPPER.txt` `runtime_boot_tested: false`, `dashboard_on_first_boot: false`; `RUNTIME_BOOT.txt` `dashboard_reachable: false`, reason `no_vm_harness` | No qemu/KVM on build host; RAM live only; no persistent install | **Highest priority:** boot harness proof of dashboard reachability |
| 6 | First-login / owner setup | Setup wizard / first user | Loopback bootstrap + login/logout sessions + CSRF | **real** | `POST /api/v1/identity/bootstrap`, `login`, `logout`; `internal/application/identity/*` | No multi-step TrueNAS-style wizard; bootstrap loopback-gated | Extend after persistent install exists |
| 7 | Web admin UI shell/navigation | Full NAS admin | Appliance shell: Overview, Settings, Storage | **partial** | `ApplianceApp.tsx`, `OverviewShell`, `SettingsShell`, `StorageShell` | Narrow surface vs TrueNAS; many RBAC perms have no UI | Grow only when backend services exist |
| 8 | Disk inventory | Disk list, SMART, replace, etc. | Read-only classified inventory with eligibility reasons | **partial** | `GET /api/v1/storage/disks`, `internal/platform/storageinventory`, Storage UI device cards | No SMART, wipe, replace, import, expand | Keep as foundation; add wipe only behind confirmed jobs |
| 9 | Storage pool creation/management | Create/import/expand/export ZFS pools | **Declared pools** persist disk reservations in `{state_dir}/storage/pools.json`; UI can declare; **`disk_format_applied: false`** — no on-disk pool | **partial** | `POST/GET /api/v1/storage/pools`, `internal/platform/storagepool`, `StorageShell` declare flow | Not TrueNAS pools: no format, no RAID/ZFS/btrfs create, no expand/import | Next core: confirmed destructive pool create on eligible disks |
| 10 | Dataset / volume / zvol equivalent | Dataset tree, zvols, quotas | Dataset **plans** (`path_intent`, status `planned`) under declared pools | **partial** | `POST /api/v1/storage/pools/{id}/datasets` | No filesystem create, no quotas, no zvols | After on-disk pools |
| 11 | Share creation/management | Create/edit shares bound to datasets | Share **plans** only (`protocol` forced to `planned`) | **partial** | `POST/GET /api/v1/storage/shares` | No live share publishing | After datasets on disk + protocol daemon |
| 12 | SMB | smbd shares | None | **absent** | No samba code under `internal/`; `protocol_support: not_available` | Entire protocol stack missing | After on-disk datasets |
| 13 | NFS | nfsd exports | None | **absent** | Same | Missing | Same |
| 14 | iSCSI | Block targets | None | **absent** | No iscsi code | Missing | Later |
| 15 | WebDAV / web share | Optional web share | None | **absent** | — | Missing | Later |
| 16 | Users / groups model | Multi-user, directory sync concepts | Owner bootstrap + RBAC roles; **no user/group management UI or APIs beyond auth** | **partial** | `identity/permission.go`, `policy.go`; no users CRUD handlers | Permissions reserved for `users:*` unused | After core NAS storage |
| 17 | ACL / permissions model | Filesystem ACL editors | Session RBAC only; no filesystem ACL | **absent** (FS ACL) / **partial** (API RBAC) | RBAC middleware only | No NFSv4/SMB ACL product | With shares |
| 18 | Snapshot surface | ZFS snapshots UI | None for datasets | **absent** | `docs/05_STORAGE.md` deferred | Missing | After on-disk pools |
| 19 | Replication / backup surface | Replication tasks; NAS data backup | **Appliance state** backup/restore CLI + overview status card; **not** NAS dataset backup | **partial** (state) / **absent** (NAS data) | `cmd/backup`, `cmd/restore`, `backupstatus` | Do not conflate state backup with TrueNAS replication | Keep separate from NAS backup roadmap |
| 20 | Apps / containers | App catalog + containers | None | **absent** | Empty `cmd/worker`; `docs/13_DOCKER.md` aspirational; RBAC `containers:*` unused | Missing | After shares |
| 21 | Virtual machines | KVM VM manager | None in product | **absent** | `docs/14_VM.md` aspirational; `scripts/run-install-media-vm.sh` is **dev helper** only | Missing | After apps or in parallel later |
| 22 | System settings | Broad system config | Instance display-name write + read settings | **partial** | `GET /api/v1/settings`, `PATCH /settings/instance` | No services, NTP, mail, etc. | Incremental |
| 23 | Network settings | Interfaces, IP, DNS, bonds | Overview **presence** read only (`available`/`unknown`/`unavailable`) | **stub/demo/test-only** for config / **partial** for presence | `netpresence`; no `/api/v1/network*`; `network:write` unused | No configuration UI | After installability |
| 24 | Update/release operations | In-UI updates / trains | Local release staging + verify; `cmd/update-agent` empty `main` | **stub/demo/test-only** (agent) / **partial** (staging) | `cmd/update-agent/main.go` empty; no update API routes | No in-product update | After boot-proven releases |
| 25 | Remote/cloud access | Cloud sync / remote UI | Deliberately deferred | **absent** | `docs/03_ROADMAP.md` Stages 4–5; public copy denies remote cloud | Correctly absent | Do not start until local core real |
| 26 | Licensing/edition controls | Edition features | Docs only; RBAC `licensing:*` unused | **absent** | `docs/15_LICENSE.md` | No product licensing | After local core |
| 27 | Public website / product accuracy | Matches shipped product | Download honesty is strong; EN/DE public + OverviewShell updated to declared-pools honesty in this audit run | **partial** | `/download` real; `public.json` EN/DE; `OverviewShell.tsx` | Still no CDN; product surface narrower than TrueNAS | Keep copy tied to real capabilities only |

---

## Phase B — Repository deep check

### Genuinely usable today
- Public download + staged BIOS image download (when staging env set)
- Writer CLI downloads (Win/macOS/Linux) with checksums
- Owner bootstrap, login, logout, CSRF
- Overview host metrics + readiness + backup status read
- Settings instance rename
- Disk inventory with eligibility
- Declare pool / prepare dataset / prepare share plan (JSON state)
- State backup/restore CLI tooling

### Read-only or plan-only (not TrueNAS equivalent)
- Declared pools (`disk_format_state: pending`, `disk_format_applied: false`)
- Dataset `path_intent` plans
- Share plans with `protocol: planned`
- Network presence status
- Overview `pool_count` / `share_count` now reflect declared store via `SummarizeLayout` (fixed in this audit run; previously inventory-only always 0)

### Tests/docs/scaffolds that overstate maturity if misread
- RBAC permissions for containers/VMs/network/updates/licensing without handlers
- `docs/12_NETWORK.md`, `13_DOCKER.md`, `14_VM.md`, `16_UPDATE.md` roadmap tone
- Empty `update-agent`
- ~~Overview/public copy that said “pools not available” after declare APIs shipped~~ — copy corrected in this audit run; `mutation_available: true` without disk format remains easy to over-read

### Blocked on this host
- Runtime boot proof / dashboard-on-boot (`no_vm_harness`: no qemu, no `/dev/kvm`)
- macOS `.dmg`/notarized app (no macOS build host / Apple credentials)

---

## Phase C — Parity scoring

### Capability count (27 rows)
| Class | Count |
|-------|-------|
| real | 3 (download clarity, auth core, writer downloads as artifacts — writer scored partial overall because no GUI/.dmg) |
| Real-ish rows treated strictly | Download clarity **real**; auth **real**; writer row **partial**; state backup under #19 **partial** |

Strict recount of the 27 rows above:
- **real:** 2 (#1 download clarity, #6 first-login/auth)
- **partial:** 12
- **blocked/partial wiring:** 1 (#5)
- **stub:** 2 (#23 config side, #24 agent)
- **absent:** 10

### Weighted estimate (NAS core weighted higher)

Weights (importance 1–5): installability/boot 5, pools/datasets/shares/protocols 5, ACL/users 3, apps/VMs 3, ops/network/update 2, website 1, remote/licensing 1.

**Estimated TrueNAS-class product parity: ~12–18%.**

Rationale: install path and admin shell exist, but the defining NAS loop (format pool → dataset → SMB/NFS share → ACL) is not operational. Declared JSON pools must not be scored as pool management parity.

### Top 10 missing capabilities
1. Boot-proven dashboard on install media (qemu/KVM harness)
2. Persistent target-disk OS install
3. On-disk pool create (format eligible disks)
4. Real datasets/filesystems
5. SMB service
6. NFS service
7. Filesystem ACL / multi-user NAS identity
8. UEFI install image
9. Apps/containers
10. Virtual machines

### Top 5 misleading maturity signals
1. `mutation_available: true` on pools/shares without disk format or protocol daemons (easy to over-read as live NAS mutation)
2. RBAC permissions for containers/VMs/updates/licensing with no product handlers
3. Roadmap/docs describing Docker/VM/network as if near-term product without code
4. Declared pools / share plans that look like storage management but never format disks or publish protocols
5. Empty `cmd/update-agent` + release staging that can be mistaken for in-product updates

### Top 5 strongest real capabilities
1. Honest public install-media API + `/download` metadata (checksums, support status)
2. Firmware-bootable BIOS raw image build/stage/download path
3. Cross-platform writer CLI packages with live download links
4. Session auth (bootstrap/login/logout) + CSRF + RBAC middleware
5. Real block-device inventory with fail-closed eligibility classification

---

## Phase D — Recommended next implementation block

**#1 priority: boot-proven runtime / installability** — prove `dashboard_reachable_on_boot: true` via qemu/KVM harness (or document hard host blocker and obtain harness).

### Why this beats other candidates now
- Media creator downloads (**priority 2**) are already staged and downloadable.
- Declared pool mutation (**priority 3**) exists as JSON foundation; expanding it without boot proof still leaves users unable to trust the appliance after USB boot.
- TrueNAS-class products assume: download → write media → boot → working admin. Vyntrio fails the third step on evidence (`dashboard_reachable: false`).
- Storage protocol work on an unproven live image creates more surface without closing the installability gap.

If qemu/KVM remains impossible on this host, the block is **blocked** and the next executable block is **on-disk pool create** (destructive, confirmed) — still after documenting the boot-proof gap as open.

---

## Evidence anchors (non-exhaustive)

| Area | Paths |
|------|-------|
| Public API | `internal/interfaces/http/router.go`, `handlers/public_install_media.go`, `handlers/release_files.go` |
| Install media | `WRAPPER.txt`, `RUNTIME_BOOT.txt`, `scripts/stage-release-install-media.sh` |
| Writer | `cmd/write-media`, `distro/release/staging/writer/` |
| Storage | `internal/platform/storageinventory`, `storagepool`, `application/storage/*`, `StorageShell.tsx` |
| Auth | `application/identity/*`, `domain/identity/permission.go` |
| Update stub | `cmd/update-agent/main.go` |
| Prior short audit | `docs/ops/truenas-parity-audit-2026-07-17.md` |

---

## Phase E — Factual consistency fixes (this audit run)

| Fix | Path |
|-----|------|
| Overview honesty copy | `frontend/src/features/overview/OverviewShell.tsx` |
| EN/DE public capability strings | `frontend/src/shared/i18n/locales/{en,de}/public.json` |
| Overview counts include declared pools/shares | `internal/application/overview/loader.go`, `cmd/api/main.go`, tests |
| Storage docs overview row | `docs/05_STORAGE.md` |
| Short audit → intensive pointer | `docs/ops/truenas-parity-audit-2026-07-17.md` |

No large implementation started. No commit. No push.
