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

## Motion (Block 11R.5+)

- GSAP + `@gsap/react` + ScrollTrigger on `/design-preview/*` only
- `PreviewPageMotion` — `default` variant for download/docs; `landing` variant for richer hero, pillar stagger, showcase split
- Header gains `vyn-public-header-scrolled` after ~140px scroll (preview surfaces)
- **Lenis deferred** — native scroll keeps anchor links and reduced-motion behavior predictable
- `prefers-reduced-motion` mandatory — no GSAP setup when enabled
- Visual references: restrained grain/depth from [animated-electrician](https://github.com/Ismail-Khan-Dev/animated-electrician); appliance density from [Unraid-PWA](https://github.com/laurensguijt/Unraid-PWA); homelab theme discipline from [theme.park](https://github.com/themepark-dev/theme.park)
- Responsive: fluid `clamp()` gutters/spacing, `auto-fit`/`minmax()` grids, container queries on preview sections (`public-responsive.css`); layout reference patterns from [innovate-tech-landing](https://github.com/CodeWithKarol/innovate-tech-landing) and GitHub.com layout kit (bento/grid rhythm, not visual clone)
- Performance: preview routes and GSAP motion load via `React.lazy` + `Suspense`; Vite `manualChunks` isolates `preview-gsap` from the shipped `/` entry path
- No fake live metric animation, no looping decoration

## Block 11R slice map

| Slice | Status | Focus |
|-------|--------|-------|
| **11R.1** | **Implemented** | Tokens v2, i18n scaffold, `/design-preview/landing` |
| **11R.2** | **Implemented** | Reusable public components in `surfaces/public/components/` |
| **11R.3** | **Implemented** | Preview landing v2 section architecture on `/design-preview/landing` |
| **11R.4** | **Implemented** | Public component system for download/docs preview surfaces |
| **11R.5** | **Implemented** | Restrained GSAP motion layer on `/design-preview/*` |
| 11R.6 | Planned | Appliance token convergence |
| 11R.7 | Planned | PWA manifest |
| 11R.8 | Planned | Cutover `/` → v2 |
| 11R.9 | Planned | `vyntrio.xyz` deploy contract |

## Preview artifacts

- Routed: `/design-preview/landing` (`LandingPreviewV2.tsx`)
- Routed: `/design-preview/download` (`DownloadPreviewV2.tsx`)
- Routed: `/design-preview/docs` (`DocsPreviewV2.tsx`)
- Shared shell config: `surfaces/public/preview/previewShellConfig.ts`
- Motion scope: `surfaces/public/preview/motion/PreviewPageMotion.tsx`
- Static moodboard: `frontend/design-preview/direction-board.html` (not embedded)

## Coherence rule

Public site, login, and `/app` must converge on `vyntrio.tokens.css` by slice 11R.6. Until then, legacy `--dashboard-*` aliases remain for appliance stability.
