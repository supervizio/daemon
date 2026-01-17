# Domain Config Package

Domain value objects for service configuration (renamed from service/).

This package defines the configuration model for services managed by the supervisor,
including service definitions, restart policies, listener configurations, logging
settings, and health check configurations.

## Files

| File | Purpose |
|------|---------|
| `config.go` | `Config` - root configuration structure |
| `serviceconfig.go` | `ServiceConfig` - service definition |
| `listener.go` | `ListenerConfig` - network listener with probe |
| `probeconfig.go` | `ProbeConfig` - health probe configuration |
| `restart.go` | `RestartConfig`, `RestartPolicy` - restart behavior |
| `loggingconfig.go` | `LoggingConfig` - global logging defaults |
| `logdefaults.go` | `LogDefaults` - default logging settings |
| `logstreamconfig.go` | `LogStreamConfig` - per-stream log settings |
| `servicelogging.go` | `ServiceLogging` - per-service logging |
| `rotationconfig.go` | `RotationConfig` - log rotation settings |
| `healthcheck.go` | `HealthCheckConfig` - legacy health checks (deprecated) |
| `validate.go` | Configuration validation functions |

## Key Types

### Config (Root Configuration)

```go
type Config struct {
    Version    string
    Logging    LoggingConfig
    Services   []ServiceConfig
    ConfigPath string
}
```

### ServiceConfig (Service Definition)

```go
type ServiceConfig struct {
    Name             string
    Command          string
    Args             []string
    User             string
    Group            string
    WorkingDirectory string
    Environment      map[string]string
    Restart          RestartConfig
    HealthChecks     []HealthCheckConfig  // deprecated
    Listeners        []ListenerConfig
    Logging          ServiceLogging
    DependsOn        []string
    Oneshot          bool
}
```

### ListenerConfig (Network Listener)

```go
type ListenerConfig struct {
    Name     string
    Port     int
    Protocol string       // "tcp", "udp"
    Address  string       // optional bind address
    Probe    *ProbeConfig // optional probe config
}
```

### ProbeConfig (Health Probe)

```go
type ProbeConfig struct {
    Type             string  // "tcp", "http", "grpc", "exec"
    Path             string  // HTTP path
    Service          string  // gRPC service name
    Command          string  // exec command
    Args             []string
    Interval         shared.Duration
    Timeout          shared.Duration
    SuccessThreshold int
    FailureThreshold int
}
```

### RestartConfig (Restart Behavior)

```go
type RestartConfig struct {
    Policy     RestartPolicy
    MaxRetries int
    Delay      shared.Duration
    DelayMax   shared.Duration  // for exponential backoff
}
```

### RestartPolicy (Enum)

```go
const (
    RestartAlways    RestartPolicy = "always"
    RestartOnFailure RestartPolicy = "on-failure"
    RestartNever     RestartPolicy = "never"
    RestartUnless    RestartPolicy = "unless-stopped"
)
```

### LoggingConfig (Global Logging)

```go
type LoggingConfig struct {
    Defaults LogDefaults
    BaseDir  string
}
```

### RotationConfig (Log Rotation)

```go
type RotationConfig struct {
    MaxSize  string  // "100MB", "1GB"
    MaxFiles int
}
```

## Factory Functions

- `NewConfig(services)` - Create config with services
- `DefaultConfig()` - Create config with defaults
- `NewServiceConfig(name, command)` - Create service config
- `NewListenerConfig(name, port)` - Create TCP listener
- `NewRestartConfig(policy)` - Create restart config
- `DefaultRestartConfig()` - Default restart on failure
- `DefaultLoggingConfig()` - Default logging settings
- `DefaultProbeConfig(probeType)` - Default probe config
- `DefaultRotationConfig()` - Default rotation settings

## Builder Methods

### ListenerConfig
- `WithProbe(probe)` - Add probe configuration
- `WithTCPProbe()` - Add TCP connection probe
- `WithHTTPProbe(path)` - Add HTTP endpoint probe
- `WithGRPCProbe(service)` - Add gRPC health probe

## Dependencies

- Depends on: `domain/shared` (for Duration)
- Used by: `application/supervisor`, `infrastructure/persistence/config/yaml`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/shared` | Duration value object |
| `infrastructure/persistence/config/yaml` | YAML config loader implements parsing |
| `application/supervisor` | Uses Config to orchestrate services |
