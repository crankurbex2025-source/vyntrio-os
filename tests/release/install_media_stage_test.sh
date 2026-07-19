#!/usr/bin/env bash
# Verifies release install-media staging output.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
STAGING="${ROOT}/distro/release/staging"
ARTIFACT="${STAGING}/vyntrio-install-media.img"
MANIFEST="${STAGING}/release-manifest.json"
PUBLIC="${STAGING}/install-media-public.json"
RECORD="${STAGING}/STAGING.txt"

fail() { echo "release install media stage test: $*" >&2; exit 1; }

[[ -f "${ARTIFACT}" ]] || fail "missing staged artifact"
[[ -f "${MANIFEST}" ]] || fail "missing release-manifest.json"
[[ -f "${PUBLIC}" ]] || fail "missing install-media-public.json"
[[ -f "${RECORD}" ]] || fail "missing STAGING.txt"

"${ROOT}/bin/vyntrio-verify-artifact" --base-dir "${STAGING}" "${MANIFEST}" \
  || fail "manifest verification failed"

grep -q '^publication_status: local_staging' "${RECORD}" || fail "STAGING.txt missing publication_status"
grep -q '^firmware_bootable: true' "${RECORD}" || fail "STAGING.txt missing firmware_bootable"
grep -q '^uefi_support: true' "${RECORD}" || fail "STAGING.txt missing uefi_support"
grep -q '^dual_mode: true' "${RECORD}" || fail "STAGING.txt missing dual_mode"
grep -q '"uefi_support": true' "${PUBLIC}" || fail "public metadata missing uefi_support"
grep -q '"dual_mode": true' "${PUBLIC}" || fail "public metadata missing dual_mode"
grep -q '"name": "vyntrio-install-media.img"' "${PUBLIC}" || fail "public metadata wrong artifact name"
[[ ! -f "${STAGING}/vyntrio-install-media-bios.img" ]] || fail "BIOS-only artifact must not remain staged as primary"

echo "release install media stage test: pass"
