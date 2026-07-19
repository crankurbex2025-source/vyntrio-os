SHELL := /bin/bash
ROOT := $(shell pwd)
GO ?= go
NPM ?= npm

# Generated embed input for the Go binary. Staged from frontend/dist by
# ui-stage; never committed. Go compilation fails if this input is absent.
UI_STAGE_DIR := internal/interfaces/http/ui/dist

.PHONY: help bootstrap verify fmt lint test test-go test-frontend ui-stage build build-write-media package-write-media test-write-media install-media install-media-stage test-install-media-stage install-media-envelope test-install-media-envelope install-media-bootability test-install-media-bootability install-media-image test-install-media-image install-media-wrap test-install-media-wrap install-media-runtime test-install-media-runtime install-media-live-rootfs test-install-media-live-rootfs install-media-hardware test-install-media-hardware install-media-initrd-swap test-install-media-initrd-swap install-media-appliance test-install-media-appliance install-media-runtime-verify test-install-media-runtime-verify release-install-media-stage test-release-install-media-stage test-write-install-media-usb installer-preflight test-installer-preflight installer-layout-plan test-installer-layout-plan installer-mutation-stub test-installer-mutation-stub installer-mutate-directories test-installer-mutate-directories installer-copy-payloads test-installer-copy-payloads installer-prepare-service test-installer-prepare-service installer-enable-service test-installer-enable-service run-api docs-check clean sqlc-generate generate

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
	@echo "  make install-media-bootability  Initialize boot/live_root stubs + image stub (USB-first foundation)"
	@echo "  make test-install-media-bootability  Verify bootability foundation output"
	@echo "  make install-media-image  Build real boot chain + emit best-available image (USB-first)"
	@echo "  make test-install-media-image  Verify boot chain / image emission (stub vs real)"
	@echo "  make install-media-wrap  Wrap boot chain into firmware-bootable image or honest fallback"
	@echo "  make test-install-media-wrap  Verify wrapper (bootable vs fallback, three axes)"
	@echo "  make install-media-runtime  Runtime boot verify (if VM) + wire first-boot dashboard path"
	@echo "  make test-install-media-runtime  Verify runtime_boot_tested honesty + first-boot wiring"
	@echo "  make install-media-live-rootfs  Compose minimal live-rootfs userland + live initramfs"
	@echo "  make test-install-media-live-rootfs  Verify live userland executes + record honesty"
	@echo "  make install-media-hardware  Add storage/network kernel modules + DHCP bring-up to live initramfs"
	@echo "  make test-install-media-hardware  Verify hardware-enabled live initramfs + record honesty"
	@echo "  make install-media-initrd-swap  Swap image initrd to the Vyntrio live initramfs + rebuild image"
	@echo "  make test-install-media-initrd-swap  Verify image boots the live initramfs (proven) + honesty"
	@echo "  make install-media-appliance  Build full USB appliance image (kernel+Vyntrio initrd+squashfs+persistence)"
	@echo "  make test-install-media-appliance  Verify appliance image layers (not a stub)"
	@echo "  make install-media  Alias for install-media-appliance (product USB appliance image)"
	@echo "  make release-install-media-stage  Stage dual-mode appliance image + version feed for /release/*"
	@echo "  make test-release-install-media-stage  Verify staged release manifest + artifact integrity"
	@echo "  make test-write-install-media-usb      Validate USB/VM install-media helper scripts (dry-run)"
	@echo "  make build-write-media                 Build install-media writer CLI"
	@echo "  make package-write-media               Cross-compile CLI writer for linux/darwin/windows"
	@echo "  make build-media-creator-native        Build native Tauri Media Creator (Linux host)"
	@echo "  make package-media-creator-native      Stage Tauri packages (NSIS/.deb/.AppImage)"
	@echo "  make test-write-media                  Test install-media writer package + CLI smoke"
	@echo "  make install-media-runtime-verify  Boot the image on a harness for real runtime/dashboard proof (fails closed w/o harness)"
	@echo "  make test-install-media-runtime-verify  Verify runtime-verify honesty + read-only (no chain mutation)"
	@echo "  make installer-preflight  Run read-only installer preflight checks"
	@echo "  make test-installer-preflight  Verify installer preflight behavior"
	@echo "  make installer-layout-plan  Validate installer target-layout manifest"
	@echo "  make test-installer-layout-plan  Verify layout plan validation"
	@echo "  make installer-mutation-stub  Run preflight-gated mutation dry-run stub"
	@echo "  make test-installer-mutation-stub  Verify mutation stub behavior"
	@echo "  make installer-mutate-directories  Create empty state dirs in target-sandbox"
	@echo "  make test-installer-mutate-directories  Verify directory mutation step"
	@echo "  make installer-copy-payloads  Copy manifest payloads to target-sandbox"
	@echo "  make test-installer-copy-payloads  Verify payload copy step"
	@echo "  make installer-prepare-service  Prepare service enablement in target-sandbox"
	@echo "  make test-installer-prepare-service  Verify service preparation step"
	@echo "  make installer-enable-service  Enable service in target-sandbox (no start)"
	@echo "  make test-installer-enable-service  Verify controlled service enablement"
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
	@$(GO) build -o bin/vyntrio-restore ./cmd/restore
	@$(GO) build -o bin/vyntrio-verify-artifact ./cmd/verify-artifact
	@$(GO) build -o bin/vyntrio-write-media ./cmd/write-media

build-write-media:
	@mkdir -p bin
	@$(GO) build -o bin/vyntrio-write-media ./cmd/write-media

package-write-media: build-write-media
	@chmod +x ./scripts/package-write-media.sh
	@./scripts/package-write-media.sh

build-media-creator-native:
	@cd desktop/vyntrio-media-creator && npm install && npm run build && npm run tauri -- build --bundles deb
	@cd desktop/vyntrio-media-creator && npm run tauri -- build --target x86_64-pc-windows-gnu --bundles nsis
	@# AppImage: prefer tauri; if linuxdeploy fails, package script accepts manually built AppImage
	@cd desktop/vyntrio-media-creator && (npm run tauri -- build --bundles appimage || true)
	@if [[ ! -f desktop/vyntrio-media-creator/src-tauri/target/release/bundle/appimage/vyntrio-media-creator-0.2.0-amd64.AppImage ]]; then \
		echo "build-media-creator-native: attempting manual AppImage via appimagetool"; \
		command -v file >/dev/null || apt-get install -y -qq file; \
		test -x /tmp/appimagetool || curl -fsSL -o /tmp/appimagetool https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage && chmod +x /tmp/appimagetool; \
		cd "desktop/vyntrio-media-creator/src-tauri/target/release/bundle/appimage" && ARCH=x86_64 APPIMAGE_EXTRACT_AND_RUN=1 /tmp/appimagetool "Vyntrio Media Creator.AppDir" "vyntrio-media-creator-0.2.0-amd64.AppImage"; \
	fi

package-media-creator-native:
	@chmod +x ./scripts/package-media-creator-native.sh
	@./scripts/package-media-creator-native.sh

test-write-media: build-write-media package-write-media release-install-media-stage
	@$(GO) test ./internal/platform/writemedia/... ./internal/platform/mediacreator/... ./internal/platform/installmediapublic/... -count=1
	@chmod +x ./tests/writemedia/write_media_cli_test.sh
	@./tests/writemedia/write_media_cli_test.sh

install-media-stage: build
	@./scripts/stage-install-media.sh

test-install-media-stage: install-media-stage
	@./tests/installmedia/stage_test.sh

install-media-envelope: install-media-stage
	@./scripts/assemble-install-media-envelope.sh

test-install-media-envelope: install-media-envelope
	@./tests/installmedia/envelope_test.sh

install-media-bootability: install-media-envelope
	@./scripts/initialize-install-media-bootability.sh

test-install-media-bootability: install-media-bootability
	@./tests/installmedia/bootability_test.sh

install-media-image: install-media-bootability
	@./scripts/build-install-media-image.sh

test-install-media-image: install-media-image
	@./tests/installmedia/image_test.sh

install-media-wrap: install-media-image
	@./scripts/wrap-install-media-image.sh

test-install-media-wrap: install-media-wrap
	@./tests/installmedia/wrapper_test.sh

install-media-runtime: install-media-wrap
	@./scripts/verify-runtime-boot.sh

test-install-media-runtime: install-media-runtime
	@./tests/installmedia/runtime_test.sh

install-media-live-rootfs: install-media-runtime
	@./scripts/compose-live-rootfs.sh

test-install-media-live-rootfs: install-media-live-rootfs
	@./tests/installmedia/live_rootfs_test.sh
	@./tests/installmedia/tls_readiness_test.sh

install-media-hardware: install-media-live-rootfs
	@./scripts/enable-live-initramfs-hardware.sh

test-install-media-hardware: install-media-hardware
	@./tests/installmedia/hardware_enable_test.sh

install-media-initrd-swap: install-media-hardware
	@./scripts/swap-live-initramfs-into-image.sh

test-install-media-initrd-swap: install-media-initrd-swap
	@./tests/installmedia/initrd_swap_test.sh

# Full USB appliance image (kernel + Vyntrio initramfs + squashfs root + persistence).
install-media-appliance: install-media-live-rootfs
	@chmod +x ./scripts/build-appliance-usb-image.sh
	@./scripts/build-appliance-usb-image.sh

test-install-media-appliance: install-media-appliance
	@chmod +x ./tests/installmedia/appliance_image_test.sh
	@./tests/installmedia/appliance_image_test.sh

# Product alias: complete appliance USB image (not host-initrd stub).
install-media: install-media-appliance

release-install-media-stage: install-media-appliance build
	@./scripts/stage-release-install-media.sh

test-release-install-media-stage: release-install-media-stage
	@./tests/release/install_media_stage_test.sh

test-write-install-media-usb: release-install-media-stage
	@./tests/release/write_install_media_usb_test.sh

# Stage 2: read-only runtime boot verification. Fails closed (non-zero) when no
# boot harness exists; the test target validates honesty + no chain mutation and
# therefore does NOT depend on this target succeeding.
install-media-runtime-verify: install-media-initrd-swap
	@./scripts/verify-live-boot-runtime.sh

test-install-media-runtime-verify: install-media-initrd-swap
	@./tests/installmedia/runtime_verify_test.sh

installer-preflight: install-media-envelope
	@./scripts/installer-preflight.sh

test-installer-preflight: installer-preflight
	@./tests/installer/preflight_test.sh

installer-layout-plan:
	@./scripts/validate-installer-layout-plan.sh

test-installer-layout-plan: installer-layout-plan
	@./tests/installer/layout_plan_test.sh

installer-mutation-stub: install-media-envelope
	@./scripts/installer-mutation-stub.sh

test-installer-mutation-stub: installer-mutation-stub
	@./tests/installer/mutation_stub_test.sh

installer-mutate-directories: install-media-envelope
	@./scripts/installer-mutate-directories.sh

test-installer-mutate-directories: installer-mutate-directories
	@./tests/installer/mutate_directories_test.sh

installer-copy-payloads: installer-mutate-directories
	@./scripts/installer-copy-payloads.sh

test-installer-copy-payloads: installer-copy-payloads
	@./tests/installer/copy_payloads_test.sh

installer-prepare-service: installer-copy-payloads
	@./scripts/installer-prepare-service.sh

test-installer-prepare-service: installer-prepare-service
	@./tests/installer/prepare_service_test.sh

installer-enable-service: installer-prepare-service
	@./scripts/installer-enable-service.sh

test-installer-enable-service: installer-enable-service
	@./tests/installer/enable_service_test.sh

run-api:
	@$(GO) run ./cmd/api

docs-check:
	@./scripts/docs-check.sh

clean:
	@rm -rf bin dist coverage.out coverage.html
	@rm -rf frontend/dist frontend/build frontend/.next
	@rm -rf "$(ROOT)/$(UI_STAGE_DIR)"
	@rm -rf distro/install-media/staging distro/install-media/envelope distro/install-media/build distro/release/staging distro/installer/dry-run distro/installer/target-sandbox
	@echo "Clean complete."

sqlc-generate:
	@command -v sqlc >/dev/null 2>&1 || { echo "sqlc not installed; install from https://docs.sqlc.dev/en/latest/overview/install.html" >&2; exit 1; }
	@sqlc generate

generate: sqlc-generate
