#!/usr/bin/env bash
# Verifies installer directory mutation (Slice 10.6).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SANDBOX="${ROOT}/distro/installer/target-sandbox"

readonly -a EXPECTED_DIRS=(
	"etc/vyntrio"
	"var/lib/vyntrio"
	"var/lib/vyntrio/backups"
)

if ! "${ROOT}/scripts/installer-mutate-directories.sh" >/dev/null; then
	echo "installer mutate directories test: mutation script failed" >&2
	exit 1
fi

for rel in "${EXPECTED_DIRS[@]}"; do
	if [[ ! -d "${SANDBOX}/${rel}" ]]; then
		echo "installer mutate directories test: missing directory: ${rel}" >&2
		exit 1
	fi
	if [[ -n "$(find "${SANDBOX}/${rel}" -type f -print -quit)" ]]; then
		echo "installer mutate directories test: unexpected file in ${rel}" >&2
		exit 1
	fi
done

if [[ ! -f "${SANDBOX}/MUTATION_RECORD.txt" ]]; then
	echo "installer mutate directories test: MUTATION_RECORD.txt missing" >&2
	exit 1
fi

if ! grep -q '^host_paths_mutated: false$' "${SANDBOX}/MUTATION_RECORD.txt"; then
	echo "installer mutate directories test: host path guard missing" >&2
	exit 1
fi

if ! grep -q '^payload_copy: deferred$' "${SANDBOX}/MUTATION_RECORD.txt"; then
	echo "installer mutate directories test: payload copy must remain deferred" >&2
	exit 1
fi

# Fail-closed: unsafe target root rejected.
if VYNTRIO_INSTALL_TARGET_ROOT="/tmp/vyntrio-unsafe-target-$$" "${ROOT}/scripts/installer-mutate-directories.sh" >/dev/null 2>&1; then
	echo "installer mutate directories test: expected failure for unsafe target root" >&2
	exit 1
fi

echo "installer mutate directories test: ok"
