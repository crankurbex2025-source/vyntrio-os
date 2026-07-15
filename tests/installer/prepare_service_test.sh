#!/usr/bin/env bash
# Verifies installer service preparation behavior (Slice 10.8).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SANDBOX="${ROOT}/distro/installer/target-sandbox"

if ! "${ROOT}/scripts/installer-prepare-service.sh" >/dev/null; then
	echo "installer prepare service test: prepare script failed" >&2
	exit 1
fi

if [[ ! -f "${SANDBOX}/SERVICE_PREP.txt" ]]; then
	echo "installer prepare service test: SERVICE_PREP.txt missing" >&2
	exit 1
fi

if ! grep -q '^enablement_status: prepared_not_enabled$' "${SANDBOX}/SERVICE_PREP.txt"; then
	echo "installer prepare service test: enablement must remain prepared only" >&2
	exit 1
fi

if ! grep -q '^service_started: false$' "${SANDBOX}/SERVICE_PREP.txt"; then
	echo "installer prepare service test: service must not be started" >&2
	exit 1
fi

if ! grep -q '^host_services_started: false$' "${SANDBOX}/SERVICE_PREP.txt"; then
	echo "installer prepare service test: host service guard missing" >&2
	exit 1
fi

if [[ ! -f "${SANDBOX}/etc/systemd/system/vyntrio-api.service.enable-prep" ]]; then
	echo "installer prepare service test: enable-prep marker missing" >&2
	exit 1
fi

# Fail-closed: missing payload copy record blocks service prep.
TMP_SANDBOX="$(mktemp -d)"
trap 'rm -rf "${TMP_SANDBOX}"' EXIT
mkdir -p "${TMP_SANDBOX}/etc/systemd/system"
cp "${SANDBOX}/etc/systemd/system/vyntrio-api.service" "${TMP_SANDBOX}/etc/systemd/system/"

if VYNTRIO_INSTALL_TARGET_ROOT="${TMP_SANDBOX}" "${ROOT}/scripts/installer-prepare-service.sh" >/dev/null 2>&1; then
	echo "installer prepare service test: expected failure without payload copy gate" >&2
	exit 1
fi

echo "installer prepare service test: ok"
