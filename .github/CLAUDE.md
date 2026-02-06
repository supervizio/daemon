# GitHub - CI/CD Workflows

GitHub Actions and instructions configuration.

## Structure

```
.github/
├── workflows/
│   ├── ci-x86.yml       # CI: x86/amd64 (lint, test, build, e2e, packages)
│   ├── ci-arm64.yml     # CI: ARM64 (native ARM runner + Zig cross)
│   ├── release.yml      # Semantic release + packages + e2e tests
│   └── deploy-repo.yml  # Package repository deployment (GitHub Pages)
└── instructions/
    └── codacy.instructions.md
```

## Workflows

| Workflow | Trigger | Action |
|----------|---------|--------|
| `ci-x86.yml` | Push/PR | x86 builds: Rust lint/test, probe, Go test, binaries, e2e |
| `ci-arm64.yml` | Push/PR | ARM64 builds: native ARM runner + Zig cross-compile |
| `release.yml` | Tag/Manual | Packages + e2e tests + GitHub release |
| `deploy-repo.yml` | Release | Deploy apt/yum/apk repos to GitHub Pages |

## Versioning

Automatic detection from conventional commits:
- `feat:` → minor, `fix:` → patch, `BREAKING CHANGE:` → major

## Platforms

- Linux: amd64, arm64 (glibc + musl)
- BSD: FreeBSD, OpenBSD, NetBSD (amd64, arm64 via Zig cross)
- macOS: amd64, arm64

## Related Directories

| Directory | See |
|-----------|-----|
| `workflows/` | `workflows/CLAUDE.md` for detailed architecture |
| `../src/` | Source code being built |
