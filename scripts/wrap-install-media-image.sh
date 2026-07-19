#!/usr/bin/env bash
# Wrap the Slice 9.12 real boot chain into a firmware-bootable image (Block 9 /
# Slice 9.13), or fall back honestly when host tooling cannot produce one.
#
# Strategies, in priority order:
#   A. ISO9660/El Torito   — grub-mkrescue + (xorriso|xorrisofs|genisoimage)
#   B. Raw BIOS disk image — mke2fs + debugfs + sfdisk + grub-bios-setup
#                            (the same method grub-install uses), on non-tmpfs
#   C. Honest fallback     — keep the real boot-chain tar, emit a documented
#                            blocker naming exactly which tool(s) are missing.
#
# Honesty rules:
#   - firmware_bootable: true ONLY when a real bootable-format image is emitted
#     and passes structural verification. No fake bootability.
#   - Runtime boot is NOT claimed unless a VM/boot test actually ran. This host
#     has no qemu, so raw/ISO images are marked runtime_boot_tested: false.
#   - No target-disk writes, no installer/apply behavior, USB creator deferred.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENVELOPE_ROOT="${VYNTRIO_INSTALL_ENVELOPE_ROOT:-${ROOT}/distro/install-media/envelope}"
BUILD_ROOT="${VYNTRIO_INSTALL_MEDIA_BUILD_ROOT:-${ROOT}/distro/install-media/build}"
GRUB_PC_DIR="${VYNTRIO_GRUB_I386_PC_DIR:-/usr/lib/grub/i386-pc}"

BOOT="${ENVELOPE_ROOT}/boot"
ISO_NAME="vyntrio-install-media.iso"
HYBRID_NAME="vyntrio-install-media.img"
RAW_NAME="vyntrio-install-media-bios.img"
REAL_TAR_NAME="vyntrio-install-media-REAL-BOOTCHAIN-NO-ISO.tar"
GRUB_EFI_DIR="${VYNTRIO_GRUB_X86_64_EFI_DIR:-/usr/lib/grub/x86_64-efi}"

die() { echo "wrap-install-media-image: $*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

# --- precondition: a real boot chain must exist (Slice 9.12) ---
[[ -d "${BOOT}" ]] || die "boot layer missing; run 'make install-media-image' first"

BOOT_CHAIN="stub"
if [[ -f "${BOOT}/vmlinuz" && -f "${BOOT}/initrd.img" && -f "${BOOT}/grub/grub.cfg" ]] \
	&& have grub-file && grub-file --is-x86-linux "${BOOT}/vmlinuz" 2>/dev/null; then
	BOOT_CHAIN="real"
fi

# --- capability detection ---
ISO_WRITER=""
for w in xorriso xorrisofs genisoimage mkisofs; do have "${w}" && { ISO_WRITER="${w}"; break; }; done

CAN_ISO=false
if [[ "${BOOT_CHAIN}" == "real" ]] && have grub-mkrescue && [[ -n "${ISO_WRITER}" ]]; then
	CAN_ISO=true
fi

# Raw BIOS image needs: FS creation (mke2fs+debugfs), partitioning (sfdisk),
# GRUB BIOS install (grub-bios-setup+grub-mkimage), and a NON-tmpfs build dir
# (grub-bios-setup cannot canonicalize a tmpfs/overlay backing device).
BUILD_FSTYPE="$(stat -f -c %T "${BUILD_ROOT}" 2>/dev/null || echo unknown)"
RAW_TOOLS_OK=true
declare -a RAW_MISSING=()
for t in mke2fs debugfs sfdisk grub-bios-setup grub-mkimage; do
	have "${t}" || { RAW_TOOLS_OK=false; RAW_MISSING+=("${t}"); }
done
NONTMPFS_OK=true
case "${BUILD_FSTYPE}" in
	tmpfs|ramfs|overlayfs) NONTMPFS_OK=false ;;
esac
CAN_RAW=false
if [[ "${BOOT_CHAIN}" == "real" && "${RAW_TOOLS_OK}" == true && "${NONTMPFS_OK}" == true ]]; then
	CAN_RAW=true
fi

# Dual-mode GPT hybrid (UEFI ESP + BIOS boot partition) — required product baseline.
HYBRID_TOOLS_OK=true
declare -a HYBRID_MISSING=()
for t in mkfs.vfat mcopy mmd sgdisk grub-mkimage grub-bios-setup; do
	have "${t}" || { HYBRID_TOOLS_OK=false; HYBRID_MISSING+=("${t}"); }
done
[[ -d "${GRUB_EFI_DIR}" ]] || { HYBRID_TOOLS_OK=false; HYBRID_MISSING+=("grub_x86_64_efi_modules"); }
CAN_HYBRID=false
if [[ "${BOOT_CHAIN}" == "real" && "${HYBRID_TOOLS_OK}" == true && "${NONTMPFS_OK}" == true && "${RAW_TOOLS_OK}" == true ]]; then
	CAN_HYBRID=true
fi

mkdir -p "${BUILD_ROOT}"
rm -f "${BUILD_ROOT}/${ISO_NAME}" "${BUILD_ROOT}/${HYBRID_NAME}" "${BUILD_ROOT}/${RAW_NAME}"
GEN_UTC="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

ARTIFACT_PATH=""
ARTIFACT_FORMAT=""
BOOT_METHOD=""
BOOT_MODE=""
FIRMWARE_BOOTABLE=false
STRUCTURAL="not_run"
declare -a BLOCKERS=()

emit_iso() {
	local src; src="$(mktemp -d)"
	cp -a "${BOOT}" "${src}/boot"
	cp -a "${ENVELOPE_ROOT}/live_root" "${src}/live_root"
	cp -a "${ENVELOPE_ROOT}/payload" "${src}/payload"
	if grub-mkrescue -o "${BUILD_ROOT}/${ISO_NAME}" "${src}" >/dev/null 2>&1 \
		&& [[ -s "${BUILD_ROOT}/${ISO_NAME}" ]]; then
		rm -rf "${src}"; return 0
	fi
	rm -rf "${src}"; return 1
}

emit_raw_hybrid() {
	# Dual-mode GPT disk: BIOS boot (EF02) + ESP FAT32 (EF00) with BOOTX64.EFI,
	# shared grub.cfg, and /boot/{vmlinuz,initrd.img}. BIOS via grub-bios-setup;
	# UEFI via removable-media path EFI/BOOT/BOOTX64.EFI.
	local work esp_fs img kbytes ibytes esp_mib disk_mib
	local bios_start=2048 bios_sectors=2048 esp_start esp_sectors
	local core_bios core_efi cfg bd esp_off
	work="$(mktemp -d -p "${BUILD_ROOT}")"
	esp_fs="${work}/esp.img"
	img="${BUILD_ROOT}/${HYBRID_NAME}"
	core_bios="${work}/core.img"
	core_efi="${work}/BOOTX64.EFI"
	cfg="${work}/grub.cfg"
	bd="${work}/bootdir"

	kbytes="$(stat -c '%s' "${BOOT}/vmlinuz")"
	ibytes="$(stat -c '%s' "${BOOT}/initrd.img")"
	esp_mib=$(( (kbytes + ibytes) / 1048576 + 48 ))
	(( esp_mib < 192 )) && esp_mib=192

	esp_start=$((bios_start + bios_sectors))
	esp_sectors=$((esp_mib * 2048))
	# GPT primary + partitions + backup GPT slack.
	disk_mib=$(( (esp_start + esp_sectors) / 2048 + 2 ))

	truncate -s "${esp_mib}M" "${esp_fs}"
	mkfs.vfat -F 32 -n VYNTRIO "${esp_fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }

	grub-mkimage -O x86_64-efi -p /EFI/BOOT -o "${core_efi}" \
		fat part_gpt part_msdos normal linux configfile echo search search_fs_file search_label \
		test boot ls cat halt reboot all_video gzio \
		>/dev/null 2>&1 || { rm -rf "${work}"; return 1; }

	grub-mkimage -O i386-pc -p '(hd0,gpt2)/EFI/BOOT' -o "${core_bios}" \
		biosdisk part_gpt fat normal linux configfile echo search search_fs_file search_label \
		test boot ls cat halt reboot \
		>/dev/null 2>&1 || { rm -rf "${work}"; return 1; }

	cat >"${cfg}" <<'EOF'
set default=0
set timeout=5
insmod part_gpt
insmod fat
search --set=root --file /boot/vmlinuz
menuentry "Vyntrio Install (live)" {
	linux /boot/vmlinuz boot=live vyntrio.media_role=install vyntrio.bootability=real quiet
	initrd /boot/initrd.img
}
menuentry "Vyntrio Install (verbose)" {
	linux /boot/vmlinuz boot=live vyntrio.media_role=install vyntrio.bootability=real
	initrd /boot/initrd.img
}
EOF

	export MTOOLS_SKIP_CHECK=1
	mmd -i "${esp_fs}" ::EFI ::EFI/BOOT ::boot >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	mcopy -i "${esp_fs}" "${core_efi}" ::EFI/BOOT/BOOTX64.EFI >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	mcopy -i "${esp_fs}" "${cfg}" ::EFI/BOOT/grub.cfg >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	mcopy -i "${esp_fs}" "${BOOT}/vmlinuz" ::boot/vmlinuz >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	mcopy -i "${esp_fs}" "${BOOT}/initrd.img" ::boot/initrd.img >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }

	truncate -s "${disk_mib}M" "${img}"
	sgdisk -o "${img}" >/dev/null 2>&1 || { rm -rf "${work}"; rm -f "${img}"; return 1; }
	sgdisk -n "1:${bios_start}:+1M" -t 1:EF02 -c 1:"BIOS boot" "${img}" >/dev/null 2>&1 \
		|| { rm -rf "${work}"; rm -f "${img}"; return 1; }
	sgdisk -n "2:0:+${esp_mib}M" -t 2:EF00 -c 2:"EFI System" "${img}" >/dev/null 2>&1 \
		|| { rm -rf "${work}"; rm -f "${img}"; return 1; }

	dd if="${esp_fs}" of="${img}" bs=512 seek="${esp_start}" conv=notrunc status=none \
		|| { rm -rf "${work}"; rm -f "${img}"; return 1; }

	mkdir -p "${bd}"
	cp "${GRUB_PC_DIR}/boot.img" "${bd}/boot.img"
	cp "${core_bios}" "${bd}/core.img"
	if ! grub-bios-setup -d "${bd}" -b boot.img -c core.img "${img}" >/dev/null 2>&1; then
		rm -rf "${work}"; rm -f "${img}"; return 1
	fi

	# Record ESP byte offset for swap/extract tooling.
	esp_off=$((esp_start * 512))
	printf '%s\n' "${esp_off}" >"${BUILD_ROOT}/HYBRID_ESP_OFFSET.txt"
	rm -rf "${work}"
	return 0
}

emit_raw() {
	# Build a real MBR/BIOS disk image: 1MiB gap for GRUB core, one bootable
	# ext2 partition holding /boot/{vmlinuz,initrd.img,grub/...}.
	local work fs img kbytes ibytes fs_mib
	work="$(mktemp -d -p "${BUILD_ROOT}")"   # keep on the same (non-tmpfs) FS
	fs="${work}/fs.img"
	img="${BUILD_ROOT}/${RAW_NAME}"

	kbytes="$(stat -c '%s' "${BOOT}/vmlinuz")"
	ibytes="$(stat -c '%s' "${BOOT}/initrd.img")"
	# FS size = kernel+initrd + 32MiB slack (modules, grub, fs overhead), min 128MiB.
	fs_mib=$(( (kbytes + ibytes) / 1048576 + 32 ))
	(( fs_mib < 128 )) && fs_mib=128

	truncate -s "${fs_mib}M" "${fs}"
	mke2fs -q -t ext2 -F "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }

	# Core image with a prefix pointing at the ext2 on the first MBR partition.
	local core="${work}/core.img"
	grub-mkimage -O i386-pc -p '(hd0,msdos1)/boot/grub' -o "${core}" \
		biosdisk part_msdos ext2 normal linux configfile echo search test boot ls cat halt reboot \
		>/dev/null 2>&1 || { rm -rf "${work}"; return 1; }

	# grub.cfg for the raw layout (kernel/initrd at partition root /boot).
	local cfg="${work}/grub.cfg"
	cat >"${cfg}" <<'EOF'
set default=0
set timeout=5
menuentry "Vyntrio Install (live)" {
	linux /boot/vmlinuz boot=live vyntrio.media_role=install vyntrio.bootability=real quiet
	initrd /boot/initrd.img
}
menuentry "Vyntrio Install (verbose)" {
	linux /boot/vmlinuz boot=live vyntrio.media_role=install vyntrio.bootability=real
	initrd /boot/initrd.img
}
EOF

	# Populate the ext2 image without mounting/root, via debugfs.
	local d
	for d in /boot /boot/grub /boot/grub/i386-pc; do
		debugfs -w -R "mkdir ${d}" "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	done
	debugfs -w -R "write ${BOOT}/vmlinuz /boot/vmlinuz" "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	debugfs -w -R "write ${BOOT}/initrd.img /boot/initrd.img" "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	debugfs -w -R "write ${cfg} /boot/grub/grub.cfg" "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	debugfs -w -R "write ${core} /boot/grub/i386-pc/core.img" "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	[[ -f "${GRUB_PC_DIR}/boot.img" ]] && debugfs -w -R "write ${GRUB_PC_DIR}/boot.img /boot/grub/i386-pc/boot.img" "${fs}" >/dev/null 2>&1
	local m
	for m in "${GRUB_PC_DIR}"/*.mod; do
		[[ -f "${m}" ]] && debugfs -w -R "write ${m} /boot/grub/i386-pc/$(basename "${m}")" "${fs}" >/dev/null 2>&1
	done

	# Verify FS content before embedding (structural, pre-write).
	debugfs -R "stat /boot/vmlinuz" "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	debugfs -R "stat /boot/initrd.img" "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }
	debugfs -R "stat /boot/grub/grub.cfg" "${fs}" >/dev/null 2>&1 || { rm -rf "${work}"; return 1; }

	# Assemble disk: 1MiB gap + partition.
	truncate -s "$((fs_mib + 2))M" "${img}"
	printf 'label: dos\nstart=2048, type=83, bootable\n' | sfdisk "${img}" >/dev/null 2>&1 || { rm -rf "${work}"; rm -f "${img}"; return 1; }
	dd if="${fs}" of="${img}" bs=1M seek=1 conv=notrunc status=none || { rm -rf "${work}"; rm -f "${img}"; return 1; }

	# Install GRUB boot.img (MBR) + core.img (embed area). Same as grub-install.
	local bd="${work}/bootdir"
	mkdir -p "${bd}"; cp "${GRUB_PC_DIR}/boot.img" "${bd}/boot.img"; cp "${core}" "${bd}/core.img"
	if ! grub-bios-setup -d "${bd}" -b boot.img -c core.img "${img}" >/dev/null 2>&1; then
		rm -rf "${work}"; rm -f "${img}"; return 1
	fi
	rm -rf "${work}"
	return 0
}

# --- try strategies in priority order ---
# Dual-mode hybrid is the required product baseline. BIOS-only is incomplete.
if [[ "${CAN_HYBRID}" == true ]] && emit_raw_hybrid; then
	ARTIFACT_PATH="distro/install-media/build/${HYBRID_NAME}"
	ARTIFACT_FORMAT="raw_gpt_hybrid_disk"
	BOOT_METHOD="grub_bios_setup+grub_mkimage_x86_64_efi"
	BOOT_MODE="bios+uefi"
	FIRMWARE_BOOTABLE=true
elif [[ "${CAN_ISO}" == true ]] && [[ -d "${GRUB_EFI_DIR}" ]] && emit_iso; then
	ARTIFACT_PATH="distro/install-media/build/${ISO_NAME}"
	ARTIFACT_FORMAT="iso9660_eltorito"
	BOOT_METHOD="grub_mkrescue_${ISO_WRITER}"
	BOOT_MODE="bios+uefi"
	FIRMWARE_BOOTABLE=true
elif [[ "${CAN_HYBRID}" != true && "${CAN_RAW}" == true ]] && emit_raw; then
	ARTIFACT_PATH="distro/install-media/build/${RAW_NAME}"
	ARTIFACT_FORMAT="raw_mbr_bios_disk"
	BOOT_METHOD="grub_bios_setup"
	BOOT_MODE="bios_legacy"
	FIRMWARE_BOOTABLE=true
	BLOCKERS+=("uefi_incomplete: BIOS-only fallback only — dual-mode hybrid tools unavailable; product baseline requires UEFI")
fi

# --- structural verification of an emitted image; downgrade on any doubt ---
if [[ "${FIRMWARE_BOOTABLE}" == true && "${ARTIFACT_FORMAT}" == "raw_gpt_hybrid_disk" ]]; then
	img="${ROOT}/${ARTIFACT_PATH}"
	ok=true
	# Protective/legacy MBR signature still present on GPT disks.
	[[ "$(dd if="${img}" bs=1 skip=510 count=2 status=none | od -An -tx1 | tr -d ' \n')" == "55aa" ]] || ok=false
	# GPT signature at LBA 1.
	[[ "$(dd if="${img}" bs=512 skip=1 count=1 status=none | head -c 8)" == "EFI PART" ]] || ok=false
	# ESP hosts BOOTX64.EFI (PE MZ header) via mtools.
	esp_off="$(cat "${BUILD_ROOT}/HYBRID_ESP_OFFSET.txt" 2>/dev/null || echo 1048576)"
	export MTOOLS_SKIP_CHECK=1
	mz="$(mcopy -i "${img}@@${esp_off}" ::EFI/BOOT/BOOTX64.EFI - 2>/dev/null | head -c 2 || true)"
	[[ "${mz}" == "MZ" ]] || ok=false
	esp_listing="$(mdir -i "${img}@@${esp_off}" ::boot 2>/dev/null || true)"
	grep -qiE 'vmlinuz' <<<"${esp_listing}" || ok=false
	grep -qiE 'initrd' <<<"${esp_listing}" || ok=false
	# BIOS GRUB stage present.
	[[ "$(dd if="${img}" bs=512 count=2048 status=none 2>/dev/null | grep -a -c GRUB)" -ge 1 ]] || ok=false
	# Partition types: BIOS boot + EFI system.
	sgdisk -p "${img}" 2>/dev/null | grep -qi 'EF02\|BIOS boot' || ok=false
	sgdisk -p "${img}" 2>/dev/null | grep -qi 'EF00\|EFI System' || ok=false
	if [[ "${ok}" == true ]]; then
		STRUCTURAL="pass"
	else
		STRUCTURAL="fail"
		FIRMWARE_BOOTABLE=false
		rm -f "${img}"
		ARTIFACT_PATH=""; ARTIFACT_FORMAT=""; BOOT_METHOD=""; BOOT_MODE=""
		BLOCKERS+=("hybrid_strategy: dual-mode image failed structural verification (GPT/ESP/BOOTX64.EFI/BIOS GRUB)")
	fi
elif [[ "${FIRMWARE_BOOTABLE}" == true && "${ARTIFACT_FORMAT}" == "raw_mbr_bios_disk" ]]; then
	img="${ROOT}/${ARTIFACT_PATH}"
	ok=true
	# MBR boot signature 0x55AA.
	[[ "$(dd if="${img}" bs=1 skip=510 count=2 status=none | od -An -tx1 | tr -d ' \n')" == "55aa" ]] || ok=false
	# GRUB stage present in the first sectors (boot.img + embedded core.img).
	[[ "$(dd if="${img}" bs=512 count=2048 status=none 2>/dev/null | grep -a -c GRUB)" -ge 1 ]] || ok=false
	# Partition table parses and has the bootable ext2 partition at sector 2048.
	sfdisk -d "${img}" 2>/dev/null | grep -q 'start=[[:space:]]*2048, size=.*type=83, bootable' || ok=false
	if [[ "${ok}" == true ]]; then
		STRUCTURAL="pass"
	else
		STRUCTURAL="fail"
		FIRMWARE_BOOTABLE=false
		rm -f "${img}"
		ARTIFACT_PATH=""; ARTIFACT_FORMAT=""; BOOT_METHOD=""; BOOT_MODE=""
	fi
elif [[ "${FIRMWARE_BOOTABLE}" == true && "${ARTIFACT_FORMAT}" == "iso9660_eltorito" ]]; then
	# grub-mkrescue output is bootable by construction; sanity-check it is a real,
	# non-empty ISO with a boot catalog signature region.
	[[ -s "${ROOT}/${ARTIFACT_PATH}" ]] && STRUCTURAL="pass" || { STRUCTURAL="fail"; FIRMWARE_BOOTABLE=false; ARTIFACT_PATH=""; ARTIFACT_FORMAT=""; }
fi

# --- honest fallback: no firmware-bootable image emitted ---
if [[ "${FIRMWARE_BOOTABLE}" != true ]]; then
	# Prefer the real boot-chain tar from Slice 9.12; else the stub tar.
	if [[ -f "${BUILD_ROOT}/${REAL_TAR_NAME}" ]]; then
		ARTIFACT_PATH="distro/install-media/build/${REAL_TAR_NAME}"
		ARTIFACT_FORMAT="tar_real_bootchain_no_wrapper"
	elif [[ -f "${BUILD_ROOT}/vyntrio-install-media-NOT-BOOTABLE.stub.tar" ]]; then
		ARTIFACT_PATH="distro/install-media/build/vyntrio-install-media-NOT-BOOTABLE.stub.tar"
		ARTIFACT_FORMAT="tar_stub"
	fi
	# Concrete blockers, per strategy.
	if [[ "${BOOT_CHAIN}" != "real" ]]; then
		BLOCKERS+=("real_boot_chain_absent: run 'make install-media-image' on a host with /boot/vmlinuz-*, /boot/initrd.img-*, grub tools")
	fi
	if [[ "${HYBRID_TOOLS_OK}" != true ]]; then
		BLOCKERS+=("hybrid_strategy: missing tool(s)/modules: ${HYBRID_MISSING[*]} — apt: grub-efi-amd64-bin dosfstools mtools gdisk")
	elif [[ "${NONTMPFS_OK}" != true ]]; then
		BLOCKERS+=("hybrid_strategy: build dir is ${BUILD_FSTYPE}; needs non-tmpfs backing — set VYNTRIO_INSTALL_MEDIA_BUILD_ROOT")
	fi
	if [[ -z "${ISO_WRITER}" ]]; then
		BLOCKERS+=("iso_strategy: missing iso9660/eltorito writer (need one of: xorriso, xorrisofs, genisoimage) — apt: xorriso")
	fi
	if [[ "${RAW_TOOLS_OK}" != true ]]; then
		BLOCKERS+=("raw_strategy: missing tool(s): ${RAW_MISSING[*]} — apt: e2fsprogs (mke2fs/debugfs), util-linux (sfdisk), grub-pc-bin (grub-bios-setup)")
	elif [[ "${NONTMPFS_OK}" != true ]]; then
		BLOCKERS+=("raw_strategy: build dir is ${BUILD_FSTYPE}; grub-bios-setup needs a non-tmpfs backing filesystem — set VYNTRIO_INSTALL_MEDIA_BUILD_ROOT to a disk-backed path")
	elif [[ "${STRUCTURAL}" == "fail" ]]; then
		BLOCKERS+=("raw_strategy: image construction ran but failed structural verification (see grub-bios-setup output)")
	fi
	BLOCKERS+=("uefi: dual-mode hybrid image is the required product baseline")
fi

# --- provenance record (Slice 9.13) ---
{
	echo "# Generated by scripts/wrap-install-media-image.sh — do not commit"
	echo "schema_version: vyntrio-install-media-wrapper-v1"
	echo "slice: 9.13"
	echo "generated_at: ${GEN_UTC}"
	echo "media_role: install"
	echo "boot_chain: ${BOOT_CHAIN}"
	echo "firmware_bootable: ${FIRMWARE_BOOTABLE}"
	echo "image_wrapper: $([[ "${FIRMWARE_BOOTABLE}" == true ]] && echo present || echo absent)"
	echo "usb_creator: deferred"
	echo "dashboard_on_first_boot: false"
	echo "artifact: ${ARTIFACT_PATH}"
	echo "artifact_format: ${ARTIFACT_FORMAT}"
	if [[ "${FIRMWARE_BOOTABLE}" == true ]]; then
		echo "boot_method: ${BOOT_METHOD}"
		echo "firmware_boot_mode: ${BOOT_MODE}"
		echo "bios_support: $([[ "${BOOT_MODE}" == *bios* || "${BOOT_MODE}" == bios_legacy ]] && echo true || echo false)"
		echo "uefi_support: $([[ "${BOOT_MODE}" == *uefi* ]] && echo true || echo false)"
		echo "dual_mode: $([[ "${BOOT_MODE}" == "bios+uefi" ]] && echo true || echo false)"
		echo "product_baseline_complete: $([[ "${BOOT_MODE}" == "bios+uefi" ]] && echo true || echo false)"
		echo "structural_verification: ${STRUCTURAL}"
		echo "runtime_boot_tested: false"
		echo "runtime_boot_reason: no_qemu_or_vm_on_build_host"
		if [[ "${BOOT_MODE}" == "bios+uefi" ]]; then
			echo "blockers: none"
		else
			echo "blockers:"
			for b in "${BLOCKERS[@]}"; do echo "  - ${b}"; done
			[[ ${#BLOCKERS[@]} -eq 0 ]] && echo "  - uefi_incomplete: BIOS-only is not the product baseline"
		fi
	else
		echo "structural_verification: ${STRUCTURAL}"
		echo "bios_support: false"
		echo "uefi_support: false"
		echo "dual_mode: false"
		echo "product_baseline_complete: false"
		echo "blockers:"
		for b in "${BLOCKERS[@]}"; do echo "  - ${b}"; done
	fi
	echo "capabilities:"
	echo "  iso_writer: ${ISO_WRITER:-none}"
	echo "  grub_mkrescue: $(have grub-mkrescue && echo yes || echo no)"
	echo "  grub_x86_64_efi: $([[ -d "${GRUB_EFI_DIR}" ]] && echo yes || echo no)"
	echo "  hybrid_tools_ok: ${HYBRID_TOOLS_OK}"
	echo "  raw_tools_ok: ${RAW_TOOLS_OK}"
	echo "  build_fs_type: ${BUILD_FSTYPE}"
	echo "  nontmpfs_ok: ${NONTMPFS_OK}"
	echo "missing_for_usb_creator:"
	if [[ "${BOOT_MODE}" == "bios+uefi" ]]; then
		echo "  - complete_live_root_userland"
		echo "  - live_session_vyntrio_api_host"
		echo "  - host_usb_writer_tool"
	else
		echo "  - uefi_boot_support"
		echo "  - complete_live_root_userland"
		echo "  - live_session_vyntrio_api_host"
		echo "  - host_usb_writer_tool"
	fi
} >"${BUILD_ROOT}/WRAPPER.txt"

echo "install-media wrapper: boot_chain=${BOOT_CHAIN} firmware_bootable=${FIRMWARE_BOOTABLE} artifact=${ARTIFACT_PATH}"
if [[ "${FIRMWARE_BOOTABLE}" == true ]]; then
	echo "install-media wrapper: ${ARTIFACT_FORMAT} via ${BOOT_METHOD} (${BOOT_MODE}); structural=${STRUCTURAL}, runtime_boot_tested=false"
else
	echo "install-media wrapper: NOT firmware-bootable — see distro/install-media/build/WRAPPER.txt blockers" >&2
fi
