#!/bin/bash
# ============================================================================
# postStart.sh - Runs EVERY TIME the container starts
# ============================================================================
# This script runs after postCreateCommand and before postAttachCommand.
# Runs each time the container is successfully started.
# Use it for: MCP setup, services startup, recurring initialization.
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../shared/utils.sh"

log_info "postStart: Container starting..."

# ============================================================================
# Restore Claude commands/scripts from image defaults
# ============================================================================
# Volume mounts overwrite image content, so we restore from /etc/claude-defaults/
CLAUDE_DEFAULTS="/etc/claude-defaults"

if [ -d "$CLAUDE_DEFAULTS" ]; then
    log_info "Restoring Claude configuration from image defaults..."

    # Ensure base directory exists
    mkdir -p "$HOME/.claude"

    # CLEAN commands, scripts and agents to avoid legacy pollution
    # Only these directories are managed by the image - sessions/plans are user data
    rm -rf "$HOME/.claude/commands" "$HOME/.claude/scripts" "$HOME/.claude/agents"

    # Restore commands (fresh copy from image)
    if [ -d "$CLAUDE_DEFAULTS/commands" ]; then
        mkdir -p "$HOME/.claude/commands"
        cp -r "$CLAUDE_DEFAULTS/commands/"* "$HOME/.claude/commands/" 2>/dev/null || true
    fi

    # Restore scripts (fresh copy from image)
    if [ -d "$CLAUDE_DEFAULTS/scripts" ]; then
        mkdir -p "$HOME/.claude/scripts"
        cp -r "$CLAUDE_DEFAULTS/scripts/"* "$HOME/.claude/scripts/" 2>/dev/null || true
        chmod -R 755 "$HOME/.claude/scripts/"
    fi

    # Restore agents (fresh copy from image)
    if [ -d "$CLAUDE_DEFAULTS/agents" ]; then
        mkdir -p "$HOME/.claude/agents"
        cp -r "$CLAUDE_DEFAULTS/agents/"* "$HOME/.claude/agents/" 2>/dev/null || true
        chmod -R 755 "$HOME/.claude/agents/"
    fi

    # Restore settings.json only if it does not exist (user customizations preserved)
    if [ -f "$CLAUDE_DEFAULTS/settings.json" ] && [ ! -f "$HOME/.claude/settings.json" ]; then
        cp "$CLAUDE_DEFAULTS/settings.json" "$HOME/.claude/settings.json"
    fi

    log_success "Claude configuration restored (clean)"
fi

# ============================================================================
# Ensure Claude directories exist (volume mount point)
# ============================================================================
mkdir -p "$HOME/.claude/sessions" "$HOME/.claude/plans"
log_success "Claude directories initialized"

# ============================================================================
# GNOME Keyring Setup (for credential storage - libsecret/Secret Service API)
# ============================================================================
# Required by: CodeRabbit CLI, GitHub CLI, VS Code, Claude Code
# Works on all platforms: Mac, Windows, Linux, WSL (container is always Linux)
setup_gnome_keyring() {
    # Check if gnome-keyring-daemon is available
    if ! command -v gnome-keyring-daemon &> /dev/null; then
        log_warning "gnome-keyring-daemon not found - credential storage may fail"
        return 1
    fi

    # Check if already running
    if pgrep -u "$(id -u)" gnome-keyring-daemon &> /dev/null; then
        log_info "gnome-keyring-daemon already running"
        return 0
    fi

    # Start D-Bus session if not available
    if [ -z "${DBUS_SESSION_BUS_ADDRESS:-}" ]; then
        log_info "Starting D-Bus session bus..."
        if command -v dbus-launch &> /dev/null; then
            eval "$(dbus-launch --sh-syntax)"
            export DBUS_SESSION_BUS_ADDRESS
        else
            log_warning "dbus-launch not found - using fallback"
            export DBUS_SESSION_BUS_ADDRESS="unix:path=/run/user/$(id -u)/bus"
        fi
    fi

    # Start gnome-keyring-daemon with secrets component
    log_info "Starting gnome-keyring-daemon..."
    # Use --unlock with empty password for headless operation
    eval "$(echo '' | gnome-keyring-daemon --unlock --components=secrets 2>/dev/null)" || {
        log_warning "gnome-keyring-daemon failed to start with unlock, trying without..."
        eval "$(gnome-keyring-daemon --start --components=secrets 2>/dev/null)" || {
            log_warning "gnome-keyring-daemon failed to start"
            return 1
        }
    }

    log_success "gnome-keyring-daemon started successfully"
    return 0
}

# Run keyring setup and export env vars for shell sessions
if setup_gnome_keyring; then
    DC_ENV="$HOME/.devcontainer-env.sh"
    if [ -f "$DC_ENV" ]; then
        # Remove existing entries to avoid duplicates
        sed -i '/^export DBUS_SESSION_BUS_ADDRESS=/d' "$DC_ENV"
        sed -i '/^export GNOME_KEYRING_CONTROL=/d' "$DC_ENV"
        sed -i '/^export SSH_AUTH_SOCK=/d' "$DC_ENV"
    fi
    # Export D-Bus and keyring variables for all shells
    if [ -n "${DBUS_SESSION_BUS_ADDRESS:-}" ]; then
        echo "export DBUS_SESSION_BUS_ADDRESS=\"$DBUS_SESSION_BUS_ADDRESS\"" >> "$DC_ENV"
    fi
    if [ -n "${GNOME_KEYRING_CONTROL:-}" ]; then
        echo "export GNOME_KEYRING_CONTROL=\"$GNOME_KEYRING_CONTROL\"" >> "$DC_ENV"
    fi
    if [ -n "${SSH_AUTH_SOCK:-}" ]; then
        echo "export SSH_AUTH_SOCK=\"$SSH_AUTH_SOCK\"" >> "$DC_ENV"
    fi
    log_success "Keyring environment variables exported to $DC_ENV"
fi

# Reload .env file to get updated tokens
ENV_FILE="/workspace/.devcontainer/.env"
if [ -f "$ENV_FILE" ]; then
    log_info "Reloading environment from .env..."
    set -a
    source "$ENV_FILE"
    set +a
fi

# ============================================================================
# MCP Configuration Setup (inject secrets into template)
# ============================================================================
VAULT_ID="ypahjj334ixtiyjkytu5hij2im"
MCP_TPL="/etc/mcp/mcp.json.tpl"
MCP_OUTPUT="/workspace/.mcp.json"

# Helper function to get 1Password field (tries multiple field names)
# Usage: get_1password_field <item_name> <vault_id>
get_1password_field() {
    local item="$1"
    local vault="$2"
    local fields=("credential" "password" "identifiant" "mot de passe")

    for field in "${fields[@]}"; do
        local value
        value=$(op item get "$item" --vault "$vault" --fields "$field" --reveal 2>/dev/null || echo "")
        if [ -n "$value" ]; then
            echo "$value"
            return 0
        fi
    done
    echo ""
}

# Initialize tokens from environment variables (fallback)
CODACY_TOKEN="${CODACY_API_TOKEN:-}"
GITHUB_TOKEN="${GITHUB_API_TOKEN:-}"
CODERABBIT_TOKEN="${CODERABBIT_API_KEY:-}"

# ============================================================================
# 1Password CLI Config Directory Permissions Fix
# ============================================================================
# Docker named volumes create directories with root ownership.
# 1Password CLI requires: ownership by current user + permissions 700.
# See: https://github.com/kodflow/devcontainer-template/issues/86
OP_CONFIG_DIRS=("$HOME/.config/op" "$HOME/.op")

for OP_DIR in "${OP_CONFIG_DIRS[@]}"; do
    if [ -d "$OP_DIR" ]; then
        # Fix ownership if not current user
        if [ "$(stat -c '%U' "$OP_DIR" 2>/dev/null)" != "$(whoami)" ]; then
            log_info "Fixing ownership of $OP_DIR..."
            sudo chown -R "$(whoami):$(whoami)" "$OP_DIR"
        fi
        # Ensure correct permissions (700 = owner only)
        chmod 700 "$OP_DIR"
    fi
done
log_success "1Password config directories configured"

# ============================================================================
# npm Cache Permissions Fix
# ============================================================================
# Docker named volumes create directories with root ownership.
# npm requires write access to its cache for npx/MCP servers to work.
# See: https://github.com/kodflow/devcontainer-template/issues/88
NPM_CACHE_DIR="$HOME/.cache/npm"

if [ -d "$NPM_CACHE_DIR" ]; then
    # Fix ownership if not current user
    if [ "$(stat -c '%U' "$NPM_CACHE_DIR" 2>/dev/null)" != "$(whoami)" ]; then
        log_info "Fixing ownership of npm cache..."
        sudo chown -R "$(whoami):$(whoami)" "$NPM_CACHE_DIR"
    fi
fi
log_success "npm cache configured"

# Try 1Password if OP_SERVICE_ACCOUNT_TOKEN is defined
if [ -n "$OP_SERVICE_ACCOUNT_TOKEN" ] && command -v op &> /dev/null; then
    log_info "Retrieving secrets from 1Password..."

    OP_CODACY=$(get_1password_field "mcp-codacy" "$VAULT_ID")
    OP_GITHUB=$(get_1password_field "mcp-github" "$VAULT_ID")
    OP_CODERABBIT=$(get_1password_field "coderabbit" "$VAULT_ID")

    [ -n "$OP_CODACY" ] && CODACY_TOKEN="$OP_CODACY"
    [ -n "$OP_GITHUB" ] && GITHUB_TOKEN="$OP_GITHUB"
    [ -n "$OP_CODERABBIT" ] && CODERABBIT_TOKEN="$OP_CODERABBIT"
fi

# Show warnings if tokens are missing
[ -z "$CODACY_TOKEN" ] && log_warning "Codacy token not available"
[ -z "$GITHUB_TOKEN" ] && log_warning "GitHub token not available"
[ -z "$CODERABBIT_TOKEN" ] && log_warning "CodeRabbit token not available"

# Generate mcp.json from template (baked in Docker image)
if [ -f "$MCP_TPL" ]; then
    log_info "Generating .mcp.json from template..."
    sed -e "s|{{CODACY_TOKEN}}|${CODACY_TOKEN}|g" \
        -e "s|{{GITHUB_TOKEN}}|${GITHUB_TOKEN}|g" \
        "$MCP_TPL" > "$MCP_OUTPUT"
    log_success "mcp.json generated successfully"

    # =========================================================================
    # Add optional MCPs based on installed features
    # =========================================================================
    # Helper function to add a conditional MCP server
    add_optional_mcp() {
        local name="$1"
        local binary="$2"
        local output="$3"

        if [ -x "$binary" ]; then
            log_info "Adding $name MCP (binary found at $binary)"
            jq --arg name "$name" --arg bin "$binary" \
               '.mcpServers[$name] = {"command": $bin, "args": [], "env": {}}' \
               "$output" > "$output.tmp" && mv "$output.tmp" "$output"
        else
            log_info "Skipping $name MCP (binary not found)"
        fi
    }

    # Rust: rust-analyzer-mcp (only if Rust feature is installed)
    add_optional_mcp "rust-analyzer" "$HOME/.cache/cargo/bin/rust-analyzer-mcp" "$MCP_OUTPUT"

    # Future conditional MCPs can be added here:
    # add_optional_mcp "gopls" "$HOME/.cache/go/bin/gopls-mcp" "$MCP_OUTPUT"
    # add_optional_mcp "pyright" "$HOME/.cache/pyenv/shims/pyright-mcp" "$MCP_OUTPUT"
else
    log_warning "MCP template not found at $MCP_TPL"
fi

# ============================================================================
# Git Credential Cleanup (remove macOS-specific helpers)
# ============================================================================
log_info "Cleaning git credential helpers..."
git config --global --unset-all credential.https://github.com.helper 2>/dev/null || true
git config --global --unset-all credential.https://gist.github.com.helper 2>/dev/null || true
log_success "Git credential helpers cleaned"

# ============================================================================
# Export dynamic environment variables (appended to ~/.devcontainer-env.sh)
# ============================================================================
# Note: ~/.devcontainer-env.sh is created by postCreate.sh with static content
# We only append dynamic variables here (secrets from 1Password)
DC_ENV="$HOME/.devcontainer-env.sh"

# Export CodeRabbit API key if available (append to existing file)
if [ -n "$CODERABBIT_TOKEN" ]; then
    # Remove any existing CODERABBIT_API_KEY line to avoid duplicates
    if [ -f "$DC_ENV" ]; then
        sed -i '/^export CODERABBIT_API_KEY=/d' "$DC_ENV"
    fi
    echo "export CODERABBIT_API_KEY=\"$CODERABBIT_TOKEN\"" >> "$DC_ENV"
    log_success "CODERABBIT_API_KEY exported to $DC_ENV"
fi

# ============================================================================
# Auto-run /init for project initialization check
# ============================================================================
# Runs at every container start to verify project is properly initialized
# (compares CLAUDE.md and README.md footprints with template)
# Skipped in CI environment

INIT_LOG="$HOME/.devcontainer-init.log"

if command -v claude &> /dev/null && [ -z "${CI:-}" ]; then
    log_info "Running project initialization check..."
    # Run /init in background to not block container startup
    # Logs persisted to $HOME for debugging (survives container restarts)
    nohup bash -c "sleep 2 && claude \"/init\" || echo \"[\$(date -Iseconds)] Init check failed with exit code \$?\" >> \"$INIT_LOG\"" >> "$INIT_LOG" 2>&1 &
    log_success "Init check scheduled (logs: ~/.devcontainer-init.log)"
elif [ -n "${CI:-}" ]; then
    log_info "CI environment detected, skipping init"
fi

# ============================================================================
# Final message
# ============================================================================
echo ""
log_success "postStart: Container ready!"
