<!-- updated: 2026-02-15T21:30:00Z -->
# DevContainer Images

## Purpose

Docker images for the devcontainer environment.

## Workflow Modes

### PLAN MODE (Analysis)

**Required phases:**
1. Analyze request
2. Research docs (WebSearch, project docs)
3. Analyze existing code (Glob/Grep/Read)
4. Cross-reference information
5. Define epics/tasks → **USER VALIDATION**

### EXECUTION MODE

Tasks must be tracked before modifying code.

## MCP Servers

**RULE:** Always check `/workspace/.mcp.json` and use MCP first.

| Action | MCP (priority) | CLI (fallback) |
|--------|----------------|----------------|
| GitHub PR | `mcp__github__create_pull_request` | `gh pr create` |
| Issues | `mcp__github__create_issue` | `gh issue create` |
| Merge | `mcp__github__merge_pull_request` | `gh pr merge` |

## ABSOLUTE SAFEGUARDS

| Action | Status |
|--------|--------|
| Auto merge | FORBIDDEN |
| Push to main/master | FORBIDDEN |
| Skip PLAN MODE | FORBIDDEN |
