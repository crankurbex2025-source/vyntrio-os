#!/usr/bin/env bash
# Verifies install-media envelope assembly output (Slice 9.8).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"

readonly -a EXPECTED_PAYLOAD=(
	"usr/bin/vyntrio-api"
	"usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service"
	"usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf"
	"etc/vyntrio/config.toml"
)

if [[ ! -d "${ENVELOPE_ROOT}" ]]; then
	echo "installmedia envelope test: envelope directory missing; run 'make install-media-envelope' first" >&2
	exit 1
fi

for layer in boot live_root payload; do
	if [[ ! -d "${ENVELOPE_ROOT}/${layer}" ]]; then
		echo "installmedia envelope test: missing layer directory: ${layer}" >&2
		exit 1
	fi
done

for rel in "${EXPECTED_PAYLOAD[@]}"; do
	path="${ENVELOPE_ROOT}/payload/${rel}"
	if [[ ! -f "${path}" ]]; then
		echo "installmedia envelope test: missing payload file: ${rel}" >&2
		exit 1
	fi
done

mapfile -t payload_files < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
if [[ "${#payload_files[@]}" -ne "${#EXPECTED_PAYLOAD[@]}" ]]; then
	echo "installmedia envelope test: unexpected payload file count" >&2
	exit 1
fi

for path in "${payload_files[@]}"; do
	rel="${path#${ENVELOPE_ROOT}/payload/}"
	found=false
	for allowed in "${EXPECTED_PAYLOAD[@]}"; do
		if [[ "${rel}" == "${allowed}" ]]; then
			found=true
			break
		fi
	done
	if [[ "${found}" != true ]]; then
		echo "installmedia envelope test: non-manifest payload file: ${rel}" >&2
		exit 1
	fi
done

if [[ ! -f "${ENVELOPE_ROOT}/boot/LAYER.txt" ]]; then
	echo "installmedia envelope test: boot/LAYER.txt missing" >&2
	exit 1
fi

if [[ ! -f "${ENVELOPE_ROOT}/live_root/LAYER.txt" ]]; then
	echo "installmedia envelope test: live_root/LAYER.txt missing" >&2
	exit 1
fi

if [[ ! -f "${ENVELOPE_ROOT}/ENVELOPE.txt" ]]; then
	echo "installmedia envelope test: ENVELOPE.txt missing" >&2
	exit 1
fi

echo "installmedia envelope test: ok (${#EXPECTED_PAYLOAD[@]} payload files, 3 layers)"
