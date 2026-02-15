---
name: test
description: |
  E2E and frontend testing with Playwright MCP and RLM decomposition.
  Automates browser interactions, visual testing, and debugging.
  Use when: running E2E tests, debugging frontend, generating test code.
allowed-tools:
  - "mcp__playwright__*"
  - "Bash(npm:*)"
  - "Bash(npx:*)"
  - "Read(**/*)"
  - "Write(**/*)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "Task(*)"
---

# /test - E2E & Frontend Testing (RLM Architecture)

$ARGUMENTS

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Use `grepai_trace_callers`/`grepai_trace_callees` for impact analysis.
Fallback to Grep ONLY for exact string matches or regex patterns.

---

## Overview

Tests E2E et debugging frontend avec patterns **RLM** :

- **Peek** - Analyser la page avant interaction
- **Decompose** - Diviser le test en étapes
- **Parallelize** - Assertions et captures simultanées
- **Synthesize** - Rapport de test consolidé

**Capacités Playwright MCP :**

- **Navigation** - Ouvrir URLs, naviguer, screenshots
- **Interaction** - Click, type, select, hover, drag
- **Assertions** - Vérifier texte, éléments, états
- **Tracing** - Enregistrer les sessions pour debug
- **PDF** - Générer des PDFs de pages
- **Codegen** - Générer du code de test

---

## Arguments

| Pattern | Action |
|---------|--------|
| `<url>` | Ouvre l'URL et explore la page |
| `--run` | Exécute les tests Playwright du projet |
| `--debug <url>` | Mode debug interactif |
| `--trace` | Active le tracing pour la session |
| `--screenshot <url>` | Capture d'écran de la page |
| `--pdf <url>` | Génère un PDF de la page |
| `--codegen <url>` | Génère du code de test |
| `--help` | Affiche l'aide |

---

## --help

```
═══════════════════════════════════════════════════════════════
  /test - E2E & Frontend Testing (RLM)
═══════════════════════════════════════════════════════════════

Usage: /test <url|action> [options]

Actions:
  <url>               Ouvre et explore la page
  --run               Exécute les tests du projet
  --debug <url>       Mode debug interactif
  --trace             Active le tracing
  --screenshot <url>  Capture d'écran
  --pdf <url>         Génère un PDF
  --codegen <url>     Génère du code de test

RLM Patterns:
  1. Peek       - Analyser la page (snapshot)
  2. Decompose  - Diviser en étapes de test
  3. Parallelize - Assertions simultanées
  4. Synthesize - Rapport consolidé

MCP Tools:
  browser_navigate    Ouvrir une URL
  browser_click       Cliquer sur un élément
  browser_type        Saisir du texte
  browser_snapshot    Capturer l'état
  browser_expect      Assertions

Exemples:
  /test https://example.com
  /test --screenshot https://myapp.com/login
  /test --run
  /test --codegen https://myapp.com

═══════════════════════════════════════════════════════════════
```

---

## Phase 1.0 : Peek (RLM Pattern)

**Analyser la page AVANT interaction :**

```yaml
peek_workflow:
  1_navigate:
    tool: mcp__playwright__browser_navigate
    params:
      url: "<url>"

  2_snapshot:
    tool: mcp__playwright__browser_snapshot
    output: "Accessibility tree de la page"

  3_analyze:
    action: "Identifier les éléments interactifs"
    extract:
      - forms: "input, select, textarea"
      - buttons: "button, [type=submit]"
      - links: "a[href]"
      - content: "main content areas"
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════════════════════
  /test - Peek Analysis
═══════════════════════════════════════════════════════════════

  URL: https://myapp.com/login

  Page Structure:
    ├─ Header (nav, logo, menu)
    ├─ Main
    │   ├─ Form#login
    │   │   ├─ Input[email]
    │   │   ├─ Input[password]
    │   │   └─ Button[Submit]
    │   └─ Link[Forgot password]
    └─ Footer

  Interactive Elements: 5
  Forms: 1
  Testable: YES

═══════════════════════════════════════════════════════════════
```

---

## Phase 2.0 : Decompose (RLM Pattern)

**Diviser le test en étapes :**

```yaml
decompose_workflow:
  example_login_test:
    steps:
      - step: "Navigate to login"
        action: browser_navigate
        url: "/login"

      - step: "Fill email"
        action: browser_type
        element: "Email input"
        value: "user@test.com"

      - step: "Fill password"
        action: browser_type
        element: "Password input"
        value: "******"

      - step: "Submit form"
        action: browser_click
        element: "Submit button"

      - step: "Verify redirect"
        action: browser_expect
        expectation: "URL contains /dashboard"
```

---

## Phase 3.0 : Parallelize (RLM Pattern)

**Assertions et captures simultanées :**

```yaml
parallel_validation:
  mode: "PARALLEL (single message, multiple MCP calls)"

  actions:
    - task: "Visibility check"
      tool: mcp__playwright__browser_expect
      params:
        expectation: "to_be_visible"
        ref: "<dashboard_ref>"

    - task: "Text check"
      tool: mcp__playwright__browser_expect
      params:
        expectation: "to_have_text"
        ref: "<welcome_ref>"
        expected: "Welcome"

    - task: "Screenshot"
      tool: mcp__playwright__browser_screenshot
      params:
        fullPage: true
```

**IMPORTANT** : Lancer TOUTES les assertions dans UN SEUL message.

---

## Phase 4.0 : Synthesize (RLM Pattern)

**Rapport de test consolidé :**

```yaml
synthesize_workflow:
  1_collect:
    action: "Rassembler tous les résultats"
    data:
      - step_results
      - assertions_passed
      - screenshots
      - timing

  2_analyze:
    action: "Identifier échecs et causes"

  3_generate_report:
    format: "Structured test report"
```

**Output Final :**

```
═══════════════════════════════════════════════════════════════
  /test - Test Report
═══════════════════════════════════════════════════════════════

  URL: https://myapp.com/login
  Scenario: Login flow

  Steps:
    ✓ Navigate to /login (245ms)
    ✓ Fill email input (32ms)
    ✓ Fill password input (28ms)
    ✓ Click submit button (156ms)
    ✓ Verify dashboard redirect (1.2s)

  Assertions:
    ✓ Dashboard visible
    ✓ Welcome message present
    ✓ User avatar displayed

  Artifacts:
    - Screenshot: /tmp/test-login-success.png
    - Trace: /tmp/trace-login.zip

  Result: PASS (5/5 steps, 3/3 assertions)

═══════════════════════════════════════════════════════════════
```

---

## Workflows

### --run (Execute project tests)

```yaml
run_workflow:
  1_peek:
    action: "Scan test files"
    tools: [Glob]
    patterns: ["**/*.spec.ts", "**/*.test.ts", "**/e2e/**"]

  2_decompose:
    action: "Categorize tests"
    categories:
      - unit: "**/unit/**"
      - integration: "**/integration/**"
      - e2e: "**/e2e/**"

  3_parallelize:
    action: "Run test suites in parallel"
    tools: [Task agents]

  4_synthesize:
    action: "Consolidated test report"
```

### --trace (Debug with tracing)

```yaml
trace_workflow:
  1_start:
    tool: mcp__playwright__browser_start_tracing
    params:
      name: "debug-session"

  2_interact:
    action: "Perform interactions"

  3_stop:
    tool: mcp__playwright__browser_stop_tracing
    output: "trace.zip (viewable in trace.playwright.dev)"
```

### --codegen (Generate test code)

```yaml
codegen_workflow:
  1_peek:
    action: "Analyze page structure"

  2_record:
    action: "Record interactions"

  3_synthesize:
    action: "Generate Playwright test code"
    output: "*.spec.ts file"
```

---

## MCP Tools Reference

### Navigation

| Tool | Description |
|------|-------------|
| `browser_navigate` | Ouvrir une URL |
| `browser_go_back` | Page précédente |
| `browser_go_forward` | Page suivante |
| `browser_reload` | Rafraîchir |

### Interaction

| Tool | Description |
|------|-------------|
| `browser_click` | Cliquer sur élément |
| `browser_type` | Saisir du texte |
| `browser_fill` | Remplir un champ |
| `browser_select_option` | Sélectionner option |
| `browser_hover` | Survoler élément |
| `browser_press_key` | Appuyer touche |

### Capture

| Tool | Description |
|------|-------------|
| `browser_snapshot` | Accessibility tree |
| `browser_screenshot` | Capture d'écran |
| `browser_pdf_save` | Générer PDF |

### Testing

| Tool | Description |
|------|-------------|
| `browser_expect` | Assertions |
| `browser_generate_locator` | Générer sélecteur |
| `browser_start_tracing` | Démarrer trace |
| `browser_stop_tracing` | Arrêter trace |

---

## GARDE-FOUS (ABSOLUS)

| Action | Status | Raison |
|--------|--------|--------|
| Skip Phase 1 (Peek/Snapshot) | ❌ **INTERDIT** | Analyser page avant interaction |
| Naviguer vers sites malveillants | ❌ **INTERDIT** | Sécurité |
| Saisir des credentials réels | ⚠ **WARNING** | Utiliser fixtures |
| Modifier données en production | ❌ **INTERDIT** | Environnement test only |

### Parallélisation légitime

| Élément | Parallèle? | Raison |
|---------|------------|--------|
| Étapes E2E (navigate→fill→click) | ❌ Séquentiel | Ordre d'interaction requis |
| Assertions finales indépendantes | ✅ Parallèle | Vérifications sans dépendance |
| Screenshots + validations | ✅ Parallèle | Opérations indépendantes |
