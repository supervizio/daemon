# superviz.io - Process Supervisor

PID1-capable process supervisor in Go for containers and Unix systems.

## Project Structure

```
/workspace
├── src/                      # Go source code (mandatory)
│   ├── cmd/daemon/           # CLI entry point
│   └── internal/             # Internal packages
│       ├── config/           # YAML parsing and validation
│       ├── supervisor/       # Service orchestration
│       ├── process/          # Process lifecycle management
│       ├── health/           # Health checks (HTTP/TCP/cmd)
│       ├── kernel/           # OS abstraction (hexagonal)
│       └── logging/          # Log rotation and capture
├── examples/                 # Example configurations
├── .github/workflows/        # CI/CD (lint, test, release)
└── .devcontainer/            # Development environment
```

## Tech Stack

- **Language**: Go 1.25
- **Dependencies**: gopkg.in/yaml.v3, testify
- **Architecture**: Hexagonal (ports & adapters) for OS abstraction

## Development Rules

**STRICT**: Follow `.devcontainer/features/languages/go/RULES.md`

- Go tests alongside code (`*_test.go`)
- Linting with `golangci-lint`
- Race detection required (`go test -race`)

## Workflow

```
/build --context    # Generate contextual docs
/feature "desc"     # Branch feat/, planning, PR
/fix "desc"         # Branch fix/, planning, PR
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
| Examples | `examples/CLAUDE.md` |
| CI/CD | `.github/CLAUDE.md` |
| DevContainer | `.devcontainer/CLAUDE.md` |

## MCP-First

Always use MCP tools before CLI:
- `mcp__github__*` before `gh`
- `mcp__codacy__*` before `codacy-cli`
