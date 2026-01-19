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

# Setup shell integration
setup_shell_integration() {
    echo ""
    info "Shell Integration Setup"
    echo ""
    echo "To enable the 'cd' command feature, add this to your shell configuration:"
    echo ""

    # Detect shell
    SHELL_NAME=$(basename "$SHELL")
    
    case "$SHELL_NAME" in
        bash)
            RC_FILE="$HOME/.bashrc"
            show_bash_integration
            ;;
        zsh)
            RC_FILE="$HOME/.zshrc"  
            show_zsh_integration
            ;;
        fish)
            RC_FILE="$HOME/.config/fish/config.fish"
            show_fish_integration
            ;;
        *)
            RC_FILE=""
            show_generic_integration
            ;;
    esac

    if [ -n "$RC_FILE" ] && [ "$SHELL_NAME" != "fish" ]; then
        echo "Detected shell: $SHELL_NAME"
        echo "Add to: $RC_FILE"
        echo ""
        read -p "Would you like to add this automatically? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            add_shell_integration_$SHELL_NAME
            success "Added to $RC_FILE"
            warn "Run 'source $RC_FILE' or restart your shell to apply changes"
        fi
    elif [ "$SHELL_NAME" = "fish" ]; then
        echo "Detected shell: fish"
        echo "Add to: $RC_FILE"
        echo ""
        read -p "Would you like to add this automatically? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            add_fish_integration
            success "Added to $RC_FILE"
            warn "Restart your shell to apply changes"
        fi
    fi
}

show_bash_integration() {
    echo "  # proj - TUI project navigator"
    echo "  proj() {"
    echo "    local output=\$(mktemp)"
    echo "    PROJ_CD_FILE=\"\$output\" command $INSTALL_DIR/proj \"\$@\""
    echo "    if [ -s \"\$output\" ]; then"
    echo "      cd \"\$(cat \"\$output\")\""
    echo "    fi"
    echo "    rm -f \"\$output\""
    echo "  }"
    echo ""
}

show_zsh_integration() {
    echo "  # proj - TUI project navigator"
    echo "  proj() {"
    echo "    local output=\$(mktemp)"
    echo "    PROJ_CD_FILE=\"\$output\" command $INSTALL_DIR/proj \"\$@\""
    echo "    if [ -s \"\$output\" ]; then"
    echo "      cd \"\$(cat \"\$output\")\""
    echo "    fi"
    echo "    rm -f \"\$output\""
    echo "  }"
    echo ""
}

show_fish_integration() {
    echo "  # proj - TUI project navigator"
    echo "  function proj"
    echo "    set output (mktemp)"
    echo "    env PROJ_CD_FILE=\"\$output\" $INSTALL_DIR/proj \$argv"
    echo ""
    echo "    if test -s \"\$output\""
    echo "      cd (cat \"\$output\")"
    echo "    end"
    echo ""
    echo "    rm -f \"\$output\""
    echo "  end"
    echo ""
}

show_generic_integration() {
    echo "  Unknown shell: $SHELL_NAME"
    echo "  Please refer to the documentation for shell integration:"
    echo "  https://github.com/${REPO}#shell-integration"
    echo ""
    echo "  Or adapt one of these examples:"
    echo ""
    show_bash_integration
}

add_shell_integration_bash() {
    echo "" >> "$RC_FILE"
    echo "# proj - TUI project navigator" >> "$RC_FILE"
    echo "proj() {" >> "$RC_FILE"
    echo "  local output=\$(mktemp)" >> "$RC_FILE"
    echo "  PROJ_CD_FILE=\"\$output\" command $INSTALL_DIR/proj \"\$@\"" >> "$RC_FILE"
    echo "  if [ -s \"\$output\" ]; then" >> "$RC_FILE"
    echo "    cd \"\$(cat \"\$output\")\"" >> "$RC_FILE"
    echo "  fi" >> "$RC_FILE"
    echo "  rm -f \"\$output\"" >> "$RC_FILE"
    echo "}" >> "$RC_FILE"
}

add_shell_integration_zsh() {
    echo "" >> "$RC_FILE"
    echo "# proj - TUI project navigator" >> "$RC_FILE"
    echo "proj() {" >> "$RC_FILE"
    echo "  local output=\$(mktemp)" >> "$RC_FILE"
    echo "  PROJ_CD_FILE=\"\$output\" command $INSTALL_DIR/proj \"\$@\"" >> "$RC_FILE"
    echo "  if [ -s \"\$output\" ]; then" >> "$RC_FILE"
    echo "    cd \"\$(cat \"\$output\")\"" >> "$RC_FILE"
    echo "  fi" >> "$RC_FILE"
    echo "  rm -f \"\$output\"" >> "$RC_FILE"
    echo "}" >> "$RC_FILE"
}

add_fish_integration() {
    mkdir -p "$(dirname "$RC_FILE")"
    echo "" >> "$RC_FILE"
    echo "# proj - TUI project navigator" >> "$RC_FILE"
    echo "function proj" >> "$RC_FILE"
    echo "  set output (mktemp)" >> "$RC_FILE"
    echo "  env PROJ_CD_FILE=\"\$output\" $INSTALL_DIR/proj \$argv" >> "$RC_FILE"
    echo "" >> "$RC_FILE"
    echo "  if test -s \"\$output\"" >> "$RC_FILE"
    echo "    cd (cat \"\$output\")" >> "$RC_FILE"
    echo "  end" >> "$RC_FILE"
    echo "" >> "$RC_FILE"
    echo "  rm -f \"\$output\"" >> "$RC_FILE"
    echo "end" >> "$RC_FILE"
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
