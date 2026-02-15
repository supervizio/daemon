# Domain Config Package

Configuration value objects for services managed by the supervisor.

## Files (48 files)

| Category | Key Files | Purpose |
|----------|-----------|---------|
| **Core** | `config.go`, `serviceconfig.go`, `validate.go` | Root config, service definition, validation |
| **Restart** | `restart.go` | RestartConfig, RestartPolicy enum |
| **Network** | `listener.go`, `probeconfig.go`, `healthcheck.go` | Listener, probe, health check configs |
| **Logging** | `loggingconfig.go`, `logstreamconfig.go`, `rotationconfig.go` | Global logging, per-stream, rotation |
| **Logging (ext)** | `daemonlogging.go`, `servicelogging.go`, `logdefaults.go` | Daemon logging, service logging, defaults |
| **Writers** | `writer_config.go`, `file_writer_config.go`, `json_writer_config.go` | Log output destinations |
| **Monitoring** | `monitoring.go`, `monitoring_defaults.go`, `metrics_config.go` | External monitoring, metrics config |
| **Targets** | `target_config.go`, `discovery_config.go` | Target and discovery base configs |
| **Discovery** | `docker_discovery_config.go`, `kubernetes_discovery_config.go` | Docker, K8s discovery |
|  | `nomad_discovery_config.go`, `systemd_discovery_config.go` | Nomad, systemd discovery |
|  | `podman_discovery_config.go`, `bsdrc_discovery_config.go` | Podman, BSD rc discovery |
|  | `openrc_discovery_config.go`, `portscan_discovery_config.go` | OpenRC, port scan discovery |

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
