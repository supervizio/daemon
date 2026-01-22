---
name: warmup
description: |
  Project context pre-loading with RLM decomposition.
  Reads CLAUDE.md hierarchy using funnel strategy (root → leaves).
  Use when: starting a session, preparing for complex tasks, or updating documentation.
allowed-tools:
  - "Read(**/*)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "Write(**/*)"
  - "Edit(**/*)"
  - "Task(*)"
  - "TodoWrite(*)"
  - "Bash(git:*)"
---

# /warmup - Project Context Pre-loading (RLM Architecture)

$ARGUMENTS

---

## Overview

Préchargement du contexte projet avec patterns **RLM** :

- **Peek** - Découvrir la hiérarchie CLAUDE.md
- **Funnel** - Lecture en entonnoir (racine → feuilles)
- **Parallelize** - Analyse parallèle par domaine
- **Synthesize** - Contexte consolidé prêt à l'emploi

**Principe** : Charger le contexte → Être plus efficace sur les tâches

---

## Arguments

| Pattern | Action |
|---------|--------|
| (none) | Précharge tout le contexte projet |
| `--update` | Met à jour tous les CLAUDE.md + crée les manquants |
| `--dry-run` | Affiche ce qui serait mis à jour (avec --update) |
| `--help` | Affiche l'aide |

---

## --help

```
═══════════════════════════════════════════════════════════════
  /warmup - Project Context Pre-loading (RLM)
═══════════════════════════════════════════════════════════════

Usage: /warmup [options]

Options:
  (none)            Précharge le contexte complet
  --update          Met à jour + crée les CLAUDE.md manquants
  --dry-run         Affiche les changements (avec --update)
  --help            Affiche cette aide

Line Thresholds (CLAUDE.md):
  IDEAL       :   0-60 lignes (simple directories)
  ACCEPTABLE  :  61-80 lignes (medium complexity)
  WARNING     : 81-100 lignes (review recommended)
  CRITICAL    : 101-150 lignes (must be condensed)
  FORBIDDEN   :  150+ lignes (split required)

Exclusions (STRICT .gitignore respect):
  - vendor/, node_modules/, .git/
  - All patterns from .gitignore are honored
  - bin/, dist/, build/ (generated outputs)

RLM Patterns:
  1. Peek       - Découvrir la hiérarchie CLAUDE.md
  2. Funnel     - Lecture entonnoir (root → leaves)
  3. Parallelize - Analyse par domaine
  4. Synthesize - Contexte consolidé

Exemples:
  /warmup                       Précharge le contexte
  /warmup --update              Met à jour + crée manquants
  /warmup --update --dry-run    Preview des changements

Workflow:
  /warmup → /plan → /do → /git

═══════════════════════════════════════════════════════════════
```

**SI `$ARGUMENTS` contient `--help`** : Afficher l'aide ci-dessus et STOP.

---

## Mode Normal (Préchargement)

### Phase 1 : Peek (Découverte hiérarchie)

```yaml
peek_workflow:
  1_discover:
    action: "Découvrir tous les CLAUDE.md du projet"
    tool: Glob
    pattern: "**/CLAUDE.md"
    output: [claude_files]

  2_build_tree:
    action: "Construire l'arbre de contexte par profondeur"
    algorithm: |
      POUR chaque fichier:
        depth = path.count('/') - base.count('/')
      Trier par profondeur croissante
      depth 0: /CLAUDE.md (racine)
      depth 1: /src/CLAUDE.md, /.devcontainer/CLAUDE.md
      depth 2+: sous-dossiers

  3_detect_project:
    action: "Identifier le type de projet"
    tools: [Glob]
    patterns:
      - "go.mod" → Go
      - "package.json" → Node.js
      - "Cargo.toml" → Rust
      - "pyproject.toml" → Python
      - "*.tf" → Terraform
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════════════════════
  /warmup - Peek Analysis
═══════════════════════════════════════════════════════════════

  Project: /workspace
  Type   : <detected_type>

  CLAUDE.md Hierarchy (<n> files):
    depth 0 : /CLAUDE.md (project root)
    depth 1 : /.devcontainer/CLAUDE.md, /src/CLAUDE.md
    depth 2 : /.devcontainer/features/CLAUDE.md
    ...

  Strategy: Funnel (root → leaves, decreasing detail)

═══════════════════════════════════════════════════════════════
```

---

### Phase 2 : Funnel (Lecture en entonnoir)

```yaml
funnel_strategy:
  principle: "Lire du plus général au plus spécifique"

  levels:
    depth_0:
      files: ["/CLAUDE.md"]
      extract: ["project_rules", "structure", "workflow", "safeguards"]
      detail_level: "HIGH"

    depth_1:
      files: ["src/CLAUDE.md", ".devcontainer/CLAUDE.md"]
      extract: ["conventions", "key_files", "domain_rules"]
      detail_level: "MEDIUM"

    depth_2_plus:
      files: ["**/CLAUDE.md"]
      extract: ["specific_rules", "attention_points"]
      detail_level: "LOW"

  extraction_rules:
    include:
      - "Règles MANDATORY/ABSOLUES"
      - "Structure du dossier"
      - "Conventions spécifiques"
      - "GARDE-FOUS"
    exclude:
      - "Exemples de code complets"
      - "Détails d'implémentation"
      - "Longs blocs de code"
```

**Algorithme de lecture :**

```
POUR profondeur DE 0 À max_profondeur:
    fichiers = filtrer(claude_files, profondeur)

    PARALLÈLE POUR chaque fichier DANS fichiers:
        contenu = Read(fichier)
        contexte[fichier] = extraire_essentiel(contenu, niveau_détail)

    consolider(contexte, profondeur)
```

---

### Phase 3 : Parallelize (Analyse par domaine)

```yaml
parallel_analysis:
  mode: "PARALLEL (single message, 4 Task calls)"

  agents:
    - task: "source-analyzer"
      type: "Explore"
      scope: "src/"
      prompt: |
        Analyser la structure du code source:
        - Packages/modules principaux
        - Patterns architecturaux détectés
        - Points d'attention (TODO, FIXME, HACK)
        Return: {packages[], patterns[], attention_points[]}

    - task: "config-analyzer"
      type: "Explore"
      scope: ".devcontainer/"
      prompt: |
        Analyser la configuration DevContainer:
        - Features installées
        - Services configurés
        - MCP servers disponibles
        Return: {features[], services[], mcp_servers[]}

    - task: "test-analyzer"
      type: "Explore"
      scope: "tests/ OR **/*_test.go OR **/*.test.ts"
      prompt: |
        Analyser la couverture de tests:
        - Fichiers de test trouvés
        - Patterns de test utilisés
        Return: {test_files[], patterns[], coverage_estimate}

    - task: "docs-analyzer"
      type: "Explore"
      scope: ".claude/docs/"
      prompt: |
        Analyser la base de connaissances:
        - Catégories de patterns disponibles
        - Nombre de patterns par catégorie
        Return: {categories[], pattern_count}
```

**IMPORTANT** : Lancer les 4 agents dans UN SEUL message.

---

### Phase 4 : Synthesize (Contexte consolidé)

```yaml
synthesize_workflow:
  1_merge:
    action: "Fusionner les résultats des agents"
    inputs:
      - "context_tree (Phase 2)"
      - "source_analysis (Phase 3)"
      - "config_analysis (Phase 3)"
      - "test_analysis (Phase 3)"
      - "docs_analysis (Phase 3)"

  2_prioritize:
    action: "Prioriser les informations"
    levels:
      - CRITICAL: "Règles absolues, garde-fous, conventions obligatoires"
      - HIGH: "Structure projet, patterns utilisés, MCP disponibles"
      - MEDIUM: "Features, services, couverture tests"
      - LOW: "Détails spécifiques, points d'attention mineurs"

  3_format:
    action: "Formater le contexte pour session"
    output: "Session context ready"
```

**Output Final (Mode Normal) :**

```
═══════════════════════════════════════════════════════════════
  /warmup - Context Loaded Successfully
═══════════════════════════════════════════════════════════════

  Project: <project_name>
  Type   : <detected_type>

  Context Summary:
    ├─ CLAUDE.md files read: <n>
    ├─ Source packages: <n>
    ├─ Test files: <n>
    ├─ Design patterns: <n>
    └─ MCP servers: <n>

  Key Rules Loaded:
    ✓ MCP-FIRST: Always use MCP before CLI
    ✓ GREPAI-FIRST: Semantic search before Grep
    ✓ Code in /src: All code MUST be in /src
    ✓ SAFEGUARDS: Never delete .claude/ or .devcontainer/

  Attention Points Detected:
    ├─ <n> TODO items in src/
    ├─ <n> FIXME in config
    └─ <n> deprecated APIs flagged

  Ready for:
    → /plan <feature>
    → /review
    → /do <task>

═══════════════════════════════════════════════════════════════
```

---

## Mode --update (Mise à jour documentation)

### Phase 1 : Scan complet du code

```yaml
scan_workflow:
  0_load_gitignore:
    action: "Charger les patterns .gitignore"
    command: "cat /workspace/.gitignore 2>/dev/null"
    rule: "TOUS les patterns sont STRICTEMENT respectés"

  1_discover_code:
    action: "Scanner tous les fichiers de code (respectant .gitignore)"
    tools: [Bash, Glob]
    command: |
      # Utilise git ls-files pour respecter .gitignore
      git ls-files --cached --others --exclude-standard \
        '*.go' '*.ts' '*.py' '*.sh' '*.rs' '*.java'
    patterns:
      - "src/**/*.go"
      - "src/**/*.ts"
      - "src/**/*.py"
      - "**/*.sh"
    exclude_source: ".gitignore (STRICT)"
    always_excluded:
      - ".git/"

  2_extract_metadata:
    action: "Extraire les métadonnées par dossier"
    parallel_per_directory:
      - "Fonctions/types publics"
      - "Patterns utilisés"
      - "TODO/FIXME/HACK"
      - "Imports critiques"
      - "Éléments obsolètes"

  3_check_claude_files:
    action: "Vérifier cohérence avec CLAUDE.md existants"
    for_each: claude_files
    checks:
      - "Structure documentée vs structure réelle"
      - "Fichiers mentionnés existent encore"
      - "Conventions documentées respectées"
      - "Informations obsolètes à supprimer"
```

---

### Phase 1.5 : Création des CLAUDE.md manquants

**Comportement par défaut de --update** (pas une option séparée).

```yaml
create_missing_workflow:
  trigger: "Toujours exécuté avec --update"

  gitignore_respect:
    rule: "STRICT - Tout pattern .gitignore est honoré"
    implementation: |
      # Lire et parser .gitignore
      gitignore_patterns = parse_gitignore("/workspace/.gitignore")

      # Utiliser git ls-files pour lister uniquement les fichiers trackés
      tracked_dirs = git ls-files --directory | get_unique_dirs()

      # OU utiliser git check-ignore pour valider
      for dir in candidate_dirs:
        if git check-ignore -q "$dir":
          skip(dir)  # Ignoré par .gitignore

  scan_directories:
    action: "Trouver les dossiers sans CLAUDE.md (respectant .gitignore)"
    tool: Bash + Glob
    command: |
      # Liste uniquement les dossiers NON ignorés par git
      find /workspace -type d \
        -not -path '*/.git/*' \
        -exec sh -c 'git check-ignore -q "$1" 2>/dev/null || echo "$1"' _ {} \; \
        | while read dir; do
            # Vérifie si contient du code source
            if ls "$dir"/*.{go,ts,py,rs,java,sh,html,tf} 2>/dev/null | head -1 > /dev/null; then
              [ ! -f "$dir/CLAUDE.md" ] && echo "$dir"
            fi
          done

    include_criteria:
      code_files:
        - "*.go, *.ts, *.py, *.rs, *.java"
        - "*.sh (scripts)"
        - "*.html, *.css (web)"
        - "Dockerfile*, *.tf (infra)"

    exclude_sources:
      primary: ".gitignore (STRICT)"
      always_excluded:
        - ".git/"
        - "**/testdata/**"
        - "**/__pycache__/**"

  create_template:
    format: |
      # <Directory Name>

      ## Purpose
      TODO: Describe the purpose of this directory.

      ## Structure
      ```text
      <auto-generated tree>
      ```

      ## Key Files
      | File | Description |
      |------|-------------|
      | <files> | TODO |

    max_lines: 30  # Template minimal, enrichi ensuite

  output: |
    ═══════════════════════════════════════════════════════════
      /warmup --update - Phase 1.5: Missing CLAUDE.md
    ═══════════════════════════════════════════════════════════

    .gitignore patterns loaded: <n> patterns

    Directories without CLAUDE.md (not in .gitignore):
      ├─ /workspace/website/ (HTML/CSS detected)
      ├─ /workspace/api/ (Proto files detected)
      └─ /workspace/setup/scripts/ (Shell scripts detected)

    Skipped (in .gitignore):
      ├─ /workspace/vendor/ (gitignored)
      ├─ /workspace/node_modules/ (gitignored)
      └─ /workspace/bin/ (gitignored)

    Action: Create template CLAUDE.md for each?
      [Apply all] [Select individually] [Skip]

    ═══════════════════════════════════════════════════════════
```

**RÈGLE ABSOLUE : .gitignore est la source de vérité pour les exclusions.**

| Source d'exclusion | Priorité | Exemples |
|--------------------|----------|----------|
| `.gitignore` | **STRICTE** | vendor/, node_modules/, *.log |
| Toujours exclus | Hardcoded | .git/, testdata/, __pycache__/ |

**Heuristiques de création :**

| Contenu détecté | Créer CLAUDE.md? | Condition |
|-----------------|------------------|-----------|
| Code source (*.go, *.ts, *.py) | ✅ OUI | Si non gitignored |
| Scripts (*.sh) | ✅ OUI | Si non gitignored |
| Web assets (*.html, *.css) | ✅ OUI | Si non gitignored |
| Config infra (Dockerfile, *.tf) | ✅ OUI | Si non gitignored |
| Tout dossier gitignored | ❌ NON | .gitignore respecté |

---

### Phase 2 : Détection des obsolescences

```yaml
obsolete_detection:
  file_references:
    description: "Fichiers mentionnés dans CLAUDE.md mais supprimés"
    action: |
      POUR chaque CLAUDE.md:
        extraire les chemins de fichiers mentionnés
        vérifier que chaque fichier existe
        marquer comme obsolète si non trouvé

  structure_changes:
    description: "Structure de dossier changée"
    action: |
      POUR chaque CLAUDE.md avec section 'Structure':
        comparer la structure documentée vs réelle
        identifier les différences

  api_changes:
    description: "APIs/fonctions renommées ou supprimées"
    action: |
      utiliser grepai pour chercher les références
      si 0 résultat → possiblement obsolète

  deprecated_patterns:
    description: "Patterns dépréciés encore documentés"
    action: |
      vérifier les imports/usages dans le code
      comparer avec ce qui est documenté
```

---

### Phase 3 : Génération des mises à jour

```yaml
update_generation:
  for_each: directory_with_claude_md

  format: |
    # <Directory Name>

    ## Purpose
    <Description courte du rôle du dossier>

    ## Structure
    ```text
    <arborescence actuelle>
    ```

    ## Key Files
    | File | Description |
    |------|-------------|
    | <file> | <description> |

    ## Conventions
    - <convention 1>
    - <convention 2>

    ## Attention Points
    - <point d'attention détecté dans le code>

  constraints:
    max_lines: 100  # WARNING threshold
    critical_threshold: 150  # Must be condensed or split
    no_implementation_details: true
    no_obsolete_info: true
    maintain_existing_structure: true
```

---

### Phase 4 : Application des changements

```yaml
apply_workflow:
  dry_run:
    condition: "--dry-run flag present"
    action: "Afficher les différences sans modifier"
    output: |
      ═══════════════════════════════════════════════════════════
        /warmup --update --dry-run
      ═══════════════════════════════════════════════════════════

      Files to update:
        ├─ /src/CLAUDE.md
        │   - Remove: "<file>" (deleted)
        │   + Add: "<file>" (new)
        │
        └─ /.devcontainer/features/CLAUDE.md
            + Add: New feature detected

      Total: <n> files, <n> changes
      Run without --dry-run to apply.
      ═══════════════════════════════════════════════════════════

  interactive:
    condition: "No --dry-run flag"
    for_each_file:
      action: "Afficher diff et demander confirmation"
      tool: AskUserQuestion
      options:
        - "Apply this change"
        - "Skip this file"
        - "Edit manually"
        - "Apply all remaining"

    on_apply:
      action: "Écrire le fichier mis à jour"
      tool: Edit or Write
      backup: true

  validation:
    post_apply:
      - "Verify file lines: IDEAL(0-60), ACCEPTABLE(61-80), WARNING(81-100), CRITICAL(101-150)"
      - "Flag files > 150 lines as FORBIDDEN (must split)"
      - "Verify no obsolete references"
      - "Verify structure section matches reality"
```

### Phase 5 : GrepAI Config Update (Project-Specific Exclusions)

**Met à jour la configuration grepai avec les exclusions spécifiques au projet.**

```yaml
grepai_config_update:
  trigger: "Always executed with --update"
  config_path: "/workspace/.grepai/config.yaml"
  template_path: "/etc/grepai/config.yaml"

  workflow:
    1_detect_project_patterns:
      action: "Analyser les patterns spécifiques au projet"
      checks:
        - ".gitignore patterns non couverts par template"
        - "Dossiers générés dynamiquement (logs, cache)"
        - "Frameworks spécifiques (Next.js .next/, Nuxt .nuxt/)"

    2_compare_with_template:
      action: "Comparer config actuelle vs template"
      detect:
        - "Nouvelles exclusions à ajouter"
        - "Exclusions obsolètes à retirer"

    3_merge_exclusions:
      action: "Fusionner les exclusions"
      rules:
        - "Garder toutes les exclusions du template"
        - "Ajouter les exclusions projet-spécifiques"
        - "Marquer les ajouts avec commentaire # Project-specific"

    4_apply_config:
      action: "Écrire la config mise à jour"
      tool: Write
      backup: true

  project_detection:
    nextjs:
      detect: "next.config.{js,ts,mjs}"
      add: [".next", ".vercel"]
    nuxt:
      detect: "nuxt.config.{js,ts}"
      add: [".nuxt", ".output"]
    vite:
      detect: "vite.config.{js,ts}"
      add: [".vite"]
    turbo:
      detect: "turbo.json"
      add: [".turbo"]
    nx:
      detect: "nx.json"
      add: [".nx", "nx-cloud.env"]
    docker:
      detect: "docker-compose*.{yml,yaml}"
      add: [".docker"]
    terraform:
      detect: "*.tf"
      add: [".terraform", "*.tfstate*"]

  output: |
    ═══════════════════════════════════════════════════════════
      /warmup --update - Phase 5: GrepAI Config
    ═══════════════════════════════════════════════════════════

    Config: /workspace/.grepai/config.yaml

    Project patterns detected:
      ├─ Next.js → adding .next, .vercel
      └─ Terraform → adding .terraform

    Exclusions updated:
      + .next (Project-specific)
      + .vercel (Project-specific)
      + .terraform (Project-specific)

    ✓ grepai config updated

    ═══════════════════════════════════════════════════════════
```

---

**Output Final (Mode --update) :**

```
═══════════════════════════════════════════════════════════════
  /warmup --update - Documentation Updated
═══════════════════════════════════════════════════════════════

  Files analyzed: <n> source files, <n> CLAUDE.md

  Changes applied:
    ✓ /src/CLAUDE.md - Updated structure
    ✓ /src/handlers/CLAUDE.md - Removed obsolete refs
    ○ /tests/CLAUDE.md - Skipped (user choice)

  Obsolete items removed:
    - <obsolete_file> reference
    - <old_function> signature

  New attention points added:
    + <n> TODO items documented
    + <n> FIXME flagged

  GrepAI config:
    ✓ Project-specific exclusions added

  Validation:
    ✓ Line thresholds: 0 FORBIDDEN, 0 CRITICAL, 2 WARNING
    ✓ Structure sections match reality
    ✓ No broken file references

═══════════════════════════════════════════════════════════════
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Status | Raison |
|--------|--------|--------|
| Skip Phase 1 (Peek) | ❌ **INTERDIT** | Découverte hiérarchie obligatoire |
| Modifier .claude/commands/ | ❌ **INTERDIT** | Fichiers protégés |
| Supprimer CLAUDE.md | ❌ **INTERDIT** | Seule mise à jour autorisée |
| Ignorer .gitignore | ❌ **INTERDIT** | Source de vérité pour exclusions |
| Créer CLAUDE.md dans gitignored | ❌ **INTERDIT** | vendor/, node_modules/, etc. |
| CLAUDE.md > 150 lignes | ❌ **FORBIDDEN** | Doit être splitté |
| CLAUDE.md 101-150 lignes | 🔴 **CRITICAL** | Condensation obligatoire |
| CLAUDE.md 81-100 lignes | ⚠ **WARNING** | Révision recommandée |
| Lecture aléatoire | ❌ **INTERDIT** | Funnel (root→leaves) obligatoire |
| Détails d'implémentation | ❌ **INTERDIT** | Contexte, pas code |
| --update sans backup | ⚠ **WARNING** | Risque de perte |

**Seuils de lignes CLAUDE.md :**

```
┌────────────┬─────────┬───────────────────────────────────────┐
│   Niveau   │ Lignes  │             Action                    │
├────────────┼─────────┼───────────────────────────────────────┤
│ IDEAL      │ 0-60    │ ✅ Aucune action                      │
├────────────┼─────────┼───────────────────────────────────────┤
│ ACCEPTABLE │ 61-80   │ ✅ Dossier moyen, acceptable          │
├────────────┼─────────┼───────────────────────────────────────┤
│ WARNING    │ 81-100  │ ⚠️ Révision recommandée à la prochaine│
├────────────┼─────────┼───────────────────────────────────────┤
│ CRITICAL   │ 101-150 │ 🔴 Condensation obligatoire           │
├────────────┼─────────┼───────────────────────────────────────┤
│ FORBIDDEN  │ 150+    │ ❌ Doit être splitté ou restructuré   │
└────────────┴─────────┴───────────────────────────────────────┘
```

**Justification des seuils :**

| Critère | 100 lignes (WARNING) | 150 lignes (CRITICAL) |
|---------|----------------------|-----------------------|
| Temps lecture | ~5 min | ~7-8 min |
| Tokens LLM | ~1000 | ~1500 |
| Flexibilité | Projets volumineux OK | Limite absolue |

**Quand 150+ lignes ?** → Le dossier doit être splitté en sous-dossiers avec leurs propres CLAUDE.md.

---

## Intégration Workflow

```
/warmup                     # Précharger contexte
    ↓
/plan "feature X"           # Planifier avec contexte
    ↓
/do                         # Exécuter le plan
    ↓
/warmup --update            # Mettre à jour doc
    ↓
/git --commit               # Commiter les changements
```

**Intégration avec autres skills :**

| Avant /warmup | Après /warmup |
|---------------|---------------|
| Container start | /plan, /review, /do |
| /init | Toute tâche complexe |

---

## Design Patterns Applied

| Pattern | Category | Usage |
|---------|----------|-------|
| Cache-Aside | Cloud | Vérifier cache avant chargement |
| Lazy Loading | Performance | Charger par phases (funnel) |
| Progressive Disclosure | DevOps | Détail croissant par profondeur |

**Références :**
- `.claude/docs/cloud/cache-aside.md`
- `.claude/docs/performance/lazy-load.md`
- `.claude/docs/devops/feature-toggles.md`
