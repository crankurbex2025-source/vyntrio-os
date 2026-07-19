#!/usr/bin/env bash
# Cross-compile Vyntrio CLI writer only.
# Native GUI packages are produced by scripts/package-media-creator-native.sh (Tauri).
# This script must NOT claim loopback localhost wizards as native desktop GUI.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="${ROOT}/distro/release/writer"
VERSION="${VYNTRIO_RELEASE_VERSION:-0.2.0-dev}"
GO="${GO:-go}"

mkdir -p "${OUT}"
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

sha256_file() {
	local path="$1"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "${path}" > "${path}.sha256"
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "${path}" > "${path}.sha256"
	fi
}

build_binary() {
	local goos="$1"
	local goarch="$2"
	local out="$3"
	echo "package-write-media: building CLI ${goos}/${goarch} -> ${out}"
	GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 \
		"${GO}" build -trimpath -ldflags "-s -w" \
		-o "${out}" "${ROOT}/cmd/write-media"
}

build_binary linux amd64 "${TMP}/vyntrio-write-media-linux-amd64"
build_binary darwin arm64 "${TMP}/vyntrio-write-media-darwin-arm64"
build_binary darwin amd64 "${TMP}/vyntrio-write-media-darwin-amd64"
build_binary windows amd64 "${TMP}/vyntrio-write-media-windows-amd64.exe"

cp -f "${TMP}/vyntrio-write-media-linux-amd64" "${OUT}/vyntrio-write-media-linux-amd64"
cp -f "${TMP}/vyntrio-write-media-darwin-arm64" "${OUT}/vyntrio-write-media-darwin-arm64"
cp -f "${TMP}/vyntrio-write-media-darwin-amd64" "${OUT}/vyntrio-write-media-darwin-amd64"
cp -f "${TMP}/vyntrio-write-media-windows-amd64.exe" "${OUT}/vyntrio-write-media-windows-amd64.exe"
sha256_file "${OUT}/vyntrio-write-media-linux-amd64"
sha256_file "${OUT}/vyntrio-write-media-darwin-arm64"
sha256_file "${OUT}/vyntrio-write-media-darwin-amd64"
sha256_file "${OUT}/vyntrio-write-media-windows-amd64.exe"

# Remove withdrawn loopback GUI artifacts if present.
rm -f \
	"${OUT}/vyntrio-media-creator-windows-amd64.exe" \
	"${OUT}/vyntrio-media-creator-windows-amd64.exe.sha256" \
	"${OUT}/vyntrio-media-creator-linux-amd64.tar.gz" \
	"${OUT}/vyntrio-media-creator-linux-amd64.tar.gz.sha256" \
	"${OUT}/vyntrio-media-creator-darwin-arm64" \
	"${OUT}/vyntrio-media-creator-darwin-arm64.sha256" \
	"${OUT}/vyntrio-media-creator-darwin-amd64" \
	"${OUT}/vyntrio-media-creator-darwin-amd64.sha256" \
	"${OUT}/"*.app.zip \
	2>/dev/null || true

cat > "${OUT}/README.txt" <<EOF
Vyntrio writer packages (${VERSION})

Native desktop Media Creator (Tauri):
  make package-media-creator-native
  -> vyntrio-media-creator-windows-amd64-setup.exe (NSIS)
  -> vyntrio-media-creator-linux-amd64.deb
  -> vyntrio-media-creator-linux-amd64.AppImage
  macOS .app/.dmg: blocked without a macOS build host

CLI writer (this script):
  vyntrio-write-media-{linux,darwin,windows}-*

Withdrawn as native GUI:
  Go loopback localhost wizard packages (not desktop parity)
EOF

STAGING_WRITER="${ROOT}/distro/release/staging/writer"
mkdir -p "${STAGING_WRITER}"
# Copy CLI only; do not delete native Tauri packages already staged.
for f in vyntrio-write-media-linux-amd64 vyntrio-write-media-darwin-arm64 \
	vyntrio-write-media-darwin-amd64 vyntrio-write-media-windows-amd64.exe README.txt; do
	[[ -f "${OUT}/${f}" ]] || continue
	cp -f "${OUT}/${f}" "${STAGING_WRITER}/${f}"
	[[ -f "${OUT}/${f}.sha256" ]] && cp -f "${OUT}/${f}.sha256" "${STAGING_WRITER}/${f}.sha256"
done
# Ensure withdrawn loopback names are gone from staging too.
rm -f \
	"${STAGING_WRITER}/vyntrio-media-creator-windows-amd64.exe" \
	"${STAGING_WRITER}/vyntrio-media-creator-windows-amd64.exe.sha256" \
	"${STAGING_WRITER}/vyntrio-media-creator-linux-amd64.tar.gz" \
	"${STAGING_WRITER}/vyntrio-media-creator-linux-amd64.tar.gz.sha256" \
	"${STAGING_WRITER}/vyntrio-media-creator-darwin-arm64" \
	"${STAGING_WRITER}/vyntrio-media-creator-darwin-arm64.sha256" \
	"${STAGING_WRITER}/vyntrio-media-creator-darwin-amd64" \
	"${STAGING_WRITER}/vyntrio-media-creator-darwin-amd64.sha256" \
	2>/dev/null || true

echo "package-write-media: CLI packages ready under ${OUT} (native GUI via package-media-creator-native)"
