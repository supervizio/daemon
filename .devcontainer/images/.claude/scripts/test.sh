#!/bin/bash
# Run tests for modified files
# Usage: test.sh <file_path>
#
# Strategy:
#   1. If Makefile exists with test target → make test FILE=<path>
#   2. Otherwise → direct test runner (jest, pytest, go test, cargo test)
#
# Makefile convention:
#   make test              # Run all tests
#   make test FILE=path    # Run tests for specific file

set -e

FILE="$1"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

EXT="${FILE##*.}"
BASENAME=$(basename "$FILE")
DIR=$(dirname "$FILE")

# Find project root (look for Makefile or common config files)
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

# Check if Makefile has test target
has_makefile_test() {
    if [ -f "$PROJECT_ROOT/Makefile" ]; then
        grep -qE "^test:" "$PROJECT_ROOT/Makefile" 2>/dev/null
        return $?
    fi
    return 1
}

# === Makefile-first approach ===
if has_makefile_test; then
    cd "$PROJECT_ROOT"
    # Try make test with FILE parameter (common convention)
    if grep -qE "FILE\s*[:?]?=" "$PROJECT_ROOT/Makefile" 2>/dev/null; then
        make test FILE="$FILE" 2>/dev/null || true
    else
        # Just run make test (will run all tests)
        make test 2>/dev/null || true
    fi
    exit 0
fi

# === Fallback: Direct test runners ===

# Check if this is a test file
IS_TEST=0
case "$BASENAME" in
    *.test.*|*.spec.*|*_test.*|test_*)
        IS_TEST=1
        ;;
esac

case "$EXT" in
    # JavaScript/TypeScript
    js|jsx|ts|tsx)
        if [ $IS_TEST -eq 1 ]; then
            if [ -f "$PROJECT_ROOT/package.json" ]; then
                cd "$PROJECT_ROOT"
                # Check for test script in package.json
                if grep -q '"test"' package.json 2>/dev/null; then
                    npm test -- "$FILE" 2>/dev/null || \
                    pnpm test "$FILE" 2>/dev/null || \
                    yarn test "$FILE" 2>/dev/null || true
                elif command -v vitest &>/dev/null; then
                    vitest run "$FILE" 2>/dev/null || true
                elif command -v jest &>/dev/null; then
                    jest "$FILE" --passWithNoTests 2>/dev/null || true
                fi
            fi
        fi
        ;;

    # Python
    py)
        if [ $IS_TEST -eq 1 ]; then
            cd "$PROJECT_ROOT"
            if command -v pytest &>/dev/null; then
                pytest "$FILE" -v 2>/dev/null || true
            elif command -v python &>/dev/null; then
                python -m pytest "$FILE" -v 2>/dev/null || true
            fi
        fi
        ;;

    # Go - tests alongside source files
    go)
        if [[ "$BASENAME" == *"_test.go" ]]; then
            if command -v go &>/dev/null; then
                (cd "$DIR" && go test -v -run . 2>/dev/null) || true
            fi
        fi
        ;;

    # Rust
    rs)
        [[ -f "$HOME/.cache/cargo/env" ]] && source "$HOME/.cache/cargo/env"
        if command -v cargo &>/dev/null; then
            if [[ "$FILE" == *"tests"* ]] || grep -q "#\[test\]" "$FILE" 2>/dev/null; then
                (cd "$PROJECT_ROOT" && cargo test 2>/dev/null) || true
            fi
        fi
        ;;

    # Elixir
    ex|exs)
        if [[ "$BASENAME" == *"_test.exs" ]]; then
            if command -v mix &>/dev/null && [ -f "$PROJECT_ROOT/mix.exs" ]; then
                (cd "$PROJECT_ROOT" && mix test "$FILE" 2>/dev/null) || true
            fi
        fi
        ;;

    # Ruby
    rb)
        if [[ "$BASENAME" == *"_spec.rb" ]] || [[ "$BASENAME" == *"_test.rb" ]]; then
            cd "$PROJECT_ROOT"
            if command -v rspec &>/dev/null && [[ "$BASENAME" == *"_spec.rb" ]]; then
                rspec "$FILE" 2>/dev/null || true
            elif command -v ruby &>/dev/null; then
                ruby -Itest "$FILE" 2>/dev/null || true
            fi
        fi
        ;;

    # PHP
    php)
        if [[ "$BASENAME" == *"Test.php" ]]; then
            if command -v phpunit &>/dev/null; then
                phpunit "$FILE" 2>/dev/null || true
            elif [ -f "$PROJECT_ROOT/vendor/bin/phpunit" ]; then
                "$PROJECT_ROOT/vendor/bin/phpunit" "$FILE" 2>/dev/null || true
            fi
        fi
        ;;
esac

exit 0
