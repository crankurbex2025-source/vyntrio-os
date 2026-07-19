#!/usr/bin/env bash
# Runtime boot proof for the Vyntrio USB appliance image (UEFI qemu).
# Proves or fails closed: kernel → early initramfs → squashfs → firstboot → HTTP.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD="${ROOT}/distro/install-media/build"
STAGED="${ROOT}/distro/release/staging/vyntrio-install-media.img"
BUILT="${BUILD}/vyntrio-install-media.img"
IMG="${VYNTRIO_APPLIANCE_IMG:-}"
[[ -z "${IMG}" ]] && [[ -f "${STAGED}" ]] && IMG="${STAGED}"
[[ -z "${IMG}" ]] && [[ -f "${BUILT}" ]] && IMG="${BUILT}"
[[ -n "${IMG}" && -f "${IMG}" ]] || { echo "prove-appliance-boot: image missing" >&2; exit 2; }

SERIAL_LOG="${BUILD}/appliance-boot-serial.log"
RECORD="${BUILD}/APPLIANCE_BOOT_PROOF.txt"
HTTP_PORT="${VYNTRIO_RUNTIME_HTTP_PORT:-18080}"
BOOT_TIMEOUT="${VYNTRIO_RUNTIME_BOOT_TIMEOUT:-240}"
MEMORY="${VYNTRIO_RUNTIME_MEMORY:-2048}"
FIRMWARE="${VYNTRIO_RUNTIME_FIRMWARE:-uefi}" # uefi|bios

QEMU="$(command -v qemu-system-x86_64 || true)"
[[ -n "${QEMU}" ]] || { echo "prove-appliance-boot: qemu-system-x86_64 required" >&2; exit 2; }

OVMF_CODE=""
for c in /usr/share/OVMF/OVMF_CODE_4M.fd /usr/share/OVMF/OVMF_CODE.fd /usr/share/qemu/OVMF.fd; do
	[[ -f "$c" ]] && { OVMF_CODE="$c"; break; }
done
OVMF_VARS_SRC=""
for c in /usr/share/OVMF/OVMF_VARS_4M.fd /usr/share/OVMF/OVMF_VARS.fd; do
	[[ -f "$c" ]] && { OVMF_VARS_SRC="$c"; break; }
done

mkdir -p "${BUILD}"
WORK="$(mktemp -d)"
cleanup() {
	[[ -n "${QPID:-}" ]] && kill "${QPID}" 2>/dev/null || true
	wait "${QPID:-}" 2>/dev/null || true
	rm -rf "${WORK}"
}
trap cleanup EXIT

# Work on a copy so persistence writes don't mutate the staged artifact unexpectedly.
IMG_COPY="${WORK}/appliance.img"
cp --reflink=auto -f "${IMG}" "${IMG_COPY}" 2>/dev/null || cp -f "${IMG}" "${IMG_COPY}"

# Clear any prior overlay upper dir so squashfs updates are not masked by old files.
if [[ -f "${ROOT}/distro/install-media/build/APPLIANCE_DATA_OFFSET.txt" ]] \
	&& command -v losetup >/dev/null 2>&1 && command -v mount >/dev/null 2>&1; then
	DATA_OFF="$(cat "${ROOT}/distro/install-media/build/APPLIANCE_DATA_OFFSET.txt")"
	LOOP_CLR="$(losetup -f --show -o "${DATA_OFF}" "${IMG_COPY}" 2>/dev/null || true)"
	if [[ -n "${LOOP_CLR}" ]]; then
		CLR_MNT="${WORK}/clr"
		mkdir -p "${CLR_MNT}"
		if mount "${LOOP_CLR}" "${CLR_MNT}" 2>/dev/null; then
			rm -rf "${CLR_MNT}/vyntrio/overlay/upper" "${CLR_MNT}/vyntrio/overlay/work"
			mkdir -p "${CLR_MNT}/vyntrio/overlay/upper" "${CLR_MNT}/vyntrio/overlay/work"
			umount "${CLR_MNT}" 2>/dev/null || true
		fi
		losetup -d "${LOOP_CLR}" 2>/dev/null || true
	fi
fi

ACCEL_ARGS=()
ACCEL=tcg
if [[ -e /dev/kvm && -r /dev/kvm ]]; then
	ACCEL_ARGS=(-accel kvm)
	ACCEL=kvm
fi

rm -f "${SERIAL_LOG}"
: >"${SERIAL_LOG}"

declare -a QEMU_ARGS=(
	-m "${MEMORY}"
	-no-reboot
	-display none
	-serial "file:${SERIAL_LOG}"
	-netdev "user,id=net0,hostfwd=tcp:127.0.0.1:${HTTP_PORT}-:8080"
	-device "e1000,netdev=net0"
)

if [[ "${FIRMWARE}" == "uefi" ]]; then
	[[ -n "${OVMF_CODE}" ]] || { echo "prove-appliance-boot: OVMF CODE missing" >&2; exit 2; }
	VARS="${WORK}/OVMF_VARS.fd"
	if [[ -n "${OVMF_VARS_SRC}" ]]; then
		cp -f "${OVMF_VARS_SRC}" "${VARS}"
	else
		# Some hosts ship combined OVMF.fd only.
		VARS=""
	fi
	QEMU_ARGS+=(
		-machine q35
		"${ACCEL_ARGS[@]}"
		-drive "if=pflash,format=raw,readonly=on,file=${OVMF_CODE}"
	)
	[[ -n "${VARS}" ]] && QEMU_ARGS+=(-drive "if=pflash,format=raw,file=${VARS}")
	QEMU_ARGS+=(-drive "file=${IMG_COPY},format=raw,if=virtio,cache=writeback")
else
	QEMU_ARGS+=(
		-machine pc
		"${ACCEL_ARGS[@]}"
		-drive "file=${IMG_COPY},format=raw,if=ide,cache=writeback"
		-boot c
	)
fi

echo "prove-appliance-boot: image=${IMG} firmware=${FIRMWARE} accel=${ACCEL} timeout=${BOOT_TIMEOUT}s hostfwd=:${HTTP_PORT}->8080"
echo "prove-appliance-boot: qemu ${QEMU} ${QEMU_ARGS[*]}" | sed "s|${WORK}|\$WORK|g"

set +e
"${QEMU}" "${QEMU_ARGS[@]}" >/dev/null 2>&1 &
QPID=$!
set -e

DASH_STATUS=""
DASH_BODY=""
DASH_REACHABLE=false
READYZ_STATUS=""
for i in $(seq 1 "${BOOT_TIMEOUT}"); do
	if ! kill -0 "${QPID}" 2>/dev/null; then
		echo "prove-appliance-boot: qemu exited early at t=${i}s"
		break
	fi
	# Appliance prefers LAN HTTPS once TLS prep succeeds; also accept loopback HTTP.
	for url in \
		"https://127.0.0.1:${HTTP_PORT}/" \
		"http://127.0.0.1:${HTTP_PORT}/"; do
		curl_opts=( -s -o "${WORK}/body" -w '%{http_code}' --max-time 2 )
		[[ "${url}" == https://* ]] && curl_opts+=( -k )
		DASH_STATUS="$(curl "${curl_opts[@]}" "${url}" 2>/dev/null || true)"
		if [[ "${DASH_STATUS}" == "200" || "${DASH_STATUS}" == "302" || "${DASH_STATUS}" == "401" ]]; then
			DASH_REACHABLE=true
			DASH_BODY="$(head -c 200 "${WORK}/body" 2>/dev/null || true)"
			ready_url="${url%/}/readyz"
			ready_opts=( -s -o /dev/null -w '%{http_code}' --max-time 2 )
			[[ "${ready_url}" == https://* ]] && ready_opts+=( -k )
			READYZ_STATUS="$(curl "${ready_opts[@]}" "${ready_url}" 2>/dev/null || true)"
			echo "prove-appliance-boot: ${url} -> ${DASH_STATUS} at t=${i}s readyz=${READYZ_STATUS}"
			break 2
		fi
	done
	# Progress breadcrumbs from serial
	if (( i % 15 == 0 )); then
		tail -n 5 "${SERIAL_LOG}" 2>/dev/null | sed 's/^/  serial> /' || true
	fi
	sleep 1
done

kill "${QPID}" 2>/dev/null || true
wait "${QPID}" 2>/dev/null || true
QPID=""

# Analyze serial
mark() { grep -aEiq "$1" "${SERIAL_LOG}" 2>/dev/null; }
MARKERS=()
mark 'GNU GRUB|GRUB version|Welcome to GRUB' && MARKERS+=("grub")
mark 'Linux version|Decompressing Linux|Booting the kernel' && MARKERS+=("kernel")
mark 'vyntrio early:' && MARKERS+=("early_init")
mark 'mounting squashfs|squashfs from' && MARKERS+=("squashfs")
mark 'switch_root' && MARKERS+=("switch_root")
mark 'overlay mounted|overlay failed' && MARKERS+=("overlay")
mark 'vyntrio appliance:' && MARKERS+=("appliance_init")
mark 'persistence' && MARKERS+=("persistence")
mark 'vyntrio firstboot:' && MARKERS+=("firstboot")
mark 'starting local dashboard|using runtime LAN HTTPS|TLS' && MARKERS+=("dashboard_start")
mark 'certificate ready|LAN dashboard HTTPS ready' && MARKERS+=("tls_ready")

KERNEL_OK=false; [[ " ${MARKERS[*]} " == *" kernel "* ]] && KERNEL_OK=true
EARLY_OK=false; [[ " ${MARKERS[*]} " == *" early_init "* ]] && EARLY_OK=true
SQUASH_OK=false; [[ " ${MARKERS[*]} " == *" squashfs "* ]] && SQUASH_OK=true
FIRSTBOOT_OK=false; [[ " ${MARKERS[*]} " == *" firstboot "* ]] && FIRSTBOOT_OK=true

BLOCKER="none"
if [[ "${DASH_REACHABLE}" != true ]]; then
	if [[ "${KERNEL_OK}" != true ]]; then BLOCKER="kernel_did_not_start_or_no_serial"
	elif [[ "${EARLY_OK}" != true ]]; then BLOCKER="early_initramfs_did_not_run"
	elif [[ "${SQUASH_OK}" != true ]]; then BLOCKER="squashfs_mount_or_handoff_failed"
	elif [[ "${FIRSTBOOT_OK}" != true ]]; then BLOCKER="firstboot_did_not_execute"
	else BLOCKER="api_or_webui_unreachable"; fi
fi

{
	echo "# Generated by scripts/prove-appliance-boot.sh — do not commit"
	echo "schema_version: vyntrio-appliance-boot-proof-v1"
	echo "generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
	echo "image: ${IMG}"
	echo "image_sha256: $(sha256sum "${IMG}" | awk '{print $1}')"
	echo "firmware: ${FIRMWARE}"
	echo "accel: ${ACCEL}"
	echo "serial_log: distro/install-media/build/appliance-boot-serial.log"
	echo "boot_timeout_s: ${BOOT_TIMEOUT}"
	echo "hostfwd_port: ${HTTP_PORT}"
	echo "markers: ${MARKERS[*]:-none}"
	echo "kernel_started: ${KERNEL_OK}"
	echo "early_init_started: ${EARLY_OK}"
	echo "squashfs_handoff: ${SQUASH_OK}"
	echo "firstboot_started: ${FIRSTBOOT_OK}"
	echo "dashboard_http_status: ${DASH_STATUS:-none}"
	echo "readyz_http_status: ${READYZ_STATUS:-none}"
	echo "dashboard_reachable_on_boot: ${DASH_REACHABLE}"
	echo "blocker: ${BLOCKER}"
	echo "runtime_boot_to_dashboard_verified: ${DASH_REACHABLE}"
} >"${RECORD}"

echo "prove-appliance-boot: markers=${MARKERS[*]:-none}"
echo "prove-appliance-boot: dashboard_reachable=${DASH_REACHABLE} http=${DASH_STATUS:-none} blocker=${BLOCKER}"
echo "prove-appliance-boot: serial_log=${SERIAL_LOG}"
echo "prove-appliance-boot: record=${RECORD}"

# Always dump last 80 lines of serial for the operator.
echo "----- serial tail -----"
tail -n 80 "${SERIAL_LOG}" 2>/dev/null || echo "(empty serial log)"
echo "----- end serial -----"

[[ "${DASH_REACHABLE}" == true ]]
