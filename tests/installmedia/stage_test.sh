#!/usr/bin/env bash
# Verifies install-media payload staging output (Slice 9.6).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PAYLOAD_ROOT="${ROOT}/distro/install-media/staging/payload"

readonly -a EXPECTED=(
	"usr/bin/vyntrio-api"
	"usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service"
	"usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf"
	"etc/vyntrio/config.toml"
)

if [[ ! -d "${PAYLOAD_ROOT}" ]]; then
	echo "installmedia stage test: staging directory missing; run 'make install-media-stage' first" >&2
	exit 1
fi

for rel in "${EXPECTED[@]}"; do
	path="${PAYLOAD_ROOT}/${rel}"
	if [[ ! -f "${path}" ]]; then
		echo "installmedia stage test: missing staged file: ${rel}" >&2
		exit 1
	fi
done

if [[ ! -f "${ROOT}/distro/install-media/staging/STAGING.txt" ]]; then
	echo "installmedia stage test: STAGING.txt missing" >&2
	exit 1
fi

echo "installmedia stage test: ok (${#EXPECTED[@]} payload files)"
