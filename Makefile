.PHONY: help generate build test lint clean

.DEFAULT_GOAL := help

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# Directories
SRC_DIR := src
CMD_DIR := $(SRC_DIR)/cmd/daemon
BIN_DIR := bin

# Binary name
BINARY := daemon

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

generate: ## Generate code (Wire dependency injection)
	@echo "==> Generating Wire code..."
	@cd $(SRC_DIR) && wire ./internal/bootstrap/

build: generate ## Build the daemon binary
	@echo "==> Building $(BINARY)..."
	@mkdir -p $(BIN_DIR)
	@cd $(SRC_DIR) && CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -o ../$(BIN_DIR)/$(BINARY) ./cmd/daemon

test: generate ## Run tests with race detection
	@echo "==> Running tests..."
	@cd $(SRC_DIR) && $(GOTEST) -race -coverprofile=coverage.out ./...

lint: generate ## Run linters (ktn-linter + golangci-lint)
	@echo "==> Running linters..."
	@cd $(SRC_DIR) && golangci-lint run
	@ktn-linter lint -c $(SRC_DIR)/.ktn-linter.yaml ./...

clean: ## Clean build artifacts and generated files
	@echo "==> Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f $(SRC_DIR)/coverage.out
	@find $(SRC_DIR) -name '*_gen.go' -delete

deps: ## Download and tidy dependencies
	@echo "==> Tidying dependencies..."
	@cd $(SRC_DIR) && $(GOMOD) tidy

all: clean generate lint test build ## Run full pipeline
