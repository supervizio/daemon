# E2E Testing - Platform Matrix

End-to-end testing across all supported platforms and init systems (AMD64 only).

## Coverage (9 jobs)

| Init System | Distribution | Pkg | VM | Docker |
|-------------|--------------|-----|:--:|:------:|
| systemd | Debian 13 | .deb | ✅ | ✅ |
| systemd | Fedora 38 | .rpm | ✅ | ✅ |
| OpenRC | Alpine 3.21 | .apk | ✅ | ✅ |
| SysVinit | Devuan 6 | .deb | ✅ | ✅ |
| runit | Alpine 3.21 | .apk | ✅ | ✅ |
| BSD rc.d | FreeBSD 14 | pkg | ✅ | - |
| BSD rc.d | OpenBSD 7 | pkg | ✅ | - |
| BSD rc.d | NetBSD 9 | pkgin | ✅ | - |
| BSD rc.d | DragonFlyBSD 6 | pkg | ✅ | - |

## Structure

```
e2e/
├── Vagrantfile           # VM config (libvirt)
├── test-install.sh       # Universal test script
├── Dockerfile.debian     # systemd, .deb
├── Dockerfile.fedora     # systemd, .rpm
├── Dockerfile.alpine     # OpenRC, .apk
├── Dockerfile.devuan     # SysVinit, .deb
└── Dockerfile.alpine-runit # runit, .apk
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
6. Uninstall removes binary

## Local Testing

```bash
# Vagrant VM
cd e2e && vagrant up debian13 --provider=libvirt
vagrant ssh debian13 -c "sudo /vagrant/test-install.sh"

# Docker
cd src && CGO_ENABLED=0 go build -o ../bin/supervizio ./cmd/daemon
docker build -f e2e/Dockerfile.debian -t supervizio-debian .
docker run --rm supervizio-debian
```

## Notes

- BSD: VM tests only (no Docker support)
- Alpine-runit: Same box as OpenRC, provisioned with runit
- All binaries: `CGO_ENABLED=0` (static)
