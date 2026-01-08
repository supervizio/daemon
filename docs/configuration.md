# Configuration Reference

Complete YAML configuration reference for superviz.io.

## Configuration File Structure

```yaml
# supervizio.yaml - Complete example
version: "1.0"

# Global settings
settings:
  log_level: info              # debug, info, warn, error
  shutdown_timeout: 30s        # Grace period for shutdown
  pid_file: /var/run/supervizio.pid

# Logging defaults
logging:
  defaults:
    directory: /var/log/supervizio
    rotation:
      max_size: 100MB
      max_age: 7d
      max_files: 10
      compress: true

# Service definitions
services:
  - name: api
    command: /app/api-server
    # ... service configuration
```

---

## Services Configuration

### Basic Service

```yaml
services:
  - name: myservice           # Required: unique identifier
    command: /path/to/binary  # Required: executable path
    args:                     # Optional: command arguments
      - --port
      - "8080"
    working_dir: /app         # Optional: working directory
    enabled: true             # Optional: enable/disable (default: true)
```

### Environment Variables

```yaml
services:
  - name: api
    command: /app/server
    environment:
      NODE_ENV: production
      DATABASE_URL: postgres://localhost/db
      LOG_LEVEL: info
    env_file:                 # Load from files
      - /etc/app/defaults.env
      - /etc/app/secrets.env
```

### User and Group

```yaml
services:
  - name: webapp
    command: /app/server
    user: appuser             # Run as user (name or UID)
    group: appgroup           # Run as group (name or GID)
```

---

## Restart Configuration

### Restart Policies

```
┌──────────────────────────────────────────────────────────────────┐
│                       RESTART POLICIES                            │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  always         Restart regardless of exit code                  │
│  on-failure     Restart only if exit code != 0                   │
│  never          Never restart automatically                       │
│  unless-stopped Restart unless manually stopped                   │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

### Complete Restart Configuration

```yaml
services:
  - name: worker
    command: /app/worker
    restart:
      policy: on-failure      # Restart policy
      max_retries: 5          # Maximum restart attempts (0 = unlimited)
      delay: 5s               # Initial delay before restart
      delay_max: 5m           # Maximum delay (cap for backoff)
      multiplier: 2           # Exponential backoff multiplier
```

### Backoff Calculation

```
delay_n = min(delay * (multiplier ^ n), delay_max)

Example with delay=5s, multiplier=2, delay_max=5m:
  Attempt 1: 5s
  Attempt 2: 10s
  Attempt 3: 20s
  Attempt 4: 40s
  Attempt 5: 80s
  Attempt 6: 160s
  Attempt 7: 300s (capped at max)
```

---

## Health Checks

### HTTP Health Check

```yaml
services:
  - name: api
    command: /app/server
    health_checks:
      - type: http
        endpoint: http://localhost:8080/health
        method: GET                    # GET, POST, HEAD
        headers:                       # Optional headers
          Authorization: Bearer token
          Content-Type: application/json
        expected_status: 200           # Expected HTTP status code
        expected_body: "ok"            # Optional: expected response body
        interval: 30s                  # Check interval
        timeout: 5s                    # Request timeout
        retries: 3                     # Failures before unhealthy
        start_period: 10s              # Grace period after start
```

### TCP Health Check

```yaml
services:
  - name: database
    command: /usr/bin/postgres
    health_checks:
      - type: tcp
        host: localhost                # Target host
        port: 5432                     # Target port
        interval: 30s
        timeout: 5s
        retries: 3
```

### Command Health Check

```yaml
services:
  - name: redis
    command: /usr/bin/redis-server
    health_checks:
      - type: command
        command: /usr/bin/redis-cli
        args:
          - ping
        interval: 30s
        timeout: 10s
        retries: 3
```

### Multiple Health Checks

```yaml
services:
  - name: api
    command: /app/server
    health_checks:
      # Primary: HTTP endpoint
      - type: http
        endpoint: http://localhost:8080/health
        interval: 30s
        timeout: 5s
        retries: 3

      # Secondary: Database connectivity
      - type: tcp
        host: localhost
        port: 5432
        interval: 60s
        timeout: 5s
        retries: 2
```

---

## Logging Configuration

### Service-Level Logging

```yaml
services:
  - name: api
    command: /app/server
    logging:
      stdout:
        file: /var/log/api/stdout.log
        rotation:
          max_size: 50MB
          max_files: 5
          compress: true
      stderr:
        file: /var/log/api/stderr.log
        rotation:
          max_size: 50MB
          max_files: 5
          compress: true
```

### Global Logging Defaults

```yaml
logging:
  defaults:
    directory: /var/log/supervizio
    rotation:
      max_size: 100MB          # Rotate when file exceeds size
      max_age: 7d              # Delete files older than this
      max_files: 10            # Keep at most N files
      compress: true           # Gzip rotated files
    timestamp: true            # Add timestamps to lines
    passthrough: false         # Also write to supervisor stdout/stderr
```

### Rotation Options

```
┌─────────────────────────────────────────────────────────────────┐
│                      ROTATION OPTIONS                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  max_size     Size trigger for rotation                         │
│               Formats: 10KB, 100MB, 1GB                         │
│                                                                  │
│  max_age      Age trigger for deletion                          │
│               Formats: 1h, 24h, 7d, 30d                         │
│                                                                  │
│  max_files    Maximum number of files to keep                   │
│               Oldest files deleted when exceeded                 │
│                                                                  │
│  compress     Gzip old files (saves ~90% space)                 │
│               Files become: app.log.1.gz                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Service Dependencies

```yaml
services:
  - name: database
    command: /usr/bin/postgres
    health_checks:
      - type: tcp
        port: 5432
        interval: 5s

  - name: api
    command: /app/server
    depends_on:
      - database              # Wait for database to be healthy

  - name: worker
    command: /app/worker
    depends_on:
      - database
      - api                   # Wait for both
```

### Dependency Resolution

```
┌─────────────────────────────────────────────────────────────────┐
│                   DEPENDENCY RESOLUTION                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Start Order:                                                    │
│                                                                  │
│    1. Services with no dependencies start first                 │
│    2. Services wait for dependencies to be "healthy"            │
│    3. If dependency fails, dependent services don't start       │
│                                                                  │
│  Example:                                                        │
│                                                                  │
│    database ──┬──► api ──┬──► worker                            │
│               │          │                                       │
│               └──────────┘                                       │
│                                                                  │
│  Shutdown Order:                                                 │
│                                                                  │
│    Reverse of start: worker → api → database                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Signals Configuration

### Stop Signals

```yaml
services:
  - name: nginx
    command: /usr/sbin/nginx
    stop_signal: SIGQUIT       # Signal for graceful stop (default: SIGTERM)
    stop_timeout: 30s          # Time before SIGKILL (default: 10s)
```

### Available Signals

| Signal | Value | Description |
|--------|-------|-------------|
| SIGTERM | 15 | Graceful termination (default) |
| SIGINT | 2 | Interrupt |
| SIGQUIT | 3 | Quit with core dump |
| SIGHUP | 1 | Hangup / reload |
| SIGUSR1 | 10 | User-defined 1 |
| SIGUSR2 | 12 | User-defined 2 |

---

## Complete Example

```yaml
version: "1.0"

settings:
  log_level: info
  shutdown_timeout: 60s

logging:
  defaults:
    directory: /var/log/myapp
    rotation:
      max_size: 100MB
      max_files: 10
      compress: true

services:
  # PostgreSQL Database
  - name: postgres
    command: /usr/lib/postgresql/15/bin/postgres
    args:
      - -D
      - /var/lib/postgresql/data
    user: postgres
    group: postgres
    working_dir: /var/lib/postgresql
    restart:
      policy: always
      delay: 5s
    health_checks:
      - type: tcp
        port: 5432
        interval: 10s
        timeout: 5s
        retries: 3

  # Redis Cache
  - name: redis
    command: /usr/bin/redis-server
    args:
      - /etc/redis/redis.conf
    user: redis
    restart:
      policy: always
    health_checks:
      - type: command
        command: /usr/bin/redis-cli
        args: [ping]
        interval: 10s
        timeout: 5s

  # API Server
  - name: api
    command: /app/api-server
    working_dir: /app
    user: app
    environment:
      DATABASE_URL: postgres://localhost/myapp
      REDIS_URL: redis://localhost:6379
      PORT: "8080"
    depends_on:
      - postgres
      - redis
    restart:
      policy: on-failure
      max_retries: 5
      delay: 5s
      delay_max: 2m
    health_checks:
      - type: http
        endpoint: http://localhost:8080/health
        interval: 30s
        timeout: 5s
        retries: 3
        start_period: 10s
    logging:
      stdout:
        file: api.log
      stderr:
        file: api-error.log

  # Background Worker
  - name: worker
    command: /app/worker
    working_dir: /app
    user: app
    environment:
      DATABASE_URL: postgres://localhost/myapp
      REDIS_URL: redis://localhost:6379
    depends_on:
      - postgres
      - redis
      - api
    restart:
      policy: always
      delay: 10s
    logging:
      stdout:
        file: worker.log
      stderr:
        file: worker-error.log
```

---

## Size and Duration Formats

### Size Format

```
┌───────────────────────────────────────┐
│          SIZE FORMATS                  │
├───────────────────────────────────────┤
│  10      → 10 bytes                   │
│  10KB    → 10 * 1024 bytes            │
│  10MB    → 10 * 1024² bytes           │
│  10GB    → 10 * 1024³ bytes           │
└───────────────────────────────────────┘
```

### Duration Format

```
┌───────────────────────────────────────┐
│        DURATION FORMATS                │
├───────────────────────────────────────┤
│  100ms   → 100 milliseconds           │
│  5s      → 5 seconds                  │
│  30s     → 30 seconds                 │
│  5m      → 5 minutes                  │
│  1h      → 1 hour                     │
│  24h     → 24 hours                   │
│  7d      → 7 days                     │
└───────────────────────────────────────┘
```

---

## Environment Variable Expansion

```yaml
services:
  - name: api
    command: /app/server
    environment:
      # Direct value
      PORT: "8080"

      # From shell environment
      DATABASE_URL: ${DATABASE_URL}

      # With default value
      LOG_LEVEL: ${LOG_LEVEL:-info}

      # Required (fails if not set)
      API_KEY: ${API_KEY:?API_KEY is required}
```
