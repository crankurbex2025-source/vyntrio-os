# Native Tauri Media Creator audit (2026-07-17)

## Framework

**Tauri 2** + React/TypeScript UI under `desktop/vyntrio-media-creator/`.

## Packages produced on this host

| File | Path |
|------|------|
| Windows NSIS | `distro/release/staging/writer/vyntrio-media-creator-windows-amd64-setup.exe` |
| Linux .deb | `distro/release/staging/writer/vyntrio-media-creator-linux-amd64.deb` |
| Linux AppImage | `distro/release/staging/writer/vyntrio-media-creator-linux-amd64.AppImage` |

## Withdrawn (not native desktop)

- Go loopback `vyntrio-media-creator-windows-amd64.exe`
- Linux loopback `.tar.gz`
- macOS Terminal helpers presented as GUI
- macOS `.app.zip`

## Blocked

| Item | Exact blocker |
|------|----------------|
| macOS `.app` / `.dmg` | No macOS build host; `hdiutil` unavailable; notarization impossible on Linux |
| Code signing / SmartScreen / Gatekeeper trust | No signing keys / Apple notarization |
| MSI | NSIS produced instead; WiX/`wixl` not used |
| AppImage via `tauri build --bundles appimage` alone | `linuxdeploy` failed; recovered with `appimagetool` + `file` |

## UEFI

Install image baseline remains dual-mode `bios+uefi`. Creator UI refuses incomplete BIOS-only releases (`uefi_support` required to continue).
