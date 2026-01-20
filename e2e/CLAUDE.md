# E2E Testing - Native Architecture Testing

End-to-end testing for supervizio using native macOS runners.

## Architecture

**Native hardware testing on macOS runners**

| Runner | Architecture | VM | Container |
|--------|--------------|-----|-----------|
| macos-13 | AMD64 (Intel) | Debian | Debian |
| macos-14 | ARM64 (Apple Silicon) | Debian | Debian |

**Total: 2 jobs** (1 AMD64 + 1 ARM64)

Each job tests:
1. **VM Test**: Debian via Vagrant + QEMU with HVF acceleration
2. **Container Test**: Debian via Docker

## Structure

```
e2e/
├── Vagrantfile           # VM configuration (QEMU provider)
├── test-install.sh       # VM installation test script
├── Dockerfile.debian     # Debian container for testing
├── Dockerfile.scratch    # Minimal scratch image
├── Dockerfile.pid1       # PID1 test container
├── config-pid1.yaml      # Config for PID1 container
├── config-scratch.yaml   # Config for scratch container
└── start-nginx.sh        # Nginx wrapper for PID1 tests
```

## CI Workflow

```
E2E Tests (2 jobs)
├── E2E AMD64 (Debian VM + Container)
│   ├── Build linux/amd64 binary
│   ├── Vagrant up debian (QEMU)
│   ├── Run test-install.sh
│   ├── Docker build Dockerfile.debian
│   └── Docker run --version test
│
└── E2E ARM64 (Debian VM + Container)
    ├── Build linux/arm64 binary
    ├── Vagrant up debian (QEMU)
    ├── Run test-install.sh
    ├── Docker build Dockerfile.debian
    └── Docker run --version test
```

## Runners

| Runner | Hardware | Provider |
|--------|----------|----------|
| `macos-13` | Intel x86_64 | QEMU + HVF |
| `macos-14` | Apple Silicon M1 | QEMU + HVF |

Both use HVF (Hypervisor Framework) for native virtualization performance.

## VM Tests (test-install.sh)

1. Install script completes successfully
2. Binary exists at `/usr/local/bin/supervizio`
3. Config directory exists
4. Config file `config.yaml` exists
5. Service file installed for platform
6. `--version` command works
7. Uninstall removes binary

## Container Tests

| Test | Description |
|------|-------------|
| Debian | Validates binary runs in standard Debian container |
| Scratch | Validates static binary runs in minimal container |
| PID1 | Verifies supervizio as container init |

## Usage

### VM Tests (Vagrant)

```bash
cd e2e

# Start VM with QEMU (macOS)
vagrant up debian --provider=qemu
vagrant ssh debian -c "sudo /vagrant/test-install.sh"
vagrant destroy debian -f
```

### Container Tests (Docker)

```bash
# Build binary first
cd src && CGO_ENABLED=0 go build -o ../bin/supervizio ./cmd/daemon

# Debian container test
docker build -f e2e/Dockerfile.debian -t supervizio-debian .
docker run --rm supervizio-debian

# Scratch test
docker build -f e2e/Dockerfile.scratch -t supervizio-scratch .
docker run --rm --entrypoint /supervizio supervizio-scratch --version
```

## Extensibility

Additional distros can be enabled in Vagrantfile:

```ruby
# Uncomment to enable Ubuntu
config.vm.define "ubuntu", autostart: false do |v|
  v.vm.box = "generic/ubuntu2204"
  v.vm.hostname = "e2e-ubuntu"
end

# Uncomment to enable Alpine
config.vm.define "alpine", autostart: false do |v|
  v.vm.box = "generic/alpine319"
  v.vm.hostname = "e2e-alpine"
end

# Uncomment to enable FreeBSD
config.vm.define "freebsd", autostart: false do |v|
  v.vm.box = "generic/freebsd14"
  v.vm.hostname = "e2e-freebsd"
end
```

## Build Requirements

### Go Cross-Compilation

| Platform | GOOS | GOARCH |
|----------|------|--------|
| Linux AMD64 | linux | amd64 |
| Linux ARM64 | linux | arm64 |

### CI Dependencies (macOS)

- Vagrant (via Homebrew cask)
- QEMU (via Homebrew)
- vagrant-qemu plugin
- Docker (via setup-buildx-action)
