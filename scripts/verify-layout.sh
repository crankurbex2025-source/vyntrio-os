#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

REQUIRED=(
  "go.mod"
  "Makefile"
  "docs/README.md"
  "docs/00_PROJECT.md"
  "docs/02_ARCHITECTURE.md"
  "docs/04_TECH_STACK.md"
  "docs/ADR/0001-use-clean-architecture.md"
  "cmd/api/main.go"
  "frontend/package.json"
  "scripts/bootstrap.sh"
)

DIRS=(
  "cmd"
  "internal"
  "frontend"
  "packages"
  "tests"
  "docs"
  "docs/ADR"
  "scripts"
  ".github/workflows"
)

missing=0
for f in "${REQUIRED[@]}"; do
  if [[ ! -f "$f" ]]; then
    echo "MISSING FILE: $f"
    missing=$((missing + 1))
  fi
done

for d in "${DIRS[@]}"; do
  if [[ ! -d "$d" ]]; then
    echo "MISSING DIR: $d"
    missing=$((missing + 1))
  fi
done

if [[ "$missing" -gt 0 ]]; then
  echo "verify-layout failed: $missing issue(s)"
  exit 1
fi

echo "verify-layout passed"
