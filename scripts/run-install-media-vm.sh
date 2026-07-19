#!/usr/bin/env bash
# Non-destructive VM launcher for the Vyntrio dual-mode (BIOS + UEFI) install image.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_IMAGE="${ROOT}/distro/release/staging/vyntrio-install-media.img"
FALLBACK_IMAGE="${ROOT}/distro/install-media/build/vyntrio-install-media.img"

IMAGE=""
DRY_RUN=false
MEMORY="2048"
FIRMWARE="uefi" # uefi | bios — UEFI is the preferred smoke path for modern hardware

usage() {
	cat <<EOF
Usage: $(basename "$0") [options]

Boot the Vyntrio dual-mode install image in a local VM (non-destructive — reads the .img file).

Options:
  --image PATH       Image file (default: staged, else build output)
  --firmware MODE    uefi (default, OVMF) or bios (legacy SeaBIOS)
  --memory MB        RAM in MiB (default: 2048)
  --dry-run          Print the qemu command without running

Requires: qemu-system-x86_64 (preferred) or qemu-system-i386
UEFI: OVMF firmware (e.g. /usr/share/OVMF/OVMF_CODE_4M.fd or OVMF_CODE.fd)
EOF
}

fail() { echo "run-install-media-vm: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
	case "$1" in
	--image)
		shift
		IMAGE="${1:-}"
		;;
	--firmware)
		shift
		FIRMWARE="${1:-}"
		;;
	--memory)
		shift
		MEMORY="${1:-}"
		;;
	--dry-run)
		DRY_RUN=true
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		fail "unknown argument: $1"
		;;
	esac
	shift
done

case "${FIRMWARE}" in
uefi | bios) ;;
*) fail "firmware must be uefi or bios (got: ${FIRMWARE})" ;;
esac

if [[ -z "${IMAGE}" ]]; then
	if [[ -f "${DEFAULT_IMAGE}" ]]; then
		IMAGE="${DEFAULT_IMAGE}"
	elif [[ -f "${FALLBACK_IMAGE}" ]]; then
		IMAGE="${FALLBACK_IMAGE}"
	else
		fail "no image found; run 'make release-install-media-stage' or pass --image"
	fi
fi

[[ -f "${IMAGE}" ]] || fail "image not found: ${IMAGE}"

QEMU=""
for candidate in qemu-system-x86_64 qemu-system-i386; do
	if command -v "${candidate}" >/dev/null 2>&1; then
		QEMU="${candidate}"
		break
	fi
done

find_ovmf() {
	local candidate
	for candidate in \
		/usr/share/OVMF/OVMF_CODE_4M.fd \
		/usr/share/OVMF/OVMF_CODE.fd \
		/usr/share/edk2/ovmf/OVMF_CODE.fd \
		/usr/share/qemu/OVMF.fd; do
		if [[ -f "${candidate}" ]]; then
			printf '%s' "${candidate}"
			return 0
		fi
	done
	return 1
}

print_qemu_cmd() {
	local qemu="$1"
	local accel_note=""
	[[ -e /dev/kvm && -r /dev/kvm ]] && accel_note=" -accel kvm"
	if [[ "${FIRMWARE}" == "uefi" ]]; then
		local ovmf
		ovmf="$(find_ovmf)" || fail "OVMF firmware not found (install ovmf / edk2-ovmf)"
		printf '%s%s -machine q35 -m %s -drive if=pflash,format=raw,readonly=on,file=%s -drive file=%s,format=raw,if=virtio -nographic\n' \
			"${qemu}" "${accel_note}" "${MEMORY}" "${ovmf}" "${IMAGE}"
	else
		printf '%s%s -machine pc -m %s -drive file=%s,format=raw,if=ide -boot c -nographic\n' \
			"${qemu}" "${accel_note}" "${MEMORY}" "${IMAGE}"
	fi
}

if ${DRY_RUN}; then
	echo "run-install-media-vm: image=${IMAGE}"
	echo "run-install-media-vm: firmware=${FIRMWARE}"
	echo "run-install-media-vm: dry-run command:"
	echo -n "  "
	print_qemu_cmd "${QEMU:-qemu-system-x86_64}"
	exit 0
fi

if [[ -z "${QEMU}" ]]; then
	fail "install qemu-system-x86_64 or qemu-system-i386"
fi

echo "run-install-media-vm: starting VM firmware=${FIRMWARE} (Ctrl+A X to quit in nographic mode)"

ACCEL=()
if [[ -e /dev/kvm && -r /dev/kvm ]]; then
	ACCEL=(-accel kvm)
fi

if [[ "${FIRMWARE}" == "uefi" ]]; then
	OVMF="$(find_ovmf)" || fail "OVMF firmware not found (install ovmf / edk2-ovmf)"
	exec "${QEMU}" -machine q35 "${ACCEL[@]}" -m "${MEMORY}" \
		-drive "if=pflash,format=raw,readonly=on,file=${OVMF}" \
		-drive "file=${IMAGE},format=raw,if=virtio" -nographic
fi

exec "${QEMU}" -machine pc "${ACCEL[@]}" -m "${MEMORY}" \
	-drive "file=${IMAGE},format=raw,if=ide" -boot c -nographic
