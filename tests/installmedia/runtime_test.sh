#!/usr/bin/env bash
# Verifies runtime boot verification + first-boot dashboard path (Slice 9.14).
# Asserts runtime_boot_tested is honest, the first-boot wiring is present and
# well-formed, and the firmware image is not regressed.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
LR="${ENVELOPE_ROOT}/live_root"
REC="${BUILD_ROOT}/RUNTIME_BOOT.txt"
WRAP="${BUILD_ROOT}/WRAPPER.txt"

fail() { echo "installmedia runtime test: $*" >&2; exit 1; }

[[ -f "${REC}" ]] || fail "RUNTIME_BOOT.txt missing; run 'make install-media-runtime' first"
[[ -f "${WRAP}" ]] || fail "WRAPPER.txt missing (Slice 9.13 regression?)"

rget() { sed -n "s/^$1: //p" "${REC}"; }
wget_() { sed -n "s/^$1: //p" "${WRAP}"; }

# --- runtime_boot_tested honesty ---
TESTED="$(rget runtime_boot_tested)"
RESULT="$(rget runtime_boot_result)"
[[ -n "${TESTED}" ]] || fail "RUNTIME_BOOT.txt missing runtime_boot_tested"
[[ -n "${RESULT}" ]] || fail "RUNTIME_BOOT.txt missing runtime_boot_result"
case "${TESTED}" in
false)
	# Skipped: a concrete reason is mandatory (no faking / no silent skip).
	grep -q '^runtime_boot_reason: ' "${REC}" || fail "runtime_boot_tested=false must record a concrete runtime_boot_reason"
	[[ "$(rget runtime_boot_reason)" != "" ]] || fail "runtime_boot_reason must be non-empty"
	[[ "${RESULT}" == "skipped" || "${RESULT}" == "no_boot_markers_observed" ]] || fail "unexpected result for untested boot: ${RESULT}"
	;;
true)
	# Actually ran a VM: harness + markers must be recorded, result must be a real outcome.
	grep -q '^vm_harness: ' "${REC}" || fail "runtime_boot_tested=true must record vm_harness"
	grep -q '^boot_markers: ' "${REC}" || fail "runtime_boot_tested=true must record boot_markers"
	[[ "${RESULT}" == "boot_reached_kernel" || "${RESULT}" == "no_boot_markers_observed" ]] || fail "unexpected result: ${RESULT}"
	;;
*) fail "unknown runtime_boot_tested: ${TESTED}" ;;
esac

# --- consistency with the firmware wrapper (no divergence) ---
[[ "$(rget firmware_bootable)" == "$(wget_ firmware_bootable)" ]] \
	|| fail "firmware_bootable disagrees between RUNTIME_BOOT.txt and WRAPPER.txt"

# --- first-boot dashboard path present + well-formed ---
FB="${LR}/usr/lib/vyntrio/firstboot.sh"
SVC="${LR}/etc/systemd/system/vyntrio-firstboot.service"
LI="${LR}/usr/lib/vyntrio/live-init.sh"
[[ -x "${FB}" ]] || fail "firstboot.sh missing or not executable"
[[ -f "${SVC}" ]] || fail "vyntrio-firstboot.service missing"
[[ -x "${LI}" ]] || fail "live-init.sh missing or not executable"

bash -n "${FB}" || fail "firstboot.sh has a syntax error"
bash -n "${LI}" || fail "live-init.sh has a bash syntax error"

# --- Slice 10.2: firstboot.sh must run in the busybox live runtime (no bash) ---
head -1 "${FB}" | grep -q '^#!/bin/sh$' \
	|| fail "firstboot.sh must use '#!/bin/sh' (live runtime ships busybox, not bash)"
grep -q '\[\[' "${FB}" && fail "firstboot.sh must avoid bash-only [[ ]] (busybox-sh compatibility)"
if command -v dash >/dev/null 2>&1; then
	dash -n "${FB}" || fail "firstboot.sh is not POSIX/busybox-sh compatible (dash -n failed)"
fi
# --- Slice 10.2: the local dashboard must be supervised (respawn), not exec-once ---
grep -q 'MAX_RESTARTS' "${FB}" || fail "firstboot.sh must supervise the dashboard (bounded respawn)"
grep -qE 'while[[:space:]]*:' "${FB}" || fail "firstboot.sh must keep the WebGUI up via a supervise loop"

grep -q 'vyntrio-api' "${FB}" || fail "firstboot.sh must launch vyntrio-api"
grep -q '8080' "${FB}" || fail "firstboot.sh must reference the dashboard port"
grep -q '^  dashboard_primary_surface: true' "${REC}" || fail "RUNTIME_BOOT.txt must record dashboard_primary_surface: true"
grep -q '^  dashboard_supervised: true' "${REC}" || fail "RUNTIME_BOOT.txt must record dashboard_supervised: true"
grep -q 'firstboot.sh' "${LI}" || fail "live-init.sh must invoke the first-boot entrypoint"
grep -q 'firstboot.sh' "${SVC}" || fail "service must run firstboot.sh"
grep -q 'WantedBy=multi-user.target' "${SVC}" || fail "service must be enable-able on first boot"

# --- dashboard honesty: not claimed reachable yet ---
grep -q '^  dashboard_reachable: false' "${REC}" || fail "RUNTIME_BOOT.txt must keep dashboard_reachable: false"
grep -q 'dashboard_reachable: false' "${LR}/usr/lib/vyntrio/FIRST_BOOT.txt" \
	|| fail "FIRST_BOOT.txt must keep dashboard_reachable: false"

# --- Slice 10.4: TLS readiness replaces the prior LAN/TLS config blocker ---
grep -q '^  tls_readiness: ready' "${REC}" || fail "RUNTIME_BOOT.txt must record tls_readiness: ready"
grep -q '^tls_readiness: ready' "${LR}/usr/lib/vyntrio/FIRST_BOOT.txt" \
	|| fail "FIRST_BOOT.txt must record tls_readiness: ready"
grep -q 'lan_exposure_blocked_on: tls_required_for_non_loopback_bind' "${REC}" \
	&& fail "RUNTIME_BOOT.txt must not keep the old TLS config blocker (Slice 10.4)"
PREP="${LR}/usr/lib/vyntrio/prepare-live-dashboard-tls.sh"
[[ -x "${PREP}" ]] || fail "prepare-live-dashboard-tls.sh missing or not executable"
head -1 "${PREP}" | grep -q '^#!/bin/sh$' \
	|| fail "prepare-live-dashboard-tls.sh must use '#!/bin/sh' (busybox compatibility)"
grep -q 'prepare-live-dashboard-tls.sh' "${FB}" || fail "firstboot.sh must invoke TLS runtime prep"
grep -q 'VYNTRIO_LAN_BIND_IP' "${FB}" || fail "firstboot.sh must document VYNTRIO_LAN_BIND_IP LAN bind path"

# --- Slice 10.3: first-boot onboarding clarity ---
grep -q 'Vyntrio first-boot setup' "${FB}" || fail "firstboot.sh must show a clear first-boot setup prompt"
grep -q 'Open the dashboard in a browser on this device' "${FB}" || fail "firstboot.sh must instruct the user to open the local dashboard"
grep -q 'Storage, services, and remote access are added later' "${FB}" || fail "firstboot.sh must defer storage/services/remote"
grep -q '^first_boot_onboarding_clarity: clear' "${LR}/usr/lib/vyntrio/FIRST_BOOT.txt" \
	|| fail "FIRST_BOOT.txt must record onboarding clarity"
grep -q '^local_onboarding_steps:' "${LR}/usr/lib/vyntrio/FIRST_BOOT.txt" \
	|| fail "FIRST_BOOT.txt must list the local onboarding steps"
grep -q '^  first_boot_onboarding_clarity: clear' "${REC}" \
	|| fail "RUNTIME_BOOT.txt must record onboarding clarity"

# --- no regression: firmware image still bootable + structurally valid ---
[[ "$(wget_ firmware_bootable)" == "true" ]] || fail "firmware image regressed (WRAPPER firmware_bootable != true)"
grep -q '^structural_verification: pass' "${WRAP}" || fail "firmware image structural verification regressed"
IMG_REL="$(wget_ artifact)"
if [[ "${IMG_REL}" == *.img ]]; then
	sig="$(dd if="${ROOT}/${IMG_REL}" bs=1 skip=510 count=2 status=none | od -An -tx1 | tr -d ' \n')"
	[[ "${sig}" == "55aa" ]] || fail "firmware raw image lost its MBR boot signature"
fi

# --- no regression: payload allowlist + no secrets in live_root ---
readonly -a EXPECTED_PAYLOAD=(
	"usr/bin/vyntrio-api" "usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service" "usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf" "etc/vyntrio/config.toml"
)
mapfile -t pf < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
[[ "${#pf[@]}" -eq "${#EXPECTED_PAYLOAD[@]}" ]] || fail "payload file count regressed"
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			fail "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${LR}" -type f -print0)

# --- fail-closed: script rejects a missing wrapper/live_root ---
tmp="$(mktemp -d)"; trap 'rm -rf "${tmp}"' EXIT
if VYNTRIO_INSTALL_ENVELOPE_ROOT="${tmp}/missing" \
	VYNTRIO_INSTALL_MEDIA_BUILD_ROOT="${tmp}/build" \
	bash "${ROOT}/scripts/verify-runtime-boot.sh" >/dev/null 2>&1; then
	fail "expected failure when live_root/WRAPPER missing"
fi

echo "installmedia runtime test: ok (runtime_boot_tested=${TESTED}, result=${RESULT}, first-boot path wired)"
