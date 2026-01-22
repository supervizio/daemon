# E2E Testing - Platform Matrix

End-to-end testing for supervizio across all supported platforms and init systems.
AMD64 only (setup scripts are architecture-agnostic).

## Init System Coverage

| Init System | Distribution | Pkg Format | VM | Docker |
|-------------|--------------|------------|:--:|:------:|
| **systemd** | Debian 13 | .deb | ✅ Vagrant | ✅ |
| **systemd** | Rocky 10 | .rpm | ✅ Vagrant | ✅ |
| **OpenRC** | Alpine 3.21 | .apk | ✅ Vagrant | ✅ |
| **SysVinit** | Devuan 6 | .deb | ✅ Vagrant | ✅ |
| **runit** | Alpine 3.21 | .apk | ✅ Vagrant | ✅ |
| **BSD rc.d** | FreeBSD 14 | pkg | ✅ Vagrant | - |
| **BSD rc.d** | OpenBSD 7 | pkg | ✅ Vagrant | - |
| **BSD rc.d** | NetBSD 10 | pkgin | ✅ Vagrant | - |
| **BSD rc.d** | DragonFlyBSD 6 | pkg | ✅ Vagrant | - |

**Total: 9 jobs**

## Structure

```
e2e/
├── Vagrantfile           # VM configuration (libvirt)
├── test-install.sh       # Universal installation test
├── Dockerfile.debian     # Debian (systemd, .deb)
├── Dockerfile.rocky      # Rocky Linux (systemd, .rpm)
├── Dockerfile.alpine     # Alpine (OpenRC, .apk)
├── Dockerfile.devuan     # Devuan (SysVinit, .deb)
├── Dockerfile.alpine-runit # Alpine (runit, .apk)
└── CLAUDE.md             # This file
```

## CI Workflow

```
E2E Tests (9 jobs, 2 workflow jobs)
│
├── e2e-linux (5 matrix jobs - Vagrant + Docker)
│   ├── Debian (systemd, .deb)
│   ├── Rocky (systemd, .rpm)
│   ├── Alpine (OpenRC, .apk)
│   ├── Devuan (SysVinit, .deb)
│   └── Alpine-runit (runit, .apk)
│
└── e2e-bsd (4 matrix jobs - Vagrant)
    ├── FreeBSD (rc.d, pkg)
    ├── OpenBSD (rc.d, pkg)
    ├── NetBSD (rc.d, pkgin)
    └── DragonFlyBSD (rc.d, pkg)
```

## Init Systems

| Init System | Service Path | Enable Command |
|-------------|--------------|----------------|
| **systemd** | `/etc/systemd/system/` | `systemctl enable` |
| **OpenRC** | `/etc/init.d/` | `rc-update add` |
| **SysVinit** | `/etc/init.d/` | `update-rc.d` |
| **runit** | `/etc/sv/` | `ln -s /etc/sv/X /var/service/` |
| **BSD rc.d** | `/usr/local/etc/rc.d/` | `sysrc enable` |

## VM Tests (test-install.sh)

1. Install script completes successfully
2. Binary exists at `/usr/local/bin/supervizio`
3. Config directory exists
4. Config file `config.yaml` exists
5. Service file installed for platform
6. `--version` command works
7. Uninstall removes binary

## Container Tests

| Test | Base Image | Init System | Pkg Format |
|------|------------|-------------|------------|
| Debian | `debian:trixie-slim` | systemd | .deb |
| Rocky | `rockylinux:10-minimal` | systemd | .rpm |
| Alpine | `alpine:3.21` | OpenRC | .apk |
| Devuan | `dyne/devuan:excalibur` | SysVinit | .deb |
| Alpine-runit | `alpine:3.21` + runit | runit | .apk |

## Local Testing

### Linux VMs (Vagrant)

```bash
cd e2e
vagrant up debian13 --provider=libvirt
vagrant ssh debian13 -c "sudo /vagrant/test-install.sh"
vagrant destroy debian13 -f
```

### Docker Containers

```bash
cd src && CGO_ENABLED=0 go build -o ../bin/supervizio ./cmd/daemon
docker build -f e2e/Dockerfile.debian -t supervizio-debian .
docker run --rm supervizio-debian
```

## Platform Notes

- **Alpine-runit**: Same Alpine box as OpenRC, but provisioned with runit
- **BSD**: No Docker support - VM tests only
- All binaries: `CGO_ENABLED=0` (static)
