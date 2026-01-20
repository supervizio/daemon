# E2E Testing - Virtual Machine Tests

End-to-end testing using Vagrant VMs.

## Structure

```
e2e/
├── Vagrantfile         # Multi-VM configuration
└── test-install.sh     # Installation test script
```

## Requirements

- Vagrant 2.4+
- QEMU (for macOS/Linux)
- VirtualBox (alternative)

### macOS Setup

```bash
brew install vagrant qemu
vagrant plugin install vagrant-qemu
```

## Available VMs

| VM | Box | Init System | autostart |
|----|-----|-------------|-----------|
| debian | generic/debian12 | systemd | yes |
| alpine | generic/alpine319 | OpenRC | no |
| ubuntu | generic/ubuntu2404 | systemd | no |
| rocky | generic/rocky9 | systemd | no |
| freebsd | generic/freebsd14 | rc.d | no |
| openbsd | generic/openbsd75 | rc.d | no |
| netbsd | generic/netbsd10 | rc.d | no |

## Usage

### From Makefile (Recommended)

```bash
# Run default E2E test (Debian)
make test-e2e

# Run specific platform
make test-e2e-debian
make test-e2e-alpine
make test-e2e-freebsd

# Clean up all VMs
make test-e2e-clean
```

### Direct Vagrant

```bash
cd e2e

# Start and provision Debian
vagrant up debian

# SSH into VM
vagrant ssh debian

# Run tests manually
vagrant ssh debian -c "sudo /vagrant/test-install.sh"

# Destroy VM
vagrant destroy debian -f
```

## Test Script

`test-install.sh` validates:

1. Install script completes successfully
2. Binary exists at `/usr/local/bin/supervizio`
3. Config directory exists
4. Config file `config.yaml` exists
5. Service file installed for platform
6. `--version` command works
7. Uninstall removes binary

## CI Integration

For GitHub Actions on macOS runners:

```yaml
jobs:
  e2e:
    runs-on: macos-15  # ARM64
    steps:
      - uses: actions/checkout@v4
      - name: Setup Vagrant
        run: brew install vagrant qemu
      - name: Run E2E
        run: make test-e2e
```
