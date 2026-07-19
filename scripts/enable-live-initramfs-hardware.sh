#!/usr/bin/env bash
# Live-initramfs hardware enablement (Block 9 / Slice 9.16).
#
# Goal: make the live-rootfs closer to a REAL first boot by adding the smallest
# viable hardware bring-up to the live initramfs:
#   1. a curated, dependency-resolved set of storage + network kernel modules,
#      matched to the same host kernel the boot chain ships (Slice 9.12);
#   2. a module-load + networking (DHCP) bring-up path invoked from /init;
#   3. a re-emitted live initramfs that now carries those modules.
#
# Honesty (read carefully):
#   - This runs AFTER Slice 9.15 (compose-live-rootfs). It augments live_root/;
#     it does not recompose the userland and does not touch the 9.15 record.
#   - Kernel modules are staged and their metadata regenerated (depmod), and a
#     sample module is validated with modinfo. Modules are NOT loaded here:
#     modprobe would load into the BUILD HOST kernel, which is out of scope and
#     unsafe. Actual module load happens only inside a booted live system.
#   - The live initramfs is still NOT wired into the image initrd, the image is
#     still NOT booted (no VM harness), and the dashboard is still NOT reachable
#     on boot. Those remain deferred and are recorded honestly.
#   - Live runtime vs target-disk payload boundary is preserved: modules land in
#     live_root/lib/modules (live runtime), never in payload/. No target writes.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVELOPE_ROOT="${VYNTRIO_INSTALL_ENVELOPE_ROOT:-${ROOT}/distro/install-media/envelope}"
BUILD_ROOT="${VYNTRIO_INSTALL_MEDIA_BUILD_ROOT:-${ROOT}/distro/install-media/build}"
LR="${ENVELOPE_ROOT}/live_root"
INITRAMFS_NAME="vyntrio-live-initramfs.cpio.gz"

KVER="${VYNTRIO_LIVE_KVER:-$(uname -r)}"
MODSRC="${VYNTRIO_MODULES_DIR:-/lib/modules/${KVER}}"
MODDST="${LR}/lib/modules/${KVER}"

# Smallest viable driver set for a first boot on common PC/server hardware.
# Missing modules on a given kernel are recorded, not fatal (host-dependent).
STORAGE_MODULES="ahci libahci sd_mod sr_mod nvme virtio_blk virtio_scsi virtio_pci ata_piix ata_generic xhci_pci ehci_pci uhci_hcd usb_storage"
FS_MODULES="ext4 vfat isofs overlay squashfs"
NET_MODULES="virtio_net e1000 e1000e r8169 igb tg3"

die() { echo "enable-live-initramfs-hardware: $*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

[[ -d "${LR}" ]] || die "live_root missing; run 'make install-media-live-rootfs' first"
[[ -f "${LR}/init" ]] || die "live_root/init missing; run 'make install-media-live-rootfs' first"
mkdir -p "${BUILD_ROOT}"

# --- module staging (honest skip if this host has no matching modules tree) ---
MODULES_ENABLED=false
MODULES_SKIP_REASON=""
declare -A WANT=()
RESOLVED_TOP=()
MISSING_NAMES=()
STORAGE_PRESENT=()
NET_PRESENT=()

DEPFILE="${MODSRC}/modules.dep"

find_module_line() {
	# Print the modules.dep line whose key basename matches $1, else nothing.
	awk -v n="$1" '
		{
			key = $1; sub(/:$/, "", key)
			b = key; sub(/.*\//, "", b); sub(/\.ko.*$/, "", b)
			if (b == n) { print $0; exit }
		}' "${DEPFILE}"
}

resolve_group() {
	# $1=group label (storage|net|fs), $2..=module names
	local label="$1"; shift
	local name line key rest d
	for name in "$@"; do
		line="$(find_module_line "${name}" 2>/dev/null || true)"
		if [[ -z "${line}" ]]; then
			MISSING_NAMES+=("${name}")
			continue
		fi
		key="${line%%:*}"
		WANT["${key}"]=1
		rest="${line#*:}"
		# modules.dep lists the FULL transitive dep chain per line.
		for d in ${rest}; do
			WANT["${d}"]=1
		done
		RESOLVED_TOP+=("${name}")
		case "${label}" in
			storage) STORAGE_PRESENT+=("${name}") ;;
			net) NET_PRESENT+=("${name}") ;;
		esac
	done
}

if [[ ! -f "${DEPFILE}" ]]; then
	MODULES_SKIP_REASON="no modules.dep at ${MODSRC} (host kernel modules unavailable)"
elif ! have awk; then
	MODULES_SKIP_REASON="awk unavailable; cannot resolve module dependencies"
else
	resolve_group storage ${STORAGE_MODULES}
	resolve_group fs ${FS_MODULES}
	resolve_group net ${NET_MODULES}

	COPIED=0
	MISSING_FILES=()
	rm -rf "${MODDST}"
	mkdir -p "${MODDST}"
	for rel in "${!WANT[@]}"; do
		src="${MODSRC}/${rel}"
		if [[ ! -f "${src}" ]]; then
			MISSING_FILES+=("${rel}")
			continue
		fi
		dst="${MODDST}/${rel}"
		mkdir -p "$(dirname "${dst}")"
		cp -aL "${src}" "${dst}"
		COPIED=$((COPIED + 1))
	done

	# Metadata that helps depmod produce a complete, correct dep graph.
	for meta in modules.order modules.builtin modules.builtin.modinfo; do
		[[ -f "${MODSRC}/${meta}" ]] && cp -aL "${MODSRC}/${meta}" "${MODDST}/${meta}" || true
	done

	DEP_REGENERATED=false
	if [[ "${COPIED}" -gt 0 ]] && have depmod; then
		if depmod -b "${LR}" "${KVER}" >/dev/null 2>&1; then
			DEP_REGENERATED=true
		fi
	fi

	if [[ "${COPIED}" -gt 0 && "${DEP_REGENERATED}" == true ]]; then
		MODULES_ENABLED=true
	else
		MODULES_SKIP_REASON="no modules copied or depmod failed (copied=${COPIED})"
	fi
fi

# --- module load list (top-level names in a sensible load order) ---
mkdir -p "${LR}/etc/vyntrio"
if [[ "${MODULES_ENABLED}" == true ]]; then
	{
		echo "# Vyntrio live-boot kernel module load list (Slice 9.16)."
		echo "# Loaded by usr/lib/vyntrio/hw-init.sh at live boot, in order."
		for m in "${RESOLVED_TOP[@]}"; do echo "${m}"; done
	} >"${LR}/etc/vyntrio/modules.load"
else
	rm -f "${LR}/etc/vyntrio/modules.load"
fi

# --- busybox applets required for hardware bring-up (additive to 9.15) ---
if [[ -x "${LR}/bin/busybox" ]]; then
	for a in modprobe insmod rmmod lsmod depmod udhcpc route; do
		ln -sf /bin/busybox "${LR}/bin/${a}"
	done
fi

# --- udhcpc lease handler: configure the interface DHCP hands us ---
mkdir -p "${LR}/usr/share/udhcpc"
cat >"${LR}/usr/share/udhcpc/default.script" <<'EOF'
#!/bin/sh
# Minimal busybox udhcpc lease handler (Slice 9.16).
[ -n "$1" ] || exit 1
case "$1" in
	deconfig)
		ifconfig "$interface" 0.0.0.0 2>/dev/null || true
		ip link set "$interface" up 2>/dev/null || true
		;;
	bound|renew)
		ifconfig "$interface" "$ip" netmask "${subnet:-255.255.255.0}" up 2>/dev/null || true
		if [ -n "$router" ]; then
			ip route del default 2>/dev/null || true
			ip route add default via "$router" 2>/dev/null || true
		fi
		if [ -n "$dns" ]; then
			: >/etc/resolv.conf
			for d in $dns; do echo "nameserver $d" >>/etc/resolv.conf; done
		fi
		;;
esac
exit 0
EOF
chmod 0755 "${LR}/usr/share/udhcpc/default.script"

# --- hw-init.sh: load modules + bring up networking beyond loopback ---
cat >"${LR}/usr/lib/vyntrio/hw-init.sh" <<'EOF'
#!/bin/sh
# Vyntrio live hardware bring-up (Slice 9.16). Runs inside the booted live
# system BEFORE the first-boot dashboard path. Loads storage/network kernel
# modules, then attempts DHCP on each non-loopback interface.
#
# Honest limit: this executes only in a real boot with a matching kernel. It is
# not exercised during build-host verification (that would touch the host kernel).

# Loopback first (retained from 9.15).
ip link set lo up 2>/dev/null || ifconfig lo up 2>/dev/null || true

# Load curated kernel modules for storage + network.
if [ -f /etc/vyntrio/modules.load ]; then
	while IFS= read -r m; do
		case "$m" in ''|'#'*) continue ;; esac
		modprobe "$m" 2>/dev/null || true
	done </etc/vyntrio/modules.load
fi

# Settle device nodes if busybox mdev is available.
if [ -e /proc/sys/kernel/hotplug ]; then
	echo /bin/busybox >/proc/sys/kernel/hotplug 2>/dev/null || true
fi
busybox mdev -s 2>/dev/null || true

# Networking beyond loopback: DHCP each ethernet-like interface.
DHCP_SCRIPT=/usr/share/udhcpc/default.script
for netdev in /sys/class/net/*; do
	[ -e "$netdev" ] || continue
	dev="${netdev##*/}"
	[ "$dev" = "lo" ] && continue
	ip link set "$dev" up 2>/dev/null || ifconfig "$dev" up 2>/dev/null || true
	if command -v udhcpc >/dev/null 2>&1; then
		udhcpc -i "$dev" -s "$DHCP_SCRIPT" -t 3 -T 2 -n -q 2>/dev/null || true
	fi
done

echo "vyntrio hw-init: modules loaded (best-effort), network brought up" >&2
exit 0
EOF
chmod 0755 "${LR}/usr/lib/vyntrio/hw-init.sh"

# --- /init: invoke hw-init before the first-boot path (retains firstboot handoff) ---
cat >"${LR}/init" <<'EOF'
#!/bin/sh
# Vyntrio live-rootfs init (Slice 9.15 userland + Slice 9.16 hardware enablement).
# Runs as PID 1 in the live runtime: mount pseudo-fs, bring up hardware, then
# hand off to the first-boot dashboard path.
/bin/busybox mount -t proc proc /proc 2>/dev/null
/bin/busybox mount -t sysfs sysfs /sys 2>/dev/null
/bin/busybox mount -t devtmpfs devtmpfs /dev 2>/dev/null
/bin/busybox mount -t tmpfs tmpfs /run 2>/dev/null

# Hardware bring-up: kernel modules (storage/network) + networking (Slice 9.16).
if [ -x /usr/lib/vyntrio/hw-init.sh ]; then
	/usr/lib/vyntrio/hw-init.sh || echo "vyntrio live-init: hw-init returned non-zero" >&2
else
	/bin/busybox ip link set lo up 2>/dev/null || /bin/busybox ifconfig lo up 2>/dev/null
fi

echo "vyntrio live-init: minimal userland up (hardware-enabled)" >&2

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

# --- verification (no module load; safe on the build host) ---
SHELL_EXEC=false
if [[ -x "${LR}/bin/busybox" ]] && have chroot; then
	chroot "${LR}" /bin/busybox sh -c 'exit 0' >/dev/null 2>&1 && SHELL_EXEC=true
fi

MODINFO_OK=false
SAMPLE_MODULE=""
if [[ "${MODULES_ENABLED}" == true ]] && have modinfo; then
	SAMPLE_MODULE="$(find "${MODDST}" -name 'ext4.ko*' -o -name 'virtio_net.ko*' 2>/dev/null | head -1)"
	[[ -z "${SAMPLE_MODULE}" ]] && SAMPLE_MODULE="$(find "${MODDST}" -name '*.ko*' 2>/dev/null | head -1)"
	if [[ -n "${SAMPLE_MODULE}" ]] && modinfo "${SAMPLE_MODULE}" >/dev/null 2>&1; then
		MODINFO_OK=true
	fi
fi

DEP_PRESENT=false
[[ -f "${MODDST}/modules.dep" ]] && DEP_PRESENT=true

# --- re-emit the live initramfs (now carrying kernel modules) ---
INITRAMFS_REEMITTED=false
INITRAMFS_HAS_MODULES=false
if [[ -x "${LR}/bin/busybox" ]] && have cpio && have gzip; then
	( cd "${LR}" && find . -print0 | cpio --null -o --format=newc 2>/dev/null | gzip -9 ) \
		>"${BUILD_ROOT}/${INITRAMFS_NAME}" 2>/dev/null
	if [[ -s "${BUILD_ROOT}/${INITRAMFS_NAME}" ]]; then
		INITRAMFS_REEMITTED=true
		# Capture the listing first: `grep -q` in a pipeline short-circuits and,
		# with pipefail, the upstream SIGPIPE would mask a real match.
		listing="$(gzip -dc "${BUILD_ROOT}/${INITRAMFS_NAME}" 2>/dev/null | cpio -t 2>/dev/null || true)"
		if grep -q "lib/modules/${KVER}/modules.dep" <<<"${listing}"; then
			INITRAMFS_HAS_MODULES=true
		fi
	fi
fi

INITRAMFS_BYTES=0
[[ -f "${BUILD_ROOT}/${INITRAMFS_NAME}" ]] && INITRAMFS_BYTES="$(stat -c '%s' "${BUILD_ROOT}/${INITRAMFS_NAME}")"

STORAGE_COUNT="${#STORAGE_PRESENT[@]}"
NET_COUNT="${#NET_PRESENT[@]}"
RESOLVED_COUNT="${#RESOLVED_TOP[@]}"
STAGED_FILES=0
[[ -d "${MODDST}" ]] && STAGED_FILES="$(find "${MODDST}" -name '*.ko*' 2>/dev/null | wc -l | tr -d ' ')"

# --- provenance ---
{
	echo "# Generated by scripts/enable-live-initramfs-hardware.sh — do not commit"
	echo "schema_version: vyntrio-install-media-hardware-v1"
	echo "slice: 9.16"
	echo "generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
	echo "media_role: install"
	echo "kernel_version: ${KVER}"
	echo "modules_source: ${MODSRC}"
	echo "modules:"
	echo "  enabled: ${MODULES_ENABLED}"
	[[ -n "${MODULES_SKIP_REASON}" ]] && echo "  skip_reason: ${MODULES_SKIP_REASON}"
	echo "  resolved_top_level: ${RESOLVED_COUNT}"
	echo "  staged_ko_files: ${STAGED_FILES}"
	echo "  storage_drivers: ${STORAGE_COUNT} (${STORAGE_PRESENT[*]:-none})"
	echo "  network_drivers: ${NET_COUNT} (${NET_PRESENT[*]:-none})"
	if [[ "${#MISSING_NAMES[@]}" -gt 0 ]]; then
		echo "  not_present_on_host_kernel: ${MISSING_NAMES[*]}"
	fi
	echo "  modules_dep_regenerated: ${DEP_PRESENT}"
	echo "  load_list: distro/install-media/envelope/live_root/etc/vyntrio/modules.load"
	echo "network:"
	echo "  bring_up_beyond_loopback: true"
	echo "  method: busybox_udhcpc_on_ethernet_interfaces"
	echo "  lease_handler: live_root/usr/share/udhcpc/default.script"
	echo "  bring_up_script: live_root/usr/lib/vyntrio/hw-init.sh"
	echo "init_integration:"
	echo "  hw_init_invoked_from: live_root/init"
	echo "  order: mount_pseudo_fs -> load_modules -> network_up -> firstboot"
	echo "verified:"
	echo "  shell_exec_in_chroot: ${SHELL_EXEC}"
	echo "  modules_present_in_live_root: ${DEP_PRESENT}"
	echo "  sample_module_modinfo_ok: ${MODINFO_OK}"
	[[ -n "${SAMPLE_MODULE}" ]] && echo "  sample_module: ${SAMPLE_MODULE#${LR}/}"
	echo "  note: >-"
	echo "    Modules are staged + metadata regenerated + modinfo-validated. They are"
	echo "    NOT loaded here: modprobe would target the build-host kernel. Actual"
	echo "    load happens only inside a real live boot."
	echo "initramfs:"
	echo "  reemitted: ${INITRAMFS_REEMITTED}"
	echo "  path: distro/install-media/build/${INITRAMFS_NAME}"
	echo "  bytes: ${INITRAMFS_BYTES}"
	echo "  contains_kernel_modules: ${INITRAMFS_HAS_MODULES}"
	echo "honest_limits:"
	echo "  wired_into_image_initrd: false"
	echo "  boot_verified: false"
	echo "  boot_verified_reason: no qemu/VM harness on this host (see RUNTIME_BOOT.txt)"
	echo "  dashboard_reachable_on_boot: false"
	echo "supersedes_gap: kernel_modules_for_storage_and_network_in_initramfs (LIVE_ROOTFS.txt)"
	echo "missing_for_dashboard_reachable_boot:"
	echo "  - image_initrd_swapped_to_live_initramfs"
	echo "  - runtime_boot_verification_vm_or_hardware"
	echo "  - module_autoload_coverage_for_arbitrary_hardware"
	echo "missing_for_usb_creator:"
	echo "  - host_usb_writer_tool"
	echo "  - uefi_boot_support"
} >"${BUILD_ROOT}/HARDWARE_ENABLE.txt"

echo "install-media hardware: modules_enabled=${MODULES_ENABLED} storage=${STORAGE_COUNT} net=${NET_COUNT} staged_ko=${STAGED_FILES}"
echo "install-media hardware: shell_exec=${SHELL_EXEC} modinfo_ok=${MODINFO_OK} initramfs_has_modules=${INITRAMFS_HAS_MODULES}"
echo "install-media hardware: network bring-up beyond loopback wired (udhcpc) — not booted here"
echo "install-media hardware: wired_into_image_initrd=false dashboard_reachable_on_boot=false — see build/HARDWARE_ENABLE.txt" >&2
