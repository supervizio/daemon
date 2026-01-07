.PHONY: help lint

.DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

lint: ## Run linters (ktn-linter + golangci-lint)
	@ktn-linter lint -c src/.ktn-linter.yaml ./...
