#!/usr/bin/env bash
# Read-only installer preflight per docs/ADR/0007-appliance-installer-contract.md (Slice 10.3).
# Validates install-media context and payload inventory only — no target-disk mutation.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVELOPE_ROOT="${VYNTRIO_INSTALL_ENVELOPE_ROOT:-${ROOT}/distro/install-media/envelope}"
PAYLOAD_ROOT="${VYNTRIO_INSTALL_PAYLOAD_ROOT:-${ENVELOPE_ROOT}/payload}"

readonly -a REQUIRED_PAYLOAD=(
	"usr/bin/vyntrio-api"
	"usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service"
	"usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf"
	"etc/vyntrio/config.toml"
)

readonly -a EXCLUDED_GLOBS=(
	"*.db"
	"*.sqlite"
	"*.sqlite3"
	"*.tar"
	"*.tar.gz"
	"*.tgz"
)

fail() {
	echo "installer-preflight: FAIL: $*" >&2
	exit 1
}

ok() {
	echo "installer-preflight: ok: $*"
}

defer() {
	echo "installer-preflight: deferred: $*"
}

# ADR-0007 §D — media role: install media context, not recovery.
if [[ ! -d "${ENVELOPE_ROOT}" ]]; then
	fail "install envelope missing: ${ENVELOPE_ROOT} (run 'make install-media-envelope' first)"
fi

envelope_record="${ENVELOPE_ROOT}/ENVELOPE.txt"
if [[ ! -f "${envelope_record}" ]]; then
	fail "ENVELOPE.txt missing in install envelope"
fi

if ! grep -q '^media_role: install$' "${envelope_record}"; then
	fail "ENVELOPE.txt media_role is not install"
fi
ok "media_role is install"

if [[ -f "${ROOT}/distro/recovery-media/manifest.yaml" ]]; then
	if grep -q 'media_role: recovery' "${ENVELOPE_ROOT}/ENVELOPE.txt" 2>/dev/null; then
		fail "recovery media context detected; installer preflight requires install media"
	fi
fi
ok "recovery separation (envelope is install context)"

# ADR-0007 §D — payload presence and inventory.
if [[ ! -d "${PAYLOAD_ROOT}" ]]; then
	fail "install payload missing: ${PAYLOAD_ROOT}"
fi

for rel in "${REQUIRED_PAYLOAD[@]}"; do
	if [[ ! -f "${PAYLOAD_ROOT}/${rel}" ]]; then
		fail "missing manifest payload file: ${rel}"
	fi
done
ok "payload presence (${#REQUIRED_PAYLOAD[@]} manifest files)"

mapfile -t payload_files < <(find "${PAYLOAD_ROOT}" -type f | LC_ALL=C sort)
if [[ "${#payload_files[@]}" -ne "${#REQUIRED_PAYLOAD[@]}" ]]; then
	fail "unexpected payload file count (expected ${#REQUIRED_PAYLOAD[@]}, found ${#payload_files[@]})"
fi

for path in "${payload_files[@]}"; do
	rel="${path#${PAYLOAD_ROOT}/}"
	found=false
	for allowed in "${REQUIRED_PAYLOAD[@]}"; do
		if [[ "${rel}" == "${allowed}" ]]; then
			found=true
			break
		fi
	done
	if [[ "${found}" != true ]]; then
		fail "non-manifest payload file: ${rel}"
	fi
done
ok "payload inventory matches manifest"

# ADR-0007 §D — media exclusions (forbidden artifacts on generic install media).
for pattern in "${EXCLUDED_GLOBS[@]}"; do
	matches="$(find "${PAYLOAD_ROOT}" -type f -name "${pattern}" -print || true)"
	if [[ -n "${matches}" ]]; then
		fail "excluded payload pattern on media: ${pattern}"
	fi
done

while IFS= read -r -d '' path; do
	base="$(basename "${path}")"
	lower="${base,,}"
	if [[ "${lower}" == *license* ]] || [[ "${lower}" == *bootstrap*token* ]]; then
		fail "excluded payload basename on media: ${base}"
	fi
	if [[ "${base}" == "vyntrio-restore" ]]; then
		fail "recovery tooling in install payload: ${base}"
	fi
done < <(find "${PAYLOAD_ROOT}" -type f -print0)

ok "media exclusions (no forbidden artifacts in payload)"

# ADR-0007 §D — bootstrap boundary: read-only; no bootstrap invocation in preflight.
ok "bootstrap boundary (preflight does not invoke bootstrap endpoints)"

# ADR-0007 §D — target disk: deferred in Slice 10.3 scaffold.
defer "target disk accessibility (not implemented in Slice 10.3)"
defer "operator reprovision confirmation (not implemented in Slice 10.3)"

echo "installer-preflight: ready — media context valid; install execution not authorized"
