# E2E Testing - Virtual Machine & Container Tests

End-to-end testing for supervizio across VMs and containers.

## Architecture

**AMD64 only (reliable, KVM-accelerated)**

| Matrix | Jobs | Description |
|--------|------|-------------|
| Linux | 3 | debian, ubuntu, alpine |
| BSD | 1 | freebsd |
| Container | 1 | amd64 |
| **TOTAL** | **5** | All required, must pass |

**Note**: ARM64/experimental tests removed to ensure merge stability.
They can be re-added once QEMU emulation is stabilized.

## Structure

```
e2e/
├── Vagrantfile           # Multi-VM configuration (14 VMs)
├── test-install.sh       # VM installation test script
├── test-container.sh     # Container PID1 test script
├── Dockerfile.pid1       # Container with supervizio as PID1
├── Dockerfile.scratch    # Minimal scratch image
├── config-pid1.yaml      # Config for PID1 container
├── config-scratch.yaml   # Config for scratch container
└── start-nginx.sh        # Nginx wrapper for PID1 tests
```

## VM Matrix

### Linux VMs (3 distros, AMD64 only)

| Distro | Init System | VM Name |
|--------|-------------|---------|
| Debian 12 | systemd | `debian` |
| Ubuntu 22.04 | systemd | `ubuntu` |
| Alpine 3.19 | OpenRC | `alpine` |

### BSD VMs (1 distro, AMD64 only)

| Distro | Init System | VM Name |
|--------|-------------|---------|
| FreeBSD 14 | rc.d | `freebsd` |

## Virtualization

| Architecture | Method | Speed |
|--------------|--------|-------|
| AMD64 | KVM (native) | Fast (~5min) |

## Container Tests

| Test | Description |
|------|-------------|
| Scratch | Validates static binary runs in minimal container |
| PID1 | Verifies supervizio as container init |
| Services | Tests nginx/redis management |
| Zombies | Tests zombie process reaping |
| Signals | Validates signal forwarding |
| Restart | Tests automatic service restart |

## CI Workflow

```
E2E Tests (5 jobs)
├── vm-linux (3 jobs)
│   ├── Linux debian (amd64)
│   ├── Linux ubuntu (amd64)
│   └── Linux alpine (amd64)
├── vm-bsd (1 job)
│   └── BSD freebsd (amd64)
└── container (1 job)
    └── Container (amd64)
        ├── scratch test
        └── pid1 test
```

## Usage

### From Makefile

```bash
make test-e2e          # Run default E2E test (Debian)
make test-e2e-debian   # Run Debian specifically
make test-e2e-clean    # Clean up all VMs
```

### VM Tests (Vagrant)

```bash
cd e2e

# Start VM with KVM acceleration
vagrant up debian --provider=libvirt
vagrant ssh debian -c "sudo /vagrant/test-install.sh"
vagrant destroy debian -f
```

### Container Tests (Docker)

```bash
# Build binary first
cd src && CGO_ENABLED=0 go build -o ../bin/supervizio ./cmd/daemon

# PID1 test
docker build -f e2e/Dockerfile.pid1 -t supervizio-pid1 .
docker run -d --name test-pid1 supervizio-pid1
./e2e/test-container.sh
docker rm -f test-pid1

# Scratch test
docker build -f e2e/Dockerfile.scratch -t supervizio-scratch .
docker run --rm --entrypoint /supervizio supervizio-scratch --version
```

## Test Coverage

### VM Tests (test-install.sh)

1. Install script completes successfully
2. Binary exists at `/usr/local/bin/supervizio`
3. Config directory exists
4. Config file `config.yaml` exists
5. Service file installed for platform
6. `--version` command works
7. Uninstall removes binary

### Container Tests (test-container.sh)

1. supervizio is PID1
2. nginx service running
3. redis service running
4. No zombie processes
5. Signals forwarded correctly
6. Service restart on crash
7. HTTP health check (nginx)
8. TCP health check (redis)

## Build Requirements

### Go Cross-Compilation

| Platform | GOOS | GOARCH |
|----------|------|--------|
| Linux AMD64 | linux | amd64 |
| Linux ARM64 | linux | arm64 |
| FreeBSD | freebsd | amd64/arm64 |
| OpenBSD | openbsd | amd64/arm64 |
| NetBSD | netbsd | amd64/arm64 |

### CI Dependencies

- libvirt-daemon-system
- qemu-kvm
- vagrant + vagrant-libvirt plugin
