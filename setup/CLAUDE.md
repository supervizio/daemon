# Setup - Installation Scripts

Platform-agnostic installation for supervizio.

## Structure

```
setup/
├── install.sh              # Universal install script
├── uninstall.sh            # Universal uninstall script
├── scripts/                # Package manager lifecycle scripts
│   └── (see scripts/CLAUDE.md)
└── init/                   # Init system service files
    ├── systemd/supervizio.service
    ├── openrc/supervizio
    ├── runit/supervizio/{run,log/run}
    ├── freebsd/supervizio
    ├── openbsd/supervizio
    ├── netbsd/supervizio
    └── launchd/io.superviz.daemon.plist
```

## Supported Platforms

| OS | Init System | Config Dir |
|----|-------------|------------|
| Linux (systemd) | systemd | `/etc/supervizio` |
| Linux (Alpine) | OpenRC | `/etc/supervizio` |
| Linux (Void) | runit | `/etc/supervizio` |
| Linux (legacy) | SysVinit | `/etc/supervizio` |
| FreeBSD | rc.d | `/usr/local/etc/supervizio` |
| OpenBSD | rc.d | `/etc/supervizio` |
| NetBSD | rc.d | `/etc/supervizio` |
| DragonFlyBSD | rc.d | `/usr/local/etc/supervizio` |
| macOS | launchd | `/etc/supervizio` |

## Usage

### Install

```bash
# Download latest and install
curl -sSL https://raw.githubusercontent.com/supervizio/daemon/main/setup/install.sh | sudo sh

# Or with specific version
SUPERVIZIO_VERSION=v1.0.0 sudo ./setup/install.sh
```

### Uninstall

```bash
sudo ./setup/uninstall.sh
```

## Detection Logic

### OS Detection
Uses `uname -s` to detect: Linux, Darwin, FreeBSD, OpenBSD, NetBSD, DragonFly

### Architecture Detection
Uses `uname -m` to detect: amd64, arm64, arm, 386, riscv64

### Init System Detection (Linux)
1. `systemctl` + `/run/systemd/system` → systemd
2. `rc-service` command → OpenRC
3. `sv` command + `/etc/sv` exists → runit
4. `/etc/init.d` exists → SysVinit

## Service Management

| Platform | Start | Stop | Status |
|----------|-------|------|--------|
| systemd | `systemctl start supervizio` | `systemctl stop supervizio` | `systemctl status supervizio` |
| OpenRC | `rc-service supervizio start` | `rc-service supervizio stop` | `rc-service supervizio status` |
| runit | `sv start supervizio` | `sv stop supervizio` | `sv status supervizio` |
| FreeBSD | `service supervizio start` | `service supervizio stop` | `service supervizio status` |
| OpenBSD | `rcctl start supervizio` | `rcctl stop supervizio` | `rcctl check supervizio` |
| macOS | `launchctl load ...` | `launchctl unload ...` | `launchctl list | grep superviz` |
