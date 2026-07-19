# TrueNAS-class parity audit (2026-07-17)

**Canonical product-class audit:** [`truenas-unraid-product-class-audit-2026-07-17.md`](./truenas-unraid-product-class-audit-2026-07-17.md)

Also see: [`truenas-intensive-parity-audit-2026-07-17.md`](./truenas-intensive-parity-audit-2026-07-17.md)

## Headline

Weighted TrueNAS/Unraid-class parity: **~14–20%**.

This run shipped **GUI media creator packages** (local web wizard): Windows `.exe`, macOS `.app.zip`, Linux `.tar.gz` — not Electron/Qt, not notarized `.dmg`, not AppImage.

## Recommended next block

**Boot-proven runtime** (`dashboard_reachable_on_boot: true`) or, if harness remains impossible, **on-disk pool create**.
