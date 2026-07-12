#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

REQUIRED_DOCS=(
  "docs/README.md"
  "docs/00_PROJECT.md"
  "docs/01_MASTERPLAN.md"
  "docs/02_ARCHITECTURE.md"
  "docs/03_ROADMAP.md"
  "docs/04_TECH_STACK.md"
  "docs/17_SECURITY.md"
  "docs/18_TESTING.md"
  "docs/20_TASKS.md"
  "docs/21_CURSOR_REPO_SETUP.md"
  "docs/AUDIT_FOUNDATION.md"
)

missing=0
for doc in "${REQUIRED_DOCS[@]}"; do
  if [[ ! -f "$doc" ]]; then
    echo "MISSING: $doc"
    missing=$((missing + 1))
  fi
done

if [[ "$missing" -gt 0 ]]; then
  echo "docs-check failed: $missing required document(s) missing"
  exit 1
fi

echo "docs-check passed: all required documents present"
