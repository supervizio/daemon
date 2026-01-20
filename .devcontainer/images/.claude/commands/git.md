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

  5_check_gpg:
    action: "Vérifier si GPG signing est configuré"
    commands:
      - "git config --get commit.gpgsign"
      - "git config --get user.signingkey"

  6_configure_gpg_if_missing:
    condition: "commit.gpgsign != true OR user.signingkey is empty"
    action: "Lister les clés GPG et demander sélection si nécessaire"
    workflow:
      1_list_keys: "gpg --list-secret-keys --keyid-format LONG"
      2_find_matching:
        rule: "Chercher clé correspondant à GIT_EMAIL"
        action: "grep -B1 '{email}' dans output gpg"
      3_if_no_match_but_keys_exist:
        tool: AskUserQuestion
        questions:
          - question: "Quelle clé GPG utiliser pour signer les commits ?"
            header: "GPG Key"
            options: "<dynamically generated from gpg output>"
      4_configure:
        commands:
          - "git config --global user.signingkey {selected_key}"
          - "git config --global commit.gpgsign true"
          - "git config --global tag.forceSignAnnotated true"
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
  /git --commit - Git Identity & GPG Validation
═══════════════════════════════════════════════════════════════

  .env check:
    ├─ File: /workspace/.env
    ├─ GIT_USER: "John Doe" ✓
    └─ GIT_EMAIL: "john.doe@example.com" ✓

  Git config:
    ├─ user.name: "John Doe" ✓ (match)
    └─ user.email: "john.doe@example.com" ✓ (match)

  GPG config:
    ├─ commit.gpgsign: true ✓
    └─ user.signingkey: ABCD1234EF567890 ✓

  Status: ✓ Identity & GPG validated, proceeding to Phase 1

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

### Phase 2 : CI Monitoring avec Backoff Exponentiel

```yaml
ci_monitoring:
  description: "Suivi intelligent du statut CI avec polling adaptatif"

  #---------------------------------------------------------------------------
  # CONFIGURATION
  #---------------------------------------------------------------------------
  config:
    initial_interval: 10s          # Intervalle initial (compatible gh --interval)
    max_interval: 120s             # Plafonné à 2 minutes
    backoff_multiplier: 1.5        # 10s → 15s → 22s → 33s → 50s → 75s → 112s → 120s
    jitter_percent: 20             # +/- 20% aléatoire (évite thundering herd)
    timeout: 600s                  # 10 minutes timeout total
    max_poll_attempts: 30          # Limite de sécurité

  #---------------------------------------------------------------------------
  # STRATÉGIE DE POLLING (MCP-FIRST)
  #---------------------------------------------------------------------------
  polling_strategy:
    github:
      primary:
        tool: mcp__github__get_pull_request_status
        params:
          pull_number: "{pr_number}"
        response_fields: ["state", "statuses[]", "check_runs[]"]
      fallback:
        command: "gh pr checks {pr_number} --watch --interval 10"

    gitlab:
      primary:
        tool: mcp__gitlab__list_pipelines
        params:
          project_id: "{project_id}"
          ref: "{branch}"
          per_page: 1
        response_fields: ["status", "id", "web_url"]
      fallback:
        command: "glab ci status --branch {branch}"

  #---------------------------------------------------------------------------
  # MAPPING DES STATUTS
  #---------------------------------------------------------------------------
  status_mapping:
    github:
      SUCCESS: [success, neutral]
      PENDING: [pending, queued, in_progress, waiting]
      FAILURE: [failure, action_required, timed_out, cancelled]
      ERROR: [error, stale]

    gitlab:
      SUCCESS: [success, manual]
      PENDING: [created, waiting_for_resource, preparing, pending, running]
      FAILURE: [failed]
      ERROR: [canceled, skipped]

  #---------------------------------------------------------------------------
  # ALGORITHME DE BACKOFF EXPONENTIEL
  #---------------------------------------------------------------------------
  backoff_algorithm:
    pseudocode: |
      interval = initial_interval
      elapsed = 0
      attempt = 0

      WHILE elapsed < timeout AND attempt < max_poll_attempts:
        status = poll_ci_status()

        IF status == SUCCESS:
          RETURN {status: "passed", duration: elapsed}
        IF status in [FAILURE, ERROR, CANCELED]:
          RETURN {status: "failed", duration: elapsed, details: get_failure_details()}
        IF status in [PENDING, RUNNING]:
          # Appliquer jitter
          jitter = interval * (random(-jitter_percent, +jitter_percent) / 100)
          sleep(interval + jitter)
          elapsed += interval + jitter

          # Backoff exponentiel
          interval = min(interval * backoff_multiplier, max_interval)
          attempt++

      RETURN {status: "timeout", duration: elapsed}

  #---------------------------------------------------------------------------
  # PARALLEL TASKS (pendant le polling)
  #---------------------------------------------------------------------------
  parallel_tasks:
    - task: "Check conflicts"
      action: "git fetch && git merge-base --is-ancestor origin/main HEAD"
      on_conflict: "Rebase automatique si --auto-rebase"

    - task: "Sync with main"
      action: "Rebase si behind (max 10 commits)"
      on_behind: "git rebase origin/main"
```

**Output Phase 2 :**

```
═══════════════════════════════════════════════════════════════
  /git --merge - CI Monitoring (Phase 2)
═══════════════════════════════════════════════════════════════

  PR/MR    : #42 (feat/add-auth)
  Platform : GitHub
  Timeout  : 10 minutes

  Polling CI status...
    [10:30:15] Poll #1: pending (10s elapsed, next in 10s)
    [10:30:27] Poll #2: running (22s elapsed, next in 15s)
    [10:30:45] Poll #3: running (40s elapsed, next in 22s)
    [10:31:12] Poll #4: running (67s elapsed, next in 33s)
    [10:31:50] ✓ CI PASSED (95s)

  Checks:
    ├─ build: passed (1m 23s)
    ├─ test: passed (2m 45s)
    └─ lint: passed (45s)

  Proceeding to Phase 3...

═══════════════════════════════════════════════════════════════
```

---

### Phase 3 : Auto-fix Loop avec Catégories d'Erreurs

```yaml
autofix_loop:
  description: "Détection, catégorisation et correction automatique des erreurs CI"

  #---------------------------------------------------------------------------
  # CONFIGURATION
  #---------------------------------------------------------------------------
  config:
    max_attempts: 3
    cooldown_between_attempts: 30s    # Attente avant re-trigger CI
    autofix_per_attempt_timeout: 120s # 2 min max par tentative de fix
    require_human_for:
      - security_scan
      - timeout
      - "confidence == LOW after 2 attempts"

  #---------------------------------------------------------------------------
  # CATÉGORIES D'ERREURS
  #---------------------------------------------------------------------------
  error_categories:
    #-------------------------------------------------------------------------
    # LINT ERRORS - Auto-fixable (HIGH confidence)
    #-------------------------------------------------------------------------
    lint_error:
      patterns:
        - "eslint.*error"
        - "prettier.*differ"
        - "golangci-lint.*"
        - "ruff.*error"
        - "shellcheck.*SC[0-9]+"
        - "stylelint.*"
      severity: LOW
      auto_fixable: true
      confidence: HIGH
      fix_strategy: "run_linter_fix"

    #-------------------------------------------------------------------------
    # TYPE ERRORS - Partially auto-fixable
    #-------------------------------------------------------------------------
    type_error:
      patterns:
        - "TS[0-9]+:"                    # TypeScript errors
        - "type.*incompatible"
        - "cannot find name"
        - "go build.*undefined:"         # Go type errors
        - "mypy.*error:"                 # Python mypy
      severity: MEDIUM
      auto_fixable: partial
      confidence: MEDIUM
      fix_strategy: "type_fix"

    #-------------------------------------------------------------------------
    # TEST FAILURES - Conditional auto-fix
    #-------------------------------------------------------------------------
    test_failure:
      patterns:
        - "FAIL.*test"
        - "AssertionError"
        - "expected.*but got"
        - "Error: expect\\("
        - "--- FAIL:"                    # Go test failures
        - "FAILED.*::.*::"               # pytest
      severity: HIGH
      auto_fixable: conditional
      confidence: MEDIUM
      fix_strategy: "test_analysis"

    #-------------------------------------------------------------------------
    # BUILD ERRORS - Requires careful analysis
    #-------------------------------------------------------------------------
    build_error:
      patterns:
        - "error: cannot find module"
        - "Module not found"
        - "compilation failed"
        - "SyntaxError:"
        - "package.*not found"
      severity: HIGH
      auto_fixable: partial
      confidence: LOW
      fix_strategy: "build_analysis"

    #-------------------------------------------------------------------------
    # SECURITY SCAN - NEVER auto-fix
    #-------------------------------------------------------------------------
    security_scan:
      patterns:
        - "CRITICAL.*vulnerability"
        - "HIGH.*CVE-"
        - "security.*violation"
        - "secret.*detected"
        - "trivy.*CRITICAL"
      severity: CRITICAL
      auto_fixable: false
      confidence: N/A
      fix_strategy: "user_intervention_required"

    #-------------------------------------------------------------------------
    # DEPENDENCY ERRORS - Often auto-fixable
    #-------------------------------------------------------------------------
    dependency_error:
      patterns:
        - "npm ERR!.*peer dep"
        - "cannot resolve dependency"
        - "go: module.*not found"
        - "pip.*ResolutionImpossible"
      severity: MEDIUM
      auto_fixable: true
      confidence: MEDIUM
      fix_strategy: "dependency_fix"

    #-------------------------------------------------------------------------
    # INFRASTRUCTURE ERRORS - Retry only
    #-------------------------------------------------------------------------
    infrastructure_error:
      patterns:
        - "rate limit"
        - "connection refused"
        - "503 Service Unavailable"
        - "ECONNRESET"
      severity: LOW
      auto_fixable: retry
      confidence: HIGH
      fix_strategy: "retry_ci"

  #---------------------------------------------------------------------------
  # ALGORITHME DE LA BOUCLE
  #---------------------------------------------------------------------------
  loop_algorithm:
    pseudocode: |
      attempt = 0
      fix_history = []

      WHILE attempt < max_attempts:
        attempt++

        # Step 1: Récupérer détails de l'échec CI
        failure = get_ci_failure_details()
        category = categorize_error(failure)

        # Step 2: Vérifier si auto-fixable
        IF NOT category.auto_fixable:
          RETURN abort_with_report(category, failure)

        # Step 3: Détecter fix circulaire
        IF is_circular_fix(category, fix_history):
          RETURN abort_with_circular_warning(fix_history)

        # Step 4: Appliquer stratégie de fix
        fix_result = apply_fix_strategy(category)
        fix_history.append({category, fix_result})

        IF fix_result.success:
          # Step 5: Commit et push
          commit_fix(fix_result)
          push_to_remote()

          # Step 6: Attendre cooldown puis re-poll CI
          sleep(cooldown_between_attempts)
          ci_status = poll_ci_with_backoff()  # Re-use Phase 2

          IF ci_status == SUCCESS:
            RETURN success_report(attempt, fix_history)
        ELSE:
          RETURN abort_with_fix_failure(fix_result)

      # Max attempts reached
      RETURN abort_max_attempts(fix_history)

  #---------------------------------------------------------------------------
  # FIX STRATEGIES
  #---------------------------------------------------------------------------
  fix_strategies:
    run_linter_fix:
      detect_linter:
        - check: "package.json"
          command: "npm run lint -- --fix"
        - check: ".golangci.yml"
          command: "golangci-lint run --fix"
        - check: "pyproject.toml [tool.ruff]"
          command: "ruff check --fix"
      commit_format: "fix(lint): auto-fix {linter} errors"

    type_fix:
      workflow:
        1_extract: "Parser CI log pour erreurs de type spécifiques"
        2_analyze: "Identifier le fichier et la ligne"
        3_fix: "Appliquer correction minimale"
        4_verify: "npm run typecheck OR go build"
      commit_format: "fix(types): resolve {error_code} in {file}"

    test_analysis:
      conditions:
        assertion_mismatch:
          pattern: "expected.*but got"
          auto_fix: true
          strategy: "Update assertion si implementation changed"
        snapshot_mismatch:
          pattern: "snapshot.*differ"
          auto_fix: true
          strategy: "npm test -- -u"
        timeout:
          pattern: "exceeded timeout"
          auto_fix: false
      commit_format: "fix(test): update {test_name}"

    dependency_fix:
      strategies:
        npm: "npm install --legacy-peer-deps"
        go: "go mod tidy"
        pip: "pip install --upgrade"
      commit_format: "fix(deps): resolve {package} conflict"

    retry_ci:
      wait: 60s
      retrigger:
        github: "gh run rerun --failed"
        gitlab: "glab ci retry"

    user_intervention_required:
      action: "Generate detailed failure report"
      include:
        - "Error category and severity"
        - "Relevant CI log snippets (max 50 lines)"
        - "Affected files"
        - "Suggested manual steps"
      block_merge: true
```

**Output Phase 3 (Auto-fix Success) :**

```
═══════════════════════════════════════════════════════════════
  /git --merge - Auto-fix Loop (Phase 3)
═══════════════════════════════════════════════════════════════

  Attempt 1/3 - lint_error
  -------------------------
    Category : lint_error (LOW severity)
    Confidence: HIGH
    Auto-fix : YES

    Error: eslint: 3 errors in src/utils/parser.ts
      ├─ Line 45: no-unused-vars
      ├─ Line 67: prefer-const
      └─ Line 89: no-console

    Fix: npm run lint -- --fix
    Result: ✓ Fixed

    Commit: fix(lint): auto-fix eslint errors in parser.ts
    Push: origin/feat/add-parser

    Re-polling CI...
      [10:32:45] ✓ CI PASSED (67s)

═══════════════════════════════════════════════════════════════
  ✓ Auto-fix Successful (1 attempt)
═══════════════════════════════════════════════════════════════

  Commits added: 1
    └─ fix(lint): auto-fix eslint errors in parser.ts

  Proceeding to Phase 4 (Merge)...

═══════════════════════════════════════════════════════════════
```

**Output Phase 3 (Security Block) :**

```
═══════════════════════════════════════════════════════════════
  /git --merge - BLOCKED (Security Issue)
═══════════════════════════════════════════════════════════════

  ⛔ AUTO-FIX DISABLED for security issues

  Category: security_scan
  Severity: CRITICAL

  Vulnerability:
    ┌─────────────────────────────────────────────────────────┐
    │ CRITICAL CVE-2023-44487                                 │
    │ Package: golang.org/x/net v0.7.0                        │
    │ Fixed in: v0.17.0                                       │
    └─────────────────────────────────────────────────────────┘

  Required Actions:
    1. go get golang.org/x/net@v0.17.0 && go mod tidy
    2. trivy fs --severity CRITICAL .
    3. Re-run /git --merge

  ⚠️  Force merge NOT available for security issues.

═══════════════════════════════════════════════════════════════
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
| Skip Phase 2 (CI Polling) | ❌ **INTERDIT** | CI validation obligatoire |
| Merge automatique sans CI | ❌ **INTERDIT** | Qualité code |
| Push sur main/master | ❌ **INTERDIT** | Branche protégée |
| Force merge si CI échoue x3 | ❌ **INTERDIT** | Limite tentatives |
| Push sans --force-with-lease | ❌ **INTERDIT** | Sécurité |
| Mentions IA dans commits | ❌ **INTERDIT** | Discrétion |
| Commit sans identité validée | ❌ **INTERDIT** | Traçabilité |

### Auto-fix Safeguards

| Action | Status | Raison |
|--------|--------|--------|
| Auto-fix vulnerabilités sécurité | ❌ **INTERDIT** | Review humain requis |
| Merge avec issues CRITICAL | ❌ **INTERDIT** | Sécurité first |
| Fix circulaire (même erreur 3x) | ❌ **INTERDIT** | Prévient boucle infinie |
| Modifier .claude/ via auto-fix | ❌ **INTERDIT** | Config protégée |
| Modifier .devcontainer/ via auto-fix | ❌ **INTERDIT** | Config protégée |
| Auto-fix sans commit message | ❌ **INTERDIT** | Traçabilité |

### Timeouts Auto-fix

| Élément | Valeur | Raison |
|---------|--------|--------|
| CI Polling total | 600s (10min) | Éviter attente infinie |
| Par tentative de fix | 120s (2min) | Éviter blocage |
| Cooldown entre tentatives | 30s | Laisser CI démarrer |
| Jitter polling | ±20% | Éviter thundering herd |

### Parallélisation légitime

| Élément | Parallèle? | Raison |
|---------|------------|--------|
| Pré-commit checks (lint+test+build) | ✅ Parallèle | Indépendants |
| CI polling + conflict check | ✅ Parallèle | Indépendants |
| Opérations git (branch→commit→push→PR) | ❌ Séquentiel | Chaîne de dépendances |
| Tentatives auto-fix | ❌ Séquentiel | Dépend du résultat CI |
| CI checks en attente | ❌ Séquentiel | Attendre résultat |
