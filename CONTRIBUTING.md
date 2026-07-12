# Contributing to Vyntrio OS

Thank you for contributing. This project is documentation-first and phase-gated. Read the docs before writing code.

## Before you start

1. Read `docs/00_PROJECT.md` and `docs/01_MASTERPLAN.md`
2. Read `docs/02_ARCHITECTURE.md` and `docs/04_TECH_STACK.md`
3. Check `docs/20_TASKS.md` for the current phase and allowed work
4. Read `docs/17_SECURITY.md` and `docs/18_TESTING.md`

## Workflow

1. Create a branch from `main`
2. Make focused changes aligned with the current phase in `docs/20_TASKS.md`
3. Update documentation when behavior or structure changes
4. Run local checks:

   ```bash
   make bootstrap
   make verify
   make test
   ```

5. Open a pull request with:
   - Summary of what changed and why
   - Links to relevant docs/tasks
   - Test plan (even if "docs-only")

## What we do not accept (yet)

- Unplanned features outside the active phase
- Demo or placeholder application code without an approved task
- Secrets, credentials, or `.env` files
- Large refactors without an ADR when architecture is affected

## Architecture decisions

Significant technical decisions require an ADR in `adr/`. Use `adr/0000-template.md` as the starting point.

## Code style

- Go: `gofmt`, `go vet`; follow standard Go conventions
- Frontend: follow conventions defined in `docs/04_TECH_STACK.md` once finalized
- All text files: LF line endings (see `.editorconfig`)

## Questions

If scope is unclear, open a discussion or issue referencing `docs/20_TASKS.md` before implementing.
