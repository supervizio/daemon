---
name: update
description: |
  DevContainer Environment Update from official template.
  Updates features, hooks, commands, and settings from kodflow/devcontainer-template.
  Use when: syncing local devcontainer with latest template improvements.
allowed-tools:
  - "Bash(curl:*)"
  - "Bash(git:*)"
  - "Read(**/*)"
  - "Write(.devcontainer/**/*)"
  - "WebFetch(*)"
  - "Task(*)"
---

# Update - DevContainer Environment Update

$ARGUMENTS

---

## Description

Met à jour l'environnement DevContainer depuis le template officiel.

**Composants mis à jour :**

- **Features** - Language features et leurs RULES.md
- **Hooks** - Scripts Claude (format, lint, security, etc.)
- **Commands** - Commandes slash (/git, /search)
- **p10k** - Configuration Powerlevel10k
- **Settings** - Configuration Claude

**Source** : `github.com/kodflow/devcontainer-template`

---

## Arguments

| Pattern | Action |
|---------|--------|
| (none) | Mise à jour complète |
| `--check` | Vérifie les mises à jour disponibles |
| `--component <name>` | Met à jour un composant spécifique |
| `--help` | Affiche l'aide |

### Composants disponibles

| Composant | Chemin |
|-----------|--------|
| `features` | `.devcontainer/features/languages/` |
| `hooks` | `.devcontainer/images/.claude/scripts/` |
| `commands` | `.devcontainer/images/.claude/commands/` |
| `p10k` | `.devcontainer/images/.p10k.zsh` |
| `settings` | `.devcontainer/images/.claude/settings.json` |

---

## --help

```
═══════════════════════════════════════════════
  /update - DevContainer Environment Update
═══════════════════════════════════════════════

Usage: /update [options]

Options:
  (none)              Mise à jour complète
  --check             Vérifie les mises à jour
  --component <name>  Met à jour un composant
  --help              Affiche cette aide

Composants:
  features    Language features (RULES.md)
  hooks       Scripts Claude (format, lint...)
  commands    Commandes slash (/git, /search)
  p10k        Powerlevel10k config
  settings    Claude settings.json

Exemples:
  /update                       Tout mettre à jour
  /update --check               Voir les mises à jour
  /update --component hooks     Hooks seulement

Source: kodflow/devcontainer-template (main)
═══════════════════════════════════════════════
```

---

## Overview

Mise à jour de l'environnement DevContainer avec patterns **RLM** :

- **Peek** - Vérifier connectivité et versions
- **Decompose** - Identifier les composants à mettre à jour
- **Parallelize** - Analyser les 5 composants simultanément
- **Synthesize** - Appliquer les mises à jour et rapport consolidé

---

## Configuration

```yaml
REPO: "kodflow/devcontainer-template"
BRANCH: "main"
BASE_URL: "https://raw.githubusercontent.com/${REPO}/${BRANCH}"
```

---

## Phase 1 : Peek (RLM Pattern)

**Vérifications AVANT toute mise à jour :**

```yaml
peek_workflow:
  1_connectivity:
    action: "Vérifier la connectivité GitHub"
    tools: [WebFetch, Bash(curl)]
    check: "API GitHub accessible"

  2_version_check:
    action: "Récupérer le dernier commit du template"
    tools: [WebFetch]
    url: "https://api.github.com/repos/kodflow/devcontainer-template/commits/main"

  3_local_version:
    action: "Lire la version locale"
    tools: [Read]
    file: ".devcontainer/.template-version"
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════
  /update - Peek Analysis
═══════════════════════════════════════════════

  Connectivity: ✓ GitHub API accessible
  Local version  : abc1234 (2024-01-15)
  Remote version : def5678 (2024-01-20)

  Status: UPDATE AVAILABLE

═══════════════════════════════════════════════
```

---

## Phase 2 : Decompose (RLM Pattern)

**Identifier les composants à analyser :**

```yaml
decompose_workflow:
  components:
    features:
      path: ".devcontainer/features/languages/"
      files: ["*/RULES.md"]
      description: "Language features et conventions"

    hooks:
      path: ".devcontainer/images/.claude/scripts/"
      files: ["*.sh"]
      description: "Scripts Claude (format, lint, security)"

    commands:
      path: ".devcontainer/images/.claude/commands/"
      files: ["*.md"]
      description: "Commandes slash (/git, /search)"

    p10k:
      path: ".devcontainer/images/.p10k.zsh"
      files: [".p10k.zsh"]
      description: "Configuration Powerlevel10k"

    settings:
      path: ".devcontainer/images/.claude/settings.json"
      files: ["settings.json"]
      description: "Configuration Claude"

  output: "5 composants à analyser"
```

---

## Phase 3 : Parallelize (RLM Pattern)

**Lancer 5 Task agents en PARALLÈLE pour analyser chaque composant :**

```yaml
parallel_analysis:
  mode: "PARALLEL (single message, 5 Task calls)"

  agents:
    - task: "features-analyzer"
      type: "Explore"
      model: "haiku"
      prompt: |
        Compare features/languages/ local vs remote
        For each language: check RULES.md differences
        Return: {language, status, changes[]}

    - task: "hooks-analyzer"
      type: "Explore"
      model: "haiku"
      prompt: |
        Compare .claude/scripts/ local vs remote
        For each script: check content differences
        Return: {script, status, changes[]}

    - task: "commands-analyzer"
      type: "Explore"
      model: "haiku"
      prompt: |
        Compare .claude/commands/ local vs remote
        For each command: check content differences
        Return: {command, status, changes[]}

    - task: "p10k-analyzer"
      type: "Explore"
      model: "haiku"
      prompt: |
        Compare .p10k.zsh local vs remote
        Return: {status, changes[]}

    - task: "settings-analyzer"
      type: "Explore"
      model: "haiku"
      prompt: |
        Compare settings.json local vs remote
        Return: {status, changes[]}
```

**IMPORTANT** : Lancer les 5 agents dans UN SEUL message.

**Output Phase 3 :**

```
═══════════════════════════════════════════════
  Component Analysis (Parallel)
═══════════════════════════════════════════════

  features:
    + languages/zig/           (new)
    ~ languages/go/RULES.md    (modified)

  hooks:
    ~ format.sh                (modified)
    ~ lint.sh                  (modified)

  commands:
    ~ git.md                   (modified)

  p10k:
    (no changes)

  settings:
    ~ settings.json            (modified)

═══════════════════════════════════════════════
```

---

## Phase 4 : Synthesize (RLM Pattern)

### 4.1 : Appliquer les mises à jour

**Pour chaque composant avec changements :**

#### Features

```bash
BASE="https://raw.githubusercontent.com/kodflow/devcontainer-template/main"
for lang in go nodejs python rust java ruby php elixir dart-flutter scala carbon cpp; do
    curl -sL "$BASE/.devcontainer/features/languages/$lang/RULES.md" \
         -o ".devcontainer/features/languages/$lang/RULES.md" 2>/dev/null
done
```

#### Hooks (scripts)

```bash
for script in format imports lint security test commit-validate bash-validate pre-validate post-edit; do
    curl -sL "$BASE/.devcontainer/images/.claude/scripts/$script.sh" \
         -o ".devcontainer/images/.claude/scripts/$script.sh" 2>/dev/null
    chmod +x ".devcontainer/images/.claude/scripts/$script.sh"
done
```

#### Commands

```bash
for cmd in git search update; do
    curl -sL "$BASE/.devcontainer/images/.claude/commands/$cmd.md" \
         -o ".devcontainer/images/.claude/commands/$cmd.md" 2>/dev/null
done
```

#### p10k

```bash
curl -sL "$BASE/.devcontainer/images/.p10k.zsh" \
     -o ".devcontainer/images/.p10k.zsh" 2>/dev/null
```

#### Settings

```bash
curl -sL "$BASE/.devcontainer/images/.claude/settings.json" \
     -o ".devcontainer/images/.claude/settings.json" 2>/dev/null
```

### 4.2 : Validation finale

```yaml
validation_workflow:
  1_verify_files:
    action: "Vérifier que tous les fichiers sont valides"
    check: "Pas de 404, syntaxe correcte"

  2_run_hooks:
    action: "Exécuter les hooks pour valider"
    tools: [Bash]

  3_update_version:
    action: "Mettre à jour .template-version"
    tools: [Write]
```

```bash
# Enregistrer la version
COMMIT=$(curl -sL "https://api.github.com/repos/kodflow/devcontainer-template/commits/main" | jq -r '.sha[:7]')
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo "{\"commit\": \"$COMMIT\", \"updated\": \"$DATE\"}" > .devcontainer/.template-version
```

### 4.3 : Rapport consolidé

**Output Final :**

```
═══════════════════════════════════════════════
  ✓ DevContainer updated successfully
═══════════════════════════════════════════════

  Template: kodflow/devcontainer-template
  Version : def5678 (2024-01-20)

  Updated components:
    ✓ features    (2 files)
    ✓ hooks       (5 files)
    ✓ commands    (1 file)
    - p10k        (unchanged)
    ✓ settings    (1 file)

  Total: 9 files updated

  Note: Restart terminal to apply p10k changes.

═══════════════════════════════════════════════
```

---

## --check

Mode dry-run : affiche les différences sans appliquer.

```
═══════════════════════════════════════════════
  /update --check
═══════════════════════════════════════════════

  Updates available:

  features (2 changes):
    ~ go/RULES.md      → Go 1.24 (was 1.23)
    + zig/             → New language support

  hooks (1 change):
    ~ lint.sh          → Added ktn-linter support

  commands (0 changes):
    (up to date)

  Run '/update' to apply all changes.
═══════════════════════════════════════════════
```

---

## --component NAME

Met à jour un seul composant.

```
/update --component hooks

═══════════════════════════════════════════════
  /update --component hooks
═══════════════════════════════════════════════

  Updating: hooks only

  ✓ format.sh      updated
  ✓ imports.sh     updated
  ✓ lint.sh        updated
  ✓ security.sh    updated
  ✓ test.sh        updated
  - pre-validate   (unchanged)
  - post-edit      (unchanged)

  Done: 5 files updated

═══════════════════════════════════════════════
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Status | Raison |
|--------|--------|--------|
| Skip Phase 1 (Peek) | ❌ **INTERDIT** | Vérifier versions avant MAJ |
| Mettre à jour depuis source non-officielle | ❌ **INTERDIT** | Sécurité |
| Modifier fichiers hors .devcontainer/ | ❌ **INTERDIT** | Scope limité |
| Écraser fichiers modifiés sans backup | ⚠ WARNING | Afficher diff d'abord |

### Parallélisation légitime

| Élément | Parallèle? | Raison |
|---------|------------|--------|
| Analyse des 5 composants | ✅ Parallèle | Comparaisons indépendantes |
| Application des mises à jour | ❌ Séquentiel | Ordre peut importer |
| Validation finale | ✅ Parallèle | Checks indépendants |

---

## Fichiers concernés

**Mis à jour par /update :**
```
.devcontainer/
├── features/languages/*/RULES.md
├── images/
│   ├── .p10k.zsh
│   └── .claude/
│       ├── commands/*.md
│       ├── scripts/*.sh
│       └── settings.json
└── .template-version
```

**JAMAIS modifiés :**
```
.devcontainer/
├── devcontainer.json      # Config projet
├── docker-compose.yml     # Services locaux
├── Dockerfile             # Customisations
└── hooks/                 # Hooks lifecycle
```
