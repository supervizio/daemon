<!-- updated: 2026-02-15T21:30:00Z -->
# Packaging - BSD Native Packages

Build scripts for native BSD package formats.

## Structure

```
packaging/
├── freebsd/build-pkg.sh   # FreeBSD .pkg via pkg create
├── openbsd/build-pkg.sh   # OpenBSD .tgz via pkg_create
└── netbsd/build-pkg.sh    # NetBSD .tgz via pkg_create
```

## Package Formats

| OS | Format | Tool | Output |
|----|--------|------|--------|
| FreeBSD | .pkg | `pkg create` | supervizio-{version}.pkg |
| OpenBSD | .tgz | `pkg_create` | supervizio-{version}.tgz |
| NetBSD | .tgz | `pkg_create` | supervizio-{version}.tgz |

## Usage

```bash
# Build on target BSD system
./setup/packaging/freebsd/build-pkg.sh <version> <binary_path>
```

## CI Integration

Built natively on Proxmox BSD VMs during `build-bsd-native` CI jobs.

## Related

| Directory | See |
|-----------|-----|
| `../` | `setup/CLAUDE.md` for install/uninstall |
| `../init/` | Init system service files |
