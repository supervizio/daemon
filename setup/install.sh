#!/bin/sh
# supervizio install script
# Detects platform, package manager, and installs supervizio.
# Strategy: native package first, fallback to raw binary install.
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

    # Check for s6 (Alpine s6 overlay, Artix s6)
    if command -v s6-svc >/dev/null 2>&1 || [ -d /etc/s6 ]; then
        echo "s6"
        return
    fi

    # Check for dinit (Artix, Chimera Linux)
    if command -v dinitctl >/dev/null 2>&1; then
        echo "dinit"
        return
    fi

    # Check for runit (Void Linux, Artix)
    if command -v sv >/dev/null 2>&1 && [ -d /etc/sv ]; then
        echo "runit"
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

# Detect package manager
detect_package_manager() {
    OS="$1"

    case "$OS" in
        linux)
            if command -v dpkg >/dev/null 2>&1 && command -v apt-get >/dev/null 2>&1; then
                echo "deb"
            elif command -v dnf >/dev/null 2>&1; then
                echo "rpm-dnf"
            elif command -v zypper >/dev/null 2>&1; then
                echo "rpm-zypper"
            elif command -v apk >/dev/null 2>&1; then
                echo "apk"
            elif command -v pacman >/dev/null 2>&1; then
                echo "pacman"
            elif command -v xbps-install >/dev/null 2>&1; then
                echo "xbps"
            else
                echo "none"
            fi
            ;;
        freebsd)
            if command -v pkg >/dev/null 2>&1; then
                echo "freebsd-pkg"
            else
                echo "none"
            fi
            ;;
        openbsd)
            if command -v pkg_add >/dev/null 2>&1; then
                echo "openbsd-pkg"
            else
                echo "none"
            fi
            ;;
        netbsd)
            if command -v pkg_add >/dev/null 2>&1; then
                echo "netbsd-pkg"
            elif command -v pkgin >/dev/null 2>&1; then
                echo "netbsd-pkg"
            else
                echo "none"
            fi
            ;;
        *)
            echo "none"
            ;;
    esac
}

# Download a file using available tool
# Note: || return 1 prevents set -e from killing the script on download failure
download_file() {
    URL="$1"
    DEST="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$URL" -o "$DEST" || return 1
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$URL" -O "$DEST" || return 1
    elif command -v fetch >/dev/null 2>&1; then
        fetch -q "$URL" -o "$DEST" || return 1
    elif command -v ftp >/dev/null 2>&1; then
        ftp -o "$DEST" "$URL" || return 1
    else
        log_error "No download tool found (curl, wget, fetch, or ftp required)"
        return 1
    fi
}

# Get package file (local or download)
get_package() {
    PKG_MGR="$1"
    ARCH="$2"

    # E2E mode: use local package if SUPERVIZIO_LOCAL_PKG is set
    if [ -n "${SUPERVIZIO_LOCAL_PKG:-}" ] && [ -f "$SUPERVIZIO_LOCAL_PKG" ]; then
        log_info "Using local package: $SUPERVIZIO_LOCAL_PKG"
        # Preserve original extension (apk-tools 3.x requires .apk extension)
        PKG_EXT="${SUPERVIZIO_LOCAL_PKG##*.}"
        cp "$SUPERVIZIO_LOCAL_PKG" "/tmp/supervizio-pkg.${PKG_EXT}"
        return 0
    fi

    # Construct package filename based on package manager
    case "$PKG_MGR" in
        deb)          PKG_FILE="supervizio-${ARCH}.deb" ;;
        rpm-dnf|rpm-zypper) PKG_FILE="supervizio-${ARCH}.rpm" ;;
        apk)          PKG_FILE="supervizio-${ARCH}.apk" ;;
        pacman)       PKG_FILE="supervizio-${ARCH}.pkg.tar.zst" ;;
        xbps)         PKG_FILE="supervizio-${ARCH}.xbps" ;;
        freebsd-pkg)  PKG_FILE="supervizio-freebsd-${ARCH}.pkg" ;;
        openbsd-pkg)  PKG_FILE="supervizio-openbsd-${ARCH}.tgz" ;;
        netbsd-pkg)   PKG_FILE="supervizio-netbsd-${ARCH}.tgz" ;;
        *)            return 1 ;;
    esac

    # Construct download URL
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/supervizio/daemon/releases/latest/download/${PKG_FILE}"
    else
        DOWNLOAD_URL="https://github.com/supervizio/daemon/releases/download/${VERSION}/${PKG_FILE}"
    fi

    log_info "Downloading package from $DOWNLOAD_URL"
    if download_file "$DOWNLOAD_URL" "/tmp/supervizio-pkg.${PKG_FILE##*.}"; then
        return 0
    else
        log_warn "Package download failed"
        return 1
    fi
}

# Install package using native package manager
install_package() {
    PKG_MGR="$1"

    # Find the package file (with extension preserved)
    PKG_FILE=$(ls /tmp/supervizio-pkg.* 2>/dev/null | head -1)
    if [ -z "$PKG_FILE" ]; then
        log_error "Package file not found"
        return 1
    fi

    log_info "Installing via package manager: $PKG_MGR"

    case "$PKG_MGR" in
        deb)
            dpkg -i "$PKG_FILE" || true
            apt-get install -f -y
            ;;
        rpm-dnf)
            dnf install -y "$PKG_FILE"
            ;;
        rpm-zypper)
            zypper install -y --allow-unsigned-rpm "$PKG_FILE"
            ;;
        apk)
            apk add --allow-untrusted "$PKG_FILE"
            ;;
        pacman)
            pacman -U --noconfirm "$PKG_FILE"
            ;;
        xbps)
            # xbps-install requires a repository index; create local repo
            REPO_DIR=$(dirname "$PKG_FILE")
            if command -v xbps-rindex >/dev/null 2>&1; then
                xbps-rindex -a "$PKG_FILE" 2>/dev/null || true
                xbps-install -y --repository="$REPO_DIR" supervizio
            else
                log_warn "xbps-rindex not available"
                return 1
            fi
            ;;
        freebsd-pkg)
            pkg add "$PKG_FILE"
            ;;
        openbsd-pkg)
            pkg_add "$PKG_FILE"
            ;;
        netbsd-pkg)
            pkg_add "$PKG_FILE"
            ;;
        *)
            log_error "Unknown package manager: $PKG_MGR"
            return 1
            ;;
    esac
    INSTALL_RESULT=$?

    # Cleanup
    rm -f "$PKG_FILE"

    return "$INSTALL_RESULT"
}

# Get binary (local or download)
get_binary() {
    OS="$1"
    ARCH="$2"

    # E2E mode: use local binary if SUPERVIZIO_LOCAL_BIN is set or /bin-local/supervizio exists
    LOCAL_BIN="${SUPERVIZIO_LOCAL_BIN:-/bin-local/supervizio}"
    if [ -f "$LOCAL_BIN" ]; then
        log_info "Using local binary: $LOCAL_BIN"
        # Avoid copying file to itself (e.g., when LOCAL_BIN is /tmp/supervizio)
        DEST="/tmp/$BINARY_NAME"
        if [ "$(readlink -f "$LOCAL_BIN" 2>/dev/null || echo "$LOCAL_BIN")" != "$(readlink -f "$DEST" 2>/dev/null || echo "$DEST")" ]; then
            cp "$LOCAL_BIN" "$DEST"
        fi
        chmod +x "$DEST"
        return
    fi

    # Download from GitHub releases
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/supervizio/daemon/releases/latest/download/supervizio-${OS}-${ARCH}"
    else
        DOWNLOAD_URL="https://github.com/supervizio/daemon/releases/download/${VERSION}/supervizio-${OS}-${ARCH}"
    fi

    log_info "Downloading supervizio from $DOWNLOAD_URL"
    download_file "$DOWNLOAD_URL" "/tmp/$BINARY_NAME"
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
        freebsd)
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
                    cp "$SCRIPT_DIR/init/sysvinit/supervizio" /etc/init.d/
                    chmod +x /etc/init.d/supervizio
                    update-rc.d supervizio defaults 2>/dev/null || true
                    log_info "Service enabled. Start with: /etc/init.d/supervizio start"
                    ;;
                s6)
                    log_info "Installing s6 service"
                    mkdir -p /etc/s6/supervizio/log
                    cp "$SCRIPT_DIR/init/s6/supervizio/run" /etc/s6/supervizio/
                    cp "$SCRIPT_DIR/init/s6/supervizio/type" /etc/s6/supervizio/
                    chmod +x /etc/s6/supervizio/run
                    if [ -f "$SCRIPT_DIR/init/s6/supervizio/log/run" ]; then
                        cp "$SCRIPT_DIR/init/s6/supervizio/log/run" /etc/s6/supervizio/log/
                        chmod +x /etc/s6/supervizio/log/run
                        mkdir -p /var/log/supervizio
                    fi
                    log_info "Service installed. Enable with: s6-rc-bundle-update add default supervizio"
                    ;;
                dinit)
                    log_info "Installing dinit service"
                    mkdir -p /etc/dinit.d
                    cp "$SCRIPT_DIR/init/dinit/supervizio" /etc/dinit.d/
                    mkdir -p /var/log/supervizio
                    dinitctl enable supervizio 2>/dev/null || true
                    log_info "Service enabled. Start with: dinitctl start supervizio"
                    ;;
                runit)
                    log_info "Installing runit service"
                    mkdir -p /etc/sv/supervizio/log
                    cp "$SCRIPT_DIR/init/runit/supervizio/run" /etc/sv/supervizio/
                    chmod +x /etc/sv/supervizio/run
                    if [ -f "$SCRIPT_DIR/init/runit/supervizio/log/run" ]; then
                        cp "$SCRIPT_DIR/init/runit/supervizio/log/run" /etc/sv/supervizio/log/
                        chmod +x /etc/sv/supervizio/log/run
                        mkdir -p /var/log/supervizio
                    fi
                    # Enable service by creating symlink
                    ln -sf /etc/sv/supervizio /var/service/
                    log_info "Service enabled. Start with: sv start supervizio"
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
            # Ensure rc.conf exists and add enable line if not present
            touch /etc/rc.conf
            if ! grep -q "^supervizio=" /etc/rc.conf; then
                echo "supervizio=YES" >> /etc/rc.conf
            fi
            log_info "Service enabled. Start with: /etc/rc.d/supervizio start"
            ;;
        darwin)
            log_info "Installing launchd service"
            cp "$SCRIPT_DIR/init/launchd/io.superviz.daemon.plist" /Library/LaunchDaemons/
            log_info "Service installed. Start with: launchctl load /Library/LaunchDaemons/io.superviz.daemon.plist"
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

    # Detect package manager
    PKG_MGR=$(detect_package_manager "$OS")
    log_info "Package manager: $PKG_MGR"

    # Strategy: try native package first, fallback to raw binary
    if [ "$PKG_MGR" != "none" ]; then
        if get_package "$PKG_MGR" "$ARCH"; then
            if install_package "$PKG_MGR"; then
                # Verify the binary was actually installed (not just metadata)
                if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
                    log_info "=== Installation complete (via package) ==="
                    "$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null && \
                        log_info "Version: $("$INSTALL_DIR/$BINARY_NAME" --version 2>&1)"
                    return 0
                fi
                log_warn "Package installed but binary missing, falling back to manual install"
            fi
            log_warn "Package install failed, falling back to manual install"
        else
            log_warn "Package unavailable, falling back to manual install"
        fi
    fi

    # Fallback: raw binary install
    log_info "Installing via raw binary"
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
