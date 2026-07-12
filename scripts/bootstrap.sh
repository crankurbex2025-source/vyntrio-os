#!/usr/bin/env bash
# Bootstrap local development environment for Vyntrio OS.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[bootstrap]${NC} $*"; }
warn()  { echo -e "${YELLOW}[bootstrap]${NC} $*"; }
fail()  { echo -e "${RED}[bootstrap]${NC} $*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "Missing required command: $1"
}

info "Checking toolchain..."

require_cmd git
require_cmd make

if command -v go >/dev/null 2>&1; then
  GO_VERSION="$(go version | awk '{print $3}')"
  info "Go: $GO_VERSION"
  go mod download 2>/dev/null || true
else
  warn "Go not installed — required for backend work (see docs/04_TECH_STACK.md)"
fi

if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then
  info "Node: $(node --version)"
  info "npm:  $(npm --version)"
  if [[ -f frontend/package.json ]]; then
    (cd frontend && npm install --ignore-scripts 2>/dev/null || true)
  fi
else
  warn "Node.js/npm not installed — required for frontend work (see docs/04_TECH_STACK.md)"
fi

if [[ ! -f .env ]] && [[ -f .env.example ]]; then
  warn "No .env file — copy .env.example to .env for local configuration"
fi

info "Running layout verification..."
./scripts/verify-layout.sh

info "Running docs check..."
./scripts/docs-check.sh

info "Bootstrap complete."
info "Next: read docs/README.md and docs/21_CURSOR_REPO_SETUP.md"
