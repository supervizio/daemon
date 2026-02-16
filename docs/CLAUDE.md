<!-- updated: 2026-02-15T21:30:00Z -->
# Documentation - MkDocs Material

MkDocs Material documentation source for supervizio.

## Structure

```
docs/
├── index.md                # Home page
├── changelog.md            # Release history
├── METRICS_CONFIGURATION.md # Metrics config reference
├── api/                    # gRPC API reference
│   ├── daemon-service.md
│   └── metrics-service.md
├── architecture/           # System architecture
│   ├── hexagonal.md
│   └── data-flow.md
├── components/             # Component documentation
│   ├── supervisor.md, lifecycle.md, health.md
│   ├── metrics.md, discovery.md, tui.md
├── configuration/          # Configuration reference
│   ├── services.md
│   └── monitoring.md
├── deployment/             # Deployment guides
│   ├── systemd.md, container.md, platforms.md
├── examples/               # Config examples
├── guides/                 # Getting started, development
├── reference/              # CLI and proto reference
└── stylesheets/            # Custom CSS (extra.css)
```

## Build

```bash
mkdocs serve        # Local preview
mkdocs build        # Build static site (→ site/)
```

## Conventions

- MkDocs Material theme
- Markdown with admonitions (`!!! note`, `!!! warning`)
- Auto-generated navigation from directory structure
- Custom CSS in `stylesheets/extra.css`

## Related

| Directory | See |
|-----------|-----|
| `api/proto/` | Canonical protobuf definitions |
| `examples/` | YAML configuration examples |
