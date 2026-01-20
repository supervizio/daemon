# E2E Testing - Virtual Machine & Container Tests

End-to-end testing for supervizio across VMs and containers.

## Architecture

**AMD64 (required) + ARM64/experimental (soft-fail)**

| Matrix | Required | Experimental | Total |
|--------|----------|--------------|-------|
| Linux | 3 (debian, ubuntu, alpine) | 5 (ARM64 + rocky) | 8 |
| BSD | 1 (freebsd) | 5 (openbsd, netbsd, ARM64) | 6 |
| Container | 1 (amd64) | 1 (arm64) | 2 |
| **TOTAL** | **5** | **11** | **16** |

Jobs marked `[exp]` use `continue-on-error: true` - failures don't block the PR.

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

### Linux VMs (4 distros × 2 archs = 8 VMs)

| Distro | Init System | AMD64 | ARM64 |
|--------|-------------|-------|-------|
| Debian 12 | systemd | `debian` | `debian-arm64` |
| Ubuntu 22.04 | systemd | `ubuntu` | `ubuntu-arm64` |
| Alpine 3.19 | OpenRC | `alpine` | `alpine-arm64` |
| Rocky 9 | systemd | `rocky` | `rocky-arm64` |

### BSD VMs (3 distros × 2 archs = 6 VMs)

| Distro | Init System | AMD64 | ARM64 |
|--------|-------------|-------|-------|
| FreeBSD 14 | rc.d | `freebsd` | `freebsd-arm64` |
| OpenBSD 7.5 | rc.d | `openbsd` | `openbsd-arm64` |
| NetBSD 10 | rc.d | `netbsd` | `netbsd-arm64` |

## Virtualization Methods

| Architecture | Method | Speed | Use Case |
|--------------|--------|-------|----------|
| AMD64 | KVM (native) | Fast | CI default |
| ARM64 | QEMU emulation | Slow (~45min) | Cross-arch testing |

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
E2E Tests (16 parallel jobs)
├── vm-linux (8 jobs)
│   ├── Linux debian (amd64)
│   ├── Linux debian (arm64)
│   ├── Linux ubuntu (amd64)
│   ├── Linux ubuntu (arm64)
│   ├── Linux alpine (amd64)
│   ├── Linux alpine (arm64)
│   ├── Linux rocky (amd64)
│   └── Linux rocky (arm64)
├── vm-bsd (6 jobs)
│   ├── BSD freebsd (amd64)
│   ├── BSD freebsd (arm64)
│   ├── BSD openbsd (amd64)
│   ├── BSD openbsd (arm64)
│   ├── BSD netbsd (amd64)
│   └── BSD netbsd (arm64)
└── container (2 jobs)
    ├── Container (amd64)
    │   ├── scratch test
    │   └── pid1 test
    └── Container (arm64)
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

# AMD64 (KVM - fast)
vagrant up debian --provider=libvirt
vagrant ssh debian -c "sudo /vagrant/test-install.sh"
vagrant destroy debian -f

# ARM64 (QEMU - slow)
vagrant up debian-arm64 --provider=libvirt
vagrant ssh debian-arm64 -c "uname -m"  # aarch64
vagrant destroy debian-arm64 -f
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
- qemu-kvm (AMD64)
- qemu-system-arm (ARM64 emulation)
- qemu-efi-aarch64 (ARM64 UEFI)
- vagrant + vagrant-libvirt plugin
