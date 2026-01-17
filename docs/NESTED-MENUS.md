# Nested Action Menus Implementation

## Overview

Implemented hierarchical/nested action menus to organize the growing list of actions into logical groups. Actions are now grouped by category (npm, Docker, Compose, etc.) with submenu navigation.

## Changes Made

### 1. Updated Action Structure (`internal/tui/views/action_menu.go`)

Added fields to support submenus:
```go
type Action struct {
    ID       string
    Label    string
    Desc     string
    Icon     string
    Command  string
    Source   string
    IsSubmenu bool    // NEW: Whether this action opens a submenu
    Children []Action // NEW: Submenu actions
}
```

### 2. Grouping Logic

**Script Actions:**
- If a source (e.g., `package.json`, `Makefile`) has **3+ scripts**: Create a submenu
- If a source has **< 3 scripts**: Show individual actions (no submenu)
- Submenu naming:
  - `package.json` → "npm Scripts"
  - `Makefile` → "Make"
  - `cargo` → "Cargo"
  - etc.

**Docker Actions:**
- Always grouped into submenus when present:
  - 🐳 **Docker** → Single container actions (build, run, logs, ps)
  - 🐙 **Docker Compose** → Multi-service actions (up, down, build, logs, ps)

### 3. App Model Updates (`internal/app/app.go`)

Added submenu stack for navigation:
```go
type Model struct {
    // ... existing fields ...
    submenuStack []views.ActionMenuModel // Stack for nested submenus
}
```

### 4. Navigation Behavior

**Entering a Submenu:**
- Press Enter on a submenu item (shown with `→` indicator)
- Current menu is pushed onto the stack
- Submenu is displayed

**Exiting a Submenu:**
- Press Esc or select "← Back"
- Previous menu is popped from the stack
- User returns to parent menu

**Back Button:**
- Always present at bottom of submenu
- Returns to parent menu or project list

### 5. Visual Indicators

- **Submenu items**: Shown with `→` arrow suffix
- **Submenu actions**: Shown with `▸` prefix icon
- **Help text**: Changes based on context
  - Main menu: "enter: execute"
  - Submenu: "enter: select"

## Example Menu Structure

```
Actions for my-project:

▸ 🚀 Open in Editor
  📂 Change Directory
  🔍 View Git Log
  📜 npm Scripts →              ← Submenu (3+ scripts)
      5 available commands
  🐳 Docker →                   ← Submenu (Docker actions)
      5 container actions
  🐙 Docker Compose →           ← Submenu (Compose actions)
      6 service actions
  📦 Install Dependencies
  🗑️ Clean Build Artifacts
  ← Back
```

**Inside "npm Scripts" submenu:**
```
npm Scripts:

▸ ▸ dev
      Run development server
  ▸ build
      Build for production
  ▸ test
      Run test suite
  ▸ lint
      Lint code
  ▸ format
      Format code
  ← Back
```

**Inside "Docker" submenu:**
```
Docker:

▸ 🏗️ Build Image
      Build Docker image
  ▶️ Run Container
      Run container interactively
  🔄 Run Detached
      Run container in background
  📋 List Containers
      Show running containers
  📜 View Logs
      Stream container logs
  ← Back
```

## Benefits

1. **Reduced Clutter**: Main menu is much cleaner with fewer items
2. **Logical Grouping**: Related actions are grouped together
3. **Scalability**: Can handle projects with many scripts/actions
4. **Familiar UX**: Submenu pattern is common in TUIs
5. **Flexible**: Automatically groups only when needed (3+ items)

## Technical Details

### Thresholds

- **Script grouping threshold**: 3+ scripts from same source
- **Docker/Compose**: Always grouped (when present)

### Navigation Stack

Uses a stack-based approach:
- Push current menu when entering submenu
- Pop from stack when exiting submenu
- Allows for potential multi-level nesting (future enhancement)

### Helper Functions

Added `getSourceDisplayName()` to convert technical source names to friendly display names:
- `package.json` → "npm Scripts"
- `Makefile` → "Make"
- `cargo` → "Cargo"
- etc.

## Testing

- ✅ Builds successfully
- ✅ All existing tests pass
- ✅ No breaking changes
- ✅ Backward compatible (single scripts still show individually)

## Future Enhancements

Potential improvements:
- Configurable grouping threshold
- Multi-level nested menus (submenus within submenus)
- Breadcrumb trail in header showing menu path
- Quick access keys (e.g., `1-9` for top 9 items)
