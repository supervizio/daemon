# GitHub - CI/CD Workflows

GitHub Actions and instructions configuration.

## Structure

```
.github/
├── workflows/
│   ├── ci.yml           # CI: lint, test, build
│   └── release.yml      # Release: multi-platform builds
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
