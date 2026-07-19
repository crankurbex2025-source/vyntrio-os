#!/usr/bin/env bash
# Verifies the image-initrd swap to the Vyntrio live initramfs (Slice 9.17).
# Independently re-proves that the firmware image now boots the live initramfs
# (not the host Debian initrd), asserts record honesty, exercises fail-closed +
# host-initrd restore, and checks prior slices are not regressed.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
BOOT="${ENVELOPE_ROOT}/boot"
REC="${BUILD_ROOT}/INITRD_SWAP.txt"
WRAP="${BUILD_ROOT}/WRAPPER.txt"
LIVE_INITRAMFS="${BUILD_ROOT}/vyntrio-live-initramfs.cpio.gz"
HOST_BACKUP="${BUILD_ROOT}/initrd.host-backup.img"

fail() { echo "installmedia initrd-swap test: $*" >&2; exit 1; }
rget() { sed -n "s/^ *$1: //p" "${REC}" | head -1; }
sha() { sha256sum "$1" 2>/dev/null | cut -d' ' -f1; }

[[ -f "${REC}" ]] || fail "INITRD_SWAP.txt missing; run 'make install-media-initrd-swap' first"
[[ "$(rget swap_status)" == "applied" ]] || fail "swap_status is not 'applied'"

LIVE_SHA_REC="$(rget sha256 | head -1)"   # first sha256 = initrd_before (host)
# Pull the specific fields we care about explicitly.
HOST_SHA="$(sed -n '/^initrd_before:/,/^initrd_after:/s/^  sha256: //p' "${REC}" | head -1)"
LIVE_SHA="$(sed -n '/^initrd_after:/,/^image:/s/^  sha256: //p' "${REC}" | head -1)"
IMG_SHA="$(rget image_initrd_sha256)"
FMT="$(rget artifact_format)"
ART="$(rget artifact)"

[[ -n "${HOST_SHA}" && -n "${LIVE_SHA}" ]] || fail "record missing before/after sha256"
[[ "${HOST_SHA}" != "${LIVE_SHA}" ]] || fail "before/after initrd sha are identical — no actual swap"

# --- the live initramfs on disk must match the recorded 'after' sha ---
[[ -f "${LIVE_INITRAMFS}" ]] || fail "live initramfs missing: ${LIVE_INITRAMFS}"
[[ "$(sha "${LIVE_INITRAMFS}")" == "${LIVE_SHA}" ]] || fail "live initramfs sha != recorded initrd_after sha"

# --- INDEPENDENT re-proof: extract the image's initrd and match the live one ---
case "${FMT}" in
raw_gpt_hybrid_disk)
	img="${ROOT}/${ART}"
	[[ -f "${img}" ]] || fail "hybrid image artifact missing: ${ART}"
	ESP_OFF_FILE="${BUILD_ROOT}/HYBRID_ESP_OFFSET.txt"
	[[ -f "${ESP_OFF_FILE}" ]] || fail "HYBRID_ESP_OFFSET.txt missing"
	ESP_OFF="$(tr -d '[:space:]' <"${ESP_OFF_FILE}")"
	if command -v mcopy >/dev/null 2>&1; then
		work="$(mktemp -d)"; trap 'rm -rf "${work}"' EXIT
		mcopy -i "${img}@@${ESP_OFF}" ::/boot/initrd.img "${work}/x" >/dev/null 2>&1 \
			|| fail "could not extract /boot/initrd.img from hybrid ESP"
		[[ -s "${work}/x" ]] || fail "extracted image initrd is empty"
		got="$(sha "${work}/x")"
		[[ "${got}" == "${LIVE_SHA}" ]] || fail "image initrd sha ${got} != live initramfs ${LIVE_SHA}"
		[[ "${got}" != "${HOST_SHA}" ]] || fail "image still boots the host Debian initrd"
		il="$(gzip -dc "${work}/x" 2>/dev/null | cpio -t 2>/dev/null || true)"
		grep -qE '^\.?/?init$' <<<"${il}" || fail "image initrd has no /init at root"
		grep -q 'bin/busybox' <<<"${il}" || fail "image initrd has no busybox (not the Vyntrio live runtime)"
		grep -qE 'lib/modules/[^/]+/modules\.dep' <<<"${il}" || fail "image initrd carries no kernel modules"
		rm -rf "${work}"; trap - EXIT
	fi
	;;
raw_mbr_bios_disk)
	img="${ROOT}/${ART}"
	[[ -f "${img}" ]] || fail "raw image artifact missing: ${ART}"
	if command -v debugfs >/dev/null 2>&1; then
		work="$(mktemp -d)"; trap 'rm -rf "${work}"' EXIT
		dd if="${img}" bs=1M skip=1 of="${work}/part.img" status=none 2>/dev/null || true
		debugfs -R "dump /boot/initrd.img ${work}/x" "${work}/part.img" >/dev/null 2>&1 \
			|| fail "could not extract /boot/initrd.img from the image"
		[[ -s "${work}/x" ]] || fail "extracted image initrd is empty"
		got="$(sha "${work}/x")"
		[[ "${got}" == "${LIVE_SHA}" ]] || fail "image initrd sha ${got} != live initramfs ${LIVE_SHA}"
		[[ "${got}" != "${HOST_SHA}" ]] || fail "image still boots the host Debian initrd"
		# Prove it is the Vyntrio live runtime, not a Debian initrd.
		il="$(gzip -dc "${work}/x" 2>/dev/null | cpio -t 2>/dev/null || true)"
		grep -qE '^\.?/?init$' <<<"${il}" || fail "image initrd has no /init at root"
		grep -q 'bin/busybox' <<<"${il}" || fail "image initrd has no busybox (not the Vyntrio live runtime)"
		grep -qE 'lib/modules/[^/]+/modules\.dep' <<<"${il}" || fail "image initrd carries no kernel modules"
		rm -rf "${work}"; trap - EXIT
	fi
	;;
iso9660_eltorito)
	[[ "$(sha "${BOOT}/initrd.img")" == "${LIVE_SHA}" ]] || fail "boot/initrd.img != live initramfs (ISO path)"
	;;
*) fail "unexpected artifact_format: ${FMT}" ;;
esac

# --- record self-consistency ---
[[ "${IMG_SHA}" == "${LIVE_SHA}" ]] || fail "record image_initrd_sha256 != initrd_after sha"
[[ "$(rget boots_live_initramfs)" == "true" ]] || fail "record must state boots_live_initramfs: true"
[[ "$(rget firmware_bootable)" == "true" ]] || fail "record must state firmware_bootable: true"
[[ "$(rget structural_verification)" == "pass" ]] || fail "record must state structural_verification: pass"

# --- honesty: no unearned runtime/dashboard claims ---
[[ "$(rget runtime_boot_tested)" == "false" ]] || fail "runtime_boot_tested must stay false (no VM)"
[[ "$(rget dashboard_reachable_on_boot)" == "false" ]] || fail "dashboard_reachable_on_boot must stay false"

# --- no regression: wrapper still reports a structurally verified bootable image ---
if [[ -f "${WRAP}" ]]; then
	[[ "$(sed -n 's/^firmware_bootable: //p' "${WRAP}")" == "true" ]] || fail "wrapper firmware_bootable regressed"
	grep -q '^structural_verification: pass' "${WRAP}" || fail "wrapper structural verification regressed"
fi

# --- no regression: payload allowlist + state-free live_root ---
mapfile -t pf < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
[[ "${#pf[@]}" -eq 6 ]] || fail "payload file count regressed (${#pf[@]} != 6)"
[[ ! -d "${ENVELOPE_ROOT}/payload/lib/modules" ]] || fail "kernel modules leaked into target-disk payload"
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			fail "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${ENVELOPE_ROOT}/live_root" -type f -print0)

# --- fail-closed: missing live initramfs must fail (before any swap) ---
tmp="$(mktemp -d)"; trap 'rm -rf "${tmp}"' EXIT
if VYNTRIO_INSTALL_ENVELOPE_ROOT="${tmp}/missing" \
	VYNTRIO_INSTALL_MEDIA_BUILD_ROOT="${tmp}/build" \
	bash "${ROOT}/scripts/swap-live-initramfs-into-image.sh" >/dev/null 2>&1; then
	fail "expected failure when live initramfs/boot chain missing"
fi

# --- fail-closed WITH restore: force wrapper failure post-swap, host initrd restored ---
# Hermetic copy of boot tree + a live initramfs; empty GRUB dir makes the raw
# wrapper fail, which must trigger restore of the host initrd and a non-zero exit.
if command -v debugfs >/dev/null 2>&1 && [[ -f "${HOST_BACKUP}" ]]; then
	henv="${tmp}/env"; hbuild="${tmp}/hbuild"; hgrub="${tmp}/emptygrub"
	mkdir -p "${henv}/boot/grub" "${hbuild}" "${hgrub}"
	cp -a "${BOOT}/vmlinuz" "${henv}/boot/vmlinuz"
	cp -a "${BOOT}/grub/grub.cfg" "${henv}/boot/grub/grub.cfg"
	cp -a "${HOST_BACKUP}" "${henv}/boot/initrd.img"          # hermetic "host" initrd
	cp -a "${LIVE_INITRAMFS}" "${hbuild}/vyntrio-live-initramfs.cpio.gz"
	host_ref_sha="$(sha "${henv}/boot/initrd.img")"
	if VYNTRIO_INSTALL_ENVELOPE_ROOT="${henv}" \
		VYNTRIO_INSTALL_MEDIA_BUILD_ROOT="${hbuild}" \
		VYNTRIO_GRUB_I386_PC_DIR="${hgrub}" \
		bash "${ROOT}/scripts/swap-live-initramfs-into-image.sh" >/dev/null 2>&1; then
		fail "expected non-zero exit when wrapper cannot produce a bootable image post-swap"
	fi
	# host initrd must have been restored (not left as the live initramfs)
	[[ "$(sha "${henv}/boot/initrd.img")" == "${host_ref_sha}" ]] \
		|| fail "restore-on-failure did not restore the host initrd"
	[[ "$(sed -n 's/^swap_status: //p' "${hbuild}/INITRD_SWAP.txt")" == "failed" ]] \
		|| fail "failed swap must record swap_status: failed"
fi

echo "installmedia initrd-swap test: ok (image boots Vyntrio live initramfs; host=${HOST_SHA:0:12} live=${LIVE_SHA:0:12}; runtime unverified)"
