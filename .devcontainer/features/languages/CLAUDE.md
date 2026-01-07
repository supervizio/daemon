# Language Features

## Purpose

Language-specific installation and conventions.

## Available Languages

| Language | Version | Status |
|----------|---------|--------|
| Go | >= 1.25.0 | Stable |
| Node.js | >= 25.0.0 | Current |
| Python | >= 3.14.0 | Stable |
| Rust | >= 1.92.0 | Stable |
| Elixir | >= 1.19.0 | Stable |
| Java | >= 25 | LTS |
| PHP | >= 8.5.0 | Stable |
| Ruby | >= 3.4.0 | Stable |
| Scala | >= 3.7.0 | Stable |
| Dart/Flutter | >= 3.10/3.38 | Stable |
| C++ | >= C++23 | Standard |
| Carbon | >= 0.1.0 | Experimental |

## Per-Language Structure

```text
<language>/
├── install.sh    # Installation script
└── RULES.md      # Coding conventions
```

## Conventions

- RULES.md line 1: Minimum version requirement
- All code in /src regardless of language
- Tests in /tests (except Go: alongside code)
