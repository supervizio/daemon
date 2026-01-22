# DevContainer Template

## Project Structure (MANDATORY)

```
/workspace
├── src/                    # ALL source code (mandatory)
├── tests/                  # Unit tests (Go: alongside code in /src)
├── docs/                   # Documentation
└── CLAUDE.md
```

**Rules:** ALL code MUST be in `/src` regardless of language.

## Language Rules

**STRICT**: Follow `.devcontainer/features/languages/<lang>/RULES.md`

- Line 1: Required version (NEVER downgrade)
- Code style and conventions
- Testing standards

## Workflow

| Command | Action |
|---------|--------|
| `/build --context` | Generate CLAUDE.md in subdirs + fetch versions |
| `/feature <desc>` | Create `feat/<desc>` branch, plan mode, CI, PR |
| `/fix <desc>` | Create `fix/<desc>` branch, plan mode, CI, PR |

## Branch Conventions

| Type | Branch | Commit |
|------|--------|--------|
| Feature | `feat/<desc>` | `feat(scope): message` |
| Bugfix | `fix/<desc>` | `fix(scope): message` |

## SAFEGUARDS (ABSOLUTE)

**NEVER without explicit user approval:**
- Delete files in `.claude/` or `.devcontainer/`
- Modify `.claude/commands/*.md` destructively
- Remove hooks from `.devcontainer/hooks/`

## Hooks (Auto-applied)

| Hook | Action |
|------|--------|
| `pre-validate.sh` | Protect sensitive files |
| `post-edit.sh` | Format + Imports + Lint |

## Context Hierarchy

```
/CLAUDE.md        → Overview
/src/CLAUDE.md    → Details
```

**Principle:** More details deeper, <80 lines each.

## MCP-FIRST RULE

**ALWAYS use MCP tools BEFORE CLI binaries.**

| Action | MCP Tool | CLI Fallback |
|--------|----------|--------------|
| GitHub | `mcp__github__*` | `gh` |
| Codacy | `mcp__codacy__*` | `codacy-cli` |

**Rules:**
1. Check `.mcp.json` for available MCP servers
2. Use `mcp__<server>__<action>` first
3. Only fallback if MCP fails
