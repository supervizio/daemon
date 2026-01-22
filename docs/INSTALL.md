# Installation Guide

supervizio can be installed via package managers on Linux or as a standalone binary on all supported platforms.

## Package Repository (Recommended)

### Debian / Ubuntu / Devuan

```bash
# Add GPG key
curl -fsSL https://supervizio.github.io/daemon/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/supervizio.gpg

# Add repository
echo "deb [signed-by=/etc/apt/keyrings/supervizio.gpg] https://supervizio.github.io/daemon/apt stable main" | sudo tee /etc/apt/sources.list.d/supervizio.list

# Install
sudo apt update
sudo apt install supervizio
```

### Rocky Linux / RHEL / Fedora

```bash
# Add GPG key
sudo rpm --import https://supervizio.github.io/daemon/gpg.key

# Add repository
sudo curl -fsSL https://supervizio.github.io/daemon/rpm/supervizio.repo -o /etc/yum.repos.d/supervizio.repo

# Install
sudo dnf install supervizio
```

### Alpine Linux

```bash
# Add signing key
sudo wget -O /etc/apk/keys/supervizio.rsa.pub https://supervizio.github.io/daemon/apk/supervizio.rsa.pub

# Add repository
echo "https://supervizio.github.io/daemon/apk/v3.21/main" | sudo tee -a /etc/apk/repositories

# Install
sudo apk add supervizio
```

## Direct Download

### Binary Installation (All Platforms)

```bash
# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/armv7l/arm/;s/i686/386/')

# Download latest release
VERSION=$(curl -s https://api.github.com/repos/supervizio/daemon/releases/latest | grep tag_name | cut -d'"' -f4)
curl -fsSL "https://github.com/supervizio/daemon/releases/download/${VERSION}/supervizio-${OS}-${ARCH}" -o supervizio

# Install
chmod +x supervizio
sudo mv supervizio /usr/local/bin/

# Verify
supervizio --version
```

### Package Files

Download `.deb`, `.rpm`, or `.apk` packages directly from [GitHub Releases](https://github.com/supervizio/daemon/releases).

```bash
# Debian/Ubuntu/Devuan
sudo dpkg -i supervizio_*.deb

# Rocky/RHEL/Fedora
sudo rpm -i supervizio-*.rpm

# Alpine (unsigned)
sudo apk add --allow-untrusted supervizio-*.apk
```

## Supported Platforms

| OS | Architectures | Package Format |
|----|---------------|----------------|
| Debian/Ubuntu/Devuan | amd64, arm64 | `.deb` |
| Rocky/RHEL/Fedora | amd64, arm64 | `.rpm` |
| Alpine | amd64, arm64 | `.apk` |
| FreeBSD | amd64, arm64, arm, 386, riscv64 | Binary |
| OpenBSD | amd64, arm64, arm, 386 | Binary |
| NetBSD | amd64, arm64, arm, 386 | Binary |
| DragonFlyBSD | amd64 | Binary |
| macOS | amd64, arm64 | Binary |

## Post-Installation

### Configure

```bash
# Copy example configuration
sudo cp /etc/supervizio/config.example.yaml /etc/supervizio/config.yaml

# Edit configuration
sudo nano /etc/supervizio/config.yaml
```

### Start Service

```bash
# systemd (Debian, Rocky, etc.)
sudo systemctl enable --now supervizio

# OpenRC (Alpine)
sudo rc-update add supervizio
sudo rc-service supervizio start

# SysVinit (Devuan)
sudo update-rc.d supervizio defaults
sudo service supervizio start
```

### Verify

```bash
supervizio --version
sudo systemctl status supervizio  # or equivalent
```

## Uninstall

### Package Manager

```bash
# Debian/Ubuntu/Devuan
sudo apt remove supervizio

# Rocky/RHEL/Fedora
sudo dnf remove supervizio

# Alpine
sudo apk del supervizio
```

### Manual

```bash
sudo rm /usr/local/bin/supervizio
sudo rm -rf /etc/supervizio
sudo rm -rf /var/log/supervizio
```

## GPG Key Verification

The package repository is signed with GPG. The public key fingerprint is published at:
- https://supervizio.github.io/daemon/gpg.key

To verify a downloaded package manually:

```bash
# Import key
curl -fsSL https://supervizio.github.io/daemon/gpg.key | gpg --import

# Verify DEB
dpkg-sig --verify supervizio_*.deb

# Verify RPM
rpm --checksig supervizio-*.rpm
```
