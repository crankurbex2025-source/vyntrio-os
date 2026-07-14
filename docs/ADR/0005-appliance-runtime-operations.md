# ADR-0005: Appliance runtime configuration and operations model

- **Status:** Accepted
- **Date:** 2026-07-13
- **Deciders:** Vyntrio OS engineering (Architecture Block 7; Slice 7.1 acceptance)

## Context

Block 6 is complete: one Go binary embeds the production frontend
(`go:embed`), serves UI and `/api/v1` same-origin, and enforces session,
memory-only CSRF, owner-only settings mutation, API isolation, CSP, and cache
policy. The next concern is **appliance operation**, not UI functionality.

Current runtime truth (verified against `cmd/api/main.go`,
`internal/platform/config`, `internal/infrastructure/persistence/sqlite`):

- **Implemented (Slice 7.2):** the API server loads host-admin runtime
  configuration from **`/etc/vyntrio/config.toml`** (or `--config <absolute-path>`
  for development/tests). TOML schema: `bind_address`, `listen_port`,
  `state_dir`, `log_level`, `cookie_secure` — all required, no implicit
  defaults; strictness via explicit required-key/exact-key-count/type validation
  and parser duplicate-key rejection. Legacy environment variables
  (`VYNTRIO_API_HOST`, `VYNTRIO_API_PORT`, `VYNTRIO_DATA_DIR`, `VYNTRIO_LOG_LEVEL`,
  `VYNTRIO_ENV`, `VYNTRIO_COOKIE_SECURE`) no longer affect the API server.
- **Startup-time path validation (implemented):** config-controlled traversal
  blocked; `state_dir` must equal exactly `/var/lib/vyntrio` in production;
  symlink `state_dir` rejected; pre-existing symlinks for `vyntrio.db` and known
  SQLite sidecar names (`vyntrio.db-journal`, `vyntrio.db-wal`, `vyntrio.db-shm`)
  rejected before SQLite open; fixed DB name and server-built DSN only.
- **Not guaranteed (deferred):** race-free confinement against a concurrent local
  actor who can mutate the state directory after validation; `os.Root`-bound
  SQLite I/O; complete local-filesystem attacker resistance. `os.OpenRoot` is a
  startup accessibility probe only and does not constrain pathname-based SQLite
  opens through `modernc.org/sqlite`.
- **Operational assumption:** `/var/lib/vyntrio` is administered by the trusted
  host operator and is not concurrently writable by an untrusted local actor.
  **Slice 7.3** operationalizes this via static `vyntrio` identity, restrictive
  directory permissions, and `StateDirectory=vyntrio`; it does not make the
  current pathname SQLite open race-free.
- SQLite (pure-Go `modernc.org/sqlite` v1.46.1, CGO-free) at
  `<state_dir>/vyntrio.db`; default **DELETE** journal mode (rollback journal
  sidecar `vyntrio.db-journal` when active); WAL/SHM are not enabled now but
  are checked at startup if present; migrations embedded at startup; state
  directory created `0o750`; `foreign_keys` enforced; pool limited to one
  connection.
- Startup is fail-closed: invalid configuration (including path-boundary
  violations), database open/migrate/ping failure, settings load/verification
  failure, credential-service init failure, or embedded-UI init failure aborts
  the process with a clear error.
- `/healthz` is process liveness; `/readyz` is process plus database readiness.
- Shutdown is graceful: SIGINT/SIGTERM → `http.Server.Shutdown` with a
  configured timeout → database close.
- The binary requires no source checkout, `frontend/dist`, `node_modules`,
  Vite, dev proxy, or static-files directory at runtime.

**Implemented (Slice 7.9):** root-only local backup command `vyntrio-backup`
(`cmd/backup`). **Not implemented:** restore CLI, backup timer, retention,
encryption, or remote upload. **Slice 7.8 (accepted contract):** binding
backup/restore safety architecture documented in sections **G** and **H** below.
**Slice 7.3 (implemented):** systemd unit,
static `vyntrio` service identity declaration, tmpfiles layout for
`/etc/vyntrio`, and conservative unit sandboxing. Config-file ownership is
documented for administrators; runtime enforcement of config ownership modes
beyond startup readability checks remains packaging/operator responsibility.
Later implementation slices must follow the contract below without expanding
scope beyond what is explicitly approved.

## Decision

### A. Appliance model

- Vyntrio v1 is a **single-node Linux appliance service**.
- The production deliverable is **one Go binary with the frontend embedded**.
  The runtime must never require a source checkout, `frontend/dist`,
  `node_modules`, Vite, a dev proxy, or a static-files directory.
- **systemd is the first supported service-manager target** for Block 7.
- Deferred: Docker/OCI images, ISO/installer packaging, Kubernetes,
  clustering, high availability, public-cloud deployment, and multi-node
  operation.

### B. Runtime configuration boundary (host admin)

- There is exactly **one host-admin-controlled runtime configuration source**:
  a configuration file at the canonical location **`/etc/vyntrio/config.toml`**.
- **Format: TOML.** Rationale: comment support and human editability for
  operators, deterministic parsing without YAML's implicit-typing ambiguity,
  flat-table fit for a small schema, and mature Go libraries. Consequence: the
  future runtime-config slice adds one vetted TOML parser dependency; **implemented
  in Slice 7.2** (`github.com/pelletier/go-toml/v2`).
- Minimal v1 schema (architectural level only):
  - bind address (default loopback);
  - listen port;
  - persistent state directory (SQLite location derives from it);
  - log level (`debug`/`info`/`warn`/`error` — supported by the current slog
    architecture);
  - production cookie-security mode (the existing `Secure`-cookie switch;
    already required by the current cookie policy for non-TLS LAN operation).
- Explicitly **not** part of the v1 schema: TLS certificates, proxy trust,
  CORS, external services, analytics, session or CSRF secrets, database
  credentials, backup destinations, cloud storage, feature flags.
- Fail-closed parsing contract for the future loader: unknown keys, duplicate
  keys, invalid types, invalid values, unsafe combinations, and prohibited
  paths (see D) **must fail startup**; the service must not run on a partially
  understood configuration.
- Defaults must be secure, deterministic, and documented. **No default may be
  silently inferred from the current working directory**; the existing
  `./data` fallback is a development convenience that the runtime-config slice
  must eliminate for production operation.
- Runtime configuration changes require a **controlled service restart**.
  v1 has no SIGHUP reload, no live reconfiguration, and no web-based runtime
  configuration editing.

### C. Product-settings boundary (application state)

- Database-backed product settings — in particular the instance display name —
  remain **application state** in SQLite, managed exclusively through the
  existing server-side authentication, owner permission, CSRF validation,
  transaction/audit behavior, and API contracts (ADR-0004, `docs/09_API.md`).
- Runtime configuration is **never** exposed through the UI, the API, or the
  settings mutation endpoint, and must not grow into a second generic
  settings system. The two stores answer different questions: *how the host
  runs the service* (root-administered file) versus *what the product is
  configured to be* (owner-administered database rows).

### D. Filesystem layout, ownership and path boundaries

Intended v1 layout:

| Path | Purpose | Service access |
|------|---------|----------------|
| `/usr/bin/vyntrio-api` (or `/usr/lib/vyntrio/`) | installed immutable binary | read/execute only |
| `/etc/vyntrio/` | root-administered configuration | read-only |
| `/var/lib/vyntrio/` | persistent state incl. SQLite | read/write (only writable state) |
| `/var/lib/vyntrio/backups/` | root-only local backup destination (future CLI) | **no service-account write** (future slice) |
| `/run/vyntrio/` | ephemeral runtime files | read/write, only if a future slice needs it |

- The database file and SQLite sidecar files must remain within the dedicated
  state directory. Current runtime uses **DELETE** journal mode; WAL/SHM are
  possible future sidecars if journal mode changes.
- **Implemented startup checks:** config-controlled traversal blocked; exact
  `state_dir` root; symlink `state_dir` rejected; pre-existing symlinks for the
  main DB and known sidecar names rejected before SQLite open.
- **Not implemented:** race-free persistence I/O confinement; `os.Root`-bound
  SQLite open; rejection of post-validation filesystem mutation by a local actor.
- Prohibited by configuration contract: relative database paths, runtime paths
  derived from the current working directory, arbitrary host paths, path
  traversal, and writing adjacent to the binary.
- Intended ownership/mode expectations (**Slice 7.3 artifacts + operator
  install**): state directory owned by the service account with restrictive
  modes (directory `0750`, database files `0640` or stricter via `UMask=0027`);
  configuration owned by root, group-readable by `vyntrio`, never writable by it
  (`/etc/vyntrio` `0750`, `config.toml` `0640` recommended).

### E. systemd service identity and sandboxing

- **Static dedicated service account `vyntrio`** (non-login, no shell,
  dedicated group) instead of `DynamicUser=yes`. Rationale: persistent SQLite
  ownership, offline restore, backup handling, and incident recovery require
  **stable file ownership** across service lifecycles; DynamicUser's mapped
  UIDs complicate all four. The service **must not run as root**.
- **Implemented (Slice 7.3):** `distro/systemd/vyntrio-api.service` with
  `User=`/`Group=vyntrio`, restrictive `UMask=0027`, `StateDirectory=vyntrio`
  (`StateDirectoryMode=0750`), `Restart=on-failure`, `NoNewPrivileges=yes`,
  empty `CapabilityBoundingSet=`, `ProtectSystem=strict`, `ProtectHome=yes`,
  `PrivateTmp=yes`, `PrivateDevices=yes`, and
  `RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX`. Account provisioning via
  `distro/systemd/vyntrio.sysusers`; `/etc/vyntrio` layout via
  `distro/systemd/vyntrio.tmpfiles.conf`. Install steps:
  `distro/systemd/README.md`.
- `ProtectSystem=strict` is paired with the explicit approved writable state
  location (`StateDirectory=vyntrio` → `/var/lib/vyntrio`).
- **Not guaranteed:** race-free SQLite I/O against a concurrent local writer in
  the state directory; pathname opens remain outside descriptor-bound confinement.
- **Deferred:** aggressive `SystemCallFilter=`/seccomp policy, until a
  distribution-tested hardening slice proves compatibility with SQLite,
  networking, DNS, and any future TLS/proxy requirements.

#### E.1 Overview host metrics (Slice 8.3, implemented)

- The API collects read-only host metrics **in-process** for
  `GET /api/v1/overview` only: CPU logical core count and 1-minute load from
  bounded `/proc/loadavg` parsing; memory totals from bounded `/proc/meminfo`
  (`MemTotal`, `MemAvailable`); filesystem capacity via `statfs` **only** on the
  startup-validated `state_dir` (`/var/lib/vyntrio` in production).
- Expose only stable IDs (`filesystems[].id = "state"`), never mount paths,
  devices, UUIDs, backup directories, or mount enumeration.
- Per-metric collector failure sets that section to `unavailable` and omits
  numeric fields; overview remains HTTP **200**. No privileged helper, no sandbox
  relaxation, no `/proc/stat`, no network interface reads, no host inventory API.

### F. Startup, liveness, readiness and shutdown

- **Fail-closed startup (implemented):** invalid runtime configuration,
  inaccessible state path, startup-time path/symlink validation failure,
  unavailable/unopenable database, settings load/verification failure,
  credential-service (hasher) init failure, session-service init failure, or
  embedded-UI init failure prevents readiness and terminates startup with a
  clear error.
- **Deferred:** race-resistant persistence I/O confinement (custom VFS,
  descriptor-based open, or equivalent) for pathname-based SQLite access.
- Endpoint semantics are preserved: **`/healthz` is liveness**, **`/readyz` is
  service readiness** (process plus database). Later implementation must
  ensure readiness reports success only after required persistent state and
  server initialization are complete.
- Graceful shutdown contract (matches current implementation): stop accepting
  new work, close listeners, finish or safely interrupt in-flight work within
  the configured timeout, flush/close persistence according to the current
  SQLite policy, then exit.

### G. Backup contract (Slice 7.8 architecture; Slice 7.9 backup CLI)

**Status:** backup CLI **implemented** (`vyntrio-backup`). No restore command,
timer, API, UI, encryption, retention, or remote target exists.

#### G.0 Operator command (implemented)

- **Command:** `vyntrio-backup` (build output: `bin/vyntrio-backup`; production
  install convention: `/usr/bin/vyntrio-backup` when packaged like `vyntrio-api`).
- **Invocation:** run as **root** only; no arguments; fixed paths only.
- **Success output:** artifact basename and member count only — no config,
  database, digest, or credential content.
- **Verification:** local loopback `/healthz` and `/readyz` after restart with a
  finite retry/backoff policy (50 ms initial delay, 100 ms retry interval, 15 s
  deadline) to accommodate normal listener startup; never public HTTPS.
- **Manifest integrity:** per-member SHA-256 only; manifest itself is not
  cryptographically authenticated in v1.

#### G.1 Authority and access model

- Backups are a **local root-operator CLI feature** — not a web/API/UI feature,
  not a `vyntrio` service-account action on completed artifacts, not an
  automated background job, and not a cloud or network upload.
- Root/administrator access is already inside the operational trust boundary
  for appliance administration. This contract does **not** add protection
  against a malicious or concurrent local writer in the state directory, nor
  race-free path containment beyond what startup validation already documents.

#### G.2 Backup scope (included / excluded)

**Included in a backup set:**

| Member | Path / scope | Notes |
|--------|----------------|-------|
| SQLite main database | `<state_dir>/vyntrio.db` | Required |
| SQLite sidecars (if present) | `vyntrio.db-journal`, `vyntrio.db-wal`, `vyntrio.db-shm` | Copy only files that exist at backup time |
| Runtime configuration (companion) | `/etc/vyntrio/config.toml` | Separate manifest member; see G.5 |

**Excluded (never implied by v1 backup):**

- Installed binary (`/usr/bin/vyntrio-api`), embedded frontend build output,
  systemd unit files, sysusers/tmpfiles artifacts, Nginx/Certbot/TLS material,
  Cloudflare/DNS/firewall state, OS packages, logs, `/tmp`, credentials stored
  outside the explicitly included config file, operator shell history, and
  third-party data.
- Optional future application-managed persistent files **do not exist today**;
  if added later, they require an explicit contract extension before inclusion.

`/etc/vyntrio/config.toml` must **not** be documented as always safe to copy
blindly: the v1 schema has no secrets, but future keys may hold sensitive
material. The future CLI must treat config as operator-protected data, apply
restrictive permissions to any copied config member, and never emit raw config
contents in logs or operator-visible success output.

#### G.3 SQLite consistency model (first implementation decision)

**Current persistence facts (implemented):**

- Database path: `<state_dir>/vyntrio.db` with `state_dir` fixed to
  `/var/lib/vyntrio` in production.
- Driver: `modernc.org/sqlite` v1.46.1; DSN is pathname-based; pool limited to
  one connection; only `foreign_keys` pragma is set — **journal mode is not
  configured explicitly** and therefore remains SQLite default **DELETE**
  (rollback journal sidecar `vyntrio.db-journal` when a transaction is active).
- WAL/SHM are not enabled now but are rejected as symlinks at startup if
  present; a future journal-mode change would make sidecar handling mandatory.
- Graceful shutdown is implemented: `SIGINT`/`SIGTERM` → `http.Server.Shutdown`
  with configured timeout → `store.Close()` (`cmd/api/main.go`). systemd uses
  `KillSignal=SIGTERM`. Shutdown drains in-flight HTTP work but is **not** a
  separate maintenance/quiesce API.

**Approaches considered (conceptual only — not implemented):**

| Approach | Role in v1 contract |
|----------|---------------------|
| 1. Stop `vyntrio-api.service`, then copy state files | **Approved first implementation** |
| 2. SQLite online backup API via application/driver | Deferred; requires explicit driver verification and new code |
| 3. External `sqlite3 .backup` / equivalent | Deferred; adds tooling dependency and operator foot-guns |
| 4. Raw filesystem copy while service is running | **Forbidden** as a documented backup mechanism |

**Approved first-implementation sequence (contract level):**

1. Verify `vyntrio-api.service` is running or intentionally stopped for
   maintenance; refuse backup if state is ambiguous without operator intent.
2. `systemctl stop vyntrio-api.service` and prove **inactive** (or equivalent
   verified stop). If stop fails, abort with a non-zero exit and **no**
   published artifact.
3. Copy `vyntrio.db` and every **existing** known sidecar file from
   `/var/lib/vyntrio/` into a root-only temporary staging area inside the
   approved backup destination (see G.6). Copying **only** `vyntrio.db` while
   the service is active is **not** a guaranteed-consistent backup because
   DELETE-journal and future WAL sidecars may be mid-write.
4. Optionally copy `/etc/vyntrio/config.toml` as a separate companion member
   per G.5.
5. Build manifest, compute SHA-256 digests, atomically publish the completed
   artifact (G.5).
6. `systemctl start vyntrio-api.service` — required after successful artifact
   publication **and** after any backup failure path that stopped the service
   (including copy/manifest/checksum failures before publication).
7. Prove restart success before reporting backup completion: service **active**,
   local loopback `http://127.0.0.1:<listen_port>/healthz` returns success, and
   local loopback `/readyz` returns success. **Implemented (Slice 7.9):**
   `vyntrio-backup` performs bounded local post-restart verification (50 ms
   initial delay, 100 ms retry interval, 15 s deadline). Verified in the approved
   production smoke test (Slice 7.9).
8. The backup command must **not** report success unless steps 6–7 succeed. If
   restart or either local probe fails, report backup **failure** clearly,
   preserve a completed artifact only according to the operator contract, and
   do **not** claim the appliance returned to service.
9. On copy/manifest/checksum failure **before** publication: remove temporary
   files, attempt service restart per steps 6–7, exit non-zero if restart or
   local probes fail, and do not leave a restore candidate behind.

**Explicit non-guarantees:** no crash-consistent snapshot of a live database;
no protection against a malicious local actor replacing files between stop and
copy; no symlink-traversal-resistant copy until a future implementation adds
one.

#### G.4 Artifact format, integrity, and configuration packaging

**Approved v1 local artifact model:**

- **Format identifier:** `vyntrio-backup-v1` (future CLI must reject unknown
  format versions fail-closed).
- **Publication pattern:** write under a root-only temporary name inside the
  backup destination, verify digests, then **atomic rename** to the final
  artifact name. Incomplete/temporary prefixes are **never** restore candidates.
- **Final name (contract):**
  `vyntrio-backup-v1_<UTC-timestamp>.tar` (exact archive layout is
  implementation detail; members below are normative).
- **Restrictive permissions:** completed artifact and manifest readable/writable
  only by root (mode `0600` or tighter directory `0700` policy — see G.6).
- **Manifest (required, JSON or equivalent):** artifact format version, creation
  timestamp (UTC), installed `vyntrio-api` version/commit metadata when safely
  determinable, database schema/migration revision metadata when safely
  determinable, per-member relative path, byte size, SHA-256 digest.
- **Checksum rule:** future restore tooling must verify manifest digests before
  touching production state. A per-member SHA-256 match does **not** make an
  unsafe archive path or disallowed member name safe; path validation remains
  mandatory (see H.2).
- **Config packaging decision (approved):** configuration is a **separate
  manifest member** (`config.toml`) inside the same backup artifact, not merged
  into the database file. Rationale: keeps SQLite state and host-runtime config
  distinguishable for partial recovery and makes future secret-bearing config
  keys explicit in operator handling. Operators may omit config capture via a
  future explicit CLI flag; default is include-with-warning.

**Out of scope for first implementation:** encryption, signing, deduplication,
compression, remote upload, retention pruning, and key management.

#### G.5 Observability (future CLI)

Safe operator-visible reporting only:

- start / completion / failure status;
- artifact identifier and destination **classification** (not contents);
- SHA-256 digests in logs **allowed** (not treated as secrets);
- **never** passwords, raw config values, cookies, session IDs, CSRF tokens,
  database rows, or credential material.

Structured audit events for backup are **deferred** unless a later slice extends
the audit schema without widening this slice.

#### G.6 Backup destination and access boundary

- **Approved destination root:** `/var/lib/vyntrio/backups/` (root-controlled).
  The `vyntrio` service account must **not** receive write access to completed
  backup artifacts or the destination directory unless a future slice explicitly
  revisits this boundary with justification. Backup creation runs as **root**.
- Destination and temp paths must be **absolute**, fixed or allowlisted, and
  validated; arbitrary operator-supplied paths are rejected.
- Symlink traversal resistance is **not** claimed until implemented.
- No HTTP/API/UI download or upload surface.

### H. Restore contract (Slice 7.8 — approved for future implementation)

**Status:** architecture contract only. No restore command exists today.

Detailed fail-closed restore **safety requirements** (preflight gates, lifecycle
ordering, deferrals, and mandatory pre-implementation test categories) are in
`docs/ops/restore-safety-contract.md`. That document elaborates section H; it does
not replace this ADR as the architecture authority.

#### H.1 Access model

- Restore is an **explicitly invoked offline root-operator procedure** — never
  an API endpoint, UI feature, background action, live in-place rollback, or
  automatic recovery.
- Restore does **not** replace application binaries, systemd units, Nginx, TLS,
  Certbot certificates, DNS, firewall rules, host users/groups, or non-Vyntrio
  packages.

#### H.2 Required operator sequence

1. Operator selects a **completed** backup artifact (reject temp/incomplete names).
2. Before touching active Vyntrio state or extracting the archive, a future
   implementation must validate the artifact fail-closed:
   - recognize only the approved backup artifact format and manifest;
   - enumerate every archive member;
   - reject absolute paths;
   - reject `..` traversal components;
   - reject empty, malformed, duplicate, unexpected, or disallowed member names;
   - reject symlink, hard-link, device, FIFO, socket, or other non-regular-file
     archive entries unless a later approved contract explicitly supports them;
   - allow only the exact approved member set for the recognized artifact format;
   - verify per-member SHA-256 digests from the manifest (checksum match alone
     does not authorize extraction of an unsafe or disallowed path).
   If any validation fails, abort without modifying active Vyntrio state.
3. `systemctl stop vyntrio-api.service` and prove **inactive**. **Refuse** restore
   if stop cannot be proven.
4. Create a **root-only pre-restore preservation copy** of current
   `/var/lib/vyntrio/` state (and current `/etc/vyntrio/config.toml` if config
   restore is requested) under a timestamped directory beneath
   `/var/lib/vyntrio/backups/` or another fixed root-only preserve root.
5. Extract and restore validated SQLite member files only into `/var/lib/vyntrio/`
   and validated config member only into `/etc/vyntrio/config.toml` when included.
6. Enforce ownership/mode: state directory and DB files `vyntrio:vyntrio` with
   directory `0750` and database `0640` (or stricter); config `root:vyntrio`
   `0640`.
7. `systemctl start vyntrio-api.service` and prove service **active**.
8. Prove application recovery via **local** `http://127.0.0.1:<listen_port>/healthz`
   and `/readyz` only (not public HTTPS in the restore tool). Readiness success
   is the final application-level validation; it is **not** proof of every
   historical workflow, session continuity, or transactional host recovery.
9. Emit operator-visible outcome summary without config contents, credentials,
   session data, or database rows.

#### H.3 Rollback boundary

- If manifest or archive-member validation, file placement, ownership repair,
  service start, or readiness proof fails, the operator procedure must document
  attempting rollback to the pre-restore preservation copy created in step 4.
- **Not guaranteed:** rollback after external interruption (power loss, kill -9,
  manual deletion mid-restore). Operators must treat a failed restore as an
  incident requiring manual inspection of preserve copies.

#### H.4 Post-restore semantics (operator expectations)

- **Session invalidation:** restored session rows will not match live browser
  cookies; users must re-authenticate.
- **Audit semantics:** restore is operator-driven; application audit tables reflect
  restored history — a future CLI may append an operator-visible audit entry only
  if a later slice extends schema support.
- **Readiness success** is the final application-level validation; it is **not**
  proof that every historical workflow or browser session survived.

#### H.5 Compatibility and migrations

- Unknown or newer `vyntrio-backup-v1` successor formats must **fail closed**.
- Before restore, compare backup metadata (when present) against installed binary
  and embedded migration revision. **Schema downgrade is unsupported** unless
  explicitly implemented later.
- Forward upgrade after restore may run embedded migrations on next start only
  where existing migration code proves upgrade paths; restore must not assume
  downgrade safety.
- Cross-major incompatibility must abort before overwriting live state.

#### H.6 Explicit restore exclusions

No live restore; no web/API/UI restore; no restore from arbitrary unsigned
archives outside the approved layout; no credential/key recovery beyond what
exists in restored files; no multi-node orchestration; no encryption, signing, or
key management in v1.

#### H.7 Deferred implementation topics (explicit)

The following remain **out of scope** until dedicated future slices:

- restore CLI implementation (see section H and
  `docs/ops/restore-safety-contract.md`);
- systemd backup timer or scheduling;
- retention/pruning;
- compression;
- encryption and key management;
- signatures;
- remote/cloud/network targets (S3, WebDAV, rclone, etc.);
- UI/API backup or restore;
- incremental backup;
- online SQLite backup API / custom VFS;
- filesystem snapshots;
- restore preview;
- role-based delegated backup authority;
- installer/package integration beyond documenting paths;
- monitoring/alerting integration;
- tested disaster-recovery runbook execution;
- symlink-race-resistant file operations;
- full-host or infrastructure configuration backup.

### I. Networking and transport boundary

- Block 7 introduces **no** public internet exposure, reverse proxying,
  trusted forwarded headers, TLS termination, ACME, CORS changes, firewall
  management, or domain routing for vyntrio.xyz.
- A future proxy/TLS slice must define its trust boundary **before** any
  forwarded-header handling or secure-cookie deployment-mode behavior is
  added. Current same-origin HTTP/API behavior is unchanged.

### J. Upgrade and rollback boundary

- Later packaging must keep the persistent state directory **independent of
  the immutable binary**: replacing the binary must never touch
  `/var/lib/vyntrio/`.
- Application/database compatibility checks and a backup-before-migration
  policy belong to future dedicated upgrade slices.
- **No promise** of binary rollback compatibility across all future schema
  versions; rollback across migrations is governed by the future upgrade
  slices together with the backup/restore contract (G, H).

## Consequences

### Positive

- Later runtime-config, systemd, backup, restore, and upgrade slices implement
  against one stable, reviewed contract instead of inventing behavior.
- Host administration (root, `/etc/vyntrio/`, restart-based) and product
  administration (owner, session/CSRF/audit, database) stay cleanly separated.
- Stable `vyntrio` file ownership keeps offline restore and incident recovery
  tractable.
- Fail-closed startup and path-boundary rules prevent silent misconfiguration.

### Negative / trade-offs

- A static service account requires packaging-time user creation (deferred
  work) that `DynamicUser` would avoid.
- Restart-only reconfiguration is less convenient than live reload but far
  simpler to reason about for a security-sensitive appliance.
- TOML adds one future parser dependency.

### Follow-up

- [x] Runtime-config slice: TOML loader for `/etc/vyntrio/config.toml`,
      fail-closed validation, removal of legacy environment-variable runtime
      inputs and the CWD-derived `./data` default.
- [x] systemd slice: unit file, service account provisioning, tmpfiles/state
      directory ownership, sandboxing per section E.
- [x] Slice 7.8 backup/restore safety architecture contract (sections G, H).
- [x] Backup CLI implementation per section G (`vyntrio-backup`).
- [ ] Restore CLI implementation per section H.
- [ ] Upgrade/migration-policy slice per section J.
- [ ] Distribution-tested seccomp/`SystemCallFilter` hardening slice.

## Alternatives considered

| Option | Why not chosen |
|--------|----------------|
| Keep environment variables as the only production config source | No comments, weak types, hard for operators to audit; env stays as dev convenience/explicit mapping only |
| YAML configuration | Implicit typing ambiguity and larger spec; unnecessary for a flat schema |
| JSON configuration | No comments; hostile to hand-editing by host admins |
| CLI flags as primary config | Unit files would embed configuration, mixing service identity with settings |
| `DynamicUser=yes` | Unstable UID mapping complicates persistent SQLite ownership, offline restore, backups, incident recovery |
| Web/API-triggered backup or restore | Expands the remote attack surface with the most destructive operations available |
| SIGHUP/live reload | Reconfiguration races against sessions/DB state; restart is deterministic and cheap on an appliance |

## References

- `docs/ADR/0001-use-clean-architecture.md`
- `docs/ADR/0003-sqlite-migrations.md`
- `docs/ADR/0004-identity-and-access.md`
- `docs/02_ARCHITECTURE.md`, `docs/17_SECURITY.md`, `docs/19_RELEASE.md`
- `docs/ops/restore-safety-contract.md` (Slice 7.11 restore safety elaboration)
- `internal/platform/config/config.go`, `cmd/api/main.go`,
  `internal/infrastructure/persistence/sqlite/sqlite.go`
