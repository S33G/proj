#!/bin/bash
# Shell integration setup script for proj
# Usage: ./shell-integration.sh [shell] [install_dir]

set -e

SHELL_TYPE="${1:-auto}"
INSTALL_DIR="${2:-$HOME/.local/bin}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() {
    echo -e "${GREEN}ℹ${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Auto-detect shell if not specified
if [ "$SHELL_TYPE" = "auto" ]; then
    SHELL_TYPE=$(basename "$SHELL")
fi

case "$SHELL_TYPE" in
    bash)
        RC_FILE="$HOME/.bashrc"
        FUNCTION_DEF="# proj - TUI project navigator
proj() {
  local output=\$(mktemp)
  PROJ_CD_FILE=\"\$output\" command $INSTALL_DIR/proj \"\$@\"
  if [ -s \"\$output\" ]; then
    cd \"\$(cat \"\$output\")\"
  fi
  rm -f \"\$output\"
}"
        SOURCE_CMD="source ~/.bashrc"
        ;;
    zsh)
        RC_FILE="$HOME/.zshrc"
        FUNCTION_DEF="# proj - TUI project navigator
proj() {
  local output=\$(mktemp)
  PROJ_CD_FILE=\"\$output\" command $INSTALL_DIR/proj \"\$@\"
  if [ -s \"\$output\" ]; then
    cd \"\$(cat \"\$output\")\"
  fi
  rm -f \"\$output\"
}"
        SOURCE_CMD="source ~/.zshrc"
        ;;
    fish)
        RC_FILE="$HOME/.config/fish/config.fish"
        FUNCTION_DEF="# proj - TUI project navigator
function proj
  set output (mktemp)
  env PROJ_CD_FILE=\"\$output\" $INSTALL_DIR/proj \$argv

  if test -s \"\$output\"
    cd (cat \"\$output\")
  end

  rm -f \"\$output\"
end"
        SOURCE_CMD="restart your terminal"
        ;;
    nushell|nu)
        RC_FILE="$HOME/.config/nushell/config.nu"
        FUNCTION_DEF="# proj - TUI project navigator
def proj [...args] {
  let output = (mktemp)
  with-env [PROJ_CD_FILE $output] { ^$INSTALL_DIR/proj ...\$args }
  if (\$output | path exists) and ((\$output | open | str length) > 0) {
    cd (\$output | open)
  }
  rm \$output
}"
        SOURCE_CMD="restart your terminal"
        ;;
    elvish)
        RC_FILE="$HOME/.elvish/rc.elv"
        FUNCTION_DEF="# proj - TUI project navigator
fn proj {|@args|
  var output = (mktemp)
  E:PROJ_CD_FILE=\$output $INSTALL_DIR/proj \$@args
  if (test -s \$output) {
    cd (slurp < \$output)
  }
  rm -f \$output
}"
        SOURCE_CMD="restart your terminal"
        ;;
    powershell|pwsh)
        if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
            RC_FILE="$HOME/Documents/PowerShell/Microsoft.PowerShell_profile.ps1"
        else
            RC_FILE="$HOME/.config/powershell/Microsoft.PowerShell_profile.ps1"
        fi
        FUNCTION_DEF="# proj - TUI project navigator
function proj {
    \$output = New-TemporaryFile
    \$env:PROJ_CD_FILE = \$output.FullName
    & \"$INSTALL_DIR/proj\" @args
    if (Test-Path \$output -PathType Leaf) {
        \$path = Get-Content \$output -Raw
        if (\$path.Trim()) {
            Set-Location \$path.Trim()
        }
    }
    Remove-Item \$output -ErrorAction SilentlyContinue
}"
        SOURCE_CMD="restart your terminal or run: . \$PROFILE"
        ;;
    *)
        echo "Unsupported shell: $SHELL_TYPE"
        echo "Supported shells: bash, zsh, fish, nushell, elvish, powershell"
        echo ""
        echo "For other shells, adapt this pattern:"
        echo "1. Create a temporary file with mktemp"
        echo "2. Set PROJ_CD_FILE environment variable to that file"
        echo "3. Run proj with your arguments"
        echo "4. If the file exists and has content, cd to that path"
        echo "5. Clean up the temporary file"
        exit 1
        ;;
esac

echo "Setting up proj shell integration for: $SHELL_TYPE"
echo "Configuration file: $RC_FILE"
echo ""

# Check if function already exists
if [ -f "$RC_FILE" ] && grep -q "proj.*TUI project navigator" "$RC_FILE"; then
    warn "proj function already exists in $RC_FILE"
    echo ""
    read -p "Replace existing function? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Keeping existing function"
        exit 0
    fi
    
    # Remove existing function (simplified - removes from # proj comment to next function/end of file)
    case "$SHELL_TYPE" in
        fish|nushell|elvish)
            # For these shells, we'll append and let user clean up manually if needed
            ;;
        *)
            # Remove between # proj comment and next empty line or end of function
            sed -i '/# proj - TUI project navigator/,/^}/d' "$RC_FILE" 2>/dev/null || true
            ;;
    esac
fi

# Create directory if it doesn't exist
mkdir -p "$(dirname "$RC_FILE")"

# Add the function
echo "" >> "$RC_FILE"
echo "$FUNCTION_DEF" >> "$RC_FILE"

success "Added proj shell integration to $RC_FILE"
echo ""
info "To activate the changes, $SOURCE_CMD"
echo ""
echo "Test with: proj --help"