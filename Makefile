# superviz.io - Hybrid Go+Rust Build System
#
# This Makefile handles building both the Rust probe library (libprobe.a)
# and the Go daemon binary, with cross-compilation support.

.PHONY: help build build-e2e build-probe build-daemon build-go-only build-all
.PHONY: run run-dev run-tui run-go-only
.PHONY: lint lint-golangci lint-ktn lint-probe fmt
.PHONY: test test-unit test-e2e test-probe coverage
.PHONY: clean clean-probe clean-go dirs header info
.PHONY: install-rust install-rust-targets install-cbindgen

.DEFAULT_GOAL := help

# ==============================================================================
# CONFIGURATION
# ==============================================================================

# Colors
CYAN := \033[36m
YELLOW := \033[33m
GREEN := \033[32m
RESET := \033[0m

# Version from git
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Rust targets for cross-compilation (Unix only)
RUST_TARGETS := x86_64-unknown-linux-gnu \
                aarch64-unknown-linux-gnu \
                x86_64-apple-darwin \
                aarch64-apple-darwin \
                x86_64-unknown-freebsd

# Go targets for cross-compilation (Unix only)
GO_TARGETS := linux/amd64 \
              linux/arm64 \
              darwin/amd64 \
              darwin/arm64 \
              freebsd/amd64

# Directories
SRC_DIR := src
PROBE_DIR := $(SRC_DIR)/lib/probe
DIST_DIR := dist
DIST_BIN := $(DIST_DIR)/bin
DIST_LIB := $(DIST_DIR)/lib
INCLUDE_DIR := $(PROBE_DIR)/include
BIN_DIR := bin

# Binary name
BINARY_NAME := supervizio

# Detect current platform
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
    ifeq ($(UNAME_M),x86_64)
        CURRENT_RUST_TARGET := x86_64-unknown-linux-gnu
        CURRENT_PLATFORM := linux-amd64
        CURRENT_GO_ARCH := amd64
    else ifeq ($(UNAME_M),aarch64)
        CURRENT_RUST_TARGET := aarch64-unknown-linux-gnu
        CURRENT_PLATFORM := linux-arm64
        CURRENT_GO_ARCH := arm64
    endif
    CURRENT_GO_OS := linux
endif

ifeq ($(UNAME_S),Darwin)
    ifeq ($(UNAME_M),arm64)
        CURRENT_RUST_TARGET := aarch64-apple-darwin
        CURRENT_PLATFORM := darwin-arm64
        CURRENT_GO_ARCH := arm64
    else
        CURRENT_RUST_TARGET := x86_64-apple-darwin
        CURRENT_PLATFORM := darwin-amd64
        CURRENT_GO_ARCH := amd64
    endif
    CURRENT_GO_OS := darwin
endif

ifeq ($(UNAME_S),FreeBSD)
    CURRENT_RUST_TARGET := x86_64-unknown-freebsd
    CURRENT_PLATFORM := freebsd-amd64
    CURRENT_GO_OS := freebsd
    CURRENT_GO_ARCH := amd64
endif

# Defaults if not detected
CURRENT_PLATFORM ?= linux-amd64
CURRENT_GO_OS ?= linux
CURRENT_GO_ARCH ?= amd64
CURRENT_RUST_TARGET ?= x86_64-unknown-linux-gnu

# CGO flags for linking the Rust library
CGO_CFLAGS := -I$(abspath $(INCLUDE_DIR))
CGO_LDFLAGS := -L$(abspath $(DIST_LIB)/$(CURRENT_PLATFORM)) -lprobe -lpthread -ldl -lm

# ==============================================================================
# HELP
# ==============================================================================

help: ## Show this help
	@echo ""
	@echo "$(CYAN)superviz.io Build System$(RESET)"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "$(YELLOW)Main Targets:$(RESET)"
	@grep -E '^[a-zA-Z0-9_-]+:.*## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Run 'make info' for build configuration details.$(RESET)"
	@echo ""

# ==============================================================================
# MAIN BUILD TARGETS
# ==============================================================================

build: dirs build-probe build-daemon ## Build supervizio (Rust probe + Go daemon)
	@echo "$(GREEN)Build complete: $(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME)$(RESET)"

build-probe: dirs ## Build Rust probe library (libprobe.a)
	@echo "$(CYAN)Building Rust probe for $(CURRENT_RUST_TARGET)...$(RESET)"
	@if ! command -v cargo >/dev/null 2>&1; then \
		echo "$(YELLOW)ERROR: Rust/Cargo not installed. Install from https://rustup.rs$(RESET)"; \
		echo "$(YELLOW)Run 'make install-rust' to install.$(RESET)"; \
		exit 1; \
	fi
	cd $(PROBE_DIR) && cargo build --release --target $(CURRENT_RUST_TARGET)
	cp $(PROBE_DIR)/target/$(CURRENT_RUST_TARGET)/release/libprobe.a $(DIST_LIB)/$(CURRENT_PLATFORM)/
	@echo "$(GREEN)Probe built: $(DIST_LIB)/$(CURRENT_PLATFORM)/libprobe.a$(RESET)"

build-daemon: ## Build Go daemon with Rust probe linked
	@echo "$(CYAN)Building Go daemon for $(CURRENT_GO_OS)/$(CURRENT_GO_ARCH)...$(RESET)"
	@if [ ! -f "$(DIST_LIB)/$(CURRENT_PLATFORM)/libprobe.a" ]; then \
		echo "$(YELLOW)ERROR: libprobe.a not found. Run 'make build-probe' first.$(RESET)"; \
		exit 1; \
	fi
	@mkdir -p $(DIST_BIN)/$(CURRENT_PLATFORM)
	cd $(SRC_DIR) && \
	    CGO_ENABLED=1 \
	    CGO_CFLAGS="$(CGO_CFLAGS)" \
	    CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	    go build -ldflags="-s -w -X github.com/kodflow/daemon/internal/bootstrap.version=$(VERSION)" \
	    -o ../$(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME) ./cmd/daemon
	@echo "$(GREEN)Daemon built: $(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME)$(RESET)"

build-go-only: dirs ## Build Go daemon without CGO (fallback, no Rust)
	@echo "$(CYAN)Building Go daemon (without Rust probe)...$(RESET)"
	@mkdir -p $(DIST_BIN)/$(CURRENT_PLATFORM)
	cd $(SRC_DIR) && \
	    CGO_ENABLED=0 \
	    go build -ldflags="-s -w -X github.com/kodflow/daemon/internal/bootstrap.version=$(VERSION)" \
	    -o ../$(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME) ./cmd/daemon
	@echo "$(GREEN)Daemon built (no CGO): $(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME)$(RESET)"

build-e2e: build ## Build E2E test binaries (supervizio + crasher)
	@mkdir -p $(BIN_DIR)
	@cd e2e/behavioral/crasher && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../../../$(BIN_DIR)/crasher .

dirs: ## Create output directories
	@mkdir -p $(DIST_BIN)/$(CURRENT_PLATFORM)
	@mkdir -p $(DIST_LIB)/$(CURRENT_PLATFORM)

# ==============================================================================
# CROSS-COMPILATION
# ==============================================================================

build-all: $(addprefix build-probe-,$(RUST_TARGETS)) build-go-all ## Build for all supported platforms
	@echo "$(GREEN)All platforms built successfully!$(RESET)"

build-go-all: ## Build Go for all platforms (requires probe libs to exist)
	@echo "$(CYAN)Building Go for all platforms...$(RESET)"
	@for target in $(GO_TARGETS); do \
	    os=$$(echo $$target | cut -d/ -f1); \
	    arch=$$(echo $$target | cut -d/ -f2); \
	    platform="$$os-$$arch"; \
	    echo "Building Go for $$platform..."; \
	    mkdir -p $(DIST_BIN)/$$platform; \
	    if [ -f "$(DIST_LIB)/$$platform/libprobe.a" ]; then \
	        cd $(SRC_DIR) && \
	            GOOS=$$os GOARCH=$$arch CGO_ENABLED=1 \
	            CGO_CFLAGS="-I$(abspath $(INCLUDE_DIR))" \
	            CGO_LDFLAGS="-L$(abspath $(DIST_LIB)/$$platform) -lprobe -lpthread -ldl -lm" \
	            go build -ldflags="-s -w" -o ../$(DIST_BIN)/$$platform/$(BINARY_NAME) ./cmd/daemon 2>/dev/null || \
	        (cd $(SRC_DIR) && \
	            GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
	            go build -ldflags="-s -w" -o ../$(DIST_BIN)/$$platform/$(BINARY_NAME) ./cmd/daemon && \
	            echo "  Built without CGO"); \
	    else \
	        cd $(SRC_DIR) && \
	            GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
	            go build -ldflags="-s -w" -o ../$(DIST_BIN)/$$platform/$(BINARY_NAME) ./cmd/daemon; \
	        echo "  Built without CGO (no probe lib)"; \
	    fi; \
	done

# Build Rust probe for a specific target
build-probe-%:
	@echo "$(CYAN)Building Rust probe for $*...$(RESET)"
	@platform=$$(echo "$*" | sed 's/x86_64-unknown-linux-gnu/linux-amd64/' | \
	    sed 's/aarch64-unknown-linux-gnu/linux-arm64/' | \
	    sed 's/x86_64-apple-darwin/darwin-amd64/' | \
	    sed 's/aarch64-apple-darwin/darwin-arm64/' | \
	    sed 's/x86_64-unknown-freebsd/freebsd-amd64/'); \
	mkdir -p $(DIST_LIB)/$$platform; \
	if command -v cargo >/dev/null 2>&1; then \
	    cd $(PROBE_DIR) && cargo build --release --target $* && \
	    cp $(PROBE_DIR)/target/$*/release/libprobe.a $(DIST_LIB)/$$platform/ && \
	    echo "$(GREEN)Probe built for $*: $(DIST_LIB)/$$platform/libprobe.a$(RESET)"; \
	else \
	    echo "$(YELLOW)Warning: Cargo not available, skipping $*$(RESET)"; \
	fi

# ==============================================================================
# RUN TARGETS
# ==============================================================================

run: build ## Run the daemon (raw mode by default, requires /etc/daemon/config.yaml)
	@./$(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME)

run-dev: build ## Run the daemon with examples/config-dev.yaml (raw mode)
	@./$(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME) --config=examples/config-dev.yaml

run-tui: build ## Run the daemon in interactive TUI mode
	@./$(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME) --config=examples/config-dev.yaml --tui

run-go-only: build-go-only ## Run the Go-only build (no Rust probe)
	@./$(DIST_BIN)/$(CURRENT_PLATFORM)/$(BINARY_NAME)

# ==============================================================================
# HEADER GENERATION
# ==============================================================================

header: ## Generate C header from Rust code (requires cbindgen)
	@echo "$(CYAN)Generating C header...$(RESET)"
	@if command -v cbindgen >/dev/null 2>&1; then \
		cd $(PROBE_DIR)/crates/probe-ffi && \
		cbindgen --config cbindgen.toml --crate probe-ffi --output ../../include/probe.h; \
		echo "$(GREEN)Header generated: $(INCLUDE_DIR)/probe.h$(RESET)"; \
	else \
		echo "$(YELLOW)Warning: cbindgen not installed. Install with: cargo install cbindgen$(RESET)"; \
	fi

# ==============================================================================
# TESTING
# ==============================================================================

test: test-unit test-e2e ## Run all tests (unit + E2E)

test-unit: ## Run unit tests with race detection
	@cd $(SRC_DIR) && go test -race ./...

test-probe: ## Run Rust tests
	@echo "$(CYAN)Testing Rust probe...$(RESET)"
	@if command -v cargo >/dev/null 2>&1; then \
		cd $(PROBE_DIR) && cargo test; \
	else \
		echo "$(YELLOW)Warning: Cargo not available, skipping Rust tests$(RESET)"; \
	fi

test-e2e: build-e2e ## Run E2E behavioral tests (requires Docker)
	@cd e2e/behavioral && go test -v -timeout 15m ./...

test-hybrid: build-probe ## Run Go tests with Rust probe linked
	@echo "$(CYAN)Testing Go code with Rust probe...$(RESET)"
	cd $(SRC_DIR) && \
	    CGO_ENABLED=1 \
	    CGO_CFLAGS="$(CGO_CFLAGS)" \
	    CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	    go test -race ./...

coverage: ## Run unit tests with coverage report
	@cd $(SRC_DIR) && go test -race -coverprofile=coverage.out ./...
	@cd $(SRC_DIR) && go tool cover -func=coverage.out | tail -1
	@rm -f $(SRC_DIR)/coverage.out

# ==============================================================================
# LINTING
# ==============================================================================

lint: lint-golangci lint-ktn ## Run all linters

lint-golangci: ## Run golangci-lint (CGO required)
	@if [ -f "$(DIST_LIB)/$(CURRENT_PLATFORM)/libprobe.a" ]; then \
		cd $(SRC_DIR) && CGO_ENABLED=1 \
		CGO_CFLAGS="$(CGO_CFLAGS)" \
		CGO_LDFLAGS="$(CGO_LDFLAGS)" \
		golangci-lint run -c ../.golangci.yml; \
	else \
		echo "$(YELLOW)Warning: libprobe.a not found. Skipping probe package in lint.$(RESET)"; \
		cd $(SRC_DIR) && golangci-lint run -c ../.golangci.yml \
			--skip-dirs=internal/infrastructure/probe \
			./cmd/... ./internal/application/... ./internal/domain/... \
			./internal/infrastructure/observability/... \
			./internal/infrastructure/persistence/... \
			./internal/infrastructure/process/... \
			./internal/infrastructure/transport/... \
			./internal/infrastructure/discovery/... \
			./internal/bootstrap/...; \
	fi

lint-ktn: ## Run ktn-linter
	@cd $(SRC_DIR) && ktn-linter lint --no-cache -c ../.ktn-linter.yaml ./...

lint-probe: ## Run Rust linter (clippy)
	@echo "$(CYAN)Linting Rust probe...$(RESET)"
	@if command -v cargo >/dev/null 2>&1; then \
		cd $(PROBE_DIR) && cargo clippy -- -D warnings; \
	else \
		echo "$(YELLOW)Warning: Cargo not available, skipping Rust linting$(RESET)"; \
	fi

fmt: ## Format code with gofmt
	@cd $(SRC_DIR) && find . -name "*.go" -type f -exec gofmt -w {} +

fmt-rust: ## Format Rust code
	@if command -v cargo >/dev/null 2>&1; then \
		cd $(PROBE_DIR) && cargo fmt; \
	fi

# ==============================================================================
# CLEAN
# ==============================================================================

clean: ## Remove all build artifacts
	@rm -rf $(BIN_DIR)/
	@rm -rf $(DIST_DIR)/
	@if command -v cargo >/dev/null 2>&1 && [ -d "$(PROBE_DIR)" ]; then \
		cd $(PROBE_DIR) && cargo clean 2>/dev/null || true; \
	fi
	@cd $(SRC_DIR) && go clean
	@echo "$(GREEN)Clean complete$(RESET)"

clean-probe: ## Remove only Rust build artifacts
	@if command -v cargo >/dev/null 2>&1 && [ -d "$(PROBE_DIR)" ]; then \
		cd $(PROBE_DIR) && cargo clean; \
	fi
	@rm -rf $(DIST_LIB)

clean-go: ## Remove only Go build artifacts
	@cd $(SRC_DIR) && go clean
	@rm -rf $(BIN_DIR)
	@rm -rf $(DIST_BIN)

# ==============================================================================
# INSTALLATION HELPERS
# ==============================================================================

install-rust: ## Install Rust toolchain
	@echo "$(CYAN)Installing Rust...$(RESET)"
	curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
	@echo "$(YELLOW)Please restart your shell or run: source ~/.cargo/env$(RESET)"

install-rust-targets: ## Install Rust cross-compilation targets (Unix only)
	@echo "$(CYAN)Installing Rust cross-compilation targets...$(RESET)"
	rustup target add x86_64-unknown-linux-gnu
	rustup target add aarch64-unknown-linux-gnu
	rustup target add x86_64-apple-darwin
	rustup target add aarch64-apple-darwin
	rustup target add x86_64-unknown-freebsd
	@echo "$(GREEN)Targets installed$(RESET)"

install-cbindgen: ## Install cbindgen for header generation
	cargo install cbindgen

# ==============================================================================
# INFO
# ==============================================================================

info: ## Show build configuration
	@echo ""
	@echo "$(CYAN)Build Configuration:$(RESET)"
	@echo "  Platform:     $(CURRENT_PLATFORM)"
	@echo "  Rust Target:  $(CURRENT_RUST_TARGET)"
	@echo "  Go OS/Arch:   $(CURRENT_GO_OS)/$(CURRENT_GO_ARCH)"
	@echo ""
	@echo "$(CYAN)Directories:$(RESET)"
	@echo "  Source:       $(SRC_DIR)"
	@echo "  Probe:        $(PROBE_DIR)"
	@echo "  Bin:          $(BIN_DIR)"
	@echo "  Dist (bin):   $(DIST_BIN)"
	@echo "  Dist (lib):   $(DIST_LIB)"
	@echo ""
	@echo "$(CYAN)Tools:$(RESET)"
	@printf "  Cargo:        "; command -v cargo >/dev/null 2>&1 && cargo --version || echo "$(YELLOW)NOT INSTALLED$(RESET)"
	@printf "  Go:           "; command -v go >/dev/null 2>&1 && go version || echo "$(YELLOW)NOT INSTALLED$(RESET)"
	@printf "  cbindgen:     "; command -v cbindgen >/dev/null 2>&1 && cbindgen --version || echo "$(YELLOW)NOT INSTALLED$(RESET)"
	@echo ""
