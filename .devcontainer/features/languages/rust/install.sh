#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# Rust Development Environment Installer
# =============================================================================
# Optimized for DevContainer: heavy libs (WebKitGTK) are in base image
# This script installs: rustup, toolchain, components, cargo tools
# =============================================================================

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

log()  { echo -e "$*"; }
ok()   { log "${GREEN}✓${NC} $*"; }
warn() { log "${YELLOW}⚠${NC} $*"; }
err()  { log "${RED}✗${NC} $*" >&2; }

# Environment
export CARGO_HOME="${CARGO_HOME:-$HOME/.cache/cargo}"
export RUSTUP_HOME="${RUSTUP_HOME:-$HOME/.cache/rustup}"

echo "========================================="
echo "Installing Rust Development Environment"
echo "========================================="

# =============================================================================
# Minimal System Dependencies (libs are in base image)
# =============================================================================
log "${YELLOW}Installing minimal dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y --no-install-recommends \
    curl build-essential gcc make cmake pkg-config libssl-dev
ok "Dependencies ready"

# =============================================================================
# Rustup Installation (idempotent)
# =============================================================================
if command -v rustup &>/dev/null; then
    ok "rustup already installed"
else
    log "${YELLOW}Installing rustup...${NC}"
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path
    ok "rustup installed"
fi

export PATH="$CARGO_HOME/bin:$PATH"
# shellcheck source=/dev/null
[[ -f "$CARGO_HOME/env" ]] && source "$CARGO_HOME/env"

# =============================================================================
# Toolchain & Components
# =============================================================================
log "${YELLOW}Setting up stable toolchain...${NC}"
rustup toolchain install stable --profile minimal
rustup default stable
rustup component add rust-analyzer clippy rustfmt
ok "Toolchain: $(rustc --version)"

# =============================================================================
# Targets (idempotent installation)
# =============================================================================
ensure_target() {
    local target="$1" mode="${2:-required}"
    if rustup target list --installed 2>/dev/null | grep -q "^${target}$"; then
        ok "${target} (cached)"
        return 0
    fi
    if rustup target add "$target" 2>/dev/null; then
        ok "${target}"
    elif [[ "$mode" == "optional" ]]; then
        warn "${target} not available"
    else
        err "${target} failed" && return 1
    fi
}

log "${YELLOW}Installing compilation targets...${NC}"

# Host target (always required)
HOST_ARCH=$(uname -m)
case "$HOST_ARCH" in
    x86_64)  ensure_target "x86_64-unknown-linux-gnu" ;;
    aarch64) ensure_target "aarch64-unknown-linux-gnu" ;;
esac

# WASM targets (lightweight, useful for web dev)
ensure_target "wasm32-unknown-unknown" "optional"
ensure_target "wasm32-wasip1" "optional"
ensure_target "wasm32-wasip2" "optional"

# =============================================================================
# Cargo-binstall (fast binary installer)
# =============================================================================
log "${YELLOW}Installing cargo-binstall...${NC}"
if command -v cargo-binstall &>/dev/null; then
    ok "cargo-binstall (cached)"
else
    curl -L --proto '=https' --tlsv1.2 -sSf \
        https://raw.githubusercontent.com/cargo-bins/cargo-binstall/main/install-from-binstall-release.sh | bash
    ok "cargo-binstall installed"
fi

# =============================================================================
# Cargo Tools (via binstall for speed)
# =============================================================================
FAILED_TOOLS=()

install_tool() {
    local tool="$1"
    if command -v "${tool}" &>/dev/null 2>&1; then
        ok "${tool} (cached)"
        return 0
    fi
    log "${YELLOW}Installing ${tool}...${NC}"
    if cargo binstall --no-confirm --locked "$tool" 2>/dev/null; then
        ok "${tool} (binary)"
    elif cargo install --locked "$tool" 2>/dev/null; then
        ok "${tool} (compiled)"
    else
        warn "${tool} failed"
        FAILED_TOOLS+=("$tool")
    fi
}

log "${YELLOW}Installing development tools...${NC}"

# Core tools (always installed)
CORE_TOOLS=(
    cargo-watch       # Auto-rebuild on file changes
    cargo-nextest     # Fast test runner
    cargo-audit       # Security vulnerability scanner
    cargo-deny        # Dependency security checker
    cargo-edit        # Dependency management (add/remove)
    cargo-expand      # Macro expansion viewer
    cargo-outdated    # Dependency update checker
    cargo-tarpaulin   # Code coverage tool
)

# Desktop/WASM tools (Tauri, WebAssembly)
DESKTOP_TOOLS=(
    tauri-cli         # Tauri desktop app framework CLI
    wasm-pack         # WASM packaging and publishing
    wasm-bindgen-cli  # JS bindings generator for WASM
)

# Install all tools
for tool in "${CORE_TOOLS[@]}" "${DESKTOP_TOOLS[@]}"; do
    install_tool "$tool"
done

# MCP server (may not have prebuilt binaries)
install_tool "rust-analyzer-mcp"

# =============================================================================
# Shell Integration
# =============================================================================
log "${YELLOW}Configuring shell integration...${NC}"
CARGO_ENV='[[ -f "$HOME/.cache/cargo/env" ]] && source "$HOME/.cache/cargo/env"'
for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
    if [[ -f "$rc" ]] && ! grep -q "cargo/env" "$rc"; then
        echo -e "\n# Rust/Cargo environment\n$CARGO_ENV" >> "$rc"
    fi
done
ok "Shell integration configured"

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Rust environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - rustup (Rust toolchain manager)"
echo "  - $(rustc --version)"
echo "  - $(cargo --version)"
echo ""
echo "Development tools:"
echo "  - cargo-binstall (fast binary installer)"
echo "  - rust-analyzer (LSP)"
echo "  - clippy (linter)"
echo "  - rustfmt (formatter)"
echo "  - cargo-watch, cargo-nextest, cargo-audit"
echo "  - cargo-deny, cargo-edit, cargo-expand"
echo "  - cargo-outdated, cargo-tarpaulin"
echo ""
echo "Desktop & WASM tools:"
echo "  - tauri-cli (desktop apps)"
echo "  - wasm-pack, wasm-bindgen-cli"
echo "  - wasm32-unknown-unknown, wasm32-wasip1/2"
echo ""
echo "Cache directories:"
echo "  - CARGO_HOME: $CARGO_HOME"
echo "  - RUSTUP_HOME: $RUSTUP_HOME"
echo ""
echo "Note: WebKitGTK/Tauri libs are pre-installed in base image"
echo ""

if [[ ${#FAILED_TOOLS[@]} -gt 0 ]]; then
    warn "Some tools failed to install: ${FAILED_TOOLS[*]}"
fi
