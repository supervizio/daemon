#!/bin/bash
set -e

echo "========================================="
echo "Installing Rust Development Environment"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Environment variables
export CARGO_HOME="${CARGO_HOME:-$HOME/.cache/cargo}"
export RUSTUP_HOME="${RUSTUP_HOME:-$HOME/.cache/rustup}"

# Install dependencies
# Includes Tauri/WebKitGTK dependencies for Linux desktop apps
echo -e "${YELLOW}Installing dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y \
    curl \
    build-essential \
    gcc \
    make \
    cmake \
    pkg-config \
    libssl-dev \
    libwebkit2gtk-4.1-dev \
    libxdo-dev \
    libayatana-appindicator3-dev \
    librsvg2-dev

# Install rustup (Rust toolchain installer)
echo -e "${YELLOW}Installing rustup...${NC}"
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path

# Setup Rust environment
export PATH="$CARGO_HOME/bin:$PATH"

# Source cargo env
source "$CARGO_HOME/env"

# ─────────────────────────────────────────────────────────────────────────────
# Install cargo-binstall (for fast prebuilt binary downloads)
# ─────────────────────────────────────────────────────────────────────────────
echo -e "${YELLOW}Installing cargo-binstall...${NC}"
curl -L --proto '=https' --tlsv1.2 -sSf \
    https://raw.githubusercontent.com/cargo-bins/cargo-binstall/main/install-from-binstall-release.sh | bash

# Verify cargo-binstall installation
if ! command -v cargo-binstall &> /dev/null; then
    echo -e "${RED}✗ cargo-binstall installation failed${NC}"
    echo -e "${YELLOW}Falling back to cargo install for tools...${NC}"
    USE_BINSTALL=false
else
    echo -e "${GREEN}✓ cargo-binstall installed${NC}"
    USE_BINSTALL=true
fi

# Track failed tools for summary
FAILED_TOOLS=()

# Helper function: try binstall first (fast), fallback to cargo install (slow)
install_cargo_tool() {
    local tool=$1
    echo -e "${YELLOW}Installing ${tool}...${NC}"

    # Try binstall if available
    if [ "$USE_BINSTALL" = true ] && cargo binstall --no-confirm --locked "$tool"; then
        echo -e "${GREEN}✓ ${tool} installed (binary)${NC}"
        return 0
    fi

    # Fallback to cargo install
    if cargo install --locked "$tool"; then
        echo -e "${GREEN}✓ ${tool} installed (compiled)${NC}"
        return 0
    fi

    # Both methods failed
    echo -e "${RED}✗ ${tool} failed to install${NC}" >&2
    FAILED_TOOLS+=("$tool")
    return 1
}

RUST_VERSION=$(rustc --version)
CARGO_VERSION=$(cargo --version)
echo -e "${GREEN}✓ ${RUST_VERSION} installed${NC}"
echo -e "${GREEN}✓ ${CARGO_VERSION} installed${NC}"

# Install stable toolchain
echo -e "${YELLOW}Installing stable toolchain...${NC}"
rustup toolchain install stable
rustup default stable
echo -e "${GREEN}✓ Stable toolchain installed${NC}"

# Install essential components
echo -e "${YELLOW}Installing rustup components...${NC}"
rustup component add rust-analyzer clippy rustfmt
echo -e "${GREEN}✓ rust-analyzer installed${NC}"
echo -e "${GREEN}✓ clippy installed${NC}"
echo -e "${GREEN}✓ rustfmt installed${NC}"

# ─────────────────────────────────────────────────────────────────────────────
# Install WebAssembly Targets
# ─────────────────────────────────────────────────────────────────────────────
echo -e "${YELLOW}Installing WebAssembly targets...${NC}"

# Browser WASM target (most common)
rustup target add wasm32-unknown-unknown
echo -e "${GREEN}✓ wasm32-unknown-unknown (browser) installed${NC}"

# WASI Preview 1 (server-side WASM runtimes: Wasmtime, Wasmer)
rustup target add wasm32-wasip1
echo -e "${GREEN}✓ wasm32-wasip1 (WASI) installed${NC}"

# WASI Preview 2 (component model - newer)
rustup target add wasm32-wasip2 2>/dev/null && \
    echo -e "${GREEN}✓ wasm32-wasip2 (WASI P2) installed${NC}" || \
    echo -e "${YELLOW}⚠ wasm32-wasip2 not available on this toolchain${NC}"

# ─────────────────────────────────────────────────────────────────────────────
# Install Rust Development Tools (via binstall - prebuilt binaries)
# ─────────────────────────────────────────────────────────────────────────────
echo -e "${YELLOW}Installing Rust development tools...${NC}"

# List of cargo tools to install (binstall downloads prebuilt binaries when available)
CARGO_TOOLS=(
    "cargo-watch"       # Auto-rebuild on file changes
    "cargo-nextest"     # Fast test runner
    "cargo-audit"       # Security vulnerability scanner
    "cargo-expand"      # Macro expansion viewer
    "cargo-outdated"    # Dependency update checker
    "cargo-deny"        # Dependency security checker
    "cargo-tarpaulin"   # Code coverage tool
    "cargo-edit"        # Dependency management (add/remove)
    "tauri-cli"         # Tauri desktop app framework CLI
    "wasm-pack"         # WASM packaging and publishing
    "wasm-bindgen-cli"  # JS bindings generator for WASM
)

# Install tools - use || true to continue on failure (tracked in FAILED_TOOLS)
for tool in "${CARGO_TOOLS[@]}"; do
    install_cargo_tool "$tool" || true
done

# MCP server (may not have prebuilt binaries)
install_cargo_tool "rust-analyzer-mcp" || true

# Summary of tool installations
if [ ${#FAILED_TOOLS[@]} -gt 0 ]; then
    echo -e "${YELLOW}⚠ Some tools failed to install: ${FAILED_TOOLS[*]}${NC}"
else
    echo -e "${GREEN}✓ All Rust development tools installed successfully${NC}"
fi

# Setup shell integration
echo -e "${YELLOW}Configuring shell integration...${NC}"
CARGO_ENV_LINE='[[ -f "$HOME/.cache/cargo/env" ]] && source "$HOME/.cache/cargo/env"'
for rc_file in "$HOME/.bashrc" "$HOME/.zshrc"; do
    if [[ -f "$rc_file" ]] && ! grep -q "cargo/env" "$rc_file"; then
        echo "" >> "$rc_file"
        echo "# Rust/Cargo environment" >> "$rc_file"
        echo "$CARGO_ENV_LINE" >> "$rc_file"
    fi
done
echo -e "${GREEN}✓ Shell integration configured${NC}"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Rust environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - rustup (Rust toolchain manager)"
echo "  - ${RUST_VERSION}"
echo "  - ${CARGO_VERSION}"
echo ""
echo "Development tools:"
echo "  - cargo-binstall (fast binary installer)"
echo "  - rust-analyzer (LSP)"
echo "  - clippy (linter)"
echo "  - rustfmt (formatter)"
echo "  - cargo-watch (auto-rebuild)"
echo "  - cargo-nextest (test runner)"
echo "  - cargo-audit (security audit)"
echo "  - cargo-deny (dependency checker)"
echo "  - cargo-tarpaulin (code coverage)"
echo "  - cargo-edit (add/remove deps)"
echo "  - cargo-expand, cargo-outdated"
echo "  - rust-analyzer-mcp (MCP server)"
echo ""
echo "Desktop & WASM tools:"
echo "  - tauri-cli (desktop app framework)"
echo "  - wasm-pack (WASM packaging)"
echo "  - wasm-bindgen-cli (JS bindings)"
echo "  - wasm32-unknown-unknown (browser target)"
echo "  - wasm32-wasip1/wasip2 (WASI targets)"
echo ""
echo "Cache directories:"
echo "  - CARGO_HOME: $CARGO_HOME"
echo "  - RUSTUP_HOME: $RUSTUP_HOME"
echo ""
