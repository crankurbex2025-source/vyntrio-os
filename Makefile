.PHONY: help bootstrap verify fmt lint test test-backend test-frontend docs-check clean

ROOT := $(shell pwd)
BACKEND_DIR := $(ROOT)/backend
FRONTEND_DIR := $(ROOT)/frontend

help:
	@echo "Vyntrio OS — development commands"
	@echo ""
	@echo "  make bootstrap     Install/check toolchain dependencies"
	@echo "  make verify        Run foundation checks (no app build yet)"
	@echo "  make fmt           Format Go and frontend sources (when present)"
	@echo "  make lint          Run linters (when configured)"
	@echo "  make test          Run all tests"
	@echo "  make test-backend  Run Go tests"
	@echo "  make test-frontend Run frontend tests"
	@echo "  make docs-check    Validate docs structure"
	@echo "  make clean         Remove build artifacts"

bootstrap:
	@echo "Checking toolchain..."
	@command -v go >/dev/null 2>&1 || (echo "ERROR: Go is not installed. See docs/04_TECH_STACK.md"; exit 1)
	@command -v node >/dev/null 2>&1 || (echo "ERROR: Node.js is not installed. See docs/04_TECH_STACK.md"; exit 1)
	@command -v npm >/dev/null 2>&1 || (echo "ERROR: npm is not installed."; exit 1)
	@if [ -f "$(BACKEND_DIR)/go.mod" ]; then \
		cd "$(BACKEND_DIR)" && go mod download; \
	fi
	@if [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		cd "$(FRONTEND_DIR)" && npm ci --ignore-scripts 2>/dev/null || npm install --ignore-scripts; \
	fi
	@echo "Bootstrap complete."

verify: docs-check
	@echo "Verifying monorepo layout..."
	@test -d docs || (echo "Missing docs/"; exit 1)
	@test -d backend || (echo "Missing backend/"; exit 1)
	@test -d frontend || (echo "Missing frontend/"; exit 1)
	@test -f backend/go.mod || (echo "Missing backend/go.mod"; exit 1)
	@test -f frontend/package.json || (echo "Missing frontend/package.json"; exit 1)
	@echo "Foundation layout OK."

fmt:
	@if [ -d "$(BACKEND_DIR)" ]; then \
		cd "$(BACKEND_DIR)" && gofmt -w . 2>/dev/null || true; \
	fi
	@if [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		cd "$(FRONTEND_DIR)" && npm run fmt --if-present; \
	fi

lint:
	@if [ -f "$(BACKEND_DIR)/go.mod" ]; then \
		cd "$(BACKEND_DIR)" && go vet ./... 2>/dev/null || true; \
	fi
	@if [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		cd "$(FRONTEND_DIR)" && npm run lint --if-present; \
	fi

test: test-backend test-frontend

test-backend:
	@if [ -f "$(BACKEND_DIR)/go.mod" ]; then \
		cd "$(BACKEND_DIR)" && go test ./...; \
	else \
		echo "No backend/go.mod — skipping Go tests"; \
	fi

test-frontend:
	@if [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		cd "$(FRONTEND_DIR)" && npm test --if-present; \
	else \
		echo "No frontend/package.json — skipping frontend tests"; \
	fi

docs-check:
	@./scripts/docs-check.sh

clean:
	@rm -rf backend/bin backend/dist backend/coverage.out backend/coverage.html
	@rm -rf frontend/dist frontend/build frontend/.next frontend/coverage
	@echo "Clean complete."
