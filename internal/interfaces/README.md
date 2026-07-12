# internal/interfaces

Delivery mechanisms: REST API, WebSocket streams, CLI wrappers.

- Translates HTTP/WS requests to application use cases.
- No business rules — validation and mapping only.
- Auth middleware lives here (Phase 2+).

See `docs/09_API.md`.
