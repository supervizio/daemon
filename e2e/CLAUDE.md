<!-- updated: 2026-02-15T21:30:00Z -->
# E2E Testing - Platform Matrix

End-to-end testing across all supported platforms and init systems.

## Coverage

| Init System | Distribution | Pkg | VM | Docker | Probe |
|-------------|--------------|-----|:--:|:------:|:-----:|
| systemd | Debian 13 | .deb | ✅ | ✅ | 100% |
| systemd | Fedora 38 | .rpm | ✅ | ✅ | 100% |
| systemd | Arch | pacman | ✅ | - | 100% |
| systemd | openSUSE | .rpm | ✅ | - | 100% |
| OpenRC | Alpine 3.21 | .apk | ✅ | ✅ | 100% |
| s6 | Alpine 3.21 | .apk | ✅ | - | 100% |
| dinit | Artix | pacman | ✅ | - | 100% |
| SysVinit | Devuan 6 | .deb | ✅ | ✅ | 100% |
| runit | Void | xbps | ✅ | ✅ | 100% |
| BSD rc.d | FreeBSD 14 | pkg | ✅ | - | 99% |
| BSD rc.d | OpenBSD 7 | pkg | ✅ | - | 93% |
| BSD rc.d | NetBSD 9 | pkgin | ✅ | - | 91% |
| launchd | macOS | - | - | - | 95% |

## Structure

```
e2e/
├── Vagrantfile              # VM config (libvirt/qemu)
├── config-scratch.yaml      # Scratch container config
├── test-install.sh          # Universal install/uninstall test
├── test-container.sh        # Container-based test runner
├── test-macos.sh            # macOS-specific E2E test
├── validate-probe.sh        # Probe JSON validation (15 sections)
├── Dockerfile.debian        # systemd, .deb
├── Dockerfile.debian-script # Debian script-based install
├── Dockerfile.fedora        # systemd, .rpm
├── Dockerfile.alpine        # OpenRC, .apk
├── Dockerfile.alpine-runit  # runit, .apk
├── Dockerfile.alpine-script # Alpine script-based install
├── Dockerfile.devuan        # SysVinit, .deb
├── Dockerfile.scratch       # Scratch container (PID1 mode)
└── behavioral/              # Go behavioral tests (testcontainers-go)
```

## Test Scripts

| Script | Purpose |
|--------|---------|
| `test-install.sh` | Install, verify binary/config/service, validate probe, uninstall |
| `test-container.sh` | Container-based testing (Docker) |
| `test-macos.sh` | macOS-specific E2E test |
| `validate-probe.sh` | Validates 15 probe JSON sections per platform |

## Probe Validation Coverage

| Platform | Coverage | Exclusions |
|----------|----------|------------|
| Linux | 100% | None |
| FreeBSD | 99% | PSI, iowait, steal |
| macOS | 95% | PSI, iowait, steal, buffers |
| OpenBSD | 93% | PSI, iowait, steal, temp_max/crit |
| NetBSD | 91% | PSI, iowait, steal, temp_max/crit |

## Local Testing

```bash
vagrant up debian13 --provider=libvirt                        # VM
docker build -f e2e/Dockerfile.debian -t supervizio-debian .  # Docker
./e2e/test-macos.sh                                           # macOS
```

## Notes

- BSD: VM tests only (no Docker support)
- macOS: GitHub Actions runners only (no Vagrant)
- See `behavioral/CLAUDE.md` for runtime behavior tests
