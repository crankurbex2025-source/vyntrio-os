# Appliance product gap — final audit (2026-07-16)

## 1. Executive summary

Vyntrio remains a **local-first appliance prototype**, not an Unraid/TrueNAS-class
NAS OS. This run performed a strict 15-category gap audit and implemented **Slice
12.2 — storage layout foundation**: read-only pools/shares API endpoints, overview
storage summary, and UI sections that report honest empty layout state with
`mutation_available: false`.

**Real after this run:** block inventory (API + UI), storage layout read models
(pools/shares endpoints + overview counts + Storage/Overview panels).

**Still absent:** pool/share mutation, NAS protocols, containers, VMs, boot-proven
runtime, public install download, remote access, licensing.

No commit, push, or deploy in this run.

---

## 2. Capability matrix summary

| State | Count | Categories |
|-------|-------|------------|
| implemented | 2 | Storage inventory (#6), Website (#15) |
| partial | 6 | Boot media (#1), First boot (#2), WebGUI (#3), Onboarding (#4), LAN TLS (#5), System ops (#12) |
| stub-demo-test-only | 2 | Pools (#7), Shares (#8) |
| blocked | — | Boot verification (`no_vm_harness` on build host) |
| not implemented | 5 | Protocols (#9), Containers (#10), VMs (#11), Remote (#13), Licensing (#14) |

Full matrix: `docs/ops/appliance-product-gap-audit-2026-07-16.md`.

---

## 3. What was implemented in this run

**Slice 12.2 — Storage layout foundation (read-only, honest empty state)**

| Layer | Change |
|-------|--------|
| Application | `summary.go`, `pools.go`, `shares.go`, `InventorySource` interface |
| API | `GET /api/v1/storage/pools`, `GET /api/v1/storage/shares`; overview `storage` field |
| HTTP | `handlers/storage.go` — `ServePools`, `ServeShares`; router `/storage/*` route group |
| UI | Overview storage layout panel; Storage page layout status (inventory + pools + shares) |
| Tests | Go unit/handler tests; frontend DTO/shell/App tests updated |

---

## 4. Why this block was chosen

Per priority rules: real product capability > core appliance management > honest
parity progress > minimal scope. Pool **mutation** is not safe to fake; public
download and boot proof are blocked or out of scope. Extending storage with
layout read models gives users a truthful NAS-style navigation structure while
explicitly marking management as unavailable.

---

## 5. Exact changed files

**Go**

- `internal/application/storage/response.go`
- `internal/application/storage/summary.go`
- `internal/application/storage/summary_test.go`
- `internal/application/storage/pools.go`
- `internal/application/storage/pools_test.go`
- `internal/application/storage/shares.go`
- `internal/application/storage/shares_test.go`
- `internal/application/storage/loader.go`
- `internal/application/overview/response.go`
- `internal/application/overview/loader.go`
- `internal/application/overview/loader_test.go`
- `internal/interfaces/http/handlers/storage.go`
- `internal/interfaces/http/handlers/storage_test.go`
- `internal/interfaces/http/handlers/overview_test.go`
- `internal/interfaces/http/handlers/settings_test.go`
- `internal/interfaces/http/router.go`
- `cmd/api/main.go`

**Frontend**

- `frontend/src/features/storage/storageDto.ts`
- `frontend/src/features/storage/StorageShell.tsx`
- `frontend/src/features/storage/StorageShell.test.tsx`
- `frontend/src/features/overview/overviewDto.ts`
- `frontend/src/features/overview/overviewDto.test.ts`
- `frontend/src/features/overview/OverviewShell.tsx`
- `frontend/src/features/overview/OverviewShell.test.tsx`
- `frontend/src/surfaces/appliance/ApplianceApp.tsx`
- `frontend/src/App.test.tsx`

**Docs**

- `docs/05_STORAGE.md`
- `docs/09_API.md`
- `docs/20_TASKS.md`
- `docs/ops/appliance-product-gap-audit-2026-07-16.md` (new)
- `docs/ops/appliance-product-gap-final-2026-07-16.md` (this file)

---

## 6. Verification commands and outcomes

```bash
# Go — storage, overview, handlers
go test ./internal/application/storage/... ./internal/application/overview/... ./internal/interfaces/http/handlers/...
# ok (handlers ~3.5s)

# Frontend
cd frontend && npm run test:run
# 23 files, 128 tests passed
```

Install media artifact (unchanged this run):

```bash
wc -c distro/install-media/build/vyntrio-install-media-bios.img
# 136314880 bytes
grep firmware_bootable distro/install-media/build/WRAPPER.txt
# firmware_bootable: true
```

---

## 7. Live deploy status

**Not deployed.** Product UI changes require `make ui-stage && make build` and
service restart to embed in `vyntrio-api`; no deploy was requested or performed.
Public website unchanged.

---

## 8. Remaining top 10 gaps

1. Boot-proven firmware → dashboard path (`dashboard_reachable_on_boot: false`)
2. Public install image publication
3. Storage pool creation/mutation from eligible disks
4. Share creation and ACL model
5. SMB/NFS (or other) protocol services
6. First-boot owner setup wizard in WebGUI
7. App/container catalog and lifecycle
8. VM management
9. Operational settings (network interfaces, power, updates)
10. Remote Connect (Stage 4 gate)

---

## 9. Recommended next block

**Slice 12.3 — Pool creation design + safe non-destructive dry-run**, or if VM
harness becomes available, **runtime boot verification** to unblock boot-proven
claims. Do not advance Remote Connect or licensing before core local storage
mutation exists.

---

## 10. Commit readiness verdict

**Ready for review, not committed.** All targeted tests pass; docs match
implementation. User instructed no commit/push.
