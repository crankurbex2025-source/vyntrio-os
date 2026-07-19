# StyleSeed code gate — full appliance DEMO

Rule set: `operations-console × product-ui × infrastructure-ops × dashboard × none`

Scored surfaces: `ApplianceFullDemo.tsx`, `appliance-demo.css` (+ shared ops primitives).

## Design Score: 84 / 100

████████████████░░░░  B

| Category | Score | Notes |
|---|---|---|
| Color discipline | 14/16 | Tokenized accent/status; catalog cards stay neutral (−2 unused info token). |
| Hierarchy & typography | 13/16 | Ops headers + mono meta; catalog titles slightly even-weight (−3). |
| Layout & rhythm | 11/12 | Banner → topnav → widgets/tables → footer status strip. |
| Cards & elevation | 10/10 | Hairline panels; no marketing float shadows. |
| States & a11y | 14/18 | Demo banner; empty install pane; focus-visible via ops CSS. (−4 no live loading skeletons — mock only). |
| Motion & interaction | 6/6 | Instant section switches; Start/Stop + Install mutations. |
| Coherence | 10/12 | Same ops grammar as /app; demo banner is intentional warn soft. |
| Distinctiveness | 6/10 | Product IA + Docker install flow; still demo-widget familiar (−4). |

Gate floor ≥80: **PASS (84)**

Note: This is a **TEST DEMO** route (`/design-preview/appliance-full`). Production `/app` remains honest (no fake Docker runtime).
