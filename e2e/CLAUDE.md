# E2E Testing - Virtual Machine & Container Tests

End-to-end testing for supervizio across VMs and containers.

## Structure

```
e2e/
├── Vagrantfile           # Multi-VM configuration (libvirt + QEMU)
├── test-install.sh       # VM installation test script
├── test-container.sh     # Container PID1 test script
├── Dockerfile.pid1       # Container with supervizio as PID1
├── Dockerfile.scratch    # Minimal scratch image
├── config-pid1.yaml      # Config for PID1 container
└── config-scratch.yaml   # Config for scratch container
```

## Test Types

### 1. VM Tests (e2e.yml)

Tests installation on real virtual machines.

| Architecture | Method | Speed |
|--------------|--------|-------|
| AMD64 | KVM (native) | Fast |
| ARM64 | QEMU emulation | Slow |

### 2. Container Tests (e2e-container.yml)

Tests supervizio as container PID1.

| Test | Description |
|------|-------------|
| PID1 | Validates supervizio runs as PID1 |
| Services | Verifies nginx/redis management |
| Zombies | Tests zombie process reaping |
| Signals | Validates signal forwarding |
| Restart | Tests automatic service restart |
| Scratch | Minimal image validation |

## CI Workflows

### e2e.yml - VM Testing

- **Runner**: `ubuntu-latest` with KVM
- **Provider**: libvirt
- **Matrix**: AMD64 (KVM) + ARM64 (QEMU)

### e2e-container.yml - Container Testing

- **Runner**: `ubuntu-latest`
- **Tests**: PID1 mode, scratch image

## Available VMs

### Active (CI tested)

| VM | Box | Init | Arch |
|----|-----|------|------|
| debian | generic/debian12 | systemd | amd64 |
| debian-arm64 | generic/debian12 | systemd | arm64 |

### Available (manual)

| VM | Box | Init | Arch |
|----|-----|------|------|
| alpine | generic/alpine319 | OpenRC | amd64 |
| ubuntu | generic/ubuntu2204 | systemd | amd64 |
| rocky | generic/rocky9 | systemd | amd64 |
| freebsd | generic/freebsd14 | rc.d | amd64 |
| openbsd | generic/openbsd75 | rc.d | amd64 |
| netbsd | generic/netbsd10 | rc.d | amd64 |

### Commented (future)

See `Vagrantfile` for full matrix including:
- Debian 10, 11
- Ubuntu 20.04, 24.04
- Alpine 3.18, 3.20
- Rocky 8, Alma 9, CentOS Stream 9
- Fedora 39, 40
- Arch, openSUSE
- FreeBSD 12, 13
- OpenBSD 7.3, 7.4
- NetBSD 9
- DragonflyBSD 6

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

# AMD64 (Linux CI or local)
vagrant up debian --provider=libvirt
vagrant ssh debian -c "sudo /vagrant/test-install.sh"
vagrant destroy debian -f

# ARM64 (QEMU emulation)
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
docker run -d --name test-scratch supervizio-scratch
docker exec test-scratch cat /proc/1/comm  # supervizio
docker rm -f test-scratch
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

## Expanding Coverage

### Enable More VMs

1. Uncomment desired VM in `e2e/Vagrantfile`
2. Uncomment corresponding matrix entry in `.github/workflows/e2e.yml`
3. Test locally first: `vagrant up <vm-name> --provider=libvirt`

### Add New OS

1. Add VM definition in `Vagrantfile`
2. Add matrix entry in workflow
3. Update init system detection in `setup/install.sh` if needed
4. Add service file in `setup/init/` if needed
