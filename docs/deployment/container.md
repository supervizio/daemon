# Container Deployment (PID1)

superviz.io is designed as a PID1-capable process supervisor for containers, handling zombie process reaping and signal forwarding.

---

## PID1 Behavior

When running as PID1 (process ID 1) inside a container, superviz.io:

1. **Reaps zombie processes**: Calls `waitpid(-1)` to clean up any orphaned child processes
2. **Forwards signals**: SIGTERM/SIGINT are forwarded to all supervised children
3. **Graceful shutdown**: Sends SIGTERM to children, waits for exit, then exits itself
4. **Prevents zombie accumulation**: Continuous reaping loop prevents process table exhaustion

The `ProvideReaper` Wire provider automatically detects PID1 mode:

```go
// Returns ZombieReaper only if running as PID 1
func ProvideReaper() *reaper.UnixZombieReaper {
    if os.Getpid() == 1 {
        return reaper.New()
    }
    return nil
}
```

---

## Dockerfile Example

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

---

## Docker Compose

```yaml
services:
  supervisor:
    image: supervizio:latest
    init: false  # Important: do NOT use Docker init, supervizio IS the init
    volumes:
      - ./config.yaml:/etc/supervizio/config.yaml:ro
      - /var/run/docker.sock:/var/run/docker.sock  # For Docker discovery
    cap_add:
      - NET_RAW  # For ICMP probes
    restart: unless-stopped
```

---

## Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: supervizio
spec:
  template:
    spec:
      containers:
        - name: supervizio
          image: supervizio:latest
          args: ["--config", "/etc/supervizio/config.yaml"]
          volumeMounts:
            - name: config
              mountPath: /etc/supervizio
          livenessProbe:
            grpc:
              port: 50051
            initialDelaySeconds: 5
          readinessProbe:
            grpc:
              port: 50051
      volumes:
        - name: config
          configMap:
            name: supervizio-config
```

---

## Signal Handling

| Signal | Behavior |
|--------|----------|
| `SIGTERM` | Graceful shutdown: stop all services, wait, exit |
| `SIGINT` | Same as SIGTERM (Ctrl+C) |
| `SIGHUP` | Reload configuration without restarting |
| `SIGCHLD` | Reap zombie processes (PID1 mode only) |

---

## Capabilities

| Capability | Required For |
|-----------|-------------|
| `CAP_NET_RAW` | ICMP health probes (native mode) |
| `CAP_SETUID` | Running services as different users |
| `CAP_SETGID` | Running services as different groups |
