# Infrastructure Healthcheck Package

Protocol adapters implementing the domain Prober interface for service health checking.

## Purpose

This package provides concrete implementations for **service health checks** - verifying
that services are reachable and responding correctly via various protocols.

Note: This is different from **probe/metrics** which collect system telemetry (CPU, RAM, DISK).

## Files

| File | Purpose |
|------|---------|
| `tcp.go` | TCP connection health checks |
| `udp.go` | UDP packet health checks |
| `http.go` | HTTP endpoint health checks |
| `grpc.go` | gRPC health checks (TCP fallback) |
| `exec.go` | Command execution health checks |
| `icmp.go` | ICMP ping health checks (TCP fallback) |
| `factory.go` | Health checker factory |

## Adapters

### TCPProber
- Verifies TCP port is accepting connections
- Uses `net.DialContext` with timeout
- Measures connection latency

### UDPProber
- Sends UDP packet and optionally waits for response
- UDP is connectionless - timeout doesn't mean failure
- Useful for DNS, NTP services

### HTTPProber
- Sends HTTP request to endpoint
- Validates response status code
- Uses `http.RoundTrip` (no redirect following)

### GRPCProber
- Uses gRPC health/v1 protocol for health checks
- Supports both insecure and TLS connections
- Returns service status (SERVING, NOT_SERVING, UNKNOWN)

### ExecProber
- Executes command via `TrustedCommand`
- Exit code 0 = success, non-zero = failure
- Captures stdout/stderr

### ICMPProber
- TCP fallback (ICMP requires CAP_NET_RAW)
- Measures network latency
- Useful for node-to-node connectivity

## Factory

```go
factory := NewFactory(5 * time.Second)
prober, err := factory.Create("http", 10*time.Second)
```

Supported types: `tcp`, `udp`, `http`, `grpc`, `exec`, `icmp`

## Security Notes

- ExecProber uses `process.TrustedCommand`
- Commands must come from validated configuration
- See `infrastructure/process/CLAUDE.md` for security model

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/healthcheck/` | Port interface definition |
| `application/health/` | Orchestrates health checks |
| `infrastructure/metrics/` | System metrics collection (different concern) |
