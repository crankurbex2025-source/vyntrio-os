# ADR-0002: Use chi for HTTP Routing

- **Status:** Accepted
- **Date:** 2026-07-12
- **Deciders:** Vyntrio OS engineering (Slice 1 approval)

## Context

Vyntrio OS needs an HTTP layer for the control plane API (`cmd/api`). `docs/04_TECH_STACK.md` lists **chi** or **gin** as options. Slice 1 requires middleware for request IDs, structured logging, recovery, and timeouts — without coupling domain logic to a web framework.

## Decision

Use **[go-chi/chi/v5](https://github.com/go-chi/chi)** as the HTTP router for `internal/interfaces/http/`.

## Rationale

| Criterion | chi |
|-----------|-----|
| Built on `net/http` | Yes — aligns with Clean Architecture boundaries |
| Middleware composability | Strong — request ID, logging, recovery, timeout |
| API surface | Small — less lock-in than full frameworks |
| OpenAPI / REST versioning | Compatible with `/api/v1` route groups |
| Ecosystem fit | Common for Go appliance/control-plane services |

**gin** was not chosen because it replaces `context.Context` with `gin.Context`, which spreads framework types into handlers and complicates testing with standard `net/http/httptest`.

## Consequences

### Positive

- Handlers use standard `http.HandlerFunc` / `http.ResponseWriter`
- Middleware stays in `internal/interfaces/http/middleware/`
- Easy httptest coverage without framework bootstrapping

### Negative / trade-offs

- New dependency (chi + indirect modules)
- WebSocket routing deferred to a later ADR/slice

### Follow-up

- [ ] Slice 2+: auth middleware on chi router group
- [ ] OpenAPI generation tooling evaluation

## References

- `docs/04_TECH_STACK.md`
- `docs/09_API.md`
- `docs/ADR/0001-use-clean-architecture.md`
