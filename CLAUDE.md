<!-- updated: 2026-02-15T21:30:00Z -->
# superviz.io - Process Supervisor

PID1-capable process supervisor in Go for containers and Unix systems.

## Project Structure

```
/workspace
├── src/                          # Go source code (module: github.com/kodflow/daemon)
│   ├── cmd/daemon/               # CLI entry point
│   └── internal/                 # Internal packages (hexagonal architecture)
│       ├── bootstrap/            # Wire DI, app lifecycle, signals
│       ├── application/          # Use cases: supervisor, lifecycle, health, metrics
│       ├── domain/               # Entities, value objects, port interfaces
│       └── infrastructure/       # Adapters: process, persistence, observability, resources, transport
├── api/proto/                    # Canonical protobuf definitions (buf.yaml)
├── docs/                         # MkDocs Material documentation source
├── e2e/                          # E2E tests (Vagrant VMs + Docker containers)
├── examples/                     # Configuration examples (YAML)
├── setup/                        # Installation scripts (Linux, BSD, macOS)
├── website/                      # Static landing page (GitHub Pages)
├── .github/workflows/            # CI/CD (lint, test, release)
└── .devcontainer/                # Development environment
```

## Tech Stack

- **Language**: Go 1.25.6
- **Dependencies**: gopkg.in/yaml.v3, testify, bubbletea, wire, bbolt, grpc, protobuf
- **Architecture**: Hexagonal (ports & adapters)
- **Linting**: golangci-lint, ktn-linter

## Development Rules

**STRICT**: Follow `.devcontainer/features/languages/go/RULES.md`

- Go tests alongside code (`*_test.go`)
- External tests (`_external_test.go`) for black-box testing
- Internal tests (`_internal_test.go`) for white-box testing
- Race detection required (`go test -race`)
- Zero lint issues (no exclusions)

## Commands

```bash
go build ./cmd/daemon         # Build
go test -race ./...           # Tests with race detection
golangci-lint run -c ../.golangci.yml   # Standard linting
ktn-linter lint -c ../.ktn-linter.yaml ./...  # KTN convention linting
```

## Conventions

| Type | Branch | Commit |
|------|--------|--------|
| Feature | `feat/<desc>` | `feat(scope): message` |
| Bugfix | `fix/<desc>` | `fix(scope): message` |

## Related Directories

| Directory | See |
|-----------|-----|
| Source code | `src/CLAUDE.md` |
| Proto definitions | `api/proto/` (canonical) |
| Documentation | `docs/` (MkDocs Material) |
| E2E tests | `e2e/CLAUDE.md` |
| Installation | `setup/CLAUDE.md` |
| CI/CD | `.github/CLAUDE.md` |
| DevContainer | `.devcontainer/CLAUDE.md` |

## MCP-First

Always use MCP tools before CLI:

- `mcp__github__*` before `gh`
- `mcp__codacy__*` before `codacy-cli`
