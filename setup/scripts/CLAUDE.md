# Setup Scripts

## Purpose

Package manager lifecycle scripts for .deb, .rpm, and .apk packages.

## Structure

```text
scripts/
├── postinstall-deb.sh   # Debian/Ubuntu post-install
├── postinstall-rpm.sh   # RHEL/Fedora post-install
├── postinstall-apk.sh   # Alpine post-install
├── preremove-deb.sh     # Debian/Ubuntu pre-remove
├── preremove-rpm.sh     # RHEL/Fedora pre-remove
└── preremove-apk.sh     # Alpine pre-remove
```

## Key Files

| Script | Trigger | Action |
|--------|---------|--------|
| `postinstall-*.sh` | After package install | Setup systemd service, create user |
| `preremove-*.sh` | Before package remove | Stop service, cleanup |

## Conventions

- POSIX-compliant shell (#!/bin/sh)
- Idempotent operations (safe to run multiple times)
- Exit 0 on success, non-zero on failure
