# Vyntrio — Produktstruktur-Audit (2026-07-19)

Stand aus Code/Docs/Artefakten auf dem Build-Host. **Repo-Wahrheit, keine Roadmap.**

## Kernurteil

Vyntrio ist heute eine **bootbare Live-/USB-Appliance** mit echtem Control Center
(API + WebGUI) und deklarierten Storage-/Share-Plänen — **nicht** ein fertiges
NAS-OS mit Docker/VMs/SMB.

„Image auf einem VPS installieren“ bedeutet im Repo: Image unter **QEMU/OVMF**
auf dem Server booten — **nicht** die VPS-Systemdisk überschreiben.

| Kennzahl | Wert |
|----------|------|
| Install-Image | ~806 MiB dual-mode `.img` (BIOS+UEFI) |
| Boot→Dashboard | QEMU UEFI bewiesen (`APPLIANCE_BOOT_PROOF.txt`, 2026-07-17) |
| `/app`-Nav | 6 Einträge (Users/Tools Planned) |
| Docker/VM-Runtimes | 0 (nicht ausgeliefert) |

---

## 1. Produkt-Schichten

| Schicht | Inhalt | Status |
|---------|--------|--------|
| Öffentliche Site | `/`, `/download`, `/docs`, `/login` | Available |
| Appliance `/app` | Dashboard · Storage · Shares · Settings (+ Users/Tools Planned) | Partial |
| Design-Preview | `/design-preview/appliance-full` (Unraid-Klassen-Mock inkl. Fake-Docker) | Demo-only |

---

## 2. Oberflächen-Matrix

| Fläche | Was sie tut | Status |
|--------|-------------|--------|
| Public `/` | Landing DE/EN, honesty blocks | Available |
| Public `/download` | Metadata + staged image when env set | Available |
| Public `/docs` | Static docs topics | Available |
| Public `/login` → `/app` | Session cookie + CSRF | Available |
| `/app` Dashboard | Live overview from API | Available |
| `/app/storage` | Disk inventory + declare pools/datasets | Partial |
| `/app/shares` | Share plans only — no SMB/NFS | Partial |
| `/app/settings` | Owner: instance display name | Available |
| `/app/users` | Nav Planned shell | Planned |
| `/app/tools` | Nav Planned shell | Planned |
| Docker / Apps / VMs in `/app` | Kept out of product nav | Not available |
| `/design-preview/appliance-full` | Full Unraid-class mock UI | Demo-only |

---

## 3. Backend & Binaries

| Komponente | Fähigkeit | Status |
|------------|-----------|--------|
| `vyntrio-api` | Control plane + embedded SPA | Available |
| Auth (login/logout/bootstrap) | Opaque session, RBAC, CSRF | Available |
| `GET /api/v1/overview` | Host/storage/backup summary | Available |
| Storage disks / pools / datasets | Read + declare plans; no format | Partial |
| Storage shares | `protocol` forced planned | Partial |
| Public install-media API | Metadata; download if staged | Available |
| `vyntrio-backup` / `restore` | Root-only local state CLIs | Available |
| `vyntrio-write-media` | USB writer CLI | Available |
| `vyntrio-installer` | Sandbox / existing partition only | Partial |
| `vyntrio-worker` | Empty main stub | Not available |
| `vyntrio-update-agent` | Empty main stub | Not available |
| Docker / VM / marketplace APIs | No product runtime | Not available |
| Remote Connect / licensing | Not in surface | Not available |

---

## 4. Install-Media & Boot

| Artefakt / Fähigkeit | Detail | Status |
|----------------------|--------|--------|
| USB appliance `.img` (~806 MiB) | GPT hybrid BIOS+UEFI raw disk | Available |
| Release staging + SHA-256 | `local_staging`, no CDN | Available |
| QEMU UEFI → dashboard 200 | Proven on this host (TCG) | Available |
| Persistent install to system disk | Live/USB runtime only | Not available |
| Secure Boot signed | unsigned engineering media | Not available |
| Production CDN channel | Explicitly not shipped | Not available |
| Media Creator (Tauri) | Linux/Windows staged; no macOS dmg | Partial |
| VPS / cloud OS install path | No docs or scripts | Not available |

**Boot-Beweis:** `APPLIANCE_BOOT_PROOF.txt` — `firmware=uefi`, `accel=tcg`,
`dashboard_http=200`, `readyz=200`, `dashboard_reachable_on_boot=true`.

---

## 5. Kurzfassung

### Ja / teilweise

- USB/VM-Image bauen, stagen, SHA-256 prüfen
- Live-Boot → firstboot → `vyntrio-api` :8080
- Owner-Login, Overview mit echten Host-Daten
- Disk-Inventar lesen; Pools/Datasets/Shares deklarieren
- Instance-Name in Settings
- Lokales State-Backup/Restore (CLI)
- Öffentliche Site + Download-Metadaten

### Nein (ehrlich)

- On-Disk-Format / echte SMB/NFS/iSCSI
- Docker / Apps-Katalog / Compose
- Hypervisor / VM-Manager
- Persistente Installation auf Systemdisk
- Remote Connect, Lizenz, Updates-Agent
- Secure Boot, CDN-Releasekanal
- „Vyntrio als VPS-OS installieren“

---

## 6. Image testweise auf einem VPS

### Falsche Erwartung vermeiden

Es gibt **keinen** dokumentierten Weg, das `.img` als Cloud-ISO hochzuladen und
den VPS damit dauerhaft zu „installieren“. Das Produkt bootet eine Live-Appliance
(RAM/USB-Modell). Auf einem VPS testest du per **QEMU+OVMF** auf dem gemieteten
Linux-Host.

### Empfohlener Testpfad (Repo-konform)

| Schritt | Aktion |
|---------|--------|
| 1. VPS wählen | x86_64 Linux (Debian/Ubuntu), ≥4 GB RAM, ≥2 GB frei |
| 2. Tools | `qemu-system-x86_64` + OVMF (UEFI); KVM optional (ohne `/dev/kvm` → TCG) |
| 3. Artefakt | `make install-media` oder fertiges `.img` kopieren |
| 4. Boot-Smoke | `./scripts/run-install-media-vm.sh --firmware uefi` |
| 5. Dashboard-Beweis | `./scripts/prove-appliance-boot.sh` |
| 6. Browser | Hostfwd: `http://127.0.0.1:<port>/app` |

```bash
sudo apt-get update
sudo apt-get install -y qemu-system-x86 ovmf git make
cd /opt/vyntrio-os   # oder Clone-Pfad
make install-media   # oder Image scp'en
./scripts/run-install-media-vm.sh --firmware uefi
./scripts/prove-appliance-boot.sh
```

### Was auf dem VPS nicht funktioniert

- **Cloud „Custom ISO“:** Raw GPT-USB-Image ≠ Cloud-ISO.
- **`dd` auf VPS-Systemdisk:** überschreibt Provider-Root; liefert trotzdem nur Live-Runtime.

### Alternative: USB / Bare Metal

```bash
./scripts/write-install-media-usb.sh --dry-run
sudo ./scripts/write-install-media-usb.sh --device /dev/sdX
```

---

## 7. Quellen

- `docs/24_INSTALL_MEDIA.md`
- `docs/05_STORAGE.md`
- `docs/09_API.md`
- `frontend/src/surfaces/appliance/nav/applianceNavConfig.ts`
- `internal/interfaces/http/router.go`
- `scripts/run-install-media-vm.sh`
- `scripts/prove-appliance-boot.sh`
- `distro/install-media/build/APPLIANCE_BOOT_PROOF.txt` (lokal, gitignored build dir)

Ältere Ops-Audits (vor Boot-Proof) können `dashboard_reachable=false` noch
behaupten — die Proof-Datei hat Vorrang.
