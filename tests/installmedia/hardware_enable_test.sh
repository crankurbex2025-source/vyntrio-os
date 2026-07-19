#!/usr/bin/env bash
# Verifies live-initramfs hardware enablement (Slice 9.16).
# Asserts the module set is staged + dep-resolved, the networking/module bring-up
# path is wired into /init, the record matches reality, the live userland still
# executes, prior slices are not regressed, and no honest boot/dashboard claim is
# overstated.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
LR="${ENVELOPE_ROOT}/live_root"
REC="${BUILD_ROOT}/HARDWARE_ENABLE.txt"
WRAP="${BUILD_ROOT}/WRAPPER.txt"

fail() { echo "installmedia hardware test: $*" >&2; exit 1; }
rget() { sed -n "s/^ *$1: //p" "${REC}" | head -1; }

[[ -f "${REC}" ]] || fail "HARDWARE_ENABLE.txt missing; run 'make install-media-hardware' first"

KVER="$(rget kernel_version)"
[[ -n "${KVER}" ]] || fail "record missing kernel_version"
MOD_ENABLED="$(rget enabled)"

# --- module enablement honesty: either really enabled, or an honest skip ---
if [[ "${MOD_ENABLED}" == "true" ]]; then
	MODDIR="${LR}/lib/modules/${KVER}"
	[[ -d "${MODDIR}" ]] || fail "modules enabled but ${MODDIR} missing"
	[[ -f "${MODDIR}/modules.dep" ]] || fail "modules.dep not regenerated in live_root"

	# At least one storage driver and one network driver must be present.
	kos="$(find "${MODDIR}" -name '*.ko*' 2>/dev/null | wc -l | tr -d ' ')"
	[[ "${kos}" -ge 1 ]] || fail "no .ko modules staged despite enabled=true"
	find "${MODDIR}" -path '*drivers/net/*' -name '*.ko*' | grep -q . \
		|| fail "no network driver module staged"
	find "${MODDIR}" \( -name 'ext4.ko*' -o -path '*scsi*' -o -path '*ata*' -o -path '*block*' -o -path '*nvme*' \) -name '*.ko*' | grep -q . \
		|| fail "no storage/filesystem driver module staged"

	# --- module load list references real, staged modules ---
	LOADLIST="${LR}/etc/vyntrio/modules.load"
	[[ -f "${LOADLIST}" ]] || fail "modules.load missing"
	while IFS= read -r m; do
		case "${m}" in ''|'#'*) continue ;; esac
		find "${MODDIR}" -name "${m}.ko*" | grep -q . \
			|| fail "modules.load references '${m}' but no matching .ko is staged"
	done <"${LOADLIST}"

	# --- record's staged count matches reality ---
	rec_ko="$(rget staged_ko_files)"
	[[ "${rec_ko}" == "${kos}" ]] || fail "record staged_ko_files=${rec_ko} != actual ${kos}"
else
	# Honest skip must record a concrete reason.
	grep -q '^  skip_reason: ' "${REC}" || fail "modules enabled=false must record a skip_reason"
fi

# --- networking bring-up beyond loopback is wired ---
[[ "$(rget bring_up_beyond_loopback)" == "true" ]] || fail "network bring_up_beyond_loopback must be true"
HW_INIT="${LR}/usr/lib/vyntrio/hw-init.sh"
[[ -x "${HW_INIT}" ]] || fail "hw-init.sh missing or not executable"
grep -q 'udhcpc' "${HW_INIT}" || fail "hw-init.sh must attempt DHCP (udhcpc) beyond loopback"
grep -q '/sys/class/net' "${HW_INIT}" || fail "hw-init.sh must enumerate network interfaces"
[[ -x "${LR}/usr/share/udhcpc/default.script" ]] || fail "udhcpc lease handler missing/not executable"

# --- module-load applets present (busybox symlinks) ---
for a in modprobe insmod udhcpc; do
	[[ -L "${LR}/bin/${a}" || -x "${LR}/bin/${a}" ]] || fail "missing busybox applet /bin/${a}"
done

# --- /init integrates hw-init BEFORE first-boot, and still hands off to firstboot ---
[[ -x "${LR}/init" ]] || fail "/init missing or not executable"
grep -q 'hw-init.sh' "${LR}/init" || fail "/init must invoke hw-init.sh"
grep -q 'firstboot.sh' "${LR}/init" || fail "/init must still hand off to firstboot.sh (9.14/9.15)"
# hw-init must be invoked ahead of firstboot in the file.
hw_line="$(grep -n 'hw-init.sh' "${LR}/init" | head -1 | cut -d: -f1)"
fb_line="$(grep -n 'firstboot.sh' "${LR}/init" | head -1 | cut -d: -f1)"
[[ "${hw_line}" -lt "${fb_line}" ]] || fail "/init must run hw-init before firstboot"

# --- record matches reality: chroot exec is re-run here, not trusted ---
if command -v chroot >/dev/null 2>&1; then
	if chroot "${LR}" /bin/busybox sh -c 'exit 0' >/dev/null 2>&1; then
		[[ "$(rget shell_exec_in_chroot)" == "true" ]] || fail "shell executes but record says shell_exec_in_chroot != true"
	fi
fi

# --- initramfs re-emitted with modules (when enabled) ---
if [[ "$(rget reemitted)" == "true" ]]; then
	IMG="${BUILD_ROOT}/vyntrio-live-initramfs.cpio.gz"
	[[ -s "${IMG}" ]] || fail "initramfs recorded as reemitted but missing/empty"
	gzip -t "${IMG}" 2>/dev/null || fail "initramfs is not valid gzip"
	listing="$(gzip -dc "${IMG}" 2>/dev/null | cpio -t 2>/dev/null || true)"
	# Regression guard: still a bootable-shaped live initramfs.
	grep -qE '^\.?/?init$' <<<"${listing}" || fail "initramfs lost /init at root"
	grep -q 'bin/busybox' <<<"${listing}" || fail "initramfs lost busybox"
	if [[ "${MOD_ENABLED}" == "true" ]]; then
		grep -q "lib/modules/${KVER}/modules.dep" <<<"${listing}" \
			|| fail "initramfs does not carry kernel modules despite enabled=true"
		[[ "$(rget contains_kernel_modules)" == "true" ]] || fail "record contains_kernel_modules != true"
	fi
fi

# --- honesty: no unearned boot/dashboard/wiring claims ---
grep -q '^  wired_into_image_initrd: false' "${REC}" || fail "must not claim initramfs is wired into image yet"
grep -q '^  boot_verified: false' "${REC}" || fail "boot_verified must stay false without a VM"
grep -q '^  dashboard_reachable_on_boot: false' "${REC}" || fail "dashboard_reachable_on_boot must stay false"

# --- modules are staged only; not loaded on the build host (safety) ---
grep -q 'NOT loaded here' "${REC}" || fail "record must state modules are not loaded on the build host"

# --- shipped media stays state-free ---
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			fail "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${LR}" -type f -print0)

# --- no regression: 9.14/9.15 userland intact ---
[[ -x "${LR}/bin/busybox" ]] || fail "busybox regressed"
[[ -f "${LR}/lib64/ld-linux-x86-64.so.2" ]] || fail "dynamic loader regressed"
[[ -x "${LR}/usr/bin/vyntrio-api" || ! -e "${LR}/usr/bin/vyntrio-api" ]] || fail "vyntrio-api present but not executable"
[[ -x "${LR}/usr/lib/vyntrio/firstboot.sh" ]] || fail "9.14 firstboot.sh regressed"
[[ -f "${LR}/etc/systemd/system/vyntrio-firstboot.service" ]] || fail "9.14 first-boot service regressed"

# --- no regression: modules must NOT leak into target-disk payload ---
[[ ! -d "${ENVELOPE_ROOT}/payload/lib/modules" ]] || fail "kernel modules leaked into target-disk payload"
mapfile -t pf < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
[[ "${#pf[@]}" -eq 6 ]] || fail "payload file count regressed (${#pf[@]} != 6)"

# --- no regression: firmware image still bootable ---
if [[ -f "${WRAP}" ]]; then
	[[ "$(sed -n 's/^firmware_bootable: //p' "${WRAP}")" == "true" ]] || fail "firmware bootability regressed"
	grep -q '^structural_verification: pass' "${WRAP}" || fail "firmware structural verification regressed"
fi

# --- fail-closed: missing live_root must fail ---
tmp="$(mktemp -d)"; trap 'rm -rf "${tmp}"' EXIT
if VYNTRIO_INSTALL_ENVELOPE_ROOT="${tmp}/missing" \
	VYNTRIO_INSTALL_MEDIA_BUILD_ROOT="${tmp}/build" \
	bash "${ROOT}/scripts/enable-live-initramfs-hardware.sh" >/dev/null 2>&1; then
	fail "expected failure when live_root missing"
fi

echo "installmedia hardware test: ok (modules_enabled=${MOD_ENABLED}, staged_ko=$(rget staged_ko_files), net_bring_up=beyond_loopback)"
