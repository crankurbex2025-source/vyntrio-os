# StyleSeed visual verify — appliance WebGUI

Rule set: `operations-console × product-ui × infrastructure-ops × dashboard × none`

Viewport: desktop 1440×900 @2x  
Harness: `/design-preview/appliance-ops`  
Artifacts: `.styleseed/verify/{dashboard,storage,settings}-1440.png`

## Pixel gate

| Check | Result |
|---|---|
| Squint / not marketing tile launcher | PASS — sidebar + metric strip + panels |
| Focal | PASS — accent-rail metric strip leads |
| Balance | PASS — no dead lower third; tables fill storage |
| Fonts | PASS — body sans + mono metrics |
| One accent | PASS — copper/orange only for action/active |
| Contrast | PASS — light text on dark panels |
| Rhythm | PASS — consistent hairline panels |
| Grammar fit (operations-console) | PASS — scan → act; Planned honesty |
| Nav active on preview sections | FIXED — `forceActiveId` for harness |

## Verdict: PASS

Visual gate completed with screenshots inspected.
