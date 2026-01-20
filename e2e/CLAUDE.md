# E2E Testing - Virtual Machine Tests

End-to-end testing using native VMs with Vagrant.

## Structure

```
e2e/
├── Vagrantfile         # Multi-VM configuration (libvirt + QEMU)
└── test-install.sh     # Installation test script
```

## CI Testing (Linux + KVM)

GitHub Actions uses `ubuntu-latest` with KVM acceleration:

- **Provider**: libvirt with KVM
- **Speed**: Fast (hardware virtualization)
- **Platforms**: amd64 native, arm64 via QEMU

## Local Testing (macOS)

For local development on macOS:

### Requirements

- Vagrant 2.4+
- QEMU + vagrant-qemu plugin

### Setup

```bash
brew install vagrant qemu
vagrant plugin install vagrant-qemu
```

## Available VMs

| VM | Box | Init System | autostart |
|----|-----|-------------|-----------|
| debian | generic/debian12 | systemd | yes |
| alpine | generic/alpine319 | OpenRC | no |
| ubuntu | generic/ubuntu2204 | systemd | no |
| rocky | generic/rocky9 | systemd | no |
| freebsd | generic/freebsd14 | rc.d | no |
| openbsd | generic/openbsd75 | rc.d | no |
| netbsd | generic/netbsd10 | rc.d | no |

## Usage

### From Makefile

```bash
make test-e2e          # Run default E2E test (Debian)
make test-e2e-debian   # Run Debian specifically
make test-e2e-clean    # Clean up all VMs
```

### Direct Vagrant

```bash
cd e2e

# macOS (QEMU)
vagrant up debian --provider=qemu
vagrant ssh debian -c "sudo /vagrant/test-install.sh"
vagrant destroy debian -f

# Linux (libvirt)
vagrant up debian --provider=libvirt
vagrant ssh debian -c "sudo /vagrant/test-install.sh"
vagrant destroy debian -f
```

## Test Coverage

`test-install.sh` validates:

1. Install script completes successfully
2. Binary exists at `/usr/local/bin/supervizio`
3. Config directory exists
4. Config file `config.yaml` exists
5. Service file installed for platform
6. `--version` command works
7. Uninstall removes binary
