# Setup - Installation Scripts

Platform-agnostic installation for supervizio.

## Structure

```
setup/
├── install.sh              # Universal install script (pkg-aware + fallback)
├── uninstall.sh            # Universal uninstall script (pkg-aware + non-interactive)
├── scripts/                # Package manager lifecycle scripts
│   └── (see scripts/CLAUDE.md)
├── packaging/              # BSD native package build scripts
│   ├── freebsd/build-pkg.sh   # FreeBSD .pkg via pkg create
│   ├── openbsd/build-pkg.sh   # OpenBSD .tgz via pkg_create
│   └── netbsd/build-pkg.sh    # NetBSD .tgz via pkg_create
└── init/                   # Init system service files
    ├── systemd/supervizio.service
    ├── openrc/supervizio
    ├── runit/supervizio/{run,log/run}
    ├── s6/supervizio/{run,type,log/run}
    ├── dinit/supervizio
    ├── sysvinit/supervizio
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
| Linux (Alpine s6) | s6 | `/etc/supervizio` |
| Linux (Artix/Chimera) | dinit | `/etc/supervizio` |
| Linux (Devuan) | SysVinit | `/etc/supervizio` |
| FreeBSD | rc.d | `/usr/local/etc/supervizio` |
| OpenBSD | rc.d | `/etc/supervizio` |
| NetBSD | rc.d | `/etc/supervizio` |
| macOS | launchd | `/etc/supervizio` |

## Usage

### Install

```bash
# Auto-detect package manager, download native package, fallback to raw binary
curl -sSL https://raw.githubusercontent.com/supervizio/daemon/main/setup/install.sh | sudo sh

# With specific version
SUPERVIZIO_VERSION=v1.0.0 sudo ./setup/install.sh

# With local package (E2E testing)
SUPERVIZIO_LOCAL_PKG=/tmp/supervizio.deb sudo ./setup/install.sh

# With local binary (fallback testing)
SUPERVIZIO_LOCAL_BIN=/tmp/supervizio sudo ./setup/install.sh
```

### Uninstall

```bash
sudo ./setup/uninstall.sh

# Non-interactive (CI - auto-removes config/logs)
SUPERVIZIO_NON_INTERACTIVE=true sudo ./setup/uninstall.sh
```

## Detection Logic

### OS Detection
Uses `uname -s` to detect: Linux, Darwin, FreeBSD, OpenBSD, NetBSD

### Architecture Detection
Uses `uname -m` to detect: amd64, arm64, arm, 386, riscv64

### Package Manager Detection
| Manager | Detection | Package |
|---------|-----------|---------|
| apt/dpkg | `dpkg` command | .deb |
| dnf | `dnf` command | .rpm |
| zypper | `zypper` command | .rpm |
| apk | `apk` command | .apk |
| pacman | `pacman` command | .pkg.tar.zst |
| xbps | `xbps-install` command | .xbps |
| FreeBSD pkg | `pkg` on FreeBSD | .pkg |
| OpenBSD pkg_add | `pkg_add` on OpenBSD | .tgz |
| NetBSD pkgin | `pkgin` on NetBSD | .tgz |

### Init System Detection (Linux)
1. `systemctl` + `/run/systemd/system` → systemd
2. `rc-service` command → OpenRC
3. `s6-svc` command or `/etc/s6` exists → s6
4. `dinitctl` command → dinit
5. `sv` command + `/etc/sv` exists → runit
6. `/etc/init.d` exists → SysVinit

## Service Management

| Platform | Start | Stop | Status |
|----------|-------|------|--------|
| systemd | `systemctl start supervizio` | `systemctl stop supervizio` | `systemctl status supervizio` |
| OpenRC | `rc-service supervizio start` | `rc-service supervizio stop` | `rc-service supervizio status` |
| runit | `sv start supervizio` | `sv stop supervizio` | `sv status supervizio` |
| s6 | `s6-svc -u /etc/s6/supervizio` | `s6-svc -d /etc/s6/supervizio` | `s6-svstat /etc/s6/supervizio` |
| dinit | `dinitctl start supervizio` | `dinitctl stop supervizio` | `dinitctl status supervizio` |
| SysVinit | `/etc/init.d/supervizio start` | `/etc/init.d/supervizio stop` | `/etc/init.d/supervizio status` |
| FreeBSD | `service supervizio start` | `service supervizio stop` | `service supervizio status` |
| OpenBSD | `rcctl start supervizio` | `rcctl stop supervizio` | `rcctl check supervizio` |
| macOS | `launchctl load ...` | `launchctl unload ...` | `launchctl list | grep superviz` |
