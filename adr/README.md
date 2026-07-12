# Architecture Decision Records (ADR)

This directory records significant technical and process decisions for Vyntrio OS.

## Index

| ADR | Status | Title |
|-----|--------|-------|
| [0000](0000-template.md) | Template | ADR template |
| — | — | _No accepted ADRs yet_ |

## When to write an ADR

Create an ADR when a decision:

- affects system architecture or module boundaries
- selects a core technology (database, framework, auth model)
- changes security or deployment posture
- is costly to reverse

## Process

1. Copy `0000-template.md` to `NNNN-short-title.md`
2. Fill in context, decision, and consequences
3. Link from relevant docs (`02_ARCHITECTURE.md`, `04_TECH_STACK.md`)
4. Mark status: Proposed → Accepted → Deprecated/Superseded

See `docs/02_ARCHITECTURE.md` for architecture principles (when finalized).
