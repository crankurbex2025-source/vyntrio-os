#!/usr/bin/env bash
# Verifies install-media boot chain + image emission (Slice 9.12).
# Distinguishes a real boot chain from the Slice 9.11 stub state and asserts the
# emitted provenance matches actual content. Does not regress 9.11 checks.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENVELOPE_ROOT="${ROOT}/distro/install-media/envelope"
BUILD_ROOT="${ROOT}/distro/install-media/build"
BOOT="${ENVELOPE_ROOT}/boot"
IMAGE_RECORD="${BUILD_ROOT}/IMAGE.txt"

fail() { echo "installmedia image test: $*" >&2; exit 1; }

[[ -d "${ENVELOPE_ROOT}" ]] || fail "envelope missing; run 'make install-media-image' first"
[[ -f "${IMAGE_RECORD}" ]] || fail "IMAGE.txt missing; run 'make install-media-image' first"

# Layers preserved (no regression on the 9.11 layout contract).
for layer in boot live_root payload; do
	[[ -d "${ENVELOPE_ROOT}/${layer}" ]] || fail "missing layer: ${layer}"
done

# Payload allowlist unchanged (no regression).
readonly -a EXPECTED_PAYLOAD=(
	"usr/bin/vyntrio-api"
	"usr/bin/vyntrio-backup"
	"etc/systemd/system/vyntrio-api.service"
	"usr/lib/sysusers.d/vyntrio.conf"
	"etc/tmpfiles.d/vyntrio.conf"
	"etc/vyntrio/config.toml"
)
for rel in "${EXPECTED_PAYLOAD[@]}"; do
	[[ -f "${ENVELOPE_ROOT}/payload/${rel}" ]] || fail "missing payload: ${rel}"
done
mapfile -t payload_files < <(find "${ENVELOPE_ROOT}/payload" -type f | LC_ALL=C sort)
[[ "${#payload_files[@]}" -eq "${#EXPECTED_PAYLOAD[@]}" ]] || fail "unexpected payload file count"

# No secrets/state ever staged into live_root (no regression).
while IFS= read -r -d '' p; do
	case "$(basename "${p}")" in
		*.db|*.sqlite|*.sqlite3|*credential*|*token*|*license*|*secret*)
			fail "forbidden live_root file: ${p#${ENVELOPE_ROOT}/}" ;;
	esac
done < <(find "${ENVELOPE_ROOT}/live_root" -type f -print0)

# firmware_bootable must be declared and honest.
grep -q '^firmware_bootable: ' "${IMAGE_RECORD}" || fail "IMAGE.txt missing firmware_bootable"
grep -q '^usb_creator: deferred' "${IMAGE_RECORD}" || fail "IMAGE.txt must keep usb_creator: deferred"

BOOT_CHAIN="$(sed -n 's/^boot_chain: //p' "${IMAGE_RECORD}")"
FW="$(sed -n 's/^firmware_bootable: //p' "${IMAGE_RECORD}")"
[[ -n "${BOOT_CHAIN}" ]] || fail "IMAGE.txt missing boot_chain"

case "${BOOT_CHAIN}" in
real)
	# Real boot chain: assets must actually be real, stubs must be gone.
	[[ -f "${BOOT}/vmlinuz" ]] || fail "real boot_chain but boot/vmlinuz missing"
	[[ -f "${BOOT}/initrd.img" ]] || fail "real boot_chain but boot/initrd.img missing"
	[[ ! -e "${BOOT}/vmlinuz.stub" && ! -e "${BOOT}/initrd.img.stub" ]] \
		|| fail "real boot_chain but 9.11 stub kernel/initrd still present"
	if command -v grub-file >/dev/null 2>&1; then
		grub-file --is-x86-linux "${BOOT}/vmlinuz" 2>/dev/null \
			|| fail "boot/vmlinuz is not a valid x86 Linux kernel"
	fi
	grep -Eq '^[[:space:]]*linux[[:space:]]' "${BOOT}/grub/grub.cfg" || fail "grub.cfg missing linux directive"
	grep -Eq '^[[:space:]]*initrd[[:space:]]' "${BOOT}/grub/grub.cfg" || fail "grub.cfg missing initrd directive"
	grep -q '\.stub' "${BOOT}/grub/grub.cfg" && fail "grub.cfg still references .stub assets"
	[[ -f "${BOOT}/grub/i386-pc/core.img" ]] || fail "real boot_chain but grub core.img missing"
	[[ "$(stat -c '%s' "${BOOT}/grub/i386-pc/core.img")" -gt 20000 ]] || fail "grub core.img implausibly small"
	grep -q '^  kernel_is_x86_linux: true' "${IMAGE_RECORD}" || fail "IMAGE.txt must record verified kernel"
	grep -q '^  grub_core_built: true' "${IMAGE_RECORD}" || fail "IMAGE.txt must record grub core build"
	;;
stub)
	# Honest fallback: 9.11 stub artifacts must still be present.
	[[ -f "${BOOT}/vmlinuz.stub" && -f "${BOOT}/initrd.img.stub" ]] \
		|| fail "stub boot_chain but 9.11 stub artifacts missing"
	[[ "${FW}" == "false" ]] || fail "stub boot_chain must not claim firmware_bootable"
	;;
*)
	fail "unknown boot_chain: ${BOOT_CHAIN}" ;;
esac

# firmware_bootable honesty: only true alongside a real ISO artifact + no blockers.
ARTIFACT="$(sed -n 's/^artifact: //p' "${IMAGE_RECORD}")"
[[ -n "${ARTIFACT}" ]] || fail "IMAGE.txt missing artifact"
[[ -f "${ROOT}/${ARTIFACT}" ]] || fail "declared artifact missing on disk: ${ARTIFACT}"
if [[ "${FW}" == "true" ]]; then
	[[ "${ARTIFACT}" == *.iso ]] || fail "firmware_bootable: true but artifact is not an .iso"
	grep -q '^  - none' "${IMAGE_RECORD}" || fail "firmware_bootable: true but blockers recorded"
else
	# Non-bootable output must name concrete blockers (honesty requirement).
	grep -q 'iso9660_or_eltorito_writer_missing' "${IMAGE_RECORD}" \
		|| fail "non-bootable image must record iso9660 writer blocker"
	# The real-bootchain tar must warn against flashing.
	if [[ "${ARTIFACT}" == *.tar ]]; then
		tar -tf "${ROOT}/${ARTIFACT}" | grep -q 'NOT_BOOTABLE.txt' \
			|| fail "non-bootable tar missing NOT_BOOTABLE.txt"
	fi
fi

# live_root minimal-runtime additions present.
for rel in "etc/os-release" "etc/fstab" "usr/lib/vyntrio/live-init.sh" "usr/lib/vyntrio/FIRST_BOOT.txt"; do
	[[ -e "${ENVELOPE_ROOT}/live_root/${rel}" ]] || fail "missing live_root addition: ${rel}"
done
[[ -x "${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio/live-init.sh" ]] || fail "live-init.sh must be executable"
grep -q 'dashboard_reachable: false' "${ENVELOPE_ROOT}/live_root/usr/lib/vyntrio/FIRST_BOOT.txt" \
	|| fail "FIRST_BOOT.txt must stay honest about the dashboard"

# Fail-closed: builder must reject a missing bootability foundation.
tmpdir="$(mktemp -d)"; trap 'rm -rf "${tmpdir}"' EXIT
if VYNTRIO_INSTALL_ENVELOPE_ROOT="${tmpdir}/missing" \
	VYNTRIO_INSTALL_MEDIA_BUILD_ROOT="${tmpdir}/build" \
	bash "${ROOT}/scripts/build-install-media-image.sh" >/dev/null 2>&1; then
	fail "expected failure when bootability foundation missing"
fi

echo "installmedia image test: ok (boot_chain=${BOOT_CHAIN}, firmware_bootable=${FW})"
