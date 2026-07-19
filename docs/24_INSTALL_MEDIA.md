# Download and install media

Technical guide for obtaining and booting the current Vyntrio OS install image.
This is **early-access engineering media**, not a production release channel.

## Product baseline (hard requirement)

| Requirement | Status |
|-------------|--------|
| UEFI boot support | **Required** — modern hardware must boot via UEFI |
| BIOS/legacy fallback | Supported on dual-mode media |
| BIOS-only media | **Incomplete** — must not be staged or called complete |

## Primary artifact

| Field | Value |
|-------|-------|
| Filename | `vyntrio-install-media.img` |
| Format | Raw GPT hybrid disk (`raw_gpt_hybrid_disk`) |
| Firmware | Dual-mode: **BIOS + UEFI** (`bios+uefi`) |
| UEFI payload | ESP with `EFI/BOOT/BOOTX64.EFI` |
| Build output | `distro/install-media/build/vyntrio-install-media.img` |

Legacy BIOS-only filename `vyntrio-install-media-bios.img` is a fallback build path only and is **not** the product baseline. Release staging refuses BIOS-only images.

## 1. Build the image

From a checkout of this repository with build dependencies installed:

```bash
make install-media
```

This runs the full install-media chain through live-initramfs swap.
Confirm `distro/install-media/build/WRAPPER.txt` contains:

- `firmware_bootable: true`
- `firmware_boot_mode: bios+uefi`
- `uefi_support: true`
- `dual_mode: true`
- `product_baseline_complete: true`

Optional structural/UEFI smoke:

```bash
./scripts/verify-uefi-boot.sh
```

## 2. Stage for HTTP download (optional)

To expose the image from a running `vyntrio-api` host:

```bash
make release-install-media-stage
export VYNTRIO_RELEASE_STAGING_DIR=/absolute/path/to/distro/release/staging
# restart vyntrio-api with the env var set
```

Staged files:

| File | Purpose |
|------|---------|
| `vyntrio-install-media.img` | Dual-mode bootable raw disk image |
| `release-manifest.json` | SHA-256 manifest (`vyntrio-release-manifest-v1`) |
| `install-media-public.json` | Public metadata for `/api/v1/public/install-media` |

Verify before use:

```bash
vyntrio-verify-artifact --base-dir distro/release/staging distro/release/staging/release-manifest.json
```

Download URLs (same origin as the API):

- Image: `/release/vyntrio-install-media.img`
- Manifest: `/release/release-manifest.json`

There is **no production CDN**. If staging is not configured, the download page shows build instructions only.

## 3. Write the image

Use the guided helper scripts (recommended) or write manually. The primary image is dual-mode; there is no separate UEFI-only download.

## Cross-platform writer

Windows, macOS, and Linux users should use **`vyntrio-media-creator`** (local web GUI) or the CLI alias **`vyntrio-write-media`**:

```bash
make build-write-media    # host binary in bin/
make package-write-media  # GUI packages + CLI aliases under distro/release/writer/
```

See `docs/ops/install-media-writer.md`. Honest limits: no notarized macOS `.dmg`, no Linux AppImage, unsigned.

The Linux bash helper `./scripts/write-install-media-usb.sh` remains supported.

### Guided helpers (Linux bash, optional)

From the repository root after staging:

```bash
# Validate image + manifest checksum (no write)
./scripts/write-install-media-usb.sh --dry-run

# USB write — destructive; requires root and explicit device confirmation
sudo ./scripts/write-install-media-usb.sh --device /dev/sdX

# VM — non-destructive; UEFI (default) or BIOS
./scripts/run-install-media-vm.sh --dry-run
./scripts/run-install-media-vm.sh --firmware uefi
./scripts/run-install-media-vm.sh --firmware bios
```

The USB helper refuses mounted devices, verifies SHA-256 against `release-manifest.json` when present, and requires typing the device path to confirm (unless `--yes`).

The `/download` page includes the same steps with live checksum and download links when staging is configured. It states **BIOS support**, **UEFI support**, and **dual-mode** explicitly.

### USB stick (bare metal, manual)

Identify the target block device carefully. Writing to the wrong device destroys data.

```bash
sudo dd if=distro/release/staging/vyntrio-install-media.img of=/dev/sdX bs=4M conv=fsync status=progress
```

Replace `/dev/sdX` with the USB device. Prefer the helper script for checksum verification and confirmation gates.

### Virtual machine

Attach `vyntrio-install-media.img` as a disk. Prefer **UEFI (OVMF)**; BIOS/legacy also works on dual-mode media. Or use `./scripts/run-install-media-vm.sh`.

## 4. Boot and first dashboard

1. Power on the target with UEFI or BIOS boot from USB or VM disk.
2. GRUB loads the Vyntrio live initramfs (busybox userland).
3. `firstboot.sh` supervises `vyntrio-api` on loopback HTTP by default.
4. Open the dashboard in a browser on the appliance (`http://127.0.0.1:8080/app` on the guest).
5. Create the owner account via bootstrap (loopback API) or sign in if already created.

Optional LAN HTTPS: set `VYNTRIO_LAN_BIND_IP` before boot so `prepare-live-dashboard-tls.sh` generates ephemeral TLS. See `docs/ops/install-media-dashboard-tls.md`.

## Known limitations

- **Secure Boot unsigned** — engineering image is not Secure Boot signed.
- **No target-disk installer** — the image boots a live RAM runtime; it does not install to a persistent system disk yet.
- **Boot not fully proven on all hosts** — structural UEFI packaging is verified; qemu/OVMF smoke is best-effort when tools and KVM are present.
- **Unsigned releases** — SHA-256 manifest integrity only; Ed25519 verification is not implemented.
- **No in-browser USB writer** — use helper scripts or manual `dd`; see `/download` guided flow.

## Public website

The `/download` page loads live status from `GET /api/v1/public/install-media` and includes a guided **Create bootable install media** section (USB + VM paths, checksum, helper commands), including BIOS / UEFI / dual-mode rows.

## Related docs

- `docs/ops/install-media-audit.md` — artifact inventory
- `docs/ops/install-media-wrapper.md` — image construction
- `docs/ops/release-artifact-verification.md` — manifest verification
- `docs/19_RELEASE.md` — release build path
