---
name: update
description: |
  DevContainer Environment Update from official template.
  Updates hooks, commands, agents, and settings from kodflow/devcontainer-template.
  Use when: syncing local devcontainer with latest template improvements.
allowed-tools:
  - "Bash(curl:*)"
  - "Bash(git:*)"
  - "Bash(jq:*)"
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

**Approche API-FIRST** : Utilise l'API GitHub pour découvrir dynamiquement
les fichiers existants au lieu de listes hardcodées.

**Composants mis à jour :**

- **Hooks** - Scripts Claude (format, lint, security, etc.)
- **Commands** - Commandes slash (/git, /search, etc.)
- **Agents** - Définitions d'agents (specialists, executors)
- **Lifecycle** - Hooks de cycle de vie (postStart)
- **Config** - p10k, settings.json
- **Compose** - docker-compose.yml (update devcontainer, preserve custom)
- **Grepai** - Configuration grepai optimisée

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

| Composant | Chemin | Description |
|-----------|--------|-------------|
| `hooks` | `.devcontainer/images/.claude/scripts/` | Scripts Claude |
| `commands` | `.devcontainer/images/.claude/commands/` | Commandes slash |
| `agents` | `.devcontainer/images/.claude/agents/` | Définitions d'agents |
| `lifecycle` | `.devcontainer/hooks/lifecycle/` | Hooks de cycle de vie |
| `p10k` | `.devcontainer/images/.p10k.zsh` | Config Powerlevel10k |
| `settings` | `.../images/.claude/settings.json` | Config Claude |
| `compose` | `.devcontainer/docker-compose.yml` | Update devcontainer service |
| `grepai` | `.devcontainer/images/grepai.config.yaml` | Config grepai |

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
  hooks       Scripts Claude (format, lint...)
  commands    Commandes slash (/git, /search)
  agents      Agent definitions (specialists)
  lifecycle   Lifecycle hooks (postStart)
  p10k        Powerlevel10k config
  settings    Claude settings.json
  compose     docker-compose.yml (devcontainer service)
  grepai      grepai config (provider, model)

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
- **Discover** - Découvrir dynamiquement les fichiers via API GitHub
- **Validate** - Valider chaque téléchargement (pas de 404)
- **Synthesize** - Appliquer les mises à jour et rapport consolidé

---

## Configuration

```yaml
REPO: "kodflow/devcontainer-template"
BRANCH: "main"
BASE_URL: "https://raw.githubusercontent.com/${REPO}/${BRANCH}"
API_URL: "https://api.github.com/repos/${REPO}/contents"
```

---

## Phase 1 : Peek (Version Check)

```yaml
peek_workflow:
  1_connectivity:
    action: "Vérifier la connectivité GitHub"
    tool: WebFetch
    url: "https://api.github.com/repos/kodflow/devcontainer-template/commits/main"

  2_local_version:
    action: "Lire la version locale"
    tool: Read
    file: ".devcontainer/.template-version"
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════
  /update - Peek Analysis
═══════════════════════════════════════════════

  Connectivity   : ✓ GitHub API accessible
  Local version  : abc1234 (2024-01-15)
  Remote version : def5678 (2024-01-20)

  Status: UPDATE AVAILABLE

═══════════════════════════════════════════════
```

---

## Phase 2 : Discover (API-FIRST - Dynamic Discovery)

**RÈGLE CRITIQUE : Toujours utiliser l'API GitHub pour découvrir les fichiers.**

Ne JAMAIS utiliser de listes hardcodées. Les fichiers peuvent être ajoutés,
renommés ou supprimés dans le template source.

```yaml
discover_workflow:
  strategy: "API-FIRST"

  components:
    hooks:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/scripts"
      filter: "*.sh"
      local_path: ".devcontainer/images/.claude/scripts/"

    commands:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/commands"
      filter: "*.md"
      local_path: ".devcontainer/images/.claude/commands/"

    agents:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/agents"
      filter: "*.md"
      local_path: ".devcontainer/images/.claude/agents/"

    lifecycle:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/hooks/lifecycle"
      filter: "*.sh"
      local_path: ".devcontainer/hooks/lifecycle/"

    p10k:
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/images/.p10k.zsh"
      local_path: ".devcontainer/images/.p10k.zsh"

    settings:
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/images/.claude/settings.json"
      local_path: ".devcontainer/images/.claude/settings.json"

    compose:
      strategy: "REPLACE from template, PRESERVE custom services"
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/docker-compose.yml"
      local_path: ".devcontainer/docker-compose.yml"
      note: |
        - Si fichier absent → télécharger complet
        - Si fichier existe:
          1. Extraire services custom (pas devcontainer)
          2. Remplacer entièrement par le template (préserve ordre/commentaires)
          3. Merger les services custom extraits
        - Ordre: devcontainer → custom
        - Backup créé avant modification, restauré si échec
        - Utilise mikefarah/yq (Go version) pour le merge
        - Note: Ollama runs on HOST (installed via initialize.sh)

    grepai:
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/images/grepai.config.yaml"
      local_path: ".devcontainer/images/grepai.config.yaml"
      note: "Optimized config with qwen3-embedding model"
```

**Implémentation Discover :**

```bash
# Fonction pour lister les fichiers d'un répertoire via API GitHub
list_remote_files() {
    local api_url="$1"
    local filter="$2"

    curl -sL "$api_url" | jq -r '.[].name' | grep -E "$filter" || true
}

# Exemple: découvrir les scripts
SCRIPTS=$(list_remote_files \
    "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/scripts" \
    '\.sh$')

# Exemple: découvrir les commandes
COMMANDS=$(list_remote_files \
    "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/commands" \
    '\.md$')

# Exemple: découvrir les agents
AGENTS=$(list_remote_files \
    "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/agents" \
    '\.md$')
```

---

## Phase 3 : Validate (Download with Verification)

**RÈGLE CRITIQUE : Toujours valider les téléchargements avant écriture.**

Ne JAMAIS écrire un fichier sans vérifier que le téléchargement a réussi.
Détecter les erreurs 404 et autres échecs.

```yaml
validate_workflow:
  rule: "NEVER write files without validation"

  checks:
    - "HTTP status 200 (not 404)"
    - "Content is not empty"
    - "Content is not HTML error page"
    - "Content starts with expected pattern"

  on_failure:
    - "Log error"
    - "Skip file"
    - "Continue with next file"
```

**Implémentation Validate :**

```bash
# Fonction de téléchargement sécurisé avec validation
safe_download() {
    local url="$1"
    local output="$2"
    local temp_file=$(mktemp)

    # Télécharger avec code HTTP
    http_code=$(curl -sL -w "%{http_code}" -o "$temp_file" "$url")

    # Valider le téléchargement
    if [ "$http_code" != "200" ]; then
        echo "✗ $output (HTTP $http_code)"
        rm -f "$temp_file"
        return 1
    fi

    # Vérifier que ce n'est pas une page 404 déguisée
    if head -1 "$temp_file" | grep -qE "^404|^<!DOCTYPE|^<html"; then
        echo "✗ $output (invalid content)"
        rm -f "$temp_file"
        return 1
    fi

    # Vérifier que le fichier n'est pas vide
    if [ ! -s "$temp_file" ]; then
        echo "✗ $output (empty)"
        rm -f "$temp_file"
        return 1
    fi

    # Tout est OK, déplacer le fichier
    mv "$temp_file" "$output"
    echo "✓ $output"
    return 0
}
```

---

## Phase 4 : Synthesize (Apply Updates)

### 4.1 : Télécharger les composants

**IMPORTANT** : Utiliser `safe_download` pour chaque fichier.

#### Hooks (scripts)

```bash
BASE="https://raw.githubusercontent.com/kodflow/devcontainer-template/main"
API="https://api.github.com/repos/kodflow/devcontainer-template/contents"

# Découvrir les scripts via API
SCRIPTS=$(curl -sL "$API/.devcontainer/images/.claude/scripts" | jq -r '.[].name' | grep '\.sh$')

# Télécharger chaque script avec validation
for script in $SCRIPTS; do
    safe_download \
        "$BASE/.devcontainer/images/.claude/scripts/$script" \
        ".devcontainer/images/.claude/scripts/$script" \
    && chmod +x ".devcontainer/images/.claude/scripts/$script"
done
```

#### Commands

```bash
# Découvrir les commandes via API
COMMANDS=$(curl -sL "$API/.devcontainer/images/.claude/commands" | jq -r '.[].name' | grep '\.md$')

# Télécharger chaque commande avec validation
for cmd in $COMMANDS; do
    safe_download \
        "$BASE/.devcontainer/images/.claude/commands/$cmd" \
        ".devcontainer/images/.claude/commands/$cmd"
done
```

#### Agents

```bash
# Découvrir les agents via API
mkdir -p ".devcontainer/images/.claude/agents"
AGENTS=$(curl -sL "$API/.devcontainer/images/.claude/agents" | jq -r '.[].name' | grep '\.md$')

# Télécharger chaque agent avec validation
for agent in $AGENTS; do
    safe_download \
        "$BASE/.devcontainer/images/.claude/agents/$agent" \
        ".devcontainer/images/.claude/agents/$agent"
done
```

#### Lifecycle Hooks

```bash
# Découvrir les lifecycle hooks via API
mkdir -p ".devcontainer/hooks/lifecycle"
LIFECYCLE=$(curl -sL "$API/.devcontainer/hooks/lifecycle" | jq -r '.[].name' | grep '\.sh$')

# Télécharger chaque hook avec validation
for hook in $LIFECYCLE; do
    safe_download \
        "$BASE/.devcontainer/hooks/lifecycle/$hook" \
        ".devcontainer/hooks/lifecycle/$hook" \
    && chmod +x ".devcontainer/hooks/lifecycle/$hook"
done
```

#### Config Files (p10k, settings, compose, grepai)

```bash
# p10k
safe_download \
    "$BASE/.devcontainer/images/.p10k.zsh" \
    ".devcontainer/images/.p10k.zsh"

# settings.json
safe_download \
    "$BASE/.devcontainer/images/.claude/settings.json" \
    ".devcontainer/images/.claude/settings.json"

# docker-compose.yml (update devcontainer service, PRESERVE custom services)
# Note: Uses mikefarah/yq (Go version) - simpler syntax with -i for in-place
# Strategy: Start fresh from template, merge back custom services
# Note: Ollama runs on HOST (installed via initialize.sh), not in container
update_compose_services() {
    local compose_file=".devcontainer/docker-compose.yml"
    local temp_template=$(mktemp --suffix=.yaml)
    local temp_custom=$(mktemp --suffix=.yaml)
    local backup_file="${compose_file}.backup"

    # Download template
    if ! curl -sL -o "$temp_template" "$BASE/.devcontainer/docker-compose.yml"; then
        echo "  ✗ Failed to download template"
        rm -f "$temp_template"
        return 1
    fi

    # Validate downloaded template is not empty and contains expected content
    if [ ! -s "$temp_template" ] || ! grep -q "^services:" "$temp_template"; then
        echo "  ✗ Downloaded template is empty or invalid (check network/rate limit)"
        rm -f "$temp_template"
        return 1
    fi

    # Backup original
    cp "$compose_file" "$backup_file"

    # Extract custom services (anything that's NOT devcontainer)
    yq '.services | to_entries | map(select(.key != "devcontainer")) | from_entries' "$compose_file" > "$temp_custom"

    # Start fresh from template (devcontainer service)
    cp "$temp_template" "$compose_file"

    # Merge back custom services if any exist
    if [ -s "$temp_custom" ] && [ "$(yq '. | length' "$temp_custom")" != "0" ]; then
        yq -i ".services *= load(\"$temp_custom\")" "$compose_file"
        echo "    - Preserved custom services"
    fi

    # Cleanup temp files
    rm -f "$temp_template" "$temp_custom"

    # Verify file is not empty and contains required structure
    if [ ! -s "$compose_file" ] || ! grep -q "^services:" "$compose_file"; then
        # File is empty or missing services - restore backup
        mv "$backup_file" "$compose_file"
        echo "  ✗ Result file is empty or invalid, restored backup"
        return 1
    fi

    # Verify YAML is valid and has expected content
    if yq '.services.devcontainer' "$compose_file" > /dev/null 2>&1; then
        rm -f "$backup_file"
        echo "  ✓ docker-compose.yml updated"
        echo "    - REPLACED: devcontainer from template"
        echo "    - PRESERVED: custom services (if any)"
        return 0
    else
        # Restore backup on failure
        mv "$backup_file" "$compose_file"
        echo "  ✗ YAML validation failed (missing devcontainer service), restored backup"
        return 1
    fi
}

if [ ! -f ".devcontainer/docker-compose.yml" ]; then
    # No file exists - download full template
    safe_download \
        "$BASE/.devcontainer/docker-compose.yml" \
        ".devcontainer/docker-compose.yml"
    echo "  ✓ docker-compose.yml created from template"
else
    # File exists - update devcontainer service
    echo "  Updating devcontainer service..."
    update_compose_services
fi

# grepai config (optimized with qwen3-embedding)
safe_download \
    "$BASE/.devcontainer/images/grepai.config.yaml" \
    ".devcontainer/images/grepai.config.yaml"
```

### 4.2 : Cleanup deprecated files

```bash
# Remove deprecated configuration files
[ -f ".coderabbit.yaml" ] && rm -f ".coderabbit.yaml" \
    && echo "Removed deprecated .coderabbit.yaml"
```

### 4.3 : Update version file

```bash
COMMIT=$(curl -sL "https://api.github.com/repos/kodflow/devcontainer-template/commits/main" | jq -r '.sha[:7]')
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo "{\"commit\": \"$COMMIT\", \"updated\": \"$DATE\"}" > .devcontainer/.template-version
```

### 4.4 : Rapport consolidé

**Output Final :**

```
═══════════════════════════════════════════════
  ✓ DevContainer updated successfully
═══════════════════════════════════════════════

  Template: kodflow/devcontainer-template
  Version : def5678 (2024-01-20)

  Updated components:
    ✓ hooks       (10 scripts)
    ✓ commands    (10 commands)
    ✓ agents      (35 agents)
    ✓ lifecycle   (6 hooks)
    ✓ p10k        (1 file)
    ✓ settings    (1 file)
    ✓ compose     (devcontainer service updated)
    ✓ grepai      (1 file - qwen3-embedding config)
    ✓ user-hooks  (synchronized with template)
    ✓ validation  (all scripts exist)

  Grepai config:
    provider: ollama
    model: qwen3-embedding:0.6b
    endpoint: host.docker.internal:11434 (GPU-accelerated)

  Cleanup:
    ✓ .coderabbit.yaml removed (if existed)

  Note: Restart terminal to apply p10k changes.

═══════════════════════════════════════════════
```

---

## Phase 5 : Hook Synchronization

**But :** Synchroniser les hooks de `~/.claude/settings.json` avec le template.

**Problème résolu :** Les utilisateurs avec un ancien `settings.json` peuvent avoir
des références à des scripts obsolètes (bash-validate.sh, phase-validate.sh, etc.)
car `postStart.sh` ne copie `settings.json` que s'il n'existe pas.

```yaml
hook_sync_workflow:
  1_backup:
    action: "Backup user settings.json"
    command: "cp ~/.claude/settings.json ~/.claude/settings.json.backup"

  2_merge_hooks:
    action: "Remplacer la section hooks avec le template"
    strategy: "REPLACE (pas merge) - le template est la source de vérité"
    tool: jq
    preserves:
      - permissions
      - model
      - env
      - statusLine
      - disabledMcpjsonServers

  3_restore_on_failure:
    action: "Restaurer le backup si le merge échoue"
```

**Implémentation :**

```bash
sync_user_hooks() {
    local user_settings="$HOME/.claude/settings.json"
    local template_settings=".devcontainer/images/.claude/settings.json"

    if [ ! -f "$user_settings" ]; then
        echo "  ⚠ No user settings.json, skipping hook sync"
        return 0
    fi

    if [ ! -f "$template_settings" ]; then
        echo "  ✗ Template settings.json not found"
        return 1
    fi

    echo "  Synchronizing user hooks with template..."

    # Backup
    cp "$user_settings" "${user_settings}.backup"

    # Replace hooks section only (preserve all other settings)
    if jq --slurpfile tpl "$template_settings" '.hooks = $tpl[0].hooks' \
       "$user_settings" > "${user_settings}.tmp"; then

        # Validate JSON
        if jq empty "${user_settings}.tmp" 2>/dev/null; then
            mv "${user_settings}.tmp" "$user_settings"
            rm -f "${user_settings}.backup"
            echo "  ✓ User hooks synchronized with template"
            return 0
        else
            mv "${user_settings}.backup" "$user_settings"
            rm -f "${user_settings}.tmp"
            echo "  ✗ Hook merge produced invalid JSON, restored backup"
            return 1
        fi
    else
        mv "${user_settings}.backup" "$user_settings"
        echo "  ✗ Hook merge failed, restored backup"
        return 1
    fi
}
```

---

## Phase 6 : Script Validation

**But :** Valider que tous les scripts référencés dans les hooks existent.

```yaml
validate_workflow:
  1_extract:
    action: "Extraire tous les chemins de scripts depuis les hooks"
    tool: jq
    pattern: ".hooks | .. | .command? // empty"

  2_verify:
    action: "Vérifier que chaque script existe"
    for_each: script_path
    check: "[ -f $script_path ]"

  3_report:
    on_missing: "Lister les scripts manquants avec suggestion de fix"
    on_success: "Tous les scripts validés"
```

**Implémentation :**

```bash
validate_hook_scripts() {
    local settings_file="$HOME/.claude/settings.json"
    local scripts_dir="$HOME/.claude/scripts"
    local missing_count=0

    if [ ! -f "$settings_file" ]; then
        echo "  ⚠ No settings.json to validate"
        return 0
    fi

    # Extract all script paths from hooks
    local scripts
    scripts=$(jq -r '.hooks | .. | .command? // empty' "$settings_file" 2>/dev/null \
        | grep -oE '/home/vscode/.claude/scripts/[^ "]+' \
        | sed 's/ .*//' \
        | sort -u)

    if [ -z "$scripts" ]; then
        echo "  ⚠ No hook scripts found in settings.json"
        return 0
    fi

    echo "  Validating hook scripts..."

    for script in $scripts; do
        local script_name=$(basename "$script")

        if [ -f "$script" ]; then
            echo "    ✓ $script_name"
        else
            echo "    ✗ $script_name (MISSING)"
            missing_count=$((missing_count + 1))
        fi
    done

    if [ $missing_count -gt 0 ]; then
        echo ""
        echo "  ⚠ $missing_count missing script(s) detected!"
        echo "  → Run: /update --component hooks"
        return 1
    fi

    echo "  ✓ All hook scripts validated"
    return 0
}
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Status | Raison |
|--------|--------|--------|
| Utiliser listes hardcodées | ❌ **INTERDIT** | API-FIRST obligatoire |
| Écrire sans validation | ❌ **INTERDIT** | Risque de corruption |
| Skip vérification HTTP | ❌ **INTERDIT** | Fichiers 404 possibles |
| Source non-officielle | ❌ **INTERDIT** | Sécurité |
| Hook sync sans backup | ❌ **INTERDIT** | Toujours backup first |
| Supprimer user settings | ❌ **INTERDIT** | Seulement merge hooks |
| Skip script validation | ❌ **INTERDIT** | Détection erreurs obligatoire |

---

## Fichiers concernés

**Mis à jour par /update :**
```
.devcontainer/
├── docker-compose.yml        # Update devcontainer service
├── hooks/lifecycle/*.sh
├── images/
│   ├── .p10k.zsh
│   ├── grepai.config.yaml    # Config grepai (provider, model)
│   └── .claude/
│       ├── agents/*.md
│       ├── commands/*.md
│       ├── scripts/*.sh
│       └── settings.json
└── .template-version
```

**Dans l'image Docker (restauré au démarrage) :**
```
/etc/grepai/config.yaml            # GrepAI config template
/etc/mcp/mcp.json.tpl              # MCP template
/etc/claude-defaults/*             # Claude defaults
```

**JAMAIS modifiés :**
```
.devcontainer/
├── devcontainer.json      # Config projet (customisations)
└── Dockerfile             # Customisations image
```

---

## Script complet (référence)

```bash
#!/bin/bash
# /update implementation - API-FIRST with validation

set -euo pipefail
set +H  # Disable bash history expansion (! in YAML causes errors)

BASE="https://raw.githubusercontent.com/kodflow/devcontainer-template/main"
API="https://api.github.com/repos/kodflow/devcontainer-template/contents"

# Safe download function
safe_download() {
    local url="$1"
    local output="$2"
    local temp_file=$(mktemp)

    http_code=$(curl -sL -w "%{http_code}" -o "$temp_file" "$url")

    if [ "$http_code" != "200" ]; then
        echo "✗ $(basename "$output") (HTTP $http_code)"
        rm -f "$temp_file"
        return 1
    fi

    if head -1 "$temp_file" | grep -qE "^404|^<!DOCTYPE|^<html"; then
        echo "✗ $(basename "$output") (invalid content)"
        rm -f "$temp_file"
        return 1
    fi

    if [ ! -s "$temp_file" ]; then
        echo "✗ $(basename "$output") (empty)"
        rm -f "$temp_file"
        return 1
    fi

    mkdir -p "$(dirname "$output")"
    mv "$temp_file" "$output"
    echo "✓ $(basename "$output")"
    return 0
}

# Hook synchronization function (Phase 5)
sync_user_hooks() {
    local user_settings="$HOME/.claude/settings.json"
    local template_settings=".devcontainer/images/.claude/settings.json"

    if [ ! -f "$user_settings" ]; then
        echo "  ⚠ No user settings.json, skipping hook sync"
        return 0
    fi

    if [ ! -f "$template_settings" ]; then
        echo "  ✗ Template settings.json not found"
        return 1
    fi

    echo "  Synchronizing user hooks with template..."
    cp "$user_settings" "${user_settings}.backup"

    if jq --slurpfile tpl "$template_settings" '.hooks = $tpl[0].hooks' \
       "$user_settings" > "${user_settings}.tmp"; then
        if jq empty "${user_settings}.tmp" 2>/dev/null; then
            mv "${user_settings}.tmp" "$user_settings"
            rm -f "${user_settings}.backup"
            echo "  ✓ User hooks synchronized"
            return 0
        else
            mv "${user_settings}.backup" "$user_settings"
            rm -f "${user_settings}.tmp"
            echo "  ✗ Invalid JSON, restored backup"
            return 1
        fi
    else
        mv "${user_settings}.backup" "$user_settings"
        echo "  ✗ Hook merge failed, restored backup"
        return 1
    fi
}

# Script validation function (Phase 6)
validate_hook_scripts() {
    local settings_file="$HOME/.claude/settings.json"
    local missing_count=0

    if [ ! -f "$settings_file" ]; then
        echo "  ⚠ No settings.json to validate"
        return 0
    fi

    local scripts
    scripts=$(jq -r '.hooks | .. | .command? // empty' "$settings_file" 2>/dev/null \
        | grep -oE '/home/vscode/.claude/scripts/[^ "]+' \
        | sed 's/ .*//' | sort -u)

    if [ -z "$scripts" ]; then
        echo "  ⚠ No hook scripts found"
        return 0
    fi

    echo "  Validating hook scripts..."
    for script in $scripts; do
        local script_name=$(basename "$script")
        if [ -f "$script" ]; then
            echo "    ✓ $script_name"
        else
            echo "    ✗ $script_name (MISSING)"
            missing_count=$((missing_count + 1))
        fi
    done

    if [ $missing_count -gt 0 ]; then
        echo "  ⚠ $missing_count missing script(s)!"
        echo "  → Run: /update --component hooks"
        return 1
    fi

    echo "  ✓ All scripts validated"
    return 0
}

echo "═══════════════════════════════════════════════"
echo "  /update - DevContainer Environment Update"
echo "═══════════════════════════════════════════════"
echo ""

# Hooks
echo "Updating hooks..."
SCRIPTS=$(curl -sL "$API/.devcontainer/images/.claude/scripts" | jq -r '.[].name' | grep '\.sh$' || true)
for script in $SCRIPTS; do
    safe_download "$BASE/.devcontainer/images/.claude/scripts/$script" \
                  ".devcontainer/images/.claude/scripts/$script" \
    && chmod +x ".devcontainer/images/.claude/scripts/$script"
done

# Commands
echo ""
echo "Updating commands..."
COMMANDS=$(curl -sL "$API/.devcontainer/images/.claude/commands" | jq -r '.[].name' | grep '\.md$' || true)
for cmd in $COMMANDS; do
    safe_download "$BASE/.devcontainer/images/.claude/commands/$cmd" \
                  ".devcontainer/images/.claude/commands/$cmd"
done

# Agents
echo ""
echo "Updating agents..."
mkdir -p ".devcontainer/images/.claude/agents"
AGENTS=$(curl -sL "$API/.devcontainer/images/.claude/agents" | jq -r '.[].name' | grep '\.md$' || true)
for agent in $AGENTS; do
    safe_download "$BASE/.devcontainer/images/.claude/agents/$agent" \
                  ".devcontainer/images/.claude/agents/$agent"
done

# Lifecycle
echo ""
echo "Updating lifecycle hooks..."
mkdir -p ".devcontainer/hooks/lifecycle"
LIFECYCLE=$(curl -sL "$API/.devcontainer/hooks/lifecycle" | jq -r '.[].name' | grep '\.sh$' || true)
for hook in $LIFECYCLE; do
    safe_download "$BASE/.devcontainer/hooks/lifecycle/$hook" \
                  ".devcontainer/hooks/lifecycle/$hook" \
    && chmod +x ".devcontainer/hooks/lifecycle/$hook"
done

# Config files
echo ""
echo "Updating config files..."
safe_download "$BASE/.devcontainer/images/.p10k.zsh" ".devcontainer/images/.p10k.zsh"
safe_download "$BASE/.devcontainer/images/.claude/settings.json" ".devcontainer/images/.claude/settings.json"

# Docker compose (update devcontainer service, PRESERVE custom services)
# Note: Uses mikefarah/yq (Go version) - simpler syntax with -i for in-place
# Ollama runs on HOST (installed via initialize.sh), not in container
echo ""
echo "Updating docker-compose.yml..."

update_compose_services() {
    local compose_file=".devcontainer/docker-compose.yml"
    local temp_template=$(mktemp --suffix=.yaml)
    local temp_custom=$(mktemp --suffix=.yaml)
    local backup_file="${compose_file}.backup"

    # Download template
    if ! curl -sL -o "$temp_template" "$BASE/.devcontainer/docker-compose.yml"; then
        echo "  ✗ Failed to download template"
        rm -f "$temp_template"
        return 1
    fi

    # Validate downloaded template is not empty and contains expected content
    if [ ! -s "$temp_template" ] || ! grep -q "^services:" "$temp_template"; then
        echo "  ✗ Downloaded template is empty or invalid (check network/rate limit)"
        rm -f "$temp_template"
        return 1
    fi

    # Backup original
    cp "$compose_file" "$backup_file"

    # Extract custom services (anything that's NOT devcontainer)
    yq '.services | to_entries | map(select(.key != "devcontainer")) | from_entries' "$compose_file" > "$temp_custom"

    # Extract custom volumes (anything that's NOT in template)
    local template_volumes=$(yq '.volumes | keys | .[]' "$temp_template" 2>/dev/null | tr '\n' '|')

    # Start fresh from template (devcontainer service)
    cp "$temp_template" "$compose_file"

    # Merge back custom services if any exist
    if [ -s "$temp_custom" ] && [ "$(yq '. | length' "$temp_custom")" != "0" ]; then
        yq -i ".services *= load(\"$temp_custom\")" "$compose_file"
        echo "    - Preserved custom services"
    fi

    # Cleanup temp files
    rm -f "$temp_template" "$temp_custom"

    # Verify file is not empty and contains required structure
    if [ ! -s "$compose_file" ] || ! grep -q "^services:" "$compose_file"; then
        # File is empty or missing services - restore backup
        mv "$backup_file" "$compose_file"
        echo "  ✗ Result file is empty or invalid, restored backup"
        return 1
    fi

    # Verify YAML is valid and has expected content
    if yq '.services.devcontainer' "$compose_file" > /dev/null 2>&1; then
        rm -f "$backup_file"
        echo "  ✓ devcontainer service updated"
        return 0
    else
        mv "$backup_file" "$compose_file"
        echo "  ✗ YAML validation failed (missing devcontainer service), restored backup"
        return 1
    fi
}

if [ ! -f ".devcontainer/docker-compose.yml" ]; then
    echo "  No docker-compose.yml found, downloading template..."
    safe_download "$BASE/.devcontainer/docker-compose.yml" ".devcontainer/docker-compose.yml"
else
    echo "  Updating devcontainer service..."
    update_compose_services
fi

# Grepai config
echo ""
echo "Updating grepai config..."
safe_download "$BASE/.devcontainer/images/grepai.config.yaml" ".devcontainer/images/grepai.config.yaml"

# Phase 5: Synchronize user hooks
echo ""
echo "Phase 5: Synchronizing user hooks..."
sync_user_hooks

# Phase 6: Validate hook scripts
echo ""
echo "Phase 6: Validating hook scripts..."
validate_hook_scripts

# Version
COMMIT=$(curl -sL "https://api.github.com/repos/kodflow/devcontainer-template/commits/main" | jq -r '.sha[:7]')
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo "{\"commit\": \"$COMMIT\", \"updated\": \"$DATE\"}" > .devcontainer/.template-version

echo ""
echo "═══════════════════════════════════════════════"
echo "  ✓ Update complete - version: $COMMIT"
echo "═══════════════════════════════════════════════"
```
