# Project-local grammar: appliance-console (provisional)

Compiled from Unraid-class WebGUI **information architecture and density** evidence
(R1: WuSiYu/unraid-custom-css screenshot.png — theme screenshot of Unraid WebGUI).
**Not** a copy of Unraid CSS, assets, branding, or proprietary chrome.

Nearest built-in fallback: `operations-console`  
Adapter: `product-ui`  
Confidence: **low–medium** (single visual reference + public Unraid docs IA)

## Twelve axes (evidence-backed)

1. **User job** — Scan live appliance state, find exceptions, act on storage/apps/VMs. (R1)
2. **Attention** — Multi-widget dashboard; first viewport shows host + services + storage together. (R1)
3. **IA / composition** — Top primary sections: Dashboard · Main/Storage · Shares · Users · Settings · Plugins · Docker · VMs · Apps · Tools. Persistent footer status strip. (R1 + Unraid docs)
4. **Density** — Compact tables, thin utilization bars, card/widget chrome with settings affordance. (R1)
5. **Typography** — Neutral UI sans; mono for IDs/temps optional; small uppercase section labels. (R1)
6. **Color** — One brand accent; green=started/healthy, red=stopped, amber=in-progress — semantic only. (R1)
7. **Surface** — Tonal panels, hairlines, soft radius; avoid marketing gradients. (R1 + StyleSeed ops lock)
8. **Data role** — Live utilization bars, container/VM lists, disk tables — not decorative charts. (R1)
9. **Actions** — Section settings gears; Docker start/stop; catalog install as primary Apps action. (R1)
10. **States** — Started/stopped/in-progress explicit; Planned/Demo when runtime absent. (product honesty)
11. **Responsive** — Desktop-first dense console; collapse widgets to single column under ~900px.
12. **Tells / reject** — Require: multi-section nav, footer status, widget dashboard. Reject: Unraid logo/copy, tile-launcher marketing, fake “healthy array” without backend.

## Anti-patterns
- Shipping Unraid CSS (`example.css`) or plugin assets into Vyntrio
- Claiming live Docker/VM/array without APIs
- Equal-weight marketing feature cards instead of operational widgets
