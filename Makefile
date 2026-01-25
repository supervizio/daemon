.PHONY: help build build-e2e run run-dev run-raw lint fmt test test-unit test-e2e coverage clean

.DEFAULT_GOAL := help

# Colors
CYAN := \033[36m
RESET := \033[0m

# Version from git
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

help: ## Show this help
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""

build: ## Build supervizio binary to bin/
	@cd src && CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/kodflow/daemon/internal/bootstrap.version=$(VERSION)" -o ../bin/supervizio ./cmd/daemon

run: build ## Run the daemon with auto-detected TUI mode (requires /etc/daemon/config.yaml)
	@./bin/supervizio

run-dev: build ## Run the daemon with examples/config-dev.yaml (interactive if TTY)
	@./bin/supervizio --config=examples/config-dev.yaml

run-raw: build ## Run the daemon in raw mode (MOTD once, then silent)
	@./bin/supervizio --config=examples/config-dev.yaml --raw

run-tui: build ## Run the daemon in interactive TUI mode
	@./bin/supervizio --config=examples/config-dev.yaml --tui

build-e2e: build ## Build E2E test binaries (supervizio + crasher)
	@cd e2e/behavioral/crasher && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../../../bin/crasher .

lint: ## Run golangci-lint
	@cd src && golangci-lint run

fmt: ## Format code with gofmt
	@cd src && go fmt ./...

test: test-unit test-e2e ## Run all tests (unit + E2E)

test-unit: ## Run unit tests with race detection
	@cd src && go test -race ./...

test-e2e: build-e2e ## Run E2E behavioral tests (requires Docker)
	@cd e2e/behavioral && go test -v -timeout 15m ./...

coverage: ## Run unit tests with coverage report
	@cd src && go test -race -coverprofile=coverage.out ./...
	@cd src && go tool cover -func=coverage.out | tail -1
	@rm -f src/coverage.out

clean: ## Remove build artifacts
	@rm -rf bin/
