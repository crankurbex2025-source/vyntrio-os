#!/usr/bin/env bash
# Read-only validation of installer target-layout manifest (Slice 10.4).
# Does not access block devices or mutate disks.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MANIFEST="${VYNTRIO_INSTALLER_LAYOUT_MANIFEST:-${ROOT}/distro/installer/target-layout-manifest.yaml}"

readonly -a REQUIRED_TARGETS=(
	"/usr/bin/vyntrio-api"
	"/usr/bin/vyntrio-backup"
	"/etc/systemd/system/vyntrio-api.service"
	"/usr/lib/sysusers.d/vyntrio.conf"
	"/etc/tmpfiles.d/vyntrio.conf"
	"/etc/vyntrio/config.toml"
)

readonly -a REQUIRED_DIRECTORIES=(
	"/etc/vyntrio"
	"/var/lib/vyntrio"
	"/var/lib/vyntrio/backups"
)

if [[ ! -f "${MANIFEST}" ]]; then
	echo "validate-installer-layout-plan: manifest missing: ${MANIFEST}" >&2
	exit 1
fi

if ! grep -q '^schema_version: vyntrio-installer-target-layout-v1$' "${MANIFEST}"; then
	echo "validate-installer-layout-plan: invalid or missing schema_version" >&2
	exit 1
fi

if ! grep -q 'partition_layout:' "${MANIFEST}"; then
	echo "validate-installer-layout-plan: partition_layout section missing" >&2
	exit 1
fi

if ! grep -q 'status: deferred' "${MANIFEST}"; then
	echo "validate-installer-layout-plan: expected deferred partition_layout status" >&2
	exit 1
fi

for target in "${REQUIRED_TARGETS[@]}"; do
	if ! grep -q "target: ${target}" "${MANIFEST}"; then
		echo "validate-installer-layout-plan: missing payload target: ${target}" >&2
		exit 1
	fi
done

for dir in "${REQUIRED_DIRECTORIES[@]}"; do
	if ! grep -q "path: ${dir}" "${MANIFEST}"; then
		echo "validate-installer-layout-plan: missing directory entry: ${dir}" >&2
		exit 1
	fi
done

if ! grep -q 'initial_content: empty' "${MANIFEST}"; then
	echo "validate-installer-layout-plan: expected empty initial_content for state directories" >&2
	exit 1
fi

echo "validate-installer-layout-plan: ok (${#REQUIRED_TARGETS[@]} payload targets, ${#REQUIRED_DIRECTORIES[@]} directories, partition deferred)"
