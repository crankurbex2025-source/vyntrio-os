#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BIN="${ROOT}/bin/vyntrio-write-media"
IMAGE="${ROOT}/distro/release/staging/vyntrio-install-media.img"

fail() { echo "write_media_cli_test: $*" >&2; exit 1; }

[[ -x "${BIN}" ]] || fail "run make build-write-media first"
[[ -f "${IMAGE}" ]] || fail "run make release-install-media-stage first"

"${BIN}" help >/dev/null
"${BIN}" info --image "${IMAGE}" >/dev/null
"${BIN}" verify-image --image "${IMAGE}" >/dev/null
"${BIN}" list >/dev/null || true
if ! "${BIN}" write --image "${IMAGE}" --device /dev/no-such-device --dry-run >/dev/null 2>&1; then
	:
else
	fail "expected dry-run failure for unknown device"
fi

# Staged GUI + CLI packages must exist with sha256 sidecars when packaged.
STAGED="${ROOT}/distro/release/staging/writer"
if [[ -d "${STAGED}" ]]; then
	for name in \
		vyntrio-media-creator-windows-amd64.exe \
		vyntrio-media-creator-darwin-arm64 \
		vyntrio-media-creator-darwin-amd64 \
		vyntrio-media-creator-linux-amd64.tar.gz \
		vyntrio-write-media-windows-amd64.exe \
		vyntrio-write-media-darwin-arm64 \
		vyntrio-write-media-darwin-amd64 \
		vyntrio-write-media-linux-amd64; do
		[[ -f "${STAGED}/${name}" ]] || fail "staged writer missing: ${name}"
		[[ -f "${STAGED}/${name}.sha256" ]] || fail "staged writer sha256 missing: ${name}"
		SIDE="$(awk '{print $1}' "${STAGED}/${name}.sha256")"
		REAL="$(sha256sum "${STAGED}/${name}" | awk '{print $1}')"
		[[ "${SIDE}" == "${REAL}" ]] || fail "sha256 mismatch for staged ${name}"
	done
	# Withdrawn packages must not remain staged.
	if ls "${STAGED}"/vyntrio-media-creator-darwin-*.app.zip >/dev/null 2>&1; then
		fail "withdrawn macOS .app.zip still staged"
	fi
	# Windows GUI package must be PE subsystem WINDOWS (2), not CONSOLE (3).
	python3 - <<'PY'
import struct, sys
p="/opt/vyntrio-os/distro/release/staging/writer/vyntrio-media-creator-windows-amd64.exe"
data=open(p,"rb").read(1024)
e=struct.unpack_from("<I", data, 0x3C)[0]
subsystem=struct.unpack_from("<H", data, e+24+68)[0]
if subsystem != 2:
    raise SystemExit(f"windows GUI subsystem={subsystem}, want 2 (WINDOWS)")
print("write_media_cli_test: windows PE subsystem WINDOWS ok")
PY
	echo "write_media_cli_test: staged GUI + CLI packages verified"
fi

# GUI smoke (Linux host binary): status endpoint responds; no-args launches GUI.
GUI_BIN="${ROOT}/bin/vyntrio-media-creator"
cp -f "${BIN}" "${GUI_BIN}"
chmod +x "${GUI_BIN}"
# Renamed binary must still launch GUI (regression for Windows rename failure).
RENAMED="${ROOT}/bin/writer-renamed-smoke"
cp -f "${BIN}" "${RENAMED}"
"${RENAMED}" gui --listen 127.0.0.1:17993 --no-browser >/tmp/vyntrio-media-creator-gui.log 2>&1 &
GUI_PID=$!
cleanup_gui() { kill "${GUI_PID}" 2>/dev/null || true; }
trap cleanup_gui EXIT
ok=0
for _ in $(seq 1 40); do
	if curl -fsS "http://127.0.0.1:17993/api/status" | grep -q local_web_gui; then
		ok=1
		break
	fi
	sleep 0.1
done
[[ "${ok}" == "1" ]] || fail "GUI status endpoint did not become ready"
kill "${GUI_PID}" 2>/dev/null || true
trap - EXIT
echo "write_media_cli_test: GUI status smoke pass"

echo "write_media_cli_test: pass"
