# E2E Testing - Full Platform Matrix

End-to-end testing for supervizio across all supported platforms and init systems.

## Complete Init System Mapping

| Init System | Distribution | AMD64 | ARM64 | GOOS | Container | VM |
|-------------|--------------|-------|-------|------|-----------|-----|
| **systemd** | Debian 12 | ✅ | ✅ | linux | ✅ | ✅ |
| **systemd** | Ubuntu 22.04 | ✅ | ✅ | linux | ✅ | ✅ |
| **OpenRC** | Alpine 3.19 | ✅ | ✅ | linux | ✅ | ✅ (AMD64 only) |
| **SysVinit** | Devuan 4 | ✅ | ❌ | linux | ✅ (AMD64 only) | ✅ (AMD64 only) |
| **runit** | Void Linux | ✅ | ✅ | linux | ✅ | ❌ (no Vagrant box) |
| **BSD rc.d** | FreeBSD 14 | ✅ | ✅* | freebsd | ❌ | ✅ |
| **BSD rc.d** | OpenBSD 7 | ✅ | ✅* | openbsd | ❌ | ✅ |
| **BSD rc.d** | NetBSD 10 | ✅ | ✅* | netbsd | ❌ | ❌ (flaky box) |
| **BSD rc.d** | DragonFlyBSD 6 | ✅ | ❌ | dragonfly | ❌ | ✅ |
| **launchd** | macOS 14 | ✅ | ✅ | darwin | ❌ | ❌ (macOS runner) |

*BSD ARM64: Go supports cross-compilation, but no CI ARM64 BSD VMs available.

## Architecture

**Full coverage of init systems and platforms**

### Linux (AMD64 + ARM64)

| Distribution | Init System | VM | Container |
|--------------|-------------|-----|-----------|
| Debian 12 | systemd | Vagrant/virt-install | Docker |
| Ubuntu 22.04 | systemd | Vagrant/virt-install | Docker |
| Alpine 3.19 | OpenRC | Vagrant (AMD64 only) | Docker |
| Devuan 4 | SysVinit | Vagrant (AMD64 only) | Docker |
| Void Linux | runit | ❌ (no Vagrant box) | Docker |

### BSD (AMD64 only)

| OS | Init System | VM | Container |
|----|-------------|-----|-----------|
| FreeBSD 14 | rc.d | Vagrant | N/A |
| OpenBSD 7 | rc.d | Vagrant | N/A |
| NetBSD 10 | rc.d | ❌ (flaky) | N/A |
| DragonFlyBSD 6 | rc.d | Vagrant | N/A |

**Total: 15 jobs** (5 Linux AMD64 + 5 Linux ARM64 + 4 BSD AMD64 + 1 PID1)

## Structure

```
e2e/
├── Vagrantfile           # VM configuration (all platforms)
├── test-install.sh       # Universal installation test script
├── test-container.sh     # PID1 container tests
├── Dockerfile.debian     # Debian 13 (systemd)
├── Dockerfile.ubuntu     # Ubuntu 24.04 (systemd)
├── Dockerfile.alpine     # Alpine 3.20 (OpenRC)
├── Dockerfile.devuan     # Devuan 5 (SysVinit)
├── Dockerfile.void       # Void Linux (runit)
├── Dockerfile.pid1       # PID1 test container
├── config-scratch.yaml   # Minimal test configuration
├── config-pid1.yaml      # PID1 test configuration
├── start-nginx.sh        # Nginx wrapper for PID1 tests
└── CLAUDE.md             # This file
```

## CI Workflow

```
E2E Tests (15 jobs)
│
├── Linux AMD64 (5 jobs - Vagrant + Docker)
│   ├── Debian (systemd) - VM + Container
│   ├── Ubuntu (systemd) - VM + Container
│   ├── Alpine (OpenRC) - VM + Container
│   ├── Devuan (SysVinit) - VM + Container
│   └── Void (runit) - Container only (no Vagrant box)
│
├── Linux ARM64 (5 jobs - Container only, virt-install unreliable on GHA)
│   ├── Debian (systemd) - Container only
│   ├── Ubuntu (systemd) - Container only
│   ├── Alpine (OpenRC) - Container only
│   ├── Devuan (SysVinit) - Skip (no ARM64 Docker image)
│   └── Void (runit) - Container only
│
├── BSD AMD64 (4 jobs - Vagrant only, no containers)
│   ├── FreeBSD (rc.d)
│   ├── OpenBSD (rc.d)
│   ├── NetBSD (rc.d) - Skip (flaky Vagrant box)
│   └── DragonFlyBSD (rc.d)
│
└── PID1 Tests (1 job - Docker)
    └── supervizio as container init (debian-based)
```

## Init Systems Tested

| Init System | Service Path | Enable Command |
|-------------|--------------|----------------|
| **systemd** | `/etc/systemd/system/` | `systemctl enable` |
| **OpenRC** | `/etc/init.d/` | `rc-update add` |
| **SysVinit** | `/etc/init.d/` | `update-rc.d` |
| **runit** | `/etc/sv/` | `ln -s /etc/sv/X /var/service/` |
| **BSD rc.d** | `/usr/local/etc/rc.d/` (FreeBSD) | `sysrc enable` |
| **BSD rc.d** | `/etc/rc.d/` (OpenBSD/NetBSD) | `rcctl enable` |

## Runners

| Runner | Hardware | Use Case |
|--------|----------|----------|
| `ubuntu-24.04` | AMD64 | Linux + BSD VMs |
| `ubuntu-24.04-arm` | ARM64 | Linux ARM64 VMs |

## VM Tests (test-install.sh)

1. Install script completes successfully
2. Binary exists at `/usr/local/bin/supervizio`
3. Config directory exists
4. Config file `config.yaml` exists
5. Service file installed for platform
6. `--version` command works
7. Uninstall removes binary

## Container Tests

| Test | Base Image | Init System |
|------|------------|-------------|
| Debian | `debian:trixie-slim` | systemd |
| Ubuntu | `ubuntu:24.04` | systemd |
| Alpine | `alpine:3.20` | OpenRC |
| Devuan | `devuan/devuan:daedalus` | SysVinit |
| Void | `ghcr.io/void-linux/void-glibc:latest` | runit |

## PID1 Tests (test-container.sh)

Tests supervizio running as container PID1 with managed services:

1. **PID1 verification** - supervizio is process 1
2. **Managed services** - nginx and redis-server running
3. **Zombie reaping** - Orphan processes are reaped
4. **Signal forwarding** - Services survive SIGHUP to PID1
5. **Service restart** - nginx restarts after being killed
6. **HTTP health check** - nginx responds HTTP 200
7. **TCP health check** - redis responds to PING

### Running PID1 Tests Locally

```bash
# Build binary
cd src && CGO_ENABLED=0 go build -o ../bin/supervizio ./cmd/daemon

# Build PID1 container
docker build -f e2e/Dockerfile.pid1 -t supervizio-pid1:test .

# Start container
docker run -d --name supervizio-pid1 \
  -p 8080:80 -p 6379:6379 \
  supervizio-pid1:test --config /etc/supervizio/config.yaml

# Wait for startup
sleep 15

# Run tests
CONTAINER_NAME=supervizio-pid1 ./e2e/test-container.sh

# Cleanup
docker stop supervizio-pid1 && docker rm supervizio-pid1
```

## Usage

### Linux VM Tests (Vagrant)

```bash
cd e2e

# systemd distributions
vagrant up debian13 --provider=libvirt
vagrant ssh debian13 -c "sudo /vagrant/test-install.sh"
vagrant destroy debian13 -f

vagrant up ubuntu --provider=libvirt
vagrant ssh ubuntu -c "sudo /vagrant/test-install.sh"
vagrant destroy ubuntu -f

# OpenRC
vagrant up alpine --provider=libvirt
vagrant ssh alpine -c "sudo /vagrant/test-install.sh"
vagrant destroy alpine -f

# SysVinit
vagrant up devuan --provider=libvirt
vagrant ssh devuan -c "sudo /vagrant/test-install.sh"
vagrant destroy devuan -f

# runit (Void Linux) - No Vagrant box available, use Docker instead
docker build -f Dockerfile.void -t supervizio-void .
docker run --rm supervizio-void
```

### BSD VM Tests (Vagrant)

```bash
cd e2e

# FreeBSD
vagrant up freebsd --provider=libvirt
vagrant ssh freebsd -c "sudo /vagrant/test-install.sh"
vagrant destroy freebsd -f

# OpenBSD
vagrant up openbsd --provider=libvirt
vagrant ssh openbsd -c "sudo /vagrant/test-install.sh"
vagrant destroy openbsd -f

# NetBSD
vagrant up netbsd --provider=libvirt
vagrant ssh netbsd -c "sudo /vagrant/test-install.sh"
vagrant destroy netbsd -f

# DragonFlyBSD
vagrant up dragonfly --provider=libvirt
vagrant ssh dragonfly -c "sudo /vagrant/test-install.sh"
vagrant destroy dragonfly -f
```

### Container Tests (Docker)

```bash
# Build binary first
cd src && CGO_ENABLED=0 go build -o ../bin/supervizio ./cmd/daemon

# Debian (systemd)
docker build -f e2e/Dockerfile.debian -t supervizio-debian .
docker run --rm supervizio-debian

# Ubuntu (systemd)
docker build -f e2e/Dockerfile.ubuntu -t supervizio-ubuntu .
docker run --rm supervizio-ubuntu

# Alpine (OpenRC)
docker build -f e2e/Dockerfile.alpine -t supervizio-alpine .
docker run --rm supervizio-alpine

# Devuan (SysVinit)
docker build -f e2e/Dockerfile.devuan -t supervizio-devuan .
docker run --rm supervizio-devuan

# Void (runit)
docker build -f e2e/Dockerfile.void -t supervizio-void .
docker run --rm supervizio-void
```

## Build Requirements

### Go Cross-Compilation

| Platform | GOOS | GOARCH |
|----------|------|--------|
| Linux AMD64 | linux | amd64 |
| Linux ARM64 | linux | arm64 |
| FreeBSD AMD64 | freebsd | amd64 |
| FreeBSD ARM64 | freebsd | arm64 |
| OpenBSD AMD64 | openbsd | amd64 |
| OpenBSD ARM64 | openbsd | arm64 |
| NetBSD AMD64 | netbsd | amd64 |
| NetBSD ARM64 | netbsd | arm64 |
| DragonFlyBSD AMD64 | dragonfly | amd64 |
| macOS AMD64 | darwin | amd64 |
| macOS ARM64 | darwin | arm64 |

### CI Dependencies (Ubuntu)

- libvirt-daemon-system
- libvirt-dev, libvirt-clients
- vagrant, vagrant-libvirt
- qemu-kvm
- Docker (pre-installed)

## Platform Limitations

### AMD64

- **Void Linux**: No libvirt Vagrant box available - container test only
- **NetBSD**: generic/netbsd10 Vagrant box is flaky on GitHub Actions - skipped

### ARM64

- **All ARM64 VMs disabled**: virt-install is unreliable on GitHub Actions ARM64 runners
- **Vagrant**: Not available for ARM64 Linux (HashiCorp limitation)
- **Devuan**: No ARM64 Docker image - completely skipped
- **BSD ARM64**: Go supports cross-compilation, but no ARM64 VM images in CI
- **Container tests work**: All ARM64 container tests pass reliably

### Not Tested (out of scope)

- **launchd (macOS)**: Requires macOS runners, tested manually
- **s6/dinit**: Niche init systems, not in mainstream CI

## Notes

- All binaries must be statically compiled (`CGO_ENABLED=0`)
- Alpine requires static binary (musl libc incompatible with glibc)
- BSD systems don't support Docker containers
- runit uses symlinks for service management (`/var/service/`)
