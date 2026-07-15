SHELL := /bin/bash
ROOT := $(shell pwd)
GO ?= go
NPM ?= npm

# Generated embed input for the Go binary. Staged from frontend/dist by
# ui-stage; never committed. Go compilation fails if this input is absent.
UI_STAGE_DIR := internal/interfaces/http/ui/dist

.PHONY: help bootstrap verify fmt lint test test-go test-frontend ui-stage build install-media-stage test-install-media-stage install-media-envelope test-install-media-envelope run-api docs-check clean sqlc-generate generate

help:
	@echo "Vyntrio OS — development commands"
	@echo ""
	@echo "  make bootstrap     Check toolchain and prepare workspace"
	@echo "  make verify        Validate repo layout and required docs"
	@echo "  make fmt           Format Go sources (when present)"
	@echo "  make lint          Run available linters"
	@echo "  make test          Run all available tests"
	@echo "  make test-go       Run Go tests (stages embedded UI first)"
	@echo "  make test-frontend Run frontend tests"
	@echo "  make ui-stage      Build frontend and stage dist for go:embed"
	@echo "  make build         Build Go binaries (embeds staged frontend)"
	@echo "  make install-media-stage  Stage install-media payloads locally"
	@echo "  make test-install-media-stage  Verify install-media staging output"
	@echo "  make install-media-envelope  Assemble local install-media envelope"
	@echo "  make test-install-media-envelope  Verify install-media envelope assembly"
	@echo "  make run-api       Run API server (cmd/api)"
	@echo "  make docs-check    Validate documentation structure"
	@echo "  make sqlc-generate Regenerate sqlc query code"
	@echo "  make generate      Alias for sqlc-generate"
	@echo "  make clean         Remove build artifacts"

bootstrap:
	@./scripts/bootstrap.sh

verify: docs-check
	@./scripts/verify-layout.sh

fmt:
	@$(GO) fmt ./...

lint: ui-stage
	@$(GO) vet ./...
	@cd frontend && $(NPM) run lint --if-present

test: test-go test-frontend

test-go: ui-stage
	@$(GO) test ./...

test-frontend:
	@cd frontend && $(NPM) test --if-present

ui-stage:
	@cd frontend && $(NPM) run build
	@rm -rf "$(ROOT)/$(UI_STAGE_DIR)"
	@mkdir -p "$(ROOT)/$(UI_STAGE_DIR)"
	@cp -R "$(ROOT)/frontend/dist/." "$(ROOT)/$(UI_STAGE_DIR)/"
	@test -f "$(ROOT)/$(UI_STAGE_DIR)/index.html" || { echo "ui-stage failed: staged index.html missing" >&2; exit 1; }
	@ls "$(ROOT)/$(UI_STAGE_DIR)/assets"/* >/dev/null 2>&1 || { echo "ui-stage failed: staged assets missing" >&2; exit 1; }

build: ui-stage
	@mkdir -p bin
	@$(GO) build -o bin/vyntrio-api ./cmd/api
	@$(GO) build -o bin/vyntrio-worker ./cmd/worker
	@$(GO) build -o bin/vyntrio-installer ./cmd/installer
	@$(GO) build -o bin/vyntrio-update-agent ./cmd/update-agent
	@$(GO) build -o bin/vyntrio-backup ./cmd/backup

install-media-stage: build
	@./scripts/stage-install-media.sh

test-install-media-stage: install-media-stage
	@./tests/installmedia/stage_test.sh

install-media-envelope: install-media-stage
	@./scripts/assemble-install-media-envelope.sh

test-install-media-envelope: install-media-envelope
	@./tests/installmedia/envelope_test.sh

run-api:
	@$(GO) run ./cmd/api

docs-check:
	@./scripts/docs-check.sh

clean:
	@rm -rf bin dist coverage.out coverage.html
	@rm -rf frontend/dist frontend/build frontend/.next
	@rm -rf "$(ROOT)/$(UI_STAGE_DIR)"
	@rm -rf distro/install-media/staging distro/install-media/envelope
	@echo "Clean complete."

sqlc-generate:
	@command -v sqlc >/dev/null 2>&1 || { echo "sqlc not installed; install from https://docs.sqlc.dev/en/latest/overview/install.html" >&2; exit 1; }
	@sqlc generate

generate: sqlc-generate
