#!/usr/bin/env bash
# One-shot validation for install-media writer helpers.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
USB="${ROOT}/scripts/write-install-media-usb.sh"
VM="${ROOT}/scripts/run-install-media-vm.sh"
IMAGE="${ROOT}/distro/release/staging/vyntrio-install-media.img"

fail() {
	echo "write_install_media_usb_test: $*" >&2
	exit 1
}

[[ -x "${USB}" ]] || fail "write-install-media-usb.sh not executable"
[[ -x "${VM}" ]] || fail "run-install-media-vm.sh not executable"
[[ -f "${IMAGE}" ]] || fail "staged image missing — run make release-install-media-stage first"

"${USB}" --dry-run
"${USB}" --image "${IMAGE}" --dry-run

if "${USB}" --device /etc/hosts --dry-run >/dev/null 2>&1; then
	fail "expected non-block device rejection"
fi

"${VM}" --dry-run
"${VM}" --image "${IMAGE}" --dry-run

echo "write_install_media_usb_test: pass"
