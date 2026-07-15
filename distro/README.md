# distro

Linux distribution image construction: ISO, initramfs, package lists, first-boot
hooks, and **systemd deployment artifacts** for appliance operation.

**Status:** Phase 0.2 (Installer & Base OS) not started. **Block 7, Slice 7.3**
adds production-oriented systemd artifacts under `distro/systemd/` (service
unit, sysusers, tmpfiles). See `distro/systemd/README.md`.

**Block 9 / Slice 9.2:** Install- und Recovery-Medien-Vertrag in
`docs/ADR/0006-appliance-install-recovery-media.md`. `distro/systemd/` sind
**Runtime-Installationsartefakte** auf der Zielplatte — nicht dasselbe wie
Boot-USB/ISO-Images (zukünftige Slices). Recovery-Medien sind ein separates
Deliverable vom Install-Medium.

**Block 9 / Slice 9.3:** Deklaratives Install-Media-Gerüst unter
`distro/install-media/` (`manifest.yaml`, `config.toml.template`, README) —
Payload-Vertrag nur; kein Boot, keine Partitionierung, kein Installer-Lauf.

**Block 9 / Slice 9.4:** Deklaratives Recovery-Media-Gerüst unter
`distro/recovery-media/` (`manifest.yaml`, README) — Offline-Recovery-Tooling-
Vertrag nur; getrennt von `install-media/`; kein Restore-CLI, kein Boot/ISO-Build.

**Block 9 / Slice 9.5:** Install-Image-Build-Vertrag in
`distro/install-media/build-contract.md` — Inputs/Outputs und Ein-/Ausschlüsse;
Manifeste bleiben deklarativ (noch kein Builder).

**Block 9 / Slice 9.6:** Lokales Payload-Staging via `make install-media-stage`
(`scripts/stage-install-media.sh` → `distro/install-media/staging/payload/`);
nur Manifest-Payloads; kein ISO/USB, kein Zielplatten-Schreiben.

**Block 9 / Slice 9.7:** Live/Boot-Envelope-Gerüst in
`distro/install-media/envelope-contract.md` und `envelope-manifest.yaml` —
deklarative Schichten (`boot/`, `live_root/`, `payload/`); Verweis auf
`staging/payload/`; kein ISO/USB-Builder, kein Boot/Live-Root-Lauf.

**Block 9 / Slice 9.8:** Lokale Envelope-Assembly via `make install-media-envelope`
(`scripts/assemble-install-media-envelope.sh` → `distro/install-media/envelope/`);
konsumiert gestagte Payloads; `boot/`/`live_root/` nur Platzhalter; kein ISO/USB,
kein Zielplatten-Schreiben.

**Block 9 / Slice 9.9:** Bootability-Initialisierungs-Gerüst in
`distro/install-media/bootability-contract.md` und `bootability-manifest.yaml` —
deklarative Grenze für bootfähige ISO/USB nach Envelope-Assembly; kein Bootloader,
kein Live-Root, kein Installer-Lauf.

**Block 10 / Slice 10.3:** Read-only Installer-Preflight unter `distro/installer/`
via `make installer-preflight`; validiert Envelope/Payload gegen ADR-0007; kein
Zielplatten-Schreiben, kein Bootstrap.

**Block 10 / Slice 10.4:** Zielplatten-Layout-Gerüst unter `distro/installer/`
(`target-layout-manifest.yaml`, `target-layout-contract.md`); read-only
Validierung via `make installer-layout-plan`; Partitionierung deferred; kein
Zielplatten-Schreiben.

**Block 10 / Slice 10.5:** Preflight-gated Mutation-Dry-Run-Stub via
`make installer-mutation-stub`; fail-closed ohne Preflight; kein
Zielplatten-Schreiben.

**Block 10 / Slice 10.6:** Erste Ziel-Mutation — leere State-Verzeichnisse in
`distro/installer/target-sandbox/` via `make installer-mutate-directories`;
preflight-gated; kein Host-Pfad-Schreiben, kein Payload-Copy.

**Block 10 / Slice 10.7:** Manifest-Payload-Copy nach `target-sandbox/` via
`make installer-copy-payloads`; nur Allowlist-Payloads; kein Service-Enablement.

**Block 10 / Slice 10.8:** Service-Enablement-Vorbereitung via
`make installer-prepare-service`; Prep-Artefakte in `target-sandbox/`; kein
Service-Start, kein Bootstrap.

See `docs/03_ROADMAP.md` and `cursor-prompts/phase-02-linux-base.md`.

## Out of scope for Phase 1

No ISO builds, no Debian customization, no installer UI in this phase.
