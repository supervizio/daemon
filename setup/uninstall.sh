#!/bin/sh
# supervizio uninstall script
# Removes supervizio and its init system configuration.
# Supports package-aware uninstall and non-interactive mode.
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
    elif command -v s6-svc >/dev/null 2>&1 || [ -d /etc/s6 ]; then
        echo "s6"
    elif command -v dinitctl >/dev/null 2>&1; then
        echo "dinit"
    elif command -v sv >/dev/null 2>&1 && [ -d /etc/sv ]; then
        echo "runit"
    elif [ -d /etc/init.d ]; then
        echo "sysvinit"
    else
        echo "unknown"
    fi
}

# Detect how supervizio was installed (package or manual)
detect_install_method() {
    # Check Linux package managers
    if command -v dpkg >/dev/null 2>&1 && dpkg -l supervizio >/dev/null 2>&1; then
        echo "deb"
        return
    fi
    if command -v rpm >/dev/null 2>&1 && rpm -q supervizio >/dev/null 2>&1; then
        echo "rpm"
        return
    fi
    if command -v apk >/dev/null 2>&1 && apk info -e supervizio >/dev/null 2>&1; then
        echo "apk"
        return
    fi
    if command -v pacman >/dev/null 2>&1 && pacman -Q supervizio >/dev/null 2>&1; then
        echo "pacman"
        return
    fi
    if command -v xbps-query >/dev/null 2>&1 && xbps-query supervizio >/dev/null 2>&1; then
        echo "xbps"
        return
    fi
    # Check BSD package managers
    OS="$(detect_os)"
    if [ "$OS" = "freebsd" ] && command -v pkg >/dev/null 2>&1 && pkg info supervizio >/dev/null 2>&1; then
        echo "freebsd-pkg"
        return
    fi
    if [ "$OS" = "openbsd" ] && command -v pkg_info >/dev/null 2>&1 && pkg_info supervizio >/dev/null 2>&1; then
        echo "openbsd-pkg"
        return
    fi
    if [ "$OS" = "netbsd" ] && command -v pkg_info >/dev/null 2>&1 && pkg_info supervizio >/dev/null 2>&1; then
        echo "netbsd-pkg"
        return
    fi
    echo "manual"
}

# Uninstall via package manager
uninstall_package() {
    METHOD="$1"
    log_info "Uninstalling via package manager: $METHOD"

    case "$METHOD" in
        deb)
            apt-get remove -y supervizio || dpkg --remove supervizio
            ;;
        rpm)
            if command -v dnf >/dev/null 2>&1; then
                dnf remove -y supervizio
            elif command -v zypper >/dev/null 2>&1; then
                zypper remove -y supervizio
            else
                rpm -e supervizio
            fi
            ;;
        apk)
            apk del supervizio
            ;;
        pacman)
            pacman -R --noconfirm supervizio
            ;;
        xbps)
            xbps-remove -y supervizio
            ;;
        freebsd-pkg)
            pkg delete -y supervizio
            ;;
        openbsd-pkg)
            pkg_delete supervizio
            ;;
        netbsd-pkg)
            pkg_delete supervizio
            ;;
    esac
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
                    rm -f /usr/lib/systemd/system/supervizio.service
                    systemctl daemon-reload
                    ;;
                openrc)
                    log_info "Stopping and removing OpenRC service"
                    rc-service supervizio stop 2>/dev/null || true
                    rc-update del supervizio default 2>/dev/null || true
                    rm -f /etc/init.d/supervizio
                    ;;
                s6)
                    log_info "Stopping and removing s6 service"
                    s6-svc -d /etc/s6/supervizio 2>/dev/null || true
                    rm -rf /etc/s6/supervizio
                    ;;
                dinit)
                    log_info "Stopping and removing dinit service"
                    dinitctl stop supervizio 2>/dev/null || true
                    dinitctl disable supervizio 2>/dev/null || true
                    rm -f /etc/dinit.d/supervizio
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

    # Detect install method
    INSTALL_METHOD=$(detect_install_method)
    log_info "Install method: $INSTALL_METHOD"

    # If installed via package manager, use it for uninstall
    if [ "$INSTALL_METHOD" != "manual" ]; then
        # Stop service first (package scripts may not handle all init systems)
        remove_service "$OS" "$INIT"
        uninstall_package "$INSTALL_METHOD"
    else
        # Manual uninstall: remove service first, then binary
        remove_service "$OS" "$INIT"
        remove_binary
    fi

    # Handle config directory
    CONFIG_DIR="/etc/supervizio"
    if [ "$OS" = "freebsd" ]; then
        CONFIG_DIR="/usr/local/etc/supervizio"
    fi

    if [ -d "$CONFIG_DIR" ]; then
        if [ "${SUPERVIZIO_NON_INTERACTIVE:-}" = "true" ]; then
            rm -rf "$CONFIG_DIR"
            log_info "Configuration removed"
        else
            printf "Remove configuration directory %s? [y/N] " "$CONFIG_DIR"
            read -r REPLY
            if [ "$REPLY" = "y" ] || [ "$REPLY" = "Y" ]; then
                rm -rf "$CONFIG_DIR"
                log_info "Configuration removed"
            else
                log_info "Configuration kept at $CONFIG_DIR"
            fi
        fi
    fi

    # Handle logs
    if [ -d /var/log/supervizio ]; then
        if [ "${SUPERVIZIO_NON_INTERACTIVE:-}" = "true" ]; then
            rm -rf /var/log/supervizio
            log_info "Logs removed"
        else
            printf "Remove logs at /var/log/supervizio? [y/N] "
            read -r REPLY
            if [ "$REPLY" = "y" ] || [ "$REPLY" = "Y" ]; then
                rm -rf /var/log/supervizio
                log_info "Logs removed"
            else
                log_info "Logs kept at /var/log/supervizio"
            fi
        fi
    fi

    log_info "=== Uninstallation complete ==="
}

main "$@"
