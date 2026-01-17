#!/bin/bash
# Uninstallation script for proj - TUI project navigator

set -e

# Configuration
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
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Main uninstallation
main() {
    echo ""
    echo "╔═══════════════════════════════════════╗"
    echo "║   proj - TUI Project Navigator        ║"
    echo "║   Uninstallation Script               ║"
    echo "╚═══════════════════════════════════════╝"
    echo ""
    
    # Remove binary
    if [ -f "$INSTALL_DIR/proj" ]; then
        info "Removing binary from $INSTALL_DIR/proj..."
        rm -f "$INSTALL_DIR/proj"
        success "Binary removed"
    else
        warn "Binary not found at $INSTALL_DIR/proj"
    fi
    
    # Ask about configuration
    echo ""
    if [ -d "$CONFIG_DIR" ]; then
        echo "Configuration directory found at: $CONFIG_DIR"
        read -p "Would you like to remove configuration too? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$CONFIG_DIR"
            success "Configuration removed"
        else
            info "Configuration kept at $CONFIG_DIR"
        fi
    fi
    
    # Remind about shell integration
    echo ""
    warn "Don't forget to remove shell integration from ~/.bashrc or ~/.zshrc:"
    echo "  Remove the 'proj()' function definition"
    
    echo ""
    echo "╔═══════════════════════════════════════╗"
    echo "║   Uninstallation Complete!            ║"
    echo "╚═══════════════════════════════════════╝"
    echo ""
}

main
