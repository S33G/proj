#!/bin/bash
# Installation script for proj - TUI project navigator
# Usage: curl -sSL https://raw.githubusercontent.com/s33g/proj/main/scripts/install.sh | bash

set -e

# Configuration
REPO="s33g/proj"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="$HOME/.config/proj"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${GREEN}ℹ${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux*)
            OS="linux"
            ;;
        Darwin*)
            OS="darwin"
            ;;
        *)
            error "Unsupported operating system: $OS (only Linux and macOS are supported)"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH (only amd64 and arm64 are supported)"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Get latest release version
get_latest_version() {
    info "Fetching latest release..."

    # Try to get latest release from GitHub API
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget found. Please install one of them."
    fi

    if [ -z "$VERSION" ]; then
        error "Failed to fetch latest version"
    fi

    success "Latest version: $VERSION"
}

# Download and install binary
install_binary() {
    BINARY_NAME="proj-${PLATFORM}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"

    info "Downloading from $DOWNLOAD_URL..."

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download binary
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$DOWNLOAD_URL" -o "$TMP_DIR/proj" || error "Failed to download binary"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$TMP_DIR/proj" "$DOWNLOAD_URL" 2>&1 || error "Failed to download binary"
        if [ ! -s "$TMP_DIR/proj" ]; then
            error "Failed to download binary: empty file"
        fi
    fi

    # Verify the downloaded file is a valid binary (not an HTML error page)
    if file "$TMP_DIR/proj" | grep -qE "text|HTML"; then
        CONTENT=$(head -c 100 "$TMP_DIR/proj")
        error "Downloaded file is not a valid binary. Content: $CONTENT\nThe release may not have binaries attached yet."
    fi

    # Make executable
    chmod +x "$TMP_DIR/proj"

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Move binary to install directory
    mv "$TMP_DIR/proj" "$INSTALL_DIR/proj"

    success "Installed to $INSTALL_DIR/proj"
}

# Check PATH
check_path() {
    echo ""
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH"
        echo "Add this to your shell configuration (~/.bashrc or ~/.zshrc):"
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    else
        success "$INSTALL_DIR is in your PATH"
    fi
}

# Setup shell integration
setup_shell_integration() {
    echo ""
    info "Setting up shell integration..."
    
    # Detect current shell
    CURRENT_SHELL=$(basename "$SHELL" 2>/dev/null || echo "unknown")
    
    case "$CURRENT_SHELL" in
        bash)
            info "Detected bash shell"
            setup_bash_integration
            ;;
        zsh)
            info "Detected zsh shell"
            setup_zsh_integration
            ;;
        fish)
            info "Detected fish shell"
            setup_fish_integration
            ;;
        *)
            warn "Shell '$CURRENT_SHELL' is not directly supported"
            show_manual_integration_help
            ;;
    esac
}

# Setup bash integration
setup_bash_integration() {
    local shell_script_url="https://raw.githubusercontent.com/${REPO}/main/scripts/shells/bash.sh"
    local target_file="$HOME/.config/proj/bash_integration.sh"
    
    mkdir -p "$(dirname "$target_file")"
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$shell_script_url" -o "$target_file"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$target_file" "$shell_script_url"
    else
        warn "Cannot download shell integration file. Please see documentation for manual setup."
        show_manual_integration_help
        return
    fi
    
    success "Downloaded bash integration to $target_file"
    echo "Add this line to your ~/.bashrc:"
    echo "  ${GREEN}source $target_file${NC}"
}

# Setup zsh integration
setup_zsh_integration() {
    local shell_script_url="https://raw.githubusercontent.com/${REPO}/main/scripts/shells/zsh.sh"
    local target_file="$HOME/.config/proj/zsh_integration.sh"
    
    mkdir -p "$(dirname "$target_file")"
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$shell_script_url" -o "$target_file"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$target_file" "$shell_script_url"
    else
        warn "Cannot download shell integration file. Please see documentation for manual setup."
        show_manual_integration_help
        return
    fi
    
    success "Downloaded zsh integration to $target_file"
    echo "Add this line to your ~/.zshrc:"
    echo "  ${GREEN}source $target_file${NC}"
}

# Setup fish integration
setup_fish_integration() {
    local shell_script_url="https://raw.githubusercontent.com/${REPO}/main/scripts/shells/fish.fish"
    local target_file="$HOME/.config/fish/conf.d/proj.fish"
    
    mkdir -p "$(dirname "$target_file")"
    
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL "$shell_script_url" -o "$target_file"; then
            warn "Failed to download fish integration using curl. Please see documentation for manual setup."
            show_manual_integration_help
            return
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -qO "$target_file" "$shell_script_url"; then
            warn "Failed to download fish integration using wget. Please see documentation for manual setup."
            show_manual_integration_help
            return
        fi
    else
        warn "Cannot download shell integration file. Please see documentation for manual setup."
        show_manual_integration_help
        return
    fi
    
    if [ ! -s "$target_file" ]; then
        warn "Downloaded fish integration file is empty or missing. Please see documentation for manual setup."
        show_manual_integration_help
        return
    fi
    
    success "Fish integration installed to $target_file"
    info "Fish integration will be active in new sessions"
}

# Show manual integration help
show_manual_integration_help() {
    echo ""
    echo "For manual shell integration setup, please see:"
    echo "  ${GREEN}https://github.com/${REPO}/blob/main/docs/INSTALL.md#shell-integration${NC}"
    echo "  ${GREEN}https://github.com/${REPO}/blob/main/docs/CONTRIBUTING.md#adding-shell-support${NC}"
    echo ""
    echo "Available shell integrations:"
    echo "  • bash:  https://github.com/${REPO}/blob/main/scripts/shells/bash.sh"
    echo "  • zsh:   https://github.com/${REPO}/blob/main/scripts/shells/zsh.sh" 
    echo "  • fish:  https://github.com/${REPO}/blob/main/scripts/shells/fish.fish"
    echo ""
    echo "To contribute support for your shell, see the contribution guide above."
}

# Main installation
main() {
    echo ""
    echo "╔═══════════════════════════════════════╗"
    echo "║   proj - TUI Project Navigator        ║"
    echo "║   Installation Script                 ║"
    echo "╚═══════════════════════════════════════╝"
    echo ""

    detect_platform
    info "Platform: $PLATFORM"

    get_latest_version
    install_binary
    check_path
    setup_shell_integration

    echo ""
    echo "╔═══════════════════════════════════════╗"
    echo "║   Installation Complete!              ║"
    echo "╚═══════════════════════════════════════╝"
    echo ""
    echo "Get started:"
    echo "  1. Run: ${GREEN}proj --init${NC}      # Initialize configuration"
    echo "  2. Run: ${GREEN}proj --set-path ~/code${NC}  # Set your projects directory"
    echo "  3. Run: ${GREEN}proj${NC}             # Launch the TUI"
    echo ""
    echo "Documentation: https://github.com/${REPO}"
    echo ""
}

main
