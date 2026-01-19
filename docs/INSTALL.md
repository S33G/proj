# Installation Guide

This guide covers all installation methods for proj.

## Quick Install

### From Source (Recommended)

```bash
git clone https://github.com/s33g/proj.git
cd proj
make install
```

### One-Line Install

After releases are published:

```bash
curl -sSL https://raw.githubusercontent.com/s33g/proj/main/scripts/install.sh | bash
```

## Detailed Installation

### Method 1: From Source

#### Prerequisites

- Go 1.24 or later
- Make
- Git

#### Steps

1. **Clone the repository:**
   ```bash
   git clone https://github.com/s33g/proj.git
   cd proj
   ```

2. **Build and install:**
   ```bash
   make install
   ```

3. **The installer will:**
   - Build the binary
   - Check for existing proj installations
   - Ask before removing old versions
   - Install to `~/.local/bin/proj`
   - Display shell integration instructions

#### Custom Installation Location

```bash
# Install to /usr/local/bin (may need sudo)
PREFIX=/usr/local make install

# Install to ~/bin
PREFIX=$HOME make install
```

### Method 2: Pre-built Binaries

Download from the [releases page](https://github.com/s33g/proj/releases):

1. **Download the binary for your platform:**
   - `proj-linux-amd64` - Linux x86_64
   - `proj-linux-arm64` - Linux ARM64
   - `proj-darwin-amd64` - macOS Intel
   - `proj-darwin-arm64` - macOS Apple Silicon

2. **Install manually:**
   ```bash
   # Download (example for Linux amd64)
   curl -LO https://github.com/s33g/proj/releases/latest/download/proj-linux-amd64
   
   # Make executable
   chmod +x proj-linux-amd64
   
   # Move to bin directory
   mv proj-linux-amd64 ~/.local/bin/proj
   ```

### Method 3: Install Script

The install script automates downloading and setup:

```bash
curl -sSL https://raw.githubusercontent.com/s33g/proj/main/scripts/install.sh | bash
```

The script will:
- Detect your OS (Linux/macOS) and architecture (amd64/arm64)
- Download the latest release
- Install to `~/.local/bin`
- Optionally set up shell integration

#### Custom Install Directory

```bash
INSTALL_DIR=/usr/local/bin curl -sSL https://raw.githubusercontent.com/s33g/proj/main/scripts/install.sh | bash
```

### Method 4: Go Install

If you have Go installed:

```bash
go install github.com/s33g/proj/cmd/proj@latest
```

This installs to `$GOPATH/bin` or `$HOME/go/bin`.

## Post-Installation Setup

### 1. Verify Installation

```bash
proj --version
```

Expected output:
```
proj version v1.0.0
```

### 2. Add to PATH

If `proj` is not found, add the installation directory to your PATH.

**For `~/.local/bin`:**

Add to `~/.bashrc` or `~/.zshrc`:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then reload:
```bash
source ~/.bashrc  # or ~/.zshrc
```

### 3. Initialize Configuration

```bash
proj --init
```

This creates `~/.config/proj/config.json` with default settings.

### 4. Set Your Projects Directory

```bash
proj --set-path ~/code
```

Replace `~/code` with your actual projects directory.

### 5. Test Everything

```bash
# Launch the TUI
proj

# List projects
proj --list

# Open config
proj --config
```

## Upgrading

### From Source

```bash
cd proj
git pull
make install
```

The installer will detect the existing installation and ask before overwriting.

### From Binary

Download the new version and replace the existing binary:

```bash
curl -LO https://github.com/s33g/proj/releases/latest/download/proj-linux-amd64
chmod +x proj-linux-amd64
mv proj-linux-amd64 ~/.local/bin/proj
```

## Uninstallation

### Using Make

```bash
cd proj
make uninstall
```

### Using Script

```bash
~/.config/proj/scripts/uninstall.sh
```

### Manual

```bash
# Remove binary
rm ~/.local/bin/proj

# Remove configuration (optional)
rm -rf ~/.config/proj

# Remove shell integration from ~/.bashrc or ~/.zshrc
# (delete the proj() function)
```

## Troubleshooting

### "command not found: proj"

The installation directory is not in your PATH.

**Solution:**
```bash
export PATH="$HOME/.local/bin:$PATH"
```

Add this to your `~/.bashrc` or `~/.zshrc` to make it permanent.

### "permission denied"

The binary doesn't have execute permissions.

**Solution:**
```bash
chmod +x ~/.local/bin/proj
```

### "Change Directory" doesn't work

Shell integration is not set up.

**Solution:**
1. Add the shell function from [Step 5](#5-shell-integration-highly-recommended)
2. Reload your shell: `source ~/.bashrc`
3. Use `proj` (the function) not `~/.local/bin/proj` (the binary)

### Old version still running

There might be multiple installations.

**Solution:**
```bash
# Find all installations
which -a proj

# Remove old ones manually, or use:
make install  # Will detect and offer to remove old versions
```

### Configuration not found

Configuration hasn't been initialized.

**Solution:**
```bash
proj --init
```

### Projects not showing up

The projects path might be incorrect.

**Solution:**
```bash
# Check current path
cat ~/.config/proj/config.json | grep reposPath

# Set correct path
proj --set-path ~/your/projects/directory
```

### Build fails

Go might not be installed or is outdated.

**Solution:**
```bash
# Check Go version
go version

# Should be 1.24 or later
# Install/update Go from https://golang.org/dl/
```

## Platform-Specific Notes

### Linux

- Standard installation works on most distributions
- May need `sudo` for `/usr/local/bin` installation

### macOS

- Works on both Intel and Apple Silicon
- May need to allow in Security & Privacy on first run
- Homebrew installation coming soon

### Windows

Windows is **not supported**. Use WSL (Windows Subsystem for Linux) instead:

```bash
# In WSL
git clone https://github.com/s33g/proj.git
cd proj
make install
```

## Installation Locations

| Location | Description |
|----------|-------------|
| `~/.local/bin/proj` | Default binary location |
| `~/.config/proj/config.json` | Configuration file |
| `~/.config/proj/plugins/` | Plugin directory |

## Getting Help

- Check the [README](../README.md)
- Review [Configuration](CONFIG.md)
- Open an issue on GitHub
