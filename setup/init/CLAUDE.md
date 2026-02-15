<!-- updated: 2026-02-12T17:05:00Z -->
# Init Systems - Service Files

Service files for all supported init systems.

## Structure

```
init/
├── systemd/    # systemd unit file (.service)
├── openrc/     # OpenRC init script
├── runit/      # runit service (run + log/run)
├── s6/         # s6 service (run + type + log/run)
├── dinit/      # dinit service file
├── sysvinit/   # SysVinit init.d script
├── freebsd/    # FreeBSD rc.d script
├── openbsd/    # OpenBSD rc.d script
├── netbsd/     # NetBSD rc.d script
└── launchd/    # macOS LaunchDaemon plist
```

## Platform Matrix

| Init System | OS | Service File |
|-------------|----|--------------|
| systemd | Debian, Fedora, Arch, openSUSE | `supervizio.service` |
| OpenRC | Alpine | `supervizio` (shell script) |
| runit | Void | `run` + `log/run` |
| s6 | Alpine (s6) | `run` + `type` + `log/run` |
| dinit | Artix, Chimera | `supervizio` (dinit config) |
| SysVinit | Devuan | `supervizio` (shell script) |
| rc.d | FreeBSD | `supervizio` (rc.d script) |
| rc.d | OpenBSD | `supervizio` (rc.d script) |
| rc.d | NetBSD | `supervizio` (rc.d script) |
| launchd | macOS | `io.superviz.daemon.plist` |

## Installation

Service files are installed automatically by `setup/install.sh` based on detected init system.

## Related

| Directory | See |
|-----------|-----|
| `../` | `setup/CLAUDE.md` for install/uninstall |
| `../../e2e/` | E2E tests for all platforms |
