# E2E Behavioral Tests

Runtime behavior tests for supervizio using testcontainers-go.

## Structure

```
behavioral/
├── crasher/              # Test binary with configurable behavior
├── Dockerfile.behavioral # Image: supervizio + crasher
├── helpers_test.go       # Shared utilities
├── *_test.go             # Test files
└── testdata/             # YAML configs
```

## Crasher Binary

Configurable process for testing restart/health scenarios.

| Flag | Default | Description |
|------|---------|-------------|
| `--exit` | 0 | Exit code |
| `--delay` | 0 | Delay before exit |
| `--port` | 0 | TCP/HTTP port |
| `--http` | false | Serve /health endpoint |
| `--healthy` | true | /health returns 200 (false=503) |
| `--crash-after` | 0 | Crash after N seconds |
| `--orphan` | false | Spawn orphan process |
| `--ignore-term` | false | Ignore SIGTERM |

## Test Coverage

| File | Tests |
|------|-------|
| `restart_test.go` | Restart policies (always, on-failure, never, unless-stopped, max_retries) |
| `backoff_test.go` | Exponential backoff, delay_max cap |
| `health_test.go` | HTTP/TCP probes, healthy status |
| `pid1_test.go` | PID1 identity, zombie reaping, signal forwarding, orphan adoption |

## Execution

```bash
# Build prerequisites
make build-linux     # supervizio binary
make build-crasher   # crasher binary
make build-behavioral-image  # Docker image

# Run tests
cd e2e/behavioral && go test -v -race ./...
```

## Dependencies

- testcontainers-go v0.37.0
- Docker daemon

## CI

Job `e2e-behavioral` in `.github/workflows/ci.yml`.
