---
name: search
description: |
  Documentation Research with RLM (Recursive Language Model) patterns.
  LOCAL-FIRST: Searches internal docs (~/.claude/docs/) before external sources.
  Cross-validates sources, generates .context.md, handles conflicts.
  Use when: researching technologies, APIs, or best practices before implementation.
allowed-tools:
  - "WebSearch(*)"
  - "WebFetch(*)"
  - "Read(**/*)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "Write(.context.md)"
  - "Task(*)"
  - "AskUserQuestion(*)"
  - "mcp__context7__*"
  - "mcp__github__create_issue"
---

# Search - Documentation Research (RLM-Enhanced)

$ARGUMENTS

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Use `grepai_trace_callers`/`grepai_trace_callees` for impact analysis.
Fallback to Grep ONLY for exact string matches or regex patterns.

---

## Description

Recherche avec stratégie **LOCAL-FIRST** et patterns RLM.

### Priorité : Documentation locale validée

```
~/.claude/docs/ (LOCAL)  →  Sources officielles (EXTERNE)
     ✓ Validée              ⚠ Peut être obsolète
     ✓ Cohérente            ⚠ Peut contredire local
     ✓ Immédiate            ⚠ Nécessite validation
```

**Patterns RLM appliqués :**

- **Local-First** - Consultation `~/.claude/docs/` en priorité
- **Peek** - Aperçu rapide avant analyse complète
- **Grep** - Filtrage par keywords avant fetch sémantique
- **Partition+Map** - Recherches parallèles multi-domaines
- **Summarize** - Résumé progressif des sources
- **Conflict-Resolution** - Gestion des contradictions local/externe
- **Programmatic** - Génération structurée du context

**Principe** : Local > Externe. Fiabilité > Quantité.

---

## Arguments

| Pattern | Action |
|---------|--------|
| `<query>` | Nouvelle recherche sur le sujet |
| `--append` | Ajoute au contexte existant |
| `--status` | Affiche le contexte actuel |
| `--clear` | Supprime le fichier .context.md |
| `--help` | Affiche l'aide |

---

## --help

```
═══════════════════════════════════════════════
  /search - Documentation Research (RLM)
═══════════════════════════════════════════════

Usage: /search <query> [options]

Options:
  <query>           Sujet de recherche
  --append          Ajoute au contexte existant
  --status          Affiche le contexte actuel
  --clear           Supprime .context.md
  --help            Affiche cette aide

RLM Patterns (toujours appliqués):
  1. Peek    - Aperçu rapide des résultats
  2. Grep    - Filtrage par keywords
  3. Map     - 6 recherches parallèles
  4. Synth   - Synthèse multi-sources (3+ pour HIGH)

Exemples:
  /search OAuth2 avec JWT
  /search Kubernetes ingress --append
  /search --status

Workflow:
  /search <query> → itérer → EnterPlanMode
═══════════════════════════════════════════════
```

---

## Sources officielles (Whitelist)

**RÈGLE ABSOLUE** : UNIQUEMENT les domaines suivants.

### Langages
| Langage | Domaines |
|---------|----------|
| Node.js | nodejs.org, developer.mozilla.org |
| Python | docs.python.org, python.org |
| Go | go.dev, pkg.go.dev |
| Rust | rust-lang.org, doc.rust-lang.org |
| Java | docs.oracle.com, openjdk.org |
| C/C++ | cppreference.com, isocpp.org |

### Cloud & Infra

| Service | Domaines |
|---------|----------|
| AWS | docs.aws.amazon.com |
| GCP | cloud.google.com |
| Azure | learn.microsoft.com |
| Docker | docs.docker.com |
| Kubernetes | kubernetes.io |
| Terraform | developer.hashicorp.com |
| GitLab | docs.gitlab.com |
| GitHub | docs.github.com |

### Frameworks
| Framework | Domaines |
|-----------|----------|
| React | react.dev |
| Vue | vuejs.org |
| Next.js | nextjs.org |
| FastAPI | fastapi.tiangolo.com |

### Standards

| Type | Domaines |
|------|----------|
| Web | developer.mozilla.org, w3.org |
| Security | owasp.org |
| RFCs | rfc-editor.org, tools.ietf.org |

### Blacklist

- ❌ Blogs, Medium, Dev.to
- ❌ Stack Overflow (sauf identification problème)
- ❌ Tutoriels tiers, cours en ligne

---

## Workflow RLM (7 phases)

### Phase 1.0 : Documentation locale (LOCAL-FIRST)

**TOUJOURS exécuter en premier. La documentation locale est VALIDÉE et prioritaire.**

```yaml
local_first:
  source: "~/.claude/docs/"
  index: "~/.claude/docs/README.md"

  workflow:
    1_search_local:
      action: |
        Grep("~/.claude/docs/", pattern=<keywords>)
        Glob("~/.claude/docs/**/*.md", pattern=<topic>)
      output: [matching_files]

    2_read_matches:
      action: |
        POUR chaque matching_file:
          Read(matching_file)
          Extraire: définition, exemples, patterns liés
      output: local_knowledge

    3_evaluate_coverage:
      rule: |
        SI local_knowledge couvre >= 80% de la query:
          status = "LOCAL_COMPLETE"
          → Skip Phase 1-3, go to Phase 6
        SINON SI local_knowledge couvre >= 40%:
          status = "LOCAL_PARTIAL"
          → Continue Phase 0+ pour gaps uniquement
        SINON:
          status = "LOCAL_NONE"
          → Continue workflow normal

  categories_mapping:
    design_patterns: "creational/, structural/, behavioral/"
    performance: "performance/"
    concurrency: "concurrency/"
    enterprise: "enterprise/"
    messaging: "messaging/"
    ddd: "ddd/"
    functional: "functional/"
    architecture: "architectural/"
    cloud: "cloud/, resilience/"
    security: "security/"
    testing: "testing/"
    devops: "devops/"
    integration: "integration/"
    principles: "principles/"
```

**Output Phase 1.0 :**

```
═══════════════════════════════════════════════
  /search - Local Documentation Check
═══════════════════════════════════════════════

  Query    : <query>
  Keywords : <k1>, <k2>, <k3>

  Local Search (~/.claude/docs/):
    ├─ Matches: 3 files
    │   ├─ behavioral/observer.md (95% match)
    │   ├─ behavioral/README.md (70% match)
    │   └─ principles/solid.md (40% match)
    │
    └─ Coverage: 85% → LOCAL_COMPLETE

  Status: ✓ Using local documentation (validated)
  External search: SKIPPED (local sufficient)

═══════════════════════════════════════════════
```

**Si LOCAL_PARTIAL :**

```
═══════════════════════════════════════════════
  /search - Local Documentation Check
═══════════════════════════════════════════════

  Query    : "OAuth2 JWT authentication"
  Keywords : OAuth2, JWT, authentication

  Local Search (~/.claude/docs/):
    ├─ Matches: 1 file
    │   └─ security/README.md (50% match)
    │
    └─ Coverage: 50% → LOCAL_PARTIAL

  Status: ⚠ Partial local coverage
  Gaps identified:
    ├─ OAuth2 flow details (not in local)
    └─ JWT implementation specifics (not in local)

  Action: External search for gaps only

═══════════════════════════════════════════════
```

---

### Phase 2.0 : Décomposition (RLM Pattern: Peek + Grep)

**Analyser la query AVANT toute recherche :**

1. **Peek** - Identifier la complexité
   - Query simple (1 concept) → Phase 1 directe
   - Query complexe (2+ concepts) → Décomposer

2. **Grep** - Extraire les keywords
   ```
   Query: "OAuth2 avec JWT pour API REST"
   Keywords: [OAuth2, JWT, API, REST]
   Technologies: [OAuth2 → rfc-editor.org, JWT → tools.ietf.org]
   ```

3. **Parallélisation systématique**
   - Toujours lancer jusqu'à 6 Task agents en parallèle
   - Couvrir tous les domaines pertinents

**Output Phase 0 :**
```
═══════════════════════════════════════════════
  /search - RLM Decomposition
═══════════════════════════════════════════════

  Query    : <query>
  Keywords : <k1>, <k2>, <k3>

  Decomposition:
    ├─ Sub-query 1: <concept1> → <domain1>
    ├─ Sub-query 2: <concept2> → <domain2>
    └─ Sub-query 3: <concept3> → <domain3>

  Strategy: PARALLEL (6 Task agents max)

═══════════════════════════════════════════════
```

---

### Phase 3.0 : Recherche parallèle (RLM Pattern: Partition + Map)

**Pour chaque sous-query, lancer un Task agent :**

```
Task({
  subagent_type: "Explore",
  prompt: "Rechercher <concept> sur <domain>. Extraire: définition, usage, exemples.",
  model: "haiku"  // Rapide pour recherche
})
```

**IMPORTANT** : Lancer TOUS les agents dans UN SEUL message (parallèle).

**Exemple multi-agent :**
```
// Message unique avec 3 Task calls
Task({ prompt: "OAuth2 sur rfc-editor.org", ... })
Task({ prompt: "JWT sur tools.ietf.org", ... })
Task({ prompt: "REST API sur developer.mozilla.org", ... })
```

---

### Phase 4.0 : Peek des résultats

**Avant analyse complète, peek sur chaque résultat :**

1. Lire les 500 premiers caractères de chaque réponse
2. Vérifier la pertinence (score 0-10)
3. Filtrer les résultats non-pertinents (< 5)

```
Résultats agents:
  ✓ OAuth2 (score: 9) - RFC 6749 trouvé
  ✓ JWT (score: 8) - RFC 7519 trouvé
  ✗ REST (score: 3) - Résultat trop générique
    → Relancer avec query affinée
```

---

### Phase 5.0 : Fetch approfondi (RLM Pattern: Summarization)

**Pour les résultats pertinents, WebFetch avec summarization :**

```
WebFetch({
  url: "<url trouvée>",
  prompt: "Résumer en 5 points clés: 1) Définition, 2) Cas d'usage, 3) Implémentation, 4) Sécurité, 5) Exemples"
})
```

**Summarization progressive :**

- Niveau 1: Résumé par source (5 points)
- Niveau 2: Fusion des résumés (synthèse)
- Niveau 3: Context final (actionable)

---

### Phase 6.0 : Croisement et validation

| Situation | Confidence | Action |
|-----------|------------|--------|
| Local + 2+ externes confirment | HIGHEST | Inclure (local prioritaire) |
| Local seul | HIGH | Inclure (validé) |
| 3+ sources externes confirment | MEDIUM | Inclure + comparer avec local |
| 2 sources externes confirment | LOW | Inclure + warning |
| 1 source externe | VERIFY | Vérifier avec local |
| Sources contradictoires | CONFLICT | Résolution utilisateur |
| 0 source | NONE | Exclure |

**Détection contradictions LOCAL vs EXTERNE :**

```yaml
conflict_detection:
  trigger: |
    SI info_externe != info_locale:
      status = "CONFLICT"
      action = "user_resolution"

  comparison:
    - Versions/dates
    - Syntaxe/API
    - Breaking changes
    - Best practices

  priority_rule: |
    LOCAL est TOUJOURS considéré comme VALIDÉ.
    EXTERNE peut être obsolète ou incorrect.
```

---

### Phase 7.0 : Résolution des conflits (CONFLICT HANDLING)

**OBLIGATOIRE si conflit détecté entre documentation locale et externe.**

```yaml
conflict_resolution:
  step_1_notify_user:
    tool: AskUserQuestion
    prompt: |
      ⚠️ CONFLIT détecté entre documentation locale et externe

      **Sujet:** {topic}

      **Documentation locale (~/.claude/docs/):**
      {local_content}

      **Documentation externe ({source}):**
      {external_content}

      **Différence:**
      {diff_summary}

    questions:
      - question: "Comment résoudre ce conflit ?"
        header: "Résolution"
        options:
          - label: "Garder LOCAL"
            description: "La doc locale est correcte, ignorer l'externe"
          - label: "Mettre à jour LOCAL"
            description: "L'externe est plus récent, créer issue pour MAJ"
          - label: "Les deux valides"
            description: "Contextes différents, documenter les deux"

  step_2_create_issue:
    condition: "user_choice == 'Mettre à jour LOCAL'"
    tool: mcp__github__create_issue
    params:
      owner: "kodflow"
      repo: "devcontainer-template"
      title: "docs: Update {category}/{file} - conflict with official docs"
      body: |
        ## Conflict Report

        **Generated by:** `/search` skill
        **Date:** {ISO8601}

        ### Local Documentation
        **File:** `~/.claude/docs/{path}`
        **Content:**
        ```
        {local_excerpt}
        ```

        ### External Source
        **URL:** {external_url}
        **Content:**
        ```
        {external_excerpt}
        ```

        ### Difference
        {diff_description}

        ### Suggested Action
        - [ ] Review external source validity
        - [ ] Update local documentation if confirmed
        - [ ] Add version/date metadata

        ---
        _Auto-generated by /search conflict detection_
      labels:
        - "documentation"
        - "auto-generated"

  step_3_continue:
    action: |
      SI user_choice == "Garder LOCAL":
        → Utiliser info locale, ignorer externe
      SI user_choice == "Mettre à jour LOCAL":
        → Issue créée, utiliser externe avec warning
      SI user_choice == "Les deux valides":
        → Documenter les deux contextes
```

**Output Phase 7.0 :**

```
═══════════════════════════════════════════════
  /search - Conflict Resolution
═══════════════════════════════════════════════

  ⚠️ CONFLICT DETECTED

  Topic: Observer Pattern implementation

  Local (~/.claude/docs/behavioral/observer.md):
    → Uses EventEmitter interface
    → Recommends typed events

  External (developer.mozilla.org):
    → Uses addEventListener
    → Browser-specific API

  User Decision: "Les deux valides"
    → Local = Application patterns
    → External = Browser DOM events

  Issue: NOT CREATED (different contexts)

═══════════════════════════════════════════════
```

**Output si issue créée :**

```
═══════════════════════════════════════════════
  /search - Conflict Resolution
═══════════════════════════════════════════════

  ⚠️ CONFLICT DETECTED

  Topic: JWT expiration handling

  Local (~/.claude/docs/security/jwt.md):
    → Recommends 15min access token

  External (tools.ietf.org/html/rfc7519):
    → No specific recommendation

  User Decision: "Mettre à jour LOCAL"

  ✓ Issue created: kodflow/devcontainer-template#142
    Title: "docs: Update security/jwt.md - add RFC reference"

  Action: Using external info with warning

═══════════════════════════════════════════════
```

---

### Phase 8.0 : Questions (si nécessaire)

**UNIQUEMENT si ambiguïté détectée :**

```
AskUserQuestion({
  questions: [{
    question: "La query mentionne X et Y. Lequel prioriser ?",
    header: "Priorité",
    options: [
      { label: "X d'abord", description: "Focus sur X" },
      { label: "Y d'abord", description: "Focus sur Y" },
      { label: "Les deux", description: "Recherche complète" }
    ]
  }]
})
```

**NE PAS demander si :**

- Query claire et non-ambiguë
- Une seule technologie
- Contexte suffisant

---

### Phase 9.0 : Génération context.md (RLM Pattern: Programmatic)

**Générer le fichier de manière structurée :**

```markdown
# Context: <sujet>

Generated: <ISO8601>
Query: <query>
Iterations: <n>
RLM-Depth: <parallel_agents_count>

## Summary

<2-3 phrases résumant les findings>

## Key Information

### <Concept 1>

<Information validée>

**Sources:**
- [<Titre>](<url>) - "<extrait>"
- [<Titre2>](<url>) - "<confirmation>"

**Confidence:** HIGH

### <Concept 2>

<Information>

**Sources:**
- [<Titre>](<url>)

**Confidence:** MEDIUM

## Clarifications

| Question | Réponse |
|----------|---------|
| <Q1> | <R1> |

## Recommendations

1. <Recommandation actionable>
2. <Recommandation actionable>

## Warnings

- ⚠ <Point d'attention>

## Sources Summary

| Source | Domain | Confidence | Used In |
|--------|--------|------------|---------|
| RFC 6749 | rfc-editor.org | HIGH | §1 |
| RFC 7519 | tools.ietf.org | HIGH | §2 |

---
_Généré par /search (RLM-enhanced). Ne pas commiter._
```

---

## --append

Enrichir le contexte existant :

1. Lire `.context.md` existant
2. Identifier les gaps (sections manquantes)
3. Rechercher uniquement les gaps
4. Fusionner sans duplicata

---

## --status / --clear

Identique à la version précédente.

---

## GARDE-FOUS

| Action | Status |
|--------|--------|
| Skip Phase 1.0 (documentation locale) | ❌ **INTERDIT** |
| Ignorer conflit local/externe | ❌ **INTERDIT** |
| Préférer externe sur local sans validation | ❌ **INTERDIT** |
| Source non-officielle | ❌ INTERDIT |
| Skip Phase 2.0 (décomposition) | ❌ INTERDIT |
| Agents séquentiels si parallélisable | ❌ INTERDIT |
| Info sans source | ❌ INTERDIT |

**RÈGLE ABSOLUE LOCAL-FIRST :**

```yaml
local_first_rule:
  priority: "LOCAL > EXTERNE"
  reason: "Documentation locale est validée et cohérente"

  workflow:
    1: "TOUJOURS chercher dans ~/.claude/docs/ d'abord"
    2: "SI local suffisant → utiliser local uniquement"
    3: "SI conflit → demander à l'utilisateur"
    4: "SI mise à jour nécessaire → créer issue GitHub"
```

---

## Exemples d'exécution

### Query simple

```
/search "Go context package"

→ 1 concept, 1 domaine (go.dev)
→ WebSearch + WebFetch direct
→ Validation 3+ sources
```

### Query complexe

```
/search "OAuth2 JWT authentication pour API REST"

→ 4 concepts, 3 domaines
→ 6 Task agents parallèles
→ Fetch références croisées
→ Synthèse RLM (3+ sources pour HIGH)
```

### Query multi-domaines

```
/search "Kubernetes ingress controller comparison"

→ 6 Task agents parallèles
→ Couverture: kubernetes.io, docs.docker.com, cloud.google.com
→ Validation stricte 3+ sources
```
