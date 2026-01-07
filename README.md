# Daemon

[![CI](https://github.com/kodflow/daemon/actions/workflows/ci.yml/badge.svg)](https://github.com/kodflow/daemon/actions/workflows/ci.yml)
[![Release](https://github.com/kodflow/daemon/actions/workflows/release.yml/badge.svg)](https://github.com/kodflow/daemon/releases)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Superviseur de processus PID1 pour conteneurs et systèmes Linux/BSD. Gère plusieurs services avec health checks, politiques de redémarrage et rotation des logs.

## Fonctionnalités

- **Gestion multi-services** : Démarrage, arrêt et monitoring de multiples processus
- **Health checks** : HTTP, TCP et commande shell avec retry configurable
- **Politiques de redémarrage** : `always`, `on-failure`, `never`, `unless-stopped`
- **Backoff exponentiel** : Délais croissants entre les tentatives de redémarrage
- **Rotation des logs** : Par taille, âge et nombre de fichiers avec compression
- **Support PID 1** : Reaping des processus zombies, gestion des signaux
- **Multi-plateforme** : Linux, BSD (FreeBSD, OpenBSD, NetBSD), macOS

## Installation

### Binaires pré-compilés

Téléchargez depuis les [releases GitHub](https://github.com/kodflow/daemon/releases).

### Depuis les sources

```bash
git clone https://github.com/kodflow/daemon.git
cd daemon/src
go build -o daemon ./cmd/daemon
```

## Utilisation

```bash
# Avec fichier de configuration
daemon --config /etc/daemon/config.yaml

# Afficher la version
daemon --version
```

## Configuration

```yaml
version: "1"

logging:
  base_dir: /var/log/daemon
  defaults:
    timestamp_format: iso8601
    rotation:
      max_size: 100MB
      max_age: 7d
      max_files: 10
      compress: true

services:
  - name: webapp
    command: /usr/bin/node
    args:
      - /app/server.js
    user: www-data
    environment:
      NODE_ENV: production
    restart:
      policy: always
      max_retries: 5
      delay: 5s
      delay_max: 5m
    health_checks:
      - type: http
        endpoint: http://localhost:3000/health
        interval: 30s
        timeout: 5s
        retries: 3
    logging:
      stdout:
        file: webapp.out.log
      stderr:
        file: webapp.err.log
```

Voir [examples/config.yaml](examples/config.yaml) pour une configuration complète.

## Signaux

| Signal | Action |
|--------|--------|
| `SIGTERM` / `SIGINT` | Arrêt gracieux de tous les services |
| `SIGHUP` | Rechargement de la configuration |
| `SIGCHLD` | Reaping des processus zombies (PID 1) |

## Architecture

```
src/
├── cmd/daemon/          # Point d'entrée
└── internal/
    ├── config/          # Parsing et validation YAML
    ├── supervisor/      # Orchestration des services
    ├── process/         # Gestion du cycle de vie
    ├── health/          # Health checks (HTTP, TCP, cmd)
    ├── kernel/          # Abstraction OS (ports & adapters)
    │   ├── ports/       # Interfaces
    │   └── adapters/    # Implémentations par plateforme
    └── logging/         # Rotation et capture des logs
```

## Développement

### Prérequis

- Go 1.25+
- golangci-lint

### Build

```bash
cd src
go build ./...
```

### Tests

```bash
cd src
go test -race -cover ./...
```

### Lint

```bash
cd src
golangci-lint run
```

## Plateformes supportées

| OS | Architectures |
|----|---------------|
| Linux | amd64, arm64, 386, armv7 |
| FreeBSD | amd64, arm64 |
| OpenBSD | amd64, arm64 |
| NetBSD | amd64, arm64 |
| DragonFlyBSD | amd64 |
| macOS | amd64, arm64 |

## License

MIT
