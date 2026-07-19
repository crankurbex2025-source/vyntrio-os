#!/usr/bin/env bash
# Verifies the firmware-bootable wrapper / honest fallback (Slice 9.13).
# Distinguishes three axes and asserts the record matches actual content:
#   - boot chain: real vs stub
#   - firmware_bootable: true vs false
#   - image wrapper: present vs absent
# Does not regress the Slice 9.11/9.12 invariants.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
BOOT="${ENVELOPE_ROOT}/boot"
REC="${BUILD_ROOT}/WRAPPER.txt"

fail() { echo "installmedia wrapper test: $*" >&2; exit 1; }
field() { sed -n "s/^$1: //p" "${REC}"; }

[[ -f "${REC}" ]] || fail "WRAPPER.txt missing; run 'make install-media-wrap' first"

BOOT_CHAIN="$(field boot_chain)"
FW="$(field firmware_bootable)"
WRAP="$(field image_wrapper)"
ARTIFACT="$(field artifact)"
FMT="$(field artifact_format)"
[[ -n "${BOOT_CHAIN}" ]] || fail "WRAPPER.txt missing boot_chain"
[[ -n "${FW}" ]] || fail "WRAPPER.txt missing firmware_bootable"
[[ -n "${WRAP}" ]] || fail "WRAPPER.txt missing image_wrapper"

# --- axis 1: boot chain real vs stub must match envelope content ---
case "${BOOT_CHAIN}" in
real)
	[[ -f "${BOOT}/vmlinuz" && -f "${BOOT}/initrd.img" ]] || fail "boot_chain=real but kernel/initrd missing"
	if command -v grub-file >/dev/null 2>&1; then
		grub-file --is-x86-linux "${BOOT}/vmlinuz" 2>/dev/null || fail "boot_chain=real but boot/vmlinuz is not x86 Linux"
	fi
	[[ ! -e "${BOOT}/vmlinuz.stub" ]] || fail "boot_chain=real but stub kernel still present"
	;;
stub)
	[[ -f "${BOOT}/vmlinuz.stub" ]] || fail "boot_chain=stub but stub kernel missing"
	;;
*) fail "unknown boot_chain: ${BOOT_CHAIN}" ;;
esac

# --- axis 2 + 3: firmware_bootable vs image wrapper coherence ---
if [[ "${FW}" == "true" ]]; then
	[[ "${WRAP}" == "present" ]] || fail "firmware_bootable=true but image_wrapper != present"
	[[ -n "${ARTIFACT}" && -f "${ROOT}/${ARTIFACT}" ]] || fail "firmware_bootable=true but artifact missing: ${ARTIFACT}"
	grep -q '^structural_verification: pass' "${REC}" || fail "firmware_bootable=true requires structural_verification: pass"
	# Honesty: must not silently claim a runtime boot it never ran.
	grep -q '^runtime_boot_tested: ' "${REC}" || fail "firmware_bootable=true must declare runtime_boot_tested"

	case "${FMT}" in
	raw_gpt_hybrid_disk)
		[[ "$(field blockers)" == "none" ]] || fail "dual-mode hybrid must have blockers: none"
		img="${ROOT}/${ARTIFACT}"
		[[ "${ARTIFACT}" == *.img ]] || fail "hybrid format but artifact is not a .img"
		sig="$(dd if="${img}" bs=1 skip=510 count=2 status=none | od -An -tx1 | tr -d ' \n')"
		[[ "${sig}" == "55aa" ]] || fail "hybrid image missing MBR 0x55AA signature (got ${sig})"
		[[ "$(dd if="${img}" bs=512 skip=1 count=1 status=none | head -c 8)" == "EFI PART" ]] \
			|| fail "hybrid image missing GPT EFI PART signature"
		grep -q '^firmware_boot_mode: bios+uefi' "${REC}" || fail "hybrid image must declare firmware_boot_mode: bios+uefi"
		grep -q '^uefi_support: true' "${REC}" || fail "hybrid image must declare uefi_support: true"
		grep -q '^dual_mode: true' "${REC}" || fail "hybrid image must declare dual_mode: true"
		grep -q '^product_baseline_complete: true' "${REC}" || fail "hybrid image must declare product_baseline_complete: true"
		;;
	raw_mbr_bios_disk)
		grep -q '^blockers:' "${REC}" || fail "BIOS-only image must list incompleteness blockers"
		img="${ROOT}/${ARTIFACT}"
		[[ "${ARTIFACT}" == *.img ]] || fail "raw format but artifact is not a .img"
		sig="$(dd if="${img}" bs=1 skip=510 count=2 status=none | od -An -tx1 | tr -d ' \n')"
		[[ "${sig}" == "55aa" ]] || fail "raw image missing MBR 0x55AA boot signature (got ${sig})"
		[[ "$(dd if="${img}" bs=512 count=2048 status=none 2>/dev/null | grep -a -c GRUB)" -ge 1 ]] \
			|| fail "raw image has no GRUB stage in the embed area"
		if command -v sfdisk >/dev/null 2>&1; then
			sfdisk -d "${img}" 2>/dev/null | grep -q 'start=[[:space:]]*2048, size=.*type=83, bootable' \
				|| fail "raw image missing bootable ext2 partition at sector 2048"
		fi
		grep -q '^product_baseline_complete: false' "${REC}" \
			|| fail "BIOS-only image must declare product_baseline_complete: false"
		grep -q '^uefi_support: false' "${REC}" || fail "BIOS-only image must declare uefi_support: false"
		;;
	iso9660_eltorito)
		[[ "$(field blockers)" == "none" ]] || fail "ISO dual-mode must have blockers: none"
		[[ "${ARTIFACT}" == *.iso ]] || fail "iso format but artifact is not a .iso"
		[[ -s "${ROOT}/${ARTIFACT}" ]] || fail "iso artifact empty"
		;;
	*) fail "firmware_bootable=true with unexpected artifact_format: ${FMT}" ;;
	esac
else
	# Honest fallback: no wrapper, concrete blockers, real boot chain preserved.
	[[ "${WRAP}" == "absent" ]] || fail "firmware_bootable=false but image_wrapper != absent"
	grep -q '^blockers:' "${REC}" || fail "fallback must record a blockers list"
	[[ "$(grep -cE '^  - ' "${REC}")" -ge 1 ]] || fail "fallback must name at least one concrete blocker"
	# A named blocker must point at a real missing capability (tool or fs), not a vague note.
	grep -qE '^  - (iso_strategy|raw_strategy|real_boot_chain_absent|uefi):' "${REC}" \
		|| fail "fallback blockers must identify the missing tool/capability"
	# If a fallback artifact is present it must be a non-bootable tar that says so.
	if [[ -n "${ARTIFACT}" && "${ARTIFACT}" == *.tar ]]; then
		tar -tf "${ROOT}/${ARTIFACT}" | grep -q 'NOT_BOOTABLE.txt' || fail "fallback tar missing NOT_BOOTABLE.txt"
	fi
fi

# --- no-regression: 9.11/9.12 envelope invariants ---
readonly -a EXPECTED_PAYLOAD=(
	"usr/bin/vyntrio-api" "usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service" "usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf" "etc/vyntrio/config.toml"
)
for rel in "${EXPECTED_PAYLOAD[@]}"; do
	[[ -f "${ENVELOPE_ROOT}/payload/${rel}" ]] || fail "missing payload: ${rel}"
done
mapfile -t pf < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
[[ "${#pf[@]}" -eq "${#EXPECTED_PAYLOAD[@]}" ]] || fail "unexpected payload file count"
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			fail "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${ENVELOPE_ROOT}/live_root" -type f -print0)

# --- fail-closed: wrapper must reject a missing boot layer ---
tmp="$(mktemp -d)"; trap 'rm -rf "${tmp}"' EXIT
if VYNTRIO_INSTALL_ENVELOPE_ROOT="${tmp}/missing" \
	VYNTRIO_INSTALL_MEDIA_BUILD_ROOT="${tmp}/build" \
	bash "${ROOT}/scripts/wrap-install-media-image.sh" >/dev/null 2>&1; then
	fail "expected failure when boot layer missing"
fi

echo "installmedia wrapper test: ok (boot_chain=${BOOT_CHAIN}, firmware_bootable=${FW}, image_wrapper=${WRAP})"
