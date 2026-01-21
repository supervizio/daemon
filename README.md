# superviz.io

[![CI](https://github.com/supervizio/daemon/actions/workflows/ci.yml/badge.svg)](https://github.com/supervizio/daemon/actions/workflows/ci.yml)
[![Release](https://github.com/supervizio/daemon/actions/workflows/release.yml/badge.svg)](https://github.com/supervizio/daemon/releases)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/c66eb99290744de6ac6a6e082f83daaf)](https://app.codacy.com/gh/supervizio/daemon/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Coverage](https://app.codacy.com/project/badge/Coverage/c66eb99290744de6ac6a6e082f83daaf)](https://app.codacy.com/gh/supervizio/daemon/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)
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

Built with **Hexagonal Architecture** (Ports & Adapters):

```
src/
├── cmd/daemon/                    # Entry point
└── internal/
    ├── application/               # Use cases & orchestration
    │   ├── supervisor/            # Service orchestration
    │   ├── process/               # Process lifecycle management
    │   ├── health/                # Health monitoring
    │   └── config/                # Configuration ports
    ├── domain/                    # Core business logic
    │   ├── service/               # Service configuration entities
    │   ├── process/               # Process entities & ports
    │   ├── health/                # Health entities
    │   └── shared/                # Shared value objects
    └── infrastructure/            # External adapters
        ├── config/yaml/           # YAML configuration loader
        ├── health/                # HTTP, TCP, command checkers
        ├── process/               # Unix process executor
        ├── kernel/                # OS abstraction (signals, reaper)
        └── logging/               # File writers, rotation, capture
```

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

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

## E2E Testing Matrix

[![E2E Tests](https://github.com/supervizio/daemon/actions/workflows/e2e.yml/badge.svg)](https://github.com/supervizio/daemon/actions/workflows/e2e.yml)

Testing across all supported init systems (AMD64):

### Linux

| Distribution | Init System | VM | Docker |
|--------------|-------------|:--:|:------:|
| Debian 12 | systemd | ✅ Vagrant | ✅ |
| Ubuntu 22.04 | systemd | ✅ Vagrant | ✅ |
| Alpine 3.19 | OpenRC | ✅ Vagrant | ✅ |
| Devuan 4 | SysVinit | ✅ Vagrant | ✅ |
| Void Linux | runit | - | ✅ |

### BSD

| OS | Init System | VM |
|----|-------------|:--:|
| FreeBSD 14 | rc.d | ✅ Vagrant |
| OpenBSD 7 | rc.d | ✅ Vagrant |
| NetBSD 10 | rc.d | ✅ Vagrant |
| DragonFlyBSD 6 | rc.d | ✅ Vagrant |

**Total: 9 jobs** covering all init systems (systemd, OpenRC, SysVinit, runit, rc.d)

## License

MIT
