---
name: init
description: |
  Project initialization check with RLM decomposition.
  Validates environment, dependencies, configuration, and grepai indexing.
  Use when: starting work on a project, verifying setup,
  or troubleshooting environment issues.
allowed-tools:
  - "Bash(git:*)"
  - "Bash(docker:*)"
  - "Bash(terraform:*)"
  - "Bash(kubectl:*)"
  - "Bash(node:*)"
  - "Bash(python:*)"
  - "Bash(go:*)"
  - "Bash(grepai:*)"
  - "Bash(curl:*)"
  - "Bash(pgrep:*)"
  - "Bash(nohup:*)"
  - "Read(**/*)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "Task(*)"
  - "mcp__github__*"
  - "mcp__codacy__*"
---

# /init - Project Initialization (RLM Architecture)

$ARGUMENTS

---

## Overview

Vérification d'initialisation projet avec patterns **RLM** :

- **Peek** - Scan rapide du projet (type, structure)
- **Decompose** - Catégoriser les vérifications (tools, deps, config, env)
- **Parallelize** - Checks simultanés par catégorie
- **Synthesize** - Rapport consolidé avec actions

---

## Usage

```
/init                      # Full initialization check
/init --tools              # Check tools only
/init --deps               # Check dependencies only
/init --env                # Check environment only
/init --fix                # Attempt auto-fix issues
/init --help               # Show help
```

---

## --help

```
═══════════════════════════════════════════════════════════════
  /init - Project Initialization (RLM)
═══════════════════════════════════════════════════════════════

Usage: /init [options]

Options:
  (none)            Full initialization check
  --tools           Check tools only
  --deps            Check dependencies only
  --env             Check environment only
  --fix             Attempt auto-fix issues
  --help            Show this help

RLM Patterns:
  1. Peek       - Detect project type
  2. Decompose  - Categorize checks
  3. Parallelize - Run checks simultaneously
  4. Synthesize - Consolidated report

Exemples:
  /init                       Full check
  /init --tools               Tools versions only
  /init --fix                 Auto-fix issues

═══════════════════════════════════════════════════════════════
```

---

## Phase 1 : Peek (RLM Pattern)

**Scan rapide du projet :**

```yaml
peek_workflow:
  1_structure:
    action: "Scanner la structure du projet"
    tools: [Glob]
    patterns:
      - "package.json"
      - "go.mod"
      - "Cargo.toml"
      - "pyproject.toml"
      - "*.tf"
      - "Dockerfile"
      - "*.yaml"

  2_identify_type:
    action: "Identifier le type de projet"
    mapping:
      - "package.json → Node.js"
      - "go.mod → Go"
      - "Cargo.toml → Rust"
      - "pyproject.toml → Python"
      - "*.tf → Terraform"
      - "Dockerfile → Container"
      - "deployment.yaml → Kubernetes"

  3_detect_requirements:
    action: "Extraire les requirements"
    tools: [Grep]
    patterns:
      - "engines" in package.json
      - "go" version in go.mod
      - "rust-version" in Cargo.toml
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════════════════════
  /init - Peek Analysis
═══════════════════════════════════════════════════════════════

  Project: /workspace

  Detected Types:
    ✓ Node.js (package.json)
    ✓ Terraform (*.tf)
    ✓ Docker (Dockerfile)

  Requirements extracted:
    - Node.js >= 20.x
    - Terraform >= 1.6.x
    - Docker >= 24.x

═══════════════════════════════════════════════════════════════
```

---

## Phase 2 : Decompose (RLM Pattern)

**Catégoriser les vérifications :**

```yaml
decompose_workflow:
  categories:
    tools:
      description: "Vérifier les outils installés et versions"
      checks:
        - git
        - node/npm
        - go
        - terraform
        - docker
        - kubectl
        - grepai

    dependencies:
      description: "Vérifier les dépendances du projet"
      checks:
        - "npm ci / npm install"
        - "go mod download"
        - "terraform init"

    configuration:
      description: "Vérifier les fichiers de configuration"
      checks:
        - ".env exists if .env.example"
        - "Config files valid syntax"
        - "CLAUDE.md present"

    environment:
      description: "Vérifier les variables d'environnement"
      checks:
        - "Required env vars set"
        - "MCP servers configured"
        - "Tokens available"

    semantic_search:
      description: "Initialiser grepai pour recherche sémantique"
      checks:
        - "Ollama sidecar accessible"
        - ".grepai/ config exists"
        - "grepai watch daemon running"
        - "Index status (files indexed)"
```

---

## Phase 3 : Parallelize (RLM Pattern)

**Lancer les checks en PARALLÈLE via Task agents :**

```yaml
parallel_checks:
  mode: "PARALLEL (single message, 5 Task calls)"

  agents:
    - task: "tools-checker"
      type: "Explore"
      prompt: |
        Check installed tools:
        - git --version
        - node --version
        - go version
        - terraform version
        - docker version
        - grepai version
        Return: {tool, required, installed, status}

    - task: "deps-checker"
      type: "Explore"
      prompt: |
        Check dependencies:
        - npm ci (if package.json)
        - go mod download (if go.mod)
        - terraform init (if *.tf)
        Return: {manager, status, issues}

    - task: "config-checker"
      type: "Explore"
      prompt: |
        Check configuration:
        - .env exists if .env.example
        - CLAUDE.md present
        - Config files valid
        Return: {file, status, issue}

    - task: "env-checker"
      type: "Explore"
      prompt: |
        Check environment:
        - Required env vars
        - MCP tokens (GITHUB_TOKEN, CODACY_TOKEN)
        Return: {variable, status, source}

    - task: "grepai-checker"
      type: "Explore"
      prompt: |
        Initialize and check grepai semantic search:
        1. Check Host Ollama (GPU-accelerated): curl -sf http://host.docker.internal:11434/api/tags
        2. Check .grepai/config.yaml exists
        3. Verify endpoint in config is host.docker.internal:11434
        4. Check daemon: pgrep -f "grepai watch"
        5. If not running: nohup grepai watch >/tmp/grepai.log 2>&1 &
        6. Check index: mcp__grepai__grepai_index_status
        Return: {ollama_host, gpu_accelerated, config, daemon, index_files, status}

        If Ollama unavailable, provide HOST setup instructions:
        - macOS: brew install ollama && ollama serve && ollama pull qwen3-embedding:0.6b
        - Linux: curl -fsSL https://ollama.ai/install.sh | sh
        Note: Ollama runs on HOST for GPU acceleration (Metal/CUDA)
```

**IMPORTANT** : Lancer les 5 agents dans UN SEUL message.

---

## Phase 4 : Synthesize (RLM Pattern)

**Consolider les résultats :**

```yaml
synthesize_workflow:
  1_collect:
    action: "Rassembler les résultats des 4 agents"

  2_categorize:
    action: "Classer par sévérité"
    levels:
      - CRITICAL: "Bloquant, impossible de travailler"
      - WARNING: "Problème potentiel"
      - INFO: "Suggestion d'amélioration"
      - PASS: "OK"

  3_generate_report:
    action: "Générer rapport structuré"

  4_suggest_fixes:
    action: "Proposer des actions correctives"
```

**Output Final :**

```
═══════════════════════════════════════════════════════════════
  /init - Project Initialization Report
═══════════════════════════════════════════════════════════════

  Project: example-app
  Types  : Node.js, Terraform, Docker

## Tools Status

| Tool | Required | Installed | Status |
|------|----------|-----------|--------|
| git | 2.40+ | 2.42.0 | ✓ PASS |
| node | 20.x | 20.10.0 | ✓ PASS |
| terraform | 1.6+ | 1.7.0 | ✓ PASS |
| docker | 24+ | 24.0.7 | ✓ PASS |

## Dependencies

| Manager | Status | Issues |
|---------|--------|--------|
| npm | ✓ PASS | 0 vulnerabilities |
| terraform | ✓ PASS | Initialized |

## Configuration

| File | Status | Issue |
|------|--------|-------|
| .env | ⚠ MISSING | Copy from .env.example |
| CLAUDE.md | ✓ PASS | - |
| .gitignore | ✓ PASS | - |

## Environment

| Variable | Status | Source |
|----------|--------|--------|
| GITHUB_TOKEN | ✓ SET | mcp.json |
| CODACY_TOKEN | ⚠ MISSING | Required |
| DATABASE_URL | ⚠ MISSING | .env |

## Semantic Search (grepai)

| Component | Status | Details |
|-----------|--------|---------|
| Ollama | ✓ READY | host.docker.internal:11434 (GPU) |
| Config | ✓ EXISTS | .grepai/config.yaml |
| Daemon | ✓ RUNNING | grepai watch (PID 1234) |
| Index | ✓ INDEXED | 296 files, 1.2MB |

## Recommended Actions

1. `cp .env.example .env` - Create env file
2. Set `CODACY_TOKEN` in environment
3. Set `DATABASE_URL` in .env

## Quick Start

```bash
cp .env.example .env
# Edit .env with your values
npm install
npm run dev
```

## Search Usage

```yaml
# MANDATORY: Use grepai MCP for ALL code searches
semantic_search: mcp__grepai__grepai_search(query="...")
call_analysis: mcp__grepai__grepai_trace_callers(symbol="...")
fallback_only: Grep tool (only if grepai fails)
```

═══════════════════════════════════════════════════════════════
```

---

## --fix Mode

**Auto-fix avec parallélisation :**

```yaml
fix_workflow:
  parallel_fixes:
    - action: "cp .env.example .env"
      condition: ".env missing && .env.example exists"

    - action: "npm audit fix"
      condition: "npm vulnerabilities > 0"

    - action: "terraform init -upgrade"
      condition: "terraform not initialized"

  mode: "PARALLEL where independent"
```

---

## Detection Patterns

```yaml
project_types:
  nodejs:
    files: ["package.json"]
    tools: ["node", "npm"]
    deps: "npm ci"

  go:
    files: ["go.mod"]
    tools: ["go"]
    deps: "go mod download"

  python:
    files: ["pyproject.toml", "requirements.txt"]
    tools: ["python", "pip"]
    deps: "pip install -r requirements.txt"

  terraform:
    files: ["*.tf"]
    tools: ["terraform", "tflint"]
    deps: "terraform init"

  kubernetes:
    files: ["**/deployment.yaml", "helm/"]
    tools: ["kubectl", "helm"]

  docker:
    files: ["Dockerfile", "docker-compose.yml"]
    tools: ["docker"]
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Status |
|--------|--------|
| Skip Phase 1 (Peek) | ❌ **INTERDIT** |
| Checks séquentiels | ❌ **INTERDIT** |
| Ignorer CRITICAL issues | ❌ **INTERDIT** |
| Auto-fix sans --fix flag | ⚠ WARNING |
