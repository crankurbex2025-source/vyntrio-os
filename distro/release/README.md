# Release artifact manifest (Slice 9.10 + 13.1)

Example JSON manifest for `vyntrio-verify-artifact`. **Slice 13.1** adds a
real staging path for the dual-mode (BIOS + UEFI) install image.

## Files

| File | Purpose |
|------|---------|
| `release-manifest.example.json` | Illustrative schema only — placeholders |
| `staging/` (generated) | `make release-install-media-stage` output — not committed |

## Staging install media (Slice 13.1)

```bash
make install-media
make release-install-media-stage
make test-release-install-media-stage
```

Staging **refuses** BIOS-only images (`product_baseline_complete` / `uefi_support` / `dual_mode` required).

Outputs under `distro/release/staging/`:

- `vyntrio-install-media.img`
- `release-manifest.json`
- `install-media-public.json`
- `STAGING.txt`

Serve from `vyntrio-api` by setting `VYNTRIO_RELEASE_STAGING_DIR` to the
absolute staging directory path.

## Usage

After building or staging artifacts locally:

```bash
make build
vyntrio-verify-artifact --base-dir distro/release/staging distro/release/staging/release-manifest.json
```

## Guarantees (v1)

- SHA-256 integrity verification against manifest entries
- Fail-closed manifest parsing (`DisallowUnknownFields`)
- Public metadata endpoint without serving binaries through JSON
- Dual-mode UEFI+BIOS packaging required for staged primary artifact

## Not implemented

- Ed25519 signature verification (metadata field reserved)
- Production CDN / automated CI publication
- Secure Boot signing

See `docs/19_RELEASE.md`, `docs/24_INSTALL_MEDIA.md`, and `docs/ops/release-artifact-verification.md`.
