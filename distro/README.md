# distro

Linux distribution image construction: ISO, initramfs, package lists, first-boot
hooks, and **systemd deployment artifacts** for appliance operation.

**Status:** Phase 0.2 (Installer & Base OS) not started. **Block 7, Slice 7.3**
adds production-oriented systemd artifacts under `distro/systemd/` (service
unit, sysusers, tmpfiles). See `distro/systemd/README.md`.

See `docs/03_ROADMAP.md` and `cursor-prompts/phase-02-linux-base.md`.

## Out of scope for Phase 1

No ISO builds, no Debian customization, no installer UI in this phase.
