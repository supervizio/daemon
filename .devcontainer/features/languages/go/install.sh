#!/bin/bash
set -e

echo "========================================="
echo "Installing Go Development Environment"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Environment variables
export GO_VERSION="${GO_VERSION:-latest}"
export GOROOT="${GOROOT:-/usr/local/go}"
export GOPATH="${GOPATH:-/home/vscode/.cache/go}"
export GOCACHE="${GOCACHE:-/home/vscode/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-/home/vscode/.cache/go/pkg/mod}"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"

# Install dependencies
# Includes Wails/WebKitGTK dependencies for Linux desktop apps
echo -e "${YELLOW}Installing dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y \
    curl \
    git \
    make \
    gcc \
    build-essential \
    pkg-config \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
    libglib2.0-dev

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        GO_ARCH="amd64"
        ;;
    aarch64|arm64)
        GO_ARCH="arm64"
        ;;
    armv7l)
        GO_ARCH="armv6l"
        ;;
    *)
        echo -e "${RED}✗ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Get latest Go version if requested
if [ "$GO_VERSION" = "latest" ]; then
    echo -e "${YELLOW}Fetching latest Go version...${NC}"
    GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1 | sed 's/go//')
fi

# Download and install Go
echo -e "${YELLOW}Installing Go ${GO_VERSION} for ${GO_ARCH}...${NC}"
GO_TARBALL="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
GO_URL="https://dl.google.com/go/${GO_TARBALL}"

# Download Go tarball
curl -fsSL --retry 3 --retry-delay 5 -o "/tmp/${GO_TARBALL}" "$GO_URL"

# Remove any existing Go installation
if [ -d "$GOROOT" ]; then
    echo -e "${YELLOW}Removing existing Go installation...${NC}"
    sudo rm -rf "$GOROOT"
fi

# Extract to /usr/local
sudo tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"

# Clean up
rm "/tmp/${GO_TARBALL}"

GO_INSTALLED=$(go version)
echo -e "${GREEN}✓ ${GO_INSTALLED} installed${NC}"

# Create necessary directories
mkdir -p "$GOPATH/bin"
mkdir -p "$GOPATH/pkg"
mkdir -p "$GOPATH/src"
mkdir -p "$GOCACHE"
mkdir -p "$GOMODCACHE"

# ─────────────────────────────────────────────────────────────────────────────
# Install Go Development Tools (prebuilt binaries from GitHub Releases)
# ─────────────────────────────────────────────────────────────────────────────
echo -e "${YELLOW}Installing Go development tools...${NC}"
mkdir -p "$HOME/.local/bin"

# Helper function: download from GitHub Releases, fallback to go install
install_go_tool() {
    local name=$1
    local url=$2
    local go_pkg=$3
    local extract_type=$4  # "tar.gz", "zip", or "binary"

    echo -e "${YELLOW}Installing ${name}...${NC}"

    local tmp_file="/tmp/${name}-download"

    if curl -fsSL --connect-timeout 10 --max-time 120 "$url" -o "$tmp_file" 2>/dev/null; then
        case "$extract_type" in
            tar.gz)
                tar -xzf "$tmp_file" -C "$GOPATH/bin/" "$name" 2>/dev/null || \
                tar -xzf "$tmp_file" --wildcards --strip-components=1 -C "$GOPATH/bin/" "*/$name" 2>/dev/null || \
                tar -xzf "$tmp_file" -C "/tmp/" && mv "/tmp/$name" "$GOPATH/bin/" 2>/dev/null
                ;;
            zip)
                unzip -o "$tmp_file" "$name" -d "$GOPATH/bin/" 2>/dev/null || \
                unzip -o "$tmp_file" -d "/tmp/${name}-extracted" && mv "/tmp/${name}-extracted/$name" "$GOPATH/bin/" 2>/dev/null
                ;;
            binary)
                mv "$tmp_file" "$GOPATH/bin/$name"
                ;;
        esac
        chmod +x "$GOPATH/bin/$name"
        rm -f "$tmp_file" "/tmp/${name}-extracted" 2>/dev/null
        echo -e "${GREEN}✓ ${name} installed (binary)${NC}"
    elif [ -n "$go_pkg" ]; then
        echo -e "${YELLOW}  Fallback to go install...${NC}"
        if go install "${go_pkg}@latest" 2>/dev/null; then
            echo -e "${GREEN}✓ ${name} installed (compiled)${NC}"
        else
            echo -e "${YELLOW}⚠ ${name} failed to install${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ ${name} download failed${NC}"
    fi
}

# Helper function: fetch latest version from GitHub API with fallback
# Protects against rate limiting (60 req/h unauthenticated)
get_github_version() {
    local repo=$1
    local fallback=$2
    local version
    # Use portable sed instead of grep -oP (PCRE not available on all systems)
    version=$(curl -s --connect-timeout 5 --max-time 10 \
        "https://api.github.com/repos/${repo}/releases/latest" 2>/dev/null \
        | sed -n 's/.*"tag_name": *"v\?\([^"]*\)".*/\1/p' | head -n 1)
    echo "${version:-$fallback}"
}

# Quality & Linting - golangci-lint (prebuilt)
# Fallback 1.63.4 fixes CVE-2024-45337 (golang.org/x/crypto) and CVE-2024-45338 (golang.org/x/net)
GOLANGCI_VERSION=$(get_github_version "golangci/golangci-lint" "1.63.4")
install_go_tool "golangci-lint" \
    "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_VERSION}/golangci-lint-${GOLANGCI_VERSION}-linux-${GO_ARCH}.tar.gz" \
    "github.com/golangci/golangci-lint/cmd/golangci-lint" \
    "tar.gz"

# Security - gosec (prebuilt)
GOSEC_VERSION=$(get_github_version "securego/gosec" "2.22.11")
install_go_tool "gosec" \
    "https://github.com/securego/gosec/releases/download/v${GOSEC_VERSION}/gosec_${GOSEC_VERSION}_linux_${GO_ARCH}.tar.gz" \
    "github.com/securego/gosec/v2/cmd/gosec" \
    "tar.gz"

# Formatting - gofumpt (prebuilt binary, no archive)
GOFUMPT_VERSION=$(get_github_version "mvdan/gofumpt" "0.9.2")
# gofumpt uses 'v' prefix in URLs
[[ "$GOFUMPT_VERSION" != v* ]] && GOFUMPT_VERSION="v${GOFUMPT_VERSION}"
install_go_tool "gofumpt" \
    "https://github.com/mvdan/gofumpt/releases/download/${GOFUMPT_VERSION}/gofumpt_${GOFUMPT_VERSION}_linux_${GO_ARCH}" \
    "mvdan.cc/gofumpt" \
    "binary"

# Testing tools - gotestsum (prebuilt)
GOTESTSUM_VERSION=$(get_github_version "gotestyourself/gotestsum" "1.13.0")
install_go_tool "gotestsum" \
    "https://github.com/gotestyourself/gotestsum/releases/download/v${GOTESTSUM_VERSION}/gotestsum_${GOTESTSUM_VERSION}_linux_${GO_ARCH}.tar.gz" \
    "gotest.tools/gotestsum" \
    "tar.gz"

# Import management - goimports (no prebuilt, go install only)
echo -e "${YELLOW}Installing goimports...${NC}"
go install golang.org/x/tools/cmd/goimports@latest
echo -e "${GREEN}✓ goimports installed${NC}"

# Optional: ktn-linter (prebuilt)
install_go_tool "ktn-linter" \
    "https://github.com/kodflow/ktn-linter/releases/latest/download/ktn-linter-linux-${GO_ARCH}" \
    "" \
    "binary"

# ─────────────────────────────────────────────────────────────────────────────
# Install Wails v2 (Desktop GUI Framework)
# ─────────────────────────────────────────────────────────────────────────────
echo -e "${YELLOW}Installing Wails v2 (desktop GUI framework)...${NC}"
if go install github.com/wailsapp/wails/v2/cmd/wails@latest; then
    if command -v wails &> /dev/null; then
        WAILS_VERSION=$(wails version 2>/dev/null | head -n 1 || echo "installed")
        echo -e "${GREEN}✓ Wails ${WAILS_VERSION}${NC}"
    else
        echo -e "${YELLOW}⚠ Wails installed but not in PATH yet${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Wails installation failed${NC}"
fi

# ─────────────────────────────────────────────────────────────────────────────
# Install TinyGo (WebAssembly Compiler)
# ─────────────────────────────────────────────────────────────────────────────
echo -e "${YELLOW}Installing TinyGo (WASM/embedded compiler)...${NC}"

# Get latest TinyGo version
TINYGO_VERSION=$(get_github_version "tinygo-org/tinygo" "0.34.0")

# Download and install TinyGo
case "$GO_ARCH" in
    amd64)
        TINYGO_PKG="tinygo_${TINYGO_VERSION}_amd64.deb"
        ;;
    arm64)
        TINYGO_PKG="tinygo_${TINYGO_VERSION}_arm64.deb"
        ;;
    *)
        echo -e "${YELLOW}⚠ TinyGo not available for ${GO_ARCH}${NC}"
        TINYGO_PKG=""
        ;;
esac

if [ -n "$TINYGO_PKG" ]; then
    TINYGO_URL="https://github.com/tinygo-org/tinygo/releases/download/v${TINYGO_VERSION}/${TINYGO_PKG}"
    if curl -fsSL --connect-timeout 10 --max-time 120 -o "/tmp/${TINYGO_PKG}" "$TINYGO_URL" 2>/dev/null; then
        sudo dpkg -i "/tmp/${TINYGO_PKG}" 2>/dev/null || sudo apt-get install -f -y
        rm -f "/tmp/${TINYGO_PKG}"
        if command -v tinygo &> /dev/null; then
            TINYGO_INSTALLED=$(tinygo version 2>/dev/null | head -n 1)
            echo -e "${GREEN}✓ ${TINYGO_INSTALLED}${NC}"
        else
            echo -e "${GREEN}✓ TinyGo ${TINYGO_VERSION} installed${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ TinyGo download failed, skipping${NC}"
    fi
fi

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Go environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - ${GO_INSTALLED}"
echo "  - Go Modules (package manager)"
echo ""
echo "Development tools:"
echo "  - golangci-lint (meta-linter)"
echo "  - gosec (security scanner)"
echo "  - gofumpt (formatter)"
echo "  - goimports (import manager)"
echo "  - gotestsum (test runner)"
echo "  - ktn-linter (custom linter)"
echo ""
echo "Desktop & WASM tools:"
echo "  - wails (desktop GUI framework)"
echo "  - tinygo (WASM/embedded compiler)"
echo ""
echo "Cache directories:"
echo "  - GOPATH: $GOPATH"
echo "  - GOCACHE: $GOCACHE"
echo "  - GOMODCACHE: $GOMODCACHE"
echo ""
