---
name: plan
description: |
  Enter Claude Code planning mode with RLM decomposition.
  Analyzes codebase, designs approach, creates step-by-step plan.
  Use when: starting a new feature, refactoring, or complex task.
allowed-tools:
  - "Read(**/*)"
  - "Glob(**/*)"
  - "Grep(**/*)"
  - "mcp__grepai__*"
  - "mcp__context7__*"
  - "Task(*)"
  - "WebFetch(*)"
  - "WebSearch(*)"
  - "mcp__github__*"
  - "mcp__playwright__*"
---

# /plan - Claude Code Planning Mode (RLM Architecture)

$ARGUMENTS

---

## Overview

Mode planning avec patterns **RLM** :

- **Peek** - Scan rapide du codebase
- **Decompose** - Diviser en sous-tâches
- **Parallelize** - Exploration multi-domaine
- **Synthesize** - Plan structuré

**Principe** : Planifier → Valider → Implémenter (jamais l'inverse)

---

## Arguments

| Pattern | Action |
|---------|--------|
| `<description>` | Planifie l'implémentation de la feature/fix |
| `--context` | Charge le .context.md généré par /search |
| `--help` | Affiche l'aide |

---

## --help

```
═══════════════════════════════════════════════════════════════
  /plan - Claude Code Planning Mode (RLM)
═══════════════════════════════════════════════════════════════

Usage: /plan <description> [options]

Options:
  <description>     Ce qu'il faut implémenter
  --context         Utilise .context.md comme base
  --help            Affiche cette aide

RLM Patterns:
  1. Peek       - Scan rapide codebase
  2. Decompose  - Diviser en sous-tâches
  3. Parallelize - Exploration parallèle
  4. Synthesize - Plan structuré

Workflow:
  /search <topic> → /plan <feature> → (approve) → /do

Exemples:
  /plan "Add user authentication with JWT"
  /plan "Refactor database layer" --context
  /plan "Fix memory leak in worker process"

═══════════════════════════════════════════════════════════════
```

---

## Phase 1 : Peek (RLM Pattern)

**Scan rapide AVANT exploration approfondie :**

```yaml
peek_workflow:
  1_context_check:
    action: "Vérifier si .context.md existe"
    tool: [Read]
    output: "context_available"

  2_structure_scan:
    action: "Scanner la structure du projet"
    tools: [Glob]
    patterns:
      - "src/**/*"
      - "tests/**/*"
      - "package.json | go.mod | Cargo.toml"

  3_pattern_grep:
    action: "Identifier les patterns pertinents"
    tools: [Grep]
    searches:
      - Keywords from description
      - Related function names
      - Existing patterns
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════════════════════
  /plan - Peek Analysis
═══════════════════════════════════════════════════════════════

  Description: "Add user authentication with JWT"

  Context:
    ✓ .context.md loaded (from /search)
    ✓ 47 source files scanned
    ✓ 23 test files found

  Patterns identified:
    - Existing auth: src/middleware/auth.ts
    - User model: src/models/user.ts
    - Routes: src/routes/*.ts

  Keywords matched: 15 occurrences

═══════════════════════════════════════════════════════════════
```

---

## Phase 2 : Decompose (RLM Pattern)

**Diviser la tâche en sous-tâches :**

```yaml
decompose_workflow:
  1_analyze_description:
    action: "Extraire les objectifs"
    example:
      description: "Add user authentication with JWT"
      objectives:
        - "Setup JWT utilities"
        - "Create auth middleware"
        - "Add login/logout endpoints"
        - "Protect existing routes"
        - "Add tests"

  2_identify_domains:
    action: "Catégoriser par domaine"
    domains:
      - backend: "API, middleware, database"
      - frontend: "UI components, state"
      - infrastructure: "config, deployment"
      - testing: "unit, integration, e2e"

  3_order_dependencies:
    action: "Ordonner par dépendance"
    output: "ordered_tasks[]"
```

---

## Phase 3 : Parallelize (RLM Pattern)

**Exploration multi-domaine en parallèle :**

```yaml
parallel_exploration:
  mode: "PARALLEL (single message, multiple Task calls)"

  agents:
    - task: "backend-explorer"
      type: "Explore"
      prompt: |
        Analyze backend for: {description}
        Find: related files, existing patterns, dependencies
        Return: {files[], patterns[], recommendations[]}

    - task: "frontend-explorer"
      type: "Explore"
      prompt: |
        Analyze frontend for: {description}
        Find: components, state, API calls
        Return: {files[], components[], state_management}

    - task: "test-explorer"
      type: "Explore"
      prompt: |
        Analyze tests for: {description}
        Find: existing coverage, test patterns
        Return: {coverage, patterns[], gaps[]}

    - task: "patterns-consultant"
      type: "Explore"
      prompt: |
        Consult .claude/docs/ for: {description}
        Find: applicable design patterns
        Return: {patterns[], references[]}
```

**IMPORTANT** : Lancer TOUS les agents dans UN SEUL message.

---

## Phase 3.5 : Pattern Consultation (OBLIGATOIRE)

**Consulter `.claude/docs/` pour les patterns :**

```yaml
pattern_consultation:
  1_identify_category:
    mapping:
      - "Création d'objets?" → creational/README.md
      - "Performance/Cache?" → performance/README.md
      - "Concurrence?" → concurrency/README.md
      - "Architecture?" → architectural/*.md
      - "Intégration?" → messaging/README.md
      - "Sécurité?" → security/README.md

  2_read_patterns:
    action: "Read(.claude/docs/<category>/README.md)"
    output: "2-3 patterns applicables"

  3_integrate:
    action: "Ajouter au plan avec justification"
```

**Output :**

```
═══════════════════════════════════════════════════════════════
  Pattern Analysis
═══════════════════════════════════════════════════════════════

  Patterns identifiés:
    ✓ Repository (DDD) - Pour accès données user
    ✓ Factory (Creational) - Pour création tokens
    ✓ Middleware (Enterprise) - Pour auth chain

  Références consultées:
    → .claude/docs/ddd/README.md
    → .claude/docs/creational/README.md
    → .claude/docs/enterprise/README.md

═══════════════════════════════════════════════════════════════
```

---

## Phase 4 : Synthesize (RLM Pattern)

**Générer le plan structuré :**

```yaml
synthesize_workflow:
  1_collect:
    action: "Rassembler résultats des agents"

  2_consolidate:
    action: "Fusionner en plan cohérent"

  3_generate:
    format: "Structured plan document"
```

**Plan Output Format :**

```markdown
# Implementation Plan: <description>

## Overview
<2-3 phrases résumant l'approche>

## Design Patterns Applied

| Pattern | Category | Justification | Reference |
|---------|----------|---------------|-----------|
| Repository | DDD | Data access abstraction | .claude/docs/ddd/README.md |
| Factory | Creational | Token creation | .claude/docs/creational/README.md |

## Prerequisites
- [ ] <Dépendance ou setup requis>
- [ ] <Autre prérequis>

## Implementation Steps

### Step 1: <Titre>
**Files:** `src/file1.ts`, `src/file2.ts`
**Actions:**
1. <Action spécifique>
2. <Action spécifique>

**Code pattern:**
```<lang>
// Example of what will be implemented
```

### Step 2: <Titre>
...

## Testing Strategy
- [ ] Unit tests for `component`
- [ ] Integration test for `flow`

## Rollback Plan
Comment annuler si problème

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Risk description | Solution |
```

---

## Phase 5 : Demande de Validation

**OBLIGATOIRE : Attendre approbation utilisateur**

```
═══════════════════════════════════════════════════════════════
  Plan ready for review
═══════════════════════════════════════════════════════════════

  Summary:
    • 4 implementation steps
    • 6 files to modify
    • 2 new files to create
    • 8 tests to add

  Design Patterns:
    • Repository (DDD)
    • Factory (Creational)

  Estimated complexity: MEDIUM

  Actions:
    → Review the plan above
    → Run /do to execute (auto-detects plan)
    → Or modify the plan manually

═══════════════════════════════════════════════════════════════
```

---

## Intégration avec autres skills

| Avant /plan | Après /plan |
|-------------|-------------|
| `/search <topic>` | `/do` |
| Génère .context.md | Exécute le plan (auto-détecté) |

**Workflow complet :**

```
/search "JWT authentication best practices"
    ↓
.context.md généré
    ↓
/plan "Add JWT auth to API" --context
    ↓
Plan créé et affiché
    ↓
User: "OK, go ahead"
    ↓
/do                          # Détecte le plan automatiquement
    ↓
Implémentation exécutée
```

**Note** : `/do` détecte automatiquement le plan approuvé et l'exécute
sans poser les questions interactives.

---

## GARDE-FOUS (ABSOLUS)

| Action | Status |
|--------|--------|
| Skip Phase 1 (Peek) | ❌ **INTERDIT** |
| Exploration séquentielle | ❌ **INTERDIT** |
| Skip Pattern Consultation | ❌ **INTERDIT** |
| Implémenter sans plan approuvé | ❌ **INTERDIT** |
| Plan sans steps concrets | ❌ **INTERDIT** |
| Plan sans rollback strategy | ⚠ **WARNING** |
