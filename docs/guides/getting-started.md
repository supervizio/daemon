# Getting Started

This guide walks through building, configuring, and running superviz.io.

---

## Prerequisites

- Go 1.25.6+
- Rust stable (for the probe library)
- CGO enabled (`CGO_ENABLED=1`)

---

## Build

```bash
# Clone the repository
git clone https://github.com/supervizio/daemon.git
cd daemon/src

# Build the Rust probe library
make build-probe

# Build the Go binary
make build-daemon

# Or build both in one step
make build-hybrid
```

The binary is created at `./supervizio` (or `./cmd/daemon/daemon`).

---

## Configure

Create a configuration file:

```yaml
version: "1"

services:
  - name: my-app
    command: /usr/local/bin/my-app
    args:
      - "--port=8080"
    env:
      LOG_LEVEL: info
    restart:
      policy: on-failure
      max_retries: 5
      delay: 5s
    listeners:
      - name: http
        port: 8080
        protocol: tcp
        probe:
          type: http
          path: /health
          interval: 30s
          timeout: 5s
```

See [Configuration](../configuration/index.md) for the full reference.

---

## Run

```bash
# Run with configuration file
./supervizio --config config.yaml

# Run with interactive TUI
./supervizio --config config.yaml --tui
```

---

## Verify

### gRPC API

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Get daemon state
grpcurl -plaintext localhost:50051 daemon.v1.DaemonService/GetState

# List processes
grpcurl -plaintext localhost:50051 daemon.v1.DaemonService/ListProcesses

# Stream system metrics
grpcurl -plaintext localhost:50051 daemon.v1.MetricsService/StreamSystemMetrics
```

### Health Check

```bash
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

---

## Docker Quick Start

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY src/ .
RUN make build-hybrid

FROM alpine:3.21
COPY --from=builder /build/supervizio /usr/local/bin/supervizio
COPY config.yaml /etc/supervizio/config.yaml
ENTRYPOINT ["/usr/local/bin/supervizio"]
CMD ["--config", "/etc/supervizio/config.yaml"]
```

```bash
docker build -t supervizio .
docker run -d --name supervisor supervizio
```

---

## Next Steps

- [Configuration Reference](../configuration/index.md) --- Full YAML configuration options
- [API Reference](../api/index.md) --- gRPC service definitions
- [Architecture](../architecture/index.md) --- Understand the internal design
- [Deployment](../deployment/index.md) --- Production deployment guides
