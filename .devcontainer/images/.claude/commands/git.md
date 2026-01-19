---
name: git
description: |
  Workflow Git Automation with RLM decomposition.
  Handles branch management, conventional commits, and CI validation.
  Use when: committing changes, creating PRs/MRs, or merging with CI checks.
  Supports GitHub (PRs) and GitLab (MRs) - auto-detected from git remote.
allowed-tools:
  - "Bash(git:*)"
  - "Bash(gh:*)"
  - "Bash(glab:*)"
  - "mcp__github__*"
  - "mcp__gitlab__*"
  - "Read(**/*)"
  - "Write(.env)"
  - "Edit(.env)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "Task(*)"
  - "AskUserQuestion(*)"
---

# /git - Workflow Git Automation (RLM Architecture)

$ARGUMENTS

---

## Overview

Automatisation Git avec patterns **RLM** :

- **Identity** - Vérifier/configurer l'identité git via `.env`
- **Peek** - Analyser l'état git avant action
- **Decompose** - Identifier les fichiers par catégorie
- **Parallelize** - Checks en parallèle (lint, test, CI)
- **Synthesize** - Rapport consolidé

**Note :** L'identité git (user.name/user.email) est stockée dans `/workspace/.env` et synchronisée automatiquement avec git config à chaque exécution.

---

## Arguments

| Pattern | Action |
|---------|--------|
| `--commit` | Workflow complet : branch, commit, push, PR/MR |
| `--merge` | Merge la PR/MR avec CI validation |
| `--help` | Affiche l'aide |

### Options --commit

| Option | Action |
|--------|--------|
| `--branch <nom>` | Force le nom de branche |
| `--no-pr` | Skip la création de PR/MR |
| `--amend` | Amend le dernier commit |
| `--skip-identity` | Skip la vérification d'identité git |

### Options --merge

| Option | Action |
|--------|--------|
| `--pr <number>` | Merge une PR spécifique (GitHub) |
| `--mr <number>` | Merge une MR spécifique (GitLab) |
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
  --commit          Workflow complet (branch, commit, push, PR/MR)
  --merge           Merge avec CI validation et auto-fix

RLM Patterns:
  0.5. Identity    - Vérifier/configurer git user via .env
  1. Peek          - Analyser état git
  2. Decompose     - Catégoriser fichiers
  3. Parallelize   - Checks simultanés
  4. Synthesize    - Rapport consolidé

Options --commit:
  --branch <nom>    Force le nom de branche
  --no-pr           Skip la création de PR/MR
  --amend           Amend le dernier commit
  --skip-identity   Skip la vérification d'identité

Options --merge:
  --pr <number>     Merge une PR spécifique (GitHub)
  --mr <number>     Merge une MR spécifique (GitLab)
  --strategy <type> Méthode: merge/squash/rebase (défaut: squash)
  --dry-run         Vérifier sans merger

Identity (.env):
  - GIT_USER et GIT_EMAIL stockés dans /workspace/.env
  - Synchronisé automatiquement avec git config
  - Demandé à l'utilisateur si absent

Exemples:
  /git --commit                 Commit + PR automatique
  /git --commit --no-pr         Commit sans créer de PR
  /git --commit --skip-identity Skip vérification identité
  /git --merge                  Merge la PR/MR courante
  /git --merge --pr 42          Merge la PR #42

═══════════════════════════════════════════════════════════════
```

---

## Priorité MCP vs CLI

**IMPORTANT** : Toujours privilégier les outils MCP quand disponibles.

**Platform auto-détectée :** `git remote get-url origin` → github.com | gitlab.*

### GitHub (PRs)

| Action | Priorité 1 (MCP) | Fallback (CLI) |
|--------|------------------|----------------|
| Créer branche | `mcp__github__create_branch` | `git checkout -b` |
| Créer PR | `mcp__github__create_pull_request` | `gh pr create` |
| Lister PRs | `mcp__github__list_pull_requests` | `gh pr list` |
| Voir PR | `mcp__github__get_pull_request` | `gh pr view` |
| Status CI | `mcp__github__get_pull_request_status` | `gh pr checks` |
| Merger PR | `mcp__github__merge_pull_request` | `gh pr merge` |

### GitLab (MRs)

| Action | Priorité 1 (MCP) | Fallback (CLI) |
|--------|------------------|----------------|
| Créer branche | `git checkout -b` + push | `git checkout -b` |
| Créer MR | `mcp__gitlab__create_merge_request` | `glab mr create` |
| Lister MRs | `mcp__gitlab__list_merge_requests` | `glab mr list` |
| Voir MR | `mcp__gitlab__get_merge_request` | `glab mr view` |
| Status CI | `mcp__gitlab__list_pipelines` | `glab ci status` |
| Merger MR | `mcp__gitlab__merge_merge_request` | `glab mr merge` |

---

## Action: --commit

### Phase 0.5 : Git Identity Validation (OBLIGATOIRE)

**Vérifier et configurer l'identité git AVANT toute action :**

```yaml
identity_validation:
  env_file: "/workspace/.env"

  1_check_env:
    action: "Vérifier si .env existe et contient GIT_USER/GIT_EMAIL"
    tool: Read("/workspace/.env")
    fallback: "Fichier non trouvé → créer"

  2_extract_or_ask:
    rule: |
      SI .env existe ET contient GIT_USER ET GIT_EMAIL:
        user = extract(GIT_USER)
        email = extract(GIT_EMAIL)
      SINON:
        → AskUserQuestion (voir ci-dessous)
        → Créer/Mettre à jour .env

  3_verify_git_config:
    action: "Comparer avec git config actuel"
    commands:
      - "git config user.name"
      - "git config user.email"
    decision:
      if_match: "→ Continuer vers Phase 1"
      if_mismatch: "→ Corriger git config"

  4_fix_if_needed:
    action: "Appliquer la configuration correcte"
    commands:
      - "git config user.name '{user}'"
      - "git config user.email '{email}'"
```

**Question si .env absent ou incomplet :**

```yaml
ask_identity:
  tool: AskUserQuestion
  questions:
    - question: "Quel nom utiliser pour les commits git ?"
      header: "Git User"
      options:
        - label: "{detected_user}"
          description: "Détecté depuis git config global"
        - label: "{github_user}"
          description: "Détecté depuis GitHub/GitLab"
      # L'utilisateur peut aussi entrer "Other" avec valeur custom

    - question: "Quelle adresse email utiliser pour les commits ?"
      header: "Git Email"
      options:
        - label: "{detected_email}"
          description: "Détecté depuis git config global"
        - label: "{noreply_email}"
          description: "Email noreply GitHub/GitLab"
```

**Format .env généré/mis à jour :**

```bash
# Git identity for commits (managed by /git)
GIT_USER="John Doe"
GIT_EMAIL="john.doe@example.com"
```

**Output Phase 0.5 :**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Git Identity Validation
═══════════════════════════════════════════════════════════════

  .env check:
    ├─ File: /workspace/.env
    ├─ GIT_USER: "John Doe" ✓
    └─ GIT_EMAIL: "john.doe@example.com" ✓

  Git config:
    ├─ user.name: "John Doe" ✓ (match)
    └─ user.email: "john.doe@example.com" ✓ (match)

  Status: ✓ Identity validated, proceeding to Phase 1

═══════════════════════════════════════════════════════════════
```

**Output si correction nécessaire :**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Git Identity Validation
═══════════════════════════════════════════════════════════════

  .env check:
    ├─ File: /workspace/.env
    ├─ GIT_USER: "John Doe" ✓
    └─ GIT_EMAIL: "john.doe@example.com" ✓

  Git config:
    ├─ user.name: "johndoe" ✗ (mismatch)
    └─ user.email: "old@email.com" ✗ (mismatch)

  Action: Correcting git config...
    ├─ git config user.name "John Doe"
    └─ git config user.email "john.doe@example.com"

  Status: ✓ Identity corrected, proceeding to Phase 1

═══════════════════════════════════════════════════════════════
```

**Output si .env absent :**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Git Identity Validation
═══════════════════════════════════════════════════════════════

  .env check:
    └─ File: NOT FOUND → Creating...

  User input required...
    ├─ Git User: "John Doe" (entered)
    └─ Git Email: "john.doe@example.com" (entered)

  Actions:
    ├─ Created /workspace/.env with GIT_USER, GIT_EMAIL
    ├─ git config user.name "John Doe"
    └─ git config user.email "john.doe@example.com"

  Status: ✓ Identity configured, proceeding to Phase 1

═══════════════════════════════════════════════════════════════
```

---

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

  5_pr_mr:
    action: "Créer la PR/MR"
    tools:
      github: mcp__github__create_pull_request
      gitlab: mcp__gitlab__create_merge_request
    skip_if: "--no-pr"
```

**Output Final (GitHub) :**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Completed (GitHub)
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

**Output Final (GitLab) :**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Completed (GitLab)
═══════════════════════════════════════════════════════════════

| Étape   | Status                           |
|---------|----------------------------------|
| Peek    | ✓ 5 files analyzed               |
| Checks  | ✓ lint, test, build PASS         |
| Branch  | `feat/add-user-auth`             |
| Commit  | `feat(auth): add user auth`      |
| Push    | origin/feat/add-user-auth        |
| MR      | !42 - feat(auth): add user auth  |

URL: https://gitlab.com/<owner>/<repo>/-/merge_requests/42

═══════════════════════════════════════════════════════════════
```

---

## Action: --merge

### Phase 1 : Peek (RLM Pattern)

```yaml
peek_workflow:
  1_pr_mr_info:
    action: "Récupérer info PR/MR"
    tools:
      github: mcp__github__get_pull_request
      gitlab: mcp__gitlab__get_merge_request
    output: "pr_mr_number, status, checks"

  2_ci_status:
    action: "Vérifier statut CI"
    tools:
      github: mcp__github__get_pull_request_status
      gitlab: mcp__gitlab__list_pipelines

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
    tools:
      github: mcp__github__merge_pull_request
      gitlab: mcp__gitlab__merge_merge_request
    method: "squash"

  2_cleanup:
    actions:
      - "git push origin --delete <branch>"
      - "git branch -D <branch>"
      - "git checkout main"
      - "git pull origin main"
```

**Output Final (GitHub) :**

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

**Output Final (GitLab) :**

```
═══════════════════════════════════════════════════════════════
  ✓ MR !42 merged successfully
═══════════════════════════════════════════════════════════════

  Branch  : feat/add-auth → main
  Method  : squash
  Pipeline: ✓ Passed (#12345, 2m 34s)
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
| Skip Phase 0.5 (Identity) sans flag | ❌ **INTERDIT** | Identité git requise |
| Skip Phase 1 (Peek) | ❌ **INTERDIT** | git status avant action |
| Merge automatique sans CI | ❌ **INTERDIT** | Qualité code |
| Push sur main/master | ❌ **INTERDIT** | Branche protégée |
| Force merge si CI échoue x3 | ❌ **INTERDIT** | Limite tentatives |
| Push sans --force-with-lease | ❌ **INTERDIT** | Sécurité |
| Mentions IA dans commits | ❌ **INTERDIT** | Discrétion |
| Commit sans identité validée | ❌ **INTERDIT** | Traçabilité |

### Parallélisation légitime

| Élément | Parallèle? | Raison |
|---------|------------|--------|
| Pré-commit checks (lint+test+build) | ✅ Parallèle | Indépendants |
| Opérations git (branch→commit→push→PR) | ❌ Séquentiel | Chaîne de dépendances |
| CI checks en attente | ❌ Séquentiel | Attendre résultat |
