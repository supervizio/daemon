# GitHub Workflows

## Purpose

CI/CD automation for build, test, release, and deployment.

## Structure

```text
workflows/
├── ci-x86.yml       # CI for x86/amd64 platforms (Linux, BSD, macOS Intel)
├── ci-arm64.yml     # CI for ARM64 platforms (Linux, BSD, macOS Apple Silicon)
├── ci.yml           # Legacy CI (to be removed after migration)
├── release.yml      # Semantic release + packages + e2e tests
└── deploy-repo.yml  # Package repository deployment (GitHub Pages)
```

## Key Files

| Workflow | Trigger | Action |
|----------|---------|--------|
| `ci-x86.yml` | Push/PR | x86 builds: lint, test, cross-compile BSD, build binaries |
| `ci-arm64.yml` | Push/PR | ARM64 builds: native ARM runner + Zig cross-compile |
| `release.yml` | Tag/Manual | Packages + e2e tests + GitHub release |
| `deploy-repo.yml` | Release | Deploy apt/yum/apk repos |

## CI Architecture

### Parallel Workflows

```
ci-x86.yml                          ci-arm64.yml
───────────                         ────────────
rust-lint                           rust-lint
    │                                   │
rust-test                           rust-test
    │                                   │
build-probe-linux                   build-probe-linux (ARM runner)
build-probe-bsd (Zig cross)         build-probe-bsd (Zig cross)
build-probe-darwin (macos-13)       build-probe-darwin (macos-14)
    │                                   │
go-test                             (skipped, uses x86 results)
    │                                   │
build-binaries (all x86)            build-binaries (all arm64)
    │                                   │
e2e-tests                           e2e-tests
    │                                   │
packages                            packages
```

### Cross-Compilation Strategy

| Platform | Probe Build | Binary Build |
|----------|-------------|--------------|
| Linux glibc | Native | Native |
| Linux musl | Native (musl-gcc) | Alpine container |
| FreeBSD | Zig cross-compile | Go cross + Zig CC |
| OpenBSD | Zig cross-compile | Go cross + Zig CC |
| NetBSD | Zig cross-compile | Go cross + Zig CC |
| DragonFlyBSD | Zig cross-compile | Go cross + Zig CC |
| macOS | Native (GitHub macOS runner) | Native |

### Key Tools

- **Zig**: Cross-compiler for BSD targets (bundles BSD libc)
- **musl-gcc**: Static linking for Alpine/scratch containers
- **Wire**: Go dependency injection code generation
- **nfpm**: Linux package generation (.deb, .rpm, .apk)

## Conventions

- Use reusable workflows where possible
- Pin action versions with SHA
- Secrets via GitHub Secrets (never hardcoded)
- Zig version: 0.13.0
- Rust: stable
- Go: 1.25.6

## Migration Notes

The `ci.yml` file is legacy and will be removed after validating the new split workflows.
The new architecture eliminates VM builds for binary compilation - VMs are only used for E2E testing.
