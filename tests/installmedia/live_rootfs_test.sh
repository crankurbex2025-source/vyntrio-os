#!/usr/bin/env bash
# Verifies the minimal live-rootfs userland (Slice 9.15).
# Asserts the userland actually executes, the record matches reality, the shipped
# live_root carries no state, and prior slices are not regressed.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
LR="${ENVELOPE_ROOT}/live_root"
REC="${BUILD_ROOT}/LIVE_ROOTFS.txt"
WRAP="${BUILD_ROOT}/WRAPPER.txt"

fail() { echo "installmedia live-rootfs test: $*" >&2; exit 1; }
rget() { sed -n "s/^  *$1: //p" "${REC}" | head -1; }

[[ -f "${REC}" ]] || fail "LIVE_ROOTFS.txt missing; run 'make install-media-live-rootfs' first"

SHELL_OK="$(rget shell_busybox)"
API_OK="$(rget vyntrio_api)"
SHELL_EXEC="$(rget shell_exec_in_chroot)"
API_EXEC="$(rget api_exec_in_chroot)"
PROBE="$(rget dashboard_probe)"

# --- userland present ---
[[ "${SHELL_OK}" == "true" ]] || fail "no busybox shell composed; live runtime cannot start"
[[ -x "${LR}/bin/busybox" ]] || fail "busybox missing/not executable in live_root"
[[ -L "${LR}/bin/sh" || -x "${LR}/bin/sh" ]] || fail "no /bin/sh applet in live_root"

# --- loader + library closure present ---
[[ -f "${LR}/lib64/ld-linux-x86-64.so.2" ]] || fail "dynamic loader missing from live_root"
[[ -f "${LR}/lib/x86_64-linux-gnu/libc.so.6" ]] || fail "libc missing from live_root"
# Loader must be a real file, not a dangling symlink (a classic composition bug).
[[ ! -L "${LR}/lib64/ld-linux-x86-64.so.2" ]] || fail "loader is a symlink; must be dereferenced into live_root"

# --- /init present and reaches the first-boot path ---
[[ -x "${LR}/init" ]] || fail "/init missing or not executable"
grep -q 'firstboot.sh' "${LR}/init" || fail "/init must hand off to the first-boot path"

# --- record must match reality: chroot exec is re-run here, not trusted ---
if command -v chroot >/dev/null 2>&1; then
	if chroot "${LR}" /bin/busybox sh -c 'exit 0' >/dev/null 2>&1; then
		[[ "${SHELL_EXEC}" == "true" ]] || fail "shell executes but record says shell_exec_in_chroot=${SHELL_EXEC}"
	else
		[[ "${SHELL_EXEC}" == "false" ]] || fail "record claims shell_exec_in_chroot=true but chroot exec fails"
	fi
	if [[ "${API_OK}" == "true" && "${API_EXEC}" == "true" ]]; then
		# Capture, don't pipe: api exits non-zero on a bad flag.
		out="$(chroot "${LR}" /usr/bin/vyntrio-api --vyntrio-probe-invalid 2>&1 || true)"
		grep -q 'unknown flag' <<<"${out}" || fail "record claims api_exec_in_chroot=true but api does not execute"
	fi
fi

# --- dashboard probe honesty ---
case "${PROBE}" in
served_http_200)
	[[ "$(rget dashboard_http_status)" == "200" ]] || fail "probe says served_http_200 but status is not 200"
	[[ "${API_EXEC}" == "true" ]] || fail "dashboard served but api_exec_in_chroot is not true"
	# Slice 10.2: when the UI answered, a readiness (DB-ready) probe must be recorded.
	READY_PROBE="$(rget dashboard_ready_probe)"
	[[ -n "${READY_PROBE}" ]] || fail "served_http_200 must also record a dashboard_ready_probe (Slice 10.2)"
	READYZ="$(rget dashboard_readyz_status)"
	if [[ "${READY_PROBE}" == "ready_http_200" ]]; then
		[[ "${READYZ}" == "200" ]] || fail "dashboard_ready_probe=ready_http_200 but readyz status is not 200"
	else
		[[ "${READY_PROBE}" == "not_ready" ]] || fail "unknown dashboard_ready_probe: ${READY_PROBE}"
	fi
	;;
skipped|failed)
	grep -q '^  probe_reason: ' "${REC}" || fail "probe ${PROBE} must record a concrete probe_reason"
	;;
*) fail "unknown dashboard_probe: ${PROBE}" ;;
esac

# --- Slice 10.4: TLS HTTPS probe ---
TLS_PROBE="$(rget dashboard_tls_probe)"
case "${TLS_PROBE}" in
served_https_200)
	[[ "$(rget dashboard_tls_http_status)" == "200" ]] || fail "TLS probe says served_https_200 but status is not 200"
	TLS_READY="$(rget dashboard_tls_ready_probe)"
	[[ -n "${TLS_READY}" ]] || fail "served_https_200 must record dashboard_tls_ready_probe"
	;;
skipped|failed)
	grep -q '^  tls_probe_reason: ' "${REC}" || fail "TLS probe ${TLS_PROBE} must record tls_probe_reason"
	;;
*) fail "unknown dashboard_tls_probe: ${TLS_PROBE}" ;;
esac
grep -q '^  tls_readiness: ready' "${REC}" || fail "LIVE_ROOTFS.txt must record tls_readiness: ready"
[[ -x "${LR}/usr/lib/vyntrio/prepare-live-dashboard-tls.sh" ]] \
	|| fail "prepare-live-dashboard-tls.sh missing from live_root"

# --- shipped media must stay state-free (probe ran on a throwaway copy) ---
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			fail "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${LR}" -type f -print0)
[[ ! -e "${LR}/var/lib/vyntrio/vyntrio.db" ]] || fail "probe leaked a database into shipped live_root"

# --- initramfs artifact ---
if [[ "$(rget emitted)" == "true" ]]; then
	IMG="${BUILD_ROOT}/vyntrio-live-initramfs.cpio.gz"
	[[ -s "${IMG}" ]] || fail "initramfs recorded as emitted but missing/empty"
	gzip -t "${IMG}" 2>/dev/null || fail "initramfs is not valid gzip"
	# cpio normalizes away the leading "./": the kernel needs /init at archive root.
	gzip -dc "${IMG}" 2>/dev/null | cpio -t 2>/dev/null | grep -qE '^\.?/?init$' \
		|| fail "initramfs does not contain /init at its root"
	gzip -dc "${IMG}" 2>/dev/null | cpio -t 2>/dev/null | grep -q 'bin/busybox' \
		|| fail "initramfs does not contain busybox"
fi

# --- honesty: no unearned boot/dashboard claims ---
grep -q '^  wired_into_image_initrd: false' "${REC}" || fail "initramfs must not claim to be wired into the image yet"
grep -q '^boot_verified: false' "${REC}" || fail "boot_verified must stay false without a VM"
grep -q '^dashboard_reachable_on_boot: false' "${REC}" || fail "dashboard_reachable_on_boot must stay false"
grep -q '^boot_verified_reason: ' "${REC}" || fail "boot_verified: false must record a reason"

# --- no regression: Slice 9.14 first-boot wiring intact ---
[[ -x "${LR}/usr/lib/vyntrio/firstboot.sh" ]] || fail "9.14 firstboot.sh regressed"
[[ -f "${LR}/etc/systemd/system/vyntrio-firstboot.service" ]] || fail "9.14 first-boot service regressed"
[[ -x "${LR}/usr/lib/vyntrio/live-init.sh" ]] || fail "9.14 live-init.sh regressed"

# --- no regression: firmware image still bootable ---
if [[ -f "${WRAP}" ]]; then
	[[ "$(sed -n 's/^firmware_bootable: //p' "${WRAP}")" == "true" ]] || fail "firmware bootability regressed"
	grep -q '^structural_verification: pass' "${WRAP}" || fail "firmware structural verification regressed"
fi

# --- no regression: payload allowlist ---
mapfile -t pf < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
[[ "${#pf[@]}" -eq 6 ]] || fail "payload file count regressed (${#pf[@]} != 6)"

# --- fail-closed ---
tmp="$(mktemp -d)"; trap 'rm -rf "${tmp}"' EXIT
if VYNTRIO_INSTALL_ENVELOPE_ROOT="${tmp}/missing" \
	VYNTRIO_INSTALL_MEDIA_BUILD_ROOT="${tmp}/build" \
	bash "${ROOT}/scripts/compose-live-rootfs.sh" >/dev/null 2>&1; then
	fail "expected failure when live_root missing"
fi

echo "installmedia live-rootfs test: ok (shell_exec=${SHELL_EXEC}, api_exec=${API_EXEC}, dashboard_probe=${PROBE})"
