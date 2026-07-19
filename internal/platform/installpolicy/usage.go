package installpolicy

import "strings"

// WorkflowText returns the Block 10 internal infrastructure policy summary.
func WorkflowText() string {
	return strings.TrimSpace(`
Vyntrio Block 10 installer infrastructure (` + PolicyVersion + `)

PRIMARY product path (USB-first — see docs/00_PROJECT.md):
  USB creator → bootable install medium → boot target hardware →
  local browser dashboard → live-session install → first-boot bootstrap

THIS CLI is secondary/internal infrastructure for development, lab, and a
future live install session AFTER bootable media exists. It is NOT the
primary operator journey.

Internal infrastructure order (fail closed — do not skip gates when mutating):

  1. vyntrio-verify-artifact <manifest.json>
     Read-only artifact integrity. Supports media/build verification.
     Does NOT install anything. Does NOT create USB/ISO images.

  2. vyntrio-installer preflight --target-disk-id <opaque-id> [--envelope-root ...] [--release-manifest ...]
     Read-only target + media checks. Use before trusting a target for writes.
     Does NOT write anything.

  3a. SANDBOX (dev/lab only — NOT the product install path):
      vyntrio-installer install --target-disk-id <opaque-id> --force [--envelope-root ...]
      Copies six allowlisted payloads under distro/installer/target-sandbox/<disk-id>/ only.
      Does NOT mutate real block devices. Do NOT treat as appliance install.

  3b. PARTITION APPLY (live-session / lab infra — not USB delivery):
      vyntrio-installer apply --target-disk-id <opaque-id> --force [--envelope-root ...]
      Mounts one existing supported partition and copies the same six payloads.
      Does NOT partition, format, or claim full OS install.
      Intended for future live-boot wizard reuse; not the primary delivery step.
      Equivalent: install --force --apply-target (prefer apply for clarity).

  4. vyntrio-installer postflight --target-root <path-with-HANDOVER_RECORD.txt>
     Replays operator handover summary. Does NOT perform installation.

Infrastructure guidance:
  - Prefer USB-first delivery for end-user installs (Block 9; not yet shipped).
  - Use sandbox install only for development and lab staging.
  - Use apply only for guarded partition writes (live session or lab).
  - Use preflight before any write command.
  - All writes require explicit --force and --target-disk-id.
  - Disk IDs are opaque (disk-<hash>); never pass /dev/sdX to the CLI.

Honest limits (all stages):
  - Not a monolithic installer or full OS deployment
  - No USB creator, ISO emission, or bootable media build
  - No partition creation, formatting, or bootloader install
  - No service enablement or bootstrap (deferred)
  - Six allowlisted payload files only

See: docs/ops/installer-policy.md, docs/00_PROJECT.md
`)
}

// UsageText returns vyntrio-installer CLI usage help.
func UsageText() string {
	return strings.TrimSpace(`usage:
  vyntrio-installer workflow
  vyntrio-installer preflight --target-disk-id <opaque-id> [options]
  vyntrio-installer install --target-disk-id <opaque-id> --force [options]
  vyntrio-installer apply --target-disk-id <opaque-id> --force [options]
  vyntrio-installer postflight --target-root <record-dir>

Block 10 internal infrastructure — secondary to USB-first delivery.
Not a single monolithic install command. Not the primary product journey.

Primary product path (intended):
  USB creator → bootable install medium → boot → local browser dashboard →
  live-session install → first-boot bootstrap
  (see docs/00_PROJECT.md; bootable media / USB creator not yet shipped)

  workflow    print USB-first vs Block 10 boundaries (read-only)
  preflight   read-only target/media verification (no writes)
  install     sandbox payload copy — DEV/LAB ONLY (unless --apply-target)
  apply       existing-partition payload apply (live-session/lab infra)
  postflight  replay HANDOVER_RECORD.txt (no writes)

Internal / lab flows (not primary operator journey):
  verify-artifact → preflight → install (sandbox, dev/lab) → postflight
  verify-artifact → preflight → apply (partition) → postflight

Sandbox path (dev/lab only):
  vyntrio-installer preflight --target-disk-id <id> ...
  vyntrio-installer install --target-disk-id <id> --force ...

Partition apply path (live-session infra / lab):
  vyntrio-installer preflight --target-disk-id <id> ...
  vyntrio-installer apply --target-disk-id <id> --force ...

All write commands require --force and --target-disk-id.
install without --apply-target is sandbox-only and does NOT mutate block devices.
apply is NOT partition creation and is NOT a full OS install.
Neither command builds USB/ISO media or replaces the USB-first product path.

shared options:
  --target-disk-id <id>      required opaque disk ID (GET /api/v1/storage/disks)
  --state-dir <path>         discovery context (default: /var/lib/vyntrio)
  --envelope-root <path>     install-media envelope root
  --release-manifest <path>  release manifest for integrity verification
  --artifact-base-dir <path> base directory for release manifest artifacts

install options:
  --force                    required to authorize sandbox (dev/lab) writes
  --apply-target             partition apply (same as apply; prefer apply)
  --mount-root <path>        temporary mount parent for --apply-target
  --sandbox-root <path>      bounded write root (default: distro/installer/target-sandbox)
  --target-root <path>       optional subpath under sandbox (default: <sandbox>/<disk-id>)

apply options:
  --force                    required to authorize partition apply
  --mount-root <path>        temporary mount parent (default: /run/vyntrio-install/mnt)

preflight options:
  --min-size-bytes <n>       minimum target size (default: 8589934592)

postflight options:
  --target-root <path>       directory containing HANDOVER_RECORD.txt
                             (sandbox target tree or install-apply record directory)

Run 'vyntrio-installer workflow' for USB-first vs Block 10 infrastructure summary.`)
}

// VerifyArtifactUsageText returns vyntrio-verify-artifact CLI usage help.
func VerifyArtifactUsageText() string {
	return strings.TrimSpace(`usage:
  vyntrio-verify-artifact [--base-dir <dir>] <manifest.json>

Read-only release/install artifact integrity check (SHA-256, size, presence).
Supports the USB-first media/build pipeline and Block 10 internal tools.
Does NOT create USB/ISO images and is NOT the primary install journey.

Verifies local artifacts against a vyntrio-release-manifest-v1 file.
Checks manifest structure, file presence, byte size, and SHA-256 digests.

Does NOT write disks, mount targets, copy install payloads, or replace:
  USB creator | bootable media build | vyntrio-installer preflight | install | apply | postflight

Recommended before consuming release artifacts in media builds or lab installer runs.
Signature (Ed25519) verification is not implemented in this release.

Primary product path (intended):
  USB creator → bootable install medium → boot → local dashboard
  (see docs/00_PROJECT.md)

Lab / infrastructure next step after success:
  vyntrio-installer preflight --target-disk-id <opaque-id> --release-manifest <manifest.json>`)
}

// RequiredUsageMarkers are substrings that must appear in installer usage text.
func RequiredUsageMarkers() []string {
	return []string{
		"Not a single monolithic",
		"USB-first",
		"secondary to USB-first",
		"DEV/LAB ONLY",
		"--force",
		"--target-disk-id",
		"sandbox",
		"apply",
		"workflow",
		"does NOT mutate block devices",
		"NOT partition creation",
		"Not the primary product journey",
	}
}

// RequiredWorkflowMarkers are substrings that must appear in workflow policy text.
func RequiredWorkflowMarkers() []string {
	return []string{
		PolicyVersion,
		"PRIMARY product path",
		"USB-first",
		"secondary/internal infrastructure",
		"dev/lab only",
		"vyntrio-verify-artifact",
		"preflight",
		"install",
		"apply",
		"postflight",
		"Does NOT mutate real block devices",
		"Does NOT partition",
		"No USB creator",
	}
}
