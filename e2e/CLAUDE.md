# E2E Testing - Linux VM Testing with KVM

End-to-end testing for supervizio using Linux runners with KVM acceleration.

## Architecture

**Linux runners with KVM for native virtualization performance**

| Runner | Architecture | Cores | RAM | VM | Container |
|--------|--------------|-------|-----|-----|-----------|
| ubuntu-24.04 | AMD64 | 4 | 16GB | Debian | Debian |
| ubuntu-24.04-arm | ARM64 | 4 | 16GB | Debian | Debian |

**Total: 2 jobs** (1 AMD64 + 1 ARM64)

- **AMD64**: VM test (Vagrant + KVM) + Container test (Docker)
- **ARM64**: Container test (Docker) - VM tests require KVM (not available on GitHub ARM64 runners)

### ARM64 VM Testing

**CI Limitation**: GitHub ARM64 runners (`ubuntu-24.04-arm`) do not have KVM available.
QEMU TCG (software emulation) is too slow for reliable VM testing in CI timeouts.

**Local ARM64 VM Testing** (with KVM available):
```bash
# Download Debian ARM64 cloud image
wget https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-generic-arm64.qcow2

# Create VM with virt-install (requires KVM)
virt-install --name e2e-debian --memory 2048 --vcpus 2 \
  --disk debian.qcow2 --disk cloud-init.iso,device=cdrom \
  --os-variant debian12 --virt-type kvm --arch aarch64 --boot uefi --import
```

**Note**: HashiCorp does not provide Vagrant for ARM64 Linux.
See: https://github.com/hashicorp/vagrant-installers/issues/288

## Structure

```
e2e/
├── Vagrantfile           # VM configuration (libvirt provider)
├── test-install.sh       # VM installation test script
├── Dockerfile.debian     # Debian container for testing
└── CLAUDE.md             # This file
```

## CI Workflow

```
E2E Tests (2 jobs)
├── E2E AMD64 (Debian VM + Container)
│   ├── Build linux/amd64 binary
│   ├── Install libvirt + Vagrant
│   ├── Vagrant up debian (KVM)
│   ├── Run test-install.sh
│   ├── Docker build Dockerfile.debian
│   └── Docker run --version test
│
└── E2E ARM64 (Debian VM + Container)
    ├── Build linux/arm64 binary
    ├── Install libvirt + virt-install
    ├── Download Debian cloud image (ARM64)
    ├── Create cloud-init config
    ├── virt-install debian (KVM)
    ├── Run test-install.sh via SSH
    ├── Docker build Dockerfile.debian
    └── Docker run --version test
```

## Runners

| Runner | Hardware | Cores | RAM | Provider |
|--------|----------|-------|-----|----------|
| `ubuntu-24.04` | x86_64 | 4 | 16GB | libvirt + KVM |
| `ubuntu-24.04-arm` | ARM64 | 4 | 16GB | libvirt + KVM |

Both use KVM for native virtualization performance.

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

## Usage

### VM Tests (Vagrant)

```bash
cd e2e

# Start VM with libvirt (Linux)
vagrant up debian --provider=libvirt
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
```

## Build Requirements

### Go Cross-Compilation

| Platform | GOOS | GOARCH |
|----------|------|--------|
| Linux AMD64 | linux | amd64 |
| Linux ARM64 | linux | arm64 |

### CI Dependencies (Ubuntu)

- libvirt-daemon-system
- libvirt-dev
- vagrant
- vagrant-libvirt plugin
- qemu-kvm
- Docker (pre-installed)
