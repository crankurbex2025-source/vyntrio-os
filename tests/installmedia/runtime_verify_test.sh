#!/usr/bin/env bash
# Verifies Stage-2 runtime boot verification (Slice 10.1).
# Asserts: the verifier is READ-ONLY (mutates no image/initrd/live-rootfs), the
# record is honest (no faked boot/dashboard claims), it fails closed with exact
# blockers when no harness exists, and Stage 1 is not regressed.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
BOOT="${ENVELOPE_ROOT}/boot"
LR="${ENVELOPE_ROOT}/live_root"
REC="${BUILD_ROOT}/RUNTIME_VERIFY.txt"
SWAP_REC="${BUILD_ROOT}/INITRD_SWAP.txt"
SCRIPT="${ROOT}/scripts/verify-live-boot-runtime.sh"

fail() { echo "installmedia runtime-verify test: $*" >&2; exit 1; }
rget() { sed -n "s/^ *$1: //p" "${REC}" | head -1; }
sha() { sha256sum "$1" 2>/dev/null | cut -d' ' -f1; }

[[ -f "${SWAP_REC}" ]] || fail "INITRD_SWAP.txt missing; run 'make install-media-initrd-swap' first"
[[ -f "${BOOT}/initrd.img" ]] || fail "boot/initrd.img missing"

IMG="${BUILD_ROOT}/vyntrio-install-media.img"

# --- fingerprint the boot/image/live-rootfs BEFORE running the verifier ------
before_initrd="$(sha "${BOOT}/initrd.img")"
before_kernel="$(sha "${BOOT}/vmlinuz")"
before_grubcfg="$(sha "${BOOT}/grub/grub.cfg")"
before_img=""; [[ -f "${IMG}" ]] && before_img="$(sha "${IMG}")"
before_lr="$( (cd "${LR}" && find . -type f -exec sha256sum {} + 2>/dev/null | LC_ALL=C sort | sha256sum) | cut -d' ' -f1)"

# --- run the verifier; tolerate its fail-closed non-zero exit ---------------
set +e
"${SCRIPT}" >/dev/null 2>&1
verify_rc=$?
set -e

# --- fingerprint AFTER; the verifier must have changed NOTHING in the chain --
after_initrd="$(sha "${BOOT}/initrd.img")"
after_kernel="$(sha "${BOOT}/vmlinuz")"
after_grubcfg="$(sha "${BOOT}/grub/grub.cfg")"
after_img=""; [[ -f "${IMG}" ]] && after_img="$(sha "${IMG}")"
after_lr="$( (cd "${LR}" && find . -type f -exec sha256sum {} + 2>/dev/null | LC_ALL=C sort | sha256sum) | cut -d' ' -f1)"

[[ "${before_initrd}" == "${after_initrd}" ]] || fail "verifier modified boot/initrd.img (must be read-only)"
[[ "${before_kernel}" == "${after_kernel}" ]] || fail "verifier modified boot/vmlinuz (must be read-only)"
[[ "${before_grubcfg}" == "${after_grubcfg}" ]] || fail "verifier modified grub.cfg (must be read-only)"
[[ "${before_img}" == "${after_img}" ]] || fail "verifier modified the firmware image (must be read-only)"
[[ "${before_lr}" == "${after_lr}" ]] || fail "verifier modified live_root contents (must be read-only)"

# --- record must exist and declare no chain modification --------------------
[[ -f "${REC}" ]] || fail "RUNTIME_VERIFY.txt not written"
[[ "$(rget chain_modified_by_this_slice)" == "false" ]] || fail "record must declare chain_modified_by_this_slice: false"
[[ "$(rget slice)" == '"10.1"' ]] || fail "record slice must be 10.1"

STATUS="$(rget status)"
HARNESS="$(rget harness)"
RUNTIME_TESTED="$(rget runtime_boot_tested)"
DASH_REACHABLE="$(rget dashboard_reachable_on_boot)"

# --- honesty rules ----------------------------------------------------------
case "${STATUS}" in
dashboard_reachable)
	# Only legitimate with a real booted-guest HTTP 200 and a successful exit.
	[[ "${verify_rc}" -eq 0 ]] || fail "status=dashboard_reachable but verifier exit was ${verify_rc}"
	[[ "${RUNTIME_TESTED}" == "true" ]] || fail "dashboard_reachable requires runtime_boot_tested=true"
	[[ "${DASH_REACHABLE}" == "true" ]] || fail "status=dashboard_reachable but dashboard_reachable_on_boot!=true"
	[[ "$(rget dashboard_http_status)" == "200" ]] || fail "dashboard_reachable requires dashboard_http_status: 200"
	[[ "$(rget dashboard_probe_method)" == "booted_guest_http_hostfwd" ]] || fail "dashboard proof must come from the booted guest, not chroot/structural"
	;;
boot_without_dashboard|no_boot_markers|blocked|stage1_regression)
	# Must fail closed and never claim dashboard-on-boot.
	[[ "${verify_rc}" -ne 0 ]] || fail "status=${STATUS} must fail closed (non-zero exit)"
	[[ "${DASH_REACHABLE}" == "false" ]] || fail "status=${STATUS} but dashboard_reachable_on_boot!=false"
	grep -qE '^  - [a-z]' "${REC}" || fail "status=${STATUS} must record at least one concrete blocker"
	;;
*) fail "unknown status: ${STATUS}" ;;
esac

# --- dashboard reachable must never be true without an HTTP 200 from a boot --
if [[ "${DASH_REACHABLE}" == "true" ]]; then
	[[ "$(rget dashboard_http_status)" == "200" ]] || fail "dashboard_reachable_on_boot=true without HTTP 200 is a faked proof"
	[[ "${HARNESS}" == "qemu" ]] || fail "dashboard reachable claimed without a boot harness"
fi

# --- when no harness, blockers must name it exactly (this host) --------------
if [[ "${HARNESS}" == "none" ]]; then
	[[ "${STATUS}" == "blocked" ]] || fail "no harness but status is not 'blocked'"
	grep -q '^  - no_vm_harness:' "${REC}" || fail "missing exact no_vm_harness blocker"
	grep -q 'were NOT modified' "${REC}" || fail "must state the chain was not modified to compensate"
fi

# --- Stage 1 still intact (read-only assertion) -----------------------------
[[ "$(sed -n 's/^  boots_live_initramfs: //p' "${SWAP_REC}" | head -1)" == "true" ]] \
	|| fail "Stage 1 regressed: image no longer boots the live initramfs"
[[ "$(rget stage1_boot_to_runtime_intact)" == "true" ]] || fail "record reports Stage 1 not intact"

# --- no-regression: payload allowlist + state-free live_root ----------------
mapfile -t pf < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
[[ "${#pf[@]}" -eq 6 ]] || fail "payload file count regressed (${#pf[@]} != 6)"
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			fail "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${LR}" -type f -print0)

echo "installmedia runtime-verify test: ok (harness=${HARNESS}, status=${STATUS}, verifier_exit=${verify_rc}, read-only confirmed)"
