<!-- updated: 2026-02-15T21:30:00Z -->
# Observability - Monitoring

Logging and health checking of services.

## Role

Provide monitoring tools: log capture, service health verification.

## Navigation

| Need | Package |
|------|---------|
| Capture process stdout/stderr | `logging/` |
| Log daemon events | `logging/daemon/` |
| Publish lifecycle events | `events/` |
| Verify service health (TCP, HTTP, etc.) | `healthcheck/` |

## Structure

```
observability/
├── events/            # Event bus (lifecycle.Publisher adapter)
├── logging/           # Log capture and formatting
│   ├── capture.go     # stdout/stderr coordination
│   ├── linewriter.go  # Line-by-line writing
│   ├── multiwriter.go # Multiple destinations
│   ├── timestamp.go   # Timestamp prefixing
│   ├── writer.go      # Base file writer
│   ├── fileopener.go  # File opener (rotation-ready)
│   └── daemon/        # Daemon event logging
│       ├── logger.go         # MultiLogger (aggregates writers)
│       ├── writer_console.go # ConsoleWriter (stdout/stderr split)
│       ├── writer_file.go    # FileWriter with rotation
│       ├── writer_json.go    # JSONWriter for structured output
│       ├── writer_buffered.go # BufferedWriter
│       ├── level_filter.go   # LevelFilter wrapper
│       └── factory.go        # BuildLogger from config
│
└── healthcheck/       # Health probers
    ├── factory.go     # Factory by type
    ├── tcp.go         # TCP connect
    ├── udp.go         # UDP packet send
    ├── http.go        # HTTP GET/HEAD
    ├── grpc.go        # gRPC health/v1
    ├── exec.go        # Command execution
    ├── icmp.go        # Ping (fallback TCP)
    └── icmp_native.go # Raw ICMP (requires CAP_NET_RAW)
```

## Terminology

| Term | Description |
|------|-------------|
| **healthcheck** | Verify a service responds (TCP, HTTP...) |
| **probe/metrics** | System metrics collection (CPU, RAM) - see `../probe/` |
