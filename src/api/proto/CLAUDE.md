<!-- updated: 2026-02-15T21:30:00Z -->
# API Protocol Buffers

gRPC API definitions for supervizio daemon.

## Structure

```
proto/
└── v1/
    └── daemon/        # Daemon API v1
        ├── daemon.proto
        ├── daemon.pb.go
        └── daemon_grpc.pb.go
```

## Versioning

API versions follow `/vN/` directory convention.

| Version | Status | Description |
|---------|--------|-------------|
| v1 | Current | Process lifecycle and metrics |

## Code Generation

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       v1/daemon/daemon.proto
```

## Related

| Directory | See |
|-----------|-----|
| v1/daemon | `v1/daemon/CLAUDE.md` |
| gRPC server | `internal/infrastructure/transport/grpc/CLAUDE.md` |
