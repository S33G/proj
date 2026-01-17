# Plugin Development Guide

## Overview

The `proj` plugin system allows you to extend the functionality of the project navigator with custom actions, language detectors, and more. Plugins communicate with the main application via JSON-RPC 2.0 over stdin/stdout.

## Plugin Architecture

### Communication Protocol

Plugins use JSON-RPC 2.0 for communication:
- **Transport**: stdin/stdout
- **Format**: Newline-delimited JSON
- **Methods**: `init`, `actions`, `executeAction`, `languages`, `shutdown`

### Plugin Lifecycle

1. **Discovery**: Plugin directories are scanned in `~/.config/proj/plugins/`
2. **Loading**: Manifest (`plugin.json`) is read
3. **Initialization**: Plugin executable is started
4. **Init Call**: `init` method is called with configuration
5. **Operation**: Plugin responds to method calls
6. **Shutdown**: `shutdown` method is called, process exits

## Creating a Plugin

### 1. Directory Structure

```
~/.config/proj/plugins/
└── my-plugin/
    ├── plugin.json      # Manifest
    └── my-plugin        # Executable (any language)
```

For development, plugins can also be placed in the project's `plugins/` directory.

### 2. Plugin Manifest

Create a `plugin.json` file:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My custom plugin",
  "executable": "my-plugin",
  "capabilities": ["actions"],
  "config": {}
}
```

**Fields:**
- `name` (required): Unique plugin identifier
- `version` (required): Semantic version
- `description` (optional): Human-readable description
- `executable` (required): Name of the executable file
- `capabilities` (required): Array of capabilities (`actions`, `languages`)
- `config` (optional): Default configuration

### 3. Implement JSON-RPC Handler

Your plugin must implement these methods:

#### `init` - Initialize Plugin

**Params:**
```json
{
  "config": {
    "key": "value"
  }
}
```

**Response:**
```json
{
  "success": true
}
```

#### `actions` - Get Available Actions

**Params:**
```json
{
  "name": "project-name",
  "path": "/path/to/project",
  "language": "Go",
  "gitBranch": "main",
  "gitDirty": false,
  "isGitRepo": true
}
```

**Response:**
```json
[
  {
    "id": "my-action",
    "label": "My Action",
    "description": "Does something useful",
    "icon": "🚀",
    "priority": 100
  }
]
```

#### `executeAction` - Execute an Action

**Params:**
```json
{
  "action": "my-action",
  "project": {
    "name": "project-name",
    "path": "/path/to/project",
    "language": "Go",
    "gitBranch": "main",
    "gitDirty": false,
    "isGitRepo": true
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Action completed successfully",
  "cdPath": "",
  "execCmd": []
}
```

**Response Fields:**
- `success`: Whether the action succeeded
- `message`: Message to display to user
- `cdPath`: (Optional) Path to change to and exit
- `execCmd`: (Optional) Command to exec and replace shell

#### `shutdown` - Graceful Shutdown

**Params:** None

**Response:**
```json
{
  "success": true
}
```

### 4. Example Plugin (Go)

See `plugins/example/main.go` for a complete working example.

## Configuration

### Enable Plugin

Edit `~/.config/proj/config.json`:

```json
{
  "plugins": {
    "enabled": ["my-plugin"],
    "config": {
      "my-plugin": {
        "key": "value"
      }
    }
  }
}
```

### Plugin Configuration

Plugin-specific configuration is passed during the `init` call. Access it from the `config` parameter.

## Building Plugins in Different Languages

### Go

```go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
)

type RPCRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
    ID      int             `json:"id"`
}

type RPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    Result  interface{} `json:"result,omitempty"`
    Error   *RPCError   `json:"error,omitempty"`
    ID      int         `json:"id"`
}

type RPCError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        var req RPCRequest
        json.Unmarshal(scanner.Bytes(), &req)
        
        resp := handleRequest(&req)
        
        respData, _ := json.Marshal(resp)
        fmt.Println(string(respData))
    }
}
```

### Python

```python
#!/usr/bin/env python3
import sys
import json

def handle_request(req):
    method = req.get('method')
    params = req.get('params', {})
    
    if method == 'init':
        return {'success': True}
    elif method == 'actions':
        return [{
            'id': 'my-action',
            'label': 'My Action',
            'description': 'Does something',
            'icon': '🚀',
            'priority': 100
        }]
    elif method == 'executeAction':
        return {
            'success': True,
            'message': 'Action executed'
        }
    
    return None

for line in sys.stdin:
    req = json.loads(line)
    result = handle_request(req)
    
    resp = {
        'jsonrpc': '2.0',
        'result': result,
        'id': req.get('id')
    }
    
    print(json.dumps(resp))
    sys.stdout.flush()
```

### Node.js

```javascript
#!/usr/bin/env node
const readline = require('readline');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

function handleRequest(req) {
  const { method, params } = req;
  
  if (method === 'init') {
    return { success: true };
  } else if (method === 'actions') {
    return [{
      id: 'my-action',
      label: 'My Action',
      description: 'Does something',
      icon: '🚀',
      priority: 100
    }];
  } else if (method === 'executeAction') {
    return {
      success: true,
      message: 'Action executed'
    };
  }
  
  return null;
}

rl.on('line', (line) => {
  const req = JSON.parse(line);
  const result = handleRequest(req);
  
  const resp = {
    jsonrpc: '2.0',
    result,
    id: req.id
  };
  
  console.log(JSON.stringify(resp));
});
```

## Testing Plugins

### Manual Testing

1. Place plugin in `~/.config/proj/plugins/my-plugin/`
2. Enable in config: `proj --config`
3. Run `proj` and verify actions appear
4. Test plugin executable directly:

```bash
echo '{"jsonrpc":"2.0","method":"init","params":{"config":{}},"id":1}' | ./my-plugin
```

### Integration Testing

Create test projects and verify plugin actions appear correctly in the TUI.

## Best Practices

1. **Error Handling**: Always return proper JSON-RPC errors
2. **Performance**: Keep action detection fast (<100ms)
3. **Resource Cleanup**: Implement shutdown properly
4. **Logging**: Write debug info to stderr, not stdout
5. **Versioning**: Use semantic versioning
6. **Security**: Validate all inputs
7. **Documentation**: Document configuration options

## Capabilities

### Actions

Plugins with the `actions` capability can:
- Provide custom actions for projects
- Execute commands
- Trigger directory changes
- Replace the shell with a new command

### Languages (Future)

Plugins with the `languages` capability can:
- Detect custom programming languages
- Provide language-specific metadata

## Troubleshooting

### Plugin Not Loading

1. Check `plugin.json` is valid JSON
2. Verify executable has execute permissions: `chmod +x my-plugin`
3. Enable plugin in config: `"enabled": ["my-plugin"]`
4. Check stderr for error messages

### Actions Not Appearing

1. Verify plugin returns actions in `actions` method
2. Check action structure matches expected format
3. Ensure `actions` capability is declared

### Plugin Crashes

1. Check stderr for error messages
2. Test plugin executable directly
3. Verify JSON-RPC responses are valid
4. Check for resource leaks (file descriptors, etc.)

## API Reference

### Types

#### Project
```typescript
{
  name: string
  path: string
  language: string
  gitBranch: string
  gitDirty: boolean
  isGitRepo: boolean
}
```

#### Action
```typescript
{
  id: string
  label: string
  description: string
  icon: string
  priority: number
}
```

#### ActionResult
```typescript
{
  success: boolean
  message: string
  cdPath?: string
  execCmd?: string[]
}
```

## Examples

See `plugins/example/` for a complete working example that demonstrates:
- JSON-RPC request handling
- Action detection
- Action execution
- Configuration handling
- Proper shutdown
