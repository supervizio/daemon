---
name: lint
description: |
  Intelligent linting with ktn-linter using RLM decomposition.
  Sequences 148 rules optimally across 8 phases.
  Fixes ALL issues automatically in intelligent order.
  Detects DTOs on-the-fly and applies dto:"direction,context,security" convention.
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
---

# /lint - Intelligent Linting (RLM Architecture)

$ARGUMENTS

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Use `grepai_trace_callers`/`grepai_trace_callees` for impact analysis.
Fallback to Grep ONLY for exact string matches or regex patterns.

---

## WORKFLOW AUTOMATIQUE

Ce skill corrige **TOUTES** les issues ktn-linter sans exception.
Pas d'arguments. Pas de flags. Juste execution complete.

---

## EXECUTION IMMEDIATE

### Etape 1 : Lancer ktn-linter

```bash
./builds/ktn-linter lint ./... 2>&1
```

Si le binaire n'existe pas :

```bash
go build -o ./builds/ktn-linter ./cmd/ktn-linter && ./builds/ktn-linter lint ./...
```

### Etape 2 : Parser le retour

Pour chaque ligne d'erreur format `fichier:ligne:colonne: KTN-XXX-YYY: message` :

1. Extraire le fichier
2. Extraire la regle (KTN-XXX-YYY)
3. Extraire le message
4. Classer dans la phase appropriee

### Etape 3 : Classer par phase

**PHASE 1 - STRUCTURAL** (fixer EN PREMIER - affecte les autres phases)

```text
KTN-STRUCT-ONEFILE   → Splitter fichiers multi-structs OU ajouter dto:"..."
KTN-TEST-SUFFIX      → Renommer _test.go → _external_test.go ou _internal_test.go
KTN-TEST-INTPRIV     → Deplacer tests prives vers _internal_test.go
KTN-TEST-EXTPUB      → Deplacer tests publics vers _external_test.go
KTN-TEST-PKGNAME     → Corriger nom de package test
KTN-CONST-ORDER      → Deplacer const en haut du fichier
KTN-VAR-ORDER        → Deplacer var apres const
```

**PHASE 2 - SIGNATURES** (modifier signatures de fonctions)

```text
KTN-FUNC-ERRLAST     → Mettre error en dernier retour
KTN-FUNC-CTXFIRST    → Mettre context.Context en premier parametre
KTN-FUNC-MAXPARAM    → Grouper parametres ou creer struct
KTN-FUNC-NAMERET     → Ajouter noms aux retours si >3
KTN-FUNC-GROUPARG    → Grouper params de meme type
KTN-RECEIVER-MIXPTR  → Uniformiser receiver (pointer ou value)
KTN-RECEIVER-NAME    → Corriger nom receiver (1-2 chars)
```

**PHASE 3 - LOGIC** (corriger erreurs de logique)

```text
KTN-VAR-SHADOW       → Renommer variable qui shadow
KTN-CONST-SHADOW     → Renommer const qui shadow builtin
KTN-FUNC-DEADCODE    → Supprimer fonction non utilisee
KTN-FUNC-CYCLO       → Refactorer fonction trop complexe
KTN-FUNC-MAXSTMT     → Splitter fonction >35 statements
KTN-FUNC-MAXLOC      → Splitter fonction >50 LOC
KTN-VAR-TYPEASSERT   → Ajouter check ok sur type assertion
KTN-ERROR-WRAP       → Utiliser %w dans fmt.Errorf
KTN-ERROR-SENTINEL   → Creer sentinel error package-level
KTN-GENERIC-*        → Corriger contraintes generiques
KTN-ITER-*           → Corriger patterns iterator
KTN-GOVET-*          → Corriger tous les govet
```

**PHASE 4 - PERFORMANCE** (optimisations memoire)

```text
KTN-VAR-HOTLOOP      → Sortir allocation de la boucle
KTN-VAR-BIGSTRUCT    → Passer par pointeur si >64 bytes
KTN-VAR-SLICECAP     → Prealloc slice avec capacite
KTN-VAR-MAPCAP       → Prealloc map avec capacite
KTN-VAR-MAKEAPPEND   → Utiliser make au lieu de append
KTN-VAR-GROW         → Utiliser Buffer.Grow
KTN-VAR-STRBUILDER   → Utiliser strings.Builder
KTN-VAR-STRCONV      → Eviter string() en boucle
KTN-VAR-SYNCPOOL     → Utiliser sync.Pool
KTN-VAR-ARRAY        → Utiliser array si <=64 bytes
```

**PHASE 5 - MODERN** (idiomes Go 1.18-1.25)

```text
KTN-VAR-USEANY       → interface{} → any
KTN-VAR-USECLEAR     → boucle delete → clear()
KTN-VAR-USEMINMAX    → math.Min/Max → min/max
KTN-VAR-RANGEINT     → for i := 0; i < n → for i := range n
KTN-VAR-LOOPVAR      → Supprimer copie variable boucle (Go 1.22+)
KTN-VAR-SLICEGROW    → Utiliser slices.Grow
KTN-VAR-SLICECLONE   → Utiliser slices.Clone
KTN-VAR-MAPCLONE     → Utiliser maps.Clone
KTN-VAR-CMPOR        → Utiliser cmp.Or
KTN-VAR-WGGO         → Utiliser WaitGroup.Go (Go 1.25+)
KTN-FUNC-MINMAX      → math.Min/Max → min/max
KTN-FUNC-USECLEAR    → clear() builtin
KTN-FUNC-RANGEINT    → range over int
MODERNIZE-*          → Tous les modernize
```

**PHASE 6 - STYLE** (conventions de nommage)

```text
KTN-VAR-CAMEL        → snake_case → camelCase
KTN-CONST-CAMEL      → UPPER_CASE → UpperCase
KTN-VAR-MINLEN       → Renommer var trop courte
KTN-VAR-MAXLEN       → Renommer var trop longue
KTN-CONST-MINLEN     → Renommer const trop courte
KTN-CONST-MAXLEN     → Renommer const trop longue
KTN-FUNC-UNUSEDARG   → Prefixer _ si unused
KTN-FUNC-BLANKPARAM  → Supprimer _ si pas interface
KTN-FUNC-NOMAGIC     → Extraire magic number en const
KTN-FUNC-EARLYRET    → Supprimer else apres return
KTN-FUNC-NAKEDRET    → Ajouter return explicite
KTN-STRUCT-NOGET     → GetX() → X()
KTN-INTERFACE-ERNAME → Ajouter suffix -er
```

**PHASE 7 - DOCS** (documentation - EN DERNIER)

```text
KTN-COMMENT-PKGDOC   → Ajouter doc package
KTN-COMMENT-FUNC     → Ajouter doc fonction
KTN-COMMENT-STRUCT   → Ajouter doc struct
KTN-COMMENT-CONST    → Ajouter doc const
KTN-COMMENT-VAR      → Ajouter doc var
KTN-COMMENT-BLOCK    → Ajouter commentaire bloc
KTN-COMMENT-LINELEN  → Wrapper ligne >100 chars
KTN-GOROUTINE-LIFECYCLE → Documenter lifecycle goroutine
```

**PHASE 8 - TESTS** (patterns de test)

```text
KTN-TEST-TABLE       → Convertir en table-driven
KTN-TEST-COVERAGE    → Ajouter tests manquants
KTN-TEST-ASSERT      → Ajouter assertions
KTN-TEST-ERRCASES    → Ajouter cas d'erreur
KTN-TEST-NOSKIP      → Supprimer t.Skip()
KTN-TEST-SETENV      → Corriger t.Setenv en parallel
KTN-TEST-SUBPARALLEL → Ajouter t.Parallel aux subtests
KTN-TEST-CLEANUP     → Utiliser t.Cleanup
```

---

## Convention DTO : dto:"direction,context,security"

**Le tag dto:"..." exempte les structs de KTN-STRUCT-ONEFILE et KTN-STRUCT-CTOR.**

### Format

```go
dto:"<direction>,<context>,<security>"
```

| Position | Valeurs | Description |
|----------|---------|-------------|
| direction | `in`, `out`, `inout` | Sens du flux |
| context | `api`, `cmd`, `query`, `event`, `msg`, `priv` | Type DTO |
| security | `pub`, `priv`, `pii`, `secret` | Classification |

### Valeurs Security

| Valeur | Logging | Marshaling | Usage |
|--------|---------|------------|-------|
| `pub` | Affiche | Inclus | Donnees publiques |
| `priv` | Affiche | Inclus | IDs, timestamps |
| `pii` | Masque | Conditionnel | Email, nom (RGPD) |
| `secret` | REDACTED | Omis | Password, token |

### Exemple Complet

```go
// Fichier: user_dto.go - PLUSIEURS DTOs (grace a dto:"...")

type CreateUserRequest struct {
    Username string `dto:"in,api,pub" json:"username" validate:"required"`
    Email    string `dto:"in,api,pii" json:"email" validate:"email"`
    Password string `dto:"in,api,secret" json:"password" validate:"min=8"`
}

type UserResponse struct {
    ID        string    `dto:"out,api,pub" json:"id"`
    Username  string    `dto:"out,api,pub" json:"username"`
    Email     string    `dto:"out,api,pii" json:"email"`
    CreatedAt time.Time `dto:"out,api,pub" json:"createdAt"`
}

type UpdateUserCommand struct {
    UserID   string `dto:"in,cmd,priv" json:"userId"`
    Email    string `dto:"in,cmd,pii" json:"email,omitempty"`
}
```

### Quand Ajouter dto:"..."

| Situation | Action |
|-----------|--------|
| Struct DTO/Request/Response | Ajouter `dto:"dir,ctx,sec"` |
| Struct sans tags (DTO) | Ajouter `dto:"dir,ctx,sec"` |
| Struct avec json/yaml/xml | OK, detecte DTO |
| KTN-STRUCT-ONEFILE DTO | dto tags → OK |

### Suffixes Reconnus

```text
DTO, Request, Response, Params, Input, Output,
Payload, Message, Event, Command, Query
```

### Guide Choix Valeurs

```text
DIRECTION:
  - Entree utilisateur → in
  - Sortie vers client → out
  - Update/Patch → inout

CONTEXT:
  - API REST/GraphQL → api
  - Commande CQRS → cmd
  - Query CQRS → query
  - Event sourcing → event
  - Message queue → msg
  - Interne → priv

SECURITY:
  - Nom produit, status → pub
  - IDs, timestamps → priv
  - Email, nom, adresse → pii
  - Password, token, cle → secret
```

---

## Regles d'Application DTO

```text
SI KTN-STRUCT-ONEFILE sur un struct :
   1. Lire le fichier
   2. Verifier si le struct devrait etre un DTO (par NOM)
   3. SI oui → Ajouter dto:"dir,ctx,sec" sur chaque champ
   4. Relancer le linter → plus d'erreur ONEFILE

SI KTN-STRUCT-CTOR sur un struct :
   1. Verifier si DTO (par tags ou nom)
   2. SI DTO sans tags → Ajouter dto:"dir,ctx,sec"
   3. Relancer → plus d'erreur CTOR

SI KTN-DTO-TAG (format invalide) :
   → Corriger le format: dto:"direction,context,security"

SI KTN-STRUCT-JSONTAG :
   → Ajouter le tag manquant (json, xml, ou dto selon contexte)

SI KTN-STRUCT-PRIVTAG :
   → Supprimer les tags des champs prives
```

---

## Execution Mode: Agent Teams (Claude 4.6)

**Si `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` est active**, utiliser Agent Teams pour paralleliser les phases independantes.

### Architecture Agent Teams

```text
LEAD (Phase 1-3 : SEQUENTIAL - dependances inter-fichiers)
  Phase 1: STRUCTURAL (7 regles)
    ↓
  Phase 2: SIGNATURES (7 regles)
    ↓
  Phase 3: LOGIC (17 regles)
    ↓
  Re-run ktn-linter → valider convergence Phase 1-3
    ↓
  === SPAWN 4 TEAMMATES ===
    ├── "perf"   → Phase 4 PERFORMANCE (11 regles)
    ├── "modern" → Phase 5 MODERN (20 regles)
    ├── "polish" → Phase 6 STYLE + Phase 7 DOCS (25 regles)
    └── "tester" → Phase 8 TESTS (8 regles)
    ↓
  LEAD: attend completion de tous les teammates
    ↓
  Re-run ktn-linter final → valider convergence globale
```

### Teammate Roles

**Lead** : Orchestre le workflow. Execute Phase 1-3 (structurel + signatures + logique) qui ont des dependances inter-fichiers. Apres convergence Phase 3, spawne les 4 teammates. Collecte les resultats et lance la verification finale.

**Teammate "perf"** (Phase 4) : Specialiste optimisation memoire. Corrige : KTN-VAR-HOTLOOP, KTN-VAR-BIGSTRUCT, KTN-VAR-SLICECAP, KTN-VAR-MAPCAP, KTN-VAR-MAKEAPPEND, KTN-VAR-GROW, KTN-VAR-STRBUILDER, KTN-VAR-STRCONV, KTN-VAR-SYNCPOOL, KTN-VAR-ARRAY.

**Teammate "modern"** (Phase 5) : Specialiste Go idiomatique. Corrige : KTN-VAR-USEANY, KTN-VAR-USECLEAR, KTN-VAR-USEMINMAX, KTN-VAR-RANGEINT, KTN-VAR-LOOPVAR, KTN-VAR-SLICEGROW, KTN-VAR-SLICECLONE, KTN-VAR-MAPCLONE, KTN-VAR-CMPOR, KTN-VAR-WGGO, MODERNIZE-*.

**Teammate "polish"** (Phase 6+7) : Style et documentation. Corrige : KTN-VAR-CAMEL, KTN-CONST-CAMEL, KTN-VAR-MINLEN/MAXLEN, KTN-FUNC-UNUSEDARG, KTN-FUNC-NOMAGIC, KTN-FUNC-EARLYRET, KTN-STRUCT-NOGET, KTN-INTERFACE-ERNAME + tous les KTN-COMMENT-*.

**Teammate "tester"** (Phase 8) : Qualite tests. Corrige : KTN-TEST-TABLE, KTN-TEST-COVERAGE, KTN-TEST-ASSERT, KTN-TEST-ERRCASES, KTN-TEST-NOSKIP, KTN-TEST-SETENV, KTN-TEST-SUBPARALLEL, KTN-TEST-CLEANUP.

### User Interaction (VS Code)

- `Shift+Up/Down` pour naviguer entre teammates
- Ecrire directement a un teammate pour guider ses decisions
- Chaque teammate utilise TaskCreate/TaskUpdate pour reporter sa progression

### Fallback : Mode Sequentiel

**Si Agent Teams non disponible**, executer le mode classique :

```text
POUR chaque phase de 1 a 8 :
    POUR chaque issue de cette phase :
        1. Lire le fichier concerne
        2. SI struct DTO → appliquer convention dto:"dir,ctx,sec"
        3. Appliquer la correction
        4. TaskUpdate → completed
    FIN POUR
FIN POUR

Relancer ktn-linter pour verifier convergence
SI encore des issues : recommencer
SINON : terminer avec rapport
```

---

## Rapport Final

```text
═══════════════════════════════════════════════════════════════
  /lint - COMPLETE
═══════════════════════════════════════════════════════════════

  Mode           : Agent Teams (4 teammates) | Sequential
  Issues corrigees : 47
  Iterations       : 3
  DTOs detectes    : 4 (exclus de ONEFILE/CTOR)

  Par phase :
    STRUCTURAL  : 5 corriges (dont 2 via dto tags)  [Lead]
    SIGNATURES  : 8 corriges                         [Lead]
    LOGIC       : 12 corriges                        [Lead]
    PERFORMANCE : 4 corriges                         [perf]
    MODERN      : 10 corriges                        [modern]
    STYLE       : 5 corriges                         [polish]
    DOCS        : 3 corriges                         [polish]
    TESTS       : 0 corriges                         [tester]

  DTOs traites :
    - user_dto.go: CreateUserRequest, UserResponse (dto:"...,api,...")
    - order_dto.go: OrderCommand, OrderQuery (dto:"...,cmd/query,...")

  Verification finale : 0 issues

═══════════════════════════════════════════════════════════════
```

---

## REGLES ABSOLUES

1. **TOUT corriger** - Aucune exception, aucun skip
2. **Ordre des phases** - Phase 1→3 sequentiel, Phase 4→8 parallele (Agent Teams) ou sequentiel (fallback)
3. **DTOs au vol** - Detecter et appliquer dto:"dir,ctx,sec"
4. **Iteration** - Relancer jusqu'a 0 issues
5. **Pas de questions** - Tout est automatique
6. **Format dto strict** - Toujours 3 valeurs separees par virgule
7. **TaskCreate** - Chaque phase = 1 task avec progression

---

## DEMARRER MAINTENANT

1. Lancer `./builds/ktn-linter lint ./...`
2. Parser le retour
3. Classer par phase
4. Detecter Agent Teams disponible (`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`)
5. SI Agent Teams : Lead Phase 1-3, spawn teammates Phase 4-8
6. SINON : corriger sequentiellement 1→8 (DTOs avec convention dto:"dir,ctx,sec")
7. Relancer jusqu'a convergence
8. Afficher rapport final
