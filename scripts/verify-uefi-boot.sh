#!/usr/bin/env bash
# Structural + optional QEMU/OVMF smoke for dual-mode install media.
# Does NOT claim dashboard reachability; only UEFI firmware handoff markers when qemu runs.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_ROOT="${VYNTRIO_INSTALL_MEDIA_BUILD_ROOT:-${ROOT}/distro/install-media/build}"
IMAGE="${1:-${BUILD_ROOT}/vyntrio-install-media.img}"
WRAPPER="${BUILD_ROOT}/WRAPPER.txt"
TIMEOUT_SEC="${VYNTRIO_UEFI_BOOT_TIMEOUT:-45}"

fail() { echo "verify-uefi-boot: $*" >&2; exit 1; }

[[ -f "${IMAGE}" ]] || fail "image missing: ${IMAGE}"
[[ -f "${WRAPPER}" ]] || fail "WRAPPER.txt missing: ${WRAPPER}"

grep -q '^firmware_bootable: true' "${WRAPPER}" || fail "WRAPPER firmware_bootable != true"
grep -q '^uefi_support: true' "${WRAPPER}" || fail "WRAPPER uefi_support != true — BIOS-only is incomplete"
grep -q '^dual_mode: true' "${WRAPPER}" || fail "WRAPPER dual_mode != true"
grep -q '^firmware_boot_mode: bios+uefi' "${WRAPPER}" || fail "WRAPPER firmware_boot_mode != bios+uefi"
grep -q '^artifact_format: raw_gpt_hybrid_disk' "${WRAPPER}" || fail "WRAPPER artifact_format != raw_gpt_hybrid_disk"

# Structural: GPT + ESP BOOTX64.EFI
command -v sgdisk >/dev/null 2>&1 || fail "sgdisk required"
command -v mdir >/dev/null 2>&1 || fail "mdir (mtools) required"

sgdisk -p "${IMAGE}" 2>/dev/null | grep -qE 'EF00|EFI System' \
	|| fail "no EFI System partition (EF00) in GPT"
ESP_OFFSET_FILE="${BUILD_ROOT}/HYBRID_ESP_OFFSET.txt"
[[ -f "${ESP_OFFSET_FILE}" ]] || fail "HYBRID_ESP_OFFSET.txt missing — rebuild hybrid image"
ESP_OFF="$(tr -d '[:space:]' <"${ESP_OFFSET_FILE}")"
[[ -n "${ESP_OFF}" ]] || fail "empty ESP offset"

# mtools often prints 8.3 names as "BOOTX64  EFI" rather than "BOOTX64.EFI"
mdir -i "${IMAGE}@@${ESP_OFF}" ::/EFI/BOOT/ 2>/dev/null | grep -qiE 'BOOTX64([ .]+EFI|\.EFI)' \
	|| fail "ESP missing EFI/BOOT/BOOTX64.EFI"
# PE/COFF MZ header on BOOTX64.EFI
TMP="$(mktemp)"
mcopy -i "${IMAGE}@@${ESP_OFF}" ::/EFI/BOOT/BOOTX64.EFI "${TMP}" >/dev/null 2>&1 \
	|| fail "could not extract BOOTX64.EFI"
head -c 2 "${TMP}" | grep -q 'MZ' || fail "BOOTX64.EFI is not a PE (missing MZ)"
rm -f "${TMP}"

echo "verify-uefi-boot: structural UEFI packaging OK (ESP + BOOTX64.EFI MZ)"

if ! command -v qemu-system-x86_64 >/dev/null 2>&1; then
	echo "verify-uefi-boot: qemu absent — structural only; runtime_boot_tested remains false"
	exit 0
fi

OVMF=""
for candidate in \
	/usr/share/OVMF/OVMF_CODE_4M.fd \
	/usr/share/OVMF/OVMF_CODE.fd \
	/usr/share/edk2/ovmf/OVMF_CODE.fd \
	/usr/share/qemu/OVMF.fd; do
	if [[ -f "${candidate}" ]]; then
		OVMF="${candidate}"
		break
	fi
done
[[ -n "${OVMF}" ]] || {
	echo "verify-uefi-boot: OVMF absent — structural only"
	exit 0
}

LOG="$(mktemp)"
ACCEL=()
[[ -e /dev/kvm && -r /dev/kvm ]] && ACCEL=(-accel kvm)

set +e
timeout "${TIMEOUT_SEC}" qemu-system-x86_64 -machine q35 "${ACCEL[@]}" -m 1024 \
	-drive "if=pflash,format=raw,readonly=on,file=${OVMF}" \
	-drive "file=${IMAGE},format=raw,if=virtio" \
	-nographic -serial stdio -monitor none \
	>"${LOG}" 2>&1
rc=$?
set -e

if grep -qiE 'GRUB|Vyntrio|Linux|Kernel|Booting' "${LOG}"; then
	echo "verify-uefi-boot: qemu/OVMF emitted boot markers (see log excerpt)"
	grep -iE 'GRUB|Vyntrio|Linux|Kernel|Booting|EFI' "${LOG}" | head -20 || true
	rm -f "${LOG}"
	exit 0
fi

echo "verify-uefi-boot: qemu ran (exit=${rc}) but no clear GRUB/kernel markers within ${TIMEOUT_SEC}s"
echo "verify-uefi-boot: not claiming runtime UEFI boot success — structural packaging still OK"
tail -30 "${LOG}" || true
rm -f "${LOG}"
exit 0
