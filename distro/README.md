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

See `docs/03_ROADMAP.md` and `cursor-prompts/phase-02-linux-base.md`.

## Out of scope for Phase 1

No ISO builds, no Debian customization, no installer UI in this phase.
