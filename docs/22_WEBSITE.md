# Vyntrio Website & Product Surface (Block 11)

## Purpose

Block 11 is a **parallel frontend track** to ongoing OS/installer work (Blocks 9–10).
It delivers a public product surface for `vyntrio.xyz` and structures the appliance
UI for long-term growth — without faking backend capabilities.

## Surfaces

| Surface | Routes | API access | Delivery |
|---------|--------|------------|----------|
| **Public** | `/`, `/download`, `/docs` | **None** — static, honest marketing | Embedded SPA via `vyntrio-api` (live on `vyntrio.xyz`) |
| **Auth** | `/login` | Login POST only | Embedded SPA on appliance origin |
| **Appliance** | `/app/*` | Session probe, overview, settings | Embedded SPA on appliance IP/hostname |

Preview mirror routes (`/design-preview/*`) and legacy rollback paths
(`/design-preview/*-legacy`) remain for internal review; they are not the
production public contract.

## Honesty rules

1. **Public pages** must not import `lib/api` or display live operational data.
2. **Appliance pages** show only data returned by real API endpoints (Block 8 contract).
3. No fake download binaries, signup flows, metrics, testimonials, or uptime claims.
4. Release and docs copy must state clearly when install media or step-by-step guides are not yet available.

## Structural inspiration

- **Unraid (structure):** public site → download/docs entry → authenticated control surface.
- **Unraid-PWA (visual direction):** premium dark appliance aesthetic for future dashboard work — reference only, not a clone.

## Parallel-track boundary

Block 11 must **not** modify installer, media, distro, runtime, backup, or restore logic
unless a documented routing/CSP requirement forces a minimal change.

Delivery model (implemented):

```
frontend/npm run build → make ui-stage → go:embed → vyntrio-api
```

## Block 11 / 11R slice status

| Slice | Status | Focus |
|-------|--------|-------|
| **11.1** | **Superseded at `/`** | Original English-only landing (retained at `/design-preview/landing-legacy`) |
| **11R.1–11R.5** | **Implemented** | Tokens v2, public components, preview architecture, download/docs surfaces, GSAP motion |
| **11R.8–11R.9** | **Implemented** | Landing visual/media; product-story modules across routes |
| **11R.10** | **Implemented** | Root cutover `/` → v2 landing; preview mirror + legacy fallback |
| **11R.11** | **Implemented** | Shipped `/download` + `/docs`; production link cleanup |
| **11R.12** | **Implemented** | Public docs/release content depth; install readiness guidance |
| **11R.13** | **Implemented** | Landing **Use cases** (intended-users) section — DE/EN, anchor `#use-cases`; completes the landing foundation section set |
| **11R.14** | **Implemented** | Landing **First boot / local setup** section (`#first-boot-setup`) — mirrors the product-side local-first onboarding flow; DE/EN; directly tied to Slice 10.3 |
| **11R.15** | **Implemented** | Landing truth alignment for Stage 3 — storage pillar + first-boot copy updated for read-only dashboard storage inventory, LAN/TLS opt-in (`VYNTRIO_LAN_BIND_IP`); product-status boundary; DE/EN; deployed |
| **11R.16** | **Implemented** | Creator / media / self-hosting landing completion — `#creators` audience layer (`PublicUseCaseSection`, DE/EN) + `#local-today` honest local-capabilities outline (`PublicProcedureOutline`); original wording; no Remote Connect/licensing/install-media claims |
| 11R.6 | **Implemented** | Appliance/login token convergence on `vyntrio.tokens.css` |
| 11R.7 | **Implemented** | PWA manifest and web shortcut installability (no offline SW) |
| 11.4 | Partial | `/login` separate; session probe on `/app/*` only |
| 11.5 | Partial | Minimal appliance shell (overview + settings, not full dashboard) |
| 11.6 | Planned | Appliance overview visual refresh (parity with public v2) |
| 11.7 (Block 11) | Partial | Public DE/EN i18n shipped; appliance UI still English-only |
| 11.8 | Implemented | Responsive public surfaces + PWA manifest foundation (11R.7) |
| 11.9 | Implemented | Lazy-loaded public routes (`React.lazy` + `Suspense`) |
| 11.10 | Implemented | GSAP motion on shipped public surfaces |
| 11.11 | Implemented | `vyntrio.xyz` live deploy via embedded SPA |

Details: `docs/23_WEBSITE_DESIGN_DIRECTION.md`.

## Customer-facing website line (parallel to the product line)

The website track runs **in parallel** to the product line but stays gated to the
current product reality (see `docs/03_ROADMAP.md`). It mirrors an Unraid-style
information architecture at a high level only — no copied text, branding, layout
assets, or design details.

| # | Website stage | Status |
|---|---------------|--------|
| 1 | **Landing page foundation** — hero, value, pillars, use cases, creator/media audience, local-today honesty, first-boot local setup, product status, primary CTA, footer/nav | **Complete** (11R.16 added creator/media + local-today sections) |
| 2 | Feature marketing sections | Next |
| 3 | Pricing / licensing page | Gated (product licensing not started) |
| 4 | Remote/Connect marketing | Gated (Remote Connect not started) |

Honesty gate: the site must not imply Remote Connect, licensing/pricing, or full
storage/NAS mutation (pools, shares, format) are available. Product state today is
**Stage 3 (read-only storage inventory in dashboard)** on Stage 2 local dashboard
work; public install media is **not** available.

## Shipped public routes

- `/` — v2 landing (DE/EN, lazy-loaded, product story, honest release CTAs); foundation section set: hero, value/release, product pillars, **use cases** (`#use-cases`), **creators & media** (`#creators`), **local capabilities today** (`#local-today`), **first boot / local setup** (`#first-boot-setup`), control-surface preview, product-status boundary, primary + final CTA, footer/nav
- `/download` — v2 release surface (artifact status, readiness, install outline; no published file)
- `/docs` — v2 documentation surface (static topics, install/operate/recovery structure; no live search)
- `/login` — auth entry; successful login navigates to `/app` with CSRF handoff
- `/app` — appliance flow (session probe, read-only overview, read-only storage inventory, Owner settings)
- Unknown paths redirect to `/`

## Appliance UI scope (current)

The appliance surface is **not** a full Unraid/TrueNAS-class dashboard. It provides
read-only overview panels (Block 8), **read-only storage inventory (Stage 3 / 11.1)**,
Owner settings, and auth — without pools, shares, container, VM, or widget modules.

## Future deployment note

On a dedicated marketing domain, `/` serves the public surface (current: `vyntrio.xyz`).
On an appliance hostname, a future slice may redirect `/` → `/login` or `/app`.
