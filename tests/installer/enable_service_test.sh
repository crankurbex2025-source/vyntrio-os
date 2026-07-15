#!/usr/bin/env bash
# Verifies installer controlled service enablement behavior (Slice 10.9).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SANDBOX="${ROOT}/distro/installer/target-sandbox"
ENABLE_SYMLINK="${SANDBOX}/etc/systemd/system/multi-user.target.wants/vyntrio-api.service"

if ! "${ROOT}/scripts/installer-enable-service.sh" >/dev/null; then
	echo "installer enable service test: enable script failed" >&2
	exit 1
fi

if [[ ! -f "${SANDBOX}/SERVICE_ENABLE.txt" ]]; then
	echo "installer enable service test: SERVICE_ENABLE.txt missing" >&2
	exit 1
fi

if ! grep -q '^enablement_status: enabled_not_started$' "${SANDBOX}/SERVICE_ENABLE.txt"; then
	echo "installer enable service test: service must be enabled but not started" >&2
	exit 1
fi

if ! grep -q '^service_started: false$' "${SANDBOX}/SERVICE_ENABLE.txt"; then
	echo "installer enable service test: service must not be started" >&2
	exit 1
fi

if ! grep -q '^host_services_started: false$' "${SANDBOX}/SERVICE_ENABLE.txt"; then
	echo "installer enable service test: host service guard missing" >&2
	exit 1
fi

if ! grep -q '^systemctl_invoked: false$' "${SANDBOX}/SERVICE_ENABLE.txt"; then
	echo "installer enable service test: systemctl must not be invoked" >&2
	exit 1
fi

if ! grep -q '^service_prep_gate: passed$' "${SANDBOX}/SERVICE_ENABLE.txt"; then
	echo "installer enable service test: service prep gate must be recorded" >&2
	exit 1
fi

if [[ ! -L "${ENABLE_SYMLINK}" ]]; then
	echo "installer enable service test: wants symlink missing" >&2
	exit 1
fi

link_target="$(readlink "${ENABLE_SYMLINK}")"
if [[ "${link_target}" != "../vyntrio-api.service" ]]; then
	echo "installer enable service test: unexpected wants symlink target: ${link_target}" >&2
	exit 1
fi

if ! grep -q '^enabled: true$' "${SANDBOX}/etc/systemd/system/vyntrio-api.service.enable-prep"; then
	echo "installer enable service test: enable marker must show enabled" >&2
	exit 1
fi

# Fail-closed: missing service prep record blocks enablement.
TMP_SANDBOX="$(mktemp -d)"
trap 'rm -rf "${TMP_SANDBOX}"' EXIT
mkdir -p "${TMP_SANDBOX}/etc/systemd/system"
cp "${SANDBOX}/etc/systemd/system/vyntrio-api.service" "${TMP_SANDBOX}/etc/systemd/system/"
cp "${SANDBOX}/PAYLOAD_COPY.txt" "${TMP_SANDBOX}/PAYLOAD_COPY.txt"

if VYNTRIO_INSTALL_TARGET_ROOT="${TMP_SANDBOX}" "${ROOT}/scripts/installer-enable-service.sh" >/dev/null 2>&1; then
	echo "installer enable service test: expected failure without service prep gate" >&2
	exit 1
fi

echo "installer enable service test: ok"
