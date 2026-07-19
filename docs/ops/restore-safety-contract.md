# Restore Safety Contract

- **Status:** Restore CLI **implemented** (`vyntrio-restore`) — see section 12
- **Block:** 7
- **Slice:** 7.11 (documentation)
- **Authority:** This document elaborates fail-closed restore **requirements** for a
  future implementation. **`docs/ADR/0005-appliance-runtime-operations.md`**
  (sections D, F, G, H) remains the authoritative architecture source. Where this
  contract adds operational detail, it must not contradict the ADR.

## 1. Explicit status

Restore CLI **`vyntrio-restore`** is implemented for offline root-only restore from
completed local backup artifacts under `/var/lib/vyntrio/backups/`.

- **Command:** `vyntrio-restore validate <artifact-basename>` — manifest-first preflight only
- **Dry-run:** `vyntrio-restore <artifact-basename> --dry-run` — preflight + scope summary, no mutation
- **Destructive:** `vyntrio-restore <artifact-basename> --force` — requires explicit `--force`

**Scope (v1):** SQLite state files and `config/config.toml` from `vyntrio-backup-v1`
archives only. Does **not** replace binaries, systemd units, TLS, or install media.

**Not implemented:** remote restore, API/UI restore,
offline service worker, or full disaster-recovery runbook automation.

**Post-restore rollback (Slice 7.13):** when placement succeeds but service
startup or local health/readiness fails, `vyntrio-restore` attempts a bounded
rollback from the preservation copy, restarts the service, and reprobes. This is
**not** guaranteed after external interruption (power loss, `kill -9`).

Backup (`vyntrio-backup`, Slice 7.9) remains the sole backup writer.

## 2. Access model and non-goals

### 2.1 Permitted model

- **Root-only** operator with effective UID 0
- **Manually invoked**, **offline-only** local procedure on the appliance host
- Fixed, allowlisted filesystem targets only (see section 7)

### 2.2 Explicit exclusions

Restore must never be implemented as:

- a web or API endpoint
- a UI feature
- a background job, timer, or scheduled action
- automatic or unattended recovery
- remote, cloud, or network-target restore
- live in-place rollback while the service is running

Restore does **not** replace application binaries, systemd units, Nginx, TLS,
Certbot material, DNS, firewall rules, host users/groups, or non-Vyntrio packages.

### 2.3 Non-goals (out of scope for Block 7 restore v1)

- Encryption, signing, or key management
- Remote restore or upload/download surfaces
- Backup timers, retention, or pruning policy
- UI/API backup or restore controls
- Automatic disaster recovery
- Storage/NAS snapshot restore (`docs/05_STORAGE.md` roadmap — separate domain)
- Boot media, installer integration, or full-host recovery (as **restore
  implementation** — see ADR-0006 for what recovery **media** is, separately)
- Restore preview or delegated non-root restore authority

**Cross-reference (Block 9):** `docs/ADR/0006-appliance-install-recovery-media.md`
defines recovery boot media as an optional **offline delivery environment** that
may host future restore tooling. It does **not** implement restore, alter this
contract, or permit API/UI restore. Block 7 restore rules govern all live-state
mutation when restore is implemented.

## 3. Trust boundaries

| Boundary | Contract |
|----------|----------|
| **Root operator** | Inside the operational trust boundary. Malicious or mistaken root action is not mitigated by v1 restore design. |
| **Backup destination** | Fixed root-controlled directory (`/var/lib/vyntrio/backups/`, mode `0700`). The `vyntrio` service account must not write completed artifacts. |
| **Selected artifact** | Operator-selected **completed** local artifact only. Filename pattern alone is insufficient trust. |
| **Archive contents** | Allowlisted members only after manifest-first preflight (section 5). |
| **Live state / config** | Writable targets: `/var/lib/vyntrio/` (SQLite state) and `/etc/vyntrio/config.toml` when config restore is in scope. |
| **Service lifecycle** | Quiescent window: proven **inactive** after stop before mutation. Post-restore: proven **active** plus local loopback readiness. |

**Integrity note:** v1 per-member SHA-256 establishes **internal consistency**
(manifest agrees with tar member bytes). It does **not** establish cryptographic
authenticity, provenance, or protection against a malicious root-crafted archive.

## 4. Artifact selection

Before any validation that could lead to mutation:

1. Artifact must reside under the fixed, root-controlled local backup destination
   (default: `/var/lib/vyntrio/backups/`).
2. Reject temporary, incomplete, or in-progress names (including `.tar.tmp` and
   other non-final publication prefixes per ADR-0005 §G.4).
3. Accept only completed artifacts matching the approved publication naming
   convention: `vyntrio-backup-v1_<UTC-timestamp>.tar`.
4. **Never trust filename alone.** Selection gates are necessary but not
   sufficient; manifest-first preflight (section 5) is mandatory.

Arbitrary operator-supplied paths, relative paths, and path traversal in artifact
selection are rejected.

## 5. Manifest-first preflight (before service stop or live mutation)

A future restore implementation must perform **complete artifact preflight** while
**active Vyntrio state and configuration remain untouched**. If any step fails,
abort with **zero live-state mutation**.

### 5.1 Required preflight sequence

1. Open the selected completed artifact read-only.
2. Enumerate **every** tar archive member.
3. Decode the embedded `manifest.json` from the archive (manifest-first — do not
   accept an externally supplied manifest as the authority).
4. Verify manifest schema and content:
   - supported `format_version` (`vyntrio-backup-v1` only in v1)
   - well-formed creation timestamp and member entries (name, size, SHA-256)
   - manifest tar member present exactly once
5. Validate tar member safety for **each** entry:
   - reject absolute paths and `..` traversal components
   - reject empty, malformed, duplicate, unexpected, or disallowed member names
   - reject symlink, hard-link, device, FIFO, socket, and other non-regular-file
     entries
   - enforce the exact approved allowlist for the recognized format
6. Verify per-member SHA-256 digest consistency (tar bytes vs manifest).
7. Verify manifest-embedded tar `manifest.json` bytes match the decoded manifest
   (no split-brain manifest).
8. Apply compatibility gate (section 6) — still before service stop.

**Critical rule:** checksum match alone does **not** authorize extraction of an
unsafe path or disallowed member. Path and member validation remain mandatory.

### 5.2 Required and optional members (`vyntrio-backup-v1`)

| Archive member | Required | Host target (future) |
|----------------|----------|----------------------|
| `manifest.json` | Yes | (validation only) |
| `state/vyntrio.db` | Yes | `/var/lib/vyntrio/vyntrio.db` |
| `state/vyntrio.db-journal` | No (if present in archive) | `/var/lib/vyntrio/vyntrio.db-journal` |
| `state/vyntrio.db-wal` | No (if present in archive) | `/var/lib/vyntrio/vyntrio.db-wal` |
| `state/vyntrio.db-shm` | No (if present in archive) | `/var/lib/vyntrio/vyntrio.db-shm` |
| `config/config.toml` | Policy-controlled (see section 7) | `/etc/vyntrio/config.toml` |

No generic extraction to operator-provided paths. Mapping is fixed and normative.

### 5.3 Relationship to backup validation primitives

The backup package exposes tar/member/digest validators (for example
`ValidateArchive`, `ValidateArchiveMemberName`). These establish strong
**library-level** checks but are **not** a restore workflow:

- backup-oriented APIs may require a caller-supplied manifest;
- restore requires manifest-first self-validation from the embedded archive;
- validators do not enforce live-state non-mutation or orchestrate lifecycle steps.

## 6. Compatibility gate (before mutation)

After manifest-first preflight succeeds, and **before** `systemctl stop` or any
live-state change, a future implementation must fail closed on:

- unsupported or newer successor format identifiers
- schema **downgrade** (unsupported unless explicitly implemented later)
- incompatible API/binary metadata vs installed `vyntrio-api` (when metadata present)
- incompatible embedded migration revision vs installed binary (when metadata present)
- cross-major incompatibility (abort before overwriting live state)

### 6.1 Explicit deferrals (future implementation decisions)

The following require binding rules in a **future implementation slice** — this
contract records the requirement to decide, not the decision itself:

- **Exact comparison rules** for API version, commit, and migration revision
  matching (equality, minimum version, tolerated skew)
- **Dirty migration policy** — backup manifests may record `migration_dirty`;
  ADR-0005 §H.5 does not yet mandate reject vs warn-and-proceed

Until decided, restore implementation must not silently assume permissive compatibility.

## 7. Ordered restore lifecycle

After successful preflight and compatibility gate (**no live mutation yet**):

| Step | Action | Live state touched |
|------|--------|-------------------|
| 1 | Manifest-first preflight + compatibility gate | **No** |
| 2 | `systemctl stop vyntrio-api.service` | No (service only) |
| 3 | Prove service **inactive** (or `failed`); refuse if unproven | No |
| 4 | Create root-only **preservation copy** of current `/var/lib/vyntrio/` state and, when config restore is in scope, current `/etc/vyntrio/config.toml` | No (copy only) |
| 5 | **Staged write** validated members to fixed targets only | Staging area first — see deferrals |
| 6 | **Replacement** of live files from staged content | **Yes** |
| 7 | Enforce ownership/mode (§ADR-0005 §D, §H.2.6) | **Yes** |
| 8 | `systemctl start vyntrio-api.service`; prove **active** | No |
| 9 | Local loopback `GET /healthz` (liveness) | No |
| 10 | Local loopback `GET /readyz` (readiness) — final app-level gate | No |
| 11 | Emit safe operator-visible result (section 10) | No |

Public HTTPS must not substitute for local `/healthz` or `/readyz` in restore tooling.

### 7.1 State and configuration rules

- State SQLite files and config are **separate concerns** with separate manifest
  members and host targets.
- **No generic tar extraction** to relative or operator-chosen paths.
- **Config restore is policy-controlled** — preservation and replacement of
  `/etc/vyntrio/config.toml` occur only when a future implementation explicitly
  includes config in the restore scope (ADR-0005 §H.2.4–5).

### 7.2 Ownership and mode (required after replacement)

| Target | Owner | Mode |
|--------|-------|------|
| `/var/lib/vyntrio/` (directory) | `vyntrio:vyntrio` | `0750` (or stricter) |
| Database and sidecar files | `vyntrio:vyntrio` | `0640` (or stricter) |
| `/etc/vyntrio/config.toml` | `root:vyntrio` | `0640` |

Failure to apply ownership/mode is a restore failure, not success.

### 7.3 Preservation-copy failure

If step 4 (preservation copy) fails:

- **Do not replace live state.**
- Do not proceed to staged write or replacement.
- Attempt safe **service recovery** (restart `vyntrio-api.service` with unchanged
  live files) and prove inactive→active transition where possible.
- Emit an operator-visible **failure** state (section 10) without sensitive content.
- Leave the operator with an stopped or recovered service and **unmodified** live
  DB/config — treat as an incident if service state is ambiguous.

## 8. Explicit deferrals (implementation design)

The following are **mandatory future decisions** before restore code merges. This
contract defers them intentionally; silent defaults are forbidden:

| Topic | Why deferred |
|-------|--------------|
| **Staged layout** | Temp directory structure and write order before replacement |
| **Cross-root atomicity** | Failure semantics when `/var/lib/vyntrio` succeeds but `/etc/vyntrio` fails (or vice versa) |
| **Stale SQLite sidecar cleanup** | Whether host sidecars not present in the archive are deleted, retained, or rejected |
| **Rollback execution model** | Automatic vs operator-mediated restore from preservation copy |
| **Preserve-root location** | `/var/lib/vyntrio/backups/` vs another fixed root-only preserve root |
| **Exact compatibility matching rules** | See section 6.1 |
| **Dirty migration policy** | See section 6.1 |

## 9. Failure handling and rollback

- **No success** may be reported unless local `/readyz` succeeds within the bounded
  post-start probe policy (ADR-0005 §F, §G.0).
- On failure after placement when service startup or local health/readiness fails,
  `vyntrio-restore` **attempts rollback** from the preservation copy, repairs
  ownership, restarts the service, and reprobes. CLI outcomes:
  - `restore succeeded` — placement + service + probes passed
  - `restore failed: rollback=succeeded preserve=<basename> category=...` — live
    state restored from preservation; preservation copy retained for diagnosis
  - `restore failed: rollback=failed preserve=<basename> category=...` — manual
    recovery required; preservation copy retained
- On failure during placement or ownership (before post-restore verification),
  rollback is attempted without the extended restart/reprobe orchestration.
- **Interruption rollback is not guaranteed** (power loss, `kill -9`, manual deletion
  mid-restore). Failed restore is an incident; inspect preservation copies manually.

### 9.1 Post-restore operator expectations

- Session cookies will not match restored session rows; users re-authenticate.
- Readiness success is not proof of every historical workflow or browser session.
- Restored audit tables reflect restored history; optional operator audit entries
  require a future schema slice.

## 10. Safe reporting

Operator-visible restore output (stdout, stderr, logs intended for operators) must
**never** include:

- raw configuration contents
- database rows or file payloads
- archive or manifest body content
- credentials, passwords, tokens, cookies, or session identifiers
- CSRF material or other secret-bearing runtime values

Permitted output examples (classification only):

- start / completion / failure status
- failure category (preflight, compatibility, preservation, placement, ownership,
  service, health, readiness, rollback)
- artifact **basename** or identifier classification (not path contents)
- member **count** or scope summary (not member bytes)
- non-sensitive exit status

Digest values in operator logs are discouraged for restore unless a future slice
explicitly aligns with backup observability policy; they must never substitute for
safe failure classification.

## 11. Mandatory test categories (before restore code)

No restore implementation may merge until test plans exist (and later pass) for:

1. **Manifest-first validation** — self-contained preflight from embedded manifest
2. **Invalid-artifact no-mutation** — live DB/config unchanged on any preflight failure
3. **Service stop / inactive proof** — refuse when inactive cannot be proven
4. **Preservation-copy failure** — no replacement; safe service recovery documented
5. **Staged-write failure** — no corrupt partial live state
6. **Cross-root partial failure** — explicit behavior across state and config roots
7. **Sidecar policy** — archive with/without sidecars vs host stale sidecars
8. **Ownership/mode enforcement** — post-replacement permissions
9. **Compatibility rejection** — unsupported format, downgrade, incompatible metadata
10. **Rollback behavior** — per future execution model
11. **Service recovery** — restart paths without false success
12. **Health/readiness failure** — no success when `/readyz` fails
13. **Safe output** — no leakage under failure and retry paths

Backup unit tests (`internal/platform/backup/*_test.go`) provide reusable patterns
but do **not** satisfy these restore categories.

## 12. References

- `docs/ADR/0005-appliance-runtime-operations.md` — sections D, F, G, H (authoritative)
- `distro/systemd/README.md` — deployment and backup command
- `docs/17_SECURITY.md` — security boundary
- `docs/02_ARCHITECTURE.md` — Block 7 status
- `internal/platform/backup/` — backup implementation and validation primitives
- `internal/platform/restore/` — restore CLI implementation (`vyntrio-restore`)
- `cmd/restore/` — restore command entrypoint
