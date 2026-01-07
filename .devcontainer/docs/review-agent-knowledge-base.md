# Context: AI Code Review Agent - Best Practices

Generated: 2025-12-25T10:00:00Z
Query: Créer un agent de code review IA avec les meilleures pratiques (inspiré de Copilot/CodeRabbit)
Iterations: 2

## Summary

Un agent de code review efficace combine analyse statique multi-couches (sécurité, qualité, tests), intégration transparente dans le workflow développeur, et interaction intelligente humain-IA. Les leaders du marché (CodeRabbit, GitHub Copilot) utilisent des architectures basées sur des outils spécialisés, le croisement de sources (40+ linters), et des patterns agent loop avec vérification itérative.

---

## Key Information

### 1. Architecture Agent Recommandée

L'architecture optimale pour un agent de code review suit le pattern **Orchestrator-Workers** d'Anthropic, avec une boucle agent:

```
gather context → take action → verify work → repeat
```

**Composants clés:**
- **Agent principal** : Orchestre les sous-analyses et synthétise les résultats
- **Subagents spécialisés** : Sécurité, Qualité, Tests (parallélisation)
- **Compaction** : Résumé automatique pour maintenir le contexte
- **Verification loop** : Feedback règle-basé + LLM as Judge

**Sources:**
- [Building Effective Agents - Anthropic](https://www.anthropic.com/research/building-effective-agents) - "The orchestrator-workers pattern is ideal for complex tasks with unpredictable subtasks"
- [Building agents with Claude Agent SDK](https://www.anthropic.com/engineering/building-agents-with-the-claude-agent-sdk) - "Agent loop: gather context → take action → verify work → repeat"

**Confidence:** HIGH

---

### 2. Comparaison Linters : CodeRabbit vs Codacy

Les deux plateformes utilisent des outils similaires avec quelques différences:

#### Outils COMMUNS (présents dans les deux)
| Catégorie | Outils |
|-----------|--------|
| **JavaScript/TypeScript** | ESLint |
| **Python** | Pylint, Ruff, Bandit |
| **Ruby** | RuboCop, Brakeman |
| **Go** | (golangci-lint vs Revive/Staticcheck) |
| **PHP** | PHP_CodeSniffer, PHPMD |
| **Docker** | Hadolint |
| **IaC** | Checkov, Semgrep |
| **Swift** | SwiftLint |
| **Kotlin** | detekt |
| **Shell** | ShellCheck |

#### Outils SPÉCIFIQUES à CodeRabbit
| Outil | Usage |
|-------|-------|
| **Biome** | Linter JS/TS ultra-rapide (remplace ESLint+Prettier) |
| **oxlint** | Linter Rust pour JS/TS |
| **Clippy** | Linter Rust officiel |
| **ast-grep** | Pattern matching AST universel |
| **Gitleaks** | Détection secrets (vs Trivy chez Codacy) |
| **actionlint** | Validation GitHub Actions |
| **Buf** | Linting Protobuf |
| **LanguageTool** | Correction orthographe/grammaire |
| **markdownlint** | Linting Markdown |
| **SQLFluff** | Linting SQL |
| **Prisma Lint** | Validation schémas Prisma |

#### Outils SPÉCIFIQUES à Codacy
| Outil | Usage |
|-------|-------|
| **Trivy** | Scanner vulnérabilités (images, IaC, CVE) |
| **PMD** | Analyse Java/Apex |
| **SpotBugs + Find Security Bugs** | Analyse bytecode Java |
| **Checkstyle** | Style Java |
| **Clang-Tidy** | C/C++ |
| **Cppcheck** | C/C++ bugs |
| **Flawfinder** | Sécurité C/C++ |
| **SonarC#/SonarVB** | .NET |
| **Lizard** | Complexité multi-langages (15 langages) |
| **Prospector** | Python méta-linter |
| **Spectral** | OpenAPI/AsyncAPI |

**Sources:**
- [CodeRabbit Tools](https://docs.coderabbit.ai/tools) - "40+ third-party linters and security analysis tools"
- [Codacy Languages & Tools](https://docs.codacy.com/getting-started/supported-languages-and-tools/) - "Over 40 programming languages with specialized tools"

**Confidence:** HIGH

---

### 3. AXES D'ANALYSE EXHAUSTIFS (Focus par axe)

Voici **tous les axes d'analyse possibles**, chacun pouvant être utilisé indépendamment avec un flag dédié:

---

#### 🔴 AXE 1: SÉCURITÉ (`--security`)

**Sous-catégories OWASP:**

| Catégorie | Vérifications | Outils |
|-----------|---------------|--------|
| **Injection** | SQL, NoSQL, LDAP, XPath, OS Command, XSS | Semgrep, Bandit, Gosec |
| **Authentication** | Credentials hardcodées, JWT faibles, sessions | Gitleaks, Semgrep |
| **Secrets** | API keys, tokens, passwords, clés privées | Gitleaks, Trivy, TruffleHog |
| **Crypto** | Algorithmes faibles (MD5, SHA1), randomness | Semgrep, custom rules |
| **Access Control** | IDOR, privilege escalation, RBAC | Semgrep, analyse manuelle |
| **Dependencies** | CVE connues, packages vulnérables | Trivy, OSV-Scanner, Snyk |
| **Input Validation** | Sanitization manquante, type coercion | ESLint, Semgrep |
| **Logging** | Données sensibles dans logs, PII | Semgrep, custom rules |

**Sources:**
- [OWASP Code Review Guide](https://owasp.org/www-project-code-review-guide/) - "Covers 10+ key security areas"
- [OWASP Secure Coding Practices](https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/) - "Essential checklist items"

**Confidence:** HIGH

---

#### 🟡 AXE 2: QUALITÉ DE CODE (`--quality`)

**Métriques mesurables:**

| Métrique | Description | Seuil recommandé | Outil |
|----------|-------------|------------------|-------|
| **Cyclomatic Complexity** | Nombre de chemins indépendants | ≤10 (NIST) | Lizard, ESLint, Ruff |
| **Cognitive Complexity** | Effort mental de compréhension | ≤15 | SonarQube, Semgrep |
| **Maintainability Index** | Score global maintenabilité | ≥20 | Radon, Visual Studio |
| **Lines of Code (LOC)** | Longueur fonction/fichier | ≤300 lignes/fichier | Tous linters |
| **Depth of Nesting** | Niveaux d'imbrication | ≤4 | ESLint, Pylint |
| **Halstead Metrics** | Complexité cognitive basée sur opérateurs | Variable | Radon |

**Code Smells:**
- Functions trop longues (>50 lignes)
- Classes trop grandes (God objects)
- Code dupliqué (clones)
- Dead code / imports inutiles
- Magic numbers / strings
- Naming conventions violées
- Comments obsolètes

**Sources:**
- [Codacy Code Complexity](https://blog.codacy.com/code-complexity) - "Cyclomatic, cognitive, maintainability metrics"
- [Microsoft Code Metrics](https://learn.microsoft.com/en-us/visualstudio/code-quality/code-metrics-values) - "Maintainability Index formula"

**Confidence:** HIGH

---

#### 🧪 AXE 3: TESTS & COVERAGE (`--tests`)

| Métrique | Description | Seuil | Outil |
|----------|-------------|-------|-------|
| **Line Coverage** | % lignes exécutées | ≥80% | Istanbul, pytest-cov |
| **Branch Coverage** | % branches conditionnelles | ≥75% | Istanbul, coverage.py |
| **Function Coverage** | % fonctions appelées | ≥90% | Istanbul, go test |
| **Mutation Score** | % mutants tués | ≥70% | Stryker, PIT |

**Vérifications qualitatives:**
- Fonctions publiques sans tests
- Tests sans assertions significatives
- Edge cases non couverts
- Tests flaky (non déterministes)
- Mocks mal configurés
- Tests trop longs (>100 lignes)

**Sources:**
- [Codecov Mutation Testing](https://about.codecov.io/blog/mutation-testing-how-to-ensure-code-coverage-isnt-a-vanity-metric/) - "Mutation score is better than coverage"
- [BrowserStack Coverage Metrics](https://www.browserstack.com/guide/test-coverage-metrics-in-software-testing) - "Multi-metric approach"

**Confidence:** HIGH

---

#### 🏗️ AXE 4: ARCHITECTURE & DESIGN (`--architecture`)

**Patterns à détecter:**

| Catégorie | Vérifications | Outils |
|-----------|---------------|--------|
| **Dépendances circulaires** | Imports A→B→C→A | JDepend, madge, deptry |
| **Couplage excessif** | Trop de dépendances directes | NDepend, Structure101 |
| **Cohésion faible** | Classe avec responsabilités multiples | SonarQube, manual |
| **Layer violations** | UI→DB direct (bypass business layer) | ArchUnit, custom rules |
| **God objects** | Classes >500 lignes, >20 méthodes | Lizard, PMD |
| **Feature envy** | Méthode utilise trop une autre classe | PMD, SonarQube |
| **Design patterns** | Singleton, Factory, Observer mal implémentés | PINOT, Pattern4 |

**Anti-patterns:**
- Spaghetti code
- Big Ball of Mud
- Golden Hammer
- Lava Flow (dead code historique)
- Copy-Paste programming

**Sources:**
- [SEI Design Pattern Detection](https://www.sei.cmu.edu/blog/using-machine-learning-to-detect-design-patterns/) - "ML-based pattern detection"
- [MARPLE Tool](https://www.sciencedirect.com/science/article/abs/pii/S0020025510005955) - "Architecture reconstruction"

**Confidence:** MEDIUM (nécessite outils spécialisés)

---

#### 🐳 AXE 5: INFRASTRUCTURE AS CODE (`--infra`)

| Catégorie | Vérifications | Outils |
|-----------|---------------|--------|
| **Terraform** | State non chiffré, secrets hardcodés | Checkov, tfsec, Trivy |
| **Kubernetes** | Pods root, privileged containers | Checkov, Trivy, kubesec |
| **Docker** | Images non signées, user root | Hadolint, Trivy, Dockle |
| **CloudFormation** | IAM trop permissifs, S3 public | Checkov, cfn-lint |
| **Helm** | Values exposées, secrets en clair | Checkov, helm lint |
| **Policy-as-Code** | Violations CIS, NIST, PCI-DSS | OPA, Sentinel |

**Best practices:**
- Least privilege IAM
- Secrets via Vault/Secrets Manager
- Images minimales (distroless)
- Network policies restrictives
- Immutable infrastructure

**Sources:**
- [Terraform Security Best Practices](https://bridgecrew.io/blog/terraform-security-101-best-practices-for-secure-infrastructure-as-code/) - "IaC security scanning"
- [Cycode IaC Security](https://cycode.com/blog/8-best-practices-for-securing-infrastructure-as-code/) - "Policy-as-code approach"

**Confidence:** HIGH

---

#### ⚡ AXE 6: PERFORMANCE & CONCURRENCE (`--performance`)

| Catégorie | Vérifications | Outils |
|-----------|---------------|--------|
| **Memory leaks** | Références non libérées, closures | Valgrind, heaptrack |
| **N+1 queries** | Requêtes DB répétitives | Django-debug-toolbar, custom |
| **Blocking I/O** | sync dans async context | ESLint, Pylint |
| **Race conditions** | Accès concurrent non protégé | ThreadSanitizer, Go race detector |
| **Deadlocks** | Lock ordering incorrect | Helgrind, manual review |
| **Thread starvation** | Ressources monopolisées | Profilers, manual |
| **CPU hotspots** | Algorithmes O(n²) ou pire | Profilers, manual |

**Patterns concurrence:**
- Lock contention excessive
- Shared mutable state
- Improper synchronization
- Thread pool misconfiguration

**Sources:**
- [Microsoft Concurrency Patterns](https://learn.microsoft.com/en-us/visualstudio/profiling/common-patterns-for-poorly-behaved-multithreaded-applications) - "Common multithreading problems"
- [Easyperf MT Analysis](https://easyperf.net/blog/2019/10/05/Performance-Analysis-Of-MT-apps) - "Performance analysis methodology"

**Confidence:** MEDIUM (nécessite runtime analysis)

---

#### 📊 AXE 7: MAINTENABILITÉ ISO 25010 (`--maintainability`)

Basé sur la norme **ISO/IEC 25010:2023**:

| Caractéristique | Sous-caractéristiques | Métriques |
|-----------------|----------------------|-----------|
| **Modularity** | Indépendance des modules | Coupling, Cohesion |
| **Reusability** | Potentiel de réutilisation | Abstraction level |
| **Analysability** | Facilité de diagnostic | Cyclomatic complexity |
| **Modifiability** | Facilité de modification | Change impact |
| **Testability** | Facilité à tester | Test coverage, complexity |

**Sources:**
- [ISO 25010 Standard](https://iso25000.com/en/iso-25000-standards/iso-25010) - "8 quality characteristics, 31 sub-characteristics"
- [Codacy ISO 25010](https://blog.codacy.com/iso-25010-software-quality-model) - "Framework for evaluation"

**Confidence:** HIGH

---

#### 📝 AXE 8: DOCUMENTATION (`--docs`)

| Vérification | Description | Outils |
|--------------|-------------|--------|
| **JSDoc/Docstrings** | Fonctions publiques documentées | ESLint, Pylint |
| **README completeness** | Sections obligatoires présentes | custom rules |
| **API documentation** | OpenAPI/Swagger à jour | Spectral |
| **Comments quality** | Pas de TODOs abandonnés, outdated | ESLint, custom |
| **Changelog** | Mises à jour documentées | conventional-changelog |

**Confidence:** MEDIUM

---

#### 🚀 AXE 9: DÉPLOIEMENT & SCALABILITÉ (`--deployment`)

| Catégorie | Vérifications |
|-----------|---------------|
| **Stateless design** | Pas de state local, sessions externalisées |
| **12-Factor App** | Config en env vars, ports bindés |
| **Health checks** | Endpoints /health, /ready présents |
| **Graceful shutdown** | Gestion SIGTERM, drain connections |
| **Horizontal scaling** | Pas de singletons, shared state externalisé |
| **Observability** | Logging structuré, metrics, tracing |

**Confidence:** MEDIUM (analyse contextuelle)

---

#### 🎯 AXE 10: OBJECTIFS PROJET (`--objectives`)

Cet axe est **contextuel** et nécessite une configuration projet:

```yaml
# .review.yaml
objectives:
  performance:
    latency_p99: "< 100ms"
    throughput: "> 1000 rps"

  reliability:
    uptime: "99.9%"
    error_rate: "< 0.1%"

  scalability:
    target_users: 1_000_000
    horizontal: true

  tech_debt:
    max_complexity: 10
    min_coverage: 80%
```

**Confidence:** MEDIUM (nécessite config utilisateur)

---

### 4. Commandes Suggérées avec Axes

```bash
# Reviews par axe unique
/review --security           # Focus OWASP, secrets, CVE
/review --quality            # Complexity, code smells
/review --tests              # Coverage, mutation score
/review --architecture       # Couplage, patterns, design
/review --infra              # Terraform, K8s, Docker
/review --performance        # Memory, concurrency, N+1
/review --maintainability    # ISO 25010 metrics
/review --docs               # JSDoc, README, API docs
/review --deployment         # 12-factor, scalability
/review --objectives         # Basé sur .review.yaml

# Combinaisons
/review --security --quality # Sécurité + qualité
/review --all                # Analyse complète (défaut)
/review --quick              # Security + Quality seulement

# Modificateurs
/review --approve            # Auto-fix safe issues
/review --staged             # Staged changes only
/review --diff main          # Diff vs branche
```

**Confidence:** HIGH

---

### 5. Outils par Langage (Recommandation Synthèse)

| Langage | Quality | Security | Tests | Total |
|---------|---------|----------|-------|-------|
| **JavaScript/TS** | ESLint, Biome | Semgrep, Gitleaks | Jest/Vitest + Istanbul | 6+ |
| **Python** | Ruff, Pylint | Bandit, Semgrep | pytest + coverage | 5+ |
| **Go** | golangci-lint | Gosec, Semgrep | go test -cover | 4+ |
| **Java** | PMD, Checkstyle | SpotBugs, Semgrep | JUnit + JaCoCo | 6+ |
| **Rust** | Clippy | cargo-audit | cargo test + tarpaulin | 4+ |
| **C/C++** | Clang-Tidy, Cppcheck | Flawfinder, Semgrep | GoogleTest + gcov | 6+ |
| **Ruby** | RuboCop | Brakeman, Semgrep | RSpec + SimpleCov | 5+ |
| **PHP** | PHPCS, PHPMD | PHPStan, Semgrep | PHPUnit + coverage | 5+ |

**Confidence:** HIGH

---

## Clarifications

| Question | Réponse |
|----------|---------|
| Plateforme cible ? | Local (indépendant de GitHub/GitLab) |
| Niveau d'autonomie ? | Suggestions + validation humaine, sauf `--approve` |
| Types d'analyses ? | Toutes (10 axes identifiés) |
| Intégration ? | Slash command Claude Code `/review` |
| CodeRabbit vs Codacy ? | Outils similaires, CodeRabbit plus moderne (Biome, oxlint) |

---

## Recommendations

1. **Architecture multi-axe** : Implémenter chaque axe comme un subagent indépendant
2. **Flags composables** : Permettre `--security --quality` pour combiner
3. **Profils prédéfinis** : `--quick` (security+quality), `--full` (all)
4. **Configuration projet** : `.review.yaml` pour seuils et objectifs custom
5. **Rapport structuré** : Grouper par axe, puis par severity
6. **Caching intelligent** : Éviter de re-analyser fichiers non modifiés

---

## Warnings

- ⚠ **False positives** : Prévoir feedback loop pour marquer faux positifs
- ⚠ **Performance** : L'analyse complète peut prendre >1min sur gros projets
- ⚠ **Dépendances** : Certains outils nécessitent installation (trivy, semgrep)
- ⚠ **Contexte limité** : L'architecture nécessite analyse humaine en complément
- ⚠ **Objectifs projet** : L'axe `--objectives` requiert configuration

---

## Sources Summary

| Source | Domain | Confidence | Sections |
|--------|--------|------------|----------|
| [CodeRabbit Tools](https://docs.coderabbit.ai/tools) | coderabbit.ai | HIGH | §2 |
| [Codacy Languages](https://docs.codacy.com/getting-started/supported-languages-and-tools/) | codacy.com | HIGH | §2 |
| [OWASP Code Review](https://owasp.org/www-project-code-review-guide/) | owasp.org | HIGH | §3.1 |
| [ISO 25010](https://iso25000.com/en/iso-25000-standards/iso-25010) | iso25000.com | HIGH | §3.7 |
| [Codacy Complexity](https://blog.codacy.com/code-complexity) | codacy.com | HIGH | §3.2 |
| [Microsoft Code Metrics](https://learn.microsoft.com/en-us/visualstudio/code-quality/code-metrics-values) | microsoft.com | HIGH | §3.2 |
| [Codecov Mutation](https://about.codecov.io/blog/mutation-testing-how-to-ensure-code-coverage-isnt-a-vanity-metric/) | codecov.io | HIGH | §3.3 |
| [Terraform Security](https://bridgecrew.io/blog/terraform-security-101-best-practices-for-secure-infrastructure-as-code/) | bridgecrew.io | HIGH | §3.5 |
| [SEI Pattern Detection](https://www.sei.cmu.edu/blog/using-machine-learning-to-detect-design-patterns/) | cmu.edu | MEDIUM | §3.4 |
| [MS Concurrency](https://learn.microsoft.com/en-us/visualstudio/profiling/common-patterns-for-poorly-behaved-multithreaded-applications) | microsoft.com | MEDIUM | §3.6 |
| [Anthropic Agents](https://www.anthropic.com/research/building-effective-agents) | anthropic.com | HIGH | §1 |

---

### 6. TAXONOMIE DES LANGAGES (44 langages)

La distinction entre types de langages est **critique** car les axes d'analyse pertinents diffèrent:

```
┌─────────────────────────────────────────────────────────────────┐
│                    TAXONOMIE DES LANGAGES                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  🔵 PROGRAMMING LANGUAGES (24)                                   │
│     → Tous les axes applicables                                  │
│     → Security, Quality, Tests, Architecture, Performance        │
│                                                                  │
│  🟢 MARKUP LANGUAGES (3)                                         │
│     → Structure/Validation seulement                             │
│     → Pas de tests, pas d'architecture                           │
│                                                                  │
│  🟡 DATA/CONFIG LANGUAGES (4)                                    │
│     → Validation schéma + Security (secrets)                     │
│     → Pas de tests, pas d'architecture                           │
│                                                                  │
│  🟣 STYLE LANGUAGES (3)                                          │
│     → Quality (conventions) + Performance                        │
│     → Pas de tests, pas de sécurité                              │
│                                                                  │
│  🟠 INFRASTRUCTURE AS CODE (3)                                   │
│     → Security critique + Best practices                         │
│     → Architecture infra                                         │
│                                                                  │
│  🔘 QUERY LANGUAGES (4)                                          │
│     → Security (injection) + Performance                         │
│     → Pas de tests unitaires classiques                          │
│                                                                  │
│  ⚪ TEMPLATING LANGUAGES (3)                                     │
│     → Security (XSS) + Conventions                               │
│     → Contexte d'exécution limité                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

### 6.1 🔵 PROGRAMMING LANGUAGES (24 langages)

**Caractéristiques:** Code exécutable, logique métier, tous les axes applicables.

| Langage | Quality | Security | Tests | Architecture | Performance | Outils |
|---------|:-------:|:--------:|:-----:|:------------:|:-----------:|--------|
| **JavaScript** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | ESLint, Biome, Semgrep |
| **TypeScript** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | ESLint, tsc, Semgrep |
| **Python** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | Ruff, Pylint, Bandit |
| **Java** | ✅ | ✅ | ✅ | ✅ | ✅ | PMD, SpotBugs, JaCoCo |
| **Go** | ✅ | ✅ | ✅ | ⚠️ | ✅ | golangci-lint, Gosec |
| **Rust** | ✅ | ✅ | ✅ | ⚠️ | ✅ | Clippy, cargo-audit |
| **C** | ✅ | ✅ | ⚠️ | ⚠️ | ✅ | Clang-Tidy, Cppcheck |
| **C++** | ✅ | ✅ | ⚠️ | ⚠️ | ✅ | Clang-Tidy, Coverity |
| **C#** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | SonarC#, Roslyn |
| **Ruby** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | RuboCop, Brakeman |
| **PHP** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | PHPCS, PHPStan |
| **Kotlin** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | detekt, Semgrep |
| **Swift** | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | SwiftLint, Semgrep |
| **Scala** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | Scalafix, Scapegoat |
| **Dart** | ✅ | ⚠️ | ✅ | ⚠️ | ⚠️ | dart analyze |
| **Elixir** | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | Credo, Sobelow |
| **Erlang** | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ | Dialyzer, Elvis |
| **Objective-C** | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | OCLint, Infer |
| **VisualBasic** | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | SonarVB, ReSharper |
| **Groovy** | ✅ | 🔧 | ⚠️ | ⚠️ | ⚠️ | CodeNarc |
| **Crystal** | ✅ | ⚠️ | ⚠️ | ❌ | ⚠️ | Ameba |
| **Fortran** | ✅ | ⚠️ | ⚠️ | ⚠️ | ✅ | Fortitude, Coverity |
| **CoffeeScript** | ⚠️ | 🔧 | ❌ | ❌ | ❌ | CoffeeLint |
| **Lua** | ✅ | ⚠️ | ⚠️ | ❌ | ⚠️ | Luacheck |

**Axes applicables:** Security ✅ | Quality ✅ | Tests ✅ | Architecture ✅ | Performance ✅ | Docs ✅

---

### 6.2 🟢 MARKUP LANGUAGES (3 langages)

**Caractéristiques:** Structure de documents, pas de logique exécutable.

| Langage | Validation | Accessibilité | Security | Outils |
|---------|:----------:|:-------------:|:--------:|--------|
| **HTML** | ✅ | ✅ | ⚠️ (XSS context) | HTMLHint, axe-core |
| **XML** | ✅ | ❌ | ⚠️ (XXE) | XMLLint, PMD |
| **Markdown** | ✅ | ❌ | ❌ | markdownlint, remark |

**Axes applicables:**
- ✅ Quality (structure, conventions)
- ✅ Accessibility (HTML)
- ⚠️ Security (XSS dans HTML, XXE dans XML)
- ❌ Tests, Architecture, Performance

**Axes NON applicables:** Tests, Architecture, Performance, Docs

---

### 6.3 🟡 DATA/CONFIG LANGUAGES (4 langages)

**Caractéristiques:** Configuration, données structurées, pas de logique.

| Langage | Schema Valid. | Secrets Detect | Format | Outils |
|---------|:-------------:|:--------------:|:------:|--------|
| **JSON** | ✅ | ✅ | ✅ | JSONLint, Semgrep |
| **YAML** | ✅ | ✅ | ✅ | yamllint, Semgrep |
| **TOML** | ✅ | ✅ | ✅ | taplo |
| **ENV/.env** | ❌ | ✅ | ⚠️ | dotenv-linter, Gitleaks |

**Axes applicables:**
- ✅ Quality (format, schéma)
- ✅ Security (secrets, credentials)
- ⚠️ Schema validation (si schéma défini)
- ❌ Tests, Architecture, Performance

**Focus principal:** Détection de **secrets exposés** !

---

### 6.4 🟣 STYLE LANGUAGES (3 langages)

**Caractéristiques:** Styles visuels, pas de logique métier.

| Langage | Conventions | Perf (size) | Compatibility | Outils |
|---------|:-----------:|:-----------:|:-------------:|--------|
| **CSS** | ✅ | ✅ | ✅ | Stylelint |
| **LESS** | ✅ | ✅ | ✅ | Stylelint |
| **SASS/SCSS** | ✅ | ✅ | ✅ | Stylelint, scss-lint |

**Axes applicables:**
- ✅ Quality (conventions, nesting, sélecteurs)
- ✅ Performance (taille, redondance, spécificité)
- ✅ Compatibility (vendor prefixes, browser support)
- ❌ Security, Tests, Architecture

---

### 6.5 🟠 INFRASTRUCTURE AS CODE (3 langages)

**Caractéristiques:** Configuration d'infrastructure, sécurité critique.

| Langage | Security | Best Practices | Compliance | Outils |
|---------|:--------:|:--------------:|:----------:|--------|
| **Terraform** | ✅ | ✅ | ✅ (CIS, NIST) | tfsec, Checkov, Trivy |
| **Dockerfile** | ✅ | ✅ | ✅ | Hadolint, Trivy, Dockle |
| **Kubernetes** (YAML) | ✅ | ✅ | ✅ | kubesec, Checkov, Trivy |

**Axes applicables:**
- ✅ Security (CRITIQUE - misconfigurations, CVE)
- ✅ Best Practices (image size, least privilege)
- ✅ Compliance (CIS, NIST, PCI-DSS, SOC2)
- ✅ Architecture (infra design)
- ❌ Tests unitaires classiques (mais tests infra via Terratest)

---

### 6.6 🔘 QUERY LANGUAGES (4 langages)

**Caractéristiques:** Requêtes data, injection critique.

| Langage | Injection | Performance | Format | Outils |
|---------|:---------:|:-----------:|:------:|--------|
| **SQL** | ✅ CRITIQUE | ✅ | ✅ | SQLFluff, Semgrep |
| **PLSQL** | ✅ CRITIQUE | ✅ | ✅ | ZPA, SonarQube |
| **TSQL** | ✅ CRITIQUE | ✅ | ✅ | tsqllint, SQLFluff |
| **GraphQL** | ✅ | ✅ | ✅ | graphql-eslint |

**Axes applicables:**
- ✅ Security (injection SQL/NoSQL - CRITIQUE)
- ✅ Performance (N+1, indexes, query optimization)
- ✅ Quality (format, conventions)
- ❌ Tests, Architecture

---

### 6.7 ⚪ TEMPLATING LANGUAGES (3 langages)

**Caractéristiques:** Génération dynamique, XSS critique.

| Langage | XSS Prevention | Conventions | Context | Outils |
|---------|:--------------:|:-----------:|:-------:|--------|
| **JSP** | ✅ CRITIQUE | ✅ | Java EE | PMD, SpotBugs |
| **Velocity** | ✅ | ✅ | Java | PMD |
| **VisualForce** | ✅ CRITIQUE | ✅ | Salesforce | PMD, SF Code Analyzer |

**Axes applicables:**
- ✅ Security (XSS, CSRF - CRITIQUE)
- ✅ Quality (conventions, structure)
- ❌ Tests unitaires directs
- ❌ Architecture, Performance

---

### 6.8 🔷 SPECIALIZED (Blockchain/Smart Contracts)

| Langage | Security | Audit | Gas Optim | Outils |
|---------|:--------:|:-----:|:---------:|--------|
| **Solidity** | ✅ CRITIQUE | ✅ | ✅ | Slither, Mythril, Solhint |
| **Apex** (Salesforce) | ✅ | ✅ | ❌ | PMD, SF Code Analyzer |

**Axes applicables:** Tous (smart contracts = code critique)

---

### 6.9 📋 SCRIPTS/AUTOMATION (3 langages)

| Langage | Security | Quality | Portability | Outils |
|---------|:--------:|:-------:|:-----------:|--------|
| **Shell/Bash** | ✅ | ✅ | ✅ | ShellCheck, Semgrep |
| **Powershell** | ✅ | ✅ | ⚠️ | PSScriptAnalyzer |
| **Makefile** | ⚠️ | ✅ | ⚠️ | Checkmake |

**Axes applicables:**
- ✅ Security (command injection, secrets)
- ✅ Quality (conventions, portability)
- ❌ Tests unitaires (mais tests fonctionnels)
- ❌ Architecture

---

### 6.10 RÉCAPITULATIF AXES PAR TYPE

| Type | Sec | Qual | Tests | Arch | Perf | Docs | Spécial |
|------|:---:|:----:|:-----:|:----:|:----:|:----:|---------|
| 🔵 **Programming** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Complet |
| 🟢 **Markup** | ⚠️ | ✅ | ❌ | ❌ | ❌ | ❌ | Accessibilité |
| 🟡 **Data/Config** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | Secrets |
| 🟣 **Style** | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | Browser compat |
| 🟠 **IaC** | ✅ | ✅ | ⚠️ | ✅ | ❌ | ✅ | Compliance |
| 🔘 **Query** | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | Injection |
| ⚪ **Templating** | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | XSS |
| 🔷 **Blockchain** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Gas/Audit |
| 📋 **Scripts** | ✅ | ✅ | ⚠️ | ❌ | ❌ | ❌ | Portability |

---

### 6.11 DÉTECTION AUTOMATIQUE DU TYPE

L'agent doit **détecter automatiquement** le type pour activer les bons axes:

```yaml
# Mapping extension → type
extensions:
  # 🔵 Programming
  - [.js, .mjs, .cjs]: { type: programming, lang: javascript }
  - [.ts, .tsx]: { type: programming, lang: typescript }
  - [.py, .pyw]: { type: programming, lang: python }
  - [.java]: { type: programming, lang: java }
  - [.go]: { type: programming, lang: go }
  - [.rs]: { type: programming, lang: rust }
  - [.c, .h]: { type: programming, lang: c }
  - [.cpp, .hpp, .cc]: { type: programming, lang: cpp }
  - [.cs]: { type: programming, lang: csharp }
  - [.rb]: { type: programming, lang: ruby }
  - [.php]: { type: programming, lang: php }
  - [.kt, .kts]: { type: programming, lang: kotlin }
  - [.swift]: { type: programming, lang: swift }
  - [.scala]: { type: programming, lang: scala }
  - [.dart]: { type: programming, lang: dart }
  - [.ex, .exs]: { type: programming, lang: elixir }
  - [.erl]: { type: programming, lang: erlang }
  - [.m]: { type: programming, lang: objectivec }
  - [.vb]: { type: programming, lang: visualbasic }
  - [.groovy]: { type: programming, lang: groovy }
  - [.cr]: { type: programming, lang: crystal }
  - [.f90, .f95, .f03]: { type: programming, lang: fortran }
  - [.coffee]: { type: programming, lang: coffeescript }
  - [.lua]: { type: programming, lang: lua }

  # 🟢 Markup
  - [.html, .htm]: { type: markup, lang: html }
  - [.xml, .xsl, .xslt]: { type: markup, lang: xml }
  - [.md, .markdown]: { type: markup, lang: markdown }

  # 🟡 Data/Config
  - [.json]: { type: data, lang: json }
  - [.yaml, .yml]: { type: data, lang: yaml }
  - [.toml]: { type: data, lang: toml }
  - [.env, .env.*]: { type: data, lang: dotenv }

  # 🟣 Style
  - [.css]: { type: style, lang: css }
  - [.less]: { type: style, lang: less }
  - [.scss, .sass]: { type: style, lang: sass }

  # 🟠 Infrastructure
  - [.tf, .tfvars]: { type: iac, lang: terraform }
  - [Dockerfile, Dockerfile.*]: { type: iac, lang: dockerfile }
  - [.hcl]: { type: iac, lang: hcl }

  # 🔘 Query
  - [.sql]: { type: query, lang: sql }
  - [.pls, .plsql]: { type: query, lang: plsql }
  - [.graphql, .gql]: { type: query, lang: graphql }

  # ⚪ Templating
  - [.jsp]: { type: templating, lang: jsp }
  - [.vm, .vtl]: { type: templating, lang: velocity }
  - [.component, .page]: { type: templating, lang: visualforce }

  # 🔷 Blockchain
  - [.sol]: { type: blockchain, lang: solidity }
  - [.cls, .trigger]: { type: specialized, lang: apex }

  # 📋 Scripts
  - [.sh, .bash]: { type: script, lang: shell }
  - [.ps1, .psm1]: { type: script, lang: powershell }
  - [Makefile]: { type: script, lang: makefile }
```

---

### 7. OUTILS UNIVERSELS (Multi-langages)

Ces outils couvrent **plusieurs langages** et sont essentiels pour un agent de review:

| Outil | Langages supportés | Usage | Gratuit |
|-------|-------------------|-------|---------|
| **Semgrep** | 30+ (bash, c, c#, c++, go, java, js, ts, kotlin, php, python, ruby, rust, scala, solidity, swift, terraform, etc.) | Security patterns, SAST | ✅ OSS |
| **Mega-Linter** | 60+ langages, 40+ linters | Meta-linter orchestrator | ✅ OSS |
| **Codacy** | 40+ langages | Quality + Security SaaS | ⚠️ Free tier |
| **SonarQube** | 30+ langages | Quality + Security | ⚠️ Community |
| **Coverity** | 20+ (C, C++, C#, Java, JS, Python, Ruby, Fortran, etc.) | Deep SAST | ❌ Commercial |
| **Checkmarx** | 40+ langages | Enterprise SAST | ❌ Commercial |
| **Trivy** | IaC (Terraform, K8s, Docker) + Images | Vulnérabilités | ✅ OSS |
| **Infer** | C, C++, Objective-C, Java | Memory/null analysis | ✅ OSS (Facebook) |

---

### 8. RECOMMANDATIONS PAR AXE

#### Pour chaque axe, les langages avec couverture COMPLÈTE:

| Axe | Langages ✅ Full Support |
|-----|--------------------------|
| **Security** | JavaScript, TypeScript, Python, Java, Go, C#, Ruby, PHP, Rust, Kotlin, C, C++, Apex, Solidity |
| **Quality** | Tous les 44 langages (via linters dédiés ou Semgrep) |
| **Tests** | JavaScript, TypeScript, Python, Java, Go, C#, Ruby, Kotlin, Scala, Elixir |
| **Architecture** | Java (JDepend, ArchUnit), C#/.NET (NDepend) — Autres langages: analyse IA |
| **Infra** | Terraform, Dockerfile, YAML (K8s), JSON (CloudFormation) |
| **Performance** | Java, C, C++, Go, Rust (profilers + static analysis) |

#### Langages nécessitant analyse IA complémentaire:

Ces langages ont une couverture outillée limitée, l'IA doit compenser:

- **Architecture**: Tous sauf Java/C# (pas d'outils comme ArchUnit)
- **Performance**: Python, Ruby, JavaScript (runtime-only, pas de static analysis)
- **Documentation**: Tous (vérification sémantique par IA)
- **Deployment**: Tous (analyse contextuelle par IA)

---

### 9. SYNTHÈSE COUVERTURE GLOBALE

```
┌────────────────────────────────────────────────────────────────┐
│                    COUVERTURE PAR AXE                          │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  🔴 Security      ████████████████████░░░░  85% (38/44 langs)  │
│  🟡 Quality       ████████████████████████  100% (44/44)        │
│  🧪 Tests         ████████████████░░░░░░░░  65% (28/44)        │
│  🏗️ Architecture  ████░░░░░░░░░░░░░░░░░░░░  15% (7/44)         │
│  🐳 Infra         █████████████░░░░░░░░░░░  50% (via IaC)      │
│  ⚡ Performance   ████████░░░░░░░░░░░░░░░░  30% (13/44)        │
│  📊 Maintain.     ████████████████████░░░░  85% (via metrics)  │
│  📝 Docs          ████████████░░░░░░░░░░░░  45% (20/44)        │
│                                                                 │
│  Légende: █ = Outils disponibles  ░ = IA requis                │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

**Conclusion**: L'IA est **indispensable** pour:
1. **Architecture** (85% des langages sans outils)
2. **Performance** (70% des langages sans analyse statique)
3. **Documentation** (55% des langages sans validation sémantique)
4. **Deployment/Objectives** (100% contextuel)

**Sources:**
- [Semgrep Supported Languages](https://semgrep.dev/docs/supported-languages) - "30+ languages"
- [Analysis Tools Dev](https://analysis-tools.dev/) - "Curated list of static analysis tools"
- [Codacy Languages](https://docs.codacy.com/getting-started/supported-languages-and-tools/) - "40+ languages"

**Confidence:** HIGH

---

## Clarifications

| Question | Réponse |
|----------|---------|
| Plateforme cible ? | Local (indépendant de GitHub/GitLab) |
| Niveau d'autonomie ? | Suggestions + validation humaine, sauf `--approve` |
| Types d'analyses ? | Toutes (10 axes identifiés) |
| Intégration ? | Slash command Claude Code `/review` |
| CodeRabbit vs Codacy ? | Outils similaires, CodeRabbit plus moderne (Biome, oxlint) |
| 44 langages supportés ? | ✅ Oui, via combinaison linters + Semgrep + IA |

---

## Recommendations

1. **Architecture multi-axe** : Implémenter chaque axe comme un subagent indépendant
2. **Flags composables** : Permettre `--security --quality` pour combiner
3. **Profils prédéfinis** : `--quick` (security+quality), `--full` (all)
4. **Configuration projet** : `.review.yaml` pour seuils et objectifs custom
5. **Rapport structuré** : Grouper par axe, puis par severity
6. **Caching intelligent** : Éviter de re-analyser fichiers non modifiés
7. **Semgrep comme base** : Utiliser Semgrep pour 30+ langages avec rules custom
8. **IA pour gaps** : Architecture, Performance, Docs → analyse IA obligatoire

---

## Warnings

- ⚠ **False positives** : Prévoir feedback loop pour marquer faux positifs
- ⚠ **Performance** : L'analyse complète peut prendre >1min sur gros projets
- ⚠ **Dépendances** : Certains outils nécessitent installation (trivy, semgrep)
- ⚠ **Contexte limité** : L'architecture nécessite analyse humaine en complément
- ⚠ **Objectifs projet** : L'axe `--objectives` requiert configuration
- ⚠ **Langages niche** : Fortran, Crystal, CoffeeScript = couverture limitée

---

## Sources Summary

| Source | Domain | Confidence | Sections |
|--------|--------|------------|----------|
| [Semgrep Languages](https://semgrep.dev/docs/supported-languages) | semgrep.dev | HIGH | §6, §7 |
| [Analysis Tools Dev](https://analysis-tools.dev/) | analysis-tools.dev | HIGH | §6 |
| [Codacy Languages](https://docs.codacy.com/getting-started/supported-languages-and-tools/) | codacy.com | HIGH | §6 |
| [PMD Languages](https://pmd.github.io/) | pmd.github.io | HIGH | §6 (Apex, JSP, Velocity) |
| [Salesforce Code Analyzer](https://developer.salesforce.com/docs/platform/salesforce-code-analyzer/guide/code-analyzer.html) | salesforce.com | HIGH | §6 (Apex, VF) |
| [Slither/Mythril](https://github.com/crytic/slither) | github.com | HIGH | §6 (Solidity) |
| [PSScriptAnalyzer](https://github.com/PowerShell/PSScriptAnalyzer) | github.com | HIGH | §6 (PowerShell) |
| [Stylelint](https://stylelint.io/) | stylelint.io | HIGH | §6 (CSS/LESS/SASS) |
| [SQLFluff](https://www.sqlfluff.com/) | sqlfluff.com | HIGH | §6 (SQL variants) |
| [Dart Analysis](https://dart.dev/tools/analysis) | dart.dev | HIGH | §6 (Dart) |
| [Fortitude](https://github.com/lfortran/fortitude) | github.com | MEDIUM | §6 (Fortran) |

---

## 10. OUTILS EXHAUSTIFS PAR TAXONOMIE (Recherche Approfondie 2025)

Cette section détaille **tous les outils disponibles** pour chaque taxonomie, incluant variantes, compléments et frameworks.

---

### 10.1 🔵 PROGRAMMING LANGUAGES - OUTILS DÉTAILLÉS

#### JavaScript / TypeScript

| Outil | Type | Vitesse | Règles | Particularités |
|-------|------|---------|--------|----------------|
| **ESLint** | Linter | Standard | 300+ | Écosystème plugins énorme, standard de l'industrie |
| **Biome** | Linter+Formatter | 20x ESLint | 200+ | Remplace ESLint+Prettier, ex-Rome |
| **oxlint** | Linter | 50-100x ESLint | 400+ | Rust-based, v1.0 stable (2025), compatible ESLint |
| **deno lint** | Linter | Très rapide | 100+ | Intégré à Deno, TypeScript natif |
| **typescript-eslint** | Plugin ESLint | - | 100+ | Type-aware linting pour TypeScript |

**Sources:** [ESLint](https://eslint.org/), [Biome](https://biomejs.dev/), [oxlint](https://oxc.rs/docs/guide/usage/linter.html)

---

#### Python

| Outil | Type | Vitesse | Règles | Particularités |
|-------|------|---------|--------|----------------|
| **Ruff** | Linter+Formatter | 10-100x Flake8 | 800+ | Rust-based, remplace Flake8+Black+isort |
| **Pylint** | Linter | Lent | 409 | Très complet, type inference |
| **Flake8** | Linter | Rapide | 200+ | Modulaire via plugins |
| **Black** | Formatter | Rapide | - | Opinionated, "uncompromising" |
| **mypy** | Type Checker | Moyen | - | Type checking statique officiel |
| **Pyright** | Type Checker | Très rapide | - | Microsoft, utilisé par Pylance (VS Code) |
| **Bandit** | Security | Rapide | 40+ | SAST pour Python, OWASP focused |

**Sources:** [Ruff](https://docs.astral.sh/ruff/), [Pylint](https://pylint.org/), [mypy](https://mypy.readthedocs.io/)

---

#### Go

| Outil | Type | Vitesse | Linters intégrés | Particularités |
|-------|------|---------|------------------|----------------|
| **golangci-lint** | Meta-linter | Rapide | 50+ | Agrège 50+ linters, standard Go |
| **staticcheck** | Linter | Très rapide | 150+ | Focus bugs subtils |
| **revive** | Linter | Rapide | 30+ | Alternative à golint (deprecated) |
| **gosec** | Security | Rapide | 30+ | SAST Go, détection vulnérabilités |
| **go vet** | Linter | Très rapide | 30+ | Intégré au toolchain Go |

**Sources:** [golangci-lint](https://golangci-lint.run/), [staticcheck](https://staticcheck.io/)

---

#### Rust

| Outil | Type | Règles | Particularités |
|-------|------|--------|----------------|
| **Clippy** | Linter | 550+ | Linter officiel Rust, intégré à cargo |
| **rustfmt** | Formatter | - | Formatter officiel Rust |
| **cargo-audit** | Security | CVE DB | Audit dépendances Cargo pour CVEs |
| **miri** | Runtime Analysis | - | Détection undefined behavior |
| **cargo-deny** | Policy | - | Licenses, bans, advisories |

**Sources:** [Clippy](https://rust-lang.github.io/rust-clippy/), [cargo-audit](https://rustsec.org/)

---

#### Java

| Outil | Type | Règles | Particularités |
|-------|------|--------|----------------|
| **PMD** | Linter | 400+ | Détecte bugs, code mort, complexité |
| **SpotBugs** | Bug Finder | 400+ | Successeur de FindBugs, analyse bytecode |
| **Checkstyle** | Style Checker | 100+ | Google/Sun style guides |
| **Error Prone** | Compiler Plugin | 400+ | Détecte bugs à la compilation |
| **SonarJava** | SAST | 600+ | Partie de SonarQube, très complet |
| **find-sec-bugs** | Security | 140+ | Plugin SpotBugs pour sécurité |

**Sources:** [PMD](https://pmd.github.io/), [SpotBugs](https://spotbugs.github.io/), [Error Prone](https://errorprone.info/)

---

#### C# / .NET

| Outil | Type | Règles | Particularités |
|-------|------|--------|----------------|
| **Roslyn Analyzers** | Linter | 200+ | Analyseurs officiels Microsoft |
| **StyleCop.Analyzers** | Style Checker | 100+ | Implémentation Roslyn de StyleCop |
| **SonarAnalyzer.CSharp** | SAST | 400+ | Bugs, vulnérabilités, code smells |
| **Roslynator** | Linter | 500+ | Analyzers + refactorings |
| **ReSharper** | IDE Plugin | 2500+ | Commercial, inspections complètes |
| **Meziantou.Analyzer** | Linter | 100+ | Best practices C# |

**Sources:** [Roslyn Analyzers](https://learn.microsoft.com/en-us/visualstudio/code-quality/roslyn-analyzers-overview), [StyleCop.Analyzers](https://github.com/DotNetAnalyzers/StyleCopAnalyzers)

---

#### Ruby

| Outil | Type | Règles | Particularités |
|-------|------|--------|----------------|
| **RuboCop** | Linter+Formatter | 700+ | Standard Ruby, très configurable |
| **Standard** | Linter | Fixed | RuboCop pré-configuré, "no config" |
| **Brakeman** | Security | 40+ | SAST Rails, réduction 40% bugs sécurité |
| **Reek** | Code Smells | 30+ | Détection couplage, complexité |
| **Fasterer** | Performance | 20+ | Suggestions optimisation |
| **bundle_audit** | Security | CVE DB | Audit gems vulnérables |

**Sources:** [RuboCop](https://rubocop.org/), [Brakeman](https://brakemanscanner.org/), [Reek](https://github.com/troessner/reek)

---

#### PHP

| Outil | Type | Règles | Particularités |
|-------|------|--------|----------------|
| **PHPStan** | Static Analyzer | 9 levels | Type inference avancée, populaire |
| **Psalm** | Static Analyzer | Strict | Focus types, meilleur pour Symfony |
| **PHPCS** | Style Checker | PSR-1/2/12 | PHP_CodeSniffer, auto-fix via phpcbf |
| **PHPMD** | Code Smells | 50+ | PHP Mess Detector, complexité |
| **Phan** | Static Analyzer | - | Par Rasmus Lerdorf, utilisé par MediaWiki |
| **Larastan** | Static Analyzer | - | PHPStan pour Laravel |

**Sources:** [PHPStan](https://phpstan.org/), [Psalm](https://psalm.dev/), [PHP_CodeSniffer](https://github.com/squizlabs/PHP_CodeSniffer)

---

#### C / C++

| Outil | Type | Particularités |
|-------|------|----------------|
| **Clang-Tidy** | Linter | Analyses profondes, auto-fix, checks modernes |
| **Cppcheck** | Bug Finder | Zero false positives goal, détecte bugs subtils |
| **cpplint** | Style Checker | Google C++ Style Guide |
| **Clang Static Analyzer** | SAST | Path-sensitive analysis |
| **include-what-you-use** | Dependency | Optimise les #include |
| **Coverity** | Commercial SAST | Enterprise, très précis |
| **PVS-Studio** | Commercial | C/C++/C#/Java, analyses profondes |

**Sources:** [Clang-Tidy](https://clang.llvm.org/extra/clang-tidy/), [Cppcheck](http://cppcheck.sourceforge.net/)

---

#### Swift

| Outil | Type | Règles | Particularités |
|-------|------|--------|----------------|
| **SwiftLint** | Linter | 100+ | Standard communauté, règles custom possibles |
| **SwiftFormat** | Formatter | 70+ | Auto-format, complète SwiftLint |
| **Periphery** | Dead Code | - | Détecte code inutilisé |
| **swift-format** | Formatter | - | Formatter officiel Apple |

**Usage recommandé:** SwiftLint (build phase) + SwiftFormat (pre-commit hook)

**Sources:** [SwiftLint](https://github.com/realm/SwiftLint), [SwiftFormat](https://github.com/nicklockwood/SwiftFormat)

---

#### Kotlin

| Outil | Type | Règles | Particularités |
|-------|------|--------|----------------|
| **detekt** | Static Analyzer | 100+ | Code smells, bugs, style, très configurable |
| **ktlint** | Style Checker | ~20 | Style officiel Kotlin, peu configurable |
| **Android Lint** | Android Specific | 300+ | Intégré Android Studio |
| **Spotless** | Multi-formatter | - | Format Kotlin + autres langages |

**Recommandation 2025:** ktlint pour style + detekt pour analyse approfondie (complémentaires)

**Sources:** [detekt](https://detekt.dev/), [ktlint](https://pinterest.github.io/ktlint/)

---

#### Scala

| Outil | Type | Particularités |
|-------|------|----------------|
| **Scalafix** | Linter+Refactoring | Règles sémantiques, auto-fix, Scala Center |
| **Scalafmt** | Formatter | Standard Scala, configurable |
| **WartRemover** | Linter | Plugin compilateur, pas d'auto-fix |
| **Scalastyle** | Style Checker | Syntaxique uniquement, moins utilisé |
| **Scapegoat** | Bug Finder | Détection bugs potentiels |

**Sources:** [Scalafix](https://scalacenter.github.io/scalafix/), [Scalafmt](https://scalameta.org/scalafmt/)

---

#### Elixir

| Outil | Type | Particularités |
|-------|------|----------------|
| **Credo** | Linter | Code smells, consistency, teaching focus |
| **Dialyzer** | Type Checker | Via dialyxir, analyse bytecode BEAM |
| **Sobelow** | Security | SAST pour Phoenix Framework |

**Sources:** [Credo](https://github.com/rrrene/credo), [Dialyxir](https://github.com/jeremyjh/dialyxir)

---

#### Dart / Flutter

| Outil | Type | Particularités |
|-------|------|----------------|
| **dart analyze** | Linter officiel | Intégré au SDK Dart |
| **flutter_lints** | Règles Flutter | Google recommended, base Flutter |
| **very_good_analysis** | Règles strictes | Par Very Good Ventures, production-ready |
| **DCM (Dart Code Metrics)** | Métriques | Complexité, architecture, anti-patterns |
| **custom_lint** | Framework | Créer vos propres règles |

**Sources:** [Dart Linter Rules](https://dart.dev/tools/linter-rules), [very_good_analysis](https://github.com/VeryGoodOpenSource/very_good_analysis)

---

#### Lua

| Outil | Type | Particularités |
|-------|------|----------------|
| **Selene** | Linter | Rust-based, actif, rapide, multithreaded |
| **Luacheck** | Linter | Plus ancien (2018), toujours fonctionnel |
| **stylua** | Formatter | Rust-based, format Lua |

**Recommandation 2025:** **Selene** (maintenance active, meilleure UX)

**Sources:** [Selene](https://kampfkarren.github.io/selene/), [Luacheck](https://github.com/mpeterv/luacheck)

---

### 10.2 🟠 INFRASTRUCTURE AS CODE - OUTILS DÉTAILLÉS

#### Terraform

| Outil | Type | Focus | Règles | Sources |
|-------|------|-------|--------|---------|
| **TFLint** | Linter | Correctness | Plugins AWS/Azure/GCP | [tflint](https://github.com/terraform-linters/tflint) |
| **tfsec** | Security | Vulnérabilités | 200+ | Deprecated → Trivy |
| **Trivy** | Multi-scanner | Security + IaC | 1000+ | [Trivy](https://trivy.dev/) |
| **Checkov** | Security + Compliance | CIS, NIST, PCI-DSS | 1000+ | [Checkov](https://www.checkov.io/) |
| **KICS** | Security | Multi-IaC | 1900+ queries | [KICS](https://kics.io/) |
| **Terrascan** | Security | OPA-based | 500+ | [Terrascan](https://runterrascan.io/) |
| **terraform validate** | Syntax | Built-in | - | Intégré |

**Recommandation:** TFLint (correctness) + Checkov ou Trivy (security)

---

#### Kubernetes

| Outil | Type | Particularités |
|-------|------|----------------|
| **kubeconform** | Schema Validation | Successeur de kubeval, très rapide |
| **kubeval** | Schema Validation | Deprecated, utiliser kubeconform |
| **kube-linter** | Best Practices | Par StackRox, vérifie bonnes pratiques |
| **Kubescape** | Security | ARMO, NSA/CISA hardening |
| **Polaris** | Best Practices | Fairwinds, admission controller |
| **Datree** | Policy | Deprecated 2023 |

**Sources:** [kubeconform](https://github.com/yannh/kubeconform), [kube-linter](https://github.com/stackrox/kube-linter)

---

#### Dockerfile

| Outil | Type | Particularités |
|-------|------|----------------|
| **Hadolint** | Linter | Intègre ShellCheck pour RUN, best practices |
| **Trivy** | Security | Scan images + Dockerfile |
| **Dockle** | Security | Container image linter |
| **dockerfile-lint** | Linter | JavaScript-based, alternative Hadolint |

**Recommandation:** **Hadolint** (standard de facto, intégration ShellCheck)

**Sources:** [Hadolint](https://github.com/hadolint/hadolint)

---

#### Ansible

| Outil | Type | Particularités |
|-------|------|----------------|
| **ansible-lint** | Linter | Officiel, best practices playbooks |
| **Steampunk Spotter** | Commercial | Analyse avancée, migrations Ansible |
| **yamllint** | YAML Linter | Pour les fichiers YAML Ansible |

**Sources:** [ansible-lint](https://ansible-lint.readthedocs.io/)

---

### 10.3 🟢 MARKUP & 🟡 DATA/CONFIG LANGUAGES

#### HTML

| Outil | Type | Particularités |
|-------|------|----------------|
| **html-validate** | Validator | Meilleur linter JS rapide (2025) |
| **HTMLHint** | Linter | Simple, configurable |
| **html-eslint** | ESLint Plugin | Nouveau 2025, intégration ESLint |
| **W3C v.Nu** | Validator | Java-based, le plus complet |
| **axe-core** | Accessibility | A11y testing, intégrable |

**Sources:** [html-validate](https://html-validate.org/), [HTMLHint](https://htmlhint.com/)

---

#### Markdown

| Outil | Type | Particularités |
|-------|------|----------------|
| **markdownlint** | Linter | Node.js, 1.2M downloads/week, populaire |
| **remark-lint** | Linter | Unified ecosystem, supporte MDX |
| **pymarkdownlnt** | Linter | Python-based, 46 règles |

**Sources:** [markdownlint](https://github.com/DavidAnson/markdownlint), [remark-lint](https://github.com/remarkjs/remark-lint)

---

#### XML

| Outil | Type | Particularités |
|-------|------|----------------|
| **xmllint** | Validator | libxml2, DTD/RelaxNG/XSD support |
| **Xerces** | Validator | Java, plus complet mais complexe |

---

#### JSON / YAML

| Outil | Type | Particularités |
|-------|------|----------------|
| **Spectral** | Linter | JSON/YAML flexible, OpenAPI/AsyncAPI support |
| **yamllint** | Linter | Python, format + syntaxe YAML |
| **ajv** | JSON Schema | Validation JSON Schema |
| **Taplo** | TOML Linter | TOML 1.0, formatter intégré |
| **Tombi** | TOML LSP | Nouveau 2025, meilleur que Taplo |

**Sources:** [Spectral](https://stoplight.io/open-source/spectral), [yamllint](https://yamllint.readthedocs.io/)

---

### 10.4 🟣 STYLE LANGUAGES

#### CSS / SCSS / LESS

| Outil | Type | Particularités |
|-------|------|----------------|
| **Stylelint** | Linter | Standard CSS/SCSS/Less, 170+ règles |
| **stylelint-scss** | Plugin | Règles SCSS spécifiques |
| **scss-lint** | Legacy | Deprecated, utiliser Stylelint |
| **ESLint CSS** | New 2025 | Support CSS natif dans ESLint |
| **cssnano** | Minifier | Optimisation CSS |

**Sources:** [Stylelint](https://stylelint.io/)

---

### 10.5 🔘 QUERY LANGUAGES (EXHAUSTIF)

`★ Insight ─────────────────────────────────────`
Les "Query Languages" sont des langages de **requête de données**. Ils incluent SQL et ses variantes, mais aussi les langages de requêtes NoSQL et les APIs de recherche.
`─────────────────────────────────────────────────`

#### SQL (Relationnel)

| Outil | Type | Dialectes | Particularités |
|-------|------|-----------|----------------|
| **SQLFluff** | Linter+Formatter | 15+ (PostgreSQL, MySQL, BigQuery, Snowflake, dbt...) | Auto-fix, standard 2025 |
| **SQLint** | Linter | PLSQL, SQL | Intégré Codacy |
| **TSQLLint** | Linter | TSQL (SQL Server) | Microsoft-specific |
| **sqlcheck** | Anti-patterns | ANSI | Détecte anti-patterns |
| **sqlfmt** | Formatter | Multi | Formatting only |

**Sources:** [SQLFluff](https://www.sqlfluff.com/), [TSQLLint](https://github.com/tsqllint/tsqllint)

---

#### GraphQL

| Outil | Type | Particularités |
|-------|------|----------------|
| **graphql-eslint** | ESLint Plugin | The Guild, très complet, utilisé par Microsoft |
| **graphql-schema-linter** | Schema Linter | Standalone, règles schéma |
| **graphql-inspector** | Multi | Schema diff, validation, coverage |

**Sources:** [graphql-eslint](https://the-guild.dev/graphql/eslint/docs)

---

#### NoSQL Query Languages

| Base de données | Langage | Outils disponibles |
|-----------------|---------|-------------------|
| **MongoDB** | MQL (MongoDB Query Language) | MongoDB Compass (visual), Studio 3T |
| **Elasticsearch** | Query DSL (JSON) | Kibana Dev Tools, elasticdump |
| **Neo4j** | Cypher | Hackolade, Cypher LSP |
| **Redis** | Redis Commands | redis-cli, RedisInsight |
| **Cassandra** | CQL | DataStax Studio |

> ⚠️ **Note**: Les query languages NoSQL n'ont généralement **pas de linters dédiés**. La validation se fait via les outils natifs des bases de données ou via des IDE/clients spécialisés.

---

### 10.6 📋 SCRIPTS/AUTOMATION

#### Shell / Bash

| Outil | Type | Particularités |
|-------|------|----------------|
| **ShellCheck** | Static Analyzer | Standard de facto, détecte bugs subtils |
| **shfmt** | Formatter | Go fmt pour shell |
| **bashate** | Style Checker | PEP8-style pour bash |
| **shellharden** | Hardener | Transforme en scripts plus sûrs |

**Recommandation:** **ShellCheck + shfmt** ensemble

**Sources:** [ShellCheck](https://www.shellcheck.net/), [shfmt](https://github.com/mvdan/sh)

---

#### PowerShell

| Outil | Type | Particularités |
|-------|------|----------------|
| **PSScriptAnalyzer** | Linter | Officiel Microsoft, seul vrai linter PS |

**Sources:** [PSScriptAnalyzer](https://github.com/PowerShell/PSScriptAnalyzer)

---

#### Makefile

| Outil | Type | Particularités |
|-------|------|----------------|
| **Checkmake** | Linter | Linter pour Makefiles |

---

### 10.7 🔷 BLOCKCHAIN / SMART CONTRACTS (OPTIONNEL)

> ⚠️ **Note**: Cette section est optionnelle et ne s'applique qu'aux projets Web3/DeFi/Smart Contracts.

#### Solidity (si applicable)

| Outil | Type | Particularités |
|-------|------|----------------|
| **Slither** | Static Analyzer | Trail of Bits, le plus rapide |
| **Solhint** | Linter | Style + Security |
| **Mythril** | Symbolic Execution | Analyse profonde |

---

### 10.8 ⚪ TEMPLATING LANGUAGES (EXHAUSTIF)

`★ Insight ─────────────────────────────────────`
Les templates sont rarement "lintés" directement. La plupart utilisent des **parseurs de syntaxe** ou des **plugins ESLint/IDE**. html-eslint (2025) supporte maintenant EJS, Handlebars, ERB, Twig nativement !
`─────────────────────────────────────────────────`

#### JavaScript Templates (Node.js)

| Moteur | Linter/Validator | Particularités |
|--------|------------------|----------------|
| **EJS** | [ejs-lint](https://github.com/RyanZim/EJS-Lint), eslint-plugin-ejs | Syntax checker pour scriptlets |
| **Handlebars** | html-eslint (intégré) | Support natif `{{variable}}` |
| **Pug** | pug-lint | Linter dédié |
| **Nunjucks** | html-eslint | Inspiré de Jinja2 |
| **Mustache** | html-eslint | Logic-less templates |
| **Liquid** | theme-check (Shopify) | Pour Shopify themes |

#### Java Templates (Spring/Java EE)

| Moteur | Linter/Validator | Particularités |
|--------|------------------|----------------|
| **Thymeleaf** | IntelliJ IDEA (intégré) | Inspections th:* attributes |
| **FreeMarker** | IntelliJ IDEA (intégré) | Plugin FTL |
| **Velocity** | IntelliJ IDEA (intégré), PMD | Plugin VTL |
| **JSP** | PMD, SpotBugs | Via règles Java |

#### Python Templates

| Moteur | Linter/Validator | Particularités |
|--------|------------------|----------------|
| **Jinja2** | [j2lint](https://github.com/aristanetworks/j2lint), ansible-lint, jinjalint | AVD style guide |
| **Django Templates** | djLint | Django-specific |
| **Mako** | - | Pas de linter dédié |

#### PHP Templates

| Moteur | Linter/Validator | Particularités |
|--------|------------------|----------------|
| **Twig** | twig-lint, html-eslint | Symfony ecosystem |
| **Blade** | blade-formatter | Laravel ecosystem |

**Sources:** [ejs-lint](https://github.com/RyanZim/EJS-Lint), [html-eslint](https://html-eslint.org/), [j2lint](https://github.com/aristanetworks/j2lint)

---

### 10.9 🔴 API SPECIFICATIONS (EXHAUSTIF)

`★ Insight ─────────────────────────────────────`
**REST/SOAP ne sont PAS des langages à linter** - ce sont des styles d'architecture API.
- **REST** → Utilise **OpenAPI/Swagger** comme format de spécification
- **SOAP** → Utilise **WSDL** comme format de spécification
- Ce qu'on lint = les **fichiers de spécification** (OpenAPI, WSDL, RAML, etc.)
`─────────────────────────────────────────────────`

#### Formats de Spécification API

| Format | Usage | Part de marché (2024) | Statut |
|--------|-------|----------------------|--------|
| **OpenAPI 3.x** | REST APIs | ~55% | ✅ Standard dominant |
| **OpenAPI 2.0** (Swagger) | REST APIs | ~39% | ✅ Legacy mais supporté |
| **AsyncAPI** | Event-driven APIs | Croissant | ✅ Actif |
| **RAML** | REST APIs | ~7% | ⚠️ Déclin (MuleSoft) |
| **API Blueprint** | REST APIs | ~7% | ❌ Non maintenu (depuis 2019) |
| **WSDL** | SOAP APIs | Legacy | ⚠️ Maintenance mode |
| **GraphQL Schema** | GraphQL | Croissant | ✅ Actif |

#### Outils de Linting API

| Outil | Formats supportés | Particularités |
|-------|-------------------|----------------|
| **Spectral** | OpenAPI 2/3, AsyncAPI, JSON/YAML | Standard de l'industrie, règles custom |
| **Redocly CLI** | OpenAPI 2/3, AsyncAPI, Arazzo | Bundle + docs generation |
| **Vacuum** | OpenAPI 2/3 | Plus rapide que Spectral, compatible |
| **swagger-cli** | OpenAPI/Swagger | Validation basique |
| **oasdiff** | OpenAPI | Détection breaking changes |

**Sources:** [Spectral](https://stoplight.io/open-source/spectral), [Redocly](https://redocly.com/docs/cli/), [Postman State of API](https://www.postman.com/state-of-api/)

---

### 10.10 📡 PROTOCOL BUFFERS / gRPC

| Outil | Type | Particularités |
|-------|------|----------------|
| **Buf CLI** | Linter + Breaking Changes | Standard 2025, v2 avec workspaces |
| **protolint** | Linter | Pluggable, sans compilateur requis |
| **protoc-gen-lint** | Plugin protoc | Style violations |
| **api-linter** | Linter | Google API Design Guidelines |

**Recommandation:** **Buf CLI** (lint + breaking + format en un outil)

**Sources:** [Buf](https://buf.build/), [protolint](https://github.com/yoheimuta/protolint)

---

### 10.11 🏢 ENTERPRISE LANGUAGES

#### Salesforce (Apex, VisualForce, LWC, Aura)

| Outil | Langages | Particularités |
|-------|----------|----------------|
| **Salesforce Code Analyzer v5** | Apex, VF, LWC, Flows | PMD 7.18, ESLint 9.39, Graph Engine |
| **sfdx-scanner** | Apex, JS | Plugin SFDX, MegaLinter intégré |
| **PMD (Apex rules)** | Apex | Règles spécifiques Apex |

**Note:** SFCA v4 retiré août 2025, utiliser v5 obligatoire.

**Sources:** [Salesforce Code Analyzer](https://developer.salesforce.com/docs/platform/salesforce-code-analyzer/guide/code-analyzer.html)

#### SAP (ABAP)

| Outil | Type | Particularités |
|-------|------|----------------|
| **abaplint** | Linter OSS | TypeScript, Clean ABAP Style Guide |
| **SAP Code Inspector (SCI)** | Built-in | Intégré à SE80/ADT |
| **abapOpenChecks** | Extension SCI | Checks custom |
| **SonarQube (ABAP)** | Commercial | Enterprise SAST |

**Sources:** [abaplint](https://abaplint.org/), [abaplint GitHub](https://github.com/abaplint/abaplint)

#### Mainframe (COBOL)

| Outil | Type | Particularités |
|-------|------|----------------|
| **SonarQube** | Commercial | COBOL + JCL support |
| **COBOL Check** | Unit Testing | Open Mainframe Project |
| **Micro Focus Analyzer** | Commercial | Enterprise suite |
| **GnuCOBOL** | Compiler | Free, Area A enforcement |

**Sources:** [COBOL Check](https://openmainframeproject.org/projects/cobol-check/), [SonarQube COBOL](https://www.sonarsource.com/cobol/)

#### Perl

| Outil | Type | Particularités |
|-------|------|----------------|
| **Perl::Critic** | Linter | Basé sur Perl Best Practices |
| **Perl::Lint** | Linter | 3-4x plus rapide que Perl::Critic |
| **perltidy** | Formatter | Code formatting |

**Sources:** [Perl::Critic](https://metacpan.org/pod/Perl::Critic)

---

## 11. MATRICE SYNTHÈSE - CHOIX D'OUTILS PAR LANGAGE

```
┌────────────────────────────────────────────────────────────────────────────┐
│                    RECOMMANDATIONS OUTILS 2025                              │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  LANGAGE          │ QUALITÉ           │ SÉCURITÉ        │ FORMATTER        │
│  ─────────────────┼───────────────────┼─────────────────┼──────────────────│
│  JavaScript/TS    │ Biome OU oxlint   │ Semgrep         │ Biome/Prettier   │
│  Python           │ Ruff              │ Bandit, Semgrep │ Ruff             │
│  Go               │ golangci-lint     │ gosec           │ gofmt            │
│  Rust             │ Clippy            │ cargo-audit     │ rustfmt          │
│  Java             │ PMD + SpotBugs    │ SpotBugs+find-sec│ google-java-fmt │
│  C#               │ Roslyn + StyleCop │ SonarAnalyzer   │ dotnet format    │
│  Ruby             │ RuboCop           │ Brakeman        │ RuboCop          │
│  PHP              │ PHPStan + PHPCS   │ Psalm           │ php-cs-fixer     │
│  Kotlin           │ detekt + ktlint   │ Semgrep         │ ktlint           │
│  Swift            │ SwiftLint         │ Semgrep         │ SwiftFormat      │
│  Terraform        │ TFLint            │ Trivy/Checkov   │ terraform fmt    │
│  Docker           │ Hadolint          │ Trivy           │ -                │
│  SQL              │ SQLFluff          │ SQLFluff        │ SQLFluff         │
│  Shell            │ ShellCheck        │ ShellCheck      │ shfmt            │
│  Solidity         │ Solhint           │ Slither+Mythril │ forge fmt        │
│                                                                             │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 12. OUTILS UNIVERSELS (One Tool, Many Languages)

| Outil | Langages | Type | Installation |
|-------|----------|------|--------------|
| **Semgrep** | 30+ | SAST | `pip install semgrep` |
| **Mega-Linter** | 60+ | Meta-Linter | Docker image |
| **Trivy** | IaC + Containers | Security | Binary / Docker |
| **ast-grep** | 20+ | Pattern matching | `cargo install ast-grep` |
| **Gitleaks** | All | Secret detection | Binary / Docker |
| **TruffleHog** | All | Secret detection | `pip install trufflehog` |

**Recommandation architecture agent:**
1. **Semgrep** comme base pour 30+ langages (custom rules)
2. **Linters spécialisés** pour chaque langage (quality)
3. **Trivy** pour containers/IaC (security)
4. **Gitleaks** pour secrets (all files)

---

---

## 13. RÉFÉRENCE CODACY - OUTILS ACTIVÉS/DÉSACTIVÉS

Basé sur la configuration Codacy réelle fournie (extraction décembre 2025):

### ✅ Outils ACTIVÉS par défaut

| Catégorie | Outil | Langages |
|-----------|-------|----------|
| **Go** | Aligncheck, Deadcode, Gosec, Revive, Staticcheck | Go |
| **Shell/Docker** | ShellCheck, Hadolint | Shell, Dockerfile |
| **IaC** | Checkov | Terraform, JSON, YAML |
| **JavaScript/TS** | ESLint v8 | JavaScript, TypeScript |
| **SQL** | SQLFluff (New), SQLint, TSQLLint | SQL, PLSQL, TSQL |
| **Security Multi** | Semgrep, Trivy | Multi-langages |
| **JSON** | Jackson Linter | JSON |
| **Kotlin/Java** | PMD7 | Kotlin, Java, Apex |
| **CSS** | Stylelint | CSS, LESS, SASS |

### ⛔ Outils DÉSACTIVÉS (à activer selon besoin)

| Catégorie | Outil | Langages | Raison |
|-----------|-------|----------|--------|
| **Ruby** | Brakeman, RuboCop, Reek | Ruby | Projet-specific |
| **Python** | Bandit, Prospector, Pylint, Ruff | Python | Projet-specific |
| **Kotlin** | detekt | Kotlin | Redondant avec PMD7 |
| **Java** | Checkstyle, PMD (legacy) | Java | Remplacé par PMD7 |
| **C/C++** | Clang-Tidy, Cppcheck, Flawfinder | C, C++ | Projet-specific |
| **Scala** | Codacy ScalaMeta, Scalastyle, SpotBugs | Scala | Projet-specific |
| **Groovy** | CodeNarc | Groovy | Projet-specific |
| **CoffeeScript** | CoffeeLint | CoffeeScript | Legacy |
| **Swift** | SwiftLint | Swift | Projet-specific |
| **PowerShell** | PSScriptAnalyzer | PowerShell | Projet-specific |
| **C#** | SonarC#, Unity Roslyn | C# | Projet-specific |
| **API** | Spectral | JSON, YAML | Projet-specific |
| **Dart** | dartanalyzer | Dart | Projet-specific |

### ⚠️ Outils DÉPRÉCIÉS

| Outil | Remplacement |
|-------|--------------|
| ESLint (legacy) | ESLint v8 |
| ESLint9 | ESLint v8 (stable) |
| bundler-audit (Ruby) | Brakeman |
| CSSLint | Stylelint |
| Faux Pas (Objective-C) | Retiré |
| tailor (Swift) | SwiftLint |

### 💡 Recommandation d'activation par projet

```yaml
# Pour un projet Python/Go/TypeScript typique:
enable:
  - ESLint v8        # JavaScript/TypeScript
  - Ruff             # Python (remplace Pylint+Bandit+Black)
  - golangci-lint    # Go (si non dispo, activer Gosec+Revive)
  - Semgrep          # Security multi-langages
  - Trivy            # CVE + IaC
  - ShellCheck       # Scripts
  - Hadolint         # Dockerfiles
  - Checkov          # Terraform

# Pour Codacy spécifiquement (basé sur vos données):
# Les outils déjà activés couvrent Go, Shell, Docker, IaC, JS/TS, SQL, Security
# Activez les langages spécifiques à votre projet (Python, Ruby, C#, etc.)
```

---

---

## 14. 🧠 AGENT REASONING PROTOCOL (Couche Comportementale)

`★ Insight ─────────────────────────────────────`
**Gemini Review**: "Le contexte liste les outils, mais n'explique pas à l'IA comment PENSER comme ces outils."
Cette section transforme l'encyclopédie en **agent qui raisonne**.
`─────────────────────────────────────────────────`

### 14.1 Protocole de Raisonnement (Chain of Thought)

```yaml
reasoning_loop:
  1_identification:
    - Détecter le type de langage (via §6 Taxonomie)
    - Détecter le contexte (Full File vs Diff vs PR)
    - Identifier les axes pertinents (via §6.10 Récapitulatif)

  2_analysis_strategy:
    programming:
      order: [Security, Logic, Performance, Style]
      rationale: "Fix security first, then bugs, then perf, then style"

    iac:
      order: [Security, Compliance, Idempotency, Style]
      rationale: "Misconfigs = breach. Check secrets/policies first"

    markup:
      order: [Validation, Accessibility, Style]
      rationale: "Structure before aesthetics"

    config:
      order: [Secrets, Schema, Format]
      rationale: "Exposed secrets = instant compromise"

  3_context_awareness:
    diff_mode:
      - "Critiquer UNIQUEMENT les lignes modifiées ET leur impact direct"
      - "NE PAS critiquer le legacy code sauf faille critique"
      - "Si effet de bord suspecté → demander fichier complet"

    full_file_mode:
      - "Analyse complète tous axes pertinents"
      - "Grouper commentaires par fonction/section"

  4_filtering:
    priority_rules:
      - "Si bug crash présent → ignorer style issues"
      - "Si faille sécurité → flag immédiat, reste secondaire"
      - "Grouper les issues similaires (ne pas répéter 10x 'missing semicolon')"

    noise_reduction:
      - "Max 5 issues mineures par fichier"
      - "Regrouper duplications: 'X occurrences de Y'"
```

---

### 14.2 Simulation des Outils (Mode LLM-Only)

Quand l'agent n'a **pas accès aux linters en runtime**, il doit **simuler leur rigueur**.

```yaml
tool_simulation:
  python:
    act_as: "Ruff + Bandit + mypy"
    apply_rules:
      - PEP8 (style)
      - B101-B999 (Bandit security)
      - Type hints verification

  javascript:
    act_as: "oxlint + ESLint strict + Semgrep"
    apply_rules:
      - no-eval, no-implied-eval
      - Prototype pollution patterns
      - XSS detection patterns

  go:
    act_as: "golangci-lint (50+ linters)"
    apply_rules:
      - errcheck (unhandled errors)
      - gosec (security)
      - ineffassign, deadcode

  terraform:
    act_as: "Checkov + TFLint"
    apply_rules:
      - CIS Benchmarks
      - No hardcoded secrets
      - Least privilege IAM

  docker:
    act_as: "Hadolint + Trivy"
    apply_rules:
      - DL3000-DL3999 (Hadolint)
      - No root user
      - Pinned versions
```

---

## 15. 🎭 PERSONA & TONE OF VOICE

### 15.1 Persona: "Senior Engineer Mentor"

```yaml
persona:
  identity: "Senior Staff Engineer avec 15+ ans d'expérience"
  mindset:
    - Empathique mais rigoureux
    - Éducatif, pas punitif
    - Valorise l'effort avant de critiquer

  communication_style:
    DO:
      - "A-t-on envisagé X pour résoudre ce problème ?"
      - "Une alternative serait..."
      - "Excellent choix d'utiliser Y ici 👍"
      - "Ce pattern peut causer Z, considérez..."

    DONT:
      - "Fais ça." (ordres directs)
      - "C'est faux." (jugement brutal)
      - "Toujours/Jamais" (absolu)
      - Jargon sans explication

  feedback_structure:
    1_acknowledge: "Commencer par ce qui est bien fait"
    2_explain: "Expliquer le POURQUOI, pas juste le QUOI"
    3_suggest: "Proposer une amélioration concrète"
    4_educate: "Lien vers doc si pertinent"
```

### 15.2 Exemples de Feedback

```markdown
# ❌ Mauvais feedback (robot froid)
"Ligne 42: Variable inutilisée. Supprimer."

# ✅ Bon feedback (mentor)
"La variable `tempData` (L42) semble ne plus être utilisée après le refactoring.
Si c'est intentionnel, on peut la supprimer pour clarifier le code.
Si elle sera utilisée plus tard, un commentaire `// TODO: will be used for X` aiderait."
```

---

## 16. 📊 MATRICE DE SÉVÉRITÉ & PRIORISATION

### 16.1 Niveaux de Sévérité

| Niveau | Emoji | Définition | Action requise |
|--------|-------|------------|----------------|
| **CRITICAL** | 🚨 | Faille sécurité, secret exposé, crash production, data loss | **Blocker** - Merge interdit |
| **MAJOR** | ⚠️ | Bug potentiel, perf O(n²), code non testé, tech debt grave | **Warning** - À traiter avant merge |
| **MINOR** | 💡 | Style, typo, convention, optimisation légère | **Info** - Nice to have |
| **POSITIVE** | ✅ | Bonne pratique observée, code élégant | **Commendation** - Renforce l'adoption |

### 16.2 Critères de Classification

```yaml
severity_criteria:
  CRITICAL:
    security:
      - SQL/NoSQL/Command injection
      - XSS, CSRF, SSRF
      - Hardcoded secrets (API keys, passwords)
      - Authentication bypass
      - Path traversal
    stability:
      - Null pointer / undefined access garantis
      - Infinite loops
      - Memory leaks critiques
      - Data corruption

  MAJOR:
    quality:
      - Cyclomatic complexity > 15
      - Function > 100 lines
      - No tests on critical path
      - Race conditions potentielles
    performance:
      - O(n²) ou pire sur data sets larges
      - N+1 queries
      - Blocking I/O in async context

  MINOR:
    style:
      - Naming conventions
      - Missing JSDoc/docstrings
      - Import order
      - Trailing whitespace
    suggestions:
      - "Could use destructuring"
      - "Consider extract method"
```

---

## 17. 📝 FORMAT DE SORTIE STANDARDISÉ

### 17.1 Structure de Réponse

```markdown
# Code Review: <filename ou scope>

## Summary
<1-2 phrases résumant l'état général du code>

---

## 🚨 Critical Issues (Blockers)
> Ces issues DOIVENT être résolues avant merge.

### [CRITICAL] `filename:line` - <Titre court>
**Problème:** <Description claire du problème>
**Impact:** <Pourquoi c'est critique>
**Suggestion:**
\`\`\`<lang>
// Code corrigé proposé
\`\`\`
**Référence:** [<Doc/OWASP/CWE>](<url>)

---

## ⚠️ Major Issues (Warnings)
> Fortement recommandé de traiter avant merge.

### [MAJOR] `filename:line` - <Titre>
**Problème:** <Description>
**Suggestion:** <Solution proposée>

---

## 💡 Minor Issues (Suggestions)
> Nice to have, peut être traité plus tard.

- `filename:line`: <Issue courte>
- `filename:line`: <Issue courte>

---

## ✅ Commendations
> Ce qui est bien fait dans ce code.

- <Bonne pratique observée>
- <Pattern élégant utilisé>

---

## 📊 Metrics
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Issues Critical | X | 0 | 🔴/🟢 |
| Issues Major | X | ≤3 | 🔴/🟢 |
| Test Coverage | X% | ≥80% | 🔴/🟢 |

---

_Review générée par `/review` - [Docs](link)_
```

### 17.2 Mode JSON (CI/CD Integration)

```json
{
  "review": {
    "summary": "...",
    "timestamp": "ISO8601",
    "files_analyzed": 3,
    "issues": {
      "critical": [
        {
          "file": "src/auth.py",
          "line": 42,
          "rule": "B105",
          "title": "Hardcoded password",
          "description": "...",
          "suggestion": "Use environment variable",
          "reference": "https://owasp.org/..."
        }
      ],
      "major": [...],
      "minor": [...]
    },
    "commendations": [...],
    "metrics": {
      "critical_count": 1,
      "major_count": 2,
      "minor_count": 5,
      "pass": false
    }
  }
}
```

---

## 18. 🔄 GESTION DIFF VS FULL FILE

### 18.1 Détection du Contexte

```yaml
context_detection:
  diff_indicators:
    - Présence de `+` / `-` markers
    - Headers `@@ -X,Y +X,Y @@`
    - Input contient "PR #" ou "pull request"
    - Scope limité (< 100 lignes)

  full_file_indicators:
    - Extension de fichier complète
    - Pas de markers diff
    - Request type: "review this file"
```

### 18.2 Stratégie par Contexte

```yaml
diff_strategy:
  scope: "Lignes modifiées + contexte immédiat (5 lignes avant/après)"

  rules:
    - "NE PAS critiquer le code legacy non modifié"
    - "SAUF si: faille sécurité critique OU impact direct par le changement"

  side_effect_detection:
    triggers:
      - Modification d'une signature de fonction
      - Changement de type de retour
      - Modification d'une constante globale
      - Changement de dépendance

    action: "Demander le fichier complet ou les fichiers dépendants"

  message_template: |
    ⚠️ Ce changement modifie `{function_name}` qui est utilisée ailleurs.
    Puis-je voir les fichiers qui l'appellent pour vérifier la compatibilité ?

full_file_strategy:
  scope: "Fichier entier"

  grouping:
    - Par fonction/classe
    - Par type d'issue (Security, Quality, Style)

  limits:
    - "Max 10 issues majeures par fichier"
    - "Regrouper issues répétitives"
```

---

## 19. 🎯 DECISION TREE (Arbre de Décision)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    AGENT DECISION FLOWCHART                              │
└─────────────────────────────────────────────────────────────────────────┘

START: Recevoir code/diff
         │
         ▼
┌─────────────────────┐
│ 1. Détecter langage │ ──→ Via extension (§6.11)
│    et taxonomie     │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐     ┌─────────────────────┐
│ 2. Contexte?        │──→  │ DIFF: Focus lignes  │
│    Diff ou Full?    │     │ modifiées seulement │
└─────────────────────┘     └─────────────────────┘
         │                            │
         ▼                            ▼
┌─────────────────────┐     ┌─────────────────────┐
│ 3. Axes pertinents? │     │ Side-effects?       │
│    (§6.10 Matrix)   │     │ Oui → Ask full file │
└─────────────────────┘     └─────────────────────┘
         │
         ▼
┌─────────────────────┐
│ 4. Analyser par     │
│    priorité:        │
│    Security > Logic │
│    > Perf > Style   │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│ 5. Classifier       │
│    CRITICAL/MAJOR/  │
│    MINOR/POSITIVE   │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│ 6. Filtrer bruit    │
│    - Max 5 minor    │
│    - Grouper dupes  │
│    - Skip style si  │
│      bugs présents  │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│ 7. Formater output  │
│    (§17 Format)     │
│    + Commendations  │
└─────────────────────┘
         │
         ▼
        END
```

---

## 20. 📋 CHECKLIST PRÉ-REVIEW (Self-Validation)

Avant de générer le feedback final, l'agent vérifie :

```yaml
pre_output_checklist:
  content:
    - [ ] Au moins 1 commendation (positive feedback)
    - [ ] Pas plus de 5 issues mineures par fichier
    - [ ] Issues groupées (pas de répétition)
    - [ ] Chaque CRITICAL a une suggestion de fix

  tone:
    - [ ] Pas d'ordres directs ("Fais X")
    - [ ] Pas de jugements ("C'est mauvais")
    - [ ] Questions ou suggestions ("A-t-on envisagé...")
    - [ ] Explication du POURQUOI, pas juste QUOI

  accuracy:
    - [ ] Issues correspondent au bon type de langage
    - [ ] Pas de false positives évidents
    - [ ] Références aux bonnes règles/outils

  context:
    - [ ] Si diff: pas de critique du legacy
    - [ ] Si full: analyse tous axes pertinents
    - [ ] Side-effects signalés si détectés
```

---

## 21. 🗃️ CACHING & ANALYSE INCRÉMENTALE

`★ Insight ─────────────────────────────────────`
**Performance critique** : Sans cache, un linter peut prendre 50s. Avec cache : 14s (golangci-lint).
Ruff analyse CPython (250k LOC) en 0.4s contre 2.5min pour Pylint.
Le secret : **analyser uniquement les fichiers modifiés**.
`─────────────────────────────────────────────────`

### 21.1 Stratégies de Cache par Outil

| Outil | Cache Location | Stratégie | Gain observé |
|-------|----------------|-----------|--------------|
| **ESLint** | `.eslintcache` | metadata (default) ou content | 5-10x (11s → 1s) |
| **Ruff** | `.ruff_cache` | Hash fichiers, analyse incrémentale | 10-100x |
| **golangci-lint** | `~/.cache/golangci-lint` | Cache entre builds | 3.5x (50s → 14s) |
| **Biome** | `.biome` | Analyse incrémentale | 20x vs ESLint |
| **MegaLinter** | Par linter | Configurable par outil | Variable |

### 21.2 Stratégies de Cache CI/CD

```yaml
# ESLint - Deux stratégies
cache_strategies:
  metadata:
    description: "Compare file size + modification time"
    usage: "Local development (default)"
    flag: "--cache"

  content:
    description: "Compare file content hash"
    usage: "CI/CD (git ne préserve pas mtime)"
    flag: "--cache --cache-strategy content"
    critical: true  # ⚠️ Obligatoire en CI !

# GitHub Actions - Exemple ESLint avec cache
github_actions_eslint:
  - name: Cache ESLint
    uses: actions/cache@v4
    with:
      path: .eslintcache
      key: eslint-${{ hashFiles('**/package-lock.json') }}-${{ github.sha }}
      restore-keys: |
        eslint-${{ hashFiles('**/package-lock.json') }}-

  - name: Lint
    run: npx eslint . --cache --cache-strategy content
```

### 21.3 Analyse Différentielle (Niveau Entreprise)

```yaml
differential_analysis:
  description: |
    Technique avancée où le serveur maintient un "context système"
    et n'analyse que les fichiers changés en utilisant le contexte global.

  outils_supportant:
    - name: "Klocwork (Perforce)"
      description: "Analyse 10M+ lignes en secondes via differential"
      enterprise: true

    - name: "Codee 2025.4"
      description: "Détecte changements, track dépendances, réutilise cache"
      enterprise: false

    - name: "SonarQube"
      description: "Incremental data loads, focus sur nouveau code"
      enterprise: true

  implementation_agent:
    strategy: |
      1. Détecter fichiers modifiés (git diff --name-only)
      2. Identifier fichiers dépendants (imports, includes)
      3. Linter uniquement: modifiés + dépendants directs
      4. Réutiliser résultats cachés pour le reste
```

### 21.4 Configuration Cache pour l'Agent

```yaml
# .review.yaml - Section caching
caching:
  enabled: true
  strategy: "content"  # metadata | content | none

  locations:
    eslint: ".eslintcache"
    ruff: ".ruff_cache"
    golangci: "~/.cache/golangci-lint"
    custom: ".review-cache/"

  invalidation:
    triggers:
      - "package.json"
      - "pyproject.toml"
      - "go.mod"
      - ".eslintrc*"
      - "ruff.toml"
    max_age_days: 7

  incremental:
    enabled: true
    include_dependents: true
    dependency_depth: 1
```

**Sources:** [ESLint Performance](https://www.emmanuelgautier.com/blog/optimize-eslint-performance), [golangci-lint-action](https://github.com/golangci/golangci-lint-action), [Ruff](https://github.com/astral-sh/ruff)

---

## 22. 🔄 GESTION DES FAUX POSITIFS (Feedback Loop)

`★ Insight ─────────────────────────────────────`
**Statistique choc** : Les outils SAST ont un taux de faux positifs de **68-78%** en moyenne.
Impact : **23 minutes perdues** par investigation de faux positif.
Solution : Feedback loop qui transforme le triage manuel en amélioration continue.
`─────────────────────────────────────────────────`

### 22.1 Impact des Faux Positifs

```yaml
false_positive_impact:
  statistics:
    average_fp_rate: "68-78%"
    worst_case: "95% sur code non configuré"
    time_wasted_per_fp: "23 minutes"

  success_metric:
    target: "< 20% false positive rate"
    impact: "3.2x amélioration satisfaction développeur"
```

### 22.2 Mécanismes de Résolution

| Type | Description | Exemple |
|------|-------------|---------|
| **False Positive** | Issue incorrectement flaggée | Variable 'secret' pour jeu de cartes |
| **Won't Fix** | Issue valide mais acceptée | Dette technique pour MVP |
| **Acknowledged** | Issue valide, à corriger plus tard | Créer ticket JIRA |

### 22.3 Feedback Loop Implementation

```yaml
feedback_loop:
  collection:
    inline_comments:
      - "@review ignore: false positive - reason"
      - "@review wontfix: accepted tech debt"
      - "@review ack: will fix in JIRA-1234"

  learning:
    semgrep_memories:
      description: "Semgrep transforme le triage en règles d'exclusion"

    coderabbit_learning:
      description: "@coderabbitai ignore → Mémorisé pour futures PRs"

  refinement:
    automatic:
      - "Patterns FP récurrents → Règle d'exclusion"
      - "Règles avec > 30% FP → Flag pour review"

  metrics:
    track:
      - "FP rate par linter/règle"
      - "Temps moyen de triage"
      - "Satisfaction développeur"
```

### 22.4 Syntaxe de Suppression Unifiée

```yaml
suppression_syntax:
  eslint: "// eslint-disable-next-line rule-name"
  ruff: "# noqa: E501"
  golangci: "//nolint:errcheck"
  semgrep: "# nosemgrep: rule-id"
  generic: "// @review-ignore: reason"
```

**Sources:** [Semgrep Zero FP](https://semgrep.dev/blog/2025/making-zero-false-positive-sast-a-reality-with-ai-powered-memory/), [CodeRabbit Commands](https://docs.coderabbit.ai/guides/commands)

---

## 23. 📄 CONFIGURATION `.review.yaml` (Schéma Complet)

`★ Insight ─────────────────────────────────────`
Les meilleures plateformes (CodeRabbit, Codacy, MegaLinter) utilisent des fichiers YAML avec:
- **Schéma JSON** pour autocomplétion IDE
- **Sections modulaires** (reviews, tools, thresholds)
- **Overrides par path** (règles différentes pour tests vs src)
`─────────────────────────────────────────────────`

### 23.1 Schéma Complet `.review.yaml`

```yaml
# yaml-language-server: $schema=https://example.com/review-schema.json
version: "1.0"
language: "fr"

# ═══════════════════════════════════════════════════════════════════
# REVIEW SETTINGS
# ═══════════════════════════════════════════════════════════════════
reviews:
  profile: "balanced"  # chill | balanced | assertive | strict
  auto_approve:
    enabled: false
    max_minor_issues: 5
    require_tests: true
  scope:
    mode: "diff"  # diff | full | changed_files
    include_dependents: true
  persona: "senior_mentor"

# ═══════════════════════════════════════════════════════════════════
# AXES D'ANALYSE
# ═══════════════════════════════════════════════════════════════════
axes:
  security:    { enabled: true,  priority: 1 }
  quality:     { enabled: true,  priority: 2 }
  tests:       { enabled: true,  priority: 3 }
  architecture: { enabled: true, priority: 4 }
  performance: { enabled: false, priority: 5 }
  documentation: { enabled: false, priority: 6 }

# ═══════════════════════════════════════════════════════════════════
# THRESHOLDS & QUALITY GATES
# ═══════════════════════════════════════════════════════════════════
thresholds:
  complexity:
    cyclomatic_max: 15
    cognitive_max: 20
    function_lines_max: 100
    nesting_depth_max: 4

  coverage:
    min_line_coverage: 80
    min_branch_coverage: 75
    min_new_code_coverage: 90

  issues:
    max_critical: 0   # Bloquant
    max_major: 3      # Warning
    max_minor: 10     # Info

# ═══════════════════════════════════════════════════════════════════
# OBJECTIFS PROJET (Contextuel)
# ═══════════════════════════════════════════════════════════════════
objectives:
  performance:
    latency_p99_ms: 100
    throughput_rps: 1000
  reliability:
    uptime_target: "99.9%"
  tech_debt:
    max_todos: 5
    max_deprecated_apis: 0

# ═══════════════════════════════════════════════════════════════════
# TOOLS CONFIGURATION
# ═══════════════════════════════════════════════════════════════════
tools:
  javascript: { linter: "biome", formatter: "biome" }
  python: { linter: "ruff", type_checker: "pyright", security: "bandit" }
  go: { linter: "golangci-lint" }
  terraform: { linter: "tflint", security: "checkov" }

  universal:
    secrets: "gitleaks"
    semgrep:
      enabled: true
      rulesets: ["p/default", "p/owasp-top-ten"]

# ═══════════════════════════════════════════════════════════════════
# PATH FILTERS & OVERRIDES
# ═══════════════════════════════════════════════════════════════════
paths:
  ignore:
    - "vendor/**"
    - "node_modules/**"
    - "*.min.js"
    - "*.generated.*"

  overrides:
    - pattern: "**/*_test.go"
      settings:
        complexity: { cyclomatic_max: 20 }
        axes: { documentation: { enabled: false } }

    - pattern: "**/migrations/**"
      settings:
        axes: { quality: { enabled: false } }

# ═══════════════════════════════════════════════════════════════════
# CACHING
# ═══════════════════════════════════════════════════════════════════
caching:
  enabled: true
  strategy: "content"
  directory: ".review-cache"
  max_age_days: 7

# ═══════════════════════════════════════════════════════════════════
# OUTPUT FORMAT
# ═══════════════════════════════════════════════════════════════════
output:
  format: "markdown"  # markdown | json | sarif | github
  include_commendations: true
  group_by: "severity"
  max_issues_per_file: 10

# ═══════════════════════════════════════════════════════════════════
# FEEDBACK LOOP
# ═══════════════════════════════════════════════════════════════════
feedback:
  enabled: true
  learn_from_dismissals: true
  alert_on_high_fp_rate: 30
```

### 23.2 Comparaison avec Outils Existants

| Feature | CodeRabbit | Codacy | MegaLinter | `.review.yaml` |
|---------|------------|--------|------------|----------------|
| Schéma JSON | ✅ | ✅ | ✅ | ✅ |
| Path overrides | ✅ | ✅ | ✅ | ✅ |
| Objectifs projet | ❌ | ❌ | ❌ | ✅ **Unique** |
| Feedback loop config | Partiel | ❌ | ❌ | ✅ **Unique** |
| Persona | ✅ | ❌ | ❌ | ✅ |

### 23.3 Profils Pré-configurés

```yaml
profiles:
  chill:      # Dev rapide
    axes: { security: true, quality: false, tests: false }
    thresholds: { issues: { max_major: 10, max_minor: 999 } }

  strict:     # Pre-release
    axes: { security: true, quality: true, tests: true, performance: true }
    thresholds: { issues: { max_critical: 0, max_major: 0 }, coverage: { min_line_coverage: 90 } }
```

**Sources:** [CodeRabbit Config](https://docs.coderabbit.ai/reference/configuration), [Codacy Config](https://docs.codacy.com/repositories-configure/codacy-configuration-file/), [MegaLinter Config](https://megalinter.io/latest/configuration/)

---

## Sources Section 21-23

| Source | Domain | Confidence |
|--------|--------|------------|
| [ESLint Caching](https://www.emmanuelgautier.com/blog/optimize-eslint-performance) | emmanuelgautier.com | HIGH |
| [ESLint Multithread](https://eslint.org/blog/2025/08/multithread-linting/) | eslint.org | HIGH |
| [golangci-lint-action](https://github.com/golangci/golangci-lint-action) | github.com | HIGH |
| [Ruff Performance](https://github.com/astral-sh/ruff) | github.com | HIGH |
| [Semgrep AI Memory](https://semgrep.dev/blog/2025/making-zero-false-positive-sast-a-reality-with-ai-powered-memory/) | semgrep.dev | HIGH |
| [CodeRabbit Docs](https://docs.coderabbit.ai/reference/configuration) | coderabbit.ai | HIGH |
| [Codacy Config](https://docs.codacy.com/repositories-configure/codacy-configuration-file/) | codacy.com | HIGH |
| [MegaLinter Config](https://megalinter.io/latest/configuration/) | megalinter.io | HIGH |

---

## Sources Section 14-20

| Source | Domain | Confidence |
|--------|--------|------------|
| Gemini 3 Review | External AI Review | HIGH |
| [Anthropic Agent Patterns](https://www.anthropic.com/research/building-effective-agents) | anthropic.com | HIGH |
| [OWASP Code Review Guide](https://owasp.org/www-project-code-review-guide/) | owasp.org | HIGH |
| [Google Code Review Guidelines](https://google.github.io/eng-practices/review/) | google.github.io | HIGH |

---

## Sources Section 10-13

| Source | Domain | Confidence |
|--------|--------|------------|
| [ESLint](https://eslint.org/) | eslint.org | HIGH |
| [Biome](https://biomejs.dev/) | biomejs.dev | HIGH |
| [oxlint](https://oxc.rs/) | oxc.rs | HIGH |
| [Ruff](https://docs.astral.sh/ruff/) | astral.sh | HIGH |
| [golangci-lint](https://golangci-lint.run/) | golangci-lint.run | HIGH |
| [Clippy](https://rust-lang.github.io/rust-clippy/) | rust-lang.github.io | HIGH |
| [PHPStan](https://phpstan.org/) | phpstan.org | HIGH |
| [detekt](https://detekt.dev/) | detekt.dev | HIGH |
| [SwiftLint](https://github.com/realm/SwiftLint) | github.com | HIGH |
| [Checkov](https://www.checkov.io/) | checkov.io | HIGH |
| [Trivy](https://trivy.dev/) | trivy.dev | HIGH |
| [Hadolint](https://github.com/hadolint/hadolint) | github.com | HIGH |
| [SQLFluff](https://www.sqlfluff.com/) | sqlfluff.com | HIGH |
| [ShellCheck](https://www.shellcheck.net/) | shellcheck.net | HIGH |
| [Slither](https://github.com/crytic/slither) | github.com | HIGH |
| [Spectral](https://stoplight.io/open-source/spectral) | stoplight.io | HIGH |
| [Stylelint](https://stylelint.io/) | stylelint.io | HIGH |
| [markdownlint](https://github.com/DavidAnson/markdownlint) | github.com | HIGH |
| [Analysis Tools Dev](https://analysis-tools.dev/) | analysis-tools.dev | HIGH |

---

## 24. 🐝 ARCHITECTURE "THE HIVE" (La Ruche)

`★ Insight ─────────────────────────────────────`
**Pattern Organisationnel:** Architecture inspirée d'une ruche d'abeilles où un "Brain" (Orchestrateur) coordonne des "Drones" (Sub-agents) spécialisés par Taxonomie. Chaque drone est expert dans son domaine et ne se réveille que si nécessaire.
`─────────────────────────────────────────────────`

### 24.1 Vue d'ensemble

```
┌─────────────────────────────────────────────────────────────────┐
│                        THE HIVE                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│    ┌─────────────┐      ┌──────────────────────────────────┐   │
│    │   PR/Push   │─────→│         BRAIN (Orchestrateur)    │   │
│    │  (Trigger)  │      │  • Routing par Taxonomie          │   │
│    └─────────────┘      │  • Cache Check (SHA-256)          │   │
│                         │  • Priorisation & Filtering       │   │
│                         │  • Interface PR (seul writer)     │   │
│                         └──────────────┬───────────────────┘   │
│                                        │                        │
│         ┌──────────────────────────────┼──────────────────────┐ │
│         │                              │                      │ │
│         ▼                              ▼                      ▼ │
│   ┌───────────┐   ┌───────────┐   ┌───────────┐   ┌───────────┐│
│   │  DRONE    │   │  DRONE    │   │  DRONE    │   │  DRONE    ││
│   │  Python   │   │  JS/TS    │   │  Go       │   │  IaC      ││
│   │  Agent    │   │  Agent    │   │  Agent    │   │  Agent    ││
│   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘   └─────┬─────┘│
│         │               │               │               │      │
│         └───────────────┴───────────────┴───────────────┘      │
│                                │                                │
│                         ┌──────▼──────┐                        │
│                         │    CACHE    │                        │
│                         │  (Redis/DB) │                        │
│                         └─────────────┘                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 24.2 Flux de Données (Workflow)

```yaml
workflow_phases:
  1_ingestion:
    trigger: "PR/Push event"
    action: "git diff --name-only"
    output: "Liste des fichiers modifiés"

  2_dispatch_cache_check:
    for_each_file:
      - compute_hash: "SHA-256 du contenu"
      - query_cache: "Redis/DB lookup"
      - decision:
          cache_hit: "Récupérer JSON d'analyse stocké"
          cache_miss: "Dispatcher au Sub-Agent de la Taxonomie"

  3_parallel_analysis:
    mode: "Async (tous les agents en parallèle)"
    isolation: "Chaque agent indépendant"
    timeout: "30s par fichier"

  4_aggregation_filtering:
    actions:
      - merge_all_jsons: "Consolidation des résultats"
      - apply_priority: "CRITICAL > MAJOR > MINOR"
      - filter_noise: "Masquer mineurs si critiques présents"

  5_actuation:
    target: "API GitHub/GitLab"
    modes:
      - "Request Changes (blocking)"
      - "Comment on lines"
      - "Summary comment"
```

### 24.3 Le Cerveau : Orchestrateur (Brain)

L'orchestrateur **ne lit pas le code en détail**. Il gère la **logistique et la politique**.

```yaml
orchestrator_responsibilities:
  routing:
    description: "Assigner chaque fichier au bon Sub-Agent"
    rules:
      "*.py": "Agent Python"
      "*.js|*.ts|*.tsx": "Agent JS/TS"
      "*.go": "Agent Go"
      "*.tf|*.yml (k8s)": "Agent IaC"
      "*.css|*.scss": "Agent Style"
      "Dockerfile": "Agent Docker"

  prioritization:
    rule: "N'affiche les Warnings que si 0 Critiques"
    rationale: "Focus développeur sur l'essentiel"

  pr_interface:
    exclusive: true
    reason: "Éviter que 5 agents spamment la PR simultanément"
    actions: ["post_review", "comment_on_line", "request_changes"]

  synthesis:
    format: "Markdown unique et digeste"
    grouping: "Par sévérité, puis par fichier"
```

**Prompt Système de l'Orchestrateur :**

```
Tu es le Lead Reviewer. Tu ne vérifies pas le code toi-même.
Tu reçois des rapports JSON de tes spécialistes.

Ta tâche est de synthétiser :
1. Groupe les retours par sévérité.
2. Si un rapport contient une 'CRITICAL security flaw',
   bloque tout le reste et alerte immédiatement.
3. Formate le tout en un commentaire Markdown unique
   et digeste pour l'humain.
```

### 24.4 Stratégie de Caching "Smart Delta"

```yaml
cache_architecture:
  storage: "Key-Value Store (Redis/DB)"

  schema:
    key: "sha256_of_file_content"
    value:
      analysis_json: "{ issues: [], score: 100 }"
      metadata:
        timestamp: "ISO8601"
        agent_version: "1.2.0"
        rules_hash: "sha256 des règles utilisées"

  example:
    | Clé (Hash)              | Valeur (JSON Analysis)                              | Metadata     |
    |-------------------------|-----------------------------------------------------|--------------|
    | sha256_of_user_py_v1    | { "issues": [], "score": 100 }                      | timestamp: T-1 |
    | sha256_of_user_py_v2    | { "issues": [{"line": 10, "severity": "CRITICAL"}]} | timestamp: Now |

delta_logic:
  description: "Optimisation pour fichiers modifiés"

  strategies:
    full_rescan:
      when: "Changements structurels majeurs"
      action: "Analyser la nouvelle version complète"
      cost: "Plus cher en compute"
      safety: "Plus sûr"

    contextual_rescan:
      when: "Petites modifications"
      action: |
        L'orchestrateur envoie au Sub-Agent:
        - Le diff uniquement
        - "L'ancienne analyse disait que tout était OK"
        L'agent vérifie juste si le diff introduit une régression.
      cost: "Économique"
      safety: "Suffisant pour modifications mineures"

  decision_matrix:
    | Taille du diff | Nombre de lignes modifiées | Stratégie            |
    |----------------|----------------------------|----------------------|
    | < 10 lignes    | < 5                        | contextual_rescan    |
    | 10-100 lignes  | 5-50                       | contextual_rescan    |
    | > 100 lignes   | > 50                       | full_rescan          |
    | Nouveau fichier| N/A                        | full_rescan          |
    | Fichier supprimé| N/A                       | cache_invalidate     |
```

### 24.5 Protocole de Communication (JSON Schema)

**Standard Output pour CHAQUE Sub-Agent :**

```json
{
  "$schema": "hive-agent-output-v1",
  "taxonomy": "python_agent",
  "file": "src/api/auth.py",
  "file_hash": "a1b2c3d4e5f6...",
  "analysis_timestamp": "2024-01-15T10:30:00Z",
  "status": "FAILED",
  "score": 45,
  "findings": [
    {
      "id": "SEC-001",
      "severity": "CRITICAL",
      "category": "SECURITY",
      "subcategory": "SQL_INJECTION",
      "line_start": 42,
      "line_end": 45,
      "column_start": 12,
      "column_end": 48,
      "message": "SQL Injection detected via formatted string",
      "explanation": "User input is directly concatenated into SQL query",
      "suggestion": "Use parameterized queries instead of string formatting",
      "fix_snippet": "cursor.execute(sql, (user_id,))",
      "references": [
        "https://owasp.org/www-community/attacks/SQL_Injection",
        "https://docs.python.org/3/library/sqlite3.html#sqlite3-placeholders"
      ],
      "confidence": 0.95,
      "false_positive_likelihood": "LOW"
    },
    {
      "id": "STYLE-001",
      "severity": "MINOR",
      "category": "STYLE",
      "subcategory": "LINE_LENGTH",
      "line_start": 10,
      "line_end": 10,
      "message": "Line exceeds 120 characters (found: 145)",
      "suggestion": "Break line into multiple lines",
      "confidence": 1.0,
      "false_positive_likelihood": "NONE"
    }
  ],
  "metrics": {
    "lines_analyzed": 150,
    "complexity_score": 12,
    "test_coverage_estimated": 65
  }
}
```

**JSON Schema Validation :**

```json
{
  "type": "object",
  "required": ["taxonomy", "file", "file_hash", "status", "findings"],
  "properties": {
    "taxonomy": { "type": "string", "enum": ["python_agent", "js_agent", "go_agent", "iac_agent", "style_agent", "docker_agent", "sql_agent"] },
    "status": { "type": "string", "enum": ["PASSED", "FAILED", "ERROR", "SKIPPED"] },
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["severity", "category", "line_start", "message"],
        "properties": {
          "severity": { "type": "string", "enum": ["CRITICAL", "MAJOR", "MINOR", "INFO"] },
          "category": { "type": "string", "enum": ["SECURITY", "QUALITY", "PERFORMANCE", "STYLE", "DOCUMENTATION", "TEST"] }
        }
      }
    }
  }
}
```

### 24.6 Implémentation : PR Reviewer Mode

**Scénario concret :**

```
Utilisateur push une PR avec :
  - 1 faille SQL critique (auth.py:42)
  - 3 problèmes de style mineurs
```

**Flux d'exécution :**

```yaml
step_1_drone_analysis:
  agent: "Python Agent (Drone)"
  input: "auth.py"
  output:
    status: "FAILED"
    findings:
      - { severity: "CRITICAL", category: "SECURITY", line: 42, message: "SQL Injection" }
      - { severity: "MINOR", category: "STYLE", line: 10, message: "Line too long" }
      - { severity: "MINOR", category: "STYLE", line: 25, message: "Unused import" }
      - { severity: "MINOR", category: "STYLE", line: 78, message: "Missing docstring" }

step_2_orchestrator_filter:
  input: "JSON du drone Python"
  logic:
    - count_critical: 1
    - count_minor: 3
    - rule_applied: "Si Critical > 0, masquer Minor"
  output:
    visible_findings: [{ severity: "CRITICAL", line: 42 }]
    hidden_findings: 3
    message: "3 problèmes mineurs masqués. Corrigez d'abord la sécurité."

step_3_github_action:
  api_calls:
    - action: "Create Review"
      event: "REQUEST_CHANGES"
      blocking: true

    - action: "Comment on line"
      file: "auth.py"
      line: 42
      body: |
        🚨 **CRITICAL: SQL Injection**

        User input is directly concatenated into SQL query.

        **Current code:**
        ```python
        cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")
        ```

        **Suggested fix:**
        ```python
        cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
        ```

    - action: "Global comment"
      body: |
        ## 🐝 Hive Review Summary

        | Severity | Count | Status |
        |----------|-------|--------|
        | 🔴 Critical | 1 | **BLOCKING** |
        | 🟡 Minor | 3 | Hidden |

        ⚠️ **3 problèmes de style mineurs détectés.**
        Corrigez d'abord la faille de sécurité pour que je les affiche.
```

### 24.7 Avantages de l'Architecture Hive

```yaml
performance:
  principle: "Activation sélective des drones"
  example: "Si vous ne touchez qu'au CSS → Agent Python, Agent SQL, Agent IaC restent endormis"
  benefits:
    - "0 coût compute pour agents non concernés"
    - "0 latence additionnelle"
    - "Parallélisme natif"

clarity:
  principle: "Orchestrateur = Rédacteur en Chef"
  role: "Nettoie le bruit avant publication"
  benefits:
    - "Focus développeur sur l'essentiel"
    - "Pas de spam de 50 commentaires mineurs"
    - "Message unique et actionable"

traceability:
  principle: "Push direct sur la PR"
  format: "Commentaires sur les lignes de code"
  benefits:
    - "IA native au workflow Git"
    - "Historique conservé dans la PR"
    - "Review inline comme un humain"

cost_optimization:
  principle: "Cache SHA-256 + Smart Delta"
  benefits:
    - "Fichiers inchangés = 0 re-analyse"
    - "Petits diffs = analyse contextuelle légère"
    - "ROI mesurable par fichier"
```

### 24.8 Mapping Drones ↔ Taxonomies

```yaml
drone_registry:
  python_agent:
    taxonomies: ["Python"]
    file_patterns: ["*.py", "*.pyi", "*.pyw"]
    tools_simulated: ["Ruff", "Bandit", "mypy", "Pylint"]

  javascript_agent:
    taxonomies: ["JavaScript", "TypeScript"]
    file_patterns: ["*.js", "*.jsx", "*.ts", "*.tsx", "*.mjs", "*.cjs"]
    tools_simulated: ["ESLint", "Biome", "oxlint", "Semgrep"]

  go_agent:
    taxonomies: ["Go"]
    file_patterns: ["*.go"]
    tools_simulated: ["golangci-lint (50+ linters)", "gosec"]

  rust_agent:
    taxonomies: ["Rust"]
    file_patterns: ["*.rs"]
    tools_simulated: ["Clippy", "cargo-audit"]

  java_agent:
    taxonomies: ["Java", "Kotlin", "Scala"]
    file_patterns: ["*.java", "*.kt", "*.kts", "*.scala"]
    tools_simulated: ["SpotBugs", "PMD", "Checkstyle", "detekt"]

  csharp_agent:
    taxonomies: ["C#", "VB.NET"]
    file_patterns: ["*.cs", "*.vb"]
    tools_simulated: ["SonarC#", "Roslynator"]

  php_agent:
    taxonomies: ["PHP"]
    file_patterns: ["*.php"]
    tools_simulated: ["PHPStan", "Psalm", "PHP_CodeSniffer"]

  ruby_agent:
    taxonomies: ["Ruby"]
    file_patterns: ["*.rb", "*.rake", "Gemfile"]
    tools_simulated: ["RuboCop", "Brakeman"]

  iac_agent:
    taxonomies: ["Terraform", "Kubernetes", "Docker", "Ansible"]
    file_patterns: ["*.tf", "*.yml", "*.yaml", "Dockerfile", "docker-compose*.yml"]
    tools_simulated: ["Checkov", "TFLint", "Hadolint", "kube-linter"]

  style_agent:
    taxonomies: ["CSS", "SCSS", "LESS"]
    file_patterns: ["*.css", "*.scss", "*.sass", "*.less"]
    tools_simulated: ["Stylelint"]

  sql_agent:
    taxonomies: ["SQL", "GraphQL"]
    file_patterns: ["*.sql", "*.graphql", "*.gql"]
    tools_simulated: ["SQLFluff", "graphql-schema-linter"]

  shell_agent:
    taxonomies: ["Shell", "PowerShell"]
    file_patterns: ["*.sh", "*.bash", "*.zsh", "*.ps1"]
    tools_simulated: ["ShellCheck", "PSScriptAnalyzer"]

  markup_agent:
    taxonomies: ["Markdown", "HTML", "XML"]
    file_patterns: ["*.md", "*.html", "*.htm", "*.xml"]
    tools_simulated: ["markdownlint", "HTMLHint"]

  config_agent:
    taxonomies: ["JSON", "YAML", "TOML"]
    file_patterns: ["*.json", "*.yml", "*.yaml", "*.toml"]
    tools_simulated: ["Schema validation", "YAML lint"]
```

---

## Sources Section 24

| Source | Domain | Confidence |
|--------|--------|------------|
| [Anthropic Orchestrator-Workers](https://www.anthropic.com/research/building-effective-agents) | anthropic.com | HIGH |
| [GitHub API Reviews](https://docs.github.com/en/rest/pulls/reviews) | github.com | HIGH |
| [Redis Caching Patterns](https://redis.io/docs/manual/patterns/) | redis.io | HIGH |
| Architecture proposée par utilisateur | Internal Design | HIGH |

---

_Ce fichier est généré automatiquement par `/search`. Ne pas commiter._
