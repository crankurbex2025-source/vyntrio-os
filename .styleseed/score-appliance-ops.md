# StyleSeed code gate — appliance WebGUI

Rule set: `operations-console × product-ui × infrastructure-ops × dashboard × none`

Scored surfaces: `OverviewShell.tsx`, `StorageShell.tsx`, `SharesShell.tsx`,
`SettingsShell.tsx`, `UsersShell.tsx`, `ToolsShell.tsx`, `appliance-ops.css`,
`ApplianceShell.tsx`.

## Design Score: 86 / 100

████████████████░░░░  B

| Category | Score | Notes |
|---|---|---|
| Color discipline | 14/16 | Tokenized accent; status colors sparse. Minor: info accent token exists unused (−2). |
| Hierarchy & typography | 14/16 | Ops h1 ~22px; mono tabular metrics; dense labels OK. |
| Layout & rhythm | 11/12 | Focal strip → setup → grouped panels/tables. |
| Cards & elevation | 10/10 | Hairline panels; no marketing float shadows on ops. |
| States & a11y | 14/18 | Empty table copy; alerts; focus-visible. Loading shells exist in ApplianceApp (−4: no skeleton on preview). |
| Motion & interaction | 6/6 | Restrained; reduced-motion handled. |
| Coherence | 10/12 | Soft radius + one accent on ops. Auth/marketing shell still uses gradient language (out of lock scope). |
| Distinctiveness | 7/10 | Product-specific strip/tables. Metric cells still somewhat even-weight (−3). |

### Fix first (applied)
1. Runtime metric first in strip (operational focal) → hierarchy
2. Flat accent brand mark + focus-visible + tabular mono → coherence/a11y
3. Accent left rail on metric strip → focal tell

Gate floor ≥80: **PASS (86)**
