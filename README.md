# proj

A fast, intuitive TUI project navigator for developers. Quickly browse, open, and manage your code projects from the terminal.

[![CI](https://github.com/S33G/proj/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/S33G/proj/actions/workflows/ci.yml?query=branch%3Amain)
[![codecov](https://codecov.io/gh/S33G/proj/branch/main/graph/badge.svg)](https://codecov.io/gh/S33G/proj)
![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey)

## Features

- **Fast Project Discovery** - Instantly scan and list all projects in your code directory
- **Smart Sorting** - Cycle through Alphabetical, Last Modified, or Language grouping with `s` key
- **Language Detection** - Automatically detects 17+ programming languages
- **Git Integration** - Shows branch, dirty status, and supports git operations
- **Docker & Compose Support** - Detect and manage containerized projects with built-in actions 🐳
- **Multi-Editor Support** - VS Code, Neovim, Vim, Emacs, JetBrains IDEs, Zed, and more
- **Built-in Actions** - Open editor, run tests, install deps, git operations, Docker commands
- **Plugin System** - Extend with custom actions via JSON-RPC plugins
- **Shell Integration** - Change directory directly from the TUI
- **Quick Project Creation** - Press `n` to create new projects on the fly
- **Keyboard-Driven** - Vi-style navigation with intuitive shortcuts

## Demo

```
┌─────────────────────────────────────────────────────────┐
│  📁 Projects in ~/code (42 projects)                    │
│  Sort: Alphabetical (A-Z)                               │
├─────────────────────────────────────────────────────────┤
│  > myapp           Go       main      ●                 │
│    webapp          TypeScript feat/auth                 │
│    api-server      Rust     develop   ●                 │
│    dotfiles        Shell    main                        │
│    ml-project      Python   main                        │
└─────────────────────────────────────────────────────────┘
  ↑/↓: navigate  •  enter: select  •  s: sort  •  n: new  •  q: quit
```

## Quick Start

### Installation

**From source (recommended):**

```bash
git clone https://github.com/s33g/proj.git
cd proj
make install
```

**One-line install:**

```bash
curl -sSL https://raw.githubusercontent.com/s33g/proj/main/scripts/install.sh | bash
```

### Setup

1. **Initialize configuration:**
   ```bash
   proj --init
   ```

2. **Set your projects directory:**
   ```bash
   proj --set-path ~/code
   ```

3. **Launch the TUI:**
   ```bash
   proj
   ```

## Usage

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `Enter` | Select project / Execute action |
| `Esc` | Go back |
| `n` | New project |
| `s` | Cycle sort (Name → Modified → Language) |
| `q` | Quit |
| `/` | Search/filter |

### CLI Commands

```bash
proj                    # Launch TUI
proj <project-name>     # Jump directly to project
proj --list             # List all projects (non-interactive)
proj --init             # Initialize/reset configuration
proj --config           # Open config in $EDITOR
proj --set-path <path>  # Set projects directory
proj --version          # Show version
proj --help             # Show help
```

### Available Actions

When you select a project, these actions are available:

| Action | Description |
|--------|-------------|
| 🚀 Open in Editor | Open project in your configured editor |
| 📂 Change Directory | Navigate to project directory (requires shell integration) |
| 🔍 View Git Log | Show recent commits |
| 🔄 Git Pull | Pull latest changes |
| 🌿 Switch Branch | Checkout a different branch |
| 🧪 Run Tests | Execute test suite |
| 📦 Install Dependencies | Run package manager install |
| 🗑️ Clean Build Artifacts | Remove build directories |

**Docker Actions** (when Dockerfile or docker-compose.yml detected):

| Action | Description |
|--------|-------------|
| 🏗️ Build Image | Build Docker image |
| ▶️ Run Container | Run container interactively |
| 🔄 Run Detached | Run container in background |
| 🚀 Compose Up | Start all services |
| 🛑 Compose Down | Stop and remove services |
| 📋 Compose PS | List services |

> See [docs/DOCKER.md](docs/DOCKER.md) for full Docker integration guide

## Configuration

Configuration is stored in `~/.config/proj/config.json`.

### Quick Config

```bash
proj --config   # Opens config in your $EDITOR
```

### Example Configuration

```json
{
  "reposPath": "~/code",
  "editor": {
    "default": "code",
    "aliases": {
      "code": ["code", "--goto"],
      "nvim": ["nvim"],
      "idea": ["idea"]
    }
  },
  "theme": {
    "primaryColor": "#00CED1",
    "accentColor": "#32CD32",
    "errorColor": "#FF6347"
  },
  "display": {
    "showHiddenDirs": false,
    "sortBy": "lastModified",
    "showGitStatus": true,
    "showLanguage": true
  },
  "plugins": {
    "enabled": [],
    "config": {}
  }
}
```

See [docs/CONFIG.md](docs/CONFIG.md) for the complete configuration reference.

## Supported Languages

proj automatically detects these languages:

| Language | Detection |
|----------|-----------|
| Go | `go.mod`, `go.sum` |
| Rust | `Cargo.toml` |
| TypeScript | `tsconfig.json`, `.ts` files |
| JavaScript | `package.json`, `.js` files |
| Python | `pyproject.toml`, `setup.py`, `requirements.txt` |
| Java | `pom.xml`, `build.gradle` |
| C# | `*.csproj`, `*.sln` |
| Ruby | `Gemfile` |
| PHP | `composer.json` |
| Swift | `Package.swift` |
| Kotlin | `build.gradle.kts` |
| C/C++ | `CMakeLists.txt`, `Makefile` |
| Elixir | `mix.exs` |
| Zig | `build.zig` |
| Haskell | `*.cabal`, `stack.yaml` |
| Scala | `build.sbt` |
| Clojure | `project.clj`, `deps.edn` |

## Supported Editors

| Editor | Command | Notes |
|--------|---------|-------|
| VS Code | `code` | Default |
| Neovim | `nvim` | |
| Vim | `vim` | |
| Emacs | `emacsclient` | Uses `-n` flag |
| IntelliJ IDEA | `idea` | |
| GoLand | `goland` | |
| PyCharm | `pycharm` | |
| WebStorm | `webstorm` | |
| CLion | `clion` | |
| RubyMine | `rubymine` | |
| PhpStorm | `phpstorm` | |
| Zed | `zed` | |
| Sublime Text | `subl` | |
| Helix | `hx` | |
| Cursor | `cursor` | |

Configure your default editor:
```bash
proj --config
# Set "editor.default": "nvim"
```

## Plugins

Extend proj with custom actions using the plugin system. Plugins communicate via JSON-RPC over stdin/stdout and can be written in any language.

### Enable a Plugin

```json
{
  "plugins": {
    "enabled": ["my-plugin"],
    "config": {
      "my-plugin": {
        "option": "value"
      }
    }
  }
}
```

### Create a Plugin

See [docs/PLUGINS.md](docs/PLUGINS.md) for the complete plugin development guide.

## Building from Source

### Requirements

- Go 1.24 or later
- Make

### Build

```bash
git clone https://github.com/s33g/proj.git
cd proj
make build      # Build for current platform
make test       # Run tests
make install    # Install to ~/.local/bin
```

### Development

```bash
make dev                        # Run in development mode
make dev ARGS="--list"          # Run with arguments
make dev ARGS="--set-path ~/code"
make lint                       # Run linter
make fmt                        # Format code
```

## Project Structure

```
proj/
├── cmd/proj/           # CLI entrypoint
├── internal/
│   ├── app/            # Main TUI application
│   ├── actions/        # Built-in actions (open, test, git, etc.)
│   ├── config/         # Configuration management
│   ├── git/            # Git operations
│   ├── language/       # Language detection
│   ├── project/        # Project scanning
│   └── tui/            # TUI components and styles
├── pkg/plugin/         # Plugin system
├── plugins/example/    # Example plugin
├── scripts/            # Installation scripts
└── docs/               # Documentation
```

## Documentation

- [Configuration Reference](docs/CONFIG.md)
- [Plugin Development](docs/PLUGINS.md)
- [Installation Guide](docs/INSTALL.md)
- [Release Guide](docs/RELEASE.md)
- [Contributing](docs/CONTRIBUTING.md)

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Built with these excellent Go libraries:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
