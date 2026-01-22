.PHONY: help build lint fmt test coverage clean

.DEFAULT_GOAL := help

# Colors
CYAN := \033[36m
RESET := \033[0m

help: ## Show this help
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  $(CYAN)%-12s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

build: ## Build binary to bin/supervizio
	@cd src && go build -o ../bin/supervizio ./cmd/daemon

lint: ## Run golangci-lint
	@cd src && golangci-lint run

fmt: ## Format code with gofmt
	@cd src && go fmt ./...

test: ## Run tests with race detection
	@cd src && go test -race ./...

coverage: ## Run tests with coverage report
	@cd src && go test -race -coverprofile=coverage.out ./...
	@cd src && go tool cover -func=coverage.out | tail -1
	@rm -f src/coverage.out

clean: ## Remove build artifacts
	@rm -rf bin/
