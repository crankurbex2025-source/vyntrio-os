#!/usr/bin/env bash
# Stage the dual-mode (BIOS+UEFI) Vyntrio USB appliance image for local download.
# Refuses host-Debian-initrd stubs and requires APPLIANCE.txt from the appliance builder.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WRAPPER_REC="${ROOT}/distro/install-media/build/WRAPPER.txt"
APPLIANCE_REC="${ROOT}/distro/install-media/build/APPLIANCE.txt"
SWAP_REC="${ROOT}/distro/install-media/build/INITRD_SWAP.txt"
STAGING_DIR="${ROOT}/distro/release/staging"
MANIFEST_NAME="release-manifest.json"
PUBLIC_NAME="install-media-public.json"
VERSIONS_NAME="image-versions.json"
RECORD_NAME="STAGING.txt"

fail() { echo "stage-release-install-media: $*" >&2; exit 1; }
wfield() { sed -n "s/^$1: //p" "${WRAPPER_REC}" | head -n1; }
afield() { sed -n "s/^$1: //p" "${APPLIANCE_REC}" | head -n1; }

[[ -f "${WRAPPER_REC}" ]] || fail "missing ${WRAPPER_REC}; run 'make install-media-appliance' first"
[[ -f "${APPLIANCE_REC}" ]] || fail "missing ${APPLIANCE_REC}; refuse staging stub — build appliance image"
[[ -f "${SWAP_REC}" ]] || fail "missing ${SWAP_REC}; Vyntrio initramfs not proven in image"

grep -q '^firmware_bootable: true' "${WRAPPER_REC}" \
	|| fail "WRAPPER.txt reports firmware_bootable != true; refusing to stage"
grep -q '^product_baseline_complete: true' "${WRAPPER_REC}" \
	|| fail "WRAPPER.txt product_baseline_complete != true"
grep -q '^uefi_support: true' "${WRAPPER_REC}" \
	|| fail "WRAPPER.txt uefi_support != true; refusing to stage BIOS-only media"
grep -q '^dual_mode: true' "${WRAPPER_REC}" \
	|| fail "WRAPPER.txt dual_mode != true; refusing to stage"
grep -q '^host_debian_initrd: false' "${WRAPPER_REC}" \
	|| fail "WRAPPER.txt still uses host Debian initrd — not an appliance image"
grep -q '^swap_status: applied' "${SWAP_REC}" \
	|| fail "INITRD_SWAP.txt swap_status != applied"
grep -q '^appliance_rootfs_on_media: true' "${APPLIANCE_REC}" \
	|| fail "APPLIANCE.txt missing appliance_rootfs_on_media"
grep -q '^vyntrio_initramfs: true' "${APPLIANCE_REC}" \
	|| fail "APPLIANCE.txt missing vyntrio_initramfs"

ART_REL="$(wfield artifact)"
[[ -n "${ART_REL}" && -f "${ROOT}/${ART_REL}" ]] || fail "wrapper artifact missing: ${ART_REL}"
ARTIFACT_NAME="$(basename "${ART_REL}")"
BUILD_IMG="${ROOT}/${ART_REL}"
FW_MODE="$(wfield firmware_boot_mode)"
ARTIFACT_FORMAT="$(wfield artifact_format)"
BIOS_SUPPORT="$(wfield bios_support)"
UEFI_SUPPORT="$(wfield uefi_support)"
DUAL_MODE="$(wfield dual_mode)"
SECURE_BOOT="$(wfield secure_boot)"
[[ -n "${FW_MODE}" ]] || FW_MODE="bios+uefi"
[[ -n "${ARTIFACT_FORMAT}" ]] || ARTIFACT_FORMAT="raw_gpt_hybrid_appliance"
[[ -n "${BIOS_SUPPORT}" ]] || BIOS_SUPPORT="true"
[[ -n "${UEFI_SUPPORT}" ]] || UEFI_SUPPORT="true"
[[ -n "${DUAL_MODE}" ]] || DUAL_MODE="true"
[[ -n "${SECURE_BOOT}" ]] || SECURE_BOOT="unsupported"

mkdir -p "${STAGING_DIR}"
rm -f "${STAGING_DIR}/vyntrio-install-media-bios.img" \
	"${STAGING_DIR}/vyntrio-install-media-bios.img.sha256" 2>/dev/null || true
cp -f "${BUILD_IMG}" "${STAGING_DIR}/${ARTIFACT_NAME}"

SIZE_BYTES="$(stat -c '%s' "${STAGING_DIR}/${ARTIFACT_NAME}")"
SHA256="$(sha256sum "${STAGING_DIR}/${ARTIFACT_NAME}" | awk '{print $1}')"
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
VERSION="${VYNTRIO_RELEASE_VERSION:-0.2.0-dev}"
CHANNEL="${VYNTRIO_RELEASE_CHANNEL:-development}"
BUILD_ID="${VYNTRIO_RELEASE_BUILD_ID:-${VERSION}+$(date -u +%Y%m%d)}"

# Refuse obviously stub-sized appliance claims (< 400 MiB with firmware/rootfs expected).
MIN_APPLIANCE_BYTES=$((400 * 1024 * 1024))
if [[ "${SIZE_BYTES}" -lt "${MIN_APPLIANCE_BYTES}" ]]; then
	fail "artifact ${SIZE_BYTES} bytes is below appliance minimum (${MIN_APPLIANCE_BYTES}); refusing stub"
fi

# Merge published versions feed (keep older entries when present).
VERSIONS_PATH="${STAGING_DIR}/${VERSIONS_NAME}"
py_bool() { [[ "$1" == "true" ]] && echo True || echo False; }
BIOS_PY="$(py_bool "${BIOS_SUPPORT}")"
UEFI_PY="$(py_bool "${UEFI_SUPPORT}")"
DUAL_PY="$(py_bool "${DUAL_MODE}")"
python3 - <<PY
import json, os
path = "${VERSIONS_PATH}"
entry = {
    "version": "${VERSION}",
    "build_id": "${BUILD_ID}",
    "channel": "${CHANNEL}",
    "generated_at": "${GENERATED_AT}",
    "name": "${ARTIFACT_NAME}",
    "format": "${ARTIFACT_FORMAT}",
    "size_bytes": ${SIZE_BYTES},
    "sha256": "${SHA256}",
    "firmware_boot_mode": "${FW_MODE}",
    "bios_support": ${BIOS_PY},
    "uefi_support": ${UEFI_PY},
    "dual_mode": ${DUAL_PY},
    "secure_boot": "${SECURE_BOOT}",
    "download_available": True,
    "download_path": "/release/${ARTIFACT_NAME}",
    "support_status": "engineering_media_early_access",
    "latest": True,
    "media_role": "appliance",
}
items = []
if os.path.isfile(path):
    try:
        prev = json.load(open(path))
        items = prev.get("versions") or []
    except Exception:
        items = []
# Demote previous latest; drop duplicate sha.
out = []
for v in items:
    if v.get("sha256") == entry["sha256"]:
        continue
    v = dict(v)
    v["latest"] = False
    out.append(v)
out.insert(0, entry)
json.dump({"schema_version": "vyntrio-image-versions-v1", "versions": out}, open(path, "w"), indent=2)
print(f"versions={len(out)}")
PY

cat > "${STAGING_DIR}/${MANIFEST_NAME}" <<EOF
{
  "format_version": "vyntrio-release-manifest-v1",
  "created_at": "${GENERATED_AT}",
  "release": {
    "version": "${VERSION}",
    "channel": "${CHANNEL}",
    "build_id": "${BUILD_ID}"
  },
  "artifacts": [
    {
      "name": "vyntrio-install-media",
      "type": "archive",
      "relative_path": "${ARTIFACT_NAME}",
      "size_bytes": ${SIZE_BYTES},
      "sha256": "${SHA256}",
      "use": "appliance_usb_image"
    }
  ]
}
EOF

# Embed versions array into public metadata for the creator feed.
VERSIONS_JSON="$(python3 -c 'import json; print(json.dumps(json.load(open("'"${VERSIONS_PATH}"'"))["versions"]))')"

cat > "${STAGING_DIR}/${PUBLIC_NAME}" <<EOF
{
  "publication_status": "local_staging",
  "generated_at": "${GENERATED_AT}",
  "release": {
    "version": "${VERSION}",
    "channel": "${CHANNEL}",
    "build_id": "${BUILD_ID}"
  },
  "primary_artifact": {
    "name": "${ARTIFACT_NAME}",
    "format": "${ARTIFACT_FORMAT}",
    "firmware_boot_mode": "${FW_MODE}",
    "bios_support": ${BIOS_SUPPORT},
    "uefi_support": ${UEFI_SUPPORT},
    "dual_mode": ${DUAL_MODE},
    "secure_boot": "${SECURE_BOOT}",
    "media_role": "appliance",
    "size_bytes": ${SIZE_BYTES},
    "sha256": "${SHA256}",
    "download_available": true,
    "download_path": "/release/${ARTIFACT_NAME}",
    "manifest_path": "/release/${MANIFEST_NAME}"
  },
  "image_versions": ${VERSIONS_JSON},
  "build_target": "make install-media-appliance",
  "stage_target": "make release-install-media-stage",
  "verify_command": "vyntrio-verify-artifact --base-dir distro/release/staging distro/release/staging/${MANIFEST_NAME}",
  "support_status": "engineering_media_early_access",
  "writer": {
    "name": "vyntrio-media-creator",
    "kind": "native_desktop_tauri",
    "platforms": ["linux", "windows"],
    "binary_name": "vyntrio-media-creator",
    "build_target": "make build-media-creator-native",
    "package_target": "make package-media-creator-native",
    "documentation_path": "docs/ops/install-media-writer.md",
    "requires_elevation": true,
    "native_gui": true,
    "gui_available": true,
    "gui_kind": "tauri"
  },
  "limitations": [
    "USB appliance image: OS + config/state persistence designed for the boot USB (Unraid-style product model; Vyntrio-owned implementation)",
    "Dual-mode GPT: UEFI mandatory path (ESP BOOTX64.EFI) + BIOS/legacy fallback",
    "Secure Boot is unsupported (not signed/enrolled)",
    "Native Media Creator: Windows NSIS + Linux .deb/.AppImage staged; macOS .app/.dmg blocked (no macOS build host) — CLI helpers may still be present",
    "Runtime boot to dashboard is wired in the image but not VM-proven on every build host",
    "Storage pool mutation, Docker/containers, and VMs are not claimed complete in this image",
    "Local staging only — set VYNTRIO_RELEASE_STAGING_DIR to expose /release/* downloads",
    "Ed25519 release signatures are not verified in v1"
  ]
}
EOF

cat > "${STAGING_DIR}/${RECORD_NAME}" <<EOF
# Generated by scripts/stage-release-install-media.sh — do not commit
schema_version: vyntrio-release-staging-v1
generated_at: ${GENERATED_AT}
publication_status: local_staging
media_role: appliance
artifact: distro/release/staging/${ARTIFACT_NAME}
size_bytes: ${SIZE_BYTES}
sha256: ${SHA256}
firmware_bootable: true
firmware_boot_mode: ${FW_MODE}
artifact_format: ${ARTIFACT_FORMAT}
bios_support: ${BIOS_SUPPORT}
uefi_support: ${UEFI_SUPPORT}
dual_mode: ${DUAL_MODE}
secure_boot: ${SECURE_BOOT}
product_baseline_complete: true
host_debian_initrd: false
vyntrio_initramfs: true
appliance_rootfs_on_media: true
EOF

echo "stage-release-install-media: staged appliance ${STAGING_DIR}/${ARTIFACT_NAME} (${SIZE_BYTES} bytes, ${FW_MODE}, secure_boot=${SECURE_BOOT})"
