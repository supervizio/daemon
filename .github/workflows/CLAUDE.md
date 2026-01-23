# GitHub Workflows

## Purpose

CI/CD automation for build, test, release, and deployment.

## Structure

```text
workflows/
├── ci.yml           # Continuous integration (lint, test, build, e2e-behavioral)
├── release.yml      # Semantic release + packages + e2e-linux/bsd tests
└── deploy-repo.yml  # Package repository deployment (GitHub Pages)
```

## Key Files

| Workflow | Trigger | Action |
|----------|---------|--------|
| `ci.yml` | Push/PR | Lint, test, build, e2e-behavioral |
| `release.yml` | Main push | Packages + e2e tests before release |
| `deploy-repo.yml` | Release | Deploy apt/yum/apk repos |

## Conventions

- Use reusable workflows where possible
- Pin action versions with SHA
- Secrets via GitHub Secrets (never hardcoded)
