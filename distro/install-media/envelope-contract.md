# Install-image live/boot envelope contract

- **Status:** Envelope scaffold only — **no envelope builder is implemented**
- **Block:** 9
- **Slice:** 9.7 (documentation)
- **Authority:** `docs/ADR/0006-appliance-install-recovery-media.md` (install media
  sections B, D). Layer inventory: `distro/install-media/envelope-manifest.yaml`.
  Payload inventory: `distro/install-media/manifest.yaml`.
  Staging: `distro/install-media/build-contract.md` (Slice 9.6).

## 1. Explicit status

**Slice 9.7 (this document):** defines the **live/boot envelope scaffold** — how
a future bootable install image wraps the staged `payload/` tree. This is a
**declarative contract only**. No ISO/USB generator, boot loader, live-root
builder, or installer runs in this slice.

**Slice 9.8 (implemented):** `make install-media-envelope` runs
`scripts/assemble-install-media-envelope.sh`, which consumes staged payloads and
assembles the local disposable tree `distro/install-media/envelope/` with
`boot/`, `live_root/`, and `payload/` layers. `boot/` and `live_root/` remain
deferred placeholders only. This does **not** produce ISO/USB images or write to
target disk paths.

**Slice 9.9 (scaffold):** `distro/install-media/bootability-contract.md` and
`bootability-manifest.yaml` define the bootable initialization boundary beyond
envelope assembly. This is declarative only.

Recovery-image envelope is a **separate future contract** under
`distro/recovery-media/` — not defined here.

## 2. Envelope purpose

A future **install-image envelope** is the bootable outer shell (ISO or raw USB
image) that carries three conceptual layers:

| Layer | On-media path | Role |
|-------|---------------|------|
| **boot** | `boot/` | Firmware handoff — bootloader, kernel, initrd (deferred) |
| **live_root** | `live_root/` | Minimal runtime that hosts future install execution (deferred) |
| **payload** | `payload/` | Manifest-listed artifacts for greenfield copy to **target disk** |

The envelope **does not** install to disk, partition, enable services, or seed
bootstrap state. It only **packages** artifacts for boot-time access.

## 3. Relationship to staging

```
Repository (build prep)                    Future bootable image
─────────────────────                      ───────────────────────
distro/install-media/staging/payload/  →  distro/install-media/envelope/payload/
        ▲                                          ▲
        │ make install-media-stage (9.6)           │ make install-media-envelope (9.8)
        │                                          │
manifest.yaml (payload authority)          envelope-manifest.yaml (layer authority)
```

| Artifact | Location | Committed | Role |
|----------|----------|-----------|------|
| Staged payloads | `distro/install-media/staging/payload/` | **No** (gitignored) | Disposable local input |
| `STAGING.txt` | `distro/install-media/staging/` | **No** | Staging provenance only |
| Assembled envelope | `distro/install-media/envelope/` | **No** (gitignored) | Local disposable envelope tree |
| `ENVELOPE.txt` | `distro/install-media/envelope/` | **No** | Envelope assembly provenance only |
| Envelope manifest | `distro/install-media/envelope-manifest.yaml` | **Yes** | Declarative layer contract |
| Future image file | build output (deferred) | **No** | Ephemeral distribution artifact |

Staging remains **local and disposable**. Envelope assembly will **read** staged
payloads; it must not write to target disk paths (`/var/lib/vyntrio/`,
deployed `/etc/vyntrio/`, etc.).

## 4. Layer boundaries

### 4.1 boot (deferred)

- Bootloader configuration (UEFI/BIOS)
- Kernel and initrd selection
- No repository scaffold populated in Slice 9.7

### 4.2 live_root (deferred)

- Minimal live filesystem for future `vyntrio-installer` host
- Must not embed appliance SQLite, Owner credentials, or license files
- Must not include recovery-media tooling (`distro/recovery-media/`)

### 4.3 payload (reference — staged in 9.6)

- Authority: `distro/install-media/manifest.yaml`
- Staging source: `distro/install-media/staging/payload/` after `make install-media-stage`
- Contents: binaries, systemd drop-ins, config template only
- Target-disk paths are mirrored as relative paths under `payload/` (not absolute host paths)

## 5. Install vs recovery vs target disk

| Concern | Install envelope (this contract) | Recovery media | Target disk |
|---------|----------------------------------|----------------|-------------|
| Deliverable | Greenfield install image | Offline recovery image | Running appliance |
| Manifest | `install-media/manifest.yaml` | `recovery-media/manifest.yaml` | N/A |
| Envelope manifest | `envelope-manifest.yaml` | Future separate contract | N/A |
| Payload model | Copy **to** disk at install | Run tooling **from** live env | Persistent state home |
| Combined image | **Forbidden** (ADR-0006) | **Forbidden** | N/A |

Persistent state (`/var/lib/vyntrio/`, populated config) exists on the **target
disk** only, per ADR-0005. Bootstrap runs on the **installed** service per
ADR-0004 — not from generic install media.

## 6. Deferred executable slices

The following require future implementation slices after 9.8:

- Live root filesystem composition (minimal Debian/base)
- Boot loader and UEFI/BIOS configuration
- ISO/USB image file generation (`xorriso`, `grub`, `dd`, etc.)
- `vyntrio-installer` execution
- `envelope-record.yaml` build metadata emission (beyond local `ENVELOPE.txt`)

## 7. Out of scope (Slice 9.7 and 9.8)

- ISO/USB generation implementation
- Boot loader or live-root execution
- Partitioning, filesystem creation, disk encryption
- Installer execution or install-progress reporting
- Recovery-media envelope changes
- API/dashboard install status
- Host paths, services, secrets, or backup artifacts at build time

## 8. Related documents

- `distro/install-media/envelope-manifest.yaml` — declarative layer inventory
- `distro/install-media/bootability-contract.md` — bootable initialization boundary
- `distro/install-media/bootability-manifest.yaml` — bootability layer inventory
- `distro/install-media/build-contract.md` — install-image build I/O contract
- `distro/install-media/manifest.yaml` — payload authority
- `docs/ADR/0006-appliance-install-recovery-media.md`
- `docs/ADR/0005-appliance-runtime-operations.md`
- `docs/ADR/0004-identity-and-access.md`
- `docs/ops/restore-safety-contract.md`
- `distro/recovery-media/README.md` — separate recovery deliverable
