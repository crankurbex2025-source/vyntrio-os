#!/usr/bin/env bash
# Initialize install-media bootability foundation (Block 9 / Slice 9.11).
# Populates boot/ + live_root/ stubs from the Slice 9.8 envelope, validates
# layout, writes BOOTABILITY.txt, and emits a clearly non-bootable image stub.
#
# Honest limits: no real kernel, initrd, bootloader, or firmware-bootable ISO/USB.
# USB creator remains a future slice.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVELOPE_ROOT="${VYNTRIO_INSTALL_ENVELOPE_ROOT:-${ROOT}/distro/install-media/envelope}"
BUILD_ROOT="${VYNTRIO_INSTALL_MEDIA_BUILD_ROOT:-${ROOT}/distro/install-media/build}"
STUB_IMAGE_NAME="vyntrio-install-media-NOT-BOOTABLE.stub.tar"
STUB_IMAGE_PATH="${BUILD_ROOT}/${STUB_IMAGE_NAME}"

readonly -a PAYLOAD_FILES=(
	"usr/bin/vyntrio-api"
	"usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service"
	"usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf"
	"etc/vyntrio/config.toml"
)

if [[ ! -d "${ENVELOPE_ROOT}/payload" ]]; then
	echo "initialize-install-media-bootability: envelope missing: distro/install-media/envelope/" >&2
	echo "hint: run 'make install-media-envelope' first" >&2
	exit 1
fi

for rel in "${PAYLOAD_FILES[@]}"; do
	if [[ ! -f "${ENVELOPE_ROOT}/payload/${rel}" ]]; then
		echo "initialize-install-media-bootability: missing payload file: ${rel}" >&2
		echo "hint: run 'make install-media-envelope' first" >&2
		exit 1
	fi
done

mapfile -t payload_files < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
if [[ "${#payload_files[@]}" -ne "${#PAYLOAD_FILES[@]}" ]]; then
	echo "initialize-install-media-bootability: unexpected payload file count" >&2
	exit 1
fi

for path in "${payload_files[@]}"; do
	rel="${path#${ENVELOPE_ROOT}/payload/}"
	expected=false
	for allowed in "${PAYLOAD_FILES[@]}"; do
		if [[ "${rel}" == "${allowed}" ]]; then
			expected=true
			break
		fi
	done
	if [[ "${expected}" != true ]]; then
		echo "initialize-install-media-bootability: non-manifest payload file: ${rel}" >&2
		exit 1
	fi
done

# --- boot layer (replaces deferred LAYER.txt) ---
rm -rf "${ENVELOPE_ROOT}/boot"
mkdir -p "${ENVELOPE_ROOT}/boot/loader/entries"

cat >"${ENVELOPE_ROOT}/boot/BOOT_LAYER.txt" <<'EOF'
boot layer — foundation stub (Slice 9.11)
status: stub
firmware_bootable: false
artifacts:
  - bootloader_config: loader/entries/vyntrio-install.conf.stub
  - kernel_image: vmlinuz.stub
  - initrd_image: initrd.img.stub
note: Placeholders only. No UEFI/BIOS loader, kernel, or initrd binary is shipped.
EOF

cat >"${ENVELOPE_ROOT}/boot/loader/entries/vyntrio-install.conf.stub" <<'EOF'
# Vyntrio install bootloader entry — STUB (not consumed by firmware)
# Future: title, linux, initrd, options for live install medium.
title Vyntrio Install (stub — not bootable)
linux /vmlinuz.stub
initrd /initrd.img.stub
options vyntrio.media_role=install vyntrio.bootability=stub
EOF

printf 'vyntrio-kernel-stub: not a real kernel image\n' >"${ENVELOPE_ROOT}/boot/vmlinuz.stub"
printf 'vyntrio-initrd-stub: not a real initrd image\n' >"${ENVELOPE_ROOT}/boot/initrd.img.stub"

# --- live_root layer (replaces deferred LAYER.txt) ---
rm -rf "${ENVELOPE_ROOT}/live_root"
mkdir -p \
	"${ENVELOPE_ROOT}/live_root/bin" \
	"${ENVELOPE_ROOT}/live_root/sbin" \
	"${ENVELOPE_ROOT}/live_root/etc/vyntrio" \
	"${ENVELOPE_ROOT}/live_root/etc/network" \
	"${ENVELOPE_ROOT}/live_root/usr/bin" \
	"${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio" \
	"${ENVELOPE_ROOT}/live_root/proc" \
	"${ENVELOPE_ROOT}/live_root/sys" \
	"${ENVELOPE_ROOT}/live_root/dev" \
	"${ENVELOPE_ROOT}/live_root/run" \
	"${ENVELOPE_ROOT}/live_root/tmp" \
	"${ENVELOPE_ROOT}/live_root/var/tmp" \
	"${ENVELOPE_ROOT}/live_root/var/log" \
	"${ENVELOPE_ROOT}/live_root/mnt/install"

cat >"${ENVELOPE_ROOT}/live_root/LIVE_ROOT.txt" <<'EOF'
live_root layer — foundation stub (Slice 9.11)
status: stub
firmware_bootable: false
dashboard_on_first_boot: false
artifacts:
  - minimal_root_filesystem: directory skeleton only
  - live_init_system: usr/lib/vyntrio/live-init.sh (stub)
  - first_boot_record: usr/lib/vyntrio/FIRST_BOOT.txt
  - future_installer_host_runtime: deferred
excludes:
  - appliance_sqlite_state
  - owner_credentials
  - bootstrap_tokens
  - license_files
  - backup_artifacts
  - recovery_media_tooling
  - preinstalled_target_disk_payloads
note: Not a complete Linux rootfs. No systemd, no packaged distro root.
note: vyntrio-api lives under payload/ for target install; it is not started in live_root yet.
EOF

cat >"${ENVELOPE_ROOT}/live_root/etc/hostname" <<'EOF'
vyntrio-install
EOF

cat >"${ENVELOPE_ROOT}/live_root/etc/hosts" <<'EOF'
127.0.0.1 localhost
127.0.1.1 vyntrio-install
::1 localhost ip6-localhost ip6-loopback
EOF

cat >"${ENVELOPE_ROOT}/live_root/etc/vyntrio/live.env" <<'EOF'
# Vyntrio live install environment — STUB
VYNTRIO_MEDIA_ROLE=install
VYNTRIO_BOOTABILITY=stub
VYNTRIO_LIVE_ROOT=1
VYNTRIO_DASHBOARD_ON_BOOT=0
EOF

cat >"${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio/FIRST_BOOT.txt" <<'EOF'
# First-boot / dashboard honesty record (Slice 9.11)
#
# Intended Unraid-like flow:
#   1. bootable USB/media
#   2. local browser dashboard
#   3. onboarding / first-boot setup
#   4. later install / storage / licensing / remote
#
# This media candidate does NOT yet reach the local dashboard on first boot.
#
# Still missing for dashboard-on-first-boot:
#   - firmware-bootable image (real kernel, initrd, bootloader)
#   - complete live rootfs with networking
#   - live-session host that runs vyntrio-api from media
#   - install wizard UI served from live session
#   - USB creator to write media to removable disk
#
# payload/ contains vyntrio-api for future target-disk install only.
# It is not launched by the live-init stub in this slice.
status: deferred
dashboard_reachable: false
EOF

cat >"${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio/live-init.sh" <<'EOF'
#!/usr/bin/env bash
# Stub live init for Vyntrio install medium (Slice 9.11).
# Not a PID 1 replacement. Does not start vyntrio-api or serve the dashboard.
set -euo pipefail
echo "vyntrio live-init stub: media_role=install bootability=stub" >&2
echo "vyntrio live-init stub: dashboard_on_first_boot=false" >&2
echo "vyntrio live-init stub: not firmware-bootable; USB creator / real rootfs deferred" >&2
if [[ -f /usr/lib/vyntrio/FIRST_BOOT.txt ]]; then
	echo "vyntrio live-init stub: see /usr/lib/vyntrio/FIRST_BOOT.txt for remaining gaps" >&2
fi
exit 0
EOF
chmod 0755 "${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio/live-init.sh"

ln -sfn ../usr/lib/vyntrio/live-init.sh "${ENVELOPE_ROOT}/live_root/sbin/init.stub"

# --- layout validation ---
for layer in boot live_root payload; do
	if [[ ! -d "${ENVELOPE_ROOT}/${layer}" ]]; then
		echo "initialize-install-media-bootability: missing layer after init: ${layer}" >&2
		exit 1
	fi
done

if [[ -f "${ENVELOPE_ROOT}/boot/LAYER.txt" ]] || [[ -f "${ENVELOPE_ROOT}/live_root/LAYER.txt" ]]; then
	echo "initialize-install-media-bootability: deferred LAYER.txt must be replaced" >&2
	exit 1
fi

for req in \
	"${ENVELOPE_ROOT}/boot/BOOT_LAYER.txt" \
	"${ENVELOPE_ROOT}/boot/vmlinuz.stub" \
	"${ENVELOPE_ROOT}/boot/initrd.img.stub" \
	"${ENVELOPE_ROOT}/boot/loader/entries/vyntrio-install.conf.stub" \
	"${ENVELOPE_ROOT}/live_root/LIVE_ROOT.txt" \
	"${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio/live-init.sh" \
	"${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio/FIRST_BOOT.txt" \
	"${ENVELOPE_ROOT}/live_root/etc/vyntrio/live.env" \
	"${ENVELOPE_ROOT}/live_root/etc/hostname" \
	"${ENVELOPE_ROOT}/live_root/etc/hosts"
do
	if [[ ! -e "${req}" ]]; then
		echo "initialize-install-media-bootability: missing required bootability artifact: ${req#${ENVELOPE_ROOT}/}" >&2
		exit 1
	fi
done

# Forbidden content in live_root (fail closed).
while IFS= read -r -d '' path; do
	base="$(basename "${path}")"
	case "${base}" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			echo "initialize-install-media-bootability: forbidden live_root file: ${path#${ENVELOPE_ROOT}/}" >&2
			exit 1
			;;
	esac
done < <(find "${ENVELOPE_ROOT}/live_root" -type f -print0)

# --- provenance ---
{
	echo "# Generated by scripts/initialize-install-media-bootability.sh — do not commit"
	echo "schema_version: vyntrio-install-media-bootability-v1"
	echo "media_role: install"
	echo "slice: 9.11"
	echo "generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
	echo "envelope_root: distro/install-media/envelope/"
	echo "firmware_bootable: false"
	echo "usb_creator: deferred"
	echo "dashboard_on_first_boot: false"
	echo "product_path_step: bootable_media_foundation"
	echo "layers:"
	echo "  boot: stub"
	echo "  live_root: stub"
	echo "  payload: carried_forward"
	echo "image_stub:"
	echo "  path: distro/install-media/build/${STUB_IMAGE_NAME}"
	echo "  firmware_bootable: false"
	echo "  format: tar_provenance_archive"
	echo "first_boot:"
	echo "  dashboard_reachable: false"
	echo "  record: live_root/usr/lib/vyntrio/FIRST_BOOT.txt"
	echo "  missing:"
	echo "    - firmware_bootable_image"
	echo "    - complete_live_rootfs_with_networking"
	echo "    - live_session_vyntrio_api_host"
	echo "    - install_wizard_ui_on_media"
	echo "missing_for_usb_creator:"
	echo "  - real_kernel_and_initrd"
	echo "  - uefi_bios_bootloader"
	echo "  - complete_live_rootfs"
	echo "  - firmware_bootable_iso_or_raw_image"
	echo "  - host_usb_writer_tool"
	echo "payload_files:"
	for rel in "${PAYLOAD_FILES[@]}"; do
		echo "  - payload/${rel}"
	done
} >"${ENVELOPE_ROOT}/BOOTABILITY.txt"

# Update ENVELOPE.txt layer status if present.
if [[ -f "${ENVELOPE_ROOT}/ENVELOPE.txt" ]]; then
	tmp="$(mktemp)"
	sed \
		-e 's/^  boot: deferred$/  boot: stub/' \
		-e 's/^  live_root: deferred$/  live_root: stub/' \
		"${ENVELOPE_ROOT}/ENVELOPE.txt" >"${tmp}"
	mv "${tmp}" "${ENVELOPE_ROOT}/ENVELOPE.txt"
fi

# --- image emission stub (not firmware-bootable) ---
rm -rf "${BUILD_ROOT}"
mkdir -p "${BUILD_ROOT}"

STAGING_TAR="$(mktemp -d)"
trap 'rm -rf "${STAGING_TAR}"' EXIT

mkdir -p "${STAGING_TAR}/vyntrio-install-media"
cp -a "${ENVELOPE_ROOT}/boot" "${STAGING_TAR}/vyntrio-install-media/"
cp -a "${ENVELOPE_ROOT}/live_root" "${STAGING_TAR}/vyntrio-install-media/"
cp -a "${ENVELOPE_ROOT}/payload" "${STAGING_TAR}/vyntrio-install-media/"
cp -a "${ENVELOPE_ROOT}/BOOTABILITY.txt" "${STAGING_TAR}/vyntrio-install-media/"
if [[ -f "${ENVELOPE_ROOT}/ENVELOPE.txt" ]]; then
	cp -a "${ENVELOPE_ROOT}/ENVELOPE.txt" "${STAGING_TAR}/vyntrio-install-media/"
fi

cat >"${STAGING_TAR}/vyntrio-install-media/NOT_BOOTABLE.txt" <<'EOF'
This archive is a Vyntrio install-media IMAGE STUB / media candidate (Slice 9.11).
It is NOT firmware-bootable.
It is NOT a USB creator output.
It does NOT reach the local browser dashboard on first boot.
Do not dd or flash this file to removable media expecting a bootable system.
See live_root/usr/lib/vyntrio/FIRST_BOOT.txt for the remaining Unraid-like flow gaps.
Future slices will emit a real ISO/raw image after kernel, initrd, and bootloader land.
EOF

tar -C "${STAGING_TAR}" -cf "${STUB_IMAGE_PATH}" vyntrio-install-media

{
	echo "# Generated by scripts/initialize-install-media-bootability.sh — do not commit"
	echo "schema_version: vyntrio-install-media-image-stub-v1"
	echo "artifact: ${STUB_IMAGE_NAME}"
	echo "firmware_bootable: false"
	echo "usb_creator: deferred"
	echo "dashboard_on_first_boot: false"
	echo "generated_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
	echo "source_envelope: distro/install-media/envelope/"
} >"${BUILD_ROOT}/IMAGE_STUB.txt"

echo "install-media bootability foundation initialized under ${ENVELOPE_ROOT}"
echo "install-media image stub (NOT bootable): ${STUB_IMAGE_PATH}"
