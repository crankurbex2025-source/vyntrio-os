# Install-media writer audit addendum (2026-07-16)

Supplements `docs/ops/appliance-install-media-final-2026-07-16.md` after Slice 13.3.

## Implementation model chosen

**Option 2 — cross-platform packaged CLI** (`vyntrio-write-media`).

| Alternative | Decision |
|-------------|----------|
| Native desktop app (Electron/Tauri) | **Rejected** — no packaging stack in repository |
| In-browser writer | **Rejected** — unsafe / infeasible for block-device writes |
| Web guide + bash only | **Insufficient** for Windows/macOS users |

The CLI lists removable devices, verifies SHA-256 against `release-manifest.json`, writes the BIOS raw image, and optionally verifies the device after write. It is **not** a GUI application.

## Supported platforms

| Platform | Device listing | Write path | Package name |
|----------|----------------|------------|--------------|
| Linux | `/sys/block` removable/USB | `/dev/sdX` | `vyntrio-write-media-linux-amd64` |
| macOS | `diskutil list` external | `/dev/rdiskN` | `vyntrio-write-media-darwin-arm64`, `-darwin-amd64` |
| Windows | PowerShell `Get-Disk` (USB/SD) | `\\.\PhysicalDriveN` | `vyntrio-write-media-windows-amd64.exe` |

Build: `make build-write-media` · Package: `make package-write-media` · Test: `make test-write-media`

## Primary artifact (unchanged)

- **File:** `vyntrio-install-media-bios.img`
- **Path (build):** `distro/install-media/build/`
- **Path (staged):** `distro/release/staging/`
- **Type:** raw MBR BIOS disk · **Firmware:** legacy BIOS only

## User-visible install path status

| Surface | Status |
|---------|--------|
| `/download` install status panel | Live API metadata (version, size, SHA-256, generated_at, support_status) |
| `/download` writer section | CLI workflow + per-OS checksum commands |
| `/download` creator guide | USB bash + VM paths (Linux-oriented) |
| Writer binaries on CDN | **Not available** — build/package from source |
| Writer downloads (host-local, 2026-07-17) | **Available when staged** — `make package-write-media` stages binaries to `distro/release/staging/writer/`; served at `/release/writer/<name>` with size + SHA-256 in `GET /api/v1/public/install-media` (`writer.artifacts[]`) |

## Remaining gaps toward polished installer

1. Native GUI wrapper for non-technical users
2. Production CDN hosting for image + writer binaries
3. Code-signed writer packages (Windows/macOS)
4. UEFI install image variant
5. Boot-proven dashboard on build host
6. Target-disk persistent install
7. Ed25519 signed releases

## Honest limitations

- Destructive USB writes — user must confirm device
- Engineering / early-access media only
- Writer is CLI — do not market as a desktop app
- No auto-download of image inside writer binary
