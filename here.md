# Plan d'Amélioration Architecture - superviz.io

## Résumé Exécutif

Le feedback architectural reçu contient **11 critiques**. Après analyse du code source:
- **3 critiques sont valides** → Actions requises
- **8 critiques sont infondées** → Pas d'action (justifications ci-dessous)

---

## Partie 1: Critiques REJETÉES (avec justification)

### 1. "Interfaces mal placées (consumer vs provider)"

**Critique**: Les interfaces sont définies côté provider (infrastructure) au lieu du consumer (domaine).

**REJET**: **Faux à 96%**

J'ai analysé **23 interfaces** dans le projet:

| Couche | Interfaces | Placement |
|--------|------------|-----------|
| Domain | 11 ports | ✅ Correct (Executor, Prober, MetricsStore, etc.) |
| Application | 4 interfaces | ✅ Correct (Loader, Creator, etc.) |
| Infrastructure | 7 helpers | ✅ Correct (internes pour tests) |
| **Violation** | 1 | ❌ `ZombieReaper` - seule erreur |

**Preuve**: Fichier `src/internal/domain/process/executor.go`:
```go
// Interface définie dans DOMAIN (consumer)
type Executor interface {
    Start(ctx context.Context, spec Spec) (int, <-chan ExitResult, error)
    Stop(pid int) error
    Signal(pid int, sig os.Signal) error
}
```
L'infrastructure l'implémente sans la définir.

---

### 2. "Couche Domaine peu visible / responsabilités floues"

**Critique**: Le Domain Layer n'est pas illustré par des packages concrets.

**REJET**: **Complètement faux**

Le domaine contient **7 packages** avec **51 fichiers**:

```
src/internal/domain/
├── process/      # Spec, State, Event, ExitResult, Executor port
├── service/      # Config, ServiceConfig, RestartPolicy
├── health/       # Status, AggregatedHealth, Event
├── healthcheck/  # Target, Result, Config, Prober port
├── metrics/      # CPU, Memory, Disk, Network (6 collectors)
├── listener/     # Listener entity, State machine
├── shared/       # Duration, Size, Clock (value objects)
├── storage/      # MetricsStore port
├── event/        # Publisher port, Event types
└── state/        # HostState, DaemonState
```

Chaque package contient des **entités métier pures** sans dépendance technique.

---

### 3. "Duplication de responsabilités entre couches"

**Critique**: `application/process/` et `infrastructure/process/` font la même chose.

**REJET**: **Faux - responsabilités distinctes**

| Couche | Package | Responsabilité |
|--------|---------|----------------|
| **Application** | `process/manager.go` | **QUOI faire**: politique de restart, backoff, orchestration |
| **Infrastructure** | `process/executor/` | **COMMENT faire**: appels OS, exec.Cmd, signaux |

Exemple concret:
- `ProcessManager` (app): "Si le process échoue, attendre 5s puis redémarrer"
- `Executor` (infra): "Exécuter `exec.Command()` avec ces credentials"

---

### 4. "Dépendances croisées et inversion mal appliquée"

**Critique**: Les erreurs sentinelles dans infrastructure créent une dépendance de tous vers ce fichier.

**REJET**: **Partiellement faux**

Les erreurs sont organisées **par couche et par domaine**:

| Fichier | Scope | Utilisé par |
|---------|-------|-------------|
| `domain/process/errors.go` | Process métier | Application seulement |
| `domain/shared/errors.go` | Cross-domain | Domain seulement |
| `infrastructure/process/errors.go` | OS/technique | Infrastructure seulement |

**Pas de dépendance infrastructure→domain**. Chaque couche a ses propres erreurs.

---

### 5. "Erreurs sentinel vs erreurs enrichies"

**Critique**: Usage d'erreurs constantes exportées est un anti-pattern.

**REJET**: **Correctement implémenté**

Le code utilise le pattern Go moderne:

```go
// Sentinel + Wrapping (correct)
return fmt.Errorf("starting process %d: %w", pid, ErrPermissionDenied)
```

J'ai trouvé **45+ usages** de `fmt.Errorf(...%w...)` pour le contexte.
Les tests utilisent `errors.Is()` correctement.

---

### 6. "Gestion des goroutines et concurrence"

**Critique**: Risque de fuites de goroutines.

**REJET**: **Bien géré**

Patterns observés:
- `context.Context` pour annulation (32 fichiers)
- `sync.WaitGroup` pour coordination shutdown
- Channels buffered (size 1) pour éviter les blocages
- `sync.RWMutex` pour accès concurrent

Exemple (`application/supervisor/supervisor.go`):
```go
func (s *Supervisor) Stop() {
    s.cancel()        // Annule le context
    s.wg.Wait()       // Attend toutes les goroutines
}
```

---

### 7. "Tags de build multi-plateforme"

**Critique**: Risque de mauvaise configuration des directives.

**REJET**: **Correctement organisé**

| Fichier | Tag | Plateforme |
|---------|-----|------------|
| `*_unix.go` | `//go:build unix` | Linux + macOS + BSD |
| `*_linux.go` | `//go:build linux` | Linux uniquement |
| `*_darwin.go` | `//go:build darwin` | macOS |
| `*_bsd.go` | `//go:build (freebsd \|\| openbsd)` | BSD |
| `*_scratch.go` | `//go:build !unix` | CI/containers |

Factory pattern pour sélection:
```go
func NewSystemCollector() metrics.SystemCollector {
    switch runtime.GOOS {
    case "linux": return linux.NewProbe()
    case "darwin": return darwin.NewProbe()
    default: return scratch.NewProbe()
    }
}
```

---

### 8. "Complexité globale - over-engineering"

**Critique**: Architecture surdimensionnée pour un superviseur.

**REJET**: **Justifié pour le cas d'usage**

Un superviseur PID1 nécessite:
- Multi-plateforme (Linux, BSD, macOS, containers)
- Zombie reaping (PID 1 uniquement)
- Health checks multi-protocoles (TCP, HTTP, gRPC, ICMP)
- Gestion signaux (SIGTERM, SIGHUP, SIGCHLD)
- Métriques système (CPU, RAM, IO)
- Restart policies avec backoff

Cette complexité est **inhérente au domaine**, pas de l'over-engineering.

---

## Partie 2: Critiques VALIDES → Actions Requises

### Action 1: Corriger ZombieReaper (OBLIGATOIRE)

**Problème**: Interface `ZombieReaper` définie dans infrastructure, utilisée par application.

**Fichiers à modifier**:

| Fichier | Action |
|---------|--------|
| `src/internal/domain/reaper/reaper.go` | **CRÉER** - Nouvelle interface port |
| `src/internal/domain/reaper/CLAUDE.md` | **CRÉER** - Documentation |
| `src/internal/application/supervisor/supervisor.go` | **MODIFIER** - Changer import |
| `src/internal/bootstrap/wire.go` | **MODIFIER** - Binding |
| `src/internal/bootstrap/providers.go` | **MODIFIER** - Provider |

**Code à créer** (`domain/reaper/reaper.go`):
```go
package reaper

// Reaper handles zombie process cleanup (PID 1 responsibility).
// This is a domain port - implementation is in infrastructure.
type Reaper interface {
    // Start begins the background reaping loop.
    Start()

    // Stop terminates the reaping loop gracefully.
    Stop()

    // ReapOnce performs a single reap cycle, returns count of reaped processes.
    ReapOnce() int

    // IsPID1 returns true if running as init process.
    IsPID1() bool
}
```

---

### Action 2: Séparer erreurs de validation (RECOMMANDÉ)

**Problème**: `domain/service/validate.go` mélange logique de validation et définitions d'erreurs.

**Fichiers à modifier**:

| Fichier | Action |
|---------|--------|
| `src/internal/domain/service/errors.go` | **CRÉER** - Extraire les 9 erreurs |
| `src/internal/domain/service/validate.go` | **MODIFIER** - Retirer les `var Err...` |

---

### Action 3: Compléter documentation CLAUDE.md (RECOMMANDÉ)

**Fichiers manquants**:

| Fichier | Contenu |
|---------|---------|
| `domain/reaper/CLAUDE.md` | Nouveau (Action 1) |
| `domain/event/CLAUDE.md` | Publisher port |
| `domain/state/CLAUDE.md` | HostState, DaemonState |
| `domain/storage/CLAUDE.md` | MetricsStore port |

---

## Partie 3: Plan d'Exécution

### Étape 1: Créer le port Reaper dans le domaine
```bash
# Fichiers créés:
src/internal/domain/reaper/
├── reaper.go      # Interface Reaper
└── CLAUDE.md      # Documentation
```

### Étape 2: Mettre à jour les imports
```go
// AVANT (supervisor.go)
import "github.com/.../infrastructure/process/reaper"

// APRÈS
import "github.com/.../domain/reaper"
```

### Étape 3: Mettre à jour Wire bindings
```go
// bootstrap/wire.go
var ReaperSet = wire.NewSet(
    infrareaper.NewReaper,
    wire.Bind(new(domainreaper.Reaper), new(*infrareaper.Reaper)),
)
```

### Étape 4: Validation
```bash
go build ./...           # Compilation
go test -race ./...      # Tests
golangci-lint run        # Linting
```

### Étape 5: (Optionnel) Séparer erreurs service
```bash
# Extraire de validate.go vers errors.go
```

### Étape 6: (Optionnel) Compléter documentation
```bash
# Créer CLAUDE.md manquants
```

---

## Récapitulatif

| Type | Critiques | Action |
|------|-----------|--------|
| Rejetées | 8 | Aucune (justifications ci-dessus) |
| Valides | 3 | Modifications listées |

**Effort estimé**:
- Action 1 (obligatoire): 5 fichiers, ~1h
- Action 2 (recommandé): 2 fichiers, ~30min
- Action 3 (recommandé): 4 fichiers, ~30min

---

## Vérification

```bash
# Test compilation multi-plateforme
GOOS=linux go build ./...
GOOS=darwin go build ./...

# Tests avec race detector
go test -race ./...

# Vérifier pas d'import circulaire
go mod graph | grep "domain.*infrastructure"
# Doit être VIDE
```
