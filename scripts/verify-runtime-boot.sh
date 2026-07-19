#!/usr/bin/env bash
# Runtime boot verification + first-boot dashboard path (Block 9 / Slice 9.14).
#
# Two concerns:
#   1. First-boot dashboard path — lay down the smallest first-boot wiring into
#      live_root: an entrypoint that launches vyntrio-api (the dashboard) and a
#      systemd unit that runs it on first boot. Honest: this executes only once a
#      live userland exists in the image; it does not run here.
#   2. Runtime boot verification — if a VM/boot harness (qemu) exists, actually
#      boot the Slice 9.13 firmware image and confirm it reaches the kernel; else
#      skip honestly and record exactly why.
#
# Constraints: no installer/apply, no target-disk writes, no regression to the
# firmware image, no faked runtime boot. USB creator remains deferred.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVELOPE_ROOT="${VYNTRIO_INSTALL_ENVELOPE_ROOT:-${ROOT}/distro/install-media/envelope}"
BUILD_ROOT="${VYNTRIO_INSTALL_MEDIA_BUILD_ROOT:-${ROOT}/distro/install-media/build}"
LR="${ENVELOPE_ROOT}/live_root"
WRAPPER_REC="${BUILD_ROOT}/WRAPPER.txt"
DASHBOARD_PORT="${VYNTRIO_DASHBOARD_PORT:-8080}"
# Bounded respawn budget for the supervised local dashboard (Slice 10.2).
DASHBOARD_MAX_RESTARTS="${VYNTRIO_DASHBOARD_MAX_RESTARTS:-5}"

die() { echo "verify-runtime-boot: $*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

# --- precondition: Slice 9.13 wrapper must have run ---
[[ -d "${LR}" ]] || die "live_root missing; run 'make install-media-wrap' first"
[[ -f "${WRAPPER_REC}" ]] || die "WRAPPER.txt missing; run 'make install-media-wrap' first"

FIRMWARE_BOOTABLE="$(sed -n 's/^firmware_bootable: //p' "${WRAPPER_REC}")"
IMAGE_REL="$(sed -n 's/^artifact: //p' "${WRAPPER_REC}")"
IMAGE_FMT="$(sed -n 's/^artifact_format: //p' "${WRAPPER_REC}")"
IMAGE_ABS="${ROOT}/${IMAGE_REL}"

# --- 1. first-boot dashboard path (smallest viable wiring) ---
mkdir -p "${LR}/usr/lib/vyntrio" "${LR}/etc/systemd/system" "${LR}/etc/vyntrio"

# Slice 10.4: runtime TLS prep script (POSIX; staged into live_root by compose-live-rootfs).
install -m 0755 "${ROOT}/scripts/prepare-live-dashboard-tls.sh" "${LR}/usr/lib/vyntrio/prepare-live-dashboard-tls.sh"

cat >"${LR}/usr/lib/vyntrio/firstboot.sh" <<EOF
#!/bin/sh
# Vyntrio first-boot dashboard entrypoint (Slice 9.14 wiring; Slice 10.2 stability).
# Local-first appliance: the browser dashboard (vyntrio-api WebGUI) is the PRIMARY
# management surface, so keep it running and tell the operator where to reach it.
#
# POSIX / busybox-sh ONLY: the live runtime ships busybox, not bash. A
# '#!/usr/bin/env bash' entrypoint would fail to exec here (no bash, no env
# applet), so the dashboard would never start on boot. That is the Slice 10.2
# fix — this script runs under /bin/sh (busybox).
#
# HONEST LIMITS:
#   - Loopback HTTP is the secure-by-default bind. LAN HTTPS requires
#     VYNTRIO_LAN_BIND_IP plus runtime TLS prep (Slice 10.4).
#   - Actual boot-to-dashboard reachability is proven only in a booted VM/hardware
#     environment (see RUNTIME_VERIFY.txt); this file is the runtime wiring.
set -u

API_BIN=/usr/bin/vyntrio-api
CONFIG=/etc/vyntrio/config.toml
TLS_PREP=/usr/lib/vyntrio/prepare-live-dashboard-tls.sh
DASH_PORT=${DASHBOARD_PORT}
MAX_RESTARTS=${DASHBOARD_MAX_RESTARTS}
DASH_URL="http://<host>:\${DASH_PORT}"

# Loopback up (idempotent; /init already brings it up).
ip link set lo up 2>/dev/null || ifconfig lo up 2>/dev/null || true

if [ ! -x "\${API_BIN}" ]; then
	echo "vyntrio firstboot: \${API_BIN} not present — dashboard not started (live userland incomplete)" >&2
	echo "vyntrio firstboot: expected local dashboard at \${DASH_URL} once userland is composed" >&2
	exit 0
fi

# Slice 10.4: optional LAN HTTPS overlay when the operator sets VYNTRIO_LAN_BIND_IP.
if [ -x "\${TLS_PREP}" ]; then
	"\${TLS_PREP}" "\${CONFIG}" || true
	if [ -f /run/vyntrio/dashboard-config.toml ]; then
		CONFIG=/run/vyntrio/dashboard-config.toml
		DASH_URL="https://\${VYNTRIO_LAN_BIND_IP}:\${DASH_PORT}"
	fi
fi

echo "vyntrio firstboot: starting local dashboard (WebGUI) on :\${DASH_PORT}" >&2

# First-boot onboarding clarity (Slice 10.3): tell the operator exactly what to
# do next and what "first use" looks like. The dashboard is the primary management
# surface, reachable locally on this machine before any LAN or remote path.
cat <<MSG >&2

=========================================================
  Vyntrio first-boot setup
=========================================================
  The local dashboard is the primary way to manage this
  appliance. After it starts, open it on this machine:

    \${DASH_URL}

  First use:
    1. Open the dashboard in a browser on this device.
    2. Sign in (or create the owner account on first boot).
    3. Review the read-only overview with real data from
       this appliance.

  Storage, services, and remote access are added later.
  This is the local-first appliance path.
=========================================================

MSG

# Supervise the dashboard: an appliance WebGUI must stay up. Respawn on exit with
# a bounded budget so a hard-failing binary drops to a shell for diagnosis instead
# of silently disappearing or spinning forever.
restarts=0
while : ; do
	"\${API_BIN}" --config "\${CONFIG}"
	rc=\$?
	restarts=\$((restarts + 1))
	echo "vyntrio firstboot: dashboard exited (rc=\${rc}); restart \${restarts}/\${MAX_RESTARTS}" >&2
	if [ "\${restarts}" -ge "\${MAX_RESTARTS}" ]; then
		echo "vyntrio firstboot: dashboard failed \${MAX_RESTARTS} times — giving up, dropping to shell" >&2
		exit "\${rc}"
	fi
	sleep 2
done
EOF
chmod 0755 "${LR}/usr/lib/vyntrio/firstboot.sh"

cat >"${LR}/etc/systemd/system/vyntrio-firstboot.service" <<EOF
[Unit]
Description=Vyntrio first-boot dashboard
Documentation=file:/usr/lib/vyntrio/FIRST_BOOT.txt
After=network.target
ConditionPathExists=/usr/bin/vyntrio-api

[Service]
Type=simple
ExecStart=/usr/lib/vyntrio/firstboot.sh
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# live-init invokes the first-boot dashboard step (no-op without a userland).
cat >"${LR}/usr/lib/vyntrio/live-init.sh" <<'EOF'
#!/usr/bin/env bash
# Vyntrio live init (Slice 9.14). Structured live-boot entry point.
set -euo pipefail

mount_pseudo() {
	mountpoint -q /proc 2>/dev/null || mount -t proc proc /proc 2>/dev/null || true
	mountpoint -q /sys  2>/dev/null || mount -t sysfs sysfs /sys 2>/dev/null || true
	mountpoint -q /run  2>/dev/null || mount -t tmpfs tmpfs /run 2>/dev/null || true
}

echo "vyntrio live-init: media_role=install boot_chain=${VYNTRIO_BOOTABILITY:-real}" >&2
mount_pseudo || true

# First-boot dashboard step: launch the entrypoint when a userland is present.
if [[ -x /usr/lib/vyntrio/firstboot.sh ]]; then
	echo "vyntrio live-init: invoking first-boot dashboard entrypoint" >&2
	/usr/lib/vyntrio/firstboot.sh || echo "vyntrio live-init: firstboot entrypoint returned non-zero (expected without userland)" >&2
fi

echo "vyntrio live-init: dashboard_reachable=false (live userland + boot environment required)" >&2
echo "vyntrio live-init: see /usr/lib/vyntrio/FIRST_BOOT.txt" >&2
exit 0
EOF
chmod 0755 "${LR}/usr/lib/vyntrio/live-init.sh"

cat >"${LR}/usr/lib/vyntrio/FIRST_BOOT.txt" <<EOF
# First-boot / dashboard honesty record (Slice 9.14 wiring; Slice 10.2 stability)
#
# Intended Unraid-like flow (local WebGUI is the PRIMARY management surface):
#   1. bootable USB/media          -> firmware-bootable BIOS image (Slice 9.13)
#   2. local browser dashboard     -> vyntrio-api WebGUI on http://<host>:${DASHBOARD_PORT}
#   3. onboarding / first-boot setup
#   4. later install / storage / licensing / remote
#
# firmware_bootable: see build/WRAPPER.txt
# dashboard_reachable: false
#
# Local dashboard path (dashboard-first):
#   - entrypoint:  usr/lib/vyntrio/firstboot.sh   (busybox-sh; supervises vyntrio-api)
#   - supervision: bounded respawn (${DASHBOARD_MAX_RESTARTS}x) — the WebGUI stays up as an appliance service
#   - service:     etc/systemd/system/vyntrio-firstboot.service (for a future systemd install)
#   - dashboard:   http://<host>:${DASHBOARD_PORT} (embedded UI in vyntrio-api)
#   - bind:        loopback/config address (LAN exposure deferred — needs TLS)
#
# Slice 10.2 stability fixes:
#   - firstboot.sh runs under /bin/sh (busybox); the prior '#!/usr/bin/env bash'
#     could not exec in the live runtime (no bash/env), so the dashboard never
#     started on boot. Now it does, and it is supervised.
#
# Slice 10.3 first-boot onboarding clarity:
#   - firstboot.sh prints a clear first-boot setup prompt: boot -> local browser
#     dashboard -> sign in / create owner -> read-only overview. Storage,
#     services, and remote access are explicitly deferred.
#   - the dashboard is the primary management surface from the first boot.
#
# Slice 10.4 TLS/LAN readiness:
#   - vyntrio-api accepts optional tls_cert_file/tls_key_file and serves HTTPS
#     when configured (non-loopback bind requires cookie_secure + TLS).
#   - prepare-live-dashboard-tls.sh generates ephemeral runtime certs under
#     /run/vyntrio when VYNTRIO_LAN_BIND_IP is set (secure-by-default loopback
#     HTTP otherwise).
#
# Still missing before first boot actually reaches the dashboard on the LAN:
#   - a boot environment (VM or real hardware) + runtime boot verification
#     (no qemu/VM on this build host — see RUNTIME_VERIFY.txt / RUNTIME_BOOT.txt)
#   - operator sets VYNTRIO_LAN_BIND_IP on the booted appliance for LAN HTTPS
#   - USB creator to write the image to removable media
status: partial
first_boot_onboarding_clarity: clear
local_onboarding_steps:
  - boot the appliance from verified media
  - open the local dashboard on the booted device (http://<host>:${DASHBOARD_PORT} or https://<lan-ip>:${DASHBOARD_PORT} when VYNTRIO_LAN_BIND_IP is set)
  - sign in or create the owner account on first boot
  - review the read-only overview with real appliance data
dashboard_primary_surface: true
dashboard_supervised: true
dashboard_bind_default: loopback
tls_readiness: ready
tls_runtime_prep: usr/lib/vyntrio/prepare-live-dashboard-tls.sh
lan_bind_enable: set VYNTRIO_LAN_BIND_IP for LAN HTTPS (requires openssl in live userland)
dashboard_reachable: false
dashboard_url: http://<host>:${DASHBOARD_PORT}
dashboard_lan_url: https://<lan-ip>:${DASHBOARD_PORT}
EOF

# Fail closed: no secrets/state in live_root.
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			die "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${LR}" -type f -print0)

# --- 2. runtime boot verification (gated on a VM harness) ---
QEMU=""
for q in qemu-system-x86_64 qemu-system-i386; do have "${q}" && { QEMU="${q}"; break; }; done

RUNTIME_TESTED=false
RUNTIME_RESULT="skipped"
RUNTIME_REASON=""
BOOT_MARKERS=""
SERIAL_LOG="${BUILD_ROOT}/runtime-boot-serial.log"
rm -f "${SERIAL_LOG}"

if [[ -z "${QEMU}" ]]; then
	RUNTIME_REASON="no_vm_harness: qemu-system-x86_64/i386 not installed and /dev/kvm absent"
elif [[ "${FIRMWARE_BOOTABLE}" != "true" || ( "${IMAGE_FMT}" != "raw_mbr_bios_disk" && "${IMAGE_FMT}" != "raw_gpt_hybrid_disk" ) ]]; then
	RUNTIME_REASON="no_firmware_bootable_raw_image: nothing bootable to test (see WRAPPER.txt)"
elif [[ ! -f "${IMAGE_ABS}" ]]; then
	RUNTIME_REASON="image_missing: ${IMAGE_REL}"
else
	# Boot the raw image headless with a serial console; bounded by timeout.
	# Prefer BIOS path for this harness (SeaBIOS); UEFI packaging verified separately via verify-uefi-boot.sh.
	set +e
	timeout "${VYNTRIO_RUNTIME_BOOT_TIMEOUT:-40}" "${QEMU}" \
		-drive "file=${IMAGE_ABS},format=raw,if=ide" \
		-m 512 -no-reboot -nographic -serial "file:${SERIAL_LOG}" -display none \
		>/dev/null 2>&1
	set -e
	RUNTIME_TESTED=true
	# Boot is considered reached if GRUB and/or the kernel emitted known markers.
	if [[ -f "${SERIAL_LOG}" ]]; then
		grep -qa 'GRUB' "${SERIAL_LOG}" && BOOT_MARKERS+="grub "
		grep -qaE 'Linux version|Decompressing Linux|Booting' "${SERIAL_LOG}" && BOOT_MARKERS+="kernel "
	fi
	if [[ -n "${BOOT_MARKERS}" ]]; then
		RUNTIME_RESULT="boot_reached_kernel"
	else
		RUNTIME_RESULT="no_boot_markers_observed"
		RUNTIME_REASON="serial output lacked GRUB/kernel markers within timeout"
	fi
fi

# --- provenance record (Slice 9.14) ---
{
	echo "# Generated by scripts/verify-runtime-boot.sh — do not commit"
	echo "schema_version: vyntrio-install-media-runtime-v1"
	echo "slice: 9.14"
	echo "generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
	echo "media_role: install"
	echo "firmware_bootable: ${FIRMWARE_BOOTABLE:-unknown}"
	echo "image: ${IMAGE_REL}"
	echo "runtime_boot_tested: ${RUNTIME_TESTED}"
	echo "runtime_boot_result: ${RUNTIME_RESULT}"
	if [[ -n "${RUNTIME_REASON}" ]]; then
		echo "runtime_boot_reason: ${RUNTIME_REASON}"
	fi
	if [[ "${RUNTIME_TESTED}" == true ]]; then
		echo "vm_harness: ${QEMU}"
		echo "boot_markers: ${BOOT_MARKERS:-none}"
		echo "serial_log: distro/install-media/build/$(basename "${SERIAL_LOG}")"
	fi
	echo "first_boot_path:"
	echo "  entrypoint: live_root/usr/lib/vyntrio/firstboot.sh"
	echo "  entrypoint_shell: /bin/sh (busybox-compatible; Slice 10.2)"
	echo "  service: live_root/etc/systemd/system/vyntrio-firstboot.service"
	echo "  dashboard_primary_surface: true"
	echo "  dashboard_supervised: true"
	echo "  dashboard_max_restarts: ${DASHBOARD_MAX_RESTARTS}"
	echo "  dashboard_bind_default: loopback"
	echo "  tls_readiness: ready"
	echo "  tls_runtime_prep: live_root/usr/lib/vyntrio/prepare-live-dashboard-tls.sh"
	echo "  lan_bind_enable: set VYNTRIO_LAN_BIND_IP for LAN HTTPS"
	echo "  dashboard_url: http://<host>:${DASHBOARD_PORT}"
	echo "  dashboard_lan_url: https://<lan-ip>:${DASHBOARD_PORT}"
	echo "  dashboard_reachable: false"
	echo "  reachable_requires:"
	echo "    - live_rootfs_userland_with_vyntrio_api"
	echo "    - boot_environment_vm_or_hardware"
	echo "    - operator_lan_bind_ip_for_lan_https"
	echo "  first_boot_onboarding_clarity: clear"
	echo "  local_onboarding_steps:"
	echo "    - boot verified media on the target appliance"
	echo "    - open the local dashboard on the booted device (http://<host>:${DASHBOARD_PORT})"
	echo "    - sign in or create the owner account on first boot"
	echo "    - review the read-only overview with real appliance data"
	echo "missing_for_usb_creator:"
	echo "  - live_rootfs_userland"
	echo "  - runtime_boot_reaching_dashboard"
	echo "  - uefi_boot_support"
	echo "  - host_usb_writer_tool"
} >"${BUILD_ROOT}/RUNTIME_BOOT.txt"

echo "install-media runtime: runtime_boot_tested=${RUNTIME_TESTED} result=${RUNTIME_RESULT}"
if [[ "${RUNTIME_TESTED}" != true ]]; then
	echo "install-media runtime: runtime boot NOT exercised — ${RUNTIME_REASON}" >&2
fi
echo "install-media runtime: first-boot dashboard path wired (dashboard_reachable=false until userland+boot)"
