#!/usr/bin/env bash
# Verifies installer preflight behavior (Slice 10.3).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
PAYLOAD_ROOT="${ENVELOPE_ROOT}/payload"

if [[ ! -d "${PAYLOAD_ROOT}" ]]; then
	echo "installer preflight test: envelope payload missing; run 'make test-installer-preflight' first" >&2
	exit 1
fi

if ! "${ROOT}/scripts/installer-preflight.sh" >/dev/null; then
	echo "installer preflight test: preflight failed on valid envelope" >&2
	exit 1
fi

# Fail-closed: missing envelope.
if VYNTRIO_INSTALL_ENVELOPE_ROOT="/tmp/vyntrio-missing-envelope-$$" ./scripts/installer-preflight.sh >/dev/null 2>&1; then
	echo "installer preflight test: expected failure for missing envelope" >&2
	exit 1
fi

# Fail-closed: extra payload file.
trap 'rm -rf "${TMP_ENVELOPE}"' EXIT
TMP_ENVELOPE="$(mktemp -d)"
mkdir -p "${TMP_ENVELOPE}/payload"
cp -a "${PAYLOAD_ROOT}/." "${TMP_ENVELOPE}/payload/"
cp "${ENVELOPE_ROOT}/ENVELOPE.txt" "${TMP_ENVELOPE}/"
touch "${TMP_ENVELOPE}/payload/extra-forbidden.txt"
if VYNTRIO_INSTALL_ENVELOPE_ROOT="${TMP_ENVELOPE}" ./scripts/installer-preflight.sh >/dev/null 2>&1; then
	echo "installer preflight test: expected failure for extra payload file" >&2
	exit 1
fi

echo "installer preflight test: ok"
