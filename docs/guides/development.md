# Development Guide

This guide covers the development environment setup, testing, and contribution workflow for superviz.io.

---

## DevContainer

The project includes a full DevContainer configuration with Go, Rust, and Node.js toolchains pre-installed.

### Requirements

- Docker Desktop or compatible container runtime
- VS Code with DevContainers extension (or GitHub Codespaces)

### Setup

1. Open the project in VS Code
2. Accept the "Reopen in Container" prompt
3. Wait for container build and initialization

The container includes:

- Go 1.25.6 with `golangci-lint`, `ktn-linter`, `gofumpt`
- Rust stable with `clippy`, `cargo-nextest`
- Node.js (for MCP servers)
- Docker-in-Docker support

---

## Project Structure

```
src/
├── cmd/daemon/               # CLI entry point
├── internal/
│   ├── bootstrap/            # Wire DI
│   ├── application/          # Use cases
│   ├── domain/               # Entities, ports
│   └── infrastructure/       # Adapters
├── lib/probe/                # Rust probe library
├── api/proto/                # Protobuf definitions
├── go.mod
└── go.sum
```

---

## Building

```bash
cd src

# Build probe (Rust)
make build-probe

# Build daemon (Go)
make build-daemon

# Build both
make build-hybrid

# Quick Go build (without probe)
go build ./cmd/daemon
```

---

## Testing

```bash
# Run all tests with race detection
go test -race ./...

# Run tests for a specific package
go test -race ./internal/application/supervisor/...

# Run with verbose output
go test -race -v ./internal/domain/process/...
```

### Test Conventions

| File Pattern | Type | Description |
|-------------|------|-------------|
| `*_external_test.go` | Black-box | Tests public API via `package_test` |
| `*_internal_test.go` | White-box | Tests internals within same package |

Race detection (`-race`) is always required.

---

## Linting

```bash
# Standard Go linting
golangci-lint run -c ../.golangci.yml

# KTN convention linting
ktn-linter lint -c ../.ktn-linter.yaml ./...
```

Both linters must pass with zero issues.

---

## Code Generation

### Wire (Dependency Injection)

```bash
# Regenerate wire_gen.go
wire ./internal/bootstrap/
```

### Protobuf

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/v1/daemon/daemon.proto
```

---

## Git Workflow

| Type | Branch | Commit |
|------|--------|--------|
| Feature | `feat/<description>` | `feat(scope): message` |
| Bug fix | `fix/<description>` | `fix(scope): message` |
| Documentation | `docs/<description>` | `docs(scope): message` |
| Refactoring | `refactor/<description>` | `refactor(scope): message` |

Commits follow [Conventional Commits](https://www.conventionalcommits.org/) format.

---

## Architecture Rules

1. **Domain has no dependencies** on application or infrastructure
2. **Application depends only on domain** (uses port interfaces)
3. **Infrastructure implements domain ports** (dependency inversion)
4. **No circular dependencies** between packages
5. **Wire binds interfaces** to implementations at compile time
