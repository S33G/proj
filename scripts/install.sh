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
