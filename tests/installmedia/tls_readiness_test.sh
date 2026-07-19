#!/usr/bin/env bash
# Verifies TLS readiness for safe LAN dashboard bind (Stage 2 / Slice 10.4).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
LR="${ENVELOPE_ROOT}/live_root"
REC="${BUILD_ROOT}/LIVE_ROOTFS.txt"
RUNTIME_REC="${BUILD_ROOT}/RUNTIME_BOOT.txt"

fail() { echo "installmedia tls-readiness test: $*" >&2; exit 1; }
rget() { sed -n "s/^  *$1: //p" "${REC}" | head -1; }

[[ -f "${REC}" ]] || fail "LIVE_ROOTFS.txt missing; run 'make install-media-live-rootfs' first"
[[ -f "${RUNTIME_REC}" ]] || fail "RUNTIME_BOOT.txt missing; run 'make install-media-runtime' first"

PREP="${LR}/usr/lib/vyntrio/prepare-live-dashboard-tls.sh"
[[ -x "${PREP}" ]] || fail "prepare-live-dashboard-tls.sh missing"

# --- API TLS probe honesty (throwaway chroot copy) ---
TLS_PROBE="$(rget dashboard_tls_probe)"
case "${TLS_PROBE}" in
served_https_200)
	[[ "$(rget dashboard_tls_http_status)" == "200" ]] || fail "TLS probe says served_https_200 but status is not 200"
	TLS_READY="$(rget dashboard_tls_ready_probe)"
	[[ -n "${TLS_READY}" ]] || fail "served_https_200 must record dashboard_tls_ready_probe"
	if [[ "${TLS_READY}" == "ready_https_200" ]]; then
		[[ "$(rget dashboard_tls_readyz_status)" == "200" ]] \
			|| fail "dashboard_tls_ready_probe=ready_https_200 but readyz status is not 200"
	else
		[[ "${TLS_READY}" == "not_ready" ]] || fail "unknown dashboard_tls_ready_probe: ${TLS_READY}"
	fi
	;;
skipped|failed)
	grep -q '^  tls_probe_reason: ' "${REC}" || fail "TLS probe ${TLS_PROBE} must record tls_probe_reason"
	;;
*) fail "unknown dashboard_tls_probe: ${TLS_PROBE}" ;;
esac

grep -q '^  tls_readiness: ready' "${REC}" || fail "LIVE_ROOTFS.txt must record tls_readiness: ready"
grep -q '^  tls_readiness: ready' "${RUNTIME_REC}" || fail "RUNTIME_BOOT.txt must record tls_readiness: ready"

# --- prepare script: loopback-only path (no LAN IP) ---
tmp="$(mktemp -d)"; trap 'rm -rf "${tmp}"' EXIT
mkdir -p "${tmp}/etc/vyntrio"
cat >"${tmp}/etc/vyntrio/config.toml" <<'EOF'
bind_address = "127.0.0.1"
listen_port = 8080
state_dir = "/var/lib/vyntrio"
log_level = "info"
cookie_secure = false
EOF
VYNTRIO_TLS_STATUS="${tmp}/tls-readiness.txt" \
	VYNTRIO_RUNTIME_CONFIG="${tmp}/dashboard-config.toml" \
	sh "${PREP}" "${tmp}/etc/vyntrio/config.toml"
grep -q '^tls_readiness: ready_loopback_only' "${tmp}/tls-readiness.txt" \
	|| fail "prepare script must record ready_loopback_only without VYNTRIO_LAN_BIND_IP"
[[ ! -f "${tmp}/dashboard-config.toml" ]] || fail "LAN overlay must not be written without VYNTRIO_LAN_BIND_IP"

# --- prepare script: LAN HTTPS overlay when IP + openssl are set ---
if command -v openssl >/dev/null 2>&1; then
	VYNTRIO_LAN_BIND_IP="192.168.254.99" \
		VYNTRIO_TLS_DIR="${tmp}/tls" \
		VYNTRIO_TLS_STATUS="${tmp}/tls-readiness-lan.txt" \
		VYNTRIO_RUNTIME_CONFIG="${tmp}/dashboard-config-lan.toml" \
		sh "${PREP}" "${tmp}/etc/vyntrio/config.toml"
	grep -q '^tls_readiness: ready' "${tmp}/tls-readiness-lan.txt" \
		|| fail "prepare script must record tls_readiness: ready with VYNTRIO_LAN_BIND_IP"
	[[ -f "${tmp}/dashboard-config-lan.toml" ]] || fail "LAN runtime config overlay missing"
	grep -q 'tls_cert_file' "${tmp}/dashboard-config-lan.toml" || fail "LAN overlay must include tls_cert_file"
	grep -q 'cookie_secure = true' "${tmp}/dashboard-config-lan.toml" || fail "LAN overlay must set cookie_secure=true"
fi

# --- Go unit tests for config TLS validation ---
if command -v go >/dev/null 2>&1; then
	( cd "${ROOT}" && go test ./internal/platform/config/... ./internal/platform/tlsutil/... ./internal/interfaces/http/... -count=1 ) \
		|| fail "Go TLS/config tests failed"
fi

echo "installmedia tls-readiness test: ok (dashboard_tls_probe=${TLS_PROBE})"
