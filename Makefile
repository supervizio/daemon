.PHONY: coverage

coverage:
	@cd src && gotestsum --format pkgname -- -race -coverprofile=coverage.out \
		-coverpkg=./internal/... ./internal/...
	@echo ""
	@cd src && go tool cover -func=coverage.out | tail -1
	@rm -f src/coverage.out
