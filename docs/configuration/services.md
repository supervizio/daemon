# Service Configuration

Services are the core of superviz.io --- each service definition describes a process to supervise with its command, environment, restart policy, and health probes.

---

## Minimal Service

```yaml
services:
  - name: my-app
    command: /usr/local/bin/my-app
```

---

## Full Service Definition

```yaml
services:
  - name: web-server
    command: /usr/bin/nginx
    args:
      - "-g"
      - "daemon off;"
    working_dir: /var/lib/nginx
    user: www-data
    env:
      NODE_ENV: production
      LOG_LEVEL: info
    restart:
      policy: always
      max_retries: 5
      delay: 5s
      delay_max: 60s
    listeners:
      - name: http
        port: 80
        protocol: tcp
        probe:
          type: http
          path: /health
          interval: 10s
          timeout: 5s
          failure_threshold: 3
          success_threshold: 1
      - name: https
        port: 443
        protocol: tcp
        probe:
          type: tcp
          interval: 10s
```

---

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Unique service name (used in API and logs) |
| `command` | `string` | Yes | Executable path |
| `args` | `list[string]` | No | Command arguments |
| `working_dir` | `string` | No | Working directory for the process |
| `user` | `string` | No | Run as this user (requires root) |
| `env` | `map[string, string]` | No | Environment variables |
| `restart` | `object` | No | [Restart policy](#restart-policy) |
| `listeners` | `list[object]` | No | [Listener definitions](#listeners) |

---

## Restart Policy

```yaml
restart:
  policy: on-failure
  max_retries: 5
  delay: 5s
  delay_max: 60s
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `policy` | `string` | `never` | Restart policy (see below) |
| `max_retries` | `int` | `0` | Maximum restart attempts (0 = unlimited for `always`) |
| `delay` | `duration` | `1s` | Initial delay before restart |
| `delay_max` | `duration` | `60s` | Maximum delay (exponential backoff cap) |

### Policies

| Policy | Behavior |
|--------|----------|
| `always` | Restart on any exit (including exit code 0) |
| `on-failure` | Restart only on non-zero exit code |
| `never` | Never restart |
| `unless-stopped` | Restart unless explicitly stopped via API |

The restart delay uses exponential backoff: each consecutive failure doubles the delay until `delay_max` is reached. A successful health check resets the counter.

---

## Listeners

Listeners define network ports that a service exposes, along with optional health probes.

```yaml
listeners:
  - name: http
    port: 8080
    protocol: tcp
    probe:
      type: http
      path: /health
      interval: 30s
      timeout: 5s
      failure_threshold: 3
      success_threshold: 1
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Listener name |
| `port` | `int` | Yes | Port number |
| `protocol` | `string` | Yes | Protocol: `tcp`, `udp` |
| `probe` | `object` | No | [Health probe configuration](#probe-configuration) |

---

## Probe Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | `string` | Required | Probe type: `tcp`, `http`, `grpc`, `icmp`, `udp`, `exec` |
| `path` | `string` | `/` | HTTP probe path |
| `method` | `string` | `GET` | HTTP method |
| `status_code` | `int` | `200` | Expected HTTP status code |
| `service` | `string` | - | gRPC service name for health check |
| `command` | `string` | - | Command for exec probe |
| `args` | `list[string]` | - | Arguments for exec probe |
| `icmp_mode` | `string` | `auto` | ICMP mode: `native`, `fallback`, `auto` |
| `interval` | `duration` | `30s` | Check interval |
| `timeout` | `duration` | `5s` | Check timeout |
| `failure_threshold` | `int` | `3` | Consecutive failures before unhealthy |
| `success_threshold` | `int` | `1` | Consecutive successes before healthy |
