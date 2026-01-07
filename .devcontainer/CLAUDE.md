# DevContainer Configuration

## Purpose

Development container setup for consistent dev environments across languages.

## Structure

```text
.devcontainer/
├── devcontainer.json    # Main config
├── docker-compose.yml   # Multi-service setup
├── Dockerfile           # Base image
├── features/            # Language & tool features
├── hooks/               # Lifecycle scripts
└── images/              # Docker images
```

## Key Files

- `devcontainer.json`: VS Code devcontainer config
- `docker-compose.yml`: Services (app, MCP servers)
- `.env`: Environment variables (git-ignored)

## Usage

Features are enabled in `devcontainer.json` under `features`.
Each language feature has its own RULES.md for conventions.
