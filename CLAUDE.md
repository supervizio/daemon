# Daemon - Process Supervisor

Superviseur de processus PID1 en Go pour conteneurs et systèmes Unix.

## Structure du projet

```
/workspace
├── src/                      # Code source Go (obligatoire)
│   ├── cmd/daemon/           # Point d'entrée CLI
│   └── internal/             # Packages internes
│       ├── config/           # Parsing YAML et validation
│       ├── supervisor/       # Orchestration des services
│       ├── process/          # Cycle de vie des processus
│       ├── health/           # Health checks (HTTP/TCP/cmd)
│       ├── kernel/           # Abstraction OS (hexagonal)
│       └── logging/          # Rotation et capture des logs
├── examples/                 # Configurations d'exemple
├── .github/workflows/        # CI/CD (lint, test, release)
└── .devcontainer/            # Environnement de développement
```

## Stack technique

- **Langage** : Go 1.25
- **Dépendances** : gopkg.in/yaml.v3, testify
- **Architecture** : Hexagonale (ports & adapters) pour l'OS

## Règles de développement

**STRICT** : Suivre `.devcontainer/features/languages/go/RULES.md`

- Tests Go aux côtés du code (`*_test.go`)
- Linting avec `golangci-lint`
- Race detection obligatoire (`go test -race`)

## Workflow

```
/build --context    # Génère la doc contextuelle
/feature "desc"     # Branche feat/, planning, PR
/fix "desc"         # Branche fix/, planning, PR
```

## Conventions

| Type | Branche | Commit |
|------|---------|--------|
| Feature | `feat/<desc>` | `feat(scope): message` |
| Bugfix | `fix/<desc>` | `fix(scope): message` |

## Dossiers liés

| Dossier | Voir |
|---------|------|
| Code source | `src/CLAUDE.md` |
| Exemples | `examples/CLAUDE.md` |
| CI/CD | `.github/CLAUDE.md` |
| DevContainer | `.devcontainer/CLAUDE.md` |

## MCP-First

Toujours utiliser les outils MCP avant les CLI :
- `mcp__github__*` avant `gh`
- `mcp__codacy__*` avant `codacy-cli`
