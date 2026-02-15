---
name: do
description: |
  Iterative task execution loop with RLM decomposition.
  Transforms a task into a persistent loop with automatic iteration.
  The agent keeps going, fixing its own mistakes, until success criteria are met.
  Also executes approved plans from /plan (auto-detected).
allowed-tools:
  - "Read(**/*)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "Write(**/*)"
  - "Edit(**/*)"
  - "Bash(*)"
  - "Task(*)"
  - "TaskCreate(*)"
  - "TaskUpdate(*)"
  - "TaskList(*)"
  - "TaskGet(*)"
  - "AskUserQuestion(*)"
  - "mcp__codacy__codacy_cli_analyze(*)"
---

# /do - Iterative Task Loop (RLM Architecture)

$ARGUMENTS

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Use `grepai_trace_callers`/`grepai_trace_callees` for impact analysis.
Fallback to Grep ONLY for exact string matches or regex patterns.

---

## Overview

Boucle itérative utilisant **Recursive Language Model** decomposition :

- **Peek** - Scan rapide avant exécution
- **Decompose** - Diviser la tâche en sous-objectifs
- **Parallelize** - Validations parallèles (test, lint, build)
- **Synthesize** - Rapport consolidé

**Principe** : Itérer jusqu'au succès plutôt que viser la perfection.

---

## --help

```
═══════════════════════════════════════════════════════════════
  /do - Iterative Task Loop (RLM)
═══════════════════════════════════════════════════════════════

  DESCRIPTION
    Transforme une tâche en boucle persistante d'itérations.
    L'agent continue jusqu'à ce que les critères de succès soient
    atteints ou que la limite d'itérations soit atteinte.

    Si un plan approuvé existe (via /plan), l'exécute automatiquement
    sans poser les questions interactives.

  USAGE
    /do <task>              Lance le workflow interactif
    /do                     Exécute le plan approuvé (si existant)
    /do --help              Affiche cette aide

  RLM PATTERNS
    1. Plan    - Détection plan approuvé (skip questions si oui)
    2. Secret   - Découverte secrets 1Password
    3. Questions - Configuration interactive (si pas de plan)
    4. Peek     - Scan du codebase + git conflict check
    5. Decompose - Division en sous-objectifs mesurables
    6. Loop     - Validations simultanées (test/lint/build)
    7. Synthesize - Rapport consolidé par itération

  EXEMPLES
    /do "Migrer les tests Jest vers Vitest"
    /do "Ajouter des tests pour couvrir src/utils à 80%"
    /do                     # Exécute le plan de /plan

  GARDE-FOUS
    - Max 50 itérations (défaut: 10)
    - Critères de succès MESURABLES uniquement
    - Revue du diff obligatoire avant merge
    - Git conflict check avant modifications

═══════════════════════════════════════════════════════════════
```

**SI `$ARGUMENTS` contient `--help`** : Afficher l'aide ci-dessus et STOP.

---

## Phase 1.0 : Détection de Plan Approuvé

**TOUJOURS exécuter en premier. Vérifie si /plan a été utilisé.**

```yaml
plan_detection:
  check: "Existe-t-il un plan approuvé dans le contexte ?"

  sources:
    - "Conversation récente (plan validé par utilisateur)"
    - "Mémoire de session Claude"

  detection_signals:
    - "User a dit 'oui', 'ok', 'go', 'approuvé' après un /plan"
    - "Plan structuré avec steps numérotées visible"
    - "ExitPlanMode a été appelé avec succès"

  if_plan_found:
    mode: "PLAN_EXECUTION"
    actions:
      - "Extraire: title, steps[], scope, files[]"
      - "Skip Phase 0 (questions interactives)"
      - "Utiliser steps du plan comme sous-objectifs"
      - "Critères = plan terminé + tests/lint/build passent"

  if_no_plan:
    mode: "ITERATIVE"
    actions:
      - "Continuer vers Phase 0 (questions)"
```

**Output Phase 1.0 (plan détecté) :**

```
═══════════════════════════════════════════════════════════════
  /do - Plan Detection
═══════════════════════════════════════════════════════════════

  ✓ Approved plan detected!

  Plan   : "Add JWT authentication to API"
  Steps  : 4
  Scope  : src/auth/, src/middleware/
  Files  : 6 to modify, 2 to create

  Mode: PLAN_EXECUTION (skipping interactive questions)

  Proceeding to Phase 4.0 (Peek)...

═══════════════════════════════════════════════════════════════
```

**Output Phase 1.0 (pas de plan) :**

```
═══════════════════════════════════════════════════════════════
  /do - Plan Detection
═══════════════════════════════════════════════════════════════

  No approved plan found.

  Mode: ITERATIVE (interactive questions required)

  Proceeding to Phase 3.0 (Questions)...

═══════════════════════════════════════════════════════════════
```

---

## Phase 2.0 : Secret Discovery (1Password)

**Verifier si des secrets sont disponibles pour ce projet :**

```yaml
secret_discovery:
  trigger: "ALWAYS (avant Phase 0)"
  blocking: false  # Informatif seulement

  1_check_available:
    condition: "command -v op && test -n $OP_SERVICE_ACCOUNT_TOKEN"
    on_failure: "Skip silently (1Password not configured)"

  2_resolve_path:
    action: "Extraire org/repo depuis git remote origin"
    command: |
      REMOTE=$(git config --get remote.origin.url)
      # Extract org/repo from HTTPS, SSH, or token-embedded URLs
      PROJECT_PATH=$(echo "${REMOTE%.git}" | grep -oP '[:/]\K[^/]+/[^/]+$')

  3_list_project_secrets:
    action: "Lister les secrets du projet"
    command: |
      op item list --vault='$VAULT_ID' --format=json \
        | jq -r '.[] | select(.title | startswith("'$PROJECT_PATH'/")) | .title'
    extract: "Supprimer le prefix pour garder les noms de cles"

  4_check_task_needs:
    action: "Si la tache mentionne secret/token/credential/password/API key"
    match_keywords: ["secret", "token", "credential", "password", "api key", "api_key", "auth"]
    if_match_and_secrets_exist:
      output: |
        ═══════════════════════════════════════════════════════════════
          /do - Secrets Available
        ═══════════════════════════════════════════════════════════════

          Project: {PROJECT_PATH}
          Available secrets in 1Password:
            ├─ DB_PASSWORD
            ├─ API_KEY
            └─ JWT_SECRET

          Use /secret --get <key> to retrieve a value
          These may help with the current task.

        ═══════════════════════════════════════════════════════════════
    if_no_secrets:
      output: "(no project secrets in 1Password, continuing...)"
```

---

## Phase 3.0 : Questions Interactives (SI PAS DE PLAN)

**Poser ces 4 questions UNIQUEMENT si aucun plan approuvé n'est détecté :**

### Question 1 : Type de tâche

```yaml
AskUserQuestion:
  questions:
    - question: "Quel type de tâche veux-tu accomplir ?"
      header: "Type"
      multiSelect: false
      options:
        - label: "Refactor/Migration (Recommended)"
          description: "Migrer un framework, refactorer du code existant"
        - label: "Test Coverage"
          description: "Ajouter des tests pour atteindre un seuil de couverture"
        - label: "Standardisation"
          description: "Appliquer des patterns cohérents (erreurs, style)"
        - label: "Greenfield"
          description: "Créer un nouveau projet/module de zéro"
```

### Question 2 : Itérations max

```yaml
AskUserQuestion:
  questions:
    - question: "Combien d'itérations maximum autoriser ?"
      header: "Iterations"
      multiSelect: false
      options:
        - label: "10 (Recommended)"
          description: "Suffisant pour la plupart des tâches"
        - label: "20"
          description: "Pour les tâches moyennement complexes"
        - label: "30"
          description: "Pour les migrations/refactorings majeurs"
        - label: "50"
          description: "Pour les projets greenfield complets"
```

### Question 3 : Critères de succès

```yaml
AskUserQuestion:
  questions:
    - question: "Quels critères de succès utiliser ?"
      header: "Critères"
      multiSelect: true
      options:
        - label: "Tests passent (Recommended)"
          description: "Tous les tests unitaires doivent être verts"
        - label: "Lint propre"
          description: "Aucune erreur de linter"
        - label: "Build réussit"
          description: "La compilation doit fonctionner"
        - label: "Couverture >= X%"
          description: "Seuil de couverture à atteindre"
```

### Question 4 : Scope

```yaml
AskUserQuestion:
  questions:
    - question: "Quel scope pour cette tâche ?"
      header: "Scope"
      multiSelect: false
      options:
        - label: "Dossier src/ (Recommended)"
          description: "Tout le code source"
        - label: "Fichiers spécifiques"
          description: "Je vais préciser les fichiers"
        - label: "Tout le projet"
          description: "Inclut tests, docs, config"
        - label: "Personnalisé"
          description: "Je vais spécifier un chemin"
```

---

## Phase 4.0 : Peek (RLM Pattern)

**Scan rapide AVANT toute modification :**

```yaml
peek_workflow:
  0_git_check:
    action: "Vérifier l'état git (conflict detection)"
    tools: [Bash]
    command: "git status --porcelain"
    checks:
      - "Pas de merge/rebase en cours"
      - "Fichiers cibles pas déjà modifiés (warning si oui)"
    on_conflict:
      action: "Warning + continuer (pas bloquant)"
      message: "⚠ Uncommitted changes detected on target files"

  1_structure:
    action: "Scanner la structure du scope"
    tools: [Glob]
    patterns:
      - "src/**/*.{ts,js,go,py,rs}"
      - "tests/**/*"
      - "package.json | go.mod | Cargo.toml | pyproject.toml"

  2_patterns:
    action: "Identifier les patterns existants"
    tools: [Grep]
    searches:
      - "class.*Factory" → Factory pattern
      - "getInstance" → Singleton
      - "describe|test|it" → Tests existants

  3_stack_detect:
    action: "Détecter le stack technique"
    checks:
      - "package.json → Node.js/npm"
      - "go.mod → Go"
      - "Cargo.toml → Rust"
      - "pyproject.toml → Python"
    output: "test_command, lint_command, build_command"
```

**Output Phase 4.0 :**

```
═══════════════════════════════════════════════════════════════
  /do - Peek Analysis
═══════════════════════════════════════════════════════════════

  Git Status:
    ✓ Working tree clean (or: ⚠ 3 uncommitted changes)

  Scope      : src/
  Files      : 47 source files, 23 test files
  Stack      : Node.js (TypeScript)

  Patterns detected:
    ✓ Factory pattern (3 occurrences)
    ✓ Repository pattern (2 occurrences)
    ✓ Jest test suite (23 files)

  Commands:
    Test  : npm test
    Lint  : npm run lint
    Build : npm run build

═══════════════════════════════════════════════════════════════
```

---

## Phase 5.0 : Decompose (RLM Pattern)

**Diviser la tâche en sous-objectifs mesurables :**

```yaml
decompose_workflow:
  1_analyze_task:
    action: "Extraire les objectifs de la tâche"
    example:
      task: "Migrer Jest vers Vitest"
      objectives:
        - "Remplacer dépendances Jest par Vitest"
        - "Mettre à jour la config de test"
        - "Adapter les imports dans les fichiers de test"
        - "Corriger les APIs incompatibles"
        - "Vérifier que tous les tests passent"

  2_prioritize:
    action: "Ordonner par dépendance"
    principle: "Smallest change first"

  3_create_todos:
    action: "Initialiser TaskCreate avec les sous-objectifs"
```

**Output Phase 5.0 :**

```
═══════════════════════════════════════════════════════════════
  /do - Task Decomposition
═══════════════════════════════════════════════════════════════

  Task: "Migrer les tests Jest vers Vitest"

  Sub-objectives (ordered):
    1. [DEPS] Remplacer jest → vitest dans package.json
    2. [CONFIG] Créer vitest.config.ts
    3. [IMPORTS] Adapter imports jest → vitest (23 fichiers)
    4. [COMPAT] Corriger APIs incompatibles
    5. [VERIFY] Tous les tests passent

  Strategy: Sequential with parallel validation

═══════════════════════════════════════════════════════════════
```

---

## Phase 6.0 : Boucle Principale

```
┌──────────────────────────────────────────────────────────────┐
│  LOOP: while (iteration < max && !success)                   │
│                                                              │
│    1. Peek  → Lire l'état actuel                             │
│    2. Apply → Modifications minimales                        │
│    3. Parallelize → Validations simultanées                  │
│    4. Synthesize → Analyser résultats                        │
│    5. Décision → SUCCESS | CONTINUE | ABORT                  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Step 3.1 : Peek itératif

```yaml
peek_iteration:
  action: "Lire l'état actuel avant modification"
  inputs:
    - "Fichiers modifiés précédemment"
    - "Erreurs de la dernière validation"
    - "Progression vers les sous-objectifs"
```

### Step 3.2 : Apply (modifications minimales)

```yaml
apply_iteration:
  principle: "Smallest change that moves toward success"
  actions:
    - "Modifier uniquement les fichiers nécessaires"
    - "Suivre les patterns existants du projet"
    - "Ne pas sur-ingénierer"
  tracking:
    - "Ajouter chaque fichier modifié à la liste"
```

### Step 3.3 : Parallelize (validations simultanées)

**Lancer les validations en PARALLÈLE via Task agents :**

```yaml
parallel_validation:
  agents:
    - task: "Run tests"
      command: "{test_command}"
      output: "test_result"

    - task: "Run linter"
      command: "{lint_command}"
      output: "lint_result"

    - task: "Run build"
      command: "{build_command}"
      output: "build_result"

  mode: "PARALLEL (single message, multiple Task calls)"
```

**IMPORTANT** : Lancer les 3 validations dans UN SEUL message.

### Step 3.4 : Synthesize (analyse des résultats)

```yaml
synthesize_iteration:
  collect:
    - "test_result.exit_code"
    - "test_result.passed / test_result.total"
    - "lint_result.error_count"
    - "build_result.exit_code"

  evaluate:
    all_success:
      condition: "test_exit == 0 && lint_exit == 0 && build_exit == 0"
      action: "EXIT with success report"

    partial_success:
      condition: "Some criteria met, some not"
      action: "CONTINUE with focused fixes"

    no_progress:
      condition: "Same errors 3 iterations in a row"
      action: "ABORT with blocker analysis"

  output: "Iteration summary"
```

**Output par itération :**

```
═══════════════════════════════════════════════════════════════
  Iteration 3/10
═══════════════════════════════════════════════════════════════

  Modified: 5 files

  Validation (parallel):
    ├─ Tests : 18/23 PASS (5 failing)
    ├─ Lint  : 2 errors
    └─ Build : SUCCESS

  Analysis:
    - 5 tests use jest.mock() incompatible avec vitest
    - 2 erreurs lint sur imports non utilisés

  Decision: CONTINUE → Focus on jest.mock migration

═══════════════════════════════════════════════════════════════
```

---

## Phase 7.0 : Synthèse Finale

### Rapport de succès

```
═══════════════════════════════════════════════════════════════
  /do - Task Completed Successfully
═══════════════════════════════════════════════════════════════

  Task       : {original_task}
  Iterations : {n}/{max}

  ✓ All Criteria Met:
    - Tests: 23/23 PASS
    - Lint: 0 errors
    - Build: SUCCESS

  Files Modified ({count}):
    - package.json (+3, -3)
    - vitest.config.ts (+25, -0)
    - src/**/*.test.ts (23 files)

  Decomposition Results:
    ✓ [DEPS] Replaced dependencies
    ✓ [CONFIG] Created vitest config
    ✓ [IMPORTS] Adapted 23 test files
    ✓ [COMPAT] Fixed mock APIs
    ✓ [VERIFY] All tests pass

═══════════════════════════════════════════════════════════════
  IMPORTANT: Review the diff before merging!
  → git diff HEAD~{n}
═══════════════════════════════════════════════════════════════
```

### Rapport d'échec

```
═══════════════════════════════════════════════════════════════
  /do - Task Stopped (Max Iterations / Blocker)
═══════════════════════════════════════════════════════════════

  Task       : {original_task}
  Iterations : {n}/{max}
  Reason     : {MAX_REACHED | BLOCKER_DETECTED | CIRCULAR_FIX}

  ✗ Criteria NOT Met:
    - Tests: 20/23 PASS (3 failing)
    - Lint: 0 errors

  Blockers Identified:
    1. tests/api.test.ts:45 - Cannot mock external service
    2. tests/db.test.ts:78 - Database connection required

  Decomposition Status:
    ✓ [DEPS] Replaced dependencies
    ✓ [CONFIG] Created vitest config
    ✓ [IMPORTS] Adapted 23 test files
    ✗ [COMPAT] 3 mocks incompatibles
    ✗ [VERIFY] Tests failing

  Suggested Next Steps:
    1. Review failing tests manually
    2. Consider mocking strategy for external services
    3. Re-run with narrower scope

═══════════════════════════════════════════════════════════════
```

---

## Anti-patterns (Détection automatique)

| Pattern | Symptôme | Action |
|---------|----------|--------|
| **Circular fix** | Même fichier modifié 3+ fois | ABORT + alerte |
| **No progress** | 0 amélioration sur 3 itérations | ABORT + diagnostic |
| **Scope creep** | Fichiers hors scope modifiés | Rollback + warning |
| **Overbaking** | Changements incohérents après 15+ iter | ABORT + rapport |

---

## TaskCreate Integration

```yaml
task_pattern:
  phase_0:
    - TaskCreate: { subject: "Configuration questions", activeForm: "Asking configuration questions" }
      → TaskUpdate: { status: "completed" }

  phase_1:
    - TaskCreate: { subject: "Peek: Analyze codebase", activeForm: "Analyzing codebase" }
      → TaskUpdate: { status: "in_progress" }

  phase_2:
    - TaskCreate: { subject: "{sub_objective_1}", activeForm: "Working on {sub_objective_1}" }
    - TaskCreate: { subject: "{sub_objective_2}", activeForm: "Working on {sub_objective_2}" }

  per_iteration:
    on_start: "TaskUpdate → status: in_progress"
    on_complete: "TaskUpdate → status: completed"
    on_blocked: "TaskCreate new blocker task"
    on_success: "TaskUpdate all → completed"
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Status | Raison |
|--------|--------|--------|
| Skip Phase 1.0 (Plan detect) | ❌ **INTERDIT** | Vérifier si plan existe |
| Skip Phase 3.0 sans plan | ❌ **INTERDIT** | Questions requises |
| Skip Phase 4.0 (Peek) | ❌ **INTERDIT** | Contexte + git check |
| Ignorer max_iterations | ❌ **INTERDIT** | Boucle infinie |
| Critères subjectifs ("joli", "clean") | ❌ **INTERDIT** | Non mesurable |
| Modifier .claude/ ou .devcontainer/ | ❌ **INTERDIT** | Fichiers protégés |
| Plus de 50 itérations | ❌ **INTERDIT** | Limite de sécurité |

### Parallélisation légitime

| Élément | Parallèle? | Raison |
|---------|------------|--------|
| Boucle itérative (N → N+1) | ❌ Séquentiel | Itération dépend du résultat précédent |
| Checks par itération (lint+test+build) | ✅ Parallèle | Indépendants entre eux |
| Actions correctives | ❌ Séquentiel | Ordre logique requis |

---

## Exemples de prompts efficaces

### ✓ BON : Critères mesurables

```
/do "Migrer tous les tests Jest vers Vitest"
→ Critère: tous les tests passent avec Vitest

/do "Ajouter des tests pour src/utils avec couverture 80%"
→ Critère: coverage >= 80%

/do "Remplacer console.log par un logger structuré"
→ Critère: 0 console.log dans src/, lint propre
```

### ✗ MAUVAIS : Critères subjectifs

```
/do "Rendre le code plus propre"
→ "Propre" n'est pas mesurable

/do "Améliorer les performances"
→ Pas de métrique de benchmark définie
```

---

## Intégration avec /review (Cyclic Workflow)

**`/review --loop` génère des plans que `/do` exécute automatiquement.**

```yaml
review_integration:
  detection:
    trigger: "plan filename contains 'review-fixes-'"
    location: ".claude/plans/review-fixes-*.md"

  mode: "REVIEW_EXECUTION"

  workflow:
    1_load_plan:
      action: "Read .claude/plans/review-fixes-{timestamp}.md"
      extract:
        - findings: [{file, line, fix_patch, language, specialist}]
        - priorities: ["CRITICAL", "HIGH", "MEDIUM"]

    2_group_by_language:
      action: "Group findings by file extension"
      example:
        ".go": ["finding1", "finding2"]
        ".ts": ["finding3"]

    3_dispatch_to_specialists:
      mode: "parallel (by language)"
      for_each_language:
        agent: "developer-specialist-{lang}"
        prompt: |
          You are the {language} specialist.

          ## Findings to Fix
          {findings_json}

          ## Constraints
          - Apply fixes in priority order (CRITICAL → HIGH)
          - Use fix_patch as starting point
          - Verify fix doesn't introduce new issues
          - Follow repo conventions

          ## Output
          For each fix applied:
          - File modified
          - Lines changed
          - Brief explanation

    4_validate:
      action: "Run quick /review (no loop) on modified files"
      check:
        - "Were original issues from the plan fixed?"
        - "Were any new CRITICAL/HIGH issues introduced?"

    5_report:
      action: "Summary of fixes applied"
      format: |
        Files modified: {n}
        Findings fixed: CRIT={a}, HIGH={b}, MED={c}
        New issues: {new_count}

    6_return_to_review:
      condition: "Called from /review --loop"
      action: "Return control to /review for re-validation"
```

**Language-Specialist Routing:**

| Extension | Specialist Agent |
|-----------|------------------|
| `.go` | `developer-specialist-go` |
| `.py` | `developer-specialist-python` |
| `.java` | `developer-specialist-java` |
| `.ts`, `.js` | `developer-specialist-nodejs` |
| `.rs` | `developer-specialist-rust` |
| `.rb` | `developer-specialist-ruby` |
| `.ex`, `.exs` | `developer-specialist-elixir` |
| `.php` | `developer-specialist-php` |
| `.c`, `.h` | `developer-specialist-c` |
| `.cpp`, `.cc`, `.hpp` | `developer-specialist-cpp` |
| `.cs` | `developer-specialist-csharp` |
| `.kt`, `.kts` | `developer-specialist-kotlin` |
| `.swift` | `developer-specialist-swift` |
| `.r`, `.R` | `developer-specialist-r` |
| `.pl`, `.pm` | `developer-specialist-perl` |
| `.lua` | `developer-specialist-lua` |
| `.f90`, `.f95`, `.f03` | `developer-specialist-fortran` |
| `.adb`, `.ads` | `developer-specialist-ada` |
| `.cob`, `.cbl` | `developer-specialist-cobol` |
| `.pas`, `.dpr`, `.pp` | `developer-specialist-pascal` |
| `.vb` | `developer-specialist-vbnet` |
| `.m` (Octave) | `developer-specialist-matlab` |
| `.asm`, `.s` | `developer-specialist-assembly` |
| `.scala` | `developer-specialist-scala` |
| `.dart` | `developer-specialist-dart` |

---

## Intégration avec autres skills

| Avant /do | Après /do |
|-----------|-----------|
| `/plan` (optionnel mais recommandé) | `/git --commit` |
| `/review` (génère plan) | `/review` (re-validate si --loop) |
| `/search` (si research needed) | N/A |

**Workflow recommandé (plan standard) :**

```
/search "vitest migration from jest"  # Si besoin de recherche
    ↓
/plan "Migrer tests Jest"              # Planifier l'approche
    ↓
(user approves plan)                   # Validation humaine
    ↓
/do                                    # Détecte le plan → exécute
    ↓
(review diff)                          # Vérifier les changements
    ↓
/git --commit                          # Commiter + PR
```

**Workflow cyclique (avec /review --loop) :**

```
/review --loop 5                       # Analyse + génère plan fixes
    ↓
/do (auto-triggered)                   # Exécute via language-specialists
    ↓
/review (auto-triggered)               # Re-valide les corrections
    ↓
(loop until no CRITICAL/HIGH OR limit)
    ↓
/git --commit                          # Commiter les corrections
```

**Workflow rapide (sans plan) :**

```
/do "Fix tous les bugs de lint"        # Tâche simple + mesurable
    ↓
(iterations jusqu'à succès)
    ↓
/git --commit
```

**Note** : `/do` remplace `/apply`. Le skill `/apply` est déprécié.
