---
name: git
description: |
  Workflow Git Automation with RLM decomposition.
  Handles branch management, conventional commits, and CI validation.
  Use when: committing changes, creating PRs, or merging with CI checks.
allowed-tools:
  - "Bash(git:*)"
  - "Bash(gh:*)"
  - "mcp__github__*"
  - "Read(**/*)"
  - "Glob(**/*)"
  - "Grep(**/*)"
  - "mcp__grepai__*"
  - "Task(*)"
---

# /git - Workflow Git Automation (RLM Architecture)

$ARGUMENTS

---

## Overview

Automatisation Git avec patterns **RLM** :

- **Peek** - Analyser l'état git avant action
- **Decompose** - Identifier les fichiers par catégorie
- **Parallelize** - Checks en parallèle (lint, test, CI)
- **Synthesize** - Rapport consolidé

---

## Arguments

| Pattern | Action |
|---------|--------|
| `--commit` | Workflow complet : branch, commit, push, PR |
| `--merge` | Merge la PR avec CI validation |
| `--help` | Affiche l'aide |

### Options --commit

| Option | Action |
|--------|--------|
| `--branch <nom>` | Force le nom de branche |
| `--no-pr` | Skip la création de PR |
| `--amend` | Amend le dernier commit |

### Options --merge

| Option | Action |
|--------|--------|
| `--pr <number>` | Merge une PR spécifique |
| `--strategy <type>` | Méthode: merge/squash/rebase (défaut: squash) |
| `--dry-run` | Vérifier sans merger |

---

## --help

```
═══════════════════════════════════════════════════════════════
  /git - Workflow Git Automation (RLM)
═══════════════════════════════════════════════════════════════

Usage: /git <action> [options]

Actions:
  --commit          Workflow complet (branch, commit, push, PR)
  --merge           Merge avec CI validation et auto-fix

RLM Patterns:
  1. Peek       - Analyser état git
  2. Decompose  - Catégoriser fichiers
  3. Parallelize - Checks simultanés
  4. Synthesize - Rapport consolidé

Options --commit:
  --branch <nom>    Force le nom de branche
  --no-pr           Skip la création de PR
  --amend           Amend le dernier commit

Options --merge:
  --pr <number>     Merge une PR spécifique
  --strategy <type> Méthode: merge/squash/rebase (défaut: squash)
  --dry-run         Vérifier sans merger

Exemples:
  /git --commit                 Commit + PR automatique
  /git --commit --no-pr         Commit sans créer de PR
  /git --merge                  Merge la PR courante
  /git --merge --pr 42          Merge la PR #42

═══════════════════════════════════════════════════════════════
```

---

## Priorité MCP vs CLI

**IMPORTANT** : Toujours privilégier les outils MCP GitHub quand disponibles.

| Action | Priorité 1 (MCP) | Fallback (CLI) |
|--------|------------------|----------------|
| Créer branche | `mcp__github__create_branch` | `git checkout -b` |
| Créer PR | `mcp__github__create_pull_request` | `gh pr create` |
| Lister PRs | `mcp__github__list_pull_requests` | `gh pr list` |
| Voir PR | `mcp__github__get_pull_request` | `gh pr view` |
| Status CI | `mcp__github__get_pull_request_status` | `gh pr checks` |
| Merger PR | `mcp__github__merge_pull_request` | `gh pr merge` |

---

## Action: --commit

### Phase 1 : Peek (RLM Pattern)

**Analyser l'état git AVANT toute action :**

```yaml
peek_workflow:
  1_status:
    action: "Vérifier l'état du repo"
    commands:
      - "git status --porcelain"
      - "git branch --show-current"
      - "git log -1 --format='%h %s'"

  2_changes:
    action: "Analyser les changements"
    tools: [Bash(git diff --stat)]

  3_branch_check:
    action: "Vérifier la branche courante"
    decision:
      - "main/master → MUST create new branch"
      - "feat/* | fix/* → Check coherence"
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Peek Analysis
═══════════════════════════════════════════════════════════════

  Branch: main (protected)
  Status: 5 files modified, 2 untracked

  Changes detected:
    ├─ src/auth/login.ts (+45, -12)
    ├─ src/auth/logout.ts (+23, -5)
    ├─ tests/auth.test.ts (+80, -0) [new]
    ├─ package.json (+2, -1)
    └─ README.md (+15, -3)

  Decision: CREATE new branch (on protected main)

═══════════════════════════════════════════════════════════════
```

---

### Phase 2 : Decompose (RLM Pattern)

**Catégoriser les fichiers modifiés :**

```yaml
decompose_workflow:
  categories:
    features:
      patterns: ["src/**/*.ts", "src/**/*.js"]
      prefix: "feat"

    fixes:
      patterns: ["*fix*", "*bug*"]
      prefix: "fix"

    tests:
      patterns: ["tests/**", "**/*.test.*"]
      prefix: "test"

    docs:
      patterns: ["*.md", "docs/**"]
      prefix: "docs"

    config:
      patterns: ["*.json", "*.yaml", "*.toml"]
      prefix: "chore"

  auto_detect:
    action: "Déduire le type dominant"
    output: "commit_type, scope, branch_name"
```

---

### Phase 3 : Parallelize (RLM Pattern)

**Pré-commit checks en parallèle :**

```yaml
parallel_checks:
  mode: "PARALLEL (single message, multiple calls)"

  agents:
    - task: "lint-check"
      action: "Run linter on modified files"
      command: "{lint_command}"

    - task: "test-check"
      action: "Run related tests"
      command: "{test_command} --findRelatedTests"

    - task: "build-check"
      action: "Verify build works"
      command: "{build_command}"
```

**IMPORTANT** : Lancer les 3 checks dans UN SEUL message.

---

### Phase 4 : Execute & Synthesize

```yaml
execute_workflow:
  1_branch:
    action: "Créer ou utiliser branche"
    auto: true

  2_stage:
    action: "Stage tous les fichiers"
    command: "git add -A"

  3_commit:
    action: "Créer le commit"
    format: |
      <type>(<scope>): <description>

      [body optionnel]

  4_push:
    action: "Push vers origin"
    command: "git push -u origin <branch>"

  5_pr:
    action: "Créer la PR"
    tool: mcp__github__create_pull_request
    skip_if: "--no-pr"
```

**Output Final :**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Completed
═══════════════════════════════════════════════════════════════

| Étape   | Status                           |
|---------|----------------------------------|
| Peek    | ✓ 5 files analyzed               |
| Checks  | ✓ lint, test, build PASS         |
| Branch  | `feat/add-user-auth`             |
| Commit  | `feat(auth): add user auth`      |
| Push    | origin/feat/add-user-auth        |
| PR      | #42 - feat(auth): add user auth  |

URL: https://github.com/<owner>/<repo>/pull/42

═══════════════════════════════════════════════════════════════
```

---

## Action: --merge

### Phase 1 : Peek (RLM Pattern)

```yaml
peek_workflow:
  1_pr_info:
    action: "Récupérer info PR"
    tool: mcp__github__get_pull_request
    output: "pr_number, status, checks"

  2_ci_status:
    action: "Vérifier statut CI"
    tool: mcp__github__get_pull_request_status

  3_conflicts:
    action: "Vérifier les conflits"
    command: "git fetch && git merge-base..."
```

---

### Phase 2 : Parallelize (CI checks)

```yaml
parallel_ci:
  mode: "PARALLEL (if CI pending, wait in parallel)"

  checks:
    - task: "Wait for CI"
      poll: "every 30s"
      timeout: "10min"

    - task: "Check conflicts"
      action: "Verify no merge conflicts"

    - task: "Sync with main"
      action: "Rebase if behind"
```

---

### Phase 3 : Auto-fix Loop

```yaml
autofix_loop:
  max_attempts: 3

  on_ci_failure:
    1_analyze: "Identifier l'erreur CI"
    2_fix: "Appliquer correction automatique"
    3_commit: "Commit fix"
    4_push: "Push et attendre CI"

  on_max_reached:
    action: "Poster commentaire détaillé sur PR"
    status: "ABORT"
```

---

### Phase 4 : Synthesize (Merge & Cleanup)

```yaml
merge_workflow:
  1_merge:
    tool: mcp__github__merge_pull_request
    method: "squash"

  2_cleanup:
    actions:
      - "git push origin --delete <branch>"
      - "git branch -D <branch>"
      - "git checkout main"
      - "git pull origin main"
```

**Output Final :**

```
═══════════════════════════════════════════════════════════════
  ✓ PR #42 merged successfully
═══════════════════════════════════════════════════════════════

  Branch  : feat/add-auth → main
  Method  : squash
  Rebase  : ✓ Synced (was 3 commits behind)
  CI      : ✓ Passed (2m 34s)
  Commits : 5 commits → 1 squashed

  Cleanup:
    ✓ Remote branch deleted
    ✓ Local branch deleted
    ✓ Switched to main
    ✓ Pulled latest (now at abc1234)

═══════════════════════════════════════════════════════════════
```

---

## Conventional Commits

| Type | Usage |
|------|-------|
| `feat` | Nouvelle fonctionnalité |
| `fix` | Correction de bug |
| `refactor` | Refactoring |
| `docs` | Documentation |
| `test` | Tests |
| `chore` | Maintenance |
| `ci` | CI/CD |

---

## GARDE-FOUS (ABSOLUS)

| Action | Status | Raison |
|--------|--------|--------|
| Skip Phase 1 (Peek) | ❌ **INTERDIT** | git status avant action |
| Merge automatique sans CI | ❌ **INTERDIT** | Qualité code |
| Push sur main/master | ❌ **INTERDIT** | Branche protégée |
| Force merge si CI échoue x3 | ❌ **INTERDIT** | Limite tentatives |
| Push sans --force-with-lease | ❌ **INTERDIT** | Sécurité |
| Mentions IA dans commits | ❌ **INTERDIT** | Discrétion |

### Parallélisation légitime

| Élément | Parallèle? | Raison |
|---------|------------|--------|
| Pré-commit checks (lint+test+build) | ✅ Parallèle | Indépendants |
| Opérations git (branch→commit→push→PR) | ❌ Séquentiel | Chaîne de dépendances |
| CI checks en attente | ❌ Séquentiel | Attendre résultat |
