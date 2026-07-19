#!/usr/bin/env bash
# Build install-media boot chain + image candidate (Block 9 / Slice 9.12).
#
# Upgrades the Slice 9.11 bootability foundation from pure stubs toward a real
# boot chain: it replaces the boot/*.stub placeholders with a real Linux kernel,
# a real initrd, a real GRUB core image, and a real grub.cfg when those inputs
# are available on the build host, then emits the closest-to-bootable image the
# host tooling can produce.
#
# Honesty:
#   - When a real kernel/initrd/grub-mkimage are present -> boot_chain: real.
#   - When they are absent -> boot_chain: stub (foundation carried forward).
#   - A firmware-bootable ISO/raw image is emitted ONLY when an ISO9660/El Torito
#     writer (grub-mkrescue + xorriso/xorrisofs/genisoimage) is present. Otherwise
#     the real boot tree is emitted as a tar and firmware_bootable stays false.
#   - No target-disk writes. No installer/apply behavior. USB creator deferred.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVELOPE_ROOT="${VYNTRIO_INSTALL_ENVELOPE_ROOT:-${ROOT}/distro/install-media/envelope}"
BUILD_ROOT="${VYNTRIO_INSTALL_MEDIA_BUILD_ROOT:-${ROOT}/distro/install-media/build}"

ISO_IMAGE_NAME="vyntrio-install-media.iso"
REAL_TAR_NAME="vyntrio-install-media-REAL-BOOTCHAIN-NO-ISO.tar"
STUB_TAR_NAME="vyntrio-install-media-NOT-BOOTABLE.stub.tar"

# Real boot-chain inputs (overridable for CI / hermetic builds).
KERNEL_SRC="${VYNTRIO_HOST_KERNEL:-}"
INITRD_SRC="${VYNTRIO_HOST_INITRD:-}"
GRUB_PC_DIR="${VYNTRIO_GRUB_I386_PC_DIR:-/usr/lib/grub/i386-pc}"

readonly -a PAYLOAD_FILES=(
	"usr/bin/vyntrio-api"
	"usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service"
	"usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf"
	"etc/vyntrio/config.toml"
)

# GRUB modules embedded in the core image and staged as real boot ingredients.
readonly -a GRUB_MODULES=(
	biosdisk part_msdos part_gpt fat iso9660 normal linux configfile
	echo search search_fs_uuid test boot ls cat halt reboot
)

die() { echo "build-install-media-image: $*" >&2; exit 1; }

# --- precondition: Slice 9.11 bootability foundation must exist ---
[[ -d "${ENVELOPE_ROOT}/boot" && -d "${ENVELOPE_ROOT}/live_root" && -d "${ENVELOPE_ROOT}/payload" ]] \
	|| die "bootability foundation missing; run 'make install-media-bootability' first"

for rel in "${PAYLOAD_FILES[@]}"; do
	[[ -f "${ENVELOPE_ROOT}/payload/${rel}" ]] || die "missing payload file: ${rel} (run 'make install-media-bootability')"
done

mapfile -t payload_files < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
[[ "${#payload_files[@]}" -eq "${#PAYLOAD_FILES[@]}" ]] || die "unexpected payload file count"

# --- capability detection ---
autodetect_kernel() {
	local k
	for k in $(ls -1 /boot/vmlinuz-* 2>/dev/null | LC_ALL=C sort -V | tac); do
		[[ -f "${k}" ]] && { echo "${k}"; return 0; }
	done
	return 1
}
autodetect_initrd() {
	local ver="$1" i
	# Prefer an initrd matching the chosen kernel version, else newest.
	if [[ -n "${ver}" && -f "/boot/initrd.img-${ver}" ]]; then
		echo "/boot/initrd.img-${ver}"; return 0
	fi
	for i in $(ls -1 /boot/initrd.img-* 2>/dev/null | LC_ALL=C sort -V | tac); do
		[[ -f "${i}" ]] && { echo "${i}"; return 0; }
	done
	return 1
}

initrd_magic_ok() {
	local f="$1" hex
	hex="$(head -c 6 "${f}" | od -An -tx1 | tr -d ' \n')"
	case "${hex}" in
		1f8b*) return 0 ;;            # gzip
		fd377a5859*) return 0 ;;      # xz
		28b52ffd*) return 0 ;;        # zstd
		04224d18*) return 0 ;;        # lz4
		894c5a*) return 0 ;;          # lzop
		425a68*) return 0 ;;          # bzip2
		303730373031) return 0 ;;     # newc cpio ("070701", uncompressed initramfs)
		*) return 1 ;;
	esac
}

BOOT_CHAIN="stub"
KERNEL_VERIFIED=false
INITRD_VERIFIED=false
GRUB_CORE_BUILT=false

if command -v grub-mkimage >/dev/null 2>&1 && command -v grub-file >/dev/null 2>&1; then
	[[ -z "${KERNEL_SRC}" ]] && KERNEL_SRC="$(autodetect_kernel || true)"
	kver=""
	[[ -n "${KERNEL_SRC}" ]] && kver="$(basename "${KERNEL_SRC}" | sed -e 's/^vmlinuz-//')"
	[[ -z "${INITRD_SRC}" ]] && INITRD_SRC="$(autodetect_initrd "${kver}" || true)"

	if [[ -n "${KERNEL_SRC}" && -f "${KERNEL_SRC}" ]] \
		&& grub-file --is-x86-linux "${KERNEL_SRC}" 2>/dev/null; then
		KERNEL_VERIFIED=true
	fi
	if [[ -n "${INITRD_SRC}" && -f "${INITRD_SRC}" ]] && initrd_magic_ok "${INITRD_SRC}"; then
		INITRD_VERIFIED=true
	fi
	if [[ "${KERNEL_VERIFIED}" == true && "${INITRD_VERIFIED}" == true && -d "${GRUB_PC_DIR}" ]]; then
		BOOT_CHAIN="real"
	fi
fi

# --- ISO writer capability (needed for a firmware-bootable image) ---
ISO_WRITER=""
for w in xorriso xorrisofs genisoimage mkisofs; do
	if command -v "${w}" >/dev/null 2>&1; then ISO_WRITER="${w}"; break; fi
done
CAN_EMIT_ISO=false
if [[ "${BOOT_CHAIN}" == "real" ]] && command -v grub-mkrescue >/dev/null 2>&1 && [[ -n "${ISO_WRITER}" ]]; then
	CAN_EMIT_ISO=true
fi

# --- boot layer: populate real assets (or keep stubs) ---
BOOT="${ENVELOPE_ROOT}/boot"
if [[ "${BOOT_CHAIN}" == "real" ]]; then
	# Remove Slice 9.11 stub placeholders; replace with real artifacts.
	rm -f "${BOOT}/vmlinuz.stub" "${BOOT}/initrd.img.stub"
	rm -f "${BOOT}/loader/entries/vyntrio-install.conf.stub"
	rmdir "${BOOT}/loader/entries" "${BOOT}/loader" 2>/dev/null || true

	install -m 0644 "${KERNEL_SRC}" "${BOOT}/vmlinuz"
	install -m 0644 "${INITRD_SRC}" "${BOOT}/initrd.img"

	mkdir -p "${BOOT}/grub/i386-pc"
	cat >"${BOOT}/grub/grub.cfg" <<'EOF'
# Vyntrio install-media boot configuration (Slice 9.12 — real boot chain).
# Real kernel + initrd. Firmware bootability requires wrapping this tree in an
# ISO9660/El Torito or raw image (deferred: see IMAGE.txt blockers).
set default=0
set timeout=5

menuentry "Vyntrio Install (live)" {
	echo "Loading Vyntrio install medium..."
	linux /boot/vmlinuz boot=live vyntrio.media_role=install vyntrio.bootability=real quiet
	initrd /boot/initrd.img
}

menuentry "Vyntrio Install (verbose)" {
	linux /boot/vmlinuz boot=live vyntrio.media_role=install vyntrio.bootability=real
	initrd /boot/initrd.img
}
EOF

	# Real GRUB BIOS core image (genuine bootloader binary, not a text stub).
	if grub-mkimage -O i386-pc -p /boot/grub \
		-o "${BOOT}/grub/i386-pc/core.img" "${GRUB_MODULES[@]}" 2>/dev/null; then
		GRUB_CORE_BUILT=true
	else
		die "grub-mkimage failed to build core.img"
	fi

	# Stage the real BIOS boot building blocks + referenced modules.
	for f in boot.img cdboot.img; do
		[[ -f "${GRUB_PC_DIR}/${f}" ]] && install -m 0644 "${GRUB_PC_DIR}/${f}" "${BOOT}/grub/i386-pc/${f}"
	done
	for m in "${GRUB_MODULES[@]}"; do
		[[ -f "${GRUB_PC_DIR}/${m}.mod" ]] && install -m 0644 "${GRUB_PC_DIR}/${m}.mod" "${BOOT}/grub/i386-pc/${m}.mod"
	done

	kern_bytes="$(stat -c '%s' "${BOOT}/vmlinuz")"
	initrd_bytes="$(stat -c '%s' "${BOOT}/initrd.img")"
	core_bytes="$(stat -c '%s' "${BOOT}/grub/i386-pc/core.img")"

	cat >"${BOOT}/BOOT_LAYER.txt" <<EOF
boot layer — real boot chain (Slice 9.12)
status: real
firmware_bootable: false
boot_chain: real
artifacts:
  - kernel_image: vmlinuz (${kern_bytes} bytes, real Linux kernel)
  - initrd_image: initrd.img (${initrd_bytes} bytes, real initramfs)
  - bootloader_config: grub/grub.cfg (real GRUB config)
  - bootloader_core: grub/i386-pc/core.img (${core_bytes} bytes, real GRUB core)
  - bios_boot_blocks: grub/i386-pc/boot.img, grub/i386-pc/cdboot.img
blocker_for_firmware_bootable:
  - iso9660_or_eltorito_writer_missing (xorriso/genisoimage)
  - no_raw_image_filesystem_tooling (mkfs.vfat/mtools/loop-mount)
note: Kernel/initrd/core/grub.cfg are real. Not yet wrapped in bootable media.
EOF
else
	# Fallback: keep Slice 9.11 stub boot assets, record honestly.
	{
		echo "boot layer — foundation stub (carried forward at Slice 9.12)"
		echo "status: stub"
		echo "firmware_bootable: false"
		echo "boot_chain: stub"
		echo "reason: real kernel/initrd/grub-mkimage not available on this build host"
		echo "note: run on a host with /boot/vmlinuz-*, /boot/initrd.img-*, and GRUB tools"
	} >"${BOOT}/BOOT_LAYER.txt"
fi

# --- live_root: minimal runnable-environment additions ---
LR="${ENVELOPE_ROOT}/live_root"
mkdir -p "${LR}/etc" "${LR}/usr/lib/vyntrio"

cat >"${LR}/etc/os-release" <<'EOF'
NAME="Vyntrio Install Medium"
ID=vyntrio-live
ID_LIKE=debian
PRETTY_NAME="Vyntrio Install Medium (live)"
VARIANT="live-install"
VARIANT_ID=install
BOOTABILITY=partial
EOF

cat >"${LR}/etc/fstab" <<'EOF'
# Vyntrio live medium — pseudo-filesystems only. No target-disk mounts here.
proc            /proc   proc    nosuid,noexec,nodev     0 0
sysfs           /sys    sysfs   nosuid,noexec,nodev     0 0
devtmpfs        /dev    devtmpfs nosuid                 0 0
tmpfs           /run    tmpfs   nosuid,nodev            0 0
tmpfs           /tmp    tmpfs   nosuid,nodev            0 0
EOF

cat >"${LR}/usr/lib/vyntrio/live-init.sh" <<'EOF'
#!/usr/bin/env bash
# Vyntrio live init (Slice 9.12). Structured live-boot entry point.
# Real boot chain loads this rootfs concept, but a complete userland
# (busybox/systemd + vyntrio-api runtime) is NOT yet present, so the local
# dashboard is still unreachable on first boot. See FIRST_BOOT.txt.
set -euo pipefail

mount_pseudo() {
	mountpoint -q /proc 2>/dev/null || mount -t proc proc /proc 2>/dev/null || true
	mountpoint -q /sys  2>/dev/null || mount -t sysfs sysfs /sys 2>/dev/null || true
	mountpoint -q /run  2>/dev/null || mount -t tmpfs tmpfs /run 2>/dev/null || true
}

echo "vyntrio live-init: media_role=install boot_chain=${VYNTRIO_BOOTABILITY:-real}" >&2
mount_pseudo || true
echo "vyntrio live-init: dashboard_on_first_boot=false (userland incomplete)" >&2
echo "vyntrio live-init: see /usr/lib/vyntrio/FIRST_BOOT.txt for remaining gaps" >&2
# Future: launch dashboard from payload/usr/bin/vyntrio-api once a userland exists.
exit 0
EOF
chmod 0755 "${LR}/usr/lib/vyntrio/live-init.sh"

# Refresh first-boot honesty record for the real-bootchain state.
cat >"${LR}/usr/lib/vyntrio/FIRST_BOOT.txt" <<EOF
# First-boot / dashboard honesty record (Slice 9.12)
#
# Intended Unraid-like flow:
#   1. bootable USB/media
#   2. local browser dashboard
#   3. onboarding / first-boot setup
#   4. later install / storage / licensing / remote
#
# boot_chain: ${BOOT_CHAIN}
# firmware_bootable: false
# dashboard_reachable: false
#
# Progress this slice: real kernel, real initrd, real GRUB core + grub.cfg
# replace the previous text stubs (when built on a capable host).
#
# Still missing before first boot reaches the dashboard:
#   - firmware-bootable container (ISO9660/El Torito or raw image)
#     BLOCKED here: no xorriso/genisoimage, no mkfs.vfat/mtools, no EFI GRUB target
#   - complete live rootfs userland (busybox/systemd) inside initrd
#   - live session that runs vyntrio-api from media and serves the wizard UI
#   - USB creator to write the image to removable media
status: partial
dashboard_reachable: false
EOF

# Fail closed: never allow secrets/state into live_root.
while IFS= read -r -d '' path; do
	case "$(basename "${path}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			die "forbidden live_root file: ${path#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${LR}" -type f -print0)

# --- validation: boot_chain must match actual content (no regressions) ---
if [[ "${BOOT_CHAIN}" == "real" ]]; then
	[[ -f "${BOOT}/vmlinuz" ]] && grub-file --is-x86-linux "${BOOT}/vmlinuz" 2>/dev/null \
		|| die "real boot_chain but boot/vmlinuz is not a valid x86 Linux kernel"
	initrd_magic_ok "${BOOT}/initrd.img" || die "real boot_chain but boot/initrd.img has no known initramfs magic"
	grep -q '^\s*linux\s' "${BOOT}/grub/grub.cfg" || die "grub.cfg missing linux directive"
	grep -q '^\s*initrd\s' "${BOOT}/grub/grub.cfg" || die "grub.cfg missing initrd directive"
	[[ -f "${BOOT}/grub/i386-pc/core.img" && "$(stat -c '%s' "${BOOT}/grub/i386-pc/core.img")" -gt 20000 ]] \
		|| die "grub core.img missing or implausibly small"
	[[ -e "${BOOT}/vmlinuz.stub" || -e "${BOOT}/initrd.img.stub" ]] \
		&& die "real boot_chain but stub kernel/initrd still present"
else
	# Stub state must still carry the 9.11 placeholders.
	[[ -f "${BOOT}/vmlinuz.stub" && -f "${BOOT}/initrd.img.stub" ]] \
		|| die "stub boot_chain but 9.11 stub artifacts are missing"
fi

# --- image emission ---
rm -f "${BUILD_ROOT}/${ISO_IMAGE_NAME}" "${BUILD_ROOT}/${REAL_TAR_NAME}"
mkdir -p "${BUILD_ROOT}"
GEN_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

ARTIFACT_PATH=""
ARTIFACT_FORMAT=""
FIRMWARE_BOOTABLE=false

if [[ "${CAN_EMIT_ISO}" == true ]]; then
	# Real firmware-bootable ISO from the assembled boot tree.
	iso_src="$(mktemp -d)"; trap 'rm -rf "${iso_src}"' EXIT
	cp -a "${BOOT}" "${iso_src}/boot"
	cp -a "${LR}"   "${iso_src}/live_root"
	cp -a "${ENVELOPE_ROOT}/payload" "${iso_src}/payload"
	if grub-mkrescue -o "${BUILD_ROOT}/${ISO_IMAGE_NAME}" "${iso_src}" >/dev/null 2>&1; then
		ARTIFACT_PATH="distro/install-media/build/${ISO_IMAGE_NAME}"
		ARTIFACT_FORMAT="iso9660_eltorito"
		FIRMWARE_BOOTABLE=true
	fi
fi

if [[ -z "${ARTIFACT_PATH}" ]]; then
	# No ISO writer (or emission failed): emit the real boot tree as a tar.
	stage="$(mktemp -d)"; trap 'rm -rf "${stage}"' EXIT
	mkdir -p "${stage}/vyntrio-install-media"
	cp -a "${BOOT}"                       "${stage}/vyntrio-install-media/boot"
	cp -a "${LR}"                         "${stage}/vyntrio-install-media/live_root"
	cp -a "${ENVELOPE_ROOT}/payload"      "${stage}/vyntrio-install-media/payload"
	[[ -f "${ENVELOPE_ROOT}/ENVELOPE.txt" ]] && cp -a "${ENVELOPE_ROOT}/ENVELOPE.txt" "${stage}/vyntrio-install-media/"
	cat >"${stage}/vyntrio-install-media/NOT_BOOTABLE.txt" <<EOF
Vyntrio install-media candidate (Slice 9.12).
boot_chain: ${BOOT_CHAIN}
firmware_bootable: false

The boot chain is REAL (kernel, initrd, GRUB core, grub.cfg) when built on a
capable host, but this archive is NOT yet a firmware-bootable image.

Blocked here by missing host tooling:
  - ISO9660/El Torito writer (xorriso / genisoimage) — no bootable ISO
  - FAT tooling (mkfs.vfat / mtools) + loop-mount privileges — no raw image
  - x86_64-efi GRUB target — no UEFI image

Do NOT dd/flash this tar expecting a bootable system.
See live_root/usr/lib/vyntrio/FIRST_BOOT.txt for the remaining flow gaps.
EOF
	if [[ "${BOOT_CHAIN}" == "real" ]]; then
		tar -C "${stage}" -cf "${BUILD_ROOT}/${REAL_TAR_NAME}" vyntrio-install-media
		ARTIFACT_PATH="distro/install-media/build/${REAL_TAR_NAME}"
		ARTIFACT_FORMAT="tar_real_bootchain_no_iso"
	else
		tar -C "${stage}" -cf "${BUILD_ROOT}/${STUB_TAR_NAME}" vyntrio-install-media
		ARTIFACT_PATH="distro/install-media/build/${STUB_TAR_NAME}"
		ARTIFACT_FORMAT="tar_stub"
	fi
fi

# --- provenance record (distinct from Slice 9.11 IMAGE_STUB.txt) ---
{
	echo "# Generated by scripts/build-install-media-image.sh — do not commit"
	echo "schema_version: vyntrio-install-media-image-v1"
	echo "slice: 9.12"
	echo "generated_at: ${GEN_UTC}"
	echo "media_role: install"
	echo "boot_chain: ${BOOT_CHAIN}"
	echo "firmware_bootable: ${FIRMWARE_BOOTABLE}"
	echo "usb_creator: deferred"
	echo "dashboard_on_first_boot: false"
	echo "artifact: ${ARTIFACT_PATH}"
	echo "artifact_format: ${ARTIFACT_FORMAT}"
	echo "verified:"
	echo "  kernel_is_x86_linux: ${KERNEL_VERIFIED}"
	echo "  initrd_magic_ok: ${INITRD_VERIFIED}"
	echo "  grub_core_built: ${GRUB_CORE_BUILT}"
	if [[ "${BOOT_CHAIN}" == "real" ]]; then
		echo "sources:"
		echo "  kernel: ${KERNEL_SRC}"
		echo "  initrd: ${INITRD_SRC}"
	fi
	echo "iso_writer: ${ISO_WRITER:-none}"
	echo "blockers_for_firmware_bootable:"
	if [[ "${FIRMWARE_BOOTABLE}" == true ]]; then
		echo "  - none"
	else
		echo "  - iso9660_or_eltorito_writer_missing"
		echo "  - raw_image_filesystem_tooling_missing"
		echo "  - efi_grub_target_missing"
	fi
	echo "missing_for_usb_creator:"
	echo "  - firmware_bootable_image_container"
	echo "  - complete_live_root_userland"
	echo "  - live_session_vyntrio_api_host"
	echo "  - host_usb_writer_tool"
} >"${BUILD_ROOT}/IMAGE.txt"

echo "install-media image build: boot_chain=${BOOT_CHAIN} firmware_bootable=${FIRMWARE_BOOTABLE}"
echo "install-media image artifact: ${ARTIFACT_PATH}"
if [[ "${FIRMWARE_BOOTABLE}" != true ]]; then
	echo "install-media image: NOT firmware-bootable — see distro/install-media/build/IMAGE.txt blockers" >&2
fi
