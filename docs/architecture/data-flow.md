# Data Flow

This page traces the major data flows through the superviz.io daemon, from startup through steady-state operation.

---

## Startup Flow

```mermaid
sequenceDiagram
    participant M as main.go
    participant B as Bootstrap
    participant W as Wire DI
    participant C as Config Loader
    participant S as Supervisor
    participant LM as Lifecycle Manager
    participant HM as Health Monitor
    participant MT as Metrics Tracker
    participant G as gRPC Server

    M->>B: Run()
    B->>W: Inject dependencies
    W->>C: LoadConfig(path)
    C-->>W: Config
    W->>S: NewSupervisor(config, deps)
    B->>S: Start()
    S->>LM: Start(service) per service
    S->>HM: StartMonitoring()
    S->>MT: StartTracking()
    S->>G: Serve(:50051)
    Note over B: Block on signal (SIGTERM/SIGINT)
```

---

## Process Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Stopped
    Stopped --> Starting: Start()
    Starting --> Running: Process alive + health OK
    Running --> Stopping: Stop() / SIGTERM
    Stopping --> Stopped: Process exited
    Running --> Failed: Process crashed
    Failed --> Starting: Restart policy allows
    Failed --> Stopped: Max retries exceeded

    note right of Starting: Executor.Start() creates exec.Cmd
    note right of Running: Health probes active, metrics collected
    note right of Failed: ExitResult logged, RestartTracker updated
```

Each service is managed by a dedicated `lifecycle.Manager` that:

1. Creates the process via the `Executor` port
2. Monitors process exit via waitpid
3. Applies restart policy (always, on-failure, never, unless-stopped)
4. Tracks restart count with backoff delay (configurable `delay` to `delay_max`)

---

## Health Check Flow

```mermaid
sequenceDiagram
    participant HM as ProbeMonitor
    participant F as Prober Factory
    participant P as Prober (TCP/HTTP/gRPC/ICMP)
    participant S as Service

    HM->>F: CreateProber(config)
    F-->>HM: Prober instance
    loop Every interval
        HM->>P: Probe(ctx, target)
        P->>S: Connect/Request
        S-->>P: Response
        P-->>HM: Result{Status, Latency}
        alt failure_threshold reached
            HM->>HM: Mark unhealthy
        end
        alt success_threshold reached
            HM->>HM: Mark healthy
        end
    end
```

Supported probe types:

| Type | Protocol | Check Method |
|------|----------|-------------|
| `tcp` | TCP | Connection establishment |
| `http` | HTTP/HTTPS | Status code + optional path |
| `grpc` | gRPC | `grpc.health.v1.Health/Check` |
| `icmp` | ICMP | Ping (native or fallback mode) |
| `udp` | UDP | Packet send/receive |
| `exec` | Shell | Command exit code |

---

## Metrics Collection Flow

```mermaid
sequenceDiagram
    participant MT as Metrics Tracker
    participant PR as Probe (Rust FFI)
    participant OS as /proc /sys
    participant ST as BoltDB Store
    participant G as gRPC Stream

    loop Every collection interval
        MT->>PR: Collect(ctx)
        PR->>OS: Read procfs/sysfs
        OS-->>PR: Raw metrics
        PR-->>MT: ProcessMetrics{CPU, Memory}
        MT->>ST: Store(metrics)
        MT->>G: Push to active streams
    end
```

The probe pipeline:

1. **Go** calls `Collector.Collect()` (application port)
2. **CGO/FFI** bridges to `libprobe.a` (Rust static library)
3. **Rust** reads OS-specific APIs (`/proc`, `/sys`, syscalls)
4. **Metrics** flow back through FFI to Go types
5. **BoltDB** persists time-series data
6. **gRPC** streams push to connected clients

---

## gRPC Request Flow

```mermaid
sequenceDiagram
    participant C as gRPC Client
    participant S as gRPC Server
    participant SUP as Supervisor
    participant MP as MetricsProvider
    participant SP as StateProvider

    C->>S: GetState(Empty)
    S->>SP: GetState()
    SP-->>S: DaemonState
    S-->>C: DaemonState

    C->>S: StreamSystemMetrics(interval)
    loop Every interval
        S->>MP: GetSystemMetrics()
        MP-->>S: SystemMetrics
        S-->>C: SystemMetrics (stream)
    end

    Note over C,S: Client cancellation terminates stream
```

The gRPC server implements two provider interfaces:

```go
type MetricsProvider interface {
    GetProcessMetrics(serviceName string) (metrics.ProcessMetrics, error)
    GetAllProcessMetrics() []metrics.ProcessMetrics
}

type StateProvider interface {
    GetState() state.DaemonState
}
```

---

## Configuration Loading

```mermaid
graph LR
    YAML["config.yaml"] -->|yaml.v3| Loader["YAML Loader"]
    Loader --> Config["domain.Config"]
    Config --> SC["ServiceConfig[]"]
    Config --> LC["LoggingConfig"]
    Config --> MC["MonitoringConfig"]
    SC --> SUP["Supervisor"]
    MC --> DIS["Discovery Factory"]

    style YAML fill:#fbdf411a,stroke:#fbdf41,color:#d4d8e0
    style Loader fill:#fbdf411a,stroke:#fbdf41,color:#d4d8e0
    style Config fill:#41fbdf1a,stroke:#41fbdf,color:#d4d8e0
    style SC fill:#41fbdf1a,stroke:#41fbdf,color:#d4d8e0
    style LC fill:#41fbdf1a,stroke:#41fbdf,color:#d4d8e0
    style MC fill:#41fbdf1a,stroke:#41fbdf,color:#d4d8e0
    style SUP fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style DIS fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
```

Configuration flows:

1. **YAML file** parsed by `gopkg.in/yaml.v3`
2. **Domain types** (`ServiceConfig`, `RestartConfig`, `ProbeConfig`) populated
3. **Supervisor** receives service configurations
4. **Discovery factory** receives monitoring/discovery configurations
5. **SIGHUP** triggers config reload via `Reloader` port
