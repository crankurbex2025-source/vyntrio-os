# Install-media writer / Media Creator

## Product form (required)

**Vyntrio Media Creator** is a **native desktop app** built with **Tauri**
(`desktop/vyntrio-media-creator/`).

It is **not**:
- a browser-only wizard presented as desktop parity
- a loopback localhost Go GUI presented as native
- a CLI-first flow claimed as GUI parity

## Packages (this host)

| Platform | Package | Kind |
|----------|---------|------|
| Windows | `vyntrio-media-creator-windows-amd64-setup.exe` | NSIS installer (unsigned; cross-built) |
| Linux | `vyntrio-media-creator-linux-amd64.deb` | Debian package |
| Linux | `vyntrio-media-creator-linux-amd64.AppImage` | AppImage |
| macOS | — | **Blocked**: no macOS build host / no `hdiutil` / no notarization |

CLI aliases remain available via `make package-write-media`.

## Build / stage

```bash
# Native Tauri (after Rust + webkit deps installed)
cd desktop/vyntrio-media-creator && npm install
npm run tauri -- build --bundles deb
npm run tauri -- build --target x86_64-pc-windows-gnu --bundles nsis
# AppImage may need appimagetool + file(1) if linuxdeploy fails
make package-media-creator-native

# Optional CLI
make package-write-media
```

## App flow

1. Welcome  
2. Release selection (version, size, SHA-256, support status, BIOS/UEFI/dual-mode)  
3. USB/storage selection + local image path  
4. Confirm destructive write  
5. Progress  
6. Completion (verification + boot instructions)  
7. Help/about + EN/DE language switch  

## UEFI

The creator surfaces `bios_support` / `uefi_support` / `dual_mode` from install-media
metadata. BIOS-only images are treated as incomplete and cannot proceed past release
selection when `uefi_support` is false.

## Honesty

- Packages are unsigned.
- Windows SmartScreen may warn.
- macOS `.app`/`.dmg` not produced here.
- Writing USB requires elevation.
