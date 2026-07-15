#!/usr/bin/env bash
# Verifies installer mutation stub behavior (Slice 10.5).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DRY_RUN_ROOT="${ROOT}/distro/installer/dry-run"

if ! "${ROOT}/scripts/installer-mutation-stub.sh" >/dev/null; then
	echo "installer mutation stub test: stub failed on valid preflight path" >&2
	exit 1
fi

if [[ ! -f "${DRY_RUN_ROOT}/MUTATION_STUB.txt" ]]; then
	echo "installer mutation stub test: MUTATION_STUB.txt missing" >&2
	exit 1
fi

if ! grep -q '^target_disk_mutation: false$' "${DRY_RUN_ROOT}/MUTATION_STUB.txt"; then
	echo "installer mutation stub test: dry-run record must deny target_disk_mutation" >&2
	exit 1
fi

if ! grep -q '^preflight_gate: passed$' "${DRY_RUN_ROOT}/MUTATION_STUB.txt"; then
	echo "installer mutation stub test: preflight gate not recorded" >&2
	exit 1
fi

# Fail-closed: preflight gate blocks stub when envelope is missing.
if VYNTRIO_INSTALL_ENVELOPE_ROOT="/tmp/vyntrio-missing-envelope-$$" "${ROOT}/scripts/installer-mutation-stub.sh" >/dev/null 2>&1; then
	echo "installer mutation stub test: expected failure without preflight" >&2
	exit 1
fi

echo "installer mutation stub test: ok"
