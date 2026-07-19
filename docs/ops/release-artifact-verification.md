# Release artifact verification

Operator contract for `vyntrio-verify-artifact` (Block 9 / Slice 9.10).

## Purpose

Provide a trustworthy **local preflight** for release/install artifacts before any
install execution, disk write, or media creation exists. This tool answers:
“Do these files on disk match the published manifest?” — not “Is this release
authentic?” unless signature verification is implemented.

## Command

```bash
vyntrio-verify-artifact [--base-dir <dir>] <manifest.json>
```

- **`--base-dir`:** directory containing artifact files. Defaults to the manifest’s directory.
- **Exit 0:** manifest structure valid and all artifacts match size + SHA-256.
- **Exit 1:** malformed manifest or integrity failure (stderr lists `artifact`, `reason`, `detail`).
- **Exit 2:** usage error.

Does **not** require root. Does **not** mutate disks, partitions, or appliance state.

## Manifest format (`vyntrio-release-manifest-v1`)

JSON with strict parsing (unknown fields rejected).

| Field | Required | Notes |
|-------|----------|-------|
| `format_version` | yes | Must be `vyntrio-release-manifest-v1` |
| `created_at` | yes | RFC3339 / RFC3339Nano UTC |
| `release.version` | yes | Release version string |
| `release.channel` | no | `development`, `nightly`, `beta`, `stable`, `hotfix` |
| `artifacts[]` | yes | At least one entry |
| `signature` | no | Reserved for future Ed25519; **not verified in v1** |

### Artifact entry

| Field | Required | Notes |
|-------|----------|-------|
| `name` | yes | Stable identifier (unique in manifest) |
| `type` | yes | `binary`, `systemd_unit`, `config_template`, `static_file`, `archive`, `iso` |
| `relative_path` | yes | Relative to `--base-dir`; no `..` or absolute paths |
| `size_bytes` | yes | Exact byte length |
| `sha256` | yes | Lowercase hex SHA-256 |
| `use` | yes | `install_media`, `recovery_media`, `release_distribution`, `local_verification` |

Example: `distro/release/release-manifest.example.json`

## Integrity vs authenticity

| Outcome | Meaning |
|---------|---------|
| `integrity=ok` | All listed files present with matching size and SHA-256 |
| `integrity=failed` | Hard failure — do not use artifacts |
| `authenticity=not_signed` | No `signature` block; integrity only |
| `authenticity=unsupported` | `signature` present but Ed25519 verification not implemented |

**Checksum success does not prove publisher authenticity.** Stderr notes clarify this.

## Failure reasons

| Reason | Cause |
|--------|-------|
| `malformed_manifest` | JSON/schema violation |
| `unsupported_format_version` | Wrong `format_version` |
| `unknown_artifact_type` | Disallowed `type` |
| `unknown_artifact_use` | Disallowed `use` |
| `duplicate_artifact_name` | Repeated `name` |
| `invalid_relative_path` | Traversal or absolute path |
| `missing_file` | Artifact not found |
| `size_mismatch` | Byte length differs |
| `sha256_mismatch` | Digest differs |
| `invalid_sha256` | Manifest digest format invalid |

## Non-goals (this slice)

- ISO/USB image build or flash
- Installer execution (`vyntrio-installer`)
- Auto-download or artifact fetch
- Backup archive validation (`vyntrio-restore validate` — separate format)
- Claiming “installer complete” or installable appliance from verification alone

## References

- `internal/platform/releaseartifact/`
- `docs/ops/installer-policy.md` (staged pipeline context)
- `docs/19_RELEASE.md`
- `distro/install-media/manifest.yaml` — declarative install payload contract (YAML, separate)
