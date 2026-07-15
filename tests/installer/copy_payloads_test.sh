#!/usr/bin/env bash
# Verifies installer payload copy behavior (Slice 10.7).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SANDBOX="${ROOT}/distro/installer/target-sandbox"

readonly -a EXPECTED=(
	"usr/bin/vyntrio-api"
	"usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service"
	"usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf"
	"etc/vyntrio/config.toml"
)

if ! "${ROOT}/scripts/installer-copy-payloads.sh" >/dev/null; then
	echo "installer copy payloads test: copy script failed" >&2
	exit 1
fi

for rel in "${EXPECTED[@]}"; do
	path="${SANDBOX}/${rel}"
	if [[ ! -f "${path}" ]]; then
		echo "installer copy payloads test: missing copied file: ${rel}" >&2
		exit 1
	fi
done

mapfile -t sandbox_files < <(find "${SANDBOX}" -type f ! -name 'MUTATION_RECORD.txt' ! -name 'PAYLOAD_COPY.txt' | LC_ALL=C sort)
if [[ "${#sandbox_files[@]}" -ne "${#EXPECTED[@]}" ]]; then
	echo "installer copy payloads test: unexpected payload file count" >&2
	exit 1
fi

if [[ ! -f "${SANDBOX}/PAYLOAD_COPY.txt" ]]; then
	echo "installer copy payloads test: PAYLOAD_COPY.txt missing" >&2
	exit 1
fi

if ! grep -q '^host_paths_mutated: false$' "${SANDBOX}/PAYLOAD_COPY.txt"; then
	echo "installer copy payloads test: host path guard missing" >&2
	exit 1
fi

if ! grep -q '^bootstrap_handoff: deferred$' "${SANDBOX}/PAYLOAD_COPY.txt"; then
	echo "installer copy payloads test: bootstrap must remain deferred" >&2
	exit 1
fi

# Fail-closed: missing envelope blocks copy.
if VYNTRIO_INSTALL_ENVELOPE_ROOT="/tmp/vyntrio-missing-envelope-$$" "${ROOT}/scripts/installer-copy-payloads.sh" >/dev/null 2>&1; then
	echo "installer copy payloads test: expected failure without preflight" >&2
	exit 1
fi

echo "installer copy payloads test: ok"
