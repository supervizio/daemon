# E2E Testing - Platform Matrix

End-to-end testing across all supported platforms and init systems (AMD64 only).

## Coverage (10 jobs)

| Init System | Distribution | Pkg | VM | Docker | Probe Coverage |
|-------------|--------------|-----|:--:|:------:|:--------------:|
| systemd | Debian 13 | .deb | ✅ | ✅ | 100% |
| systemd | Fedora 38 | .rpm | ✅ | ✅ | 100% |
| OpenRC | Alpine 3.21 | .apk | ✅ | ✅ | 100% |
| SysVinit | Devuan 6 | .deb | ✅ | ✅ | 100% |
| runit | Alpine 3.21 | .apk | ✅ | ✅ | 100% |
| BSD rc.d | FreeBSD 14 | pkg | ✅ | - | 98% |
| BSD rc.d | OpenBSD 7 | pkg | ✅ | - | 92% |
| BSD rc.d | NetBSD 9 | pkgin | ✅ | - | 90% |
| BSD rc.d | DragonFlyBSD 6 | pkg | ✅ | - | 0% (stub) |
| launchd | macOS | - | - | - | 95% |

**Note**: macOS testing via GitHub Actions macOS runners only (no Vagrant support).

## Structure

```
e2e/
├── Vagrantfile           # VM config (libvirt/qemu)
├── test-install.sh       # Universal install/uninstall test
├── test-macos.sh         # macOS-specific E2E test
├── validate-probe.sh     # Comprehensive probe JSON validation
├── Dockerfile.debian     # systemd, .deb
├── Dockerfile.fedora     # systemd, .rpm
├── Dockerfile.alpine     # OpenRC, .apk
├── Dockerfile.devuan     # SysVinit, .deb
├── Dockerfile.alpine-runit # runit, .apk
└── behavioral/           # Behavioral tests (testcontainers-go)
    ├── crasher/          # Test binary
    ├── testdata/         # YAML configs
    └── *_test.go         # Go tests
```

## Init System Paths

| Init | Service Path | Enable Command |
|------|--------------|----------------|
| systemd | `/etc/systemd/system/` | `systemctl enable` |
| OpenRC | `/etc/init.d/` | `rc-update add` |
| SysVinit | `/etc/init.d/` | `update-rc.d` |
| runit | `/etc/sv/` | `ln -s /etc/sv/X /var/service/` |
| BSD rc.d | `/usr/local/etc/rc.d/` | `sysrc enable` |

## Test Script (test-install.sh)

1. Install script completes
2. Binary at `/usr/local/bin/supervizio`
3. Config directory and file exist
4. Service file installed
5. `--version` works
6. **Probe metrics validation** (via `validate-probe.sh`)
7. Uninstall removes binary

## Probe Validation (validate-probe.sh)

Comprehensive per-platform validation of `--probe` JSON output:

### Coverage by Platform

| Platform | Expected | Validates |
|----------|----------|-----------|
| Linux | 100% | CPU, Memory, PSI, Load, Process, Disk, Network, I/O, Context Switches, Thermal, Quota |
| FreeBSD | 98% | All except: PSI, iowait, steal, buffers |
| macOS | 95% | All except: PSI, iowait, steal, buffers |
| OpenBSD | 92% | All except: PSI, iowait, steal, buffers, temp_max/crit |
| NetBSD | 90% | All except: PSI, iowait, steal, buffers, temp_max/crit |
| DragonFlyBSD | 0% | Stub only (no metrics) |

### Sections Validated

1. **Metadata**: timestamp, platform, collected_at_ns
2. **CPU**: user/system/idle percent, cores, frequency_mhz
3. **Memory**: total/available/used/cached bytes, swap, used_percent
4. **PSI**: cpu/memory/io pressure (Linux 4.20+ only)
5. **Load**: load_1min/5min/15min
6. **Process**: pid, memory_rss/vms, state
7. **Disk**: partitions, usage (total/used/available)
8. **Disk I/O**: reads/writes completed, device names
9. **Network Interfaces**: name, rx/tx bytes/packets
10. **Network Connections**: tcp_stats, tcp/udp arrays
11. **Aggregated I/O**: read/write bytes/ops
12. **Context Switches**: system_total, voluntary, involuntary
13. **Thermal**: supported flag, zones with temp_celsius
14. **Container/Runtime**: is_containerized detection
15. **Quota**: supported flag (Linux cgroups)

### Usage

```bash
# Inside VM after installation
/vagrant/validate-probe.sh

# Returns exit code 0 if all expected fields present
# Returns exit code 1 if missing fields or invalid JSON
```

## Local Testing

```bash
# Vagrant VM (Linux/BSD)
cd e2e && vagrant up debian13 --provider=libvirt
vagrant ssh debian13 -c "sudo /vagrant/test-install.sh"

# Docker (Linux distributions)
cd src && CGO_ENABLED=0 go build -o ../bin/supervizio ./cmd/daemon
docker build -f e2e/Dockerfile.debian -t supervizio-debian .
docker run --rm supervizio-debian

# macOS (local machine)
cd e2e && ./test-macos.sh

# BSD VMs (FreeBSD, OpenBSD, NetBSD)
cd e2e && vagrant up freebsd --provider=libvirt
vagrant ssh freebsd -c "cd /vagrant/src && cargo build --release && sudo cp target/release/libprobe.a /usr/local/lib/"
vagrant ssh freebsd -c "cd /vagrant/src && go build -o /tmp/supervizio ./cmd/daemon && sudo cp /tmp/supervizio /usr/local/bin/"
vagrant ssh freebsd -c "sudo /vagrant/validate-probe.sh"
```

## Notes

- BSD: VM tests only (no Docker support)
- Alpine-runit: Same box as OpenRC, provisioned with runit
- All binaries: `CGO_ENABLED=0` (static)

## Behavioral Tests

See `behavioral/CLAUDE.md` for runtime behavior tests (restart policies, health probes, PID1 capabilities).
