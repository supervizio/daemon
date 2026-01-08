# Development Guide

Setup, testing, and contributing guidelines for superviz.io.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.25+ | Language runtime |
| golangci-lint | latest | Standard linting |
| ktn-linter | latest | KTN conventions |
| Docker | 20+ | DevContainer support |

---

## Development Environment

### Using DevContainer (Recommended)

```bash
# Clone repository
git clone https://github.com/kodflow/daemon.git
cd daemon

# Open in VS Code
code .
# → Click "Reopen in Container"
```

The DevContainer includes:
- Go 1.25 with tools
- golangci-lint
- ktn-linter
- All VS Code extensions

### Manual Setup

```bash
# Install Go 1.25+
# https://go.dev/dl/

# Clone repository
git clone https://github.com/kodflow/daemon.git
cd daemon/src

# Install dependencies
go mod download

# Install linters
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Project Structure

```
/workspace
├── src/                          # Go source code
│   ├── cmd/daemon/               # CLI entry point
│   ├── internal/                 # Internal packages
│   │   ├── application/          # Application layer
│   │   ├── domain/               # Domain layer
│   │   └── infrastructure/       # Infrastructure layer
│   ├── go.mod                    # Go module
│   ├── .golangci.yml             # Linter config
│   └── .ktn-linter.yaml          # KTN config
├── examples/                     # Example configurations
├── docs/                         # Documentation
├── .github/workflows/            # CI/CD
└── .devcontainer/                # Dev environment
```

---

## Common Commands

### Building

```bash
cd src

# Build binary
go build -o supervizio ./cmd/daemon

# Build with version info
go build -ldflags "-X main.version=1.0.0" -o supervizio ./cmd/daemon
```

### Testing

```bash
cd src

# Run all tests
go test ./...

# Run with race detection (REQUIRED)
go test -race ./...

# Run with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test -race ./internal/application/process/...

# Run specific test
go test -race -run TestProcessManager_Start ./internal/application/process/...

# Verbose output
go test -race -v ./...
```

### Linting

```bash
cd src

# Standard linting
golangci-lint run

# KTN convention linting
ktn-linter lint ./...

# Fix auto-fixable issues
golangci-lint run --fix
```

---

## Testing Strategy

### Test File Naming

```
┌────────────────────────────────────────────────────────────────┐
│                    TEST FILE CONVENTIONS                        │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  *_external_test.go   Black-box tests                          │
│                       package: package_test                     │
│                       Tests public API only                     │
│                       Import with path                          │
│                                                                 │
│  *_internal_test.go   White-box tests                          │
│                       package: same as source                   │
│                       Can access unexported symbols             │
│                       For testing internal logic                │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### Example External Test

```go
// manager_external_test.go
package process_test

import (
    "testing"

    "github.com/kodflow/daemon/internal/application/process"
    "github.com/stretchr/testify/assert"
)

func TestProcessManager_Start(t *testing.T) {
    // Test public API
    pm := process.NewManager(config, executor, logger)
    info, err := pm.Start()

    assert.NoError(t, err)
    assert.NotZero(t, info.PID)
}
```

### Example Internal Test

```go
// manager_internal_test.go
package process

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestCalculateBackoff(t *testing.T) {
    // Test internal function
    delay := calculateBackoff(3, 5*time.Second, 5*time.Minute, 2.0)

    assert.Equal(t, 40*time.Second, delay)
}
```

### Test Requirements

| Requirement | Description |
|-------------|-------------|
| Race Detection | All tests must pass with `-race` |
| Coverage | Aim for >80% coverage |
| External Tests | All public APIs must have external tests |
| Assertions | Use testify/assert or testify/require |

---

## Code Style

### Go Conventions

```go
// Good: Clear, concise naming
type ProcessManager struct {
    config   *ServiceConfig
    executor Executor
    state    State
}

func (pm *ProcessManager) Start() (ProcessInfo, error) {
    // Implementation
}

// Bad: Verbose, unclear
type ProcessManagerImpl struct {
    serviceConfiguration *ServiceConfigurationData
    processExecutor      ProcessExecutorInterface
    currentState         ProcessState
}
```

### Error Handling

```go
// Good: Wrap errors with context
func (pm *ProcessManager) Start() (ProcessInfo, error) {
    info, err := pm.executor.Start(pm.spec)
    if err != nil {
        return ProcessInfo{}, fmt.Errorf("start process %s: %w", pm.name, err)
    }
    return info, nil
}

// Bad: Lose context
func (pm *ProcessManager) Start() (ProcessInfo, error) {
    info, err := pm.executor.Start(pm.spec)
    if err != nil {
        return ProcessInfo{}, err
    }
    return info, nil
}
```

### Interface Design

```go
// Good: Small, focused interfaces (domain ports)
type Executor interface {
    Start(spec Spec) (ProcessInfo, error)
    Stop(pid int, timeout time.Duration) error
    Signal(pid int, sig os.Signal) error
}

// Bad: Large, unfocused interfaces
type ProcessService interface {
    Start(spec Spec) (ProcessInfo, error)
    Stop(pid int) error
    Signal(pid int, sig os.Signal) error
    GetStatus(pid int) Status
    ListProcesses() []ProcessInfo
    SetLogger(logger Logger)
    // ... many more methods
}
```

---

## Architecture Guidelines

### Hexagonal Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                           RULES                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Domain is PURE                                              │
│     • No imports from application or infrastructure             │
│     • Contains entities, value objects, port interfaces         │
│                                                                  │
│  2. Application orchestrates                                     │
│     • Imports domain only                                       │
│     • Implements use cases                                      │
│     • Defines bootstrap ports (Loader, Reloader)                │
│                                                                  │
│  3. Infrastructure implements                                    │
│     • Adapters implement domain port interfaces                 │
│     • Contains all I/O operations                               │
│     • OS-specific code in kernel/                               │
│                                                                  │
│  4. cmd/ is composition root                                    │
│     • Wires infrastructure into application                     │
│     • Handles CLI, signals, main loop                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Adding New Features

1. **Define domain types** (if needed)
   - Entities in `domain/`
   - Port interfaces in `domain/*/port.go`

2. **Implement application logic**
   - Use cases in `application/`
   - Orchestration, state management

3. **Create infrastructure adapters**
   - Implement domain ports
   - Add to `infrastructure/`

4. **Wire in cmd/**
   - Inject dependencies
   - Handle CLI arguments

---

## Git Workflow

### Branch Naming

| Type | Branch Pattern | Example |
|------|----------------|---------|
| Feature | `feat/<description>` | `feat/add-tcp-health-check` |
| Bug Fix | `fix/<description>` | `fix/restart-loop` |
| Refactor | `refactor/<description>` | `refactor/hexagonal-structure` |
| Docs | `docs/<description>` | `docs/add-configuration-guide` |

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code refactoring |
| `docs` | Documentation |
| `test` | Tests |
| `chore` | Maintenance |
| `ci` | CI/CD changes |

**Examples:**

```bash
feat(health): add TCP health check support
fix(process): prevent restart loop on immediate failure
refactor(kernel): move to infrastructure layer
docs(readme): update architecture diagram
test(supervisor): add event handling tests
```

### Pull Request Process

1. Create feature branch from `main`
2. Make changes, ensure tests pass
3. Run linters: `golangci-lint run` and `ktn-linter lint ./...`
4. Create PR with descriptive title
5. Wait for CI to pass
6. Request review
7. Squash merge when approved

---

## CI/CD

### GitHub Actions

```
┌────────────────────────────────────────────────────────────────┐
│                      CI PIPELINE                                │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  On Pull Request:                                               │
│    1. Checkout code                                             │
│    2. Setup Go                                                  │
│    3. Install dependencies                                      │
│    4. Run golangci-lint                                         │
│    5. Run ktn-linter                                            │
│    6. Run tests with race detection                            │
│    7. Upload coverage to Codacy                                 │
│                                                                 │
│  On Release Tag:                                                │
│    1. Build binaries for all platforms                         │
│    2. Create GitHub release                                     │
│    3. Upload binaries as assets                                 │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### Local CI Check

Before pushing, run:

```bash
cd src

# Lint
golangci-lint run
ktn-linter lint ./...

# Test with race detection
go test -race ./...

# Build
go build ./cmd/daemon
```

---

## Debugging

### Running with Debug Output

```bash
# Build with debug info
go build -gcflags="all=-N -l" -o supervizio ./cmd/daemon

# Run with debug logging
./supervizio --config config.yaml --log-level debug
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug ./cmd/daemon -- --config config.yaml

# Set breakpoint
(dlv) break internal/application/process/manager.go:42
(dlv) continue
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Import cycle | Check hexagonal layer rules |
| Race condition | Run with `-race`, use proper synchronization |
| Test flakiness | Avoid time.Sleep, use channels/contexts |
| Lint failures | Run `golangci-lint run --fix` |

---

## Contributing

1. Fork the repository
2. Create feature branch
3. Write tests first (TDD encouraged)
4. Implement feature
5. Ensure all tests pass with race detection
6. Ensure linters pass
7. Submit PR with clear description

### Code Review Checklist

- [ ] Tests included and passing
- [ ] Race detection clean
- [ ] Linters pass
- [ ] Documentation updated
- [ ] Follows hexagonal architecture
- [ ] Error handling with context
- [ ] No hardcoded values
