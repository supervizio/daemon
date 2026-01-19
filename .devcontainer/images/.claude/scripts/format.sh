#!/bin/bash
# Auto-format files based on extension
# Usage: format.sh <file_path>
#
# Strategy:
#   1. If Makefile exists with fmt/format target → make fmt FILE=<path>
#   2. Otherwise → direct formatter (prettier, ruff, goimports, etc.)
#
# Note: Most formatters also handle import sorting (goimports, ruff, rustfmt).

set +e  # Fail-open: hooks should never block unexpectedly

FILE="${1:-}"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

EXT="${FILE##*.}"
DIR=$(dirname "$FILE")

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

# Check if Makefile has fmt or format target
has_makefile_fmt() {
    if [ -f "$PROJECT_ROOT/Makefile" ]; then
        grep -qE "^(fmt|format):" "$PROJECT_ROOT/Makefile" 2>/dev/null
        return $?
    fi
    return 1
}

# === Makefile-first approach ===
if has_makefile_fmt; then
    cd "$PROJECT_ROOT"
    # Try fmt first (more common), then format
    TARGET="fmt"
    if ! grep -qE "^fmt:" "$PROJECT_ROOT/Makefile" 2>/dev/null; then
        TARGET="format"
    fi
    if grep -qE "FILE\s*[:?]?=" "$PROJECT_ROOT/Makefile" 2>/dev/null; then
        make "$TARGET" FILE="$FILE" 2>/dev/null || true
    else
        make "$TARGET" 2>/dev/null || true
    fi
    exit 0
fi

# === Fallback: Direct formatters ===

case "$EXT" in
    # JavaScript/TypeScript - prettier is the standard
    js|jsx|ts|tsx|mjs|cjs)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        elif command -v npx &>/dev/null && [ -f "$PROJECT_ROOT/package.json" ]; then
            (cd "$PROJECT_ROOT" && npx prettier --write "$FILE" 2>/dev/null) || true
        fi
        ;;

    # Python - ruff format (includes import sorting)
    py)
        if command -v ruff &>/dev/null; then
            ruff format "$FILE" 2>/dev/null || true
            ruff check --select I --fix "$FILE" 2>/dev/null || true  # Import sorting
        elif command -v black &>/dev/null; then
            black --quiet "$FILE" 2>/dev/null || true
            # isort for imports if black is used
            if command -v isort &>/dev/null; then
                isort --quiet "$FILE" 2>/dev/null || true
            fi
        fi
        ;;

    # Go - goimports (format + imports)
    go)
        if command -v goimports &>/dev/null; then
            goimports -w "$FILE" 2>/dev/null || true
        elif command -v gofmt &>/dev/null; then
            gofmt -w "$FILE" 2>/dev/null || true
        fi
        ;;

    # Rust - rustfmt (handles imports too)
    rs)
        [[ -f "$HOME/.cache/cargo/env" ]] && source "$HOME/.cache/cargo/env"
        if command -v rustfmt &>/dev/null; then
            rustfmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # JSON - prettier or jq
    json)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        elif command -v jq &>/dev/null; then
            TMP=$(mktemp)
            if jq '.' "$FILE" > "$TMP" 2>/dev/null; then
                mv "$TMP" "$FILE"
            else
                rm -f "$TMP"
            fi
        fi
        ;;

    # YAML - prettier or yamlfmt
    yml|yaml)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        elif command -v yamlfmt &>/dev/null; then
            yamlfmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # Markdown - prettier
    md)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        fi
        ;;

    # Terraform - terraform fmt
    tf|tfvars)
        if command -v terraform &>/dev/null; then
            terraform fmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # Shell - shfmt
    sh|bash)
        if command -v shfmt &>/dev/null; then
            shfmt -w "$FILE" 2>/dev/null || true
        fi
        ;;

    # C/C++ - clang-format (includes sorting)
    c|cpp|cc|cxx|h|hpp)
        if command -v clang-format &>/dev/null; then
            clang-format -i --sort-includes "$FILE" 2>/dev/null || true
        fi
        ;;

    # Java - google-java-format
    java)
        if command -v google-java-format &>/dev/null; then
            google-java-format --replace "$FILE" 2>/dev/null || true
        fi
        ;;

    # HTML/CSS/SCSS - prettier
    html|htm|css|scss|less)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        fi
        ;;

    # XML - xmllint
    xml)
        if command -v xmllint &>/dev/null; then
            xmllint --format "$FILE" --output "$FILE" 2>/dev/null || true
        fi
        ;;

    # SQL - sql-formatter or pg_format
    sql)
        if command -v sql-formatter &>/dev/null; then
            sql-formatter "$FILE" -o "$FILE" 2>/dev/null || true
        elif command -v pg_format &>/dev/null; then
            pg_format -i "$FILE" 2>/dev/null || true
        fi
        ;;

    # Lua - stylua
    lua)
        if command -v stylua &>/dev/null; then
            stylua "$FILE" 2>/dev/null || true
        fi
        ;;

    # Ruby - rubocop
    rb)
        if command -v rubocop &>/dev/null; then
            rubocop -a "$FILE" 2>/dev/null || true
        fi
        ;;

    # PHP - php-cs-fixer
    php)
        if command -v php-cs-fixer &>/dev/null; then
            php-cs-fixer fix "$FILE" --quiet 2>/dev/null || true
        fi
        ;;

    # Kotlin - ktlint
    kt|kts)
        if command -v ktlint &>/dev/null; then
            ktlint -F "$FILE" 2>/dev/null || true
        fi
        ;;

    # Swift - swiftformat
    swift)
        if command -v swiftformat &>/dev/null; then
            swiftformat "$FILE" 2>/dev/null || true
        fi
        ;;

    # Dart - dart format
    dart)
        if command -v dart &>/dev/null; then
            dart format "$FILE" 2>/dev/null || true
        fi
        ;;

    # Elixir - mix format
    ex|exs)
        if command -v mix &>/dev/null; then
            mix format "$FILE" 2>/dev/null || true
        fi
        ;;

    # Zig - zig fmt
    zig)
        if command -v zig &>/dev/null; then
            zig fmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # Nim - nimpretty
    nim)
        if command -v nimpretty &>/dev/null; then
            nimpretty "$FILE" 2>/dev/null || true
        fi
        ;;

    # TOML - taplo fmt
    toml)
        if command -v taplo &>/dev/null; then
            taplo fmt "$FILE" 2>/dev/null || true
        fi
        ;;
esac

exit 0
