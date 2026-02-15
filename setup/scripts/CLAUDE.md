<!-- updated: 2026-02-15T21:30:00Z -->
# Setup Scripts

## Purpose

Package manager lifecycle scripts for all supported package formats.

## Structure

```text
scripts/
├── postinstall-deb.sh    # Debian/Ubuntu/Devuan post-install
├── postinstall-rpm.sh    # RHEL/Fedora/openSUSE post-install
├── postinstall-apk.sh    # Alpine post-install
├── postinstall-arch.sh   # Arch/Artix post-install (systemd + dinit)
├── postinstall-xbps.sh   # Void Linux post-install (runit)
├── preremove-deb.sh      # Debian/Ubuntu/Devuan pre-remove
├── preremove-rpm.sh      # RHEL/Fedora/openSUSE pre-remove
├── preremove-apk.sh      # Alpine pre-remove
├── preremove-arch.sh     # Arch/Artix pre-remove (systemd + dinit)
└── preremove-xbps.sh     # Void Linux pre-remove (runit)
```

## Key Files

| Script | Trigger | Action |
|--------|---------|--------|
| `postinstall-*.sh` | After package install | Setup service, create user, detect init system |
| `preremove-*.sh` | Before package remove | Stop service, cleanup |
| `postinstall-arch.sh` | Arch/Artix install | Detects systemd vs dinit, installs correct service |
| `postinstall-xbps.sh` | Void install | Sets up runit service directory + symlink |

## Conventions

- POSIX-compliant shell (#!/bin/sh)
- Idempotent operations (safe to run multiple times)
- Exit 0 on success, non-zero on failure
- Init system auto-detection for multi-init distros (Arch/Artix)
