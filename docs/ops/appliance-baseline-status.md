# Appliance master plan — baseline status (Block 0)

Recorded: 2026-07-16. Source: one-shot verification on this build host.

## Status matrix

| Pillar | State | Verified how | Honest blocker |
|--------|-------|--------------|----------------|
| Boot/media chain (Stage 1) | **Complete** | `make test-install-media-initrd-swap` | None on chain |
| Initrd swap to live initramfs | **Complete** | `INITRD_SWAP.txt` `boots_live_initramfs=true` | Runtime boot not exercised |
| Runtime boot verification (Stage 2) | **Blocked here** | `make test-install-media-runtime-verify` | No qemu/`/dev/kvm` |
| Local dashboard stability | **Implemented** | busybox-sh `firstboot.sh`, supervised respawn, chroot `/readyz` | No booted-guest proof |
| First-boot onboarding clarity | **Implemented** | Banner + `FIRST_BOOT.txt` + landing `#first-boot-setup` | Owner flow not proven on boot |
| TLS / LAN dashboard bind | **Ready (config + runtime prep)** | API TLS, HTTPS chroot probe, `tls_readiness: ready` | Operator must set `VYNTRIO_LAN_BIND_IP`; no LAN reachability faked |
| Storage API (inventory) | **Complete** | `GET /api/v1/storage/disks`, Go tests | — |
| Storage/NAS UI foundation | **Implemented (this run)** | Frontend `StorageShell`, 126 frontend tests | Not proven on booted live runtime |
| Remote Connect | **Not started** | Roadmap Stage 4 gated | Stage 3 + boot proof |
| Licensing | **Not started** | Roadmap Stage 5 gated | Remote Connect |
| Public landing alignment | **Updated (this run)** | i18n + deploy verification pending | Install media not published |

## Commands used (Block 0)

```bash
make test-install-media-initrd-swap
make test-install-media-runtime-verify
make test-install-media-runtime
make test-install-media-live-rootfs
go test ./internal/platform/config/... ./internal/platform/tlsutil/... ./internal/interfaces/http/...
cd frontend && npm run test:run
```

## Unrelated CI / lint notes (not blocking this plan)

- Pre-existing frontend type issues may exist in files outside this plan's scope (`App.test.tsx`, `PublicDownloadView.tsx`, `vite.config.ts`) — not expanded in this run.

## Gate decision

Blocks **1–7** unblocked for product-local work. Blocks **8–9** remain gated.
