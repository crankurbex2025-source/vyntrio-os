SHELL := /bin/bash
ROOT := $(shell pwd)
GO ?= go
NPM ?= npm

.PHONY: help bootstrap verify fmt lint test test-go test-frontend build run-api docs-check clean sqlc-generate generate

help:
	@echo "Vyntrio OS — development commands"
	@echo ""
	@echo "  make bootstrap     Check toolchain and prepare workspace"
	@echo "  make verify        Validate repo layout and required docs"
	@echo "  make fmt           Format Go sources (when present)"
	@echo "  make lint          Run available linters"
	@echo "  make test          Run all available tests"
	@echo "  make test-go       Run Go tests"
	@echo "  make test-frontend Run frontend tests"
	@echo "  make build         Build Go binaries (foundation stubs)"
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

lint:
	@$(GO) vet ./...
	@cd frontend && $(NPM) run lint --if-present

test: test-go test-frontend

test-go:
	@$(GO) test ./...

test-frontend:
	@cd frontend && $(NPM) test --if-present

build:
	@mkdir -p bin
	@$(GO) build -o bin/vyntrio-api ./cmd/api
	@$(GO) build -o bin/vyntrio-worker ./cmd/worker
	@$(GO) build -o bin/vyntrio-installer ./cmd/installer
	@$(GO) build -o bin/vyntrio-update-agent ./cmd/update-agent

run-api:
	@$(GO) run ./cmd/api

docs-check:
	@./scripts/docs-check.sh

clean:
	@rm -rf bin dist coverage.out coverage.html
	@rm -rf frontend/dist frontend/build frontend/.next
	@echo "Clean complete."

sqlc-generate:
	@command -v sqlc >/dev/null 2>&1 || { echo "sqlc not installed; install from https://docs.sqlc.dev/en/latest/overview/install.html" >&2; exit 1; }
	@sqlc generate

generate: sqlc-generate
