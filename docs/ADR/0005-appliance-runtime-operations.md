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

- Configuration comes from **environment variables** with documented defaults:
  `VYNTRIO_API_HOST` (`127.0.0.1`), `VYNTRIO_API_PORT` (`8080`, validated
  1–65535), read/write/idle/shutdown timeouts, `VYNTRIO_ENV` (`development`),
  `VYNTRIO_LOG_LEVEL` (`info`), optional `VYNTRIO_COOKIE_SECURE`, and
  `VYNTRIO_DATA_DIR` with the default `./data` — a **current-working-directory
  dependency** that is unacceptable for a supervised production service.
- SQLite (pure-Go `modernc.org/sqlite`, CGO-free) lives at
  `<DataDir>/vyntrio.db`; migrations are embedded and run at startup; the data
  directory is created with mode `0o750`; `foreign_keys` is enforced and the
  pool is limited to one connection.
- Startup is fail-closed today: invalid configuration, database open/migrate/
  ping failure, settings load/verification failure, credential-service init
  failure, or embedded-UI init failure aborts the process with a clear error.
- `/healthz` is process liveness; `/readyz` is process plus database readiness.
- Shutdown is graceful: SIGINT/SIGTERM → `http.Server.Shutdown` with a
  configured timeout → database close.
- The binary requires no source checkout, `frontend/dist`, `node_modules`,
  Vite, dev proxy, or static-files directory at runtime (ADR-0004,
  `docs/19_RELEASE.md`).

There is no service-manager integration, no host filesystem contract, no
backup/restore story, and no host-admin configuration file. Later slices must
implement these against a stable contract; this ADR is that contract.

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
  future runtime-config slice adds one vetted TOML parser dependency; **no
  parser is added in this slice**, and the current environment-variable loader
  remains the implemented mechanism until that slice replaces or explicitly
  maps it.
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

- The database file **and all SQLite sidecar files** (journal/WAL/SHM) must
  remain within the dedicated state directory.
- Prohibited: relative database paths, runtime paths derived from the current
  working directory, arbitrary host paths, path traversal, symlink escape out
  of the state directory, and writing adjacent to the binary.
- Intended ownership/mode expectations (to be implemented by later packaging
  slices, **not enforced by current code**): state directory owned by the
  service account with restrictive modes (directory `0750`, database files
  `0640` or stricter); configuration owned by root, readable by the service
  account, never writable by it.

### E. systemd service identity and sandboxing

- **Static dedicated service account `vyntrio`** (non-login, no shell,
  dedicated group) instead of `DynamicUser=yes`. Rationale: persistent SQLite
  ownership, offline restore, backup handling, and incident recovery require
  **stable file ownership** across service lifecycles; DynamicUser's mapped
  UIDs complicate all four. The service **must not run as root**.
- Intended unit-level controls (architectural contract; **no `.service` file
  exists yet**): `User=`/`Group=vyntrio`, restrictive `UMask=`,
  `StateDirectory=vyntrio` (and `RuntimeDirectory=` only if needed),
  `Restart=on-failure`, `NoNewPrivileges=yes`, full capability removal
  (`CapabilityBoundingSet=`), filesystem protections (`ProtectSystem=strict`,
  `ProtectHome=yes`), `PrivateTmp=yes`, device restrictions
  (`PrivateDevices=yes`), and restrictive address families
  (`RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX`).
- `ProtectSystem=strict` must be paired **only** with the explicit approved
  writable state/runtime locations from section D.
- **Deferred:** aggressive `SystemCallFilter=`/seccomp policy, until a
  distribution-tested hardening slice proves compatibility with SQLite,
  networking, DNS, and any future TLS/proxy requirements.

### F. Startup, liveness, readiness and shutdown

- **Fail-closed startup today** (implemented): invalid runtime configuration,
  inaccessible state path, unavailable/unopenable database, settings
  load/verification failure, credential-service (hasher) init failure, session-
  service init failure, or embedded-UI init failure prevents readiness and
  terminates startup with a clear error.
- **Fail-closed startup (future runtime-config slice):** the TOML loader and
  path validation must additionally fail closed for invalid state-path
  boundaries, prohibited paths, path traversal, and symlink escape attempts
  (see section D). These checks are not implemented in the current
  environment-variable loader.
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

- [ ] Runtime-config slice: TOML loader for `/etc/vyntrio/config.toml`,
      fail-closed validation, removal of the CWD-derived `./data` default for
      production operation.
- [ ] systemd slice: unit file, service account provisioning, tmpfiles/state
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
