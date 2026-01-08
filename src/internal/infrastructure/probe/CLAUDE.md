# Infrastructure Probe Package

Protocol adapters implementing the domain Prober interface.

## Files

| File | Purpose |
|------|---------|
| `tcp.go` | TCP connection probes |
| `udp.go` | UDP packet probes |
| `http.go` | HTTP endpoint probes |
| `grpc.go` | gRPC health probes (TCP fallback) |
| `exec.go` | Command execution probes |
| `icmp.go` | ICMP ping probes (TCP fallback) |
| `factory.go` | Prober factory |

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
- Currently: TCP connectivity check
- TODO: Full gRPC health/v1 protocol
- Requires `google.golang.org/grpc` for full support

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
