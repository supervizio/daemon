# E2E Testing - Platform Matrix

End-to-end testing for supervizio across all supported platforms and init systems.
AMD64 only (setup scripts are architecture-agnostic).

## Init System Coverage

| Init System | Distribution | VM | Docker |
|-------------|--------------|:--:|:------:|
| **systemd** | Debian 12 | ✅ Vagrant | ✅ |
| **systemd** | Ubuntu 22.04 | ✅ Vagrant | ✅ |
| **OpenRC** | Alpine 3.19 | ✅ Vagrant | ✅ |
| **SysVinit** | Devuan 4 | ✅ Vagrant | ✅ |
| **runit** | Void Linux | - | ✅ |
| **BSD rc.d** | FreeBSD 14 | ✅ Vagrant | - |
| **BSD rc.d** | OpenBSD 7 | ✅ Vagrant | - |
| **BSD rc.d** | NetBSD 10 | ✅ Vagrant | - |
| **BSD rc.d** | DragonFlyBSD 6 | ✅ Vagrant | - |

**Total: 9 jobs**

## Structure

```
e2e/
├── Vagrantfile           # VM configuration (libvirt)
├── test-install.sh       # Universal installation test
├── Dockerfile.debian     # Debian (systemd)
├── Dockerfile.ubuntu     # Ubuntu (systemd)
├── Dockerfile.alpine     # Alpine (OpenRC)
├── Dockerfile.devuan     # Devuan (SysVinit)
├── Dockerfile.void       # Void Linux (runit)
└── CLAUDE.md             # This file
```

## CI Workflow

```
E2E Tests (9 jobs, 2 workflow jobs)
│
├── e2e-linux (5 matrix jobs - Vagrant + Docker)
│   ├── Debian (systemd)
│   ├── Ubuntu (systemd)
│   ├── Alpine (OpenRC)
│   ├── Devuan (SysVinit)
│   └── Void (runit) - Docker only
│
└── e2e-bsd (4 matrix jobs - Vagrant)
    ├── FreeBSD (rc.d)
    ├── OpenBSD (rc.d)
    ├── NetBSD (rc.d)
    └── DragonFlyBSD (rc.d)
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

| Test | Base Image | Init System |
|------|------------|-------------|
| Debian | `debian:trixie-slim` | systemd |
| Ubuntu | `ubuntu:24.04` | systemd |
| Alpine | `alpine:3.20` | OpenRC |
| Devuan | `devuan/devuan:daedalus` | SysVinit |
| Void | `ghcr.io/void-linux/void-glibc:latest` | runit |

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

- **Void Linux**: No Vagrant box - Docker only
- **DragonFlyBSD**: Not supported by cross-platform-actions - uses Vagrant
- **BSD**: No Docker support - VM tests only
- All binaries: `CGO_ENABLED=0` (static)
