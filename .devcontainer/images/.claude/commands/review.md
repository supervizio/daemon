---
name: review
description: |
  AI-powered code review (RLM decomposition) for PRs/MRs or local diffs.
  Focus: correctness, security, design, quality, shell safety.
  15 phases, 5 agents (opus for deep analysis, haiku for patterns).
  Cyclic workflow: /review --loop for iterative perfection.
  Local-only output with /plan generation for /do execution.
allowed-tools:
  - "Bash(git *)"
  - "Bash(gh *)"
  - "Bash(glab *)"
  - "Read(**/*)"
  - "Glob(**/*)"
  - "Grep(**/*)"
  - "mcp__github__*"
  - "mcp__gitlab__*"
  - "mcp__codacy__*"
  - "mcp__grepai__*"
  - "Task(*)"
  - "TaskCreate(*)"
  - "TaskUpdate(*)"
  - "TaskList(*)"
  - "TaskGet(*)"
---

# Review - AI Code Review (RLM Architecture)

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Use `grepai_trace_callers`/`grepai_trace_callees` for impact analysis.
Fallback to Grep ONLY for exact string matches or regex patterns.

## Overview

Intelligent code review using **Recursive Language Model** decomposition:

| Phase | Name | Action |
|-------|------|--------|
| 0 | Context | Detect PR/MR, branch, CI status (GitHub/GitLab) |
| 0.5 | **Repo Profile** | Cache conventions, architecture, ownership (7d TTL) |
| 1 | Intent | Analyze PR/MR + **Risk Model** calibration |
| 1.5 | Describe | Auto-generate PR/MR description (drift detection) |
| 2 | Feedback | Collect ALL comments/reviews |
| 2.3 | **CI Diagnostics** | Extract CI failure context (conditional) |
| 2.5 | Questions | Handle human questions |
| 3 | Peek | Snapshot diff, categorize files, route agents |
| 4 | **Analyze** | **5 parallel agents** (correctness, security, design, quality, shell) |
| 4.7 | **Merge & Dedupe** | Normalize findings, deduplicate, require evidence |
| 5 | Challenge | Evaluate feedback relevance with context |
| 6 | Output | Generate LOCAL report + /plan file (no GitHub/GitLab post) |
| 6.5 | **Dispatch** | Route fixes to language-specialist via /do |
| 7 | **Cyclic** | Loop until perfect OR --loop limit |

**Principe RLM** : Peek → Decompose → Parallelize → Synthesize

**5 Agents (opus for reasoning, haiku for patterns):**
- `developer-executor-correctness` (opus) - Algorithmic errors, invariants
- `developer-executor-security` (opus) - Taint analysis, OWASP, supply chain
- `developer-executor-design` (opus) - Antipatterns, DDD, layering, SOLID
- `developer-executor-quality` (haiku) - Style, complexity, metrics
- `developer-executor-shell` (haiku) - Shell safety, Dockerfile, CI/CD

**Platform Support:** GitHub (PRs) + GitLab (MRs) - auto-detected from git remote.
**Output:** LOCAL only (no comments posted). Generates /plan for /do execution.

---

## Usage

```
/review                    # Single review (no loop, no fix)
/review --loop             # Cyclic review until PERFECT (infinite)
/review --loop 5           # Cyclic review (max 5 iterations)
/review --pr [number]      # Review specific PR (GitHub)
/review --mr [number]      # Review specific MR (GitLab)
/review --staged           # Review staged changes only
/review --file <path>      # Review specific file
/review --security         # Security-focused review only
/review --correctness      # Correctness-focused review only
/review --design           # Design/architecture review only
/review --quality          # Quality-focused review only
/review --triage           # Large PR/MR mode (>30 files or >1500 lines)
/review --describe         # Force auto-describe even if PR/MR has description
```

**Cyclic Workflow:**
```
/review --loop
    │
    ├─→ Phase 0-6: Full analysis (5 agents)
    ├─→ Phase 6.5: Generate /plan with fixes
    ├─→ /do executes fixes via language-specialist
    ├─→ Loop: re-review → fix → re-review
    └─→ Exit when: no HIGH/CRITICAL OR limit reached
```

---

## Budget Controller (OBLIGATOIRE)

```yaml
budget_controller:
  thresholds:
    normal_mode:
      max_files: 30
      max_lines: 1500
      max_comments_ingested: 80
    triage_mode:
      trigger: "files > 30 OR lines > 1500"
      action: "Focus sur: unresolved threads, lignes modifiées, security only"

  output_limits:
    critical: unlimited
    high: 10
    medium: 5
    low: 3

  comment_priority:
    1: "Unresolved threads"
    2: "Comments on modified lines"
    3: "Human reviews"
    4: "AI bot suggestions"
```

**Décision automatique :**

| Situation | Mode |
|-----------|------|
| diff < 1500 lines, files < 30 | NORMAL |
| diff >= 1500 OR files >= 30 | TRIAGE |
| comments > 80 | FILTER (unresolved + modified lines only) |

---

## Phase 1.0 : Context Detection

**Identifier le contexte d'exécution (GitHub/GitLab auto-détecté) :**

```yaml
context_detection:
  1_git_state:
    tools:
      - "git remote -v"
      - "git branch --show-current"
      - "git rev-parse --abbrev-ref HEAD@{upstream} 2>/dev/null || echo 'no-upstream'"
    output:
      current_branch: string
      upstream: string
      remote: "origin" | "upstream"

  1.5_platform_detection:
    rule: |
      remote_url = git remote get-url origin
      SI remote_url contains "github.com":
        platform = "github"
        mcp_prefix = "mcp__github__"
      SINON SI remote_url contains "gitlab.com" OR "gitlab.":
        platform = "gitlab"
        mcp_prefix = "mcp__gitlab__"
      SINON:
        platform = "local"
        mcp_prefix = null
    output:
      platform: "github" | "gitlab" | "local"
      mcp_prefix: string | null

  2_pr_mr_detection:
    tools:
      github: "mcp__github__list_pull_requests(head: current_branch)"
      gitlab: "mcp__gitlab__list_merge_requests(source_branch: current_branch)"
    output:
      on_pr_mr: boolean
      pr_mr_number: number | null
      pr_mr_url: string | null
      target_branch: string  # base branch

  3_diff_source:
    rule: |
      SI on_pr_mr == true:
        source = "PR/MR diff via MCP"
        base = target_branch
        head = current_branch
      SINON:
        source = "local diff"
        base = "git merge-base origin/main HEAD"
        head = "HEAD"
    output:
      diff_source: "pr" | "mr" | "local"
      merge_base: string
      display: "Reviewing: {base}...{head}"

  4_ci_status:
    condition: "on_pr_mr == true"
    tools:
      github: "mcp__github__get_pull_request_status"
      gitlab: "mcp__gitlab__list_pipelines(ref: current_branch)"
    strategy:
      max_polls: 2
      poll_interval: 30s
      on_pending: "Continue avec warning 'CI pending'"
      on_failure: "Signaler dans report, ne pas bloquer"
    output:
      ci_status: "passing|pending|failing|unknown"
      ci_jobs: [{name, status, conclusion}]
```

**Output Phase 0 (GitHub) :**

```
═══════════════════════════════════════════════════════════════
  /review - Context Detection
═══════════════════════════════════════════════════════════════

  Platform: GitHub
  Branch: feat/post-compact-hook
  PR: #97 (open) → main
  Diff source: PR (mcp__github)
  Merge base: a60a896...847d6db

  CI Status: ✓ passing (3/3 jobs)
    ├─ build: passed (1m 23s)
    ├─ test: passed (2m 45s)
    └─ lint: passed (45s)

  Mode: NORMAL (18 files, 375 lines)

═══════════════════════════════════════════════════════════════
```

**Output Phase 0 (GitLab) :**

```
═══════════════════════════════════════════════════════════════
  /review - Context Detection
═══════════════════════════════════════════════════════════════

  Platform: GitLab
  Branch: feat/post-compact-hook
  MR: !42 (open) → main
  Diff source: MR (mcp__gitlab)
  Merge base: a60a896...847d6db

  CI Status: ✓ passed (pipeline #12345)
    ├─ build: passed (1m 23s)
    ├─ test: passed (2m 45s)
    └─ lint: passed (45s)

  Mode: NORMAL (18 files, 375 lines)

═══════════════════════════════════════════════════════════════
```

---

## Phase 2.0 : Repo Profile (Cacheable)

**Build stable repo understanding BEFORE analysis:**

```yaml
repo_profile:
  cache:
    location: ".claude/.cache/repo_profile.json"
    key: "repo_profile@{default_branch}"
    ttl: "7 days"

  inputs:
    priority_files:
      - "README.md"
      - "CONTRIBUTING.md"
      - "ARCHITECTURE.md"
      - ".editorconfig"
      - ".golangci*"
      - ".eslintrc*"
      - "pyproject.toml"
      - "CODEOWNERS"
      - "~/.claude/docs/**"

  extract:
    languages: [string]
    build_tools: [string]
    test_frameworks: [string]
    lint_tools: [string]
    architecture_style: "hexagonal|layered|cqrs|microservices|monolith"
    error_conventions: [string]  # "wrap with fmt.Errorf", etc.
    naming_conventions: [string]
    ownership:
      codeowners_present: boolean
      owners_by_path: [{path, owners}]

  output:
    repo_profile_summary: "max 50 lines, JSON"

  usage: |
    Injected into EVERY agent so they:
    - Adapt checks to repo conventions
    - Avoid false positives on intentional patterns
    - Respect established style
```

---

## Phase 3.0 : Intent Analysis

**Comprendre l'intention de la PR/MR AVANT analyse lourde :**

```yaml
intent_analysis:
  inputs:
    github:
      - "mcp__github__get_pull_request (title, body, labels)"
      - "mcp__github__get_pull_request_files (file list)"
    gitlab:
      - "mcp__gitlab__get_merge_request (title, description, labels)"
      - "mcp__gitlab__get_merge_request_changes (file list)"
    common:
      - "git diff --stat"

  extract:
    title: string
    description: string (first 500 chars)
    labels: [string]
    files_changed: number
    lines_added: number
    lines_deleted: number
    directories_touched: [string]
    file_categories:
      security: count
      shell: count
      config: count
      tests: count
      docs: count
      code: count

  calibration:
    rule: |
      SI files_changed <= 5 AND only docs/config:
        analysis_depth = "light"
        skip_patterns = true
      SINON SI security_files > 0 OR shell_files > 0:
        analysis_depth = "deep"
        force_security_scan = true
      SINON:
        analysis_depth = "normal"

  risk_model:
    goal: "Identify critical zones BEFORE heavy analysis"

    risk_tags:
      - "authn_authz"      # auth, jwt, oauth, rbac, acl, session
      - "crypto"           # crypto, x509, tls, sign, encrypt, hash
      - "secrets"          # secret, token, key, vault, password
      - "network"          # http, grpc, tcp, udp, dns, socket
      - "db_migrations"    # migrate, schema, sql, gorm, prisma
      - "concurrency"      # goroutine, mutex, channel, lock, atomic
      - "supply_chain"     # Dockerfile, go.sum, package-lock
      - "state_machine"    # state, transition, fsm, workflow
      - "pagination"       # cursor, offset, limit, page
      - "caching"          # cache, ttl, invalidate, redis

    calibration:
      rule: |
        SI any(risk_tags):
          analysis_depth = "deep"
          prioritize_files = "risk-touched first"
          enable_agents = ["correctness", "security", "design"]
        SI risk_tags contains ["authn_authz", "crypto", "secrets"]:
          force_security_deep = true
        SI risk_tags contains ["concurrency", "state_machine"]:
          force_correctness_deep = true

    output:
      risk_tags: [string]
      risk_files: [{path, risk_tags}]
      review_priorities: ["correctness", "security", "design", "quality"]
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════════════════════
  /review - Intent Analysis
═══════════════════════════════════════════════════════════════

  Title: feat(hooks): add SessionStart hook
  Labels: [enhancement]

  Scope:
    ├─ Files: 18 (+375, -12)
    ├─ Dirs: .devcontainer/, .claude/
    └─ Categories: shell(2), config(3), code(13)

  Calibration:
    ├─ Depth: DEEP (shell files detected)
    ├─ Security scan: FORCED
    └─ Pattern analysis: CONDITIONAL

═══════════════════════════════════════════════════════════════
```

---

## Phase 4.0 : Auto-Describe (PR-Agent inspired)

**Générer description si PR/MR vide ou insuffisante :**

```yaml
auto_describe:
  trigger:
    - "pr_mr_body is empty OR pr_mr_body.length < 50"
    - "pr_mr_body contains only template placeholders"
    - "--describe flag passed"

  skip_if:
    - "pr_mr_body.length >= 200 AND contains_summary_section"
    - "mode == 'local' (no PR/MR)"

  workflow:
    1_analyze_diff:
      inputs:
        github: "mcp__github__get_pull_request_files"
        gitlab: "mcp__gitlab__get_merge_request_changes"
        common:
          - "git diff --stat"
          - "git log --oneline {base}..HEAD"
      extract:
        main_changes: [string]  # Max 5 key changes
        breaking_changes: [string]
        new_features: [string]
        bug_fixes: [string]
        refactors: [string]

    2_generate_description:
      format: |
        ## Summary
        {1-2 sentence overview based on commit messages + diff}

        ## Changes
        {bulleted list of main_changes, max 5}

        ## Type
        - [ ] Feature
        - [ ] Bug fix
        - [ ] Refactor
        - [ ] Documentation
        - [ ] Configuration

        ## Checklist
        - [ ] Tests added/updated
        - [ ] Documentation updated (if needed)
        - [ ] No breaking changes (or documented below)

      constraints:
        max_length: 1000
        no_code_blocks_in_summary: true
        no_file_by_file_description: true  # Avoid verbose output

    3_user_validation:
      tool: AskUserQuestion
      prompt: |
        📝 Description générée pour {PR|MR} #{pr_mr_number}:

        {generated_description}

        Action?
      options:
        - label: "Poster"
          description: "Mettre à jour la description"
        - label: "Éditer"
          description: "Modifier avant de poster"
        - label: "Ignorer"
          description: "Ne pas modifier"

    4_update_pr_mr:
      condition: "user_choice in ['Poster', 'Éditer']"
      tools:
        github: "gh pr edit {pr_number} --body '{final_description}'"
        gitlab: "glab mr update {mr_number} --description '{final_description}'"
      fallback:
        github: "mcp__github__update_pull_request (body: final_description)"
        gitlab: "mcp__gitlab__update_merge_request (description: final_description)"
```

**Output Phase 1.5 :**

```
═══════════════════════════════════════════════════════════════
  /review - Auto-Describe
═══════════════════════════════════════════════════════════════

  PR Description: EMPTY (0 chars)
  Action: Generate description

  Generated:
    ## Summary
    Add SessionStart hook to restore context after compaction.

    ## Changes
    - Add post-compact.sh script for context restoration
    - Update settings.json with SessionStart hook config
    - Add hook documentation in CLAUDE.md

    ## Type: Feature

  Status: ⏳ Waiting user validation...

═══════════════════════════════════════════════════════════════
```

---

## Phase 5.0 : Feedback Collection

**Collecter les feedbacks avec budget et priorisation :**

```yaml
feedback_collection:
  1_fetch:
    tools:
      github:
        - "mcp__github__get_pull_request_reviews"
        - "mcp__github__get_pull_request_comments"
      gitlab:
        - "mcp__gitlab__list_merge_request_notes"
        - "mcp__gitlab__list_merge_request_discussions"
      common:
        - "mcp__codacy__codacy_list_pull_request_issues"

  2_budget_filter:
    rule: |
      SI count(all_feedback) > 80:
        filter = "unresolved + modified_lines_only"
      SINON:
        filter = "all"

  3_classify:
    method: |
      POUR chaque feedback:
        SI author.type == "Bot" OR author.login ends with "[bot]":
          category = "ai_review"
        SINON:
          category = "human_review"

        SI body contains "?" AND NOT suggestion_code:
          type = "question"
        SINON SI suggestion_code != null:
          type = "suggestion"
        SINON:
          type = "comment"

  4_prioritize:
    order:
      1: "unresolved human reviews"
      2: "questions (need response)"
      3: "suggestions on modified lines"
      4: "ai reviews (behavior extraction)"
      5: "resolved/outdated"
```

**Classification output :**

| Category | Type | Count | Priority |
|----------|------|-------|----------|
| Human | Question | 2 | HIGH |
| Human | Comment | 3 | MEDIUM |
| AI (qodo) | Suggestion | 3 | EXTRACT |

---

## Phase 6.0 : CI Diagnostics (Conditional)

**Extract actionable signal from CI failures:**

```yaml
ci_diagnostics:
  trigger: "on_pr_mr == true AND ci_status in ['failing', 'pending']"

  goal: "Extract exploitable signal from CI failures without noise"

  tools:
    github:
      - "mcp__github__get_workflow_run_logs"
      - "gh run view --log-failed"
    gitlab:
      - "mcp__gitlab__get_pipeline_jobs"
      - "glab ci trace"

  extract:
    failing_jobs: [{name, conclusion, url}]
    top_errors: [string]  # max 5 representative lines
    affected_files: [string]
    error_categories:
      - "build_error"
      - "test_failure"
      - "lint_error"
      - "security_scan"
      - "timeout"

  output:
    ci_first_section: |
      IF failing:
        Prepend review with CI-First section
        Focus analysis on affected_files first

    rule: |
      SI ci_status == "failing":
        priority = ["fix CI errors", "then review rest"]
        inject_ci_context = true
      SI ci_status == "pending":
        warning = "CI still running, results may change"
```

---

## Phase 7.0 : Question Handling

**Préparer réponses pour questions humaines :**

```yaml
question_handling:
  rule_absolute: "JAMAIS mentionner AI/Claude/IA dans les réponses"

  forbidden_phrases:
    - "Claude", "AI", "IA", "assistant", "LLM"
    - "j'ai été généré", "automatiquement généré"
    - "l'intelligence artificielle suggère"

  workflow:
    1_collect: "Extraire questions des reviews humains"
    2_prepare: |
      POUR chaque question:
        answer = generate_answer(question, context)
        validate: no_forbidden_phrases(answer)

    3_present:
      format: |
        ## Question by {author}
        > {question_text}

        **Proposed answer:**
        {answer}

        [Post / Edit / Skip]

    4_user_validates: "AskUserQuestion avant de poster"
    5_post:
      github: "mcp__github__add_issue_comment (si validé)"
      gitlab: "mcp__gitlab__create_merge_request_note (si validé)"
```

---

## Phase 8.0 : Behavior Extraction (AI Reviews)

**Extraire axes comportementaux des reviews AI :**

```yaml
behavior_extraction:
  filter:
    - "importance >= 6/10"
    - "not already in workflow"
    - "actionable pattern"

  extract:
    from: "{bot_suggestion_text}"
    to:
      behavior: "description courte du pattern"
      category: "shell_safety|security|quality|pattern"
      check: "question à ajouter au workflow"

  action:
    auto: false
    prompt_user: |
      Nouveau pattern détecté:
        Behavior: {behavior}
        Category: {category}

      Ajouter au workflow /review? [Oui/Non]
```

---

## Phase 9.0 : Peek & Decompose

**Snapshot du diff et catégorisation :**

```yaml
peek_decompose:
  1_diff_snapshot:
    tool: |
      SI diff_source == "pr" (GitHub):
        mcp__codacy__codacy_get_pull_request_git_diff
      SINON SI diff_source == "mr" (GitLab):
        mcp__gitlab__get_merge_request_changes
      SINON:
        git diff --merge-base {base}...HEAD

    extract:
      files: [{path, status, additions, deletions}]
      total_lines: number
      hunks_count: number

  2_categorize:
    rules:
      security:
        patterns: ["auth", "crypto", "password", "token", "secret", "jwt"]
        extensions: [".go", ".py", ".js", ".ts", ".java"]
      shell:
        extensions: [".sh"]
        files: ["Dockerfile", "Makefile"]
      config:
        extensions: [".json", ".yaml", ".yml", ".toml"]
        files: ["mcp.json", "settings.json", "*.config.*"]
      tests:
        patterns: ["*_test.*", "*.test.*", "*.spec.*", "test_*"]
      docs:
        extensions: [".md"]

  3_mode_decision:
    rule: |
      SI total_lines > 1500 OR files.count > 30:
        mode = "TRIAGE"
      SINON:
        mode = "NORMAL"
```

---

## Phase 10.0 : Parallel Analysis (5 AGENTS)

**Launch 5 sub-agents with strict JSON contract:**

```yaml
parallel_analysis:
  dispatch:
    mode: "parallel (single message, 5 Task calls)"
    agents:
      correctness:
        name: "developer-executor-correctness"
        model: opus
        trigger: "always (MANDATORY for code stability)"
        focus:
          - "Algorithmic errors (off-by-one, bounds, indexes)"
          - "Invariant violations"
          - "State machine correctness"
          - "Concurrency issues (races, deadlocks)"
          - "Error surfacing (silent failures)"
          - "Idempotence violations"
          - "Ordering/determinism issues"

      security:
        name: "developer-executor-security"
        model: opus
        trigger: "always"
        focus:
          - "OWASP Top 10"
          - "Taint analysis (source → sink)"
          - "Supply chain risks"
          - "AuthN/AuthZ issues"
          - "Crypto misuse"
          - "Secrets exposure"

      design:
        name: "developer-executor-design"
        model: opus
        trigger: "risk_tags contains architecture OR files in core/, domain/, pkg/"
        focus:
          - "Antipatterns (God object, Feature envy, etc.)"
          - "DDD violations"
          - "Layering violations"
          - "SOLID violations"
          - "Design pattern misuse"

      quality:
        name: "developer-executor-quality"
        model: haiku
        trigger: "always"
        focus:
          - "Complexity metrics"
          - "Code duplication"
          - "Style issues"
          - "DTO convention check"

      shell:
        name: "developer-executor-shell"
        model: haiku
        trigger: "shell_files > 0 OR Dockerfile exists OR ci_config exists"
        focus:
          - "Shell safety (6 axes)"
          - "Dockerfile best practices"
          - "CI/CD script safety"

  agent_contract:
    input:
      files: [string]
      diff: string
      mode: "normal|triage"
      repo_profile: object  # From Phase 0.5

    output_schema:
      agent: string
      summary: string (max 200 chars)
      findings:
        - severity: "CRITICAL|HIGH|MEDIUM|LOW"
          impact: "correctness|security|design|quality|shell"
          category: string (ex: "injection", "invariant", "antipattern")

          # Location
          file: string
          line: number
          in_modified_lines: boolean

          # Description
          title: string (max 80 chars)
          evidence: string (MANDATORY, max 300 chars, NO SECRETS)

          # For correctness/security
          oracle: "invariant|counterexample|boundary|error-surfacing|taint"
          failure_mode: string (what can go wrong)
          repro: string (scenario: input → expected vs actual)

          # For security
          source: string (taint origin)
          sink: string (vulnerable point)
          taint_path_summary: string
          references: ["CWE-XX", "OWASP-AXX"]

          # Fix
          recommendation: string (MANDATORY)
          fix_patch: string (MANDATORY for HIGH+)
          effort: "XS|S|M|L"
          confidence: "HIGH|MEDIUM|LOW"

      commendations: [string]
      metrics:
        files_scanned: number
        findings_count: number

  severity_rubric:
    CRITICAL:
      - "Vuln exploitable (RCE, injection, auth bypass)"
      - "Secret/token exposé"
      - "Supply chain non vérifiée"
      - "Data loss certain (invariant violation)"
      - "Infinite loop (pagination bug)"
    HIGH:
      - "Bug probable (null deref, race condition)"
      - "Silent failure (error swallowed)"
      - "Layering violation (domain → infra)"
      - "State machine corruption"
    MEDIUM:
      - "Dette technique"
      - "Design antipattern"
      - "SOLID violation"
      - "Missing validation"
    LOW:
      - "Style/polish"
      - "Maintainability antipattern"
      - "Naming conventions"
```

**Secret Masking Policy (OBLIGATOIRE) :**

```yaml
secret_masking:
  rule: "JAMAIS reposter tokens/secrets/URLs signées"

  patterns_to_mask:
    - "AKIA[0-9A-Z]{16}"           # AWS Access Key
    - "ghp_[a-zA-Z0-9]{36}"        # GitHub PAT
    - "sk-[a-zA-Z0-9]{48}"         # OpenAI key
    - "eyJ[a-zA-Z0-9_-]+\\.[a-zA-Z0-9_-]+\\.[a-zA-Z0-9_-]+"  # JWT
    - "-----BEGIN.*PRIVATE KEY-----"
    - "Bearer [a-zA-Z0-9._-]+"

  action: "Remplacer par [REDACTED] dans evidence/recommendation"
```

---

## Phase 11.0 : Merge & Dedupe

**Normalize, deduplicate, require evidence:**

```yaml
merge_dedupe:
  goal: "Normalize findings, remove duplicates, enforce evidence"

  inputs:
    - "correctness_agent.findings"
    - "security_agent.findings"
    - "design_agent.findings"
    - "quality_agent.findings"
    - "shell_agent.findings"

  normalize:
    required_fields:
      - severity
      - impact
      - category
      - file
      - line
      - title
      - evidence
      - recommendation
      - confidence

    optional_enriched:
      - oracle (correctness)
      - failure_mode (correctness)
      - repro (correctness)
      - source, sink (security)
      - taint_path_summary (security)
      - references (security/design)
      - fix_patch (all)
      - effort (all)

  drop_rules:
    - "evidence is missing OR evidence is empty"
    - "recommendation is missing OR recommendation is empty"
    - "impact == 'correctness' AND severity >= HIGH AND (repro is missing OR repro is empty) AND (failure_mode is missing OR failure_mode is empty)"
    - "impact == 'security' AND category == 'injection' AND severity >= HIGH AND (source is missing OR source is empty)"

  dedupe:
    key: "{impact}:{category}:{file}:{line}:{normalize(title)}"
    merge_strategy: "keep highest severity, merge evidence"

  promote:
    rule: |
      SI file has >= 3 MEDIUM findings in same impact:
        Create 1 HIGH umbrella finding
        Reference the 3 MEDIUM as sub-findings

  output:
    findings_normalized: [{...}]
    stats:
      total_before: number
      total_after: number
      dropped: number
      promoted: number
```

---

## Phase 12.0 : Challenge & Synthesize

**Évaluer pertinence avec NOTRE contexte :**

```yaml
challenge_feedback:
  timing: "APRÈS phases 3-4 (on a le contexte complet)"

  for_each_suggestion:
    evaluate:
      - "Dans le scope de la PR?"
      - "Applicable à notre stack/langage?"
      - "Pattern déjà implémenté ailleurs?"
      - "Trade-off conscient?"
      - "Suggestion générique vs cas spécifique?"

  classify:
    KEEP:
      action: "Intégrer dans findings"
      confidence: "HIGH"
    PARTIAL:
      action: "Signaler avec nuance"
      confidence: "MEDIUM"
    REJECT:
      action: "Ignorer avec raison"
      confidence: "LOW"
    DEFER:
      action: "Créer issue séparée"
      reason: "Hors scope PR"

  output_format:
    table:
      - suggestion: string
      - source: string (bot name)
      - verdict: "KEEP|PARTIAL|REJECT|DEFER"
      - rationale: string (1-2 lines)
      - action: "apply|issue|ignore"

  ask_user_if:
    - "Ambiguïté sur pertinence"
    - "Trade-off non documenté"
    - "Suggestion impacte architecture"
```

**Table de challenge :**

| Situation | Verdict | Action |
|-----------|---------|--------|
| Suggestion valide, applicable | KEEP | Apply now |
| Suggestion valide, hors scope | DEFER | Create issue |
| Suggestion générique, pas applicable | REJECT | Ignore + rationale |
| Trade-off conscient | REJECT | Document trade-off |
| Ambiguïté | ASK | User decision |

---

## Phase 13.0 : Output Generation (LOCAL ONLY)

**Generate LOCAL report + /plan file (NO GitHub/GitLab posting):**

```yaml
output_generation:
  mode: "LOCAL ONLY - No PR/MR comments"

  inputs:
    - findings_normalized: "Phase 4.7 output"
    - validated_suggestions: "Phase 5 KEEP items"
    - repo_profile: "Phase 0.5 output"

  outputs:
    1_terminal_report:
      format: "Markdown to terminal"
      sections:
        - summary
        - critical_issues
        - high_priority
        - medium (max 5)
        - low (max 3)
        - commendations
        - metrics

    2_plan_file:
      location: ".claude/plans/review-fixes-{timestamp}.md"
      content:
        header: |
          # Review Fixes Plan
          Generated: {timestamp}
          Branch: {branch}
          Files: {files_count}
          Findings: CRIT={n}, HIGH={n}, MED={n}

        sections:
          critical:
            title: "## Critical (MUST FIX)"
            items: |
              ### {title}
              - **File:** {file}:{line}
              - **Impact:** {impact}
              - **Evidence:** {evidence}
              - **Fix:** {fix_patch}
              - **Language:** {language}
              - **Specialist:** developer-specialist-{lang}

          high:
            title: "## High Priority"
            items: "Same format as critical"

          medium:
            title: "## Medium"
            items: "Same format, max 5"

  no_github_gitlab:
    rule: "NEVER post comments to PR/MR"
    reason: "Reviews are local, fixes via /do"
```

---

## Phase 14.0 : Language-Specialist Dispatch

**Route fixes to language-specialist agent via /do:**

```yaml
language_specialist_dispatch:
  goal: "Delegate fixes to language specialist"

  routing:
    ".go":    "developer-specialist-go"
    ".py":    "developer-specialist-python"
    ".java":  "developer-specialist-java"
    ".kt":    "developer-specialist-kotlin"
    ".kts":   "developer-specialist-kotlin"
    ".ts":    "developer-specialist-nodejs"
    ".js":    "developer-specialist-nodejs"
    ".rs":    "developer-specialist-rust"
    ".rb":    "developer-specialist-ruby"
    ".ex":    "developer-specialist-elixir"
    ".exs":   "developer-specialist-elixir"
    ".php":   "developer-specialist-php"
    ".scala": "developer-specialist-scala"
    ".cpp":   "developer-specialist-cpp"
    ".cc":    "developer-specialist-cpp"
    ".hpp":   "developer-specialist-cpp"
    ".c":     "developer-specialist-c"
    ".h":     "developer-specialist-c"
    ".dart":  "developer-specialist-dart"
    ".cs":    "developer-specialist-csharp"
    ".swift": "developer-specialist-swift"
    ".r":     "developer-specialist-r"
    ".R":     "developer-specialist-r"
    ".pl":    "developer-specialist-perl"
    ".pm":    "developer-specialist-perl"
    ".lua":   "developer-specialist-lua"
    ".f90":   "developer-specialist-fortran"
    ".f95":   "developer-specialist-fortran"
    ".f03":   "developer-specialist-fortran"
    ".adb":   "developer-specialist-ada"
    ".ads":   "developer-specialist-ada"
    ".cob":   "developer-specialist-cobol"
    ".cbl":   "developer-specialist-cobol"
    ".pas":   "developer-specialist-pascal"
    ".dpr":   "developer-specialist-pascal"
    ".vb":    "developer-specialist-vbnet"
    ".m":     "developer-specialist-matlab"
    ".asm":   "developer-specialist-assembly"
    ".s":     "developer-specialist-assembly"

  dispatch:
    command: "/do --plan .claude/plans/review-fixes-{timestamp}.md"
    executor: "developer-specialist-{lang}"

  integration_with_do:
    workflow:
      1: "/review generates plan with findings + fix_patch"
      2: "/do loads plan"
      3: "/do groups by language"
      4: "/do dispatches to language-specialist agents"
      5: "Language-specialists apply fixes"
      6: "/do returns control to /review"
      7: "If --loop, re-run /review"
```

---

## Phase 15.0 : Cyclic Validation

**Loop until perfect OR --loop limit:**

```yaml
cyclic_workflow:
  trigger: "/review --loop [N]"

  modes:
    no_flag: "Single review, no fix, no loop"
    loop_only: "--loop → Infinite until perfect"
    loop_N: "--loop 5 → Max 5 iterations"

  flow:
    iteration_1:
      1_review: "Full analysis (15 phases, 5 agents)"
      2_generate_plan: ".claude/plans/review-fixes-{timestamp}.md"
      3_dispatch_to_do: "/do --plan {plan_file}"

    iteration_2_to_N:
      1_review_validation: "/review (re-scan post-fix)"
      2_check_remaining:
        if: "findings.CRITICAL + findings.HIGH > 0"
        then: "Generate new plan, continue loop"
        else: "Exit loop (success)"
      3_check_loop_limit:
        if: "iteration >= N"
        then: "Exit loop (limit reached)"

  exit_conditions:
    - "No CRITICAL/HIGH findings remaining"
    - "--loop limit reached"
    - "User interrupt (Ctrl+C)"

  output_per_iteration:
    format: |
      ═══════════════════════════════════════════════════════════════
        Iteration {X}/{N}
        Findings: CRIT={a}, HIGH={b}, MED={c}, LOW={d}
        Fixes applied: {n} files modified
        Status: {CONTINUE|SUCCESS|LIMIT_REACHED}
      ═══════════════════════════════════════════════════════════════
```

---

## Output Format

```markdown
# Code Review: PR #{number}

## Summary
{1-2 sentences assessment}
Mode: {NORMAL|TRIAGE}
CI: {status}

## Critical Issues
> Must fix before merge

### [CRITICAL] `file:line` - Title
**Problem:** {description}
**Evidence:** {code snippet, REDACTED if secret}
**Fix:** {actionable recommendation}
**Confidence:** HIGH

## High Priority
> From our analysis + validated bot suggestions

## Medium
> Quality improvements (max 5)

## Low
> Style/polish (max 3)

## Shell Safety (si *.sh présents)

### Download Safety
| Check | Status | File:Line |
|-------|--------|-----------|

### Path Determinism
| Config | Issue | Fix |
|--------|-------|-----|

## Pattern Analysis (CONDITIONAL)
> Triggered only if: complexity ↑, duplication, or core/ touched

### Patterns Identified
| Pattern | Location | Status |

### Suggestions
| Problem | Pattern | Reference |

## Challenged Feedback
| Suggestion | Source | Verdict | Rationale |
|------------|--------|---------|-----------|

## Questions (pending)
| Author | Question | Proposed Answer |

## Commendations
> What's done well

## Metrics
| Metric | Value |
|--------|-------|
| Mode | NORMAL |
| Files reviewed | 18 |
| Lines | +375/-12 |
| Critical | 0 |
| High | 3 |
| Medium | 2 |
| Suggestions kept | 3/8 |
```

---

## Pattern Consultation (CONDITIONNELLE)

**Source :** `~/.claude/docs/` (Design Patterns Knowledge Base)

**Déclencher UNIQUEMENT si :**

```yaml
pattern_triggers:
  source: "~/.claude/docs/"
  index: "~/.claude/docs/README.md"

  conditions:
    - "complexity_increase > 20%"
    - "duplication_detected"
    - "directories: core/, domain/, pkg/, internal/"
    - "new_classes > 3"
    - "file size > 500 lines"

  skip_if:
    - "only docs/config changes"
    - "test files only"
    - "mode == TRIAGE"

  workflow:
    1_identify: "Lire ~/.claude/docs/README.md pour identifier catégorie"
    2_consult: "Read(~/.claude/docs/<category>/README.md)"
    3_analyze: "Vérifier patterns utilisés vs recommandés"
    4_report: "Inclure dans section 'Pattern Analysis'"

  language_aware:
    go: "No 'class' keyword, check interfaces/structs"
    ts_js: "Factory, Singleton, Observer patterns"
    python: "Metaclass, decorator patterns"
    shell: "N/A - skip pattern analysis"
```

---

## Shell Safety Checks (si *.sh présents)

```yaml
shell_safety_axes:
  1_download_safety:
    checks:
      - "mktemp pour fichiers temporaires?"
      - "curl --retry --proto '=https'?"
      - "install -m au lieu de chmod?"
      - "rm -f cleanup?"

  2_download_robustness:
    checks:
      - "Track échecs de download?"
      - "Exit si critique?"
      - "Évite silent failures?"

  3_path_determinism:
    checks:
      - "Chemins absolus dans configs?"
      - "Pas de dépendance PATH implicite?"

  4_fallback_completeness:
    checks:
      - "Fallback copie binaire au bon endroit?"

  5_input_resilience:
    checks:
      - "Gère entrée vide?"
      - "set -e avec handling graceful?"

  6_url_validation:
    checks:
      - "URL release existe?"
      - "Script officiel si dispo?"
```

---

## DTO Convention Check (Go files)

**Vérifier les DTOs Go utilisent `dto:"direction,context,security"`:**

```yaml
dto_validation:
  trigger: "*.go files in diff"
  severity: MEDIUM

  detection:
    suffixes:
      - Request
      - Response
      - DTO
      - Input
      - Output
      - Payload
      - Message
      - Event
      - Command
      - Query
    serialization_tags: ["json:", "yaml:", "xml:"]

  check: |
    Struct name matches *Request/*Response/*DTO/etc.
    AND has serialization tags
    → MUST have dto:"dir,ctx,sec" on each PUBLIC field

  valid_format: 'dto:"<direction>,<context>,<security>"'
  valid_values:
    direction: [in, out, inout]
    context: [api, cmd, query, event, msg, priv]
    security: [pub, priv, pii, secret]

  purpose: |
    Le tag dto:"..." exempte les structs de KTN-STRUCT-ONEFILE
    (groupement de plusieurs DTOs dans un même fichier autorisé)

  report_format: |
    ### DTO Convention
    | File | Struct | Status | Issue |
    |------|--------|--------|-------|
    | user_dto.go | CreateUserRequest | ✓ | - |
    | order.go | OrderResponse | ✗ | Missing dto:"..." tags |

  reference: "~/.claude/docs/conventions/dto-tags.md"
```

---

## Guard-rails

| Action | Status |
|--------|--------|
| Auto-approve/merge | FORBIDDEN |
| Skip security issues | FORBIDDEN |
| Modify code directly | FORBIDDEN |
| Post comment without user validation | FORBIDDEN |
| Mention AI in PR responses | **ABSOLUTE FORBIDDEN** |
| Skip Phase 0-1 (context/intent) | FORBIDDEN |
| Challenge without context (skip Phase 3-4) | FORBIDDEN |
| Expose secrets in evidence/output | FORBIDDEN |
| Ignore budget limits | FORBIDDEN |
| Pattern analysis on docs-only PR | SKIP (not forbidden) |

---

## No-Regression Checklist

```yaml
no_regression:
  check_in_pr:
    - "Tests ajoutés/ajustés pour changes?"
    - "Migration/rollback nécessaire?"
    - "Backward compatibility maintenue?"
    - "Config changes documentées?"
    - "Observability (logs/metrics) ajoutée?"
```

---

## Error Handling

```yaml
error_handling:
  mcp_rate_limit:
    action: "Backoff exponentiel (1s, 2s, 4s)"
    fallback: "git diff local"
    max_retries: 3

  agent_timeout:
    max_wait: 60s
    action: "Continue sans cet agent, signaler"

  large_diff:
    threshold: 5000 lines
    action: "Forcer TRIAGE mode, warning user"
```

---

## Agents Architecture

```
/review (15 phases, 5 agents)
    │
    ├─→ Phase 0-2.5: Context + Feedback (sequential)
    │     ├─→ 0: Context Detection (GitHub/GitLab auto)
    │     ├─→ 0.5: Repo Profile (cached 7d)
    │     ├─→ 1: Intent + Risk Model
    │     ├─→ 1.5: Auto-Describe (drift detection)
    │     ├─→ 2: Feedback Collection
    │     ├─→ 2.3: CI Diagnostics (conditional)
    │     └─→ 2.5: Question Handling
    │
    ├─→ Phase 3-4.7: Parallel Analysis
    │       │
    │       ├─→ 3: Peek & Route (categorize files)
    │       │
    │       ├─→ 4: PARALLEL (5 agents)
    │       │       │
    │       │       ├─→ developer-executor-correctness (opus)
    │       │       │     Focus: Invariants, bounds, state, concurrency
    │       │       │     Output: oracle, failure_mode, repro
    │       │       │
    │       │       ├─→ developer-executor-security (opus)
    │       │       │     Focus: OWASP, taint analysis, supply chain
    │       │       │     Output: source, sink, taint_path, CWE refs
    │       │       │
    │       │       ├─→ developer-executor-design (opus)
    │       │       │     Focus: Antipatterns, DDD, layering, SOLID
    │       │       │     Output: pattern_reference, official_reference
    │       │       │
    │       │       ├─→ developer-executor-quality (haiku)
    │       │       │     Focus: Complexity, duplication, style, DTOs
    │       │       │
    │       │       └─→ developer-executor-shell (haiku)
    │       │             Condition: *.sh OR Dockerfile exists
    │       │             Focus: 6 shell safety axes
    │       │
    │       └─→ 4.7: Merge & Dedupe (normalize, evidence-required)
    │
    ├─→ Phase 5: Challenge (with full context)
    │
    ├─→ Phase 6-6.5: Output (LOCAL ONLY)
    │       ├─→ 6: Generate report + /plan file
    │       └─→ 6.5: Dispatch to language-specialist via /do
    │
    └─→ Phase 7: Cyclic Validation (--loop)
          Loop: review → fix → review until perfect
```

---

## Review Iteration Loop

```yaml
iteration_loop:
  description: |
    Amélioration continue basée sur retours bots.

  process:
    1: "Collecter suggestions bots (Phase 2.6)"
    2: "Extraire COMPORTEMENT (pas le fix)"
    3: "Catégoriser (shell/security/quality)"
    4: "Ajouter au workflow (user approuve)"
    5: "Commit l'amélioration"

  example:
    input: "Use mktemp to prevent partial writes"
    output:
      behavior: "Downloads should use temp files"
      category: "shell_safety"
      axis: "1_download_safety"
```
