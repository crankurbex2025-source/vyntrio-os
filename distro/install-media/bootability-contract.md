# Install-image bootability initialization contract

- **Status:** Bootability scaffold only — **no bootable image builder is implemented**
- **Block:** 9
- **Slice:** 9.9 (documentation)
- **Authority:** `docs/ADR/0006-appliance-install-recovery-media.md` (install media
  sections B, D). Layer inventory: `distro/install-media/bootability-manifest.yaml`.
  Envelope contract: `distro/install-media/envelope-contract.md`.
  Envelope assembly: Slice 9.8 (`make install-media-envelope`).

## 1. Explicit status

**Slice 9.9 (this document):** defines the **bootable initialization boundary**
— what must be added to the local envelope before it becomes a bootable ISO/USB
deliverable. This is a **declarative contract only**. No bootloader, kernel,
initrd, live-root filesystem, ISO/USB generator, or installer runs in this slice.

**Slice 9.8 (implemented):** `make install-media-envelope` assembles
`distro/install-media/envelope/` with `boot/`, `live_root/`, and `payload/`
layers. `boot/` and `live_root/` contain **deferred placeholders only**.

**Bootable initialization is not installation.** Making an image bootable only
enables firmware to start a live runtime. Copying `payload/` to a target disk,
partitioning, enabling services, and ADR-0004 bootstrap remain **deferred**
installer slices.

Recovery-image bootability is a **separate future contract** under
`distro/recovery-media/` — not defined here.

## 2. What bootable initialization adds

Slice 9.8 envelope assembly packages structure. Bootable initialization (future
executable slices) **populates** deferred layers and **emits** a distribution
artifact:

| Step | Beyond 9.8 envelope | Status |
|------|---------------------|--------|
| `boot/` population | Bootloader config, kernel, initrd | **Deferred** |
| `live_root/` population | Minimal live filesystem and init | **Deferred** |
| Layout validation | Verify boot/live_root/payload before emission | **Deferred** |
| Image emission | Bootable `.iso` or raw USB image file | **Deferred** |

The `payload/` layer from Slice 9.6/9.8 is **carried forward unchanged** in
concept. Bootability does not add secrets, persistent DB, or target-disk state.

## 3. Relationship to staging and envelope assembly

```
Slice 9.6                    Slice 9.8                      Slice 9.9+ (deferred)
─────────                    ─────────                      ─────────────────────
staging/payload/      →      envelope/payload/       →      same payload/ in image
                             envelope/boot/ (placeholder) → populated boot/
                             envelope/live_root/ (placeholder) → populated live_root/
                                                          → bootable .iso / .img
```

| Artifact | Location | Committed | Role |
|----------|----------|-----------|------|
| Staged payloads | `distro/install-media/staging/payload/` | **No** | Slice 9.6 disposable input |
| Assembled envelope | `distro/install-media/envelope/` | **No** | Slice 9.8 disposable tree |
| Bootability manifest | `distro/install-media/bootability-manifest.yaml` | **Yes** | Declarative bootability contract |
| Future bootable image | build output (deferred) | **No** | Ephemeral distribution artifact |

Bootability initialization **reads** the local envelope; it must not write to
target disk paths (`/var/lib/vyntrio/`, deployed `/etc/vyntrio/`, etc.).

## 4. Bootability vs installer execution

| Concern | Bootable initialization | Installer execution (deferred) |
|---------|-------------------------|--------------------------------|
| Purpose | Start live runtime from media | Copy payloads to target disk |
| Operates on | Image envelope layers | Target disk partitions/paths |
| `payload/` role | Packaged for later copy | Copied to `/usr/bin`, `/etc/`, etc. |
| Services | None started on media | `vyntrio-api.service` enabled on target |
| Bootstrap | **Not invoked** | ADR-0004 after installed service starts |
| Persistent state | **None on media** | Created on target disk only |

## 5. Install vs recovery vs target disk

| Concern | Install bootability (this contract) | Recovery media | Target disk |
|---------|-------------------------------------|----------------|-------------|
| Deliverable | Greenfield install image | Offline recovery image | Running appliance |
| Bootability manifest | `bootability-manifest.yaml` | Future separate contract | N/A |
| Payload model | Copy **to** disk at install | Run tooling **from** live env | Persistent state home |
| Combined image | **Forbidden** (ADR-0006) | **Forbidden** | N/A |

## 6. Deferred executable slices

The following require future implementation slices after 9.9:

- Boot loader and UEFI/BIOS configuration
- Live root filesystem composition (minimal Debian/base)
- Bootable layout validation target
- ISO/USB image file generation (`xorriso`, `grub`, `dd`, etc.)
- `vyntrio-installer` execution
- `bootability-record.yaml` build metadata emission

## 7. Out of scope (Slice 9.9)

- Bootloader, kernel, or initrd implementation
- Live-root filesystem build or execution
- ISO/USB generation implementation
- Partitioning, filesystem creation, disk encryption
- Installer execution or install-progress reporting
- Recovery-media bootability changes
- API/dashboard install status
- Host paths, services, secrets, or backup artifacts at build time

## 8. Related documents

- `distro/install-media/bootability-manifest.yaml` — declarative bootability inventory
- `distro/install-media/envelope-manifest.yaml` — envelope layer contract
- `distro/install-media/envelope-contract.md` — envelope boundary
- `distro/install-media/build-contract.md` — install-image build I/O contract
- `distro/install-media/manifest.yaml` — payload authority
- `docs/ADR/0006-appliance-install-recovery-media.md`
- `docs/ADR/0007-appliance-installer-contract.md` — installer execution contract (Block 10)
- `docs/ADR/0005-appliance-runtime-operations.md`
- `docs/ADR/0004-identity-and-access.md`
- `docs/ops/restore-safety-contract.md`
- `distro/recovery-media/README.md` — separate recovery deliverable
