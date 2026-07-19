# UEFI dual-mode install media audit (2026-07-17)

Hard requirement: **UEFI support is mandatory.** BIOS-only media is incomplete.

## Audit result

| Check | Result |
|-------|--------|
| Prior primary artifact | `vyntrio-install-media-bios.img` — MBR + GRUB i386-pc only |
| ESP / `BOOTX64.EFI` on prior primary | **Absent** |
| Current primary | `vyntrio-install-media.img` — GPT hybrid `raw_gpt_hybrid_disk` |
| Firmware mode | `bios+uefi` |
| BIOS support | **yes** (EF02 + grub-bios-setup) |
| UEFI support | **yes** (EF00 ESP + `EFI/BOOT/BOOTX64.EFI` PE/MZ) |
| Dual-mode | **yes** — product baseline |
| Release staging of BIOS-only | **Refused** by `scripts/stage-release-install-media.sh` |
| Structural verification | Pass on wrap (`WRAPPER.txt`) |
| Runtime UEFI boot proven on every host | **No** — qemu/OVMF best-effort via `scripts/verify-uefi-boot.sh` |

## Truth rules applied

- Do **not** call install media complete if BIOS-only.
- Do **not** claim modern hardware support without UEFI packaging (ESP + BOOTX64.EFI).
- Dual-mode media is the required baseline; Secure Boot signing remains absent.

## Media creator / download

- Creator defaults to `vyntrio-install-media.img` and instructs UEFI or BIOS boot.
- `/download` states BIOS support, UEFI support, and dual-mode explicitly.

## Blockers (if hybrid cannot build)

Exact tools recorded in `WRAPPER.txt` when hybrid fails: `grub-efi-amd64-bin` (`grub-mkimage -O x86_64-efi`), `dosfstools`, `mtools`, `gdisk`/`sgdisk`. On this host those tools were present and hybrid emission succeeded.
