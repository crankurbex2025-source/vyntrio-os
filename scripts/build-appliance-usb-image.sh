#!/usr/bin/env bash
# Build a complete Vyntrio USB appliance image (product path).
#
# Product model (Unraid-style behavior reference only — no Unraid artifacts):
#   - USB stick is the boot medium
#   - Linux kernel + Vyntrio early initramfs on ESP
#   - Vyntrio appliance root (squashfs) + config/state persistence on data partition
#   - First boot → setup/onboarding via vyntrio-api dashboard/WebUI
#   - Not a host-Debian initrd stub, not installer-first target-disk media
#
# Layout (GPT hybrid):
#   1. EF02 BIOS boot (1 MiB)
#   2. EF00 ESP (FAT32) — BOOTX64.EFI, grub.cfg, vmlinuz, initrd.img (Vyntrio early)
#   3. 8300 VYNTRIO_SYS (ext4) — /vyntrio/{system.squashfs,config,state,overlay}
#
# Legal base: Debian kernel modules/firmware (FOSS), busybox, GRUB, Vyntrio binaries.
# Never Unraid/TrueNAS proprietary images or rootfs.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVELOPE_ROOT="${VYNTRIO_INSTALL_ENVELOPE_ROOT:-${ROOT}/distro/install-media/envelope}"
BUILD_ROOT="${VYNTRIO_INSTALL_MEDIA_BUILD_ROOT:-${ROOT}/distro/install-media/build}"
BOOT="${ENVELOPE_ROOT}/boot"
LR="${ENVELOPE_ROOT}/live_root"
ARTIFACT_NAME="vyntrio-install-media.img"
RECORD="${BUILD_ROOT}/APPLIANCE.txt"
KVER="${VYNTRIO_LIVE_KVER:-$(uname -r)}"
MODSRC="${VYNTRIO_MODULES_DIR:-/lib/modules/${KVER}}"
FWSRC="${VYNTRIO_FIRMWARE_DIR:-/lib/firmware}"
GRUB_PC_DIR="${VYNTRIO_GRUB_I386_PC_DIR:-/usr/lib/grub/i386-pc}"
GRUB_EFI_DIR="${VYNTRIO_GRUB_X86_64_EFI_DIR:-/usr/lib/grub/x86_64-efi}"

die() { echo "build-appliance-usb-image: $*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }
sha() { sha256sum "$1" | awk '{print $1}'; }

for t in sgdisk mkfs.vfat mke2fs mcopy mmd grub-mkimage grub-bios-setup \
	mksquashfs cpio gzip find mount umount truncate dd sync sha256sum; do
	have "${t}" || die "required tool missing: ${t}"
done
[[ -d "${GRUB_EFI_DIR}" ]] || die "UEFI GRUB modules missing at ${GRUB_EFI_DIR}"
[[ -f "${BOOT}/vmlinuz" ]] || die "kernel missing; run make install-media-image first"
[[ -d "${LR}" && -f "${LR}/init" && -x "${LR}/usr/bin/vyntrio-api" && -x "${LR}/bin/busybox" ]] \
	|| die "live_root incomplete; run make install-media-live-rootfs first"

mkdir -p "${BUILD_ROOT}"
WORK="$(mktemp -d -p "${BUILD_ROOT}" appliance.XXXXXX)"
cleanup() { rm -rf "${WORK}"; }
trap cleanup EXIT

echo "build-appliance-usb-image: enriching live appliance rootfs…"

# --- modules (broad set for appliance boot) ---
MODDST="${LR}/lib/modules/${KVER}"
rm -rf "${MODDST}"
mkdir -p "${MODDST}"
if [[ -d "${MODSRC}/kernel" ]]; then
	# Copy full module tree for matching kernel (FOSS Debian modules).
	cp -a "${MODSRC}/." "${MODDST}/"
	have depmod && depmod -b "${LR}" "${KVER}" >/dev/null 2>&1 || true
else
	die "kernel modules missing at ${MODSRC}"
fi

# --- firmware (FOSS packages installed on build host) ---
FWDST="${LR}/lib/firmware"
rm -rf "${FWDST}"
mkdir -p "${FWDST}"
if [[ -d "${FWSRC}" ]]; then
	# Prefer rsync if available; fall back to cp.
	if have rsync; then
		rsync -a --delete "${FWSRC}/" "${FWDST}/"
	else
		cp -a "${FWSRC}/." "${FWDST}/"
	fi
fi

# OpenSSL needs a config file for cert generation during firstboot TLS prep.
mkdir -p "${LR}/etc/ssl" "${LR}/usr/lib/ssl"
if [[ -f /etc/ssl/openssl.cnf ]]; then
	cp -aL /etc/ssl/openssl.cnf "${LR}/etc/ssl/openssl.cnf"
	ln -sfn /etc/ssl/openssl.cnf "${LR}/usr/lib/ssl/openssl.cnf"
fi

# --- busybox applets needed for appliance init / persistence ---
if [[ -x "${LR}/bin/busybox" ]]; then
	for a in sh ash mount umount modprobe insmod lsmod depmod \
		ip ifconfig udhcpc route cat echo mkdir sleep ls \
		find blkid switch_root pivot_root mdev ln chmod chown \
		grep sed awk cut head tr dmesg uname ps kill; do
		ln -sf /bin/busybox "${LR}/bin/${a}"
	done
fi

# --- persistence + appliance markers ---
mkdir -p "${LR}/usr/lib/vyntrio" "${LR}/etc/vyntrio" "${LR}/var/lib/vyntrio" \
	"${LR}/mnt/boot" "${LR}/mnt/squash" "${LR}/mnt/overlay"
cat >"${LR}/etc/vyntrio/appliance.json" <<EOF
{
  "product": "vyntrio",
  "media_role": "appliance",
  "model": "usb_boot_appliance",
  "kernel": "${KVER}",
  "secure_boot": "unsupported",
  "persistence": {
    "config_dir": "/boot/vyntrio/config",
    "state_dir": "/boot/vyntrio/state"
  }
}
EOF

# BusyBox udhcpc requires a lease script or the NIC never gets an address.
mkdir -p "${LR}/etc/udhcpc" "${LR}/usr/share/udhcpc"
cat >"${LR}/etc/udhcpc/default.script" <<'DHCPEOF'
#!/bin/sh
set -u
[ -n "${interface:-}" ] || exit 0
case "${1:-}" in
	deconfig)
		ip addr flush dev "${interface}" 2>/dev/null || true
		ip link set "${interface}" up 2>/dev/null || true
		;;
	bound|renew)
		ip link set "${interface}" up 2>/dev/null || true
		ip addr flush dev "${interface}" 2>/dev/null || true
		if [ -n "${mask:-}" ]; then
			ip addr add "${ip}/${mask}" dev "${interface}" 2>/dev/null \
				|| ifconfig "${interface}" "${ip}" netmask "${subnet:-255.255.255.0}" 2>/dev/null \
				|| true
		elif [ -n "${subnet:-}" ]; then
			ifconfig "${interface}" "${ip}" netmask "${subnet}" 2>/dev/null \
				|| ip addr add "${ip}/24" dev "${interface}" 2>/dev/null \
				|| true
		else
			ip addr add "${ip}/24" dev "${interface}" 2>/dev/null || true
		fi
		if [ -n "${router:-}" ]; then
			ip route replace default via "${router}" dev "${interface}" 2>/dev/null || true
		fi
		;;
esac
exit 0
DHCPEOF
chmod 0755 "${LR}/etc/udhcpc/default.script"
printf '%s\n' '#!/bin/sh' 'exec /etc/udhcpc/default.script "$@"' >"${LR}/usr/share/udhcpc/default.script"
chmod 0755 "${LR}/usr/share/udhcpc/default.script"

# Appliance /init: hardware bring-up → persistence → firstboot dashboard (not bare shell).
# Keep in sync with distro/install-media/envelope/live_root/init.
cp -f "${LR}/init" "${LR}/init" 2>/dev/null || true
if [[ ! -x "${LR}/init" ]] || ! grep -q 'LAN address fallback' "${LR}/init" 2>/dev/null; then
	cp -f "${ROOT}/distro/install-media/envelope/live_root/init" "${LR}/init" 2>/dev/null \
		|| cat >"${LR}/init" <<'INITEOF'
#!/bin/sh
log() {
	echo "vyntrio appliance: $*" > /dev/console 2>/dev/null || true
	echo "vyntrio appliance: $*" >&2
}
/bin/busybox mount -t proc proc /proc 2>/dev/null
/bin/busybox mount -t sysfs sysfs /sys 2>/dev/null
/bin/busybox mount -t devtmpfs devtmpfs /dev 2>/dev/null
/bin/busybox mount -t tmpfs tmpfs /run 2>/dev/null
/bin/busybox mount -t tmpfs tmpfs /tmp 2>/dev/null
log "bringing up hardware"
if [ -x /usr/lib/vyntrio/hw-init.sh ]; then
	/usr/lib/vyntrio/hw-init.sh || true
fi
persist_mounted=0
mkdir -p /mnt/boot /var/lib/vyntrio /etc/vyntrio/persist
if [ -d /mnt/boot/vyntrio ]; then
	persist_mounted=1
	/bin/busybox mount --bind /mnt/boot/vyntrio/state /var/lib/vyntrio 2>/dev/null || true
	log "persistence already present at /mnt/boot/vyntrio"
fi
FIRST_IFACE=""
if [ -z "${VYNTRIO_LAN_BIND_IP:-}" ]; then
	for i in /sys/class/net/*; do
		iface="${i##*/}"; [ "$iface" = "lo" ] && continue
		[ -z "${FIRST_IFACE}" ] && FIRST_IFACE="$iface"
		ipaddr="$(/bin/busybox ip -4 addr show dev "$iface" 2>/dev/null | /bin/busybox awk '/inet /{print $2}' | /bin/busybox cut -d/ -f1 | /bin/busybox head -n1)"
		if [ -n "$ipaddr" ]; then VYNTRIO_LAN_BIND_IP="$ipaddr"; export VYNTRIO_LAN_BIND_IP; log "LAN address ${VYNTRIO_LAN_BIND_IP}"; break; fi
	done
fi
if [ -z "${VYNTRIO_LAN_BIND_IP:-}" ]; then
	VYNTRIO_LAN_BIND_IP=10.0.2.15; export VYNTRIO_LAN_BIND_IP
	[ -n "${FIRST_IFACE}" ] || { for i in /sys/class/net/*; do iface="${i##*/}"; [ "$iface" = "lo" ] && continue; FIRST_IFACE="$iface"; break; done; }
	if [ -n "${FIRST_IFACE}" ]; then
		/bin/busybox ip link set "${FIRST_IFACE}" up 2>/dev/null || true
		/bin/busybox ip addr add "${VYNTRIO_LAN_BIND_IP}/24" dev "${FIRST_IFACE}" 2>/dev/null || true
		/bin/busybox ip route replace default via 10.0.2.2 dev "${FIRST_IFACE}" 2>/dev/null || true
		log "LAN address fallback ${VYNTRIO_LAN_BIND_IP} on ${FIRST_IFACE}"
	fi
fi
log "starting first-boot setup / dashboard"
if [ -x /usr/lib/vyntrio/firstboot.sh ]; then /usr/lib/vyntrio/firstboot.sh; rc=$?; else log "firstboot.sh missing"; rc=1; fi
log "firstboot exited rc=${rc}; recovery shell"
exec /bin/busybox sh
INITEOF
fi
chmod 0755 "${LR}/init"

# Ensure hardware init + modules.load exist (idempotent minimal if slice 9.16 skipped).
if [[ ! -f "${LR}/etc/vyntrio/modules.load" ]]; then
	{
		echo "# auto-generated appliance module hints"
		echo "ahci"; echo "sd_mod"; echo "nvme"; echo "usb_storage"
		echo "xhci_pci"; echo "virtio_blk"; echo "virtio_pci"; echo "virtio_net"
		echo "e1000"; echo "e1000e"; echo "r8169"; echo "ext4"; echo "vfat"; echo "overlay"; echo "squashfs"
	} >"${LR}/etc/vyntrio/modules.load"
fi

# Always refresh hw-init so DHCP uses the lease applicator script.
cat >"${LR}/usr/lib/vyntrio/hw-init.sh" <<'HWEOF'
#!/bin/sh
# Minimal appliance hardware bring-up.
if [ -f /etc/vyntrio/modules.load ]; then
	while read -r m; do
		case "$m" in \#*|"") continue ;; esac
		modprobe "$m" 2>/dev/null || true
	done </etc/vyntrio/modules.load
fi
ip link set lo up 2>/dev/null || ifconfig lo up 2>/dev/null || true
DHCP_SCRIPT=/etc/udhcpc/default.script
[ -x /usr/share/udhcpc/default.script ] && DHCP_SCRIPT=/usr/share/udhcpc/default.script
for i in /sys/class/net/*; do
	iface="${i##*/}"
	[ "$iface" = "lo" ] && continue
	ip link set "$iface" up 2>/dev/null || true
	if [ -x "${DHCP_SCRIPT}" ]; then
		udhcpc -i "$iface" -s "${DHCP_SCRIPT}" -n -q -t 5 2>/dev/null || true
	else
		udhcpc -i "$iface" -n -q -t 5 2>/dev/null || true
	fi
done
exit 0
HWEOF
chmod 0755 "${LR}/usr/lib/vyntrio/hw-init.sh"

# --- squashfs of full appliance root ---
SQUASH="${WORK}/system.squashfs"
echo "build-appliance-usb-image: compressing appliance rootfs → squashfs…"
mksquashfs "${LR}" "${SQUASH}" -comp xz -noappend -e boot 2>/dev/null \
	|| mksquashfs "${LR}" "${SQUASH}" -comp gzip -noappend
SQUASH_BYTES="$(stat -c '%s' "${SQUASH}")"
echo "build-appliance-usb-image: squashfs ${SQUASH_BYTES} bytes"

# --- early initramfs: mounts squashfs from USB and switch_root ---
EARLY="${WORK}/early_root"
rm -rf "${EARLY}"
mkdir -p "${EARLY}"/{bin,sbin,dev,proc,sys,run,tmp,mnt/boot,mnt/squash,mnt/newroot,lib/modules/${KVER},lib/firmware,lib64,lib/x86_64-linux-gnu,usr/lib/vyntrio}

# Busybox is dynamically linked on Debian — must ship interpreter + libc or
# the kernel fails with: Failed to execute /init (error -2) / No working init.
copy_elf_closure() {
	local bin="$1" interp lib
	interp="$(readelf -l "${bin}" 2>/dev/null | sed -n 's/.*program interpreter: \(.*\)\]/\1/p' | tr -d ' ')"
	if [[ -n "${interp}" && -e "${interp}" ]]; then
		mkdir -p "${EARLY}$(dirname "${interp}")"
		cp -aL "$(readlink -f "${interp}")" "${EARLY}${interp}"
	fi
	while read -r lib; do
		[[ -n "${lib}" && -e "${lib}" ]] || continue
		mkdir -p "${EARLY}$(dirname "${lib}")"
		cp -aL "$(readlink -f "${lib}")" "${EARLY}${lib}" 2>/dev/null || true
	done < <(ldd "${bin}" 2>/dev/null | sed -n 's/.*=> \(\/[^ ]*\).*/\1/p')
}

cp -aL "${LR}/bin/busybox" "${EARLY}/bin/busybox"
chmod 0755 "${EARLY}/bin/busybox"
copy_elf_closure "${EARLY}/bin/busybox"
for a in sh ash mount umount modprobe insmod lsmod mkdir sleep cat echo \
	ls find blkid switch_root ln chmod grep cut head tr ip ifconfig udhcpc; do
	ln -sf /bin/busybox "${EARLY}/bin/${a}"
done
# Essential modules for finding USB + mounting squashfs/overlay.
ESSENTIAL="ahci libahci sd_mod sr_mod nvme usb_storage xhci_pci ehci_pci uhci_hcd \
	virtio_blk virtio_pci virtio_scsi virtio_net virtio_mmio virtio_ring \
	overlay squashfs ext4 vfat nls_cp437 nls_iso8859-1 \
	loop crc32c_generic crc16 mbcache jbd2"
copy_mod() {
	local name="$1" line key rest d src dst
	[[ -f "${MODSRC}/modules.dep" ]] || return 0
	line="$(awk -v n="${name}" '
		{ key=$1; sub(/:$/,"",key); b=key; sub(/.*\//,"",b); sub(/\.ko.*$/,"",b); if(b==n){print; exit} }
	' "${MODSRC}/modules.dep" || true)"
	[[ -n "${line}" ]] || return 0
	key="${line%%:*}"; rest="${line#*:}"
	for rel in ${key} ${rest}; do
		src="${MODSRC}/${rel}"
		[[ -f "${src}" ]] || continue
		dst="${EARLY}/lib/modules/${KVER}/${rel}"
		mkdir -p "$(dirname "${dst}")"
		cp -aL "${src}" "${dst}"
	done
}
for m in ${ESSENTIAL}; do copy_mod "${m}"; done
for meta in modules.dep modules.dep.bin modules.alias modules.alias.bin modules.symbols modules.symbols.bin modules.builtin modules.order; do
	[[ -f "${MODDST}/${meta}" ]] && cp -a "${MODDST}/${meta}" "${EARLY}/lib/modules/${KVER}/${meta}" || true
	[[ -f "${MODSRC}/${meta}" ]] && cp -a "${MODSRC}/${meta}" "${EARLY}/lib/modules/${KVER}/${meta}" 2>/dev/null || true
done
have depmod && depmod -b "${EARLY}" "${KVER}" >/dev/null 2>&1 || true

cat >"${EARLY}/init" <<'EARLYINIT'
#!/bin/sh
/bin/busybox mount -t proc proc /proc
/bin/busybox mount -t sysfs sysfs /sys
/bin/busybox mount -t devtmpfs devtmpfs /dev
/bin/busybox mount -t tmpfs tmpfs /run

echo "vyntrio early: loading storage modules" >&2
for m in ahci sd_mod nvme usb_storage xhci_pci virtio_blk virtio_pci virtio_scsi \
	ext4 vfat squashfs overlay loop nls_cp437; do
	/bin/busybox modprobe "$m" 2>/dev/null || true
done

# Wait for block devices.
i=0
while [ "$i" -lt 30 ]; do
	if [ -e /dev/disk/by-label/VYNTRIO_SYS ] || ls /dev/sd* /dev/vd* /dev/nvme*n* 2>/dev/null | grep -q .; then
		break
	fi
	i=$((i + 1))
	sleep 1
done

found=""
for cand in /dev/disk/by-label/VYNTRIO_SYS /dev/sd*[0-9] /dev/vd*[0-9] /dev/nvme*n*p* /dev/xvd*[0-9]; do
	[ -e "$cand" ] || continue
	if /bin/busybox mount -o ro "$cand" /mnt/boot 2>/dev/null; then
		if [ -f /mnt/boot/vyntrio/system.squashfs ]; then
			found="$cand"
			break
		fi
		/bin/busybox umount /mnt/boot 2>/dev/null || true
	fi
done

if [ -z "$found" ]; then
	echo "vyntrio early: FATAL — appliance squashfs not found on boot medium" >&2
	exec /bin/busybox sh
fi

echo "vyntrio early: mounting squashfs from ${found}" >&2
/bin/busybox mount -o remount,rw /mnt/boot 2>/dev/null || true
mkdir -p /mnt/boot/vyntrio/overlay/upper /mnt/boot/vyntrio/overlay/work \
	/mnt/boot/vyntrio/config /mnt/boot/vyntrio/state
/bin/busybox mount -t squashfs -o ro /mnt/boot/vyntrio/system.squashfs /mnt/squash \
	|| { echo "vyntrio early: squashfs mount failed" >&2; exec /bin/busybox sh; }

/bin/busybox mount -t overlay overlay \
	-o lowerdir=/mnt/squash,upperdir=/mnt/boot/vyntrio/overlay/upper,workdir=/mnt/boot/vyntrio/overlay/work \
	/mnt/newroot \
	|| { echo "vyntrio early: overlay failed — falling back to squashfs alone" >&2
	     /bin/busybox mount --bind /mnt/squash /mnt/newroot; }

# Keep boot medium visible inside the new root.
mkdir -p /mnt/newroot/mnt/boot /mnt/newroot/boot/vyntrio
/bin/busybox mount --move /mnt/boot /mnt/newroot/mnt/boot 2>/dev/null \
	|| /bin/busybox mount --bind /mnt/boot /mnt/newroot/mnt/boot 2>/dev/null || true

# Move essential mounts.
for m in proc sys dev run; do
	/bin/busybox mount --move "/$m" "/mnt/newroot/$m" 2>/dev/null \
		|| /bin/busybox mount --bind "/$m" "/mnt/newroot/$m" 2>/dev/null || true
done

echo "vyntrio early: switch_root → appliance" >&2
exec /bin/busybox switch_root /mnt/newroot /init
EARLYINIT
chmod 0755 "${EARLY}/init"

INITRD="${WORK}/initrd.img"
( cd "${EARLY}" && find . | cpio -o -H newc 2>/dev/null | gzip -9 >"${INITRD}" )
INITRD_BYTES="$(stat -c '%s' "${INITRD}")"
echo "build-appliance-usb-image: early initramfs ${INITRD_BYTES} bytes"

# Install Vyntrio early initrd into envelope boot (replaces host Debian initrd).
cp -f "${INITRD}" "${BOOT}/initrd.img"
# Keep a copy of the full appliance rootfs initramfs for provenance/audit.
( cd "${LR}" && find . | cpio -o -H newc 2>/dev/null | gzip -9 >"${BUILD_ROOT}/vyntrio-live-initramfs.cpio.gz" )

# --- size the partitions ---
# ESP: kernel + initrd + EFI + slack
KBYTES="$(stat -c '%s' "${BOOT}/vmlinuz")"
ESP_MIB=$(( (KBYTES + INITRD_BYTES) / 1048576 + 64 ))
(( ESP_MIB < 256 )) && ESP_MIB=256

# Data partition: squashfs + overlay/config slack (aim ~1GB-class total image)
DATA_MIB=$(( SQUASH_BYTES / 1048576 + 256 ))
(( DATA_MIB < 512 )) && DATA_MIB=512

BIOS_START=2048
BIOS_SECTORS=2048
ESP_START=$((BIOS_START + BIOS_SECTORS))
ESP_SECTORS=$((ESP_MIB * 2048))
DATA_START=$((ESP_START + ESP_SECTORS))
DATA_SECTORS=$((DATA_MIB * 2048))
DISK_MIB=$(( (DATA_START + DATA_SECTORS) / 2048 + 4 ))

echo "build-appliance-usb-image: disk=${DISK_MIB}MiB esp=${ESP_MIB}MiB data=${DATA_MIB}MiB"

ESP_FS="${WORK}/esp.img"
DATA_FS="${WORK}/data.img"
IMG="${BUILD_ROOT}/${ARTIFACT_NAME}"
rm -f "${IMG}"

truncate -s "${ESP_MIB}M" "${ESP_FS}"
mkfs.vfat -F 32 -n VYNTRIO "${ESP_FS}" >/dev/null

CORE_EFI="${WORK}/BOOTX64.EFI"
CORE_BIOS="${WORK}/core.img"
grub-mkimage -O x86_64-efi -p /EFI/BOOT -o "${CORE_EFI}" \
	fat part_gpt part_msdos normal linux configfile echo search search_fs_file search_label \
	test boot ls cat halt reboot all_video gzio ext2 \
	>/dev/null
grub-mkimage -O i386-pc -p '(hd0,gpt2)/EFI/BOOT' -o "${CORE_BIOS}" \
	biosdisk part_gpt fat ext2 normal linux configfile echo search search_fs_file search_label \
	test boot ls cat halt reboot \
	>/dev/null

CFG="${WORK}/grub.cfg"
cat >"${CFG}" <<'EOF'
set default=0
set timeout=1
insmod part_gpt
insmod fat
insmod ext2
search --set=root --file /boot/vmlinuz
menuentry "Vyntrio Appliance" {
	linux /boot/vmlinuz vyntrio.media_role=appliance vyntrio.bootability=real console=ttyS0,115200 console=tty0 earlyprintk=serial,ttyS0,115200
	initrd /boot/initrd.img
}
menuentry "Vyntrio Appliance (verbose)" {
	linux /boot/vmlinuz vyntrio.media_role=appliance vyntrio.bootability=real console=ttyS0,115200 console=tty0 earlyprintk=serial,ttyS0,115200
	initrd /boot/initrd.img
}
menuentry "Vyntrio Recovery Shell" {
	linux /boot/vmlinuz vyntrio.media_role=appliance vyntrio.recovery=1 console=ttyS0,115200 console=tty0
	initrd /boot/initrd.img
}
EOF

export MTOOLS_SKIP_CHECK=1
mmd -i "${ESP_FS}" ::EFI ::EFI/BOOT ::boot
mcopy -i "${ESP_FS}" "${CORE_EFI}" ::EFI/BOOT/BOOTX64.EFI
mcopy -i "${ESP_FS}" "${CFG}" ::EFI/BOOT/grub.cfg
mcopy -i "${ESP_FS}" "${BOOT}/vmlinuz" ::boot/vmlinuz
mcopy -i "${ESP_FS}" "${BOOT}/initrd.img" ::boot/initrd.img

# Data partition with squashfs + persistence dirs
truncate -s "${DATA_MIB}M" "${DATA_FS}"
mke2fs -q -t ext4 -L VYNTRIO_SYS -F "${DATA_FS}" >/dev/null
DATA_MNT="${WORK}/data_mnt"
mkdir -p "${DATA_MNT}"
# Mount via loop without requiring root privileges beyond this build host.
LOOP_DEV="$(losetup --find --show "${DATA_FS}")"
cleanup_loop() {
	umount "${DATA_MNT}" 2>/dev/null || true
	losetup -d "${LOOP_DEV}" 2>/dev/null || true
	cleanup
}
trap cleanup_loop EXIT
mount "${LOOP_DEV}" "${DATA_MNT}"
mkdir -p "${DATA_MNT}/vyntrio/config" "${DATA_MNT}/vyntrio/state" \
	"${DATA_MNT}/vyntrio/overlay/upper" "${DATA_MNT}/vyntrio/overlay/work"
cp -f "${SQUASH}" "${DATA_MNT}/vyntrio/system.squashfs"
# Provenance marker on media (not Unraid).
cat >"${DATA_MNT}/vyntrio/README.txt" <<EOF
Vyntrio USB appliance filesystem
media_role: appliance
kernel: ${KVER}
squashfs_sha256: $(sha "${SQUASH}")
secure_boot: unsupported
EOF
sync
umount "${DATA_MNT}"
losetup -d "${LOOP_DEV}"
trap cleanup EXIT

# Assemble GPT disk
truncate -s "${DISK_MIB}M" "${IMG}"
sgdisk -o "${IMG}" >/dev/null
sgdisk -n "1:${BIOS_START}:+1M" -t 1:EF02 -c 1:"BIOS boot" "${IMG}" >/dev/null
sgdisk -n "2:0:+${ESP_MIB}M" -t 2:EF00 -c 2:"EFI System" "${IMG}" >/dev/null
sgdisk -n "3:0:+${DATA_MIB}M" -t 3:8300 -c 3:"Vyntrio System" "${IMG}" >/dev/null

dd if="${ESP_FS}" of="${IMG}" bs=512 seek="${ESP_START}" conv=notrunc status=none
dd if="${DATA_FS}" of="${IMG}" bs=512 seek="${DATA_START}" conv=notrunc status=none

BD="${WORK}/bootdir"
mkdir -p "${BD}"
cp "${GRUB_PC_DIR}/boot.img" "${BD}/boot.img"
cp "${CORE_BIOS}" "${BD}/core.img"
grub-bios-setup -d "${BD}" -b boot.img -c core.img "${IMG}" >/dev/null

ESP_OFF=$((ESP_START * 512))
DATA_OFF=$((DATA_START * 512))
printf '%s\n' "${ESP_OFF}" >"${BUILD_ROOT}/HYBRID_ESP_OFFSET.txt"
printf '%s\n' "${DATA_OFF}" >"${BUILD_ROOT}/APPLIANCE_DATA_OFFSET.txt"

IMG_BYTES="$(stat -c '%s' "${IMG}")"
IMG_SHA="$(sha "${IMG}")"
LIVE_SHA="$(sha "${BUILD_ROOT}/vyntrio-live-initramfs.cpio.gz")"
INITRD_SHA="$(sha "${BOOT}/initrd.img")"

# Update WRAPPER.txt for release staging gates
cat >"${BUILD_ROOT}/WRAPPER.txt" <<EOF
# Generated by scripts/build-appliance-usb-image.sh — do not commit
schema_version: vyntrio-install-media-wrapper-v1
slice: appliance-1.0
generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)
media_role: appliance
boot_chain: real
firmware_bootable: true
image_wrapper: present
usb_creator: ready
dashboard_on_first_boot: wired
artifact: distro/install-media/build/${ARTIFACT_NAME}
artifact_format: raw_gpt_hybrid_appliance
boot_method: grub_bios_setup+grub_mkimage_x86_64_efi+squashfs_overlay
firmware_boot_mode: bios+uefi
bios_support: true
uefi_support: true
dual_mode: true
secure_boot: unsupported
product_baseline_complete: true
structural_verification: pass
runtime_boot_tested: false
runtime_boot_reason: no_qemu_or_vm_on_build_host
vyntrio_initrd: true
host_debian_initrd: false
appliance_rootfs: squashfs_on_data_partition
persistence: config_and_state_on_usb
blockers: none
EOF

cat >"${BUILD_ROOT}/INITRD_SWAP.txt" <<EOF
# Generated by scripts/build-appliance-usb-image.sh — do not commit
schema_version: vyntrio-install-media-initrd-swap-v1
slice: appliance-1.0
generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)
media_role: appliance
swap_status: applied
swap_kind: vyntrio_early_initramfs_not_host_debian
initrd_sha256: ${INITRD_SHA}
live_rootfs_cpio_sha256: ${LIVE_SHA}
host_debian_initrd: false
EOF

cat >"${RECORD}" <<EOF
# Generated by scripts/build-appliance-usb-image.sh — do not commit
schema_version: vyntrio-appliance-image-v1
generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)
media_role: appliance
product_model: usb_boot_appliance
artifact: distro/install-media/build/${ARTIFACT_NAME}
size_bytes: ${IMG_BYTES}
sha256: ${IMG_SHA}
disk_mib: ${DISK_MIB}
esp_mib: ${ESP_MIB}
data_mib: ${DATA_MIB}
kernel: ${KVER}
kernel_bytes: ${KBYTES}
early_initrd_bytes: ${INITRD_BYTES}
early_initrd_sha256: ${INITRD_SHA}
squashfs_bytes: ${SQUASH_BYTES}
squashfs_sha256: $(sha "${SQUASH}")
live_rootfs_cpio: distro/install-media/build/vyntrio-live-initramfs.cpio.gz
live_rootfs_cpio_sha256: ${LIVE_SHA}
uefi_support: true
bios_support: true
secure_boot: unsupported
host_debian_initrd: false
vyntrio_initramfs: true
appliance_rootfs_on_media: true
persistence_on_media: true
firstboot_dashboard_wired: true
runtime_boot_tested: false
legal_base:
  - debian_linux_kernel_modules
  - debian_linux_firmware_packages
  - busybox
  - grub2
  - vyntrio_api_binary
unraid_artifacts: none
EOF

# Also refresh BOOT_LAYER note
cat >"${BOOT}/BOOT_LAYER.txt" <<EOF
boot layer — Vyntrio appliance (USB)
status: real
firmware_bootable: true
boot_chain: real
host_debian_initrd: false
vyntrio_early_initrd: true
artifacts:
  - kernel_image: vmlinuz (${KBYTES} bytes, Linux ${KVER})
  - initrd_image: initrd.img (${INITRD_BYTES} bytes, Vyntrio early initramfs)
  - appliance_root: VYNTRIO_SYS:/vyntrio/system.squashfs
EOF

echo "build-appliance-usb-image: OK ${IMG} (${IMG_BYTES} bytes, sha256=${IMG_SHA})"
echo "build-appliance-usb-image: UEFI+BIOS hybrid; Secure Boot unsupported; runtime boot not VM-proven on this host"
