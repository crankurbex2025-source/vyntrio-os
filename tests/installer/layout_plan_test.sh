#!/usr/bin/env bash
# Verifies installer target-layout plan validation (Slice 10.4).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

if ! "${ROOT}/scripts/validate-installer-layout-plan.sh" >/dev/null; then
	echo "installer layout plan test: validation failed on canonical manifest" >&2
	exit 1
fi

BAD_MANIFEST="$(mktemp)"
trap 'rm -f "${BAD_MANIFEST}"' EXIT
grep -v 'target: /usr/bin/vyntrio-api' "${ROOT}/distro/installer/target-layout-manifest.yaml" >"${BAD_MANIFEST}"

if VYNTRIO_INSTALLER_LAYOUT_MANIFEST="${BAD_MANIFEST}" "${ROOT}/scripts/validate-installer-layout-plan.sh" >/dev/null 2>&1; then
	echo "installer layout plan test: expected failure for incomplete manifest" >&2
	exit 1
fi

echo "installer layout plan test: ok"
