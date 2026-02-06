#!/bin/sh
# supervizio uninstall script
# Removes supervizio and its init system configuration
set -e

BINARY_NAME="supervizio"
INSTALL_DIR="/usr/local/bin"

# Colors
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
        *)       echo "unknown" ;;
    esac
}

# Detect init system
detect_init_system() {
    if [ "$(detect_os)" != "linux" ]; then
        echo "none"
        return
    fi

    if command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
        echo "systemd"
    elif command -v rc-service >/dev/null 2>&1; then
        echo "openrc"
    elif command -v sv >/dev/null 2>&1 && [ -d /etc/sv ]; then
        echo "runit"
    elif [ -d /etc/init.d ]; then
        echo "sysvinit"
    else
        echo "unknown"
    fi
}

# Stop and remove service
remove_service() {
    OS="$1"
    INIT="$2"

    case "$OS" in
        linux)
            case "$INIT" in
                systemd)
                    log_info "Stopping and removing systemd service"
                    systemctl stop supervizio 2>/dev/null || true
                    systemctl disable supervizio 2>/dev/null || true
                    rm -f /etc/systemd/system/supervizio.service
                    systemctl daemon-reload
                    ;;
                openrc)
                    log_info "Stopping and removing OpenRC service"
                    rc-service supervizio stop 2>/dev/null || true
                    rc-update del supervizio default 2>/dev/null || true
                    rm -f /etc/init.d/supervizio
                    ;;
                sysvinit)
                    log_info "Stopping and removing SysVinit service"
                    /etc/init.d/supervizio stop 2>/dev/null || true
                    update-rc.d -f supervizio remove 2>/dev/null || true
                    rm -f /etc/init.d/supervizio
                    ;;
                runit)
                    log_info "Stopping and removing runit service"
                    sv stop supervizio 2>/dev/null || true
                    rm -f /var/service/supervizio
                    rm -rf /etc/sv/supervizio
                    ;;
            esac
            ;;
        freebsd)
            log_info "Stopping and removing rc.d service"
            service supervizio stop 2>/dev/null || true
            sysrc -x supervizio_enable 2>/dev/null || true
            rm -f /usr/local/etc/rc.d/supervizio
            ;;
        openbsd)
            log_info "Stopping and removing rc.d service"
            rcctl stop supervizio 2>/dev/null || true
            rcctl disable supervizio 2>/dev/null || true
            rm -f /etc/rc.d/supervizio
            ;;
        netbsd)
            log_info "Stopping and removing rc.d service"
            /etc/rc.d/supervizio stop 2>/dev/null || true
            rm -f /etc/rc.d/supervizio
            # Remove from rc.conf
            sed -i '/^supervizio=/d' /etc/rc.conf 2>/dev/null || true
            ;;
        darwin)
            log_info "Unloading and removing launchd service"
            launchctl unload /Library/LaunchDaemons/io.superviz.daemon.plist 2>/dev/null || true
            rm -f /Library/LaunchDaemons/io.superviz.daemon.plist
            ;;
    esac
}

# Remove binary
remove_binary() {
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        log_info "Removing binary $INSTALL_DIR/$BINARY_NAME"
        rm -f "$INSTALL_DIR/$BINARY_NAME"
    else
        log_warn "Binary not found at $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Main
main() {
    log_info "=== supervizio uninstallation ==="

    # Check root
    if [ "$(id -u)" -ne 0 ]; then
        log_error "This script must be run as root"
        exit 1
    fi

    OS=$(detect_os)
    INIT=$(detect_init_system)

    log_info "Detected: OS=$OS INIT=$INIT"

    # Remove service first
    remove_service "$OS" "$INIT"

    # Remove binary
    remove_binary

    # Ask about config
    CONFIG_DIR="/etc/supervizio"
    if [ "$OS" = "freebsd" ]; then
        CONFIG_DIR="/usr/local/etc/supervizio"
    fi

    if [ -d "$CONFIG_DIR" ]; then
        printf "Remove configuration directory %s? [y/N] " "$CONFIG_DIR"
        read -r REPLY
        if [ "$REPLY" = "y" ] || [ "$REPLY" = "Y" ]; then
            rm -rf "$CONFIG_DIR"
            log_info "Configuration removed"
        else
            log_info "Configuration kept at $CONFIG_DIR"
        fi
    fi

    # Ask about logs
    if [ -d /var/log/supervizio ]; then
        printf "Remove logs at /var/log/supervizio? [y/N] "
        read -r REPLY
        if [ "$REPLY" = "y" ] || [ "$REPLY" = "Y" ]; then
            rm -rf /var/log/supervizio
            log_info "Logs removed"
        else
            log_info "Logs kept at /var/log/supervizio"
        fi
    fi

    log_info "=== Uninstallation complete ==="
}

main "$@"
