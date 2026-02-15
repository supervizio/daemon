# GitHub Workflows

## Purpose

CI/CD automation for build, test, release, and deployment.

## Structure

```text
workflows/
├── ci.yml           # Unified CI pipeline (32 jobs, ARM64, cross-build, musl, native BSD)
├── release.yml      # Semantic release + packages + e2e tests (mirrors CI builds)
└── deploy-repo.yml  # Package repository deployment (GitHub Pages)
```

## Key Files

| Workflow | Trigger | Action |
|----------|---------|--------|
| `ci.yml` | Push/PR | Full pipeline: lint, test, build, packages, Docker + VM E2E |
| `release.yml` | Tag/Manual | Probe builds + packages + e2e tests + GitHub release |
| `deploy-repo.yml` | Release | Deploy apt/yum/apk repos |

## CI Architecture (ci.yml)

### Pipeline Flow (32 jobs)

```
rust-lint → rust-test
  → [build-probe-linux | build-probe-linux-arm64 | build-probe-linux-cross x6
     | build-probe-darwin | build-probe-darwin-arm64]
  → verify-probes (15 probes) → go-test (probe verified)
  → [build-amd64 | build-amd64-musl | build-arm64 | build-arm64-musl
     | build-linux-cross x3 | build-linux-cross-musl x3
     | build-bsd-native x3 | build-darwin-amd64 | build-darwin-arm64]
  → verify-binaries (15 binaries)
  → [packages-glibc | packages-musl | packages-archlinux | packages-xbps]
  → verify-packages-linux → verify-packages-all
  → [e2e-docker | e2e-docker-arm64 | e2e-macos | e2e-macos-arm64]
  → [e2e-vm (12 VMs, always)] → summary → vm-cleanup (always)
```

### Release Architecture (release.yml)

```
version
  → [build-probe-linux | build-probe-linux-arm64 | build-probe-linux-cross x6
     | build-probe-darwin | build-probe-darwin-arm64]
  → [build-linux-amd64 | build-linux-amd64-musl | build-linux-arm64 | build-linux-arm64-musl
     | build-linux-cross x3 | build-linux-cross-musl x3
     | build-bsd-native x3 | build-darwin-amd64 | build-darwin-arm64]
  → [packages-glibc | packages-musl | packages-archlinux | packages-xbps]
  → [e2e-docker | e2e-docker-arm64 | e2e-macos | e2e-macos-arm64]
  → [e2e-vm (12 VMs)] → vm-cleanup (always)
  → release
  → cleanup-on-failure (if failure)
```

### Build Strategy

| Platform | Probe | Binary | Runner |
|----------|-------|--------|--------|
| Linux amd64 glibc | Native | Native | `ubuntu-24.04` |
| Linux amd64 musl | musl-gcc | musl-gcc | `ubuntu-24.04` |
| Linux arm64 glibc | Native | Native | `ubuntu-24.04-arm` |
| Linux arm64 musl | musl-gcc | musl-gcc | `ubuntu-24.04-arm` |
| Linux arm glibc | Cross (arm-gcc) | Cross (CGO) | `ubuntu-24.04` |
| Linux arm musl | Cross (Zig CC) | Cross (CGO, static) | `ubuntu-24.04` |
| Linux 386 glibc | Cross (i686-gcc) | Cross (CGO) | `ubuntu-24.04` |
| Linux 386 musl | Cross (Zig CC) | Cross (CGO, static) | `ubuntu-24.04` |
| Linux riscv64 glibc | Cross (riscv64-gcc) | Cross (CGO) | `ubuntu-24.04` |
| Linux riscv64 musl | Cross (Zig CC) | Cross (CGO, static) | `ubuntu-24.04` |
| FreeBSD | Native on VM | Native on VM | `[self-hosted, proxmox]` |
| OpenBSD | Native on VM | Native on VM | `[self-hosted, proxmox]` |
| NetBSD | Native on VM | Native on VM | `[self-hosted, proxmox]` |
| macOS amd64 | Native | Native | `macos-15-intel` |
| macOS arm64 | Native | Native | `macos-15` |

### VM Lifecycle

- **Acquire**: `.github/scripts/vm-acquire.sh` (lock + rollback + start)
- **Release**: `.github/scripts/vm-release.sh` (stop)
- **Cleanup**: `vm-cleanup` job resets all 12 VMs (always runs)

### E2E VM Matrix (12 VMs)

| VM | Init | Package | Runner |
|----|------|---------|--------|
| alpine-openrc | OpenRC | .apk | self-hosted |
| alpine-s6 | s6 | .apk | self-hosted |
| arch-systemd | systemd | .pkg.tar.zst | self-hosted |
| artix-dinit | dinit | .pkg.tar.zst | self-hosted |
| debian-systemd | systemd | .deb | self-hosted |
| devuan-sysvinit | SysVinit | .deb | self-hosted |
| fedora-systemd | systemd | .rpm | self-hosted |
| opensuse-systemd | systemd | .rpm | self-hosted |
| void-runit | runit | .xbps | self-hosted |
| freebsd | rc.d | .pkg | self-hosted |
| openbsd | rc.d | .tgz | self-hosted |
| netbsd | rc.d | .tgz | self-hosted |

### Key Tools

- **musl-gcc**: Static linking for Alpine/scratch containers (amd64/arm64)
- **Zig CC**: Musl cross-compilation via Zig's built-in musl libc (arm/386/riscv64)
- **Wire**: Go dependency injection code generation
- **nfpm**: Linux package generation (.deb, .rpm, .apk, .pkg.tar.zst)

## Conventions

- Pin action versions with SHA
- Secrets via GitHub Secrets (never hardcoded)
- CGO_ENABLED=1 always (probe requires Rust FFI)
- Rust: stable, Go: 1.25.6
