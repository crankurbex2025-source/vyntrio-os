# systemd deployment artifacts (Block 7, Slice 7.3)

Production-oriented systemd integration for the Vyntrio OS API binary
(`vyntrio-api`). These files are **installation artifacts**; they do not
create the host account or start the service during repository tests.

## Layout

| File | Installed path (typical) | Purpose |
|------|--------------------------|---------|
| `vyntrio.sysusers` | `/usr/lib/sysusers.d/vyntrio.conf` | Static `vyntrio` system user/group |
| `vyntrio.tmpfiles.conf` | `/etc/tmpfiles.d/vyntrio.conf` | `/etc/vyntrio` ownership and mode |
| `vyntrio-api.service` | `/etc/systemd/system/vyntrio-api.service` | Supervised API process |

Production binary path: `/usr/bin/vyntrio-api` (see `docs/ADR/0005-appliance-runtime-operations.md`).

## Filesystem permissions model

| Path | Owner | Mode | Service access |
|------|-------|------|----------------|
| `/etc/vyntrio/` | `root:vyntrio` | `0750` | read/execute only |
| `/etc/vyntrio/config.toml` | `root:vyntrio` (administrator-created) | `0640` recommended | read only |
| `/var/lib/vyntrio/` | `vyntrio:vyntrio` | `0750` | read/write (via `StateDirectory`) |
| SQLite files under state dir | `vyntrio:vyntrio` | `0640` typical (`UMask=0027`) | read/write |

The `vyntrio` service account must **not** be able to modify runtime
configuration. Persistent state lives only under `/var/lib/vyntrio/`.

**Security boundary:** systemd ownership, `StateDirectory`, and sandboxing
operationalize the trusted-host-administrator assumption. They do **not**
make pathname-based `modernc.org/sqlite` opens race-free against a concurrent
local actor who can mutate `/var/lib/vyntrio` after startup validation.

## Manual installation (administrator)

**Interim developer/lab path only** — not the primary Vyntrio product install journey.
The intended operator path is bootable install USB/ISO (Block 9) → boot → local
dashboard → install from live session (Block 9 + Block 10). Use this procedure for
development, CI, and appliances already running Linux until bootable install media
ships. See `docs/00_PROJECT.md` and `docs/ADR/0006-appliance-install-recovery-media.md`.

Prerequisites: built `vyntrio-api` binary and a valid
`/etc/vyntrio/config.toml` (see `docs/API_CONVENTIONS.md`).

```bash
# 1. Install binary
install -D -m 0755 bin/vyntrio-api /usr/bin/vyntrio-api

# 2. Provision service account (requires root)
install -D -m 0644 distro/systemd/vyntrio.sysusers /usr/lib/sysusers.d/vyntrio.conf
systemd-sysusers /usr/lib/sysusers.d/vyntrio.conf

# 3. Configuration directory layout
install -D -m 0644 distro/systemd/vyntrio.tmpfiles.conf /etc/tmpfiles.d/vyntrio.conf
systemd-tmpfiles --create /etc/tmpfiles.d/vyntrio.conf

# 4. Administrator creates/edits config (example permissions)
install -D -m 0640 -o root -g vyntrio /path/to/config.toml /etc/vyntrio/config.toml

# 5. Install and enable unit
install -D -m 0644 distro/systemd/vyntrio-api.service /etc/systemd/system/vyntrio-api.service
systemctl daemon-reload
systemctl enable --now vyntrio-api.service
```

Verify: `systemctl status vyntrio-api.service`, `curl -sf http://127.0.0.1:<listen_port>/healthz`.

## Unit verification

**Production unit (unmodified artifact):**

```bash
systemd-analyze verify distro/systemd/vyntrio-api.service
```

This checks the checked-in `vyntrio-api.service` with the real production
`ExecStart=/usr/bin/vyntrio-api --config /etc/vyntrio/config.toml`. On a
developer or CI host where `/usr/bin/vyntrio-api` is not yet installed,
`systemd-analyze verify` exits non-zero with:

```text
vyntrio-api.service: Command /usr/bin/vyntrio-api is not executable: No such file or directory
```

That result is an **environment limitation** (missing installed binary), not a
unit syntax or sandboxing defect. After `install -D -m 0755 bin/vyntrio-api
/usr/bin/vyntrio-api`, the same command should exit 0.

**Repository tests:** `tests/systemd/artifacts_test.go` asserts the exact
production `ExecStart` string on the real artifact and runs a separate
`systemd-analyze verify` against a temporary copy that substitutes
`ExecStart=/bin/true` only to validate unit syntax and sandbox directives where
the production binary is absent.

## Sandboxing notes

- `ProtectSystem=strict` keeps `/usr` and most of `/etc` read-only; writable
  state is limited to `StateDirectory=vyntrio` (`/var/lib/vyntrio`).
- `MemoryDenyWriteExecute` is intentionally omitted: Go runtimes may require
  executable heap mappings.
- `SystemCallFilter`/seccomp hardening remains deferred per ADR-0005.

## Backup command (implemented)

Root operators run **`vyntrio-backup`** on the appliance host. The command:

- requires effective UID 0;
- stops `vyntrio-api.service`, copies approved SQLite state files and
  `/etc/vyntrio/config.toml`, publishes `vyntrio-backup-v1_<UTC>.tar` under
  `/var/lib/vyntrio/backups/` (root `0700`, artifact `0600`);
- restarts the service and proves loopback `/healthz` and `/readyz` with a
  finite local retry policy after restart (no public HTTPS check);

Restore is implemented via **`vyntrio-restore`**. See `docs/ADR/0005-appliance-runtime-operations.md`
section G and H and `docs/ops/restore-safety-contract.md`.

## Offline restore (`vyntrio-restore`)

Root operators restore from completed local backup artifacts with
**`vyntrio-restore`** (Block 7 Slice 7.12):

```bash
vyntrio-restore validate vyntrio-backup-v1_<UTC-timestamp>.tar
vyntrio-restore vyntrio-backup-v1_<UTC-timestamp>.tar --dry-run
vyntrio-restore vyntrio-backup-v1_<UTC-timestamp>.tar --force
```

Destructive restore requires `--force`. Scope is SQLite state + config only — not
binaries, systemd units, or install media. If post-restore startup or health checks
fail, the command attempts rollback from the preservation copy and reports whether
rollback succeeded (`rollback=succeeded`) or also failed (`rollback=failed`).

See `docs/ADR/0005-appliance-runtime-operations.md` sections G and H and
`docs/ops/restore-safety-contract.md`.
