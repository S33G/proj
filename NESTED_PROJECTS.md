# Nested Project Support

## Overview

The project scanner now supports **grouped project view** with 1-level nesting, allowing organized project categories.

## Features Implemented

### 1. **Grouped View with Category Headers**
- Scans 1 level deep to detect sub-projects
- Parent folders with sub-projects become non-selectable category headers
- Only immediate children are shown (no deep recursion)
- If no nesting exists, shows normal flat list
- Example structure:
  ```
  nested-dir (category - not selectable)
    ├── project1
    └── project2
  webdev (category - not selectable)
    ├── frontend
    ├── backend
    └── shared
  standalone-project (regular project)
  ```

### 2. **Project Structure Enhancements**
New fields added to `Project` struct:
- `Depth` (int): Tracks nesting level (0 = category/top-level, 1 = nested project)
- `SubProjectCount` (int): Number of immediate sub-projects (determines if it's a category)

### 3. **UI Display Enhancements**
- **Category Headers**: Parent projects with children shown in gray, marked with ▼
- **Non-selectable**: Categories cannot be selected - only actual projects
- **Indentation**: Child projects indented with `  ` (2 spaces)
- **Sub-project Count**: Categories display count (e.g., `nested-dir (3)`)
- **Smart Detection**: Folders without sub-projects shown as regular selectable projects

### 4. **Configuration**
New config option in `DisplayConfig`:
```json
{
  "display": {
    "maxScanDepth": 1
  }
}
```

Set to `0` for completely flat view (no scanning of subdirectories).

## Example Output

```
▼ gamedev (2)              [category - not selectable]
    ▸ gamedev/unity-game     🎮 C#          develop
    ▸ gamedev/godot-project  🐍 Python      feature/ui
▼ webdev (3)               [category - not selectable]
    ▸ webdev/frontend        ⚛️  TypeScript  main
    ▸ webdev/backend         🟢 Node.js     main
    ▸ webdev/shared          🟨 JavaScript  main
▸ standalone               🔷 Go          main*
```

## Use Cases

1. **Organized Repos**: Group related projects under category folders
2. **Project Categories**: Separate by type (gamedev, webdev, mobile, etc.)
3. **Monorepos**: Show multiple sub-projects within a parent directory
4. **Mixed Structure**: Some projects standalone, others grouped

## Behavior

### With Nested Folders
When a folder contains sub-projects:
- Parent folder shown as **category header** (gray, non-selectable)
- Marked with ▼ instead of ▸
- Shows count of sub-projects: `(3)`
- Children shown indented below
- Only children are selectable

### Without Nested Folders
When a folder has no sub-projects:
- Shown as regular project (selectable)
- No special formatting
- Normal white text with ▸ marker

## Current Behavior

Nested/grouped projects are **always enabled** with a fixed scan depth of 1 level.

- Grouped view: always enabled
- Scan depth: 1 (one level of nesting)
- All directories within a group are shown (not just those with project indicators)

## Implementation Details

### Scanner Changes
- `Scan()` calls `scanWithGroups()` to detect parent/child relationships
- `findChildProjects()` finds all directories within a group (1 level deep)
- Each level respects exclusion patterns and hidden directory settings

### Performance Considerations
- Depth limit prevents infinite recursion
- Excluded directories (`.git`, `node_modules`, etc.) are skipped at all levels
- Only directories are scanned (files are ignored)

## Future Enhancements

Potential improvements:
- Collapsible tree view (expand/collapse parent projects)
- Recursive actions (apply operation to all sub-projects)
- Custom depth per project category
- Search filtering that respects hierarchy
