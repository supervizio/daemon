# Init - Project Initialization Check

$ARGUMENTS

---

## Description

Commande d'initialisation automatique du projet. Exécutée au démarrage du devcontainer via `postStart.sh`.

**Objectif :** Vérifier que le projet est personnalisé (CLAUDE.md et README.md différents du template).

---

## Arguments

| Pattern | Action |
|---------|--------|
| (vide) | Vérifie si le projet est initialisé |
| `--force` | Force la réinitialisation |
| `--status` | Affiche le statut d'initialisation |
| `--help` | Affiche l'aide |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /init - Project Initialization Check
═══════════════════════════════════════════════

Usage: /init [options]

Options:
  (vide)      Vérifie et initialise si nécessaire
  --force     Force la réinitialisation
  --status    Affiche le statut actuel
  --help      Affiche cette aide

Comportement:
  - Compare CLAUDE.md et README.md avec le template
  - Si identiques → guide l'utilisateur pour personnaliser
  - Si différents → projet déjà initialisé

Fichiers vérifiés:
  - ./CLAUDE.md (doit être personnalisé)
  - ./README.md (doit être personnalisé)

Template de référence:
  https://github.com/kodflow/devcontainer-template
═══════════════════════════════════════════════
```

---

## Logique d'initialisation

### Étape 1 : Récupérer les footprints du template

```bash
REPO="kodflow/devcontainer-template"
BRANCH="main"

# Footprint = premiers 500 caractères + hash SHA256
get_footprint() {
    local file="$1"
    echo "$(head -c 500 "$file" 2>/dev/null | sha256sum | cut -d' ' -f1)"
}

# Footprints distants
REMOTE_CLAUDE=$(curl -sL "https://raw.githubusercontent.com/$REPO/$BRANCH/CLAUDE.md" | head -c 500 | sha256sum | cut -d' ' -f1)
REMOTE_README=$(curl -sL "https://raw.githubusercontent.com/$REPO/$BRANCH/README.md" | head -c 500 | sha256sum | cut -d' ' -f1)

# Footprints locaux
LOCAL_CLAUDE=$(get_footprint "./CLAUDE.md")
LOCAL_README=$(get_footprint "./README.md")
```

### Étape 2 : Comparer les footprints

| Condition | Action |
|-----------|--------|
| CLAUDE.md identique au template | ⚠️ Demander personnalisation |
| README.md identique au template | ⚠️ Demander personnalisation |
| Les deux personnalisés | ✅ Projet initialisé |

### Étape 3 : Guider la personnalisation

Si un fichier n'est pas personnalisé, afficher les instructions :

```
═══════════════════════════════════════════════
  /init - Initialisation requise
═══════════════════════════════════════════════

  ⚠️ Fichiers non personnalisés détectés :

  ○ CLAUDE.md (identique au template)
  ○ README.md (identique au template)

─────────────────────────────────────────────

  CLAUDE.md doit contenir :
    - Description de VOTRE projet
    - Structure de VOTRE codebase
    - Conventions de VOTRE équipe
    - Règles spécifiques au projet

  README.md doit contenir :
    - Nom de VOTRE projet
    - Description du projet
    - Instructions d'installation
    - Guide d'utilisation

─────────────────────────────────────────────

  Voulez-vous que je vous aide à personnaliser
  ces fichiers maintenant ?

═══════════════════════════════════════════════
```

Puis utiliser `AskUserQuestion` :

```yaml
questions:
  - question: "Voulez-vous personnaliser ces fichiers maintenant ?"
    header: "Init"
    options:
      - label: "Oui, guidez-moi"
        description: "Questions interactives pour personnaliser"
      - label: "Non, plus tard"
        description: "Rappel au prochain démarrage"
    multiSelect: false
```

---

## --status

Afficher le statut d'initialisation :

```
═══════════════════════════════════════════════
  /init --status
═══════════════════════════════════════════════

  Projet : <nom du dossier>
  Statut : ✅ Initialisé / ⚠️ Non initialisé

─────────────────────────────────────────────
  Fichiers
─────────────────────────────────────────────

  CLAUDE.md : ✅ Personnalisé / ⚠️ Template
  README.md : ✅ Personnalisé / ⚠️ Template

─────────────────────────────────────────────
  Footprints
─────────────────────────────────────────────

  Template CLAUDE.md : <hash court>
  Local CLAUDE.md    : <hash court>
  Match              : Oui/Non

  Template README.md : <hash court>
  Local README.md    : <hash court>
  Match              : Oui/Non

═══════════════════════════════════════════════
```

---

## Workflow de personnalisation interactive

Si l'utilisateur choisit "Oui, guidez-moi" :

### Pour CLAUDE.md

```yaml
questions:
  - question: "Quel est le nom de votre projet ?"
    header: "Projet"
    options:
      - label: "Utiliser le nom du dossier"
        description: "<nom_dossier_courant>"
      - label: "Autre"
        description: "Je saisis un nom personnalisé"
    multiSelect: false

  - question: "Quel type de projet est-ce ?"
    header: "Type"
    options:
      - label: "API/Backend"
        description: "REST, GraphQL, gRPC..."
      - label: "Frontend/Web"
        description: "React, Vue, Angular..."
      - label: "CLI/Tool"
        description: "Outil en ligne de commande"
      - label: "Library"
        description: "Package/Module réutilisable"
    multiSelect: false

  - question: "Quelle architecture utilisez-vous ?"
    header: "Archi"
    options:
      - label: "Clean Architecture"
        description: "Couches indépendantes"
      - label: "Hexagonal"
        description: "Ports & Adapters"
      - label: "MVC"
        description: "Model-View-Controller"
      - label: "Flat"
        description: "Structure simple"
    multiSelect: false
```

### Pour README.md

```yaml
questions:
  - question: "Décrivez brièvement le projet (1-2 phrases)"
    header: "Desc"
    options:
      - label: "Je saisis"
        description: "Saisie libre"
    multiSelect: false
```

### Génération des fichiers

Après les réponses, générer des fichiers personnalisés :

**CLAUDE.md template personnalisé :**

```markdown
# <Nom du projet>

## Project Structure

```text
/workspace
├── src/                # Source code
│   └── ...
├── tests/              # Unit tests
└── docs/               # Documentation
```

## Type

`<Type>` project using `<Architecture>` pattern.

## Conventions

- All code in `/src`
- Tests in `/tests`
- Follow language-specific rules in `.devcontainer/features/languages/`

## Key Commands

- `/update --context` - Generate context files
- `/feature <desc>` - Start new feature
- `/fix <desc>` - Fix a bug

```

**README.md template personnalisé :**

```markdown
# <Nom du projet>

<Description du projet>

## Getting Started

### Prerequisites

- Docker
- VS Code with Remote Containers extension

### Installation

1. Clone the repository
2. Open in VS Code
3. "Reopen in Container"

## Usage

TODO: Add usage instructions

## License

TODO: Add license
```

---

## GARDE-FOUS

| Action | Status |
|--------|--------|
| Écraser CLAUDE.md personnalisé | ❌ **INTERDIT** |
| Écraser README.md personnalisé | ❌ **INTERDIT** |
| Skip vérification footprint | ❌ **INTERDIT** |
| Ignorer choix utilisateur | ❌ **INTERDIT** |

---

## Voir aussi

- `/update --context` - Générer le contexte après init
- `/feature <description>` - Commencer une fonctionnalité
