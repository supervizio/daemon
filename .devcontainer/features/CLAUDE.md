# DevContainer Features

## Purpose

Modular features for languages, tools, and architectures.

## Structure

```text
features/
├── languages/      # Language-specific (go, nodejs, python, etc.)
├── architectures/  # Architecture patterns
└── claude/         # Claude Code integration
```

## Key Components

- Each language has `install.sh` + `RULES.md`
- RULES.md line 1 = minimum version (NEVER downgrade)
- install.sh runs on devcontainer build

## Adding a Language

1. Create `languages/<name>/`
2. Add `install.sh` for installation
3. Add `RULES.md` with conventions
