# Improve - Continuous Enhancement (RLM Multi-Agent)

## Description

Amélioration continue automatique. Détecte le contexte et agit.

```
/improve
```

**Pas d'arguments.** Le skill détecte automatiquement le mode.

---

## Modes (Auto-détectés)

```
┌─────────────────────────────────────────────────────────────────────┐
│                            /improve                                  │
├──────────────────────────────┬──────────────────────────────────────┤
│  kodflow/devcontainer-template│  Autre projet                       │
│  ══════════════════════════════│══════════════════════════════════  │
│                               │                                      │
│  → Améliorer ~/.claude/docs/    │  → Analyser le code                 │
│    ├─ MAJ best practices      │    ├─ Détecter anti-patterns        │
│    ├─ Corriger incohérences   │    ├─ Comparer avec ~/.claude/docs/   │
│    ├─ Affiner exemples        │    ├─ Trouver bonnes pratiques      │
│    └─ WebSearch validations   │    └─ Créer issues sur template     │
│                               │                                      │
│  Output: Fichiers modifiés    │  Output: Issues GitHub créées       │
└──────────────────────────────┴──────────────────────────────────────┘
```

---

## Workflow RLM

### Phase 1 : Détection du contexte

```yaml
detection:
  command: git remote get-url origin 2>/dev/null

  rules:
    - if: "contains 'kodflow/devcontainer-template'"
      mode: "DOCS_IMPROVEMENT"
      scope: "~/.claude/docs/**/*.md"
      action: "Améliorer la documentation patterns"

    - else:
      mode: "ANTI_PATTERN_DETECTION"
      scope: "**/*.{md,ts,js,py,go,rs,java,rb,php}"
      action: "Détecter violations et créer issues"
      target: "github.com/kodflow/devcontainer-template/issues"
```

---

### Phase 2 : Inventaire (Partition)

```yaml
inventory:
  mode_docs:
    action: Glob("~/.claude/docs/**/*.md")
    group_by: category (principles, creational, behavioral, etc.)

  mode_antipattern:
    action: |
      Glob("**/*.md")
      Glob("**/*.{ts,js,py,go,rs,java,rb,php}")
    group_by: file_type
```

---

### Phase 3 : Agents parallèles (Map)

**Lance 1 agent par fichier, max 20 en parallèle.**

```yaml
parallel_execution:
  max_agents: 20
  model: haiku  # Fast

  mode_docs:
    prompt_per_file: |
      FICHIER: {path}
      CATÉGORIE: {category}

      TÂCHES:
      1. Lire le contenu actuel
      2. Identifier améliorations possibles:
         - Info obsolète
         - Exemples manquants
         - Incohérences
      3. WebSearch "{pattern} best practices 2024"
      4. Proposer corrections

      OUTPUT JSON:
      {
        "file": "{path}",
        "status": "OK | UPDATE | OUTDATED",
        "improvements": [{
          "type": "content | example | fix",
          "current": "...",
          "proposed": "...",
          "source": "url"
        }]
      }

  mode_antipattern:
    prompt_per_file: |
      FICHIER: {path}
      RÉFÉRENCE: ~/.claude/docs/

      TÂCHES:
      1. Lire le code
      2. Comparer avec patterns documentés
      3. Détecter:
         - Violations (anti-patterns)
         - Patterns manquants
         - Bonnes pratiques à documenter

      OUTPUT JSON:
      {
        "file": "{path}",
        "violations": [{
          "pattern": "name",
          "severity": "HIGH | MEDIUM | LOW",
          "description": "...",
          "code": "...",
          "fix": "..."
        }],
        "positive": [{
          "description": "...",
          "code": "...",
          "worth_documenting": true
        }]
      }
```

---

### Phase 4 : Validation (WebSearch)

```yaml
validation:
  for_each_improvement:
    search: "{pattern} {year} best practices"
    sources:
      - Official docs (go.dev, docs.python.org, etc.)
      - martinfowler.com, refactoring.guru
      - owasp.org (security)

    confidence:
      - 3+ sources: VALIDATED
      - 2 sources: MEDIUM
      - 1 source: LOW (flag)
      - 0 source: SKIP
```

---

### Phase 5 : Application

```yaml
application:
  mode_docs:
    action: |
      POUR chaque improvement VALIDATED:
        Edit(file, old, new)
      Afficher résumé modifications

  mode_antipattern:
    action: |
      POUR chaque violation HIGH/MEDIUM:
        mcp__github__create_issue(
          owner: "kodflow",
          repo: "devcontainer-template",
          title: "pattern: {description}",
          body: "## Violation\n{details}\n## Code\n```\n{code}\n```\n## Fix\n{suggestion}",
          labels: ["documentation", "improvement", "auto-generated"]
        )

      POUR chaque positive worth_documenting:
        mcp__github__create_issue(
          title: "new-pattern: {description}",
          labels: ["new-pattern", "auto-generated"]
        )

      Afficher liste issues créées
```

---

### Phase 6 : Rapport

```yaml
report:
  mode_docs:
    output: |
      ═══════════════════════════════════════════════
        /improve - Documentation Enhancement Complete
      ═══════════════════════════════════════════════

        Files analyzed: {total}

        Results:
          ✓ OK: {ok}
          ⚠ Updated: {updated}
          ✗ Outdated: {outdated}

        Changes applied: {changes}

      ═══════════════════════════════════════════════

  mode_antipattern:
    output: |
      ═══════════════════════════════════════════════
        /improve - Anti-Pattern Detection Complete
      ═══════════════════════════════════════════════

        Repository: {repo}
        Files analyzed: {total}

        Violations:
          🔴 HIGH: {high}
          🟡 MEDIUM: {medium}
          🟢 LOW: {low}

        Positive patterns: {positive}

        Issues created: {issues}
          {issue_list}

      ═══════════════════════════════════════════════
```

---

## Catégories patterns (~/.claude/docs/)

| Catégorie | Scope |
|-----------|-------|
| principles | SOLID, DRY, KISS, YAGNI |
| creational | Factory, Builder, Singleton |
| structural | Adapter, Decorator, Proxy |
| behavioral | Observer, Strategy, Command |
| performance | Cache, Lazy Load, Pool |
| concurrency | Thread Pool, Actor, Mutex |
| enterprise | PoEAA (Martin Fowler) |
| messaging | EIP patterns |
| ddd | Aggregate, Entity, Repository |
| functional | Monad, Functor, Either |
| architectural | Hexagonal, CQRS |
| cloud | Circuit Breaker, Saga |
| resilience | Retry, Timeout, Bulkhead |
| security | OAuth, JWT, RBAC |
| testing | Mock, Stub, Fixture |
| devops | GitOps, IaC, Blue-Green |

---

## Détection violations

| Type | Description |
|------|-------------|
| SOLID_VIOLATION | God class, mauvais couplage |
| DRY_VIOLATION | Code dupliqué |
| MISSING_PATTERN | Pattern absent mais nécessaire |
| SECURITY | Failles, secrets hardcodés |
| PERFORMANCE | N+1, cache manquant |
| ERROR_HANDLING | Silent catch, retry manquant |

---

## GARDE-FOUS

| Action | Status |
|--------|--------|
| Modifier sans WebSearch validation | ❌ INTERDIT |
| Créer issue sans code excerpt | ❌ INTERDIT |
| Agents séquentiels (si parallélisable) | ❌ INTERDIT |
| Issues sur repo autre que template | ❌ INTERDIT |
