# GitHub - CI/CD Workflows

GitHub Actions, scripts, and instructions configuration.

## Structure

```
.github/
├── workflows/
│   ├── ci.yml             # CI: unified pipeline (32 jobs, VM E2E, ARM64, cross-build, musl)
│   ├── release.yml        # Semantic release + packages + e2e tests (mirrors CI)
│   └── deploy-repo.yml    # Package repository deployment (GitHub Pages)
├── scripts/
│   ├── vm-acquire.sh      # VM lock acquisition (flambeau mechanism)
│   └── vm-release.sh      # VM release (stop)
└── instructions/
    └── codacy.instructions.md
```

## Workflows

| Workflow | Trigger | Action |
|----------|---------|--------|
| `ci.yml` | Push/PR | Unified pipeline: lint, test, build, packages, Docker E2E, VM E2E |
| `release.yml` | Tag/Manual | Build with probe + packages + e2e tests + GitHub release |
| `deploy-repo.yml` | Release | Deploy apt/yum/apk repos to GitHub Pages |

## Versioning

Automatic detection from conventional commits:
- `feat:` → minor, `fix:` → patch, `BREAKING CHANGE:` → major

## Platforms

- Linux: amd64 + arm64 (glibc + musl), arm + 386 + riscv64 (glibc + musl, cross-compiled)
- BSD: FreeBSD, OpenBSD, NetBSD (native VM builds)
- macOS: amd64 + arm64 (Apple Silicon)

## VM Locking

Scripts in `scripts/` implement a "flambeau" mechanism:
- `vm-acquire.sh`: Lock via `/usebyjob` file on VM, retry if busy (max 15min)
- `vm-release.sh`: Stop VM (lock cleared on next snapshot rollback)

## Related Directories

| Directory | See |
|-----------|-----|
| `workflows/` | `workflows/CLAUDE.md` for detailed architecture |
| `../src/` | Source code being built |
