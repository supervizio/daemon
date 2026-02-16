<!-- updated: 2026-02-15T21:30:00Z -->
# Healthcheck - Health Probers

Service health verification via multiple protocols.

## Role

Test that a service is accessible and responding correctly.

**Note**: Different from `../probe/` which collects system metrics (CPU, RAM).

## Available Probers

| Type | File | Description |
|------|------|-------------|
| TCP | `tcp.go` | Successful TCP connection |
| UDP | `udp.go` | Packet send (connectionless) |
| HTTP | `http.go` | GET/HEAD, status validation |
| gRPC | `grpc.go` | health/v1 protocol |
| Exec | `exec.go` | Command exit code 0 |
| ICMP | `icmp.go` | Ping (TCP fallback without CAP_NET_RAW) |
| ICMP Native | `icmp_native.go` | Raw ICMP (requires CAP_NET_RAW) |

## Factory

```go
factory := NewFactory(5 * time.Second)  // Default timeout
prober, err := factory.Create("http", 10*time.Second)

// Or typed methods
tcpProber := factory.CreateTCP(5 * time.Second)
httpProber := factory.CreateHTTP(10 * time.Second)
```

## Implemented Interface

```go
// domain/health/prober.go
type Prober interface {
    Probe(ctx context.Context, target Target) CheckResult
}
```

## Constructors

```go
NewTCPProber(timeout time.Duration) *TCPProber
NewHTTPProber(timeout time.Duration) *HTTPProber
NewGRPCProber(timeout time.Duration) *GRPCProber
NewExecProber(timeout time.Duration) *ExecProber
NewICMPProber(timeout time.Duration) *ICMPProber
NewUDPProber(timeout time.Duration) *UDPProber
```

## Security

`ExecProber` uses `executor.TrustedCommand()` - see `process/executor/CLAUDE.md`.
