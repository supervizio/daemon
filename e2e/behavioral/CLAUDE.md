# E2E Behavioral Tests

Tests comportementaux validant le runtime de supervizio.

## Structure

```
behavioral/
├── crasher/              # Binaire de test contrôlable
│   └── main.go           # ~140 lignes
├── Dockerfile.behavioral # Image Docker avec supervizio + crasher
├── go.mod                # Module Go pour testcontainers
├── helpers_test.go       # Fonctions utilitaires partagées
├── restart_test.go       # Tests restart policies
├── backoff_test.go       # Tests exponential backoff
├── health_test.go        # Tests health probes
├── pid1_test.go          # Tests PID1 container
└── testdata/             # Configs YAML de test
    ├── restart-always.yaml
    ├── restart-on-failure.yaml
    ├── restart-on-failure-exit0.yaml
    ├── restart-never.yaml
    ├── restart-unless-stopped.yaml
    ├── backoff.yaml
    ├── health-http.yaml
    ├── health-tcp.yaml
    ├── orphan-spawner.yaml
    ├── long-running.yaml
    └── ignore-term.yaml
```

## Crasher Binary

Binaire Go contrôlable pour simuler différents comportements de processus.

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--exit` | int | 0 | Exit code à retourner |
| `--delay` | duration | 0 | Délai avant exit |
| `--port` | int | 0 | Port TCP/HTTP à écouter |
| `--http` | bool | false | Servir /health en HTTP |
| `--orphan` | bool | false | Spawn processus orphelin |
| `--ignore-term` | bool | false | Ignorer SIGTERM |
| `--log` | string | "" | Fichier log (stdout si vide) |
| `--crash-after` | int | 0 | Crash après N secondes |
| `--healthy` | bool | true | /health retourne 200 (false = 503) |

### Exemples

```bash
# Exit immédiatement avec code 0
crasher --exit=0

# Exit après 5s avec code 1 (failure)
crasher --delay=5s --exit=1

# HTTP server avec health endpoint
crasher --http --port=8080 --delay=1h

# Ignore SIGTERM (teste SIGKILL fallback)
crasher --ignore-term --delay=1h

# Spawn orphan process
crasher --orphan --delay=1s --exit=0
```

## Exécution

### Prérequis

1. Build des binaires :
```bash
# Build supervizio
cd src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/supervizio ./cmd/daemon

# Build crasher
cd ../e2e/behavioral/crasher && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../../../bin/crasher .
```

2. Build image Docker :
```bash
docker build -f e2e/behavioral/Dockerfile.behavioral -t supervizio-behavioral:test .
```

### Lancer les tests

```bash
# Tous les tests
cd e2e/behavioral && go test -v ./...

# Avec race detection
go test -v -race ./...

# Tests spécifiques
go test -v -run TestRestartPolicy ./...
go test -v -run TestZombieReaping ./...

# Skip tests longs
go test -v -short ./...
```

## Tests Couverts

### Restart Policies (restart_test.go)

| Test | Description |
|------|-------------|
| `TestRestartPolicyAlways` | Restart sur exit code 0 |
| `TestRestartPolicyOnFailureWithFailure` | Restart sur exit code 1 |
| `TestRestartPolicyOnFailureWithSuccess` | Pas de restart sur exit code 0 |
| `TestRestartPolicyNever` | Jamais de restart |
| `TestRestartPolicyUnlessStopped` | Restart sauf si stopped |
| `TestMaxRetriesRespected` | Limite max_retries respectée |

### Exponential Backoff (backoff_test.go)

| Test | Description |
|------|-------------|
| `TestExponentialBackoff` | Délai augmente entre restarts |
| `TestBackoffRespectsMaxDelay` | Délai ne dépasse pas delay_max |

### Health Probes (health_test.go)

| Test | Description |
|------|-------------|
| `TestHTTPHealthProbe` | Probe HTTP /health fonctionne |
| `TestTCPHealthProbe` | Probe TCP fonctionne |
| `TestHealthProbeSuccessMarksHealthy` | Service marqué healthy |

### PID1 Capabilities (pid1_test.go)

| Test | Description |
|------|-------------|
| `TestSupervizioRunsAsPID1` | supervizio est PID 1 |
| `TestZombieReaping` | Pas de zombies accumulés |
| `TestNoZombiesAfterMultipleRestarts` | Zombies nettoyés après restarts |
| `TestSignalForwardingToServices` | SIGTERM forwardé aux services |
| `TestGracefulShutdownForwardsSignals` | Shutdown gracieux |
| `TestSIGKILLFallbackForUnresponsiveService` | SIGKILL si SIGTERM ignoré |
| `TestOrphanAdoption` | Orphelins adoptés par PID 1 |

## CI Integration

Le job `e2e-behavioral` dans `.github/workflows/e2e.yml` exécute ces tests automatiquement.

## Dépendances

- testcontainers-go v0.37.0
- stretchr/testify v1.10.0
- Docker (pour les tests)
