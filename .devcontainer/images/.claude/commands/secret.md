---
name: secret
description: |
  Secure secret management with 1Password CLI (op).
  Share secrets between projects via Vault-like path structure.
  Auto-detects project path from git remote origin.
  Use when: storing, retrieving, or listing project secrets.
allowed-tools:
  - "Bash(op:*)"
  - "Bash(git:*)"
  - "Read(**/*)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "AskUserQuestion(*)"
---

# /secret - Secure Secret Management (1Password + Vault-like Paths)

$ARGUMENTS

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Fallback to Grep ONLY for exact string matches or regex patterns.

---

## Overview

Gestion securisee des secrets via **1Password CLI** (`op`) avec une arborescence de paths inspiree de HashiCorp Vault :

- **Peek** - Verifier connectivite 1Password + resoudre le path projet
- **Execute** - Appeler `op` CLI pour push/get/list
- **Synthesize** - Afficher resultat formate

**Backend** : 1Password (via `OP_SERVICE_ACCOUNT_TOKEN`)
**CLI** : `op` (installe dans le devcontainer)
**Pas de MCP** : 1Password n'a pas de MCP officiel (politique deliberee)

---

## Arguments

| Pattern | Action |
|---------|--------|
| `--push <key>=<value>` | Ecrire un secret dans 1Password |
| `--get <key>` | Lire un secret depuis 1Password |
| `--list` | Lister les secrets du projet |
| `--path <path>` | Override le path projet (optionnel) |
| `--help` | Affiche l'aide |

### Exemples

```bash
# Push un secret (path auto = kodflow/devcontainer-template)
/secret --push DB_PASSWORD=mypass

# Push sur un path different (cross-projet)
/secret --push SHARED_TOKEN=abc123 --path kodflow/shared-infra

# Get un secret
/secret --get DB_PASSWORD

# Get depuis un autre path
/secret --get API_KEY --path kodflow/other-project

# Lister les secrets du projet courant
/secret --list

# Lister les secrets d'un autre path
/secret --list --path kodflow/shared-infra
```

---

## --help

```
═══════════════════════════════════════════════════════════════
  /secret - Secure Secret Management (1Password)
═══════════════════════════════════════════════════════════════

Usage: /secret <action> [options]

Actions:
  --push <key>=<value>    Store a secret in 1Password
  --get <key>             Retrieve a secret from 1Password
  --list                  List secrets for current project

Options:
  --path <org/repo>       Override project path (default: auto)
  --help                  Show this help

Path Convention (Vault-like):
  Items are named: <org>/<repo>/<key>
  Default path is auto-detected from git remote origin.
  Example: kodflow/devcontainer-template/DB_PASSWORD

  Without --path: scoped to current project ONLY
  With --path: access any project's secrets

Backend:
  1Password CLI (op) with OP_SERVICE_ACCOUNT_TOKEN
  Items stored as API_CREDENTIAL in configured vault
  Field: "credential" (matches existing MCP token pattern)

Examples:
  /secret --push DB_PASSWORD=s3cret
  /secret --get DB_PASSWORD
  /secret --list
  /secret --push TOKEN=abc --path kodflow/shared
  /secret --get TOKEN --path kodflow/shared

═══════════════════════════════════════════════════════════════
```

---

## Path Convention (Vault-like)

**Arborescence dans 1Password :**

```
<vault>/                              # 1Password vault (default: CI)
├── kodflow/
│   ├── devcontainer-template/        # Projet courant
│   │   ├── DB_PASSWORD               # Item: kodflow/devcontainer-template/DB_PASSWORD
│   │   ├── API_KEY                   # Item: kodflow/devcontainer-template/API_KEY
│   │   └── JWT_SECRET                # Item: kodflow/devcontainer-template/JWT_SECRET
│   ├── shared-infra/                 # Secrets partages
│   │   ├── AWS_CREDENTIALS            # Item: kodflow/shared-infra/AWS_CREDENTIALS
│   │   └── TF_VAR_db_password       # Item: kodflow/shared-infra/TF_VAR_db_password
│   └── other-project/
│       └── STRIPE_KEY                # Item: kodflow/other-project/STRIPE_KEY
└── mcp-github                        # Items existants (pattern legacy)
```

**Resolution du path :**

```bash
# Git remote → path
git remote get-url origin
  → https://github.com/kodflow/devcontainer-template.git
  → path: kodflow/devcontainer-template

# SSH format
  → git@github.com:kodflow/devcontainer-template.git
  → path: kodflow/devcontainer-template

# Token-embedded
  → https://ghp_xxx@github.com/kodflow/devcontainer-template.git
  → path: kodflow/devcontainer-template
```

**Regle stricte** : Sans `--path`, TOUTES les operations sont scopees au path du projet courant. Impossible d'acceder a un path different sans le specifier explicitement.

---

## 1Password Item Format

Chaque secret est stocke comme un item 1Password :

```yaml
item:
  title: "<org>/<repo>/<key>"           # Ex: kodflow/devcontainer-template/DB_PASSWORD
  category: "API_CREDENTIAL"            # Meme categorie que mcp-github, mcp-codacy
  vault: "${OP_VAULT_ID}"               # Vault configure (default: CI)
  fields:
    - name: "credential"                # Champ principal (meme pattern que les tokens MCP)
      value: "<secret_value>"
    - name: "notesPlain"                # Metadata optionnel
      value: "Managed by /secret skill"
```

---

## Phase 1.0 : Peek (OBLIGATOIRE)

**Verifier les prerequis AVANT toute operation :**

```yaml
peek_workflow:
  1_check_op:
    action: "Verifier que op CLI est disponible"
    command: "command -v op"
    on_failure: |
      ABORT avec message:
      "op CLI not found. Install 1Password CLI or run inside DevContainer."

  2_check_token:
    action: "Verifier OP_SERVICE_ACCOUNT_TOKEN"
    command: "test -n \"$OP_SERVICE_ACCOUNT_TOKEN\""
    on_failure: |
      ABORT avec message:
      "OP_SERVICE_ACCOUNT_TOKEN not set. Configure in .devcontainer/.env"

  3_check_vault:
    action: "Verifier acces au vault"
    command: "op vault list --format=json 2>/dev/null | jq -r '.[0].id'"
    store: "VAULT_ID"
    on_failure: |
      ABORT avec message:
      "Cannot access 1Password vault. Check OP_SERVICE_ACCOUNT_TOKEN."

  4_resolve_path:
    action: "Resoudre le path projet depuis git remote"
    command: |
      REMOTE_URL=$(git config --get remote.origin.url 2>/dev/null || echo "")
      # Remove .git suffix
      REMOTE_URL="${REMOTE_URL%.git}"
      # Extract org/repo (handles HTTPS, SSH, token-embedded)
      if [[ "$REMOTE_URL" =~ [:/]([^/]+)/([^/]+)$ ]]; then
        PROJECT_PATH="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
      else
        ABORT "Cannot resolve project path from git remote: $REMOTE_URL"
      fi
    store: "PROJECT_PATH"
    override: "--path argument if provided"
```

**Output Phase 1 :**

```
═══════════════════════════════════════════════════════════════
  /secret - Connection Check
═══════════════════════════════════════════════════════════════

  1Password CLI : op v2.32.0 ✓
  Service Token : OP_SERVICE_ACCOUNT_TOKEN ✓ (set)
  Vault Access  : CI (ypahjj334ixtiyjkytu5hij2im) ✓
  Project Path  : kodflow/devcontainer-template ✓

═══════════════════════════════════════════════════════════════
```

---

## Action: --push

**Ecrire un secret dans 1Password :**

```yaml
push_workflow:
  1_parse_args:
    action: "Parser key=value"
    validation:
      - "key ne contient pas de caracteres speciaux (a-zA-Z0-9_)"
      - "value n'est pas vide"
      - "format exact: KEY=VALUE (un seul =)"

  2_build_title:
    action: "Construire le titre de l'item"
    format: "<PROJECT_PATH>/<key>"
    example: "kodflow/devcontainer-template/DB_PASSWORD"

  3_check_exists:
    action: "Verifier si l'item existe deja"
    command: "op item get '<title>' --vault '$VAULT_ID' 2>/dev/null"
    decision:
      exists: "Update (op item edit)"
      not_exists: "Create (op item create)"

  4a_create:
    condition: "Item n'existe pas"
    command: |
      op item create \
        --category=API_CREDENTIAL \
        --title='<org>/<repo>/<key>' \
        --vault='$VAULT_ID' \
        'credential=<value>'
    note: "Le champ 'credential' matche le pattern des tokens MCP existants"

  4b_update:
    condition: "Item existe deja"
    command: |
      op item edit '<org>/<repo>/<key>' \
        --vault='$VAULT_ID' \
        'credential=<value>'

  5_confirm:
    action: "Verifier que l'item est bien stocke"
    command: "op item get '<title>' --vault '$VAULT_ID' --format=json | jq -r '.title'"
```

**Output --push (nouveau) :**

```
═══════════════════════════════════════════════════════════════
  /secret --push
═══════════════════════════════════════════════════════════════

  Path   : kodflow/devcontainer-template
  Key    : DB_PASSWORD
  Action : Created

  Item: kodflow/devcontainer-template/DB_PASSWORD
  Vault: CI
  Field: credential
  Status: ✓ Stored successfully

═══════════════════════════════════════════════════════════════
```

**Output --push (update) :**

```
═══════════════════════════════════════════════════════════════
  /secret --push
═══════════════════════════════════════════════════════════════

  Path   : kodflow/devcontainer-template
  Key    : DB_PASSWORD
  Action : Updated (existing item)

  Item: kodflow/devcontainer-template/DB_PASSWORD
  Vault: CI
  Field: credential
  Status: ✓ Updated successfully

═══════════════════════════════════════════════════════════════
```

---

## Action: --get

**Lire un secret depuis 1Password :**

```yaml
get_workflow:
  1_build_title:
    action: "Construire le titre"
    format: "<PROJECT_PATH>/<key>"

  2_retrieve:
    action: "Recuperer la valeur"
    command: |
      op item get '<org>/<repo>/<key>' \
        --vault='$VAULT_ID' \
        --fields='credential' \
        --reveal
    fallback_fields: ["credential", "password", "identifiant", "mot de passe"]
    note: "Meme logique de fallback que get_1password_field dans postStart.sh"

  3_display:
    action: "Afficher le resultat"
    security: "La valeur est revelee UNE SEULE FOIS dans l'output"
```

**Output --get (success) :**

```
═══════════════════════════════════════════════════════════════
  /secret --get
═══════════════════════════════════════════════════════════════

  Path  : kodflow/devcontainer-template
  Key   : DB_PASSWORD
  Value : s3cr3t_p4ssw0rd

═══════════════════════════════════════════════════════════════
```

**Output --get (not found) :**

```
═══════════════════════════════════════════════════════════════
  /secret --get
═══════════════════════════════════════════════════════════════

  Path  : kodflow/devcontainer-template
  Key   : DB_PASSWORD
  Status: ✗ Not found

  Hint: Use /secret --list to see available secrets
        Use /secret --push DB_PASSWORD=<value> to create it

═══════════════════════════════════════════════════════════════
```

---

## Action: --list

**Lister les secrets d'un path :**

```yaml
list_workflow:
  1_list_items:
    action: "Lister tous les items du vault"
    command: |
      op item list \
        --vault='$VAULT_ID' \
        --format=json
    filter: "Filtrer par prefix PROJECT_PATH/"

  2_display:
    action: "Afficher la liste filtree"
    format: "Tableau avec titre, categorie, date de modification"
    extract_key: "Supprimer le prefix path/ pour n'afficher que la cle"
```

**Output --list (avec secrets) :**

```
═══════════════════════════════════════════════════════════════
  /secret --list
═══════════════════════════════════════════════════════════════

  Path: kodflow/devcontainer-template

  | Key             | Category       | Updated            |
  |-----------------|----------------|--------------------|
  | DB_PASSWORD     | API_CREDENTIAL | 2026-02-09 10:30   |
  | API_KEY         | API_CREDENTIAL | 2026-02-08 14:22   |
  | JWT_SECRET      | API_CREDENTIAL | 2026-02-07 09:15   |

  Total: 3 secrets

═══════════════════════════════════════════════════════════════
```

**Output --list (vide) :**

```
═══════════════════════════════════════════════════════════════
  /secret --list
═══════════════════════════════════════════════════════════════

  Path: kodflow/devcontainer-template

  No secrets found for this project.

  Hint: Use /secret --push KEY=VALUE to store a secret
        Use /secret --list --path / to see all paths

═══════════════════════════════════════════════════════════════
```

**Output --list --path / (tous les paths) :**

```
═══════════════════════════════════════════════════════════════
  /secret --list --path /
═══════════════════════════════════════════════════════════════

  All secrets (grouped by path):

  kodflow/devcontainer-template/ (3 secrets)
    ├─ DB_PASSWORD
    ├─ API_KEY
    └─ JWT_SECRET

  kodflow/shared-infra/ (2 secrets)
    ├─ AWS_CREDENTIALS
    └─ TF_VAR_db_password

  (legacy items without path)
    ├─ mcp-github
    ├─ mcp-codacy
    └─ Coderabbit TOKEN

  Total: 8 items (5 with paths, 3 legacy)

═══════════════════════════════════════════════════════════════
```

---

## Cross-Project Secret Sharing

**Utiliser `--path` pour partager des secrets entre projets :**

```yaml
sharing_patterns:
  # Partager un secret infra commun
  push_shared:
    command: '/secret --push AWS_CREDENTIALS=xxx... --path kodflow/shared-infra'
    note: "Accessible par tous les projets kodflow"

  # Recuperer depuis un autre projet
  get_cross_project:
    command: '/secret --get STRIPE_KEY --path kodflow/payment-service'
    note: "Debloquer une situation en recuperant un secret d'un autre projet"

  # Debloquer une situation
  unblock_workflow:
    1: '/secret --list --path /'
    2: 'Identifier le secret necessaire et son path'
    3: '/secret --get <key> --path <org>/<repo>'
```

---

## Integration avec les autres skills

### Depuis /init

```yaml
init_integration:
  phase: "Phase 3 (Parallelize)"
  agent: "vault-checker"
  check:
    - "op CLI disponible"
    - "OP_SERVICE_ACCOUNT_TOKEN set"
    - "Vault accessible"
    - "Nombre de secrets pour le projet courant"
  report_section: "1Password Secrets"
```

### Depuis /git (pre-commit)

```yaml
git_integration:
  phase: "Phase 3 (Parallelize)"
  agent: "secret-scan"
  check:
    - "Scanner git diff --cached pour des patterns de secrets"
    - "Patterns: ghp_, glpat-, sk-, pk_, postgres://, mysql://, mongodb://"
    - "Si trouve: AVERTIR (pas bloquer)"
    - "Proposer: /secret --push <key>=<detected_value>"
  behavior: "WARNING only, ne bloque PAS le commit"
```

### Depuis /do

```yaml
do_integration:
  phase: "Phase 0 (avant Questions)"
  check:
    - "Si la tache mentionne: secret, token, credential, password, API key"
    - "Lister les secrets disponibles pour le projet"
    - "Proposer de les utiliser ou d'en creer de nouveaux"
  behavior: "Informatif, aide a debloquer"
```

### Depuis /infra

```yaml
infra_integration:
  phase: "Avant --plan et --apply"
  check:
    - "Lister secrets du projet avec prefix TF_VAR_"
    - "Verifier si des variables Terraform referencent des secrets"
    - "Proposer de recuperer depuis 1Password"
  cross_path: "Permettre --path kodflow/shared-infra pour secrets partages"
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Status | Raison |
|--------|--------|--------|
| Reveler un secret sans --get explicite | ❌ **INTERDIT** | Securite |
| Ecrire un secret dans les logs | ❌ **INTERDIT** | Securite |
| Push sans confirmation si item existe | ❌ **INTERDIT** | Eviter ecrasement |
| Acceder a un path different sans --path | ❌ **INTERDIT** | Scope strict |
| Fonctionner sans OP_SERVICE_ACCOUNT_TOKEN | ❌ **INTERDIT** | Auth requise |
| Supprimer un secret (pas de --delete) | ❌ **INTERDIT** | Utiliser 1Password UI |
| Skip Phase 1 (Peek) | ❌ **INTERDIT** | Verification connexion |
