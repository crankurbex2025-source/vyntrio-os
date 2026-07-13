# ADR-0004: Local-first identity and access control

- **Status:** Accepted
- **Date:** 2026-07-12
- **Deciders:** Vyntrio OS engineering (Architecture Block 4; Slice 4.1 acceptance)

## Context

Vyntrio OS is a self-hosted appliance control plane. Remote admin access is a primary threat surface (`docs/17_SECURITY.md`). Platform Core has SQLite persistence, internal settings, and graceful API shutdown, but **no authentication or authorization**. Public settings and admin APIs must not ship without identity controls.

Constraints:

- **Local-first:** no external IdP, OAuth, OIDC, LDAP, or SAML in v1.
- Offline-capable appliance; SQLite remains the initial identity store.
- Browser dashboard (React) and REST API share the same origin in production.
- Clean Architecture: domain rules independent of HTTP/cookie details (`docs/ADR/0001-use-clean-architecture.md`).
- ADR-0003 defers public settings API until auth/RBAC exists.

`docs/04_TECH_STACK.md` mentions JWT/refresh tokens; for a single-node appliance admin UI, **server-side sessions** are the primary model. API tokens may be added later for automation.

## Decision

### 1. Authentication model (browser/API boundary)

**Primary:** Server-side sessions with an **opaque cryptographically random session ID** stored in SQLite.

- Browser receives `HttpOnly`, `Secure` (production), `SameSite=Strict` cookie: `vyntrio_session`.
- The cookie value is the raw session ID; **SQLite stores only a hash** of the session ID (see §4).
- Session record stores: user ID, created/expiry/last-seen, user agent hash, IP (optional), revocation flag.
- Session TTL: **24h idle**, **7d absolute** (constants in v1 implementation; env-configurable later).
- Logout = server-side session revoke + cookie clear.
- **No JWT for browser auth in v1.**

**Secondary (deferred):** Opaque or JWT **API tokens** for CLI/automation — separate slice after session auth is stable.

### 2. Password hashing

- **Argon2id** via `golang.org/x/crypto/argon2` (or maintained wrapper).
- Initial target parameters: memory 64 MB, iterations 3, parallelism 4, salt 16 bytes, key 32 bytes — tuned per hardware in implementation; not immutable in this ADR.
- Password policy (v1): minimum length 12; block empty passwords; no secrets in logs or API responses.
- Password hash never logged, never returned in API.

### 3. Bootstrap / initial Owner

- On first start when **zero active users** exist, appliance enters **bootstrap mode**.
- Bootstrap mode allows only:
  - existing probes: `/healthz`, `/readyz`, `/api/v1/version`
  - bootstrap endpoint(s): create first **Owner** account (exact route defined in implementation slice)
- Bootstrap completes atomically: create Owner user + mark bootstrap complete (settings key or dedicated state).
- **Exactly one Owner** required at bootstrap; additional Owners allowed later by an existing Owner only.

#### Bootstrap network boundary (mandatory)

Before the first Owner exists, bootstrap **must not be freely reachable on the LAN**.

- **Default implementation direction:** bind bootstrap to **loopback only** (`127.0.0.1`), **or** require an equivalent **one-time local console/bootstrap secret** before bootstrap endpoints accept requests.
- **No unrestricted LAN-accessible bootstrap endpoint is permitted.**
- Exact transport mechanics and installer integration remain **deferred** to Phase 0.2 / later slices; this ADR binds the security requirement only.

### 4. Secret persistence (mandatory)

- Session IDs and CSRF tokens must be **cryptographically random**.
- SQLite stores **only hashes** of session IDs and CSRF tokens (e.g. SHA-256 or HMAC with server pepper).
- **Raw session IDs and CSRF tokens are never:**
  - persisted in SQLite
  - logged
  - returned after initial issuance (session ID via `Set-Cookie` only; CSRF via response body once per rotation)
  - sent in URLs or query strings
  - included in audit metadata

### 5. CSRF strategy (cookie sessions)

Because the SPA uses cookie sessions:

- `SameSite=Strict` cookie reduces cross-site risk.
- Every **state-changing endpoint that relies on an authenticated cookie-backed session** must require a valid **`X-CSRF-Token`** header matching the **session-bound CSRF token**, enforced **after** session authentication.
- In **current v1**, the only such endpoint is **`POST /api/v1/identity/logout`**. **`POST /api/v1/identity/bootstrap`** and **`POST /api/v1/identity/login`** are explicitly **pre-session** and do **not** use CSRF validation. **`GET /api/v1/settings`** is read-only and does not use CSRF middleware.
- **Future** cookie-session-authenticated **`POST` / `PUT` / `PATCH` / `DELETE`** endpoints must attach the same CSRF middleware before activation; no session-authenticated write may ship without it.
- Safe methods (`GET`, `HEAD`, `OPTIONS`) are exempt from CSRF validation.

#### CSRF delivery (mandatory)

- On **successful login** (`POST /api/v1/identity/login`), the server returns **200 OK** with JSON body `{ "csrf_token": "<raw opaque token>" }` only. No CSRF cookie is set.
- The raw CSRF token appears **only** in that login response body (over HTTPS in production). The SPA keeps it **in memory only** and sends it in the **`X-CSRF-Token`** header for mutating requests.
- **Logout** (`POST /api/v1/identity/logout`) requires a valid session cookie **and** matching `X-CSRF-Token` after authentication.
- **Never** expose CSRF tokens via URL, query string, logs, response headers, error responses, or cookies.

### 6. User identity model (minimal lifecycle)

**User** (conceptual fields):

- `id` (UUID)
- `username` (unique, case-insensitive compare)
- `display_name` (optional)
- `password_hash` (Argon2id)
- `role` (single fixed enum — see §7)
- `status`: `active` | `disabled`
- `created_at`, `updated_at`, `last_login_at`
- `must_change_password` (bool, for future recovery flows)

**Lifecycle (v1):**

- Create (admin or bootstrap)
- Disable (admin; cannot disable last Owner)
- Password change (self or admin)
- Role change (Owner/Admin rules below)
- No self-registration, no public signup

**Deferred:** email verification, invite links, SCIM, federated identities.

### 7. RBAC model

Five fixed system roles (no custom roles in v1):

| Role | Intent |
|------|--------|
| **Owner** | Full control; can manage Owners; appliance-level decisions |
| **Administrator** | Full admin except Owner-only actions |
| **Operator** | Operate workloads and infrastructure; no identity admin |
| **User** | Standard user; limited write |
| **Read-only** | View-only |

#### V1 role storage (mandatory)

- **V1 uses exactly one role stored directly on the user record** (`users.role` enum column).
- **Do not create `roles` or `user_roles` tables in v1.**
- Fixed role enum: **Owner**, **Administrator**, **Operator**, **User**, **Read-only**.
- **Multi-role support is explicitly deferred** to a future migration and ADR amendment.

**Assignment rules:**

- Last active Owner cannot be removed or demoted.
- **Only Owner** may assign the Owner role.
- Administrator may assign Administrator, Operator, User, Read-only.

**Authorization shape:**

- Domain permission constants in `internal/domain/identity` (or `access`).
- Application layer enforces via `Authorizer` port before use cases.
- HTTP middleware extracts session → identity → role → passes to handlers.

### 8. Authorization boundaries (future modules)

Permissions are **resource + action** (`settings:read`, `storage:write`, etc.). Module gates:

| Module | Minimum role (typical) | Notes |
|--------|------------------------|-------|
| Settings | Operator read; Admin write | Public settings API requires `settings:read`/`settings:write` |
| Storage | Operator write | Destructive ops may require Admin |
| Containers / VMs | Operator write | User: limited to owned resources (future) |
| Network | Admin write | Operator read in v1 |
| Updates | Operator apply | Admin configure policies |
| Licensing | Admin read/write | Operator read |
| Audit / security logs | Admin read | Owner export policy deferred |
| Users / roles | Admin manage | Owner role assignment |

Normative first-release matrix below.

### 9. 2FA extension point (not implemented)

Reserve in model only:

- `users.mfa_enabled` (bool, default false) — column may be added in a future migration
- TOTP/recovery storage — **no migration until dedicated slice**
- Login flow may later include `mfa_required` step
- Do not implement TOTP/WebAuthn in Block 4 v1 slices

### 10. Recovery and break-glass (concept only)

Document policy; do not implement in v1:

- **Break-glass:** physical/console access or signed recovery artifact held offline by operator.
- **Owner lockout:** requires break-glass procedure in ops runbook (future `docs/17_SECURITY.md` section).
- No backdoor account, no hardcoded credentials.

### 11. Audit logging (identity/security events)

Append-only **security audit events** (separate from general application logs):

Required event types (v1 implementation namespace `identity.*`):

- `identity.bootstrap.succeeded`
- `identity.login.succeeded` / `identity.login.failure`
- `identity.logout.succeeded`
- (deferred) `identity.session.revoked`, `user.created`, `user.disabled`, `user.password.changed`, `role.changed`, `access.denied`

Each event: timestamp, actor user ID (nullable on failed login), subject user ID where applicable, action, result, metadata JSON (**no secrets** — no session IDs, CSRF tokens, passwords, or password material). Failed login events use empty metadata `{}`.

Retention: local SQLite; export/deletion policy deferred.

## Initial data model outline (no SQL)

**users** — identity, password hash, **single `role` enum column**, status, timestamps.

**sessions** — **session ID hash**, user ID, expiry, idle deadline, **CSRF token hash**, revoked_at, metadata (no raw secrets).

**security_audit_events** — append-only audit stream.

**bootstrap state** — prefer settings key `system.bootstrap_completed`; dedicated table only if richer state is required later.

Relationships:

- User 1—N Sessions
- User 1—N SecurityAuditEvents (as actor or subject)

**Not in v1:** `roles` table, `user_roles` table, `user_mfa_secrets` table.

## Role / permission matrix — first release

Legend: **R** read, **W** write/admin, **A** apply/action, **—** denied.

| Permission | Owner | Administrator | Operator | User | Read-only |
|------------|-------|---------------|----------|------|-----------|
| `system:health` (probes) | R | R | R | R | R |
| `auth:session` (self) | W | W | W | W | W |
| `users:read` | W | W | — | — | — |
| `users:write` | W | W | — | — | — |
| `roles:assign` | W | W* | — | — | — |
| `roles:assign_owner` | W | — | — | — | — |
| `settings:read` | W | W | R | R | R |
| `settings:admin:read` | W | — | — | — | — |
| `settings:write` | W | W | — | — | — |
| `storage:read` | W | W | R | R | R |
| `storage:write` | W | W | W | — | — |
| `containers:read` | W | W | R | R | R |
| `containers:write` | W | W | W | — | — |
| `vms:read` | W | W | R | R | R |
| `vms:write` | W | W | W | — | — |
| `network:read` | W | W | R | R | R |
| `network:write` | W | W | — | — | — |
| `updates:read` | W | W | R | R | R |
| `updates:apply` | W | W | A | — | — |
| `licensing:read` | W | W | R | R | R |
| `licensing:write` | W | W | — | — | — |
| `audit:read` | W | R | — | — | — |
| `audit:export` | W | — | — | — | — |

\* Administrator may assign Administrator, Operator, User, Read-only — not Owner.

## Security assumptions and threat boundaries

**In scope threats:**

- Unauthenticated LAN attacker accessing admin API/UI.
- Unrestricted bootstrap on LAN before Owner exists (**mitigated by §3 bootstrap network boundary**).
- Stolen session cookie (mitigate: HttpOnly, Secure, SameSite, short TTL, revoke, hashed storage).
- CSRF against authenticated browser session.
- XSS exfiltrating CSRF token from memory (mitigate: CSP, no CSRF in cookies/URLs).
- Credential stuffing (mitigate: rate limit, audit failures, no user enumeration in errors).
- Privilege escalation via role manipulation (mitigate: Owner-only rules, audit).

**Out of scope (v1):**

- Multi-tenant isolation.
- External IdP compromise.
- Worker-to-API mTLS (deferred).

**Assumptions:**

- Admin UI served same-origin over HTTPS in production.
- Appliance physical access is trusted for break-glass procedures only.
- SQLite file integrity relies on OS filesystem permissions (`VYNTRIO_DATA_DIR`).

## Alternatives considered

| Option | Why not chosen |
|--------|----------------|
| JWT access + refresh as primary browser auth | Poor revocation on single node; refresh token theft = long-lived compromise |
| OAuth/OIDC/LDAP/SAML | Violates local-first boundary |
| `roles` + `user_roles` tables in v1 | Unnecessary complexity; single role per user sufficient for appliance v1 |
| LAN-accessible bootstrap without local proof | Unauthenticated Owner creation on network = critical vulnerability |
| Store raw session/CSRF tokens in SQLite | DB leak would grant immediate session hijack |
| CSRF token in cookie or URL | Defeats HttpOnly model or exposes token to logs/referrers |
| Fine-grained custom roles | Five fixed roles enough for v1 |
| Bootstrap via installer only | Installer not implemented; API bootstrap unblocks Platform Core with loopback guard |

## Consequences

### Positive

- Unblocks public settings API and admin modules with consistent RBAC.
- Session revocation supports logout and compromise response.
- Hashed secret storage limits blast radius of DB disclosure.
- Bootstrap network boundary closes pre-Owner LAN takeover.
- Audit trail satisfies `docs/17_SECURITY.md` and architecture rules.

### Negative / trade-offs

- Server-side sessions require DB reads per request (acceptable on appliance).
- CSRF adds frontend/API contract complexity.
- Single-role column requires migration for multi-role later.
- Loopback-only bootstrap requires local access or SSH tunnel for first setup until installer integration.

### Follow-up

- [x] Accept ADR-0004 (Slice 4.1)
- [ ] Slice 4.2: domain types + permission constants + Authorizer port (no DB, no HTTP)
- [ ] Slice 4.3: migration — `users`, `sessions`, `security_audit_events` (no API)
- [ ] Slice 4.4: sqlc + repositories (internal tests only)
- [ ] Slice 4.5: Argon2id password service
- [ ] Slice 4.6: session + CSRF service (application layer)
- [ ] Slice 4.7: bootstrap endpoint (loopback guard) + audit
- [ ] Slice 4.8: login/logout + session cookie middleware
- [ ] Slice 4.9: RBAC middleware
- [ ] Slice 4.10: public settings API behind `settings:read/write`

**Deferred:** 2FA, break-glass implementation, API tokens, user management UI, frontend login, installer integration, external IdP, license enforcement.

## References

- `docs/02_ARCHITECTURE.md`
- `docs/04_TECH_STACK.md`
- `docs/09_API.md`
- `docs/17_SECURITY.md`
- `docs/ADR/0001-use-clean-architecture.md`
- `docs/ADR/0003-sqlite-migrations.md`
- `docs/API_CONVENTIONS.md`
