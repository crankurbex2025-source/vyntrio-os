#!/usr/bin/env bash
# Compose a minimal live-rootfs userland (Block 9 / Slice 9.15).
#
# Goal: the smallest userland that can actually START the first-boot path —
# a shell (busybox), the dynamic loader + library closure, and vyntrio-api (the
# dashboard) when present, plus an /init that reaches firstboot.sh.
#
# Verification without a VM:
#   - chroot exec: busybox sh and vyntrio-api must actually run under the
#     composed loader/libc.
#   - optional chroot HTTP probe (throwaway copy, free high port): proves the
#     dashboard really serves from this userland.
#
# Honesty:
#   - The shipped live_root never contains a database, credentials, or secrets.
#     The HTTP probe runs against a THROWAWAY copy so no state lands on media.
#   - Composing a userland is not a boot. Runtime boot remains unverified here
#     (no qemu). The image initrd is NOT swapped in this slice — see LIVE_ROOTFS.txt.
#   - vyntrio-api in live_root is the LIVE dashboard runtime; it is distinct from
#     payload/ (the target-disk copy). No target-disk writes, no installer/apply.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVELOPE_ROOT="${VYNTRIO_INSTALL_ENVELOPE_ROOT:-${ROOT}/distro/install-media/envelope}"
BUILD_ROOT="${VYNTRIO_INSTALL_MEDIA_BUILD_ROOT:-${ROOT}/distro/install-media/build}"
LR="${ENVELOPE_ROOT}/live_root"
INITRAMFS_NAME="vyntrio-live-initramfs.cpio.gz"
PROBE_PORT="${VYNTRIO_LIVE_PROBE_PORT:-18080}"

BUSYBOX_SRC="${VYNTRIO_BUSYBOX:-$(command -v busybox || true)}"
API_SRC="${VYNTRIO_API_BIN:-}"

die() { echo "compose-live-rootfs: $*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

[[ -d "${LR}" ]] || die "live_root missing; run 'make install-media-runtime' first"
mkdir -p "${BUILD_ROOT}"

# Prefer the envelope payload's api (manifest-built), else repo bin/.
if [[ -z "${API_SRC}" ]]; then
	if [[ -x "${ENVELOPE_ROOT}/payload/usr/bin/vyntrio-api" ]]; then
		API_SRC="${ENVELOPE_ROOT}/payload/usr/bin/vyntrio-api"
	elif [[ -x "${ROOT}/bin/vyntrio-api" ]]; then
		API_SRC="${ROOT}/bin/vyntrio-api"
	fi
fi

# --- skeleton ---
mkdir -p "${LR}"/{bin,sbin,usr/bin,usr/lib/vyntrio,etc/vyntrio,proc,sys,dev,run,tmp,var/lib,lib64,lib/x86_64-linux-gnu}

# --- library closure helper: copy an ELF's interpreter + NEEDED libs, dereferenced ---
copy_closure() {
	local bin="$1" lib interp
	# Dynamic loader (ELF interpreter).
	interp="$(readelf -l "${bin}" 2>/dev/null | sed -n 's/.*program interpreter: \(.*\)\]/\1/p' | tr -d ' ')"
	if [[ -n "${interp}" && -e "${interp}" ]]; then
		mkdir -p "${LR}$(dirname "${interp}")"
		cp -aL "$(readlink -f "${interp}")" "${LR}${interp}"
	fi
	# Shared libraries resolved by ldd.
	while read -r lib; do
		[[ -n "${lib}" && -e "${lib}" ]] || continue
		mkdir -p "${LR}$(dirname "${lib}")"
		cp -aL "$(readlink -f "${lib}")" "${LR}${lib}" 2>/dev/null || true
	done < <(ldd "${bin}" 2>/dev/null | sed -n 's/.*=> \(\/[^ ]*\).*/\1/p')
}

USERLAND_SHELL=false
USERLAND_API=false

# --- busybox (shell + core utilities) ---
if [[ -n "${BUSYBOX_SRC}" && -x "${BUSYBOX_SRC}" ]]; then
	cp -aL "${BUSYBOX_SRC}" "${LR}/bin/busybox"
	chmod 0755 "${LR}/bin/busybox"
	copy_closure "${BUSYBOX_SRC}"
	# Minimal applet set for a first-boot environment.
	for a in sh ash mount umount ip ifconfig ls cat echo mkdir sleep ps kill uname dmesg switch_root; do
		ln -sf /bin/busybox "${LR}/bin/${a}"
	done
	ln -sf /bin/busybox "${LR}/sbin/init.busybox"
	USERLAND_SHELL=true
fi

# --- vyntrio-api: the LIVE dashboard runtime (distinct from payload/) ---
if [[ -n "${API_SRC}" && -x "${API_SRC}" ]]; then
	cp -aL "${API_SRC}" "${LR}/usr/bin/vyntrio-api"
	chmod 0755 "${LR}/usr/bin/vyntrio-api"
	copy_closure "${API_SRC}"
	USERLAND_API=true
fi

# --- openssl (optional): runtime TLS cert generation for LAN dashboard bind (Slice 10.4) ---
OPENSSL_SRC="${VYNTRIO_OPENSSL:-$(command -v openssl || true)}"
USERLAND_OPENSSL=false
if [[ -n "${OPENSSL_SRC}" && -x "${OPENSSL_SRC}" ]]; then
	cp -aL "${OPENSSL_SRC}" "${LR}/usr/bin/openssl"
	chmod 0755 "${LR}/usr/bin/openssl"
	copy_closure "${OPENSSL_SRC}"
	USERLAND_OPENSSL=true
fi

# Slice 10.4 TLS prep script must be present (installed by verify-runtime-boot.sh).
[[ -x "${LR}/usr/lib/vyntrio/prepare-live-dashboard-tls.sh" ]] \
	|| die "prepare-live-dashboard-tls.sh missing; run 'make install-media-runtime' first"

# Live dashboard config (template only — no secrets, no state).
if [[ -f "${ENVELOPE_ROOT}/payload/etc/vyntrio/config.toml" ]]; then
	cp -aL "${ENVELOPE_ROOT}/payload/etc/vyntrio/config.toml" "${LR}/etc/vyntrio/config.toml"
fi

# --- /init: first process of the live runtime ---
cat >"${LR}/init" <<'EOF'
#!/bin/sh
# Vyntrio live-rootfs init (Slice 9.15). Runs as PID 1 in the live runtime.
# Brings up the minimal environment, then hands off to the first-boot path.
/bin/busybox mount -t proc proc /proc 2>/dev/null
/bin/busybox mount -t sysfs sysfs /sys 2>/dev/null
/bin/busybox mount -t devtmpfs devtmpfs /dev 2>/dev/null
/bin/busybox mount -t tmpfs tmpfs /run 2>/dev/null
/bin/busybox ip link set lo up 2>/dev/null || /bin/busybox ifconfig lo up 2>/dev/null

echo "vyntrio live-init: minimal userland up" >&2

if [ -x /usr/lib/vyntrio/firstboot.sh ]; then
	/usr/lib/vyntrio/firstboot.sh
fi

echo "vyntrio live-init: first-boot path returned; dropping to shell" >&2
exec /bin/busybox sh
EOF
chmod 0755 "${LR}/init"

# --- fail closed: no state/secrets on shipped media ---
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			die "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${LR}" -type f -print0)

# --- verification 1: chroot exec (proves loader/libc closure is complete) ---
SHELL_EXEC=false
API_EXEC=false
if [[ "${USERLAND_SHELL}" == true ]] && have chroot; then
	if chroot "${LR}" /bin/busybox sh -c 'exit 0' >/dev/null 2>&1; then
		SHELL_EXEC=true
	fi
fi
if [[ "${USERLAND_API}" == true ]] && have chroot; then
	# A rejected bogus flag proves the binary executed under the composed userland.
	# Capture output rather than pipe: the api exits non-zero on a bad flag and
	# `set -o pipefail` would mask a successful match.
	api_out="$(chroot "${LR}" /usr/bin/vyntrio-api --vyntrio-probe-invalid 2>&1 || true)"
	if grep -q 'unknown flag' <<<"${api_out}"; then
		API_EXEC=true
	fi
fi

# --- verification 2: dashboard HTTP probe from a THROWAWAY copy ---
# Never probe the shipped live_root: vyntrio-api creates a database at runtime.
DASHBOARD_PROBE="skipped"
DASHBOARD_STATUS=""
DASHBOARD_READY_STATUS=""
PROBE_REASON=""
if [[ "${API_EXEC}" != true ]]; then
	PROBE_REASON="vyntrio-api not executable in composed userland"
elif ! have curl || ! have chroot; then
	PROBE_REASON="curl/chroot unavailable"
elif command -v ss >/dev/null 2>&1 && ss -ltn 2>/dev/null | grep -q ":${PROBE_PORT} "; then
	PROBE_REASON="probe port ${PROBE_PORT} already in use (refusing to disturb a running service)"
else
	probe_dir="$(mktemp -d)"
	cp -a "${LR}/." "${probe_dir}/" 2>/dev/null || true
	mkdir -p "${probe_dir}/var/lib/vyntrio" "${probe_dir}/dev"
	mknod "${probe_dir}/dev/urandom" c 1 9 2>/dev/null || true
	mknod "${probe_dir}/dev/null" c 1 3 2>/dev/null || true
	sed -e "s/^listen_port = .*/listen_port = ${PROBE_PORT}/" \
		"${LR}/etc/vyntrio/config.toml" >"${probe_dir}/etc/vyntrio/config.toml" 2>/dev/null || true

	chroot "${probe_dir}" /usr/bin/vyntrio-api --config /etc/vyntrio/config.toml >/dev/null 2>&1 &
	probe_pid=$!
	for _ in $(seq 1 30); do
		sleep 0.3
		DASHBOARD_STATUS="$(curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:${PROBE_PORT}/" 2>/dev/null || true)"
		[[ "${DASHBOARD_STATUS}" == "200" ]] && break
	done
	# Readiness (Slice 10.2): /readyz returns 200 only when the backend (DB) is
	# actually ready — stronger evidence of a stable dashboard than serving the UI.
	if [[ "${DASHBOARD_STATUS}" == "200" ]]; then
		DASHBOARD_READY_STATUS="$(curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:${PROBE_PORT}/readyz" 2>/dev/null || true)"
	fi
	kill "${probe_pid}" >/dev/null 2>&1 || true
	wait "${probe_pid}" 2>/dev/null || true
	rm -rf "${probe_dir}"

	if [[ "${DASHBOARD_STATUS}" == "200" ]]; then
		DASHBOARD_PROBE="served_http_200"
	else
		DASHBOARD_PROBE="failed"
		PROBE_REASON="dashboard did not answer 200 (got '${DASHBOARD_STATUS:-no response}')"
	fi
fi

# --- verification 3: dashboard HTTPS/TLS probe from a THROWAWAY copy (Slice 10.4) ---
DASHBOARD_TLS_PROBE="skipped"
DASHBOARD_TLS_STATUS=""
DASHBOARD_TLS_READY_STATUS=""
TLS_PROBE_REASON=""
if [[ "${API_EXEC}" != true ]]; then
	TLS_PROBE_REASON="vyntrio-api not executable in composed userland"
elif ! have openssl || ! have curl || ! have chroot; then
	TLS_PROBE_REASON="openssl/curl/chroot unavailable"
elif command -v ss >/dev/null 2>&1 && ss -ltn 2>/dev/null | grep -q ":${PROBE_PORT} "; then
	TLS_PROBE_REASON="probe port ${PROBE_PORT} already in use"
else
	tls_probe_dir="$(mktemp -d)"
	cp -a "${LR}/." "${tls_probe_dir}/" 2>/dev/null || true
	mkdir -p "${tls_probe_dir}/var/lib/vyntrio" "${tls_probe_dir}/dev" "${tls_probe_dir}/run/vyntrio/tls"
	mknod "${tls_probe_dir}/dev/urandom" c 1 9 2>/dev/null || true
	mknod "${tls_probe_dir}/dev/null" c 1 3 2>/dev/null || true
	tls_cert="${tls_probe_dir}/run/vyntrio/tls/dashboard-cert.pem"
	tls_key="${tls_probe_dir}/run/vyntrio/tls/dashboard-key.pem"
	if openssl req -x509 -newkey rsa:2048 -nodes \
		-keyout "${tls_key}" -out "${tls_cert}" -days 1 \
		-subj "/CN=vyntrio-probe" \
		-addext "subjectAltName=IP:127.0.0.1" \
		2>/dev/null; then
		cat >"${tls_probe_dir}/etc/vyntrio/config.toml" <<TLSCFG
bind_address = "127.0.0.1"
listen_port = ${PROBE_PORT}
state_dir = "/var/lib/vyntrio"
log_level = "info"
cookie_secure = true
tls_cert_file = "/run/vyntrio/tls/dashboard-cert.pem"
tls_key_file = "/run/vyntrio/tls/dashboard-key.pem"
TLSCFG
		chroot "${tls_probe_dir}" /usr/bin/vyntrio-api --config /etc/vyntrio/config.toml >/dev/null 2>&1 &
		tls_probe_pid=$!
		for _ in $(seq 1 30); do
			sleep 0.3
			DASHBOARD_TLS_STATUS="$(curl -sk -o /dev/null -w '%{http_code}' "https://127.0.0.1:${PROBE_PORT}/" 2>/dev/null || true)"
			[[ "${DASHBOARD_TLS_STATUS}" == "200" ]] && break
		done
		if [[ "${DASHBOARD_TLS_STATUS}" == "200" ]]; then
			DASHBOARD_TLS_READY_STATUS="$(curl -sk -o /dev/null -w '%{http_code}' "https://127.0.0.1:${PROBE_PORT}/readyz" 2>/dev/null || true)"
		fi
		kill "${tls_probe_pid}" >/dev/null 2>&1 || true
		wait "${tls_probe_pid}" 2>/dev/null || true
		if [[ "${DASHBOARD_TLS_STATUS}" == "200" ]]; then
			DASHBOARD_TLS_PROBE="served_https_200"
		else
			DASHBOARD_TLS_PROBE="failed"
			TLS_PROBE_REASON="dashboard did not answer HTTPS 200 (got '${DASHBOARD_TLS_STATUS:-no response}')"
		fi
	else
		DASHBOARD_TLS_PROBE="failed"
		TLS_PROBE_REASON="openssl cert generation failed for TLS probe"
	fi
	rm -rf "${tls_probe_dir}"
fi

# --- emit live initramfs artifact (real cpio.gz of the userland) ---
INITRAMFS_EMITTED=false
if [[ "${USERLAND_SHELL}" == true ]] && have cpio && have gzip; then
	( cd "${LR}" && find . -print0 | cpio --null -o --format=newc 2>/dev/null | gzip -9 ) \
		>"${BUILD_ROOT}/${INITRAMFS_NAME}" 2>/dev/null
	[[ -s "${BUILD_ROOT}/${INITRAMFS_NAME}" ]] && INITRAMFS_EMITTED=true
fi

# --- provenance ---
{
	echo "# Generated by scripts/compose-live-rootfs.sh — do not commit"
	echo "schema_version: vyntrio-install-media-live-rootfs-v1"
	echo "slice: 9.15"
	echo "generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
	echo "media_role: install"
	echo "userland:"
	echo "  shell_busybox: ${USERLAND_SHELL}"
	echo "  vyntrio_api: ${USERLAND_API}"
	echo "  openssl: ${USERLAND_OPENSSL}"
	echo "  tls_runtime_prep: usr/lib/vyntrio/prepare-live-dashboard-tls.sh"
	echo "  init: /init (PID 1 -> firstboot.sh)"
	echo "verified:"
	echo "  shell_exec_in_chroot: ${SHELL_EXEC}"
	echo "  api_exec_in_chroot: ${API_EXEC}"
	echo "  dashboard_probe: ${DASHBOARD_PROBE}"
	[[ -n "${DASHBOARD_STATUS}" ]] && echo "  dashboard_http_status: ${DASHBOARD_STATUS}"
	# Slice 10.2: readiness (DB-ready) probe result, when the UI probe succeeded.
	if [[ -n "${DASHBOARD_READY_STATUS}" ]]; then
		if [[ "${DASHBOARD_READY_STATUS}" == "200" ]]; then
			echo "  dashboard_ready_probe: ready_http_200"
		else
			echo "  dashboard_ready_probe: not_ready"
		fi
		echo "  dashboard_readyz_status: ${DASHBOARD_READY_STATUS}"
	fi
	[[ -n "${PROBE_REASON}" ]] && echo "  probe_reason: ${PROBE_REASON}"
	echo "  dashboard_tls_probe: ${DASHBOARD_TLS_PROBE}"
	[[ -n "${DASHBOARD_TLS_STATUS}" ]] && echo "  dashboard_tls_http_status: ${DASHBOARD_TLS_STATUS}"
	if [[ -n "${DASHBOARD_TLS_READY_STATUS}" ]]; then
		if [[ "${DASHBOARD_TLS_READY_STATUS}" == "200" ]]; then
			echo "  dashboard_tls_ready_probe: ready_https_200"
		else
			echo "  dashboard_tls_ready_probe: not_ready"
		fi
		echo "  dashboard_tls_readyz_status: ${DASHBOARD_TLS_READY_STATUS}"
	fi
	[[ -n "${TLS_PROBE_REASON}" ]] && echo "  tls_probe_reason: ${TLS_PROBE_REASON}"
	echo "  tls_readiness: ready"
	echo "  probe_isolation: throwaway_chroot_copy (shipped live_root stays state-free)"
	echo "initramfs:"
	echo "  emitted: ${INITRAMFS_EMITTED}"
	if [[ "${INITRAMFS_EMITTED}" == true ]]; then
		echo "  path: distro/install-media/build/${INITRAMFS_NAME}"
		echo "  bytes: $(stat -c '%s' "${BUILD_ROOT}/${INITRAMFS_NAME}")"
	fi
	echo "  wired_into_image_initrd: false"
	echo "  wired_reason: >-"
	echo "    Swapping the raw image initrd is deferred: this userland carries no kernel"
	echo "    modules (storage/network drivers) and no VM exists here to verify the boot."
	echo "boot_verified: false"
	echo "boot_verified_reason: no qemu/VM harness on this host (see RUNTIME_BOOT.txt)"
	echo "dashboard_reachable_on_boot: false"
	echo "missing_for_dashboard_reachable_boot:"
	echo "  - kernel_modules_for_storage_and_network_in_initramfs"
	echo "  - image_initrd_swapped_to_live_initramfs"
	echo "  - runtime_boot_verification_vm_or_hardware"
	echo "  - network_bring_up_beyond_loopback"
	echo "missing_for_usb_creator:"
	echo "  - host_usb_writer_tool"
	echo "  - uefi_boot_support"
} >"${BUILD_ROOT}/LIVE_ROOTFS.txt"

echo "install-media live-rootfs: shell=${USERLAND_SHELL} api=${USERLAND_API} shell_exec=${SHELL_EXEC} api_exec=${API_EXEC}"
echo "install-media live-rootfs: dashboard_probe=${DASHBOARD_PROBE}${DASHBOARD_STATUS:+ (HTTP ${DASHBOARD_STATUS})}${DASHBOARD_READY_STATUS:+ readyz=${DASHBOARD_READY_STATUS}} tls_probe=${DASHBOARD_TLS_PROBE}${DASHBOARD_TLS_READY_STATUS:+ tls_readyz=${DASHBOARD_TLS_READY_STATUS}}"
echo "install-media live-rootfs: initramfs_emitted=${INITRAMFS_EMITTED} (not wired into image initrd yet)"
echo "install-media live-rootfs: dashboard_reachable_on_boot=false — see build/LIVE_ROOTFS.txt" >&2
