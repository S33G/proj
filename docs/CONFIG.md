# Configuration Reference

This document describes all configuration options for proj.

## Configuration File Location

```
~/.config/proj/config.json
```

## Quick Access

```bash
proj --init     # Create default configuration
proj --config   # Open config in $EDITOR
```

## Full Configuration Example

```json
{
  "reposPath": "~/code",
  "editor": {
    "default": "code",
    "aliases": {
      "code": ["code", "--goto"],
      "nvim": ["nvim"],
      "vim": ["vim"],
      "emacs": ["emacsclient", "-n"],
      "idea": ["idea"],
      "goland": ["goland"],
      "pycharm": ["pycharm"],
      "webstorm": ["webstorm"],
      "clion": ["clion"],
      "rubymine": ["rubymine"],
      "phpstorm": ["phpstorm"],
      "zed": ["zed"],
      "subl": ["subl"],
      "hx": ["hx"],
      "cursor": ["cursor"]
    }
  },
  "shell": "/bin/bash",
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
  "excludePatterns": [
    ".git",
    "node_modules",
    ".DS_Store",
    "__pycache__",
    "vendor"
  ],
  "actions": {
    "enableGitOperations": true,
    "enableTestRunner": true
  },
  "plugins": {
    "enabled": [],
    "config": {}
  }
}
```

## Configuration Options

### reposPath

**Type:** `string`  
**Default:** `~/code`

The root directory where your projects are located. Supports `~` expansion.

```json
{
  "reposPath": "~/code"
}
```

```json
{
  "reposPath": "/home/user/projects"
}
```

Set via CLI:
```bash
proj --set-path ~/code
```

---

### editor

**Type:** `object`

Editor configuration for the "Open in Editor" action.

#### editor.default

**Type:** `string`  
**Default:** `"code"`

The default editor to use. Must be a key in `editor.aliases`.

```json
{
  "editor": {
    "default": "nvim"
  }
}
```

#### editor.aliases

**Type:** `object`  
**Default:** See below

Map of editor names to command arrays. The first element is the command, subsequent elements are arguments.

```json
{
  "editor": {
    "aliases": {
      "code": ["code", "--goto"],
      "nvim": ["nvim"],
      "vim": ["vim"],
      "emacs": ["emacsclient", "-n"],
      "idea": ["idea"],
      "goland": ["goland"],
      "pycharm": ["pycharm"],
      "webstorm": ["webstorm"],
      "clion": ["clion"],
      "rubymine": ["rubymine"],
      "phpstorm": ["phpstorm"],
      "zed": ["zed"],
      "subl": ["subl"],
      "hx": ["hx"],
      "cursor": ["cursor"]
    }
  }
}
```

**Adding a custom editor:**

```json
{
  "editor": {
    "default": "my-editor",
    "aliases": {
      "my-editor": ["my-editor", "--some-flag", "--another-flag"]
    }
  }
}
```

---

### shell

**Type:** `string`  
**Default:** `"/bin/bash"`

The shell to use for running commands.

```json
{
  "shell": "/bin/zsh"
}
```

---

### theme

**Type:** `object`

Color theme configuration using hex color codes.

#### theme.primaryColor

**Type:** `string`  
**Default:** `"#00CED1"` (dark cyan)

Primary color used for titles and highlights.

#### theme.accentColor

**Type:** `string`  
**Default:** `"#32CD32"` (lime green)

Accent color used for success states and selections.

#### theme.errorColor

**Type:** `string`  
**Default:** `"#FF6347"` (tomato red)

Color used for errors and warnings.

```json
{
  "theme": {
    "primaryColor": "#61AFEF",
    "accentColor": "#98C379",
    "errorColor": "#E06C75"
  }
}
```

**Popular color schemes:**

One Dark:
```json
{
  "theme": {
    "primaryColor": "#61AFEF",
    "accentColor": "#98C379",
    "errorColor": "#E06C75"
  }
}
```

Dracula:
```json
{
  "theme": {
    "primaryColor": "#BD93F9",
    "accentColor": "#50FA7B",
    "errorColor": "#FF5555"
  }
}
```

Nord:
```json
{
  "theme": {
    "primaryColor": "#88C0D0",
    "accentColor": "#A3BE8C",
    "errorColor": "#BF616A"
  }
}
```

Catppuccin Mocha:
```json
{
  "theme": {
    "primaryColor": "#89B4FA",
    "accentColor": "#A6E3A1",
    "errorColor": "#F38BA8"
  }
}
```

---

### display

**Type:** `object`

Display and UI preferences.

#### display.showHiddenDirs

**Type:** `boolean`  
**Default:** `false`

Whether to show hidden directories (starting with `.`) in the project list.

```json
{
  "display": {
    "showHiddenDirs": true
  }
}
```

#### display.sortBy

**Type:** `string`  
**Default:** `"lastModified"`  
**Options:** `"name"`, `"lastModified"`

How to sort projects in the list.

- `"name"` - Alphabetical by project name
- `"lastModified"` - Most recently modified first

```json
{
  "display": {
    "sortBy": "name"
  }
}
```

#### display.showGitStatus

**Type:** `boolean`  
**Default:** `true`

Whether to show git branch and dirty status in the project list.

> **Note:** This option is currently defined but not yet implemented. Git status is always shown.

```json
{
  "display": {
    "showGitStatus": false
  }
}
```

#### display.showLanguage

**Type:** `boolean`  
**Default:** `true`

Whether to show detected language in the project list.

> **Note:** This option is currently defined but not yet implemented. Language is always shown.

```json
{
  "display": {
    "showLanguage": false
  }
}
```

---

### excludePatterns

**Type:** `array of strings`  
**Default:** `[".git", "node_modules", ".DS_Store", "__pycache__", "vendor"]`

Directory names to exclude when scanning for projects.

```json
{
  "excludePatterns": [
    ".git",
    "node_modules",
    ".DS_Store",
    "__pycache__",
    "vendor",
    "dist",
    "build",
    ".cache"
  ]
}
```

---

### actions

**Type:** `object`

Configuration for built-in actions.

#### actions.enableGitOperations

**Type:** `boolean`  
**Default:** `true`

Whether to show git-related actions (log, pull, branch) in the action menu.

```json
{
  "actions": {
    "enableGitOperations": false
  }
}
```

#### actions.enableTestRunner

**Type:** `boolean`  
**Default:** `true`

Whether to show the "Run Tests" action in the action menu.

```json
{
  "actions": {
    "enableTestRunner": false
  }
}
```

---

### plugins

**Type:** `object`

Plugin system configuration.

#### plugins.enabled

**Type:** `array of strings`  
**Default:** `[]`

List of plugin names to enable. Plugins are loaded from `~/.config/proj/plugins/`.

```json
{
  "plugins": {
    "enabled": ["my-plugin", "another-plugin"]
  }
}
```

#### plugins.config

**Type:** `object`  
**Default:** `{}`

Per-plugin configuration. Each key is a plugin name, and the value is passed to that plugin during initialization.

```json
{
  "plugins": {
    "enabled": ["github-plugin"],
    "config": {
      "github-plugin": {
        "token": "ghp_xxxxxxxxxxxx",
        "showPRs": true
      }
    }
  }
}
```

See [PLUGINS.md](PLUGINS.md) for more information on developing and using plugins.

---

## Environment Variables

### PROJ_CD_FILE

When set, proj writes the selected directory path to this file on exit. Used for shell integration.

```bash
PROJ_CD_FILE=/tmp/proj_cd proj
```

### EDITOR

Falls back to this when opening configuration if `editor.default` is not set.

```bash
EDITOR=nvim proj --config
```

---

## Resetting Configuration

To reset to defaults:

```bash
proj --init
```

Or manually:

```bash
rm ~/.config/proj/config.json
proj --init
```

---

## Configuration Directory Structure

```
~/.config/proj/
├── config.json          # Main configuration file
└── plugins/             # Plugin directory
    ├── my-plugin/
    │   ├── plugin.json  # Plugin manifest
    │   └── my-plugin    # Plugin executable
    └── another-plugin/
        ├── plugin.json
        └── another-plugin
```

---

## Tips

### Using Different Configs

While not directly supported, you can use symlinks or set `XDG_CONFIG_HOME`:

```bash
# Use a different config directory
XDG_CONFIG_HOME=~/my-config proj
```

### Minimal Configuration

For a minimal setup, you only need:

```json
{
  "reposPath": "~/code"
}
```

All other values will use defaults.

### Validating Configuration

proj validates configuration on load. If there are issues, it falls back to defaults and logs warnings to stderr.
