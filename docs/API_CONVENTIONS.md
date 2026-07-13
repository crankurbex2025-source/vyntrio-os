# API Conventions

Conventions for the Vyntrio OS HTTP API. Applies from Slice 1 onward.

Related: `docs/09_API.md`, `docs/ADR/0002-use-chi-router.md`.

## Base URL and versioning

- REST JSON API under `/api/v1/`
- Operational probes at root: `/healthz`, `/readyz` (not versioned)
- OpenAPI spec: `api/openapi.yaml`

## Request ID

- Every response includes header `X-Request-ID`
- Clients may send `X-Request-ID`; server generates UUID v4 if absent
- Error bodies include the same `request_id`

## Logging

- Server uses Go stdlib **`log/slog`** (structured JSON in production, text in development)
- Log level from `VYNTRIO_LOG_LEVEL`
- See `docs/04_TECH_STACK.md`

## Success responses

JSON object bodies. Examples (Slice 1):

```json
{"status":"ok"}
```

```json
{"status":"ready","checks":{"process":"ok","database":"ok"}}
```

```json
{"status":"not_ready","checks":{"process":"ok","database":"error"}}
```

(Slice 2: `/readyz` returns **503** when `database` is `error`.)

```json
{"version":"0.2.0-dev","commit":"unknown"}
```

## Error responses

All non-2xx JSON errors use this shape:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Human-readable description",
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `error.code` | string | yes | Machine-readable code (UPPER_SNAKE_CASE) |
| `error.message` | string | yes | Safe client-facing message (no stack traces) |
| `error.request_id` | string | yes | Matches `X-Request-ID` response header |

### Standard codes (initial set)

| HTTP | code | When |
|------|------|------|
| 400 | `BAD_REQUEST` | Malformed request |
| 404 | `NOT_FOUND` | Unknown route or resource |
| 405 | `METHOD_NOT_ALLOWED` | Wrong HTTP method |
| 408 | `REQUEST_TIMEOUT` | Handler timeout exceeded |
| 500 | `INTERNAL_ERROR` | Unexpected server error |
| 503 | `SERVICE_UNAVAILABLE` | Process not ready (future: dependency down) |

## Runtime configuration (`cmd/api`, implementiert)

Canonical file: `/etc/vyntrio/config.toml`. Development/tests may pass
`--config <absolute-path>` only.

| Key | Type | Description |
|-----|------|-------------|
| `bind_address` | string | Literal IP only (`netip.ParseAddr`); no hostnames or wildcards |
| `listen_port` | integer | 1–65535 |
| `state_dir` | string | Must equal `/var/lib/vyntrio` exactly in production |
| `log_level` | string | `debug`, `info`, `warn`, or `error` (exact match) |
| `cookie_secure` | boolean | `false` allowed only with loopback `bind_address` |

All keys are required. Unknown keys, duplicate keys, wrong types, invalid
values, and path-boundary violations fail startup. Legacy `VYNTRIO_*` runtime
environment variables no longer affect the API server. See
`docs/ADR/0005-appliance-runtime-operations.md`.
