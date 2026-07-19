#!/usr/bin/env bash
# Verify the Vyntrio USB appliance image is a real appliance artifact (not a stub).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BUILD="${ROOT}/distro/install-media/build"
IMG="${BUILD}/vyntrio-install-media.img"
REC="${BUILD}/APPLIANCE.txt"
ESP_OFF_FILE="${BUILD}/HYBRID_ESP_OFFSET.txt"

fail() { echo "appliance_image_test: FAIL $*" >&2; exit 1; }
pass() { echo "appliance_image_test: $*"; }

[[ -f "${IMG}" ]] || fail "missing ${IMG}"
[[ -f "${REC}" ]] || fail "missing ${REC}"
grep -q '^host_debian_initrd: false' "${REC}" || fail "still host debian initrd"
grep -q '^vyntrio_initramfs: true' "${REC}" || fail "vyntrio initramfs not recorded"
grep -q '^appliance_rootfs_on_media: true' "${REC}" || fail "rootfs not on media"
grep -q '^uefi_support: true' "${REC}" || fail "UEFI not recorded"
grep -q '^secure_boot: unsupported' "${REC}" || fail "Secure Boot status not explicit"

SIZE="$(stat -c '%s' "${IMG}")"
[[ "${SIZE}" -ge $((400 * 1024 * 1024)) ]] || fail "image too small for appliance (${SIZE} bytes)"

# GPT partitions: BIOS + ESP + data
sgdisk -p "${IMG}" | grep -q 'EF02' || fail "BIOS boot partition missing"
sgdisk -p "${IMG}" | grep -q 'EF00' || fail "ESP missing"
sgdisk -p "${IMG}" | grep -q '8300\|Linux filesystem' || fail "appliance data partition missing"

ESP_OFF="$(cat "${ESP_OFF_FILE}")"
export MTOOLS_SKIP_CHECK=1
mdir -i "${IMG}@@${ESP_OFF}" ::/EFI/BOOT | grep -qi BOOTX64 || fail "BOOTX64.EFI missing"
mdir -i "${IMG}@@${ESP_OFF}" ::/boot | grep -qi vmlinuz || fail "kernel missing on ESP"
mdir -i "${IMG}@@${ESP_OFF}" ::/boot | grep -qi initrd || fail "initrd missing on ESP"

# Extract initrd and prove it is Vyntrio early (not host debian microcode-only outer).
WORK="$(mktemp -d)"
mcopy -i "${IMG}@@${ESP_OFF}" ::/boot/initrd.img "${WORK}/initrd.img"
gzip -dc "${WORK}/initrd.img" 2>/dev/null | cpio -t 2>/dev/null | grep -qE '^\.?/?init$' \
	|| fail "initrd has no /init"
gzip -dc "${WORK}/initrd.img" 2>/dev/null | cpio -t 2>/dev/null | grep -q 'bin/busybox' \
	|| fail "initrd missing busybox"
# Must NOT be the huge host debian multi-part initrd signature alone.
LISTING="$(gzip -dc "${WORK}/initrd.img" 2>/dev/null | cpio -t 2>/dev/null || true)"
echo "${LISTING}" | grep -qi 'GenuineIntel.bin' && fail "initrd looks like host Debian microcode initrd"
echo "${LISTING}" | grep -q 'switch_root\|squashfs\|vyntrio' \
	|| { echo "${LISTING}" | grep -q 'mnt/squash' || fail "early initrd missing appliance mount path"; }

# Data partition contains squashfs
DATA_OFF="$(cat "${BUILD}/APPLIANCE_DATA_OFFSET.txt")"
DATA_START_SEC=$((DATA_OFF / 512))
# Probe squashfs magic via dd
dd if="${IMG}" bs=512 skip="${DATA_START_SEC}" count=2048 status=none 2>/dev/null | strings | grep -q vyntrio \
	|| true
# Mount data via loop offset
LOOP="$(losetup -f --show -o "${DATA_OFF}" "${IMG}")"
MNT="${WORK}/mnt"
mkdir -p "${MNT}"
mount -o ro "${LOOP}" "${MNT}" || fail "cannot mount appliance data partition"
[[ -f "${MNT}/vyntrio/system.squashfs" ]] || fail "system.squashfs missing on media"
[[ -d "${MNT}/vyntrio/config" ]] || fail "persistence config dir missing"
[[ -d "${MNT}/vyntrio/state" ]] || fail "persistence state dir missing"
# Squashfs contains vyntrio-api
unsquashfs -l "${MNT}/vyntrio/system.squashfs" 2>/dev/null | grep -q 'usr/bin/vyntrio-api' \
	|| fail "squashfs missing vyntrio-api"
unsquashfs -l "${MNT}/vyntrio/system.squashfs" 2>/dev/null | grep -q 'usr/lib/vyntrio/firstboot.sh' \
	|| fail "squashfs missing firstboot.sh"
umount "${MNT}"
losetup -d "${LOOP}"
rm -rf "${WORK}"

pass "ok size=${SIZE} sha=$(sha256sum "${IMG}" | awk '{print $1}')"
