#!/usr/bin/env bash
# Guided USB writer for the Vyntrio BIOS raw install image.
# Destructive: overwrites the entire target block device.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_IMAGE="${ROOT}/distro/release/staging/vyntrio-install-media.img"
FALLBACK_IMAGE="${ROOT}/distro/install-media/build/vyntrio-install-media.img"

IMAGE=""
DEVICE=""
DRY_RUN=false
ASSUME_YES=false

usage() {
	cat <<EOF
Usage: $(basename "$0") --device /dev/sdX [options]

Write the Vyntrio BIOS install image to a USB stick or other block device.

Options:
  --image PATH     Image file (default: staged image, else build output)
  --device PATH    Target block device (required for write)
  --dry-run        Print actions without writing
  --yes            Skip interactive confirmation (still requires --device)

Safety:
  - Refuses to write if the device is mounted
  - Requires typing the device path to confirm (unless --yes)
  - Verifies SHA-256 against release-manifest.json when present beside the image

Example:
  $(basename "$0") --device /dev/sdX --dry-run
  sudo $(basename "$0") --device /dev/sdX
EOF
}

fail() { echo "write-install-media-usb: $*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
	case "$1" in
	--image)
		shift
		IMAGE="${1:-}"
		;;
	--device)
		shift
		DEVICE="${1:-}"
		;;
	--dry-run)
		DRY_RUN=true
		;;
	--yes)
		ASSUME_YES=true
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

MANIFEST="${IMAGE%/*}/release-manifest.json"
if [[ -f "${MANIFEST}" ]] && command -v jq >/dev/null 2>&1; then
	EXPECTED_SHA="$(jq -r '.artifacts[0].sha256 // empty' "${MANIFEST}")"
	if [[ -n "${EXPECTED_SHA}" && "${EXPECTED_SHA}" != "null" ]]; then
		ACTUAL_SHA="$(sha256sum "${IMAGE}" | awk '{print $1}')"
		[[ "${ACTUAL_SHA}" == "${EXPECTED_SHA}" ]] || fail "SHA-256 mismatch (manifest ${EXPECTED_SHA}, file ${ACTUAL_SHA})"
		echo "write-install-media-usb: SHA-256 verified against ${MANIFEST}"
	fi
fi

if [[ -z "${DEVICE}" ]]; then
	if ${DRY_RUN}; then
		echo "write-install-media-usb: dry-run ok — image=${IMAGE}"
		echo "write-install-media-usb: would write to --device /dev/sdX (not specified)"
		exit 0
	fi
	fail "--device is required (use --dry-run to validate image only)"
fi

[[ -b "${DEVICE}" ]] || fail "not a block device: ${DEVICE}"

if findmnt -n "${DEVICE}" >/dev/null 2>&1 || findmnt -n "${DEVICE}"* >/dev/null 2>&1; then
	fail "device appears mounted: ${DEVICE}"
fi

SIZE="$(stat -c '%s' "${IMAGE}")"
echo "write-install-media-usb: image=${IMAGE} (${SIZE} bytes)"
echo "write-install-media-usb: target=${DEVICE}"
echo "write-install-media-usb: WARNING — this erases all data on ${DEVICE}"

if ${DRY_RUN}; then
	echo "write-install-media-usb: dry-run — no data written"
	exit 0
fi

if ! ${ASSUME_YES}; then
	read -r -p "Type ${DEVICE} to confirm: " CONFIRM
	[[ "${CONFIRM}" == "${DEVICE}" ]] || fail "confirmation mismatch; aborted"
fi

if [[ "${EUID}" -ne 0 ]]; then
	fail "root privileges required; re-run with sudo"
fi

dd if="${IMAGE}" of="${DEVICE}" bs=4M conv=fsync status=progress
sync
echo "write-install-media-usb: complete — boot ${DEVICE} in BIOS/legacy mode"
