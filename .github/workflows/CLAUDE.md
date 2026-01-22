# GitHub Workflows

## Purpose

CI/CD automation for build, test, release, and deployment.

## Structure

```text
workflows/
├── ci.yml           # Continuous integration (lint, test, build)
├── e2e.yml          # End-to-end tests (Vagrant/Docker)
├── release.yml      # Semantic release + package generation
└── deploy-repo.yml  # Package repository deployment (GitHub Pages)
```

## Key Files

| Workflow | Trigger | Action |
|----------|---------|--------|
| `ci.yml` | Push/PR | Lint, test, build binaries |
| `e2e.yml` | Manual/Schedule | Full E2E test matrix |
| `release.yml` | Tag push | Create release + packages |
| `deploy-repo.yml` | Release | Deploy apt/yum/apk repos |

## Conventions

- Use reusable workflows where possible
- Pin action versions with SHA
- Secrets via GitHub Secrets (never hardcoded)
