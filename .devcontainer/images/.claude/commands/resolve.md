# Resolve - Auto-fix PR Reviews

$ARGUMENTS

---

## Description

Commande automatisée pour corriger les issues de code review sur une PR :

- **Codacy** : Issues statiques (linting, security, best practices)
- **CodeRabbit** : Commentaires de review IA
- **Qodo Merge** : Commentaires PR-Agent

**Workflow itératif :**

1. Récupère les issues/commentaires via MCP
2. Applique les corrections
3. Commit + Push
4. Répète jusqu'à résolution complète
5. Marque les reviews comme résolues

---

## Arguments

| Pattern | Action |
|---------|--------|
| (vide) | Résout issues de la PR de la branche courante |
| `--pr <number>` | Résout issues d'une PR spécifique |
| `--codacy-only` | Ne traite que les issues Codacy |
| `--coderabbit-only` | Ne traite que les commentaires CodeRabbit |
| `--qodo-only` | Ne traite que les commentaires Qodo Merge |
| `--dry-run` | Affiche les issues sans les corriger |
| `--help` | Affiche l'aide |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /resolve - Auto-fix PR Reviews
═══════════════════════════════════════════════

Usage: /resolve [options]

Options:
  (vide)              Résout la PR de la branche courante
  --pr <number>       Résout une PR spécifique
  --codacy-only       Issues Codacy uniquement
  --coderabbit-only   Commentaires CodeRabbit uniquement
  --qodo-only         Commentaires Qodo Merge uniquement
  --dry-run           Affiche sans corriger
  --help              Affiche cette aide

Intégrations:
  Codacy              Issues statiques (boucle itérative)
  CodeRabbit          Commentaires PR + @coderabbitai resolve
  Qodo Merge          Commentaires PR-Agent

Exemples:
  /resolve                    Résout tout sur PR courante
  /resolve --pr 90            Résout PR #90
  /resolve --codacy-only      Codacy uniquement
  /resolve --dry-run          Preview des issues
═══════════════════════════════════════════════
```

---

## Priorité des outils

**IMPORTANT** : Toujours privilégier les outils MCP.

| Action | MCP Tool |
|--------|----------|
| Lister PRs | `mcp__github__list_pull_requests` |
| Issues Codacy | `mcp__codacy__codacy_list_pull_request_issues` |
| Commentaires PR | `mcp__github__get_pull_request_comments` |
| Poster commentaire | `mcp__github__add_issue_comment` |

**Extraction owner/repo :**
```bash
git remote get-url origin | sed -E 's#.*[:/]([^/]+)/([^/.]+)(\.git)?$#\1 \2#'
```

---

## Workflow complet

### Étape 1 : Détection de la PR

```bash
BRANCH=$(git branch --show-current)
MAIN_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")

# Bloquer si sur main
if [[ "$BRANCH" == "$MAIN_BRANCH" || "$BRANCH" == "master" ]]; then
    echo "❌ Impossible de résoudre sur la branche principale"
    exit 1
fi

# Extraire owner/repo
REMOTE=$(git remote get-url origin)
OWNER=$(echo "$REMOTE" | sed -E 's#.*[:/]([^/]+)/([^/.]+).*#\1#')
REPO=$(echo "$REMOTE" | sed -E 's#.*[:/]([^/]+)/([^/.]+).*#\2#')
```

**Trouver la PR via MCP :**
```
mcp__github__list_pull_requests:
  owner: $OWNER
  repo: $REPO
  head: "$OWNER:$BRANCH"
  state: "open"
```

**Si `--pr <number>` spécifié :** Utiliser directement ce numéro.

---

### Étape 2 : Boucle Codacy (itérative)

**Skip si `--coderabbit-only`**

```
MAX_ITERATIONS = 5
iteration = 0

WHILE iteration < MAX_ITERATIONS:
    iteration++

    # Récupérer issues Codacy
    issues = mcp__codacy__codacy_list_pull_request_issues(
        provider: "gh",
        organization: $OWNER,
        repository: $REPO,
        pullRequestNumber: $PR_NUMBER,
        status: "new"
    )

    IF issues.data.length == 0:
        log_success "✓ Codacy: 0 issues"
        BREAK

    log_info "Iteration $iteration: ${issues.data.length} issues"

    # Grouper par fichier
    files_issues = group_by(issues.data, "filePath")

    FOR each (file, file_issues) in files_issues:
        # Lire le fichier (OBLIGATOIRE avant edit)
        Read(file)

        FOR each issue in file_issues:
            # Analyser le message et la suggestion
            line = issue.lineNumber
            message = issue.message
            suggestion = issue.suggestion (si disponible)

            # Appliquer la correction
            IF suggestion:
                Edit(file, old=lineText, new=suggestion)
            ELSE:
                # Analyser et corriger selon le pattern
                fix_issue_by_pattern(file, line, message)

        # Commit atomique par fichier
        git add $file
        git commit -m "fix: resolve Codacy issues in $(basename $file)"

    # Push et attendre re-analyse
    git push
    sleep 15

# Alerte si max iterations atteint
IF iteration >= MAX_ITERATIONS AND issues.data.length > 0:
    log_warning "⚠ Max iterations reached, ${issues.data.length} issues remaining"
```

---

### Étape 3 : Boucle CodeRabbit

**Skip si `--codacy-only` ou `--qodo-only`**

```
# Récupérer tous les commentaires de la PR
comments = mcp__github__get_pull_request_comments(
    owner: $OWNER,
    repo: $REPO,
    pull_number: $PR_NUMBER
)

# Filtrer les commentaires CodeRabbit actionables
coderabbit_comments = comments.filter(c =>
    c.user.login == "coderabbitai[bot]"
    AND (
        c.body contains "Potential issue"
        OR c.body contains "suggestion"
        OR c.body contains "Committable suggestion"
    )
)

IF coderabbit_comments.length == 0:
    log_success "✓ CodeRabbit: 0 commentaires actionables"
ELSE:
    log_info "${coderabbit_comments.length} commentaires CodeRabbit à traiter"

    FOR each comment in coderabbit_comments:
        file = comment.path
        line = comment.position OR comment.original_position

        # Lire le fichier (OBLIGATOIRE)
        Read(file)

        # Extraire la suggestion du commentaire
        suggestion = extract_committable_suggestion(comment.body)

        IF suggestion:
            Edit(file, apply_suggestion)
        ELSE:
            fix_based_on_comment(file, line, comment.body)

        git add $file
        git commit -m "fix: address CodeRabbit review comment"

    git push

    # Marquer comme résolu
    mcp__github__add_issue_comment(
        owner: $OWNER,
        repo: $REPO,
        issue_number: $PR_NUMBER,
        body: "@coderabbitai resolve\n\nAll review comments have been addressed."
    )
```

---

### Étape 4 : Boucle Qodo Merge

**Skip si `--codacy-only` ou `--coderabbit-only`**

```
# Filtrer les commentaires Qodo Merge actionables
qodo_comments = comments.filter(c =>
    c.user.login == "qodo-merge-pro[bot]"
    AND (
        c.body contains "suggestion"
        OR c.body contains "Code suggestion"
        OR c.body contains "Recommended fix"
    )
)

IF qodo_comments.length == 0:
    log_success "✓ Qodo Merge: 0 commentaires actionables"
ELSE:
    log_info "${qodo_comments.length} commentaires Qodo à traiter"

    FOR each comment in qodo_comments:
        file = comment.path
        line = comment.position OR comment.original_position

        # Lire le fichier (OBLIGATOIRE)
        Read(file)

        # Extraire la suggestion
        suggestion = extract_qodo_suggestion(comment.body)

        IF suggestion:
            Edit(file, apply_suggestion)
        ELSE:
            fix_based_on_comment(file, line, comment.body)

        git add $file
        git commit -m "fix: address Qodo Merge review comment"

    git push
```

**Extraction de suggestions :**
```
extract_committable_suggestion(body):  # CodeRabbit
    # Pattern: ```suggestion ... ```
    match = regex(body, /```suggestion\n(.*?)\n```/s)
    IF match:
        return match[1]

    # Pattern: Committable suggestion block
    match = regex(body, /📝 Committable suggestion.*?```[a-z]*\n(.*?)\n```/s)
    IF match:
        return match[1]

    return null

extract_qodo_suggestion(body):  # Qodo Merge
    # Pattern: Code suggestion block
    match = regex(body, /```[a-z]*\n(.*?)\n```/s)
    IF match:
        return match[1]

    return null
```

---

### Étape 5 : Résumé final

```
═══════════════════════════════════════════════
  /resolve - PR #$PR_NUMBER
═══════════════════════════════════════════════

  Codacy Issues
─────────────────────────────────────────────
  Iterations     : $codacy_iterations
  Issues fixed   : $codacy_fixed
  Status         : ✓ Clean

  CodeRabbit Comments
─────────────────────────────────────────────
  Comments found : $coderabbit_total
  Comments fixed : $coderabbit_fixed
  @resolve posted: ✓

  Qodo Merge Comments
─────────────────────────────────────────────
  Comments found : $qodo_total
  Comments fixed : $qodo_fixed

  Summary
─────────────────────────────────────────────
  Total commits  : $total_commits
  Files modified : $files_count

  PR: https://github.com/$OWNER/$REPO/pull/$PR_NUMBER

═══════════════════════════════════════════════
```

---

## --dry-run

Mode preview sans modification :

```
═══════════════════════════════════════════════
  /resolve --dry-run - PR #90
═══════════════════════════════════════════════

  Codacy Issues (would fix)
─────────────────────────────────────────────
  1. postStart.sh:304 - SC2016 (shellcheck)
     "Expressions don't expand in single quotes"

  2. CLAUDE.md:12 - MD022 (markdownlint)
     "Expected: 1; Actual: 0; Below"

  CodeRabbit Comments (would fix)
─────────────────────────────────────────────
  1. .coderabbit.yaml - Nitpick
     "Clarifier le commentaire sur le profil"

  2. postStart.sh - Major
     "Risques de sécurité dans l'exécution"

  Summary
─────────────────────────────────────────────
  Would fix: 2 Codacy issues, 2 comments
  No changes made (dry-run mode)

═══════════════════════════════════════════════
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Status |
|--------|--------|
| Fix sans lecture préalable du fichier | ❌ **INTERDIT** |
| Ignorer issues CRITICAL/BLOCKER | ❌ **INTERDIT** |
| Plus de 5 itérations Codacy | ❌ **STOP + alerte** |
| Push sur main/master | ❌ **INTERDIT** |
| Modifier fichiers hors diff PR | ❌ **INTERDIT** |
| Skip @coderabbitai resolve | ❌ **INTERDIT** |

---

## Patterns de correction courants

### Codacy / ShellCheck

| Pattern ID | Correction |
|------------|------------|
| SC2016 | Single quotes → Double quotes avec escape |
| SC2086 | Ajouter quotes autour de $variable |
| SC2046 | Ajouter quotes autour de $(command) |

### Codacy / Markdownlint

| Pattern ID | Correction |
|------------|------------|
| MD022 | Ajouter ligne vide après heading |
| MD032 | Ajouter ligne vide autour des listes |
| MD033 | Escape `<tags>` avec backticks |

### CodeRabbit suggestions

| Type | Action |
|------|--------|
| Committable suggestion | Appliquer le bloc ```suggestion``` |
| Correction proposée | Appliquer le diff proposé |
| Nitpick/Trivial | Ignorer ou appliquer selon contexte |

---

## Voir aussi

- `/review` - Lancer une review locale ou externe
- `/git --commit` - Commit et PR workflow
- `/apply --pr` - Appliquer suggestions CodeRabbit (legacy)
