# GitHub - CI/CD Workflows

GitHub Actions and instructions configuration.

## Structure

```
.github/
├── workflows/
│   ├── ci.yml           # CI: lint, test, build
│   ├── release.yml      # Release: multi-platform builds
│   └── e2e-macos.yml    # E2E: macOS ARM64/AMD64 tests
└── instructions/
    └── codacy.instructions.md
```

## Workflows

### ci.yml

Triggered on: `push`, `pull_request`

| Job | Actions |
|-----|---------|
| `lint` | golangci-lint |
| `test` | go test -race -cover |
| `build` | go build (verify) |

Configuration:
- Go 1.25.5
- Ubuntu latest
- Coverage reporting to Codacy

### release.yml

Triggered on: `push` to `main`

| Step | Action |
|------|--------|
| Version detect | Conventional commits → semver |
| Build | Multi-platform binaries |
| Release | GitHub release + artifacts |

Platforms:
- Linux: amd64, arm64, 386, armv7
- BSD: amd64, arm64
- macOS: amd64, arm64

### e2e-macos.yml

Triggered on: `workflow_run` (after Release), `workflow_dispatch`

| Runner | Architecture | Tests |
|--------|--------------|-------|
| `macos-14` | Apple Silicon (ARM64) | Binary exec, config, signals |
| `macos-13` | Intel (AMD64) | Binary exec, config, signals |

Manual trigger: Actions → E2E macOS → Run workflow

## Versioning

Automatic detection from commits:
- `feat:` → minor version bump
- `fix:` → patch version bump
- `BREAKING CHANGE:` → major version bump

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../src/` | Source code being built |
| `../` | README.md with CI badges |
