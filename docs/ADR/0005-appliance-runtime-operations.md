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

There is no backup/restore story yet. **Slice 7.3 (implemented):** systemd unit,
static `vyntrio` service identity declaration, tmpfiles layout for
`/etc/vyntrio`, and conservative unit sandboxing. Config-file ownership is
documented for administrators; runtime enforcement of config ownership modes
beyond startup readability checks remains packaging/operator responsibility.
Later slices must implement backup/restore against the contract below.

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
| `/var/lib/vyntrio/backups/` | optional local backup destination | read/write (future backup slice) |
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

### G. Backup contract (future CLI slice)

- Backups are a **local administrator CLI feature** — not a web/API/UI
  feature, not an automated background job, not a cloud upload.
- A backup must be **SQLite-consistent**. Raw file copy of a live database is
  **never** a documented backup mechanism. The later implementation prefers
  the SQLite Backup API (e.g. `VACUUM INTO` or the online backup interface),
  chosen only after verifying compatibility with the repository's driver
  (`modernc.org/sqlite`) and its single-connection locking policy.
- Output constraints: explicit destination, destination validation, write to a
  temporary file with restrictive permissions, integrity verification, atomic
  final rename only after success, and **no session/CSRF/password/token
  material in logs or backup metadata**.
- v1 has no automatic retention, scheduler, encryption-key management, remote
  upload, or web-triggered backup.

### H. Restore contract (future offline operation)

- Restore is an **explicitly invoked offline/local administrator operation** —
  never an API endpoint, UI feature, background action, or automatic rollback.
- Intended high-level sequence:
  stop service → validate candidate backup → preserve/handle current state
  according to a future explicit policy → atomically replace state → enforce
  ownership/mode → start service → verify readiness.
- Restore design must account for **session invalidation** (restored session
  rows do not match live browser state), **audit-log semantics** (the restore
  itself must be operator-visible), **database-version compatibility**
  (migration state of the backup versus the installed binary), and
  operator-visible recovery evidence. Destructive overwrite details beyond
  this safety contract are not decided here.

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
- [ ] Backup CLI slice per section G; restore procedure per section H.
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
- `internal/platform/config/config.go`, `cmd/api/main.go`,
  `internal/infrastructure/persistence/sqlite/sqlite.go`
