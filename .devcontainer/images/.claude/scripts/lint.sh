#!/bin/bash
# Lint files based on extension
# Usage: lint.sh <file_path>
#
# Strategy:
#   1. If Makefile exists with lint target → make lint FILE=<path>
#   2. Otherwise → direct linter (eslint, ruff, golangci-lint, etc.)
#
# Note: This focuses on CODE QUALITY (style, errors, best practices).
# typecheck.sh handles strict type checking.

set +e  # Fail-open: hooks should never block unexpectedly

FILE="${1:-}"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

EXT="${FILE##*.}"
DIR=$(dirname "$FILE")
BASENAME=$(basename "$FILE")

# Find project root
find_project_root() {
    local current="$1"
    while [ "$current" != "/" ]; do
        if [ -f "$current/Makefile" ] || \
           [ -f "$current/package.json" ] || \
           [ -f "$current/pyproject.toml" ] || \
           [ -f "$current/go.mod" ] || \
           [ -f "$current/Cargo.toml" ]; then
            echo "$current"
            return
        fi
        current=$(dirname "$current")
    done
    echo "$DIR"
}

PROJECT_ROOT=$(find_project_root "$DIR")

# Check if Makefile has lint target
has_makefile_lint() {
    if [ -f "$PROJECT_ROOT/Makefile" ]; then
        grep -qE "^lint:" "$PROJECT_ROOT/Makefile" 2>/dev/null
        return $?
    fi
    return 1
}

# === Makefile-first approach ===
if has_makefile_lint; then
    cd "$PROJECT_ROOT"
    if grep -qE "FILE\s*[:?]?=" "$PROJECT_ROOT/Makefile" 2>/dev/null; then
        make lint FILE="$FILE" 2>/dev/null || true
    else
        make lint 2>/dev/null || true
    fi
    exit 0
fi

# === Fallback: Direct linters ===

case "$EXT" in
    # JavaScript/TypeScript - eslint
    js|jsx|ts|tsx|mjs|cjs)
        if command -v eslint &>/dev/null; then
            eslint --fix "$FILE" 2>/dev/null || true
        elif command -v npx &>/dev/null && [ -f "$PROJECT_ROOT/package.json" ]; then
            (cd "$PROJECT_ROOT" && npx eslint --fix "$FILE" 2>/dev/null) || true
        fi
        ;;

    # Python - ruff is faster and comprehensive
    py)
        if command -v ruff &>/dev/null; then
            ruff check --fix "$FILE" 2>/dev/null || true
        elif command -v pylint &>/dev/null; then
            pylint --errors-only "$FILE" 2>/dev/null || true
        fi
        ;;

    # Go - golangci-lint is comprehensive
    go)
        if command -v golangci-lint &>/dev/null; then
            golangci-lint run --fix "$FILE" 2>/dev/null || true
        fi
        ;;

    # Rust - clippy is the standard linter
    rs)
        [[ -f "$HOME/.cache/cargo/env" ]] && source "$HOME/.cache/cargo/env"
        if command -v cargo &>/dev/null && [ -f "$PROJECT_ROOT/Cargo.toml" ]; then
            (cd "$PROJECT_ROOT" && cargo clippy --fix --allow-dirty --allow-staged -- -D warnings 2>/dev/null) || true
        fi
        ;;

    # Shell - shellcheck
    sh|bash)
        if command -v shellcheck &>/dev/null; then
            shellcheck "$FILE" 2>/dev/null || true
        fi
        ;;

    # Dockerfile - hadolint
    Dockerfile*)
        if command -v hadolint &>/dev/null; then
            hadolint "$FILE" 2>/dev/null || true
        fi
        ;;

    # YAML - yamllint
    yml|yaml)
        if command -v yamllint &>/dev/null; then
            yamllint -d relaxed "$FILE" 2>/dev/null || true
        fi
        # Ansible-specific lint for playbooks
        if [[ "$FILE" == *"playbook"* ]] || [[ "$FILE" == *"ansible"* ]]; then
            if command -v ansible-lint &>/dev/null; then
                ansible-lint "$FILE" 2>/dev/null || true
            fi
        fi
        ;;

    # Terraform - tflint
    tf|tfvars)
        if command -v tflint &>/dev/null; then
            tflint "$FILE" 2>/dev/null || true
        fi
        ;;

    # C/C++ - clang-tidy
    c|cpp|cc|cxx|h|hpp)
        if command -v clang-tidy &>/dev/null; then
            clang-tidy "$FILE" --fix 2>/dev/null || true
        elif command -v cppcheck &>/dev/null; then
            cppcheck "$FILE" 2>/dev/null || true
        fi
        ;;

    # Java - checkstyle
    java)
        if command -v checkstyle &>/dev/null; then
            checkstyle "$FILE" 2>/dev/null || true
        fi
        ;;

    # Ruby - rubocop
    rb)
        if command -v rubocop &>/dev/null; then
            rubocop -a "$FILE" 2>/dev/null || true
        fi
        ;;

    # PHP - phpstan or syntax check
    php)
        if command -v phpstan &>/dev/null; then
            phpstan analyse "$FILE" 2>/dev/null || true
        elif command -v php &>/dev/null; then
            php -l "$FILE" 2>/dev/null || true
        fi
        ;;

    # Kotlin - ktlint
    kt|kts)
        if command -v ktlint &>/dev/null; then
            ktlint "$FILE" 2>/dev/null || true
        fi
        ;;

    # Swift - swiftlint
    swift)
        if command -v swiftlint &>/dev/null; then
            swiftlint lint --path "$FILE" 2>/dev/null || true
        fi
        ;;

    # Lua - luacheck
    lua)
        if command -v luacheck &>/dev/null; then
            luacheck "$FILE" 2>/dev/null || true
        fi
        ;;

    # SQL - sqlfluff
    sql)
        if command -v sqlfluff &>/dev/null; then
            sqlfluff lint "$FILE" 2>/dev/null || true
        fi
        ;;

    # Markdown - markdownlint
    md)
        if command -v markdownlint &>/dev/null; then
            markdownlint "$FILE" 2>/dev/null || true
        fi
        ;;

    # JSON - jsonlint
    json)
        if command -v jsonlint &>/dev/null; then
            jsonlint -q "$FILE" 2>/dev/null || true
        fi
        ;;

    # HTML - htmlhint
    html|htm)
        if command -v htmlhint &>/dev/null; then
            htmlhint "$FILE" 2>/dev/null || true
        fi
        ;;

    # CSS/SCSS - stylelint
    css|scss|less)
        if command -v stylelint &>/dev/null; then
            stylelint --fix "$FILE" 2>/dev/null || true
        fi
        ;;

    # Elixir - credo
    ex|exs)
        if command -v mix &>/dev/null && [ -f "$PROJECT_ROOT/mix.exs" ]; then
            (cd "$PROJECT_ROOT" && mix credo "$FILE" 2>/dev/null) || true
        fi
        ;;

    # Dart - dart analyze
    dart)
        if command -v dart &>/dev/null; then
            dart analyze "$FILE" 2>/dev/null || true
        fi
        ;;

    # TOML - taplo lint
    toml)
        if command -v taplo &>/dev/null; then
            taplo lint "$FILE" 2>/dev/null || true
        fi
        ;;

    # Protobuf - buf lint
    proto)
        if command -v buf &>/dev/null; then
            buf lint "$FILE" 2>/dev/null || true
        fi
        ;;
esac

exit 0
