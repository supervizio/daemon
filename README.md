# superviz.io

[![CI](https://github.com/supervizio/daemon/actions/workflows/ci.yml/badge.svg)](https://github.com/supervizio/daemon/actions/workflows/ci.yml)
[![Release](https://github.com/supervizio/daemon/actions/workflows/release.yml/badge.svg)](https://github.com/supervizio/daemon/releases)
[![codecov](https://codecov.io/gh/supervizio/daemon/branch/main/graph/badge.svg)](https://codecov.io/gh/supervizio/daemon)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

PID1-capable process supervisor for containers and Linux/BSD systems. Manages multiple services with health checks, restart policies, and log rotation.

## Features

- **Multi-service management**: Start, stop, and monitor multiple processes
- **Health checks**: HTTP, TCP, and shell command with configurable retries
- **Restart policies**: `always`, `on-failure`, `never`, `unless-stopped`
- **Exponential backoff**: Increasing delays between restart attempts
- **Log rotation**: By size, age, and file count with compression
- **PID 1 support**: Zombie process reaping, signal handling
- **Multi-platform**: Linux, BSD (FreeBSD, OpenBSD, NetBSD), macOS

## Installation

### Pre-built binaries

Download from [GitHub releases](https://github.com/supervizio/daemon/releases).

### From source

```bash
git clone https://github.com/supervizio/daemon.git
cd daemon/src
go build -o supervizio ./cmd/daemon
```

## Usage

```bash
# With configuration file
supervizio --config /etc/supervizio/config.yaml

# Show version
supervizio --version
```

## Configuration

```yaml
version: "1"

logging:
  base_dir: /var/log/supervizio
  defaults:
    timestamp_format: iso8601
    rotation:
      max_size: 100MB
      max_age: 7d
      max_files: 10
      compress: true

services:
  - name: webapp
    command: /usr/bin/node
    args:
      - /app/server.js
    user: www-data
    environment:
      NODE_ENV: production
    restart:
      policy: always
      max_retries: 5
      delay: 5s
      delay_max: 5m
    health_checks:
      - type: http
        endpoint: http://localhost:3000/health
        interval: 30s
        timeout: 5s
        retries: 3
    logging:
      stdout:
        file: webapp.out.log
      stderr:
        file: webapp.err.log
```

See [examples/config.yaml](examples/config.yaml) for a complete configuration.

## Signals

| Signal | Action |
|--------|--------|
| `SIGTERM` / `SIGINT` | Graceful shutdown of all services |
| `SIGHUP` | Configuration reload |
| `SIGCHLD` | Zombie process reaping (PID 1) |

## Architecture

```
src/
├── cmd/daemon/          # Entry point
└── internal/
    ├── config/          # YAML parsing and validation
    ├── supervisor/      # Service orchestration
    ├── process/         # Lifecycle management
    ├── health/          # Health checks (HTTP, TCP, cmd)
    ├── kernel/          # OS abstraction (ports & adapters)
    │   ├── ports/       # Interfaces
    │   └── adapters/    # Platform-specific implementations
    └── logging/         # Log rotation and capture
```

## Development

### Prerequisites

- Go 1.25+
- golangci-lint

### Build

```bash
cd src
go build ./...
```

### Tests

```bash
cd src
go test -race -cover ./...
```

### Lint

```bash
cd src
golangci-lint run
```

## Supported Platforms

| OS | Architectures |
|----|---------------|
| Linux | amd64, arm64, 386, armv7 |
| FreeBSD | amd64, arm64 |
| OpenBSD | amd64, arm64 |
| NetBSD | amd64, arm64 |
| DragonFlyBSD | amd64 |
| macOS | amd64, arm64 |

## License

MIT
