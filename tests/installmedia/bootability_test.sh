#!/usr/bin/env bash
# Verifies install-media bootability foundation output (Slice 9.11).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
STUB_IMAGE="${BUILD_ROOT}/vyntrio-install-media-NOT-BOOTABLE.stub.tar"

readonly -a EXPECTED_PAYLOAD=(
	"usr/bin/vyntrio-api"
	"usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service"
	"usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf"
	"etc/vyntrio/config.toml"
)

if [[ ! -d "${ENVELOPE_ROOT}" ]]; then
	echo "installmedia bootability test: envelope missing; run 'make install-media-bootability' first" >&2
	exit 1
fi

for layer in boot live_root payload; do
	if [[ ! -d "${ENVELOPE_ROOT}/${layer}" ]]; then
		echo "installmedia bootability test: missing layer: ${layer}" >&2
		exit 1
	fi
done

if [[ -f "${ENVELOPE_ROOT}/boot/LAYER.txt" ]] || [[ -f "${ENVELOPE_ROOT}/live_root/LAYER.txt" ]]; then
	echo "installmedia bootability test: deferred LAYER.txt still present" >&2
	exit 1
fi

for rel in \
	"boot/BOOT_LAYER.txt" \
	"boot/vmlinuz.stub" \
	"boot/initrd.img.stub" \
	"boot/loader/entries/vyntrio-install.conf.stub" \
	"live_root/LIVE_ROOT.txt" \
	"live_root/usr/lib/vyntrio/live-init.sh" \
	"live_root/etc/vyntrio/live.env" \
	"BOOTABILITY.txt"
do
	if [[ ! -e "${ENVELOPE_ROOT}/${rel}" ]]; then
		echo "installmedia bootability test: missing ${rel}" >&2
		exit 1
	fi
done

if [[ ! -x "${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio/live-init.sh" ]]; then
	echo "installmedia bootability test: live-init.sh must be executable" >&2
	exit 1
fi

if ! grep -q 'firmware_bootable: false' "${ENVELOPE_ROOT}/BOOTABILITY.txt"; then
	echo "installmedia bootability test: BOOTABILITY.txt must declare firmware_bootable: false" >&2
	exit 1
fi

if ! grep -q 'usb_creator: deferred' "${ENVELOPE_ROOT}/BOOTABILITY.txt"; then
	echo "installmedia bootability test: BOOTABILITY.txt must declare usb_creator: deferred" >&2
	exit 1
fi

for rel in "${EXPECTED_PAYLOAD[@]}"; do
	if [[ ! -f "${ENVELOPE_ROOT}/payload/${rel}" ]]; then
		echo "installmedia bootability test: missing payload: ${rel}" >&2
		exit 1
	fi
done

mapfile -t payload_files < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
if [[ "${#payload_files[@]}" -ne "${#EXPECTED_PAYLOAD[@]}" ]]; then
	echo "installmedia bootability test: unexpected payload file count" >&2
	exit 1
fi

if [[ ! -f "${STUB_IMAGE}" ]]; then
	echo "installmedia bootability test: missing image stub: ${STUB_IMAGE}" >&2
	exit 1
fi

if [[ ! -f "${BUILD_ROOT}/IMAGE_STUB.txt" ]]; then
	echo "installmedia bootability test: IMAGE_STUB.txt missing" >&2
	exit 1
fi

if ! grep -q 'firmware_bootable: false' "${BUILD_ROOT}/IMAGE_STUB.txt"; then
	echo "installmedia bootability test: IMAGE_STUB.txt must declare firmware_bootable: false" >&2
	exit 1
fi

if ! tar -tf "${STUB_IMAGE}" | grep -q 'NOT_BOOTABLE.txt'; then
	echo "installmedia bootability test: stub archive missing NOT_BOOTABLE.txt" >&2
	exit 1
fi

# Fail-closed: initializer must reject missing envelope.
tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT
if VYNTRIO_INSTALL_ENVELOPE_ROOT="${tmpdir}/missing" \
	VYNTRIO_INSTALL_MEDIA_BUILD_ROOT="${tmpdir}/build" \
	bash "${ROOT}/scripts/initialize-install-media-bootability.sh" 2>/dev/null; then
	echo "installmedia bootability test: expected failure when envelope missing" >&2
	exit 1
fi

echo "installmedia bootability test: ok (boot+live_root stubs, image stub, firmware_bootable=false)"
