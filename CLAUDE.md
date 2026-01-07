# Kodflow DevContainer Template

## Project Structure (MANDATORY)

```text
/workspace
+-- src/                    # ALL source code (mandatory)
|   +-- components/
|   +-- services/
|   +-- ...
+-- tests/                  # Unit tests (optional, not for Go)
+-- docs/                   # Documentation
+-- CLAUDE.md
```

**Rules:**
- ALL code MUST be in `/src` regardless of language
- Tests in `/tests` (except Go: tests alongside code in `/src`)
- Never put code at project root

## Language Rules

**STRICT**: Follow rules in `.devcontainer/features/languages/<lang>/RULES.md`

Each RULES.md contains:
1. **Line 1**: Required version (NEVER downgrade)
2. Code style and conventions
3. Project structure requirements
4. Testing standards

## Workflow (MANDATORY)

### 1. Context Generation
```
/build --context
```
Generates CLAUDE.md in all subdirectories + fetches latest language versions.

### 2. Feature Development
```
/feature <description>
```
Creates `feat/<description>` branch, **mandatory planning mode**, CI check, PR creation (no auto-merge).

### 3. Bug Fixes
```
/fix <description>
```
Creates `fix/<description>` branch, **mandatory planning mode**, CI check, PR creation (no auto-merge).

**Flow:**
```
/build --context → /feature "..." ou /fix "..."
```

## Branch Conventions

| Type | Branch | Commit prefix |
|------|--------|---------------|
| Feature | `feat/<desc>` | `feat(scope): message` |
| Bugfix | `fix/<desc>` | `fix(scope): message` |

## Code Quality

- Latest stable version ONLY (see RULES.md)
- No deprecated APIs
- No legacy patterns
- Security-first approach
- Full test coverage

## MCP-FIRST RULE (MANDATORY)

**ALWAYS use MCP tools BEFORE falling back to CLI binaries.**

```yaml
mcp_priority:
  rule: "MCP tools are the PRIMARY interface"
  fallback: "CLI only when MCP unavailable or fails"

  workflow:
    1_check_mcp: "Verify MCP server is available in .mcp.json"
    2_use_mcp: "Call mcp__<server>__<action> tool"
    3_on_failure: "Log error, inform user, then try CLI fallback"
    4_never_ask: "NEVER ask user for tokens if MCP is configured"

  examples:
    github:
      priority: "mcp__github__list_pull_requests"
      fallback: "gh pr list"
    codacy:
      priority: "mcp__codacy__codacy_cli_analyze"
      fallback: "codacy-cli analyze"
```

**Why MCP-first:**

- MCP servers have pre-configured authentication (tokens in .mcp.json)
- CLI tools require separate auth (`gh auth login`, etc.)
- MCP provides structured responses (JSON vs text parsing)
- Single source of truth for credentials

## SAFEGUARDS (ABSOLUTE - NO BYPASS)

**NEVER without EXPLICIT user approval:**
- Delete files in `.claude/` directory
- Delete files in `.devcontainer/` directory
- Modify `.claude/commands/*.md` destructively (removing features/logic)
- Remove hooks from `.devcontainer/hooks/`

**When simplifying/refactoring:**
- Move content to separate files, NEVER delete logic
- Ask before removing any feature, even if it seems redundant

## Hooks (Auto-applied)

| Hook | Action |
|------|--------|
| `pre-validate.sh` | Protect sensitive files |
| `post-edit.sh` | Format + Imports + Lint |
| `security.sh` | Secret detection |
| `test.sh` | Run related tests |

## Context Hierarchy

```
/CLAUDE.md              # Overview (committed)
/src/CLAUDE.md          # src details (gitignored)
/src/api/CLAUDE.md      # API details (gitignored)
```

**Principle:** More details deeper in tree, under 60 lines each.
