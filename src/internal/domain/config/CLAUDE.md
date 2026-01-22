# Domain Config Package

Configuration value objects for services managed by the supervisor.

## Files

| File | Purpose |
|------|---------|
| `config.go` | `Config` - root configuration |
| `serviceconfig.go` | `ServiceConfig` - service definition |
| `listener.go` | `ListenerConfig` - network listener with probe |
| `probeconfig.go` | `ProbeConfig` - health probe configuration |
| `restart.go` | `RestartConfig`, `RestartPolicy` |
| `loggingconfig.go` | `LoggingConfig` - global logging |
| `logstreamconfig.go` | `LogStreamConfig` - per-stream settings |
| `rotationconfig.go` | `RotationConfig` - log rotation |
| `validate.go` | Configuration validation |

## Key Types

### Config (Root)
- `Version`, `Logging`, `Services[]`, `ConfigPath`

### ServiceConfig
- `Name`, `Command`, `Args`, `User`, `Group`, `WorkingDirectory`
- `Environment`, `Restart`, `Listeners[]`, `Logging`, `DependsOn`, `Oneshot`

### ListenerConfig
- `Name`, `Port`, `Protocol` (tcp/udp), `Address`, `Probe`
- Builder: `WithProbe()`, `WithTCPProbe()`, `WithHTTPProbe(path)`, `WithGRPCProbe(svc)`

### ProbeConfig
- `Type` (tcp, http, grpc, exec), `Path`, `Service`, `Command`, `Args`
- `Interval`, `Timeout`, `SuccessThreshold`, `FailureThreshold`

### RestartConfig
- `Policy`, `MaxRetries`, `Delay`, `DelayMax` (for exponential backoff)

### RestartPolicy (Enum)
- `RestartAlways`, `RestartOnFailure`, `RestartNever`, `RestartUnless`

### RotationConfig
- `MaxSize` ("100MB", "1GB"), `MaxFiles`

## Factory Functions

- `NewConfig(services)`, `DefaultConfig()`
- `NewServiceConfig(name, command)`
- `NewListenerConfig(name, port)`
- `NewRestartConfig(policy)`, `DefaultRestartConfig()`
- `DefaultLoggingConfig()`, `DefaultProbeConfig(type)`

## Dependencies

- Depends on: `domain/shared` (Duration)
- Used by: `application/supervisor`, `infrastructure/persistence/config/yaml`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/shared` | Duration value object |
| `infrastructure/persistence/config/yaml` | Parses YAML to Config |
| `application/supervisor` | Uses Config for orchestration |
