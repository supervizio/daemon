# superviz.io - Process Supervisor

PID1-capable process supervisor in Go for containers and Unix systems.

## Project Structure

```
/workspace
├── src/                          # Go source code
│   ├── cmd/daemon/               # CLI entry point
│   └── internal/                 # Internal packages (hexagonal)
│       ├── application/          # Application layer
│       │   ├── config/           # Config port interface
│       │   ├── health/           # Health monitoring
│       │   ├── process/          # Process management
│       │   └── supervisor/       # Service orchestration
│       ├── domain/               # Domain layer
│       │   ├── health/           # Health entities
│       │   ├── process/          # Process entities
│       │   ├── service/          # Service configuration
│       │   └── shared/           # Shared value objects
│       ├── infrastructure/       # Infrastructure layer
│       │   ├── config/yaml/      # YAML config loader
│       │   ├── health/           # Health check adapters
│       │   ├── kernel/           # OS abstraction (signals, reaper)
│       │   │   ├── adapters/     # Platform adapters
│       │   │   └── ports/        # Kernel interfaces
│       │   ├── logging/          # Log management (writers, capture)
│       │   └── process/          # Process executor
├── website/                      # Documentation website
├── .github/workflows/            # CI/CD (lint, test, release)
└── .devcontainer/                # Development environment
```

## Tech Stack

- **Language**: Go 1.25.5
- **Dependencies**: gopkg.in/yaml.v3, testify
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
golangci-lint run             # Standard linting
ktn-linter lint ./...         # KTN convention linting
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
| CI/CD | `.github/CLAUDE.md` |
| DevContainer | `.devcontainer/CLAUDE.md` |

## MCP-First

Always use MCP tools before CLI:

- `mcp__github__*` before `gh`
- `mcp__codacy__*` before `codacy-cli`
