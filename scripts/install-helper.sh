#!/bin/bash
# Installation helper script that checks for existing installations

BINARY_NAME="proj"
INSTALL_DIR="${1:-$HOME/.local/bin}"
SOURCE_BINARY="${2:-./proj}"

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Common installation directories to check
CHECK_DIRS=(
    "$HOME/bin"
    "$HOME/.local/bin"
    "/usr/local/bin"
    "/usr/bin"
)

echo "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
echo ""

# Check for existing installations in other locations
for dir in "${CHECK_DIRS[@]}"; do
    # Skip the target installation directory (we'll check it separately)
    if [ "$dir" = "$INSTALL_DIR" ]; then
        continue
    fi
    
    if [ -f "$dir/$BINARY_NAME" ]; then
        echo -e "${YELLOW}⚠${NC} Found existing $BINARY_NAME in $dir"
        
        # Try to get version
        VERSION=$("$dir/$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        echo "  Version: $VERSION"
        echo ""
        
        read -p "Remove $dir/$BINARY_NAME? [y/N] " -n 1 -r
        echo ""
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            if rm -f "$dir/$BINARY_NAME" 2>/dev/null; then
                echo -e "${GREEN}✓${NC} Removed $dir/$BINARY_NAME"
            else
                echo -e "${RED}✗${NC} Failed to remove (may need sudo)"
                echo "  You can manually remove it with: sudo rm $dir/$BINARY_NAME"
            fi
        else
            echo -e "${BLUE}ℹ${NC} Kept $dir/$BINARY_NAME (you may have multiple versions)"
        fi
        echo ""
    fi
done

# Check if target location already exists
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo -e "${YELLOW}⚠${NC} $INSTALL_DIR/$BINARY_NAME already exists"
    
    # Try to get versions
    CURRENT_VERSION=$("$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null || echo "unknown")
    NEW_VERSION=$("$SOURCE_BINARY" --version 2>/dev/null || echo "unknown")
    
    echo "  Current version: $CURRENT_VERSION"
    echo "  New version:     $NEW_VERSION"
    echo ""
    
    read -p "Overwrite? [Y/n] " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo "Installation cancelled."
        exit 1
    fi
    echo ""
fi

# Create directory if needed
mkdir -p "$INSTALL_DIR"

# Copy binary
if cp "$SOURCE_BINARY" "$INSTALL_DIR/$BINARY_NAME"; then
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo -e "${GREEN}✓${NC} Installed to $INSTALL_DIR/$BINARY_NAME"
    exit 0
else
    echo -e "${RED}✗${NC} Failed to install"
    exit 1
fi
