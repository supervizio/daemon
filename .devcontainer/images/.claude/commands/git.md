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
  - "TaskCreate(*)"
  - "TaskUpdate(*)"
  - "TaskList(*)"
  - "TaskGet(*)"
  - "AskUserQuestion(*)"
---

# /git - Workflow Git Automation (RLM Architecture)

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Use `grepai_trace_callers`/`grepai_trace_callees` for impact analysis.
Fallback to Grep ONLY for exact string matches or regex patterns.

$ARGUMENTS

---

## Overview

Automatisation Git avec patterns **RLM** :

- **Identity** - Vérifier/configurer l'identité git via `.env`
- **Peek** - Analyser l'état git avant action
- **Decompose** - Identifier les fichiers par catégorie
- **Parallelize** - Checks en parallèle (lint, test, CI)
- **Context** - `/warmup --update` sur fichiers modifiés (branch diff)
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
  3.8. Context     - /warmup --update (branch diff, 5min staleness)
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

### Phase 1.0 : Git Identity Validation (OBLIGATOIRE)

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

### Phase 2.0 : Peek (RLM Pattern)

**Analyser l'état git AVANT toute action :**

```yaml
peek_workflow:
  1_status:
    action: "Vérifier l'état du repo (TOUTES les modifications, pas seulement la tâche courante)"
    commands:
      - "git status --porcelain"
      - "git branch --show-current"
      - "git log -1 --format='%h %s'"
    critical_rule: |
      LISTER TOUS les fichiers modifiés — y compris CLAUDE.md, .devcontainer/,
      .claude/commands/. Ne JAMAIS ignorer des fichiers trackés modifiés.
      git status --porcelain montre TOUT ce qui est tracké et modifié.
      Les fichiers gitignorés N'APPARAISSENT PAS → pas de risque de les inclure.

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

### Phase 3.0 : Decompose (RLM Pattern)

**Catégoriser les fichiers modifiés :**

```yaml
decompose_workflow:
  categories:
    features:
      patterns: ["src/**/*.ts", "src/**/*.js", "src/**/*.go", "src/**/*.rs", "src/**/*.py"]
      prefix: "feat"

    fixes:
      patterns: ["*fix*", "*bug*"]
      prefix: "fix"

    tests:
      patterns: ["tests/**", "**/*.test.*", "**/*_test.go"]
      prefix: "test"

    docs:
      patterns: ["*.md", "docs/**", "**/CLAUDE.md", ".claude/commands/*.md"]
      prefix: "docs"

    config:
      patterns: ["*.json", "*.yaml", "*.toml", ".devcontainer/**"]
      prefix: "chore"

    hooks:
      patterns: [".devcontainer/hooks/**", ".claude/scripts/**", ".githooks/**"]
      prefix: "fix"

  auto_detect:
    action: "Déduire le type dominant"
    output: "commit_type, scope, branch_name"

  gitignore_awareness:
    rule: |
      AVANT de catégoriser, vérifier le statut gitignore de chaque fichier.
      Utiliser `git status --porcelain` pour lister TOUS les fichiers modifiés.
      Les fichiers gitignorés n'apparaissent pas dans git status → pas de risque.
      Les fichiers trackés modifiés DOIVENT être inclus, même s'ils sont dans .claude/ ou CLAUDE.md.
    check: |
      # Lister TOUTES les modifications (staged + unstaged + untracked non-ignored)
      git status --porcelain
      # Vérifier que rien de tracké n'est oublié après staging
      git diff --name-only  # Doit être vide après git add -A
```

---

### Phase 4.0 : Parallelize (RLM Pattern) - Multi-Language Pre-commit

**Auto-detect ALL project languages and run checks for each:**

```yaml
language_detection:
  script: ".claude/scripts/pre-commit-checks.sh"

  detection_files:
    go.mod: "Go"
    Cargo.toml: "Rust"
    package.json: "Node.js"
    pyproject.toml: "Python"
    requirements.txt: "Python"
    Gemfile: "Ruby"
    pom.xml: "Java (Maven)"
    build.gradle: "Java/Kotlin (Gradle)"
    mix.exs: "Elixir"
    composer.json: "PHP"
    pubspec.yaml: "Dart/Flutter"
    build.sbt: "Scala"
    CMakeLists.txt: "C/C++ (CMake)"
    meson.build: "C/C++ (Meson)"
    "*.csproj": "C# (.NET)"
    "*.sln": "C# (.NET)"
    Package.swift: "Swift"
    DESCRIPTION: "R"
    cpanfile: "Perl"
    Makefile.PL: "Perl"
    "*.rockspec": "Lua"
    .luacheckrc: "Lua"
    fpm.toml: "Fortran"
    alire.toml: "Ada"
    "*.gpr": "Ada"
    "*.cob": "COBOL"
    "*.cbl": "COBOL"
    "*.lpi": "Pascal"
    "*.vbproj": "VB.NET"

parallel_checks:
  mode: "PARALLEL (single message, multiple calls)"

  for_each_detected_language:
    - task: "lint-check"
      priority: "Makefile target > language-specific tool"
      commands:
        go: "golangci-lint run ./..."
        rust: "cargo clippy -- -D warnings"
        nodejs: "npm run lint"
        python: "ruff check ."
        ruby: "bundle exec rubocop"
        java-maven: "mvn checkstyle:check"
        java-gradle: "./gradlew check"
        elixir: "mix credo --strict"
        php: "vendor/bin/phpstan analyse"
        dart: "dart analyze --fatal-infos"

    - task: "build-check"
      commands:
        go: "go build ./..."
        rust: "cargo build --release"
        nodejs: "npm run build"
        java-maven: "mvn compile -q"
        java-gradle: "./gradlew build -x test"
        elixir: "mix compile --warnings-as-errors"

    - task: "test-check"
      commands:
        go: "go test -race ./..."
        rust: "cargo test"
        nodejs: "npm test"
        python: "pytest"
        ruby: "bundle exec rspec"
        java-maven: "mvn test -q"
        java-gradle: "./gradlew test"
        elixir: "mix test"
        php: "vendor/bin/phpunit"
        dart: "dart test"
```

**Output Multi-Language:**

```
═══════════════════════════════════════════════════════════════
   Pre-commit Checks
═══════════════════════════════════════════════════════════════

  Languages detected: Go, Rust

--- Rust Checks ---
[CHECK] Rust lint (clippy)...
[PASS] Rust lint (clippy)
[CHECK] Rust build...
[PASS] Rust build
[CHECK] Rust tests...
[PASS] Rust tests

--- Go Checks ---
[CHECK] Go lint (golangci-lint)...
[PASS] Go lint (golangci-lint)
[CHECK] Go build...
[PASS] Go build
[CHECK] Go tests (with race detection)...
[PASS] Go tests (with race detection)

═══════════════════════════════════════════════════════════════
   All pre-commit checks passed
═══════════════════════════════════════════════════════════════
```

**IMPORTANT** : Run `.claude/scripts/pre-commit-checks.sh` which auto-detects languages.

---

### Phase 5.0 : Secret Scan (1Password Integration)

**REGLE ABSOLUE : Aucun secret/mot de passe reel ne doit fuiter dans un commit.**

**Politique secrets :**

| Type | Action | Exemple |
|------|--------|---------|
| Secret reel (token, mdp prod) | **BLOQUER le commit** | `ghp_abc123...`, `postgres://user:realpass@prod/db` |
| Mot de passe de test | **OK si dans fichier `.example`** | `.env.example`, `config.example.yaml` |
| Mot de passe de test dans le code | **OK si commente explicitement** | `// TEST ONLY - not a real credential` |
| Fichier `.env` avec vrais secrets | **JAMAIS committe** | Doit etre dans `.gitignore` |

**Fichiers `.example` :** Les mots de passe de test dans des fichiers `.example` sont acceptes car ils servent de documentation. Ils DOIVENT avoir un commentaire expliquant que ce sont des valeurs de test :

```bash
# .env.example - Test/default values only, NOT real credentials
DB_PASSWORD=test_password_change_me    # TEST ONLY
API_KEY=sk-test-fake-key-for-dev       # TEST ONLY
```

**Scanner les fichiers staged pour des secrets hardcodes :**

```yaml
secret_scan:
  trigger: "ALWAYS run in parallel with language checks"
  blocking: true  # BLOQUE le commit si secret reel detecte

  0_policy:
    real_secrets: "BLOCK - never commit real tokens, passwords, API keys"
    test_passwords_in_example_files: "ALLOW - .example files are documentation"
    test_passwords_in_code: "ALLOW if commented with '// TEST ONLY' or '# TEST ONLY'"
    env_files: "BLOCK - .env must be in .gitignore, use .env.example instead"

  1_get_staged_files:
    command: "git diff --cached --name-only"
    exclude: [".env", ".env.*", "*.lock", "*.sum"]

  1b_check_env_not_staged:
    command: "git diff --cached --name-only | grep -E '^\.env$' || true"
    action: |
      SI .env est staged:
        BLOQUER le commit
        Message: ".env contient potentiellement des secrets reels. Utilisez .env.example pour les valeurs par defaut."

  2_scan_patterns:
    patterns:
      tokens:
        - 'ghp_[a-zA-Z0-9]{36}'           # GitHub PAT
        - 'glpat-[a-zA-Z0-9\-]{20}'       # GitLab PAT
        - 'sk-[a-zA-Z0-9]{48}'            # OpenAI/Stripe secret key
        - 'pk_[a-zA-Z0-9]{24,}'           # Stripe publishable key
        - 'ops_[a-zA-Z0-9]{50,}'          # 1Password service account
        - 'AKIA[0-9A-Z]{16}'             # AWS access key
      connection_strings:
        - 'postgres://[^\s]+'
        - 'mysql://[^\s]+'
        - 'mongodb(\+srv)?://[^\s]+'
      generic:
        - '[a-zA-Z0-9+/]{40,}={0,2}'     # Long base64 (potential secrets)

    exceptions:
      - file_pattern: "*.example*"         # .env.example, config.example.yaml
      - file_pattern: "*_example.*"
      - file_pattern: "*.sample*"
      - comment_marker: "TEST ONLY"        # Inline comment marks test value
      - comment_marker: "FAKE"
      - comment_marker: "PLACEHOLDER"
      - value_pattern: "test_*"            # test_password, test_token
      - value_pattern: "fake_*"
      - value_pattern: "dummy_*"
      - value_pattern: "changeme"
      - value_pattern: "TODO:*"

  3_if_secrets_found:
    action: "BLOCK commit + suggestion"
    output: |
      ═══════════════════════════════════════════════════════════════
        ⛔ REAL SECRETS DETECTED - COMMIT BLOCKED
      ═══════════════════════════════════════════════════════════════

        Found {count} potential secret(s) in staged files:

        File: src/config.go
          Line 42: ghp_xxxx... (GitHub PAT)
          Suggestion: /secret --push GITHUB_TOKEN=<value>
                      Replace with: os.Getenv("GITHUB_TOKEN")

        File: .env.production
          Line 5: postgres://user:pass@host/db
          Suggestion: /secret --push DATABASE_URL=<value>

        Action: Use /secret --push to store in 1Password
                Then replace with env var reference

        Test passwords? Put them in .env.example with comment:
          DB_PASSWORD=test_pass  # TEST ONLY

      ═══════════════════════════════════════════════════════════════

  4_if_no_secrets:
    output: "[PASS] No hardcoded secrets detected"
```

---

### Phase 6.0 : Context Update (MANDATORY before commit)

**Met à jour les fichiers CLAUDE.md pour refléter les modifications de la branche.**

**IMPORTANT** : Cette phase s'exécute APRÈS lint/test/build (Phase 3) pour éviter de
relancer `/warmup --update` si les checks échouent et nécessitent des corrections.

```yaml
context_update_workflow:
  trigger: "ALWAYS (mandatory before commit)"
  position: "After Phase 3 + 3.5 (all checks pass), before Phase 4 (commit)"
  tool: "/warmup --update"

  1_collect_branch_diff:
    action: "Identifier TOUS les fichiers modifiés sur la branche"
    command: |
      # Fichiers modifiés dans la branche entière (vs main)
      git diff main...HEAD --name-only 2>/dev/null || git diff HEAD --name-only
      # + fichiers non-staged/non-committed (en cours)
      git diff --name-only
      git diff --cached --name-only
      # Dédupliquer
    output: "changed_files[] (unique list)"

  2_resolve_claude_files:
    action: "Trouver les CLAUDE.md concernés par les fichiers modifiés"
    algorithm: |
      POUR chaque fichier modifié:
        dir = dirname(fichier)
        TANT QUE dir != /workspace:
          SI existe(dir/CLAUDE.md):
            ajouter(dir/CLAUDE.md) au set
          dir = parent(dir)
      # Toujours inclure /workspace/CLAUDE.md (racine)
    output: "claude_files_to_update[] (unique set)"

  3_check_staleness:
    action: "Vérifier le timestamp de dernière mise à jour"
    algorithm: |
      POUR chaque claude_file DANS claude_files_to_update:
        first_line = read_first_line(claude_file)
        SI first_line match '<!-- updated: YYYY-MM-DDTHH:MM:SSZ -->':
          timestamp = parse_iso(first_line)
          age = now() - timestamp
          SI age < 5 minutes:
            skip(claude_file)  # Déjà à jour
            log("Skipping {claude_file} (updated {age} ago)")
        SINON:
          include(claude_file)  # Pas de timestamp = toujours mettre à jour
    output: "stale_claude_files[] (files needing update)"

  4_run_warmup_update:
    condition: "stale_claude_files is not empty"
    action: "Exécuter /warmup --update sur les fichiers périmés"
    tool: "Skill(warmup, --update)"
    scope: "Limité aux répertoires des stale_claude_files"
    note: |
      /warmup --update ajoutera automatiquement le timestamp ISO
      en première ligne de chaque CLAUDE.md mis à jour:
        <!-- updated: 2026-02-11T14:30:00Z -->

  5_stage_updated_docs:
    action: "Ajouter les CLAUDE.md mis à jour au staging"
    command: "git add **/CLAUDE.md"
    note: "Inclus dans le même commit que les modifications de code"

  timestamp_format:
    format: "<!-- updated: YYYY-MM-DDTHH:MM:SSZ -->"
    example: "<!-- updated: 2026-02-11T14:30:00Z -->"
    position: "Première ligne du fichier CLAUDE.md"
    purpose: "Détection de fraîcheur (staleness check 5 minutes)"
    parse: "ISO 8601 - format le plus facile à parser programmatiquement"
```

**Output Phase 3.8 :**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Context Update (Phase 3.8)
═══════════════════════════════════════════════════════════════

  Branch diff: 12 files changed

  CLAUDE.md resolution:
    ├─ /workspace/CLAUDE.md (stale, 2h ago)
    ├─ .devcontainer/CLAUDE.md (stale, no timestamp)
    ├─ .devcontainer/images/CLAUDE.md (fresh, 3m ago) → SKIP
    └─ .devcontainer/hooks/CLAUDE.md (stale, 45m ago)

  /warmup --update:
    ✓ 3 CLAUDE.md files updated
    ✓ Timestamps refreshed
    ✓ Staged for commit

═══════════════════════════════════════════════════════════════
```

---

### Phase 7.0 : Execute & Synthesize

```yaml
execute_workflow:
  1_branch:
    action: "Créer ou utiliser branche"
    auto: true

  2_stage:
    action: "Stage TOUS les fichiers modifiés trackés"
    steps:
      - command: "git add -A"
        note: "git add -A respecte .gitignore automatiquement — aucun fichier ignoré ne sera stagé"
      - command: "git diff --name-only"
        verify: "DOIT être vide — sinon des fichiers trackés ont été oubliés"
        on_failure: |
          Si des fichiers trackés restent unstaged après git add -A :
          → Les ajouter explicitement avec git add <file>
          → Ne JAMAIS ignorer des modifications de fichiers trackés (CLAUDE.md, .claude/commands/, hooks/)
    rules:
      - "TOUJOURS utiliser git add -A (jamais de staging sélectif par nom de fichier)"
      - "git add -A inclut automatiquement : CLAUDE.md, .devcontainer/, .claude/commands/"
      - "git add -A exclut automatiquement : .env, mcp.json, .grepai/, .claude/* (sauf exceptions gitignore)"
      - "Vérifier git diff --name-only après staging — si non-vide, il y a un problème"
      - "Si un fichier tracké ne doit PAS être commité → git restore <file> AVANT staging, pas après"

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
| Context | ✓ 3 CLAUDE.md updated            |
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
| Context | ✓ 3 CLAUDE.md updated            |
| Branch  | `feat/add-user-auth`             |
| Commit  | `feat(auth): add user auth`      |
| Push    | origin/feat/add-user-auth        |
| MR      | !42 - feat(auth): add user auth  |

URL: https://gitlab.com/<owner>/<repo>/-/merge_requests/42

═══════════════════════════════════════════════════════════════
```

---

## Action: --merge

### MCP-ONLY Policy (STRICT - Issue #142)

**NEVER use CLI for pipeline status. Always use MCP tools:**

```yaml
mcp_only_policy:
  MANDATORY:
    github:
      pipeline_status: "mcp__github__get_pull_request"
      check_runs: "mcp__github__list_check_runs (via pull_request_read)"
    gitlab:
      pipeline_status: "mcp__gitlab__list_pipelines"
      pipeline_jobs: "mcp__gitlab__list_pipeline_jobs"

  FORBIDDEN:
    - "gh pr checks"
    - "gh run view"
    - "glab ci status"
    - "glab ci view"
    - "curl api.github.com"
    - "curl gitlab.com/api"

  rationale: |
    CLI commands return stale/cached data and require parsing
    MCP provides structured JSON with real-time status
```

---

### Phase 1.0 : Peek + Commit-Pinned Tracking

**CRITICAL: Track pipeline for SPECIFIC commit SHA**

```yaml
peek_workflow:
  0_get_pushed_commit:
    action: "Get SHA of just-pushed commit"
    command: "git rev-parse HEAD"
    store: "pushed_commit_sha"
    critical: true

  1_pr_mr_info:
    action: "Récupérer info PR/MR"
    tools:
      github: mcp__github__get_pull_request
      gitlab: mcp__gitlab__get_merge_request
    verify: "head_sha == pushed_commit_sha"
    output: "pr_mr_number, head_sha, status, checks"

  2_find_pipeline:
    action: "Find pipeline triggered by THIS commit"
    github: |
      # Verify: check_run.head_sha == pushed_commit_sha
      mcp__github__pull_request_read(method="get")
    gitlab: |
      # Filter: pipeline.sha == pushed_commit_sha
      mcp__gitlab__list_pipelines(sha=pushed_commit_sha)

  3_validate_pipeline:
    action: "Abort if pipeline not found within 60s"
    timeout: 60s
    on_timeout: "ERROR: No pipeline triggered for commit {sha}"

  4_conflicts:
    action: "Vérifier les conflits"
    command: "git fetch && git merge-base..."
```

**Output Phase 1:**

```
═══════════════════════════════════════════════════════════════
  /git --merge - Pipeline Tracking
═══════════════════════════════════════════════════════════════

  Commit: abc1234 (verified)
  PR: #42

  Pipeline found:
    ├─ ID: 12345
    ├─ SHA: abc1234 ✓ (matches pushed commit)
    ├─ Triggered: 15s ago
    └─ Status: running

═══════════════════════════════════════════════════════════════
```

---

### Phase 2.0 : Job-Level Status Parsing (CRITICAL)

**Parse EACH job individually, not overall status:**

```yaml
status_parsing:
  github:
    statuses:
      success: ["success", "neutral"]
      pending: ["queued", "in_progress", "waiting", "pending"]
      failure: ["failure", "action_required", "timed_out"]
      cancelled: ["cancelled", "stale"]
      skipped: ["skipped"]

    aggregation_rule: |
      # CRITICAL: A single failed job = PIPELINE FAILED
      pipeline_success = ALL jobs in [success, skipped, neutral]
      pipeline_failure = ANY job in [failure, cancelled, timed_out]
      pipeline_pending = ANY job in [pending, queued, in_progress]

      # DO NOT report success if any job failed!

  gitlab:
    statuses:
      success: ["success", "manual"]
      pending: ["created", "waiting_for_resource", "preparing", "pending", "running"]
      failure: ["failed"]
      cancelled: ["canceled"]
      skipped: ["skipped"]

job_by_job_output:
  format: |
    ═══════════════════════════════════════════════════════════════
      CI Status - Commit {sha}
    ═══════════════════════════════════════════════════════════════

      Pipeline: #{id} (triggered {time_ago})
      Branch:   {branch}
      Commit:   {sha} ✓ (verified)

      Jobs:
        ├─ lint      : ✓ passed (45s)
        ├─ build     : ✓ passed (1m 23s)
        ├─ test      : ✗ FAILED (2m 15s)    <-- FAILED
        └─ deploy    : ⊘ skipped

      Overall: ✗ FAILED (1 job failed)

    ═══════════════════════════════════════════════════════════════
```

---

### Phase 3.0 : CI Monitoring avec Backoff Exponentiel et Hard Timeout

**ABSOLUTE LIMIT: 10 minutes / 30 polls**

```yaml
ci_monitoring:
  description: "Suivi intelligent du statut CI avec polling adaptatif"

  #---------------------------------------------------------------------------
  # CONFIGURATION
  #---------------------------------------------------------------------------
  config:
    initial_interval: 10s          # Intervalle initial
    max_interval: 120s             # Plafonné à 2 minutes
    backoff_multiplier: 1.5        # 10s → 15s → 22s → 33s → 50s → 75s → 112s → 120s
    jitter_percent: 20             # +/- 20% aléatoire (évite thundering herd)
    timeout: 600s                  # 10 minutes HARD timeout total
    max_poll_attempts: 30          # Limite de sécurité

  #---------------------------------------------------------------------------
  # STRATÉGIE DE POLLING (MCP-ONLY - NO CLI FALLBACK)
  #---------------------------------------------------------------------------
  polling_strategy:
    github:
      tool: mcp__github__get_pull_request
      params:
        pull_number: "{pr_number}"
      response_fields: ["state", "statuses[]", "check_runs[]"]
      # NO FALLBACK - CLI FORBIDDEN

    gitlab:
      tool: mcp__gitlab__list_pipelines
      params:
        project_id: "{project_id}"
        ref: "{branch}"
        per_page: 1
      response_fields: ["status", "id", "web_url"]
      # NO FALLBACK - CLI FORBIDDEN

  #---------------------------------------------------------------------------
  # ALGORITHME DE BACKOFF EXPONENTIEL
  #---------------------------------------------------------------------------
  backoff_algorithm:
    pseudocode: |
      interval = initial_interval
      elapsed = 0
      attempt = 0

      WHILE elapsed < timeout AND attempt < max_poll_attempts:
        status = poll_ci_status()  # MCP ONLY

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
  # ON TIMEOUT
  #---------------------------------------------------------------------------
  on_timeout:
    action: "ABORT immediately"
    output: |
      ═══════════════════════════════════════════════════════════════
        ⛔ Pipeline Timeout
      ═══════════════════════════════════════════════════════════════

        Waited: 10 minutes
        Polls:  30 attempts
        Status: Still pending

        This usually means:
        - Pipeline is stuck
        - Pipeline was cancelled externally
        - Wrong pipeline being monitored

        Actions:
        1. Check pipeline manually: {pipeline_url}
        2. Re-run: /git --merge
        3. Force: /git --merge --skip-ci (if CI is broken)

      ═══════════════════════════════════════════════════════════════

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

**Output Phase 2.5 :**

```
═══════════════════════════════════════════════════════════════
  /git --merge - CI Monitoring (Phase 2.5)
═══════════════════════════════════════════════════════════════

  PR/MR    : #42 (feat/add-auth)
  Platform : GitHub
  Timeout  : 10 minutes (HARD LIMIT)

  Polling CI status (MCP-ONLY)...
    [10:30:15] Poll #1: pending (10s elapsed, next in 10s)
    [10:30:27] Poll #2: running (22s elapsed, next in 15s)
    [10:30:45] Poll #3: running (40s elapsed, next in 22s)
    [10:31:12] Poll #4: running (67s elapsed, next in 33s)
    [10:31:50] ✓ CI PASSED (95s)

  Job-level verification:
    ├─ lint: ✓ passed (45s)
    ├─ build: ✓ passed (1m 23s)
    └─ test: ✓ passed (2m 45s)

  Proceeding to Phase 3...

═══════════════════════════════════════════════════════════════
```

---

### Phase 4.0 : Error Log Extraction (on failure)

**When pipeline fails, extract actionable information:**

```yaml
error_extraction:
  step_1_identify:
    action: "Get list of failed jobs"
    output: "[job_name, job_id, failure_reason]"

  step_2_parse_error:
    patterns:
      lint_error:
        - "eslint.*error"
        - "golangci-lint"
        - "clippy::"
        - "ruff.*error"
      build_error:
        - "cannot find module"
        - "compilation failed"
        - "cargo build.*error"
        - "tsc.*error"
      test_error:
        - "FAIL.*test"
        - "AssertionError"
        - "--- FAIL:"
        - "pytest.*FAILED"
      security_error:
        - "CRITICAL.*vulnerability"
        - "CVE-"
        - "HIGH.*severity"

  step_3_generate_debug_plan:
    output: |
      ═══════════════════════════════════════════════════════════════
        Pipeline Failed - Debug Plan
      ═══════════════════════════════════════════════════════════════

        Failed Job: {job_name}
        Error Type: {error_type}
        Exit Code:  {exit_code}

        Error Summary:
        ┌─────────────────────────────────────────────────────────────
        │ {error_excerpt_20_lines}
        └─────────────────────────────────────────────────────────────

        Suggested Actions:
        1. {action_1_based_on_error_type}
        2. {action_2_based_on_error_type}
        3. Run locally: {local_command}

        Next Step: Run `/plan debug {error_type}` to investigate

      ═══════════════════════════════════════════════════════════════
```

---

### Phase 5.0 : Auto-fix Loop avec Catégories d'Erreurs

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
          ci_status = poll_ci_with_backoff()  # Re-use Phase 2.5

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

**Output Phase 4 (Auto-fix Success) :**

```
═══════════════════════════════════════════════════════════════
  /git --merge - Auto-fix Loop (Phase 4)
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

  Proceeding to Phase 5 (Merge)...

═══════════════════════════════════════════════════════════════
```

**Output Phase 4 (Security Block) :**

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

### Phase 6.0 : Synthesize (Merge & Cleanup)

```yaml
merge_workflow:
  1_final_verify:
    action: "Verify ALL jobs passed (job-level check)"
    tools:
      github: mcp__github__get_pull_request
      gitlab: mcp__gitlab__get_merge_request
    condition: "ALL check_runs.conclusion == 'success'"

  2_merge:
    tools:
      github: mcp__github__merge_pull_request
      gitlab: mcp__gitlab__merge_merge_request
    method: "squash"

  3_cleanup:
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

  CI (job-level verification):
    ├─ lint      : ✓ passed
    ├─ build     : ✓ passed
    ├─ test      : ✓ passed
    └─ security  : ✓ passed

  Total CI Time: 2m 34s

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
| Skip Phase 3.8 (Context) | ❌ **INTERDIT** | CLAUDE.md doivent refléter les changements |
| Skip Phase 2 (CI Polling) | ❌ **INTERDIT** | CI validation obligatoire |
| Merge automatique sans CI | ❌ **INTERDIT** | Qualité code |
| Push sur main/master | ❌ **INTERDIT** | Branche protégée |
| Force merge si CI échoue x3 | ❌ **INTERDIT** | Limite tentatives |
| Push sans --force-with-lease | ❌ **INTERDIT** | Sécurité |
| Mentions IA dans commits | ❌ **INTERDIT** | Discrétion |
| Commit sans identité validée | ❌ **INTERDIT** | Traçabilité |
| CLI for CI status | ❌ **INTERDIT** | MCP-ONLY policy |
| Report success if ANY job failed | ❌ **INTERDIT** | Job-level parsing |
| Wait > 10 min for pipeline | ❌ **INTERDIT** | Hard timeout |
| Monitor wrong commit's pipeline | ❌ **INTERDIT** | Commit-pinned tracking |

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

### CLI Commands FORBIDDEN for CI Monitoring

```yaml
forbidden_cli:
  github:
    - "gh pr checks"
    - "gh run view"
    - "gh run list"
    - "gh api repos/.../check-runs"
  gitlab:
    - "glab ci status"
    - "glab ci view"
    - "glab pipeline status"
  generic:
    - "curl *api.github.com*"
    - "curl *gitlab.com/api*"

required_mcp:
  github: "mcp__github__get_pull_request, mcp__github__pull_request_read"
  gitlab: "mcp__gitlab__list_pipelines, mcp__gitlab__list_pipeline_jobs"
```

### Parallélisation légitime

| Élément | Parallèle? | Raison |
|---------|------------|--------|
| Pré-commit checks (lint+test+build) | ✅ Parallèle | Indépendants |
| Language checks (Go+Rust+Node) | ✅ Parallèle | Indépendants |
| CI polling + conflict check | ✅ Parallèle | Indépendants |
| Opérations git (branch→commit→push→PR) | ❌ Séquentiel | Chaîne de dépendances |
| Tentatives auto-fix | ❌ Séquentiel | Dépend du résultat CI |
| CI checks en attente | ❌ Séquentiel | Attendre résultat |
| Pipeline polling | ❌ Séquentiel | État change entre polls |
