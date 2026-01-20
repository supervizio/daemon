#!/bin/sh
# supervizio install script
# Detects platform and installs supervizio with appropriate init system
set -e

VERSION="${SUPERVIZIO_VERSION:-latest}"
BINARY_NAME="supervizio"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/supervizio"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Colors (if terminal supports it)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

log_info() { printf "${GREEN}[INFO]${NC} %s\n" "$1"; }
log_warn() { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }
log_error() { printf "${RED}[ERROR]${NC} %s\n" "$1"; }

# Detect OS
detect_os() {
    OS="$(uname -s)"
    case "$OS" in
        Linux)   echo "linux" ;;
        Darwin)  echo "darwin" ;;
        FreeBSD) echo "freebsd" ;;
        OpenBSD) echo "openbsd" ;;
        NetBSD)  echo "netbsd" ;;
        DragonFly) echo "dragonfly" ;;
        *)       log_error "Unsupported OS: $OS"; exit 1 ;;
    esac
}

# Detect architecture
detect_arch() {
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        armv7l|armv7)  echo "arm" ;;
        i386|i686)     echo "386" ;;
        riscv64)       echo "riscv64" ;;
        *)             log_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
}

# Detect init system (Linux only)
detect_init_system() {
    if [ "$(detect_os)" != "linux" ]; then
        echo "none"
        return
    fi

    # Check for systemd
    if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
        echo "systemd"
        return
    fi

    # Check for OpenRC (Alpine, Gentoo)
    if command -v rc-service >/dev/null 2>&1; then
        echo "openrc"
        return
    fi

    # Check for SysVinit
    if [ -d /etc/init.d ] && [ ! -d /run/systemd/system ]; then
        echo "sysvinit"
        return
    fi

    echo "unknown"
}

# Detect Linux distribution
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    elif [ -f /etc/alpine-release ]; then
        echo "alpine"
    elif [ -f /etc/debian_version ]; then
        echo "debian"
    elif [ -f /etc/redhat-release ]; then
        echo "rhel"
    else
        echo "unknown"
    fi
}

# Get binary (local or download)
get_binary() {
    OS="$1"
    ARCH="$2"

    # E2E mode: use local binary if SUPERVIZIO_LOCAL_BIN is set or /bin-local/supervizio exists
    LOCAL_BIN="${SUPERVIZIO_LOCAL_BIN:-/bin-local/supervizio}"
    if [ -f "$LOCAL_BIN" ]; then
        log_info "Using local binary: $LOCAL_BIN"
        cp "$LOCAL_BIN" "/tmp/$BINARY_NAME"
        chmod +x "/tmp/$BINARY_NAME"
        return
    fi

    # Download from GitHub releases
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/supervizio/daemon/releases/latest/download/supervizio-${OS}-${ARCH}"
    else
        DOWNLOAD_URL="https://github.com/supervizio/daemon/releases/download/${VERSION}/supervizio-${OS}-${ARCH}"
    fi

    log_info "Downloading supervizio from $DOWNLOAD_URL"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$DOWNLOAD_URL" -o "/tmp/$BINARY_NAME"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$DOWNLOAD_URL" -O "/tmp/$BINARY_NAME"
    elif command -v fetch >/dev/null 2>&1; then
        fetch -q "$DOWNLOAD_URL" -o "/tmp/$BINARY_NAME"
    else
        log_error "No download tool found (curl, wget, or fetch required)"
        exit 1
    fi

    chmod +x "/tmp/$BINARY_NAME"
}

# Install binary
install_binary() {
    log_info "Installing binary to $INSTALL_DIR/$BINARY_NAME"

    # Create install directory if needed
    mkdir -p "$INSTALL_DIR"

    # Move binary
    mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod 755 "$INSTALL_DIR/$BINARY_NAME"
}

# Create config directory
create_config_dir() {
    OS="$1"

    case "$OS" in
        freebsd|dragonfly)
            CONFIG_DIR="/usr/local/etc/supervizio"
            ;;
    esac

    log_info "Creating config directory $CONFIG_DIR"
    mkdir -p "$CONFIG_DIR"

    # Create default config if not exists
    if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
        cat > "$CONFIG_DIR/config.yaml" << 'EOF'
version: "1"

logging:
  base_dir: /var/log/supervizio
  defaults:
    timestamp_format: iso8601
    rotation:
      max_size: 100MB
      max_age: 7d
      max_files: 10
      compress: true

services: []
EOF
        log_info "Created default config at $CONFIG_DIR/config.yaml"
    fi
}

# Install init script/service
install_init() {
    OS="$1"
    INIT="$2"

    case "$OS" in
        linux)
            case "$INIT" in
                systemd)
                    log_info "Installing systemd service"
                    cp "$SCRIPT_DIR/init/systemd/supervizio.service" /etc/systemd/system/
                    systemctl daemon-reload
                    systemctl enable supervizio
                    log_info "Service enabled. Start with: systemctl start supervizio"
                    ;;
                openrc)
                    log_info "Installing OpenRC service"
                    cp "$SCRIPT_DIR/init/openrc/supervizio" /etc/init.d/
                    chmod +x /etc/init.d/supervizio
                    rc-update add supervizio default
                    log_info "Service enabled. Start with: rc-service supervizio start"
                    ;;
                sysvinit)
                    log_info "Installing SysVinit service"
                    cp "$SCRIPT_DIR/init/openrc/supervizio" /etc/init.d/
                    chmod +x /etc/init.d/supervizio
                    update-rc.d supervizio defaults 2>/dev/null || true
                    log_info "Service enabled. Start with: /etc/init.d/supervizio start"
                    ;;
                *)
                    log_warn "Unknown init system, skipping service installation"
                    ;;
            esac
            ;;
        freebsd)
            log_info "Installing FreeBSD rc.d service"
            cp "$SCRIPT_DIR/init/freebsd/supervizio" /usr/local/etc/rc.d/
            chmod +x /usr/local/etc/rc.d/supervizio
            sysrc supervizio_enable="YES"
            log_info "Service enabled. Start with: service supervizio start"
            ;;
        openbsd)
            log_info "Installing OpenBSD rc.d service"
            cp "$SCRIPT_DIR/init/openbsd/supervizio" /etc/rc.d/
            chmod +x /etc/rc.d/supervizio
            rcctl enable supervizio
            log_info "Service enabled. Start with: rcctl start supervizio"
            ;;
        netbsd)
            log_info "Installing NetBSD rc.d service"
            cp "$SCRIPT_DIR/init/netbsd/supervizio" /etc/rc.d/
            chmod +x /etc/rc.d/supervizio
            echo "supervizio=YES" >> /etc/rc.conf
            log_info "Service enabled. Start with: /etc/rc.d/supervizio start"
            ;;
        darwin)
            log_info "Installing launchd service"
            cp "$SCRIPT_DIR/init/launchd/io.superviz.daemon.plist" /Library/LaunchDaemons/
            log_info "Service installed. Start with: launchctl load /Library/LaunchDaemons/io.superviz.daemon.plist"
            ;;
        dragonfly)
            log_info "Installing DragonFlyBSD rc.d service"
            cp "$SCRIPT_DIR/init/freebsd/supervizio" /usr/local/etc/rc.d/
            chmod +x /usr/local/etc/rc.d/supervizio
            echo 'supervizio_enable="YES"' >> /etc/rc.conf
            log_info "Service enabled. Start with: service supervizio start"
            ;;
    esac
}

# Main installation
main() {
    log_info "=== supervizio installation ==="

    # Check root
    if [ "$(id -u)" -ne 0 ]; then
        log_error "This script must be run as root"
        exit 1
    fi

    # Detect platform
    OS=$(detect_os)
    ARCH=$(detect_arch)
    INIT=$(detect_init_system)

    log_info "Detected: OS=$OS ARCH=$ARCH INIT=$INIT"

    if [ "$OS" = "linux" ]; then
        DISTRO=$(detect_distro)
        log_info "Distribution: $DISTRO"
    fi

    # Get and install binary (local or download)
    get_binary "$OS" "$ARCH"
    install_binary

    # Create config
    create_config_dir "$OS"

    # Install init scripts
    install_init "$OS" "$INIT"

    # Verify installation
    if "$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null; then
        log_info "=== Installation complete ==="
    else
        log_warn "Binary installed but --version check failed"
        log_info "=== Installation complete (verify manually) ==="
    fi
}

main "$@"
