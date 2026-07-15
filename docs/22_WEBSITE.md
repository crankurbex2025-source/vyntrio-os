# Vyntrio Website & Product Surface (Block 11)

## Purpose

Block 11 is a **parallel frontend track** to ongoing OS/installer work (Blocks 9–10).
It introduces a public product surface for `vyntrio.xyz` and structures the existing
appliance UI for long-term growth — without faking backend capabilities.

## Surfaces

| Surface | Routes | API access | Delivery |
|---------|--------|------------|----------|
| **Public** | `/`, `/download` | **None** — static, honest marketing | Embedded SPA (near-term); extractable for `vyntrio.xyz` (future) |
| **Auth** | `/login` | Login POST only | Embedded SPA on appliance origin |
| **Appliance** | `/app/*` | Session probe, overview, settings | Embedded SPA on appliance IP/hostname |

## Honesty rules

1. **Public pages** must not import `lib/api` or display live operational data.
2. **Appliance pages** show only data returned by real API endpoints (Block 8 contract).
3. No fake download binaries, signup flows, metrics, testimonials, or uptime claims.
4. Placeholder routes must state clearly that content is not yet available.

## Structural inspiration

- **Unraid (structure):** public site → download/docs entry → authenticated control surface.
- **Unraid-PW (visual direction):** premium dark appliance aesthetic for future dashboard work — reference only, not a clone.

## Parallel-track boundary

Block 11 must **not** modify installer, media, distro, runtime, backup, or restore logic
unless a documented routing/CSP requirement forces a minimal change.

The existing delivery model is unchanged:

```
frontend/npm run build → make ui-stage → go:embed → vyntrio-api
```

## Block 11 slice plan

| Slice | Status | Focus |
|-------|--------|-------|
| **11.1** | **Implemented** | Router foundation + public landing at `/` |
| 11.2 | Planned | Shared design tokens + UI primitives |
| 11.3 | Planned | Download & docs entry pages |
| 11.4 | Planned | Auth decoupling — session probe only on `/app/*` |
| 11.5 | Planned | Appliance app shell scaffold |
| 11.6 | Planned | Overview visual refresh |
| 11.7 | Planned | i18n scaffold |
| 11.8 | Planned | Mobile-first + PWA stub |
| 11.9 | Planned | Route-based code splitting |
| 11.10 | Planned | Motion scaffold (GSAP-ready) |
| 11.11 | Planned | `vyntrio.xyz` deploy contract |

## Slice 11.1 behavior

- `/` — static landing page (hero, value proposition, CTAs to `/download` and `/login`)
- `/download` — honest placeholder (no binaries linked)
- `/login` — auth entry; successful login navigates to `/app` with CSRF handoff
- `/app` — existing appliance flow (session probe, overview, settings)
- Unknown paths redirect to `/`

## Future deployment note

On a dedicated marketing domain, `/` serves the public surface. On an appliance
hostname, a future slice may redirect `/` → `/login` or `/app` — not in 11.1.
