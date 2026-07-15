# Website design direction (Block 11R)

**Status:** Accepted for redesign correction track (Slice 11R.1 foundation).
**Shipped route `/` remains Slice 11.1** until Block 11R.8 cutover.
**Preview route:** `/design-preview/landing`

## Verdict on Slice 11.1

| Keep | Reject |
|------|--------|
| Router surfaces (`/`, `/login`, `/app`) | English-only copy |
| Appliance/public split | Purple SaaS template aesthetic |
| Honesty rules (no fake data) | Symmetric 3-card hero scaffold |
| Embedded delivery model | Decoupled public/dashboard token families (interim until 11R.6) |

## Product positioning

**German-first premium appliance OS** — Homelab/Server-Kontrolle. Not generic self-hosted SaaS.

## Visual language (v2)

| Token | Value | Role |
|-------|-------|------|
| `--vyn-bg` | `#06080d` | Graphite base |
| `--vyn-accent` | `#e85d2b` | Warm appliance accent (distinct from Unraid, same class) |
| `--vyn-accent-info` | `#3d8bfd` | Links, release strip |
| `--vyn-font-display/body` | IBM Plex Sans | Marketing + UI |
| `--vyn-font-mono` | IBM Plex Mono | Status labels, metrics |

**Rules:** No glassmorphism hero cards. No glow-orbs as logo. Tighter radii (`0.375–0.5rem`). One token file: `frontend/src/shared/theme/vyntrio.tokens.css`.

## Landing section architecture (preview v2 — Block 11R.3)

1. **Hero band** (`#product`, elevated) — lead headline, CTA stack with hints, companion control-surface frame
2. **Release strip** — honest status (no fake version/download)
3. **Capability pillars** (`#capabilities`, inset) — eyebrow, intro, tagged unequal grid (storage featured)
4. **Control-surface showcase** (`#control-surface`, elevated) — section intro + full-width static frame
5. **Product status** (`#product-status`, inset) — honesty boundary with bullet points
6. **Final CTA band** (elevated) — closing download/sign-in hierarchy
7. **Footer** — tagline, release/sign-in links, preview note, DE/EN switcher in header

Structural reference: [unraid.net](https://unraid.net). Dashboard density reference: [Unraid-PWA](https://github.com/laurensguijt/Unraid-PWA).

## i18n contract

| Locale | Role |
|--------|------|
| `de` | Default — authored first |
| `en` | First-class — separate copy in `locales/en/public.json` |

- Persistence: `localStorage` key `vyntrio.locale`
- Browser fallback: `navigator.languages`, then `de`
- Future: URL prefix `/de`, `/en` in slice 11R.3+
- **No machine translation pipeline**

## Motion (deferred to 11R.5)

- GSAP + `@gsap/react` + ScrollTrigger + Lenis
- `prefers-reduced-motion` mandatory
- No fake live metric animation

## Block 11R slice map

| Slice | Status | Focus |
|-------|--------|-------|
| **11R.1** | **Implemented** | Tokens v2, i18n scaffold, `/design-preview/landing` |
| **11R.2** | **Implemented** | Reusable public components in `surfaces/public/components/` |
| **11R.3** | **Implemented** | Preview landing v2 section architecture on `/design-preview/landing` |
| 11R.4 | Planned | Download/docs pages v2 |
| 11R.5 | Planned | GSAP motion layer |
| 11R.6 | Planned | Appliance token convergence |
| 11R.7 | Planned | PWA manifest |
| 11R.8 | Planned | Cutover `/` → v2 |
| 11R.9 | Planned | `vyntrio.xyz` deploy contract |

## Preview artifacts

- Routed: `/design-preview/landing` (`LandingPreviewV2.tsx`)
- Static moodboard: `frontend/design-preview/direction-board.html` (not embedded)

## Coherence rule

Public site, login, and `/app` must converge on `vyntrio.tokens.css` by slice 11R.6. Until then, legacy `--dashboard-*` aliases remain for appliance stability.
