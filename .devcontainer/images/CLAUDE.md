# DevContainer Workflow

## Modes de Travail

### PLAN MODE (Analyse)

**Phases obligatoires:**
1. Analyse demande → Comprendre l'objectif
2. Recherche docs → WebSearch, docs projet
3. Analyse existant → Glob/Grep/Read
4. Affûtage → Croiser les infos
5. Définition épics/tasks → **VALIDATION USER**
6. Écriture Taskwarrior

**INTERDIT:** Write/Edit sur fichiers code (✅ `/plans/` uniquement)

### BYPASS MODE (Exécution)

**Workflow par task:**
```bash
task-start.sh <uuid>   # TODO → WIP
# Execute task
task-done.sh <uuid>    # WIP → DONE
```

**OBLIGATOIRE:** Une task DOIT être WIP avant Write/Edit

## Commandes Taskwarrior

| Script | Usage |
|--------|-------|
| `task-init.sh <type> <desc>` | Init projet |
| `task-epic.sh <project> <num> <name>` | Créer epic |
| `task-add.sh <project> <epic> <uuid> <name> [parallel] [ctx]` | Ajouter task |
| `task-start.sh <uuid>` | TODO → WIP |
| `task-done.sh <uuid>` | WIP → DONE |

Chemin: `/home/vscode/.claude/scripts/`

## Session JSON

| Élément | Convention |
|---------|------------|
| Dossier | `~/.claude/sessions/` |
| Fichier | `<project>.json` (slug branche sans préfixe) |

**States:** `planning` → `planned` → `applying` → `applied`

## MCP Servers

**RÈGLE:** Toujours vérifier `/workspace/.mcp.json` et utiliser MCP en priorité.

| Action | MCP (priorité) | CLI (fallback) |
|--------|----------------|----------------|
| GitHub PR | `mcp__github__create_pull_request` | `gh pr create` |
| Issues | `mcp__github__create_issue` | `gh issue create` |
| Merge | `mcp__github__merge_pull_request` | `gh pr merge` |

## GARDE-FOUS ABSOLUS

| Action | Status |
|--------|--------|
| Merge automatique | ❌ INTERDIT |
| Push sur main/master | ❌ INTERDIT |
| Skip PLAN MODE | ❌ INTERDIT |
| Write/Edit sans task WIP | ❌ BLOQUÉ |
