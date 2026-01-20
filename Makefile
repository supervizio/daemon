.PHONY: coverage build test-e2e test-e2e-setup test-e2e-debian test-e2e-alpine test-e2e-freebsd test-e2e-clean

# Go build
build:
	@cd src && go build -o ../bin/supervizio ./cmd/daemon

# Coverage
coverage:
	@cd src && gotestsum --format pkgname -- -race -coverprofile=coverage.out \
		-coverpkg=./internal/... ./internal/...
	@echo ""
	@cd src && go tool cover -func=coverage.out | tail -1
	@rm -f src/coverage.out

# E2E Testing with Vagrant VMs
E2E_DIR := e2e

# Setup: ensure vagrant-qemu plugin is installed
test-e2e-setup:
	@command -v vagrant >/dev/null 2>&1 || { echo "Error: vagrant not installed"; exit 1; }
	@command -v qemu-system-aarch64 >/dev/null 2>&1 || command -v qemu-system-x86_64 >/dev/null 2>&1 || { echo "Error: QEMU not installed"; exit 1; }
	@vagrant plugin list 2>/dev/null | grep -q vagrant-qemu || { echo "Installing vagrant-qemu plugin..."; vagrant plugin install vagrant-qemu; }

test-e2e: test-e2e-debian
	@echo "E2E tests completed"

test-e2e-debian: test-e2e-setup
	@echo "=== E2E Test: Debian ==="
	@cd $(E2E_DIR) && vagrant up debian --provider=qemu --provision
	@cd $(E2E_DIR) && vagrant ssh debian -c "sudo /vagrant/test-install.sh"
	@cd $(E2E_DIR) && vagrant destroy debian -f

test-e2e-alpine: test-e2e-setup
	@echo "=== E2E Test: Alpine ==="
	@cd $(E2E_DIR) && vagrant up alpine --provider=qemu --provision
	@cd $(E2E_DIR) && vagrant ssh alpine -c "sudo /vagrant/test-install.sh"
	@cd $(E2E_DIR) && vagrant destroy alpine -f

test-e2e-freebsd: test-e2e-setup
	@echo "=== E2E Test: FreeBSD ==="
	@cd $(E2E_DIR) && vagrant up freebsd --provider=qemu --provision
	@cd $(E2E_DIR) && vagrant ssh freebsd -c "sudo /vagrant/test-install.sh"
	@cd $(E2E_DIR) && vagrant destroy freebsd -f

test-e2e-clean:
	@cd $(E2E_DIR) && vagrant destroy -f 2>/dev/null || true
