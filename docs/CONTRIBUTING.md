# Contributing to proj

Thank you for your interest in contributing to proj! This document provides guidelines and information for contributors.

## Code of Conduct

Please be respectful and considerate in all interactions. We're all here to build something useful together.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Make
- Git

### Setting Up the Development Environment

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/proj.git
   cd proj
   ```

2. **Install dependencies:**
   ```bash
   make deps
   ```

3. **Build the project:**
   ```bash
   make build
   ```

4. **Run tests:**
   ```bash
   make test
   ```

5. **Run in development mode:**
   ```bash
   make dev
   ```

## Development Workflow

### Branching Strategy

- `main` - Stable release branch
- Feature branches - `feat/description`
- Bug fix branches - `fix/description`
- Documentation - `docs/description`

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Make your changes** following the coding guidelines below.

3. **Run tests and linting:**
   ```bash
   make test
   make lint
   make fmt
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add my new feature"
   ```

5. **Push and create a pull request:**
   ```bash
   git push origin feat/my-feature
   ```

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

**Examples:**
```
feat: add support for Ruby language detection

fix: resolve crash when git directory is missing

docs: update installation instructions

refactor: simplify project scanning logic
```

## Project Structure

```
proj/
├── cmd/proj/              # CLI entrypoint
│   └── main.go            # Main function, CLI parsing
├── internal/              # Private application code
│   ├── app/               # Main TUI application model
│   ├── actions/           # Built-in action implementations
│   ├── config/            # Configuration loading/saving
│   ├── git/               # Git operations
│   ├── language/          # Language detection
│   ├── project/           # Project scanning
│   └── tui/               # TUI components
│       ├── styles.go      # Lip Gloss styles
│       ├── keys.go        # Keyboard shortcuts
│       └── views/         # View components
├── pkg/                   # Public packages (importable by others)
│   └── plugin/            # Plugin system
├── plugins/               # Example plugins
│   └── example/           # Example plugin implementation
├── scripts/               # Build and installation scripts
├── docs/                  # Documentation
├── Makefile               # Build automation
└── go.mod                 # Go module definition
```

## Coding Guidelines

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use meaningful variable and function names
- Add comments for exported functions and types

### File Organization

- One primary type per file when possible
- Tests in `*_test.go` files alongside source
- Keep files focused and under 500 lines when reasonable

### Error Handling

- Always handle errors explicitly
- Use descriptive error messages
- Wrap errors with context using `fmt.Errorf("context: %w", err)`

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

### Testing

- Write tests for new functionality
- Use table-driven tests where appropriate
- Aim for meaningful coverage, not 100%

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty input", "", ""},
        {"normal input", "hello", "HELLO"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

## Adding Features

### Adding a New Language

1. Edit `internal/language/detect.go`
2. Add detection logic in `Detect()` function
3. Add to the `languages` slice if using marker files
4. Add tests in `detect_test.go`

```go
// In detect.go
{name: "MyLang", markers: []string{"mylang.config", "*.mylang"}},
```

### Adding a New Editor

1. Edit `internal/config/config.go`
2. Add to `DefaultConfig()` aliases
3. Update `docs/CONFIG.md`

```go
// In config.go, DefaultConfig()
"myeditor": {"myeditor", "--flag"},
```

### Adding a New Action

1. Edit `internal/actions/actions.go`
2. Add action handler in `Execute()` method
3. Add to `internal/tui/views/action_menu.go` in `DefaultActions()`
4. Add tests

### Adding a Plugin Capability

1. Edit `pkg/plugin/types.go` for new types
2. Update `pkg/plugin/loader.go` for new methods
3. Update `docs/PLUGINS.md`
4. Add example in `plugins/example/main.go`

## Testing

### Running Tests

```bash
# All tests
make test

# Specific package
go test ./internal/language/

# With verbose output
go test -v ./...

# With coverage
make test-coverage
```

### Writing Tests

- Place tests in `*_test.go` files
- Use `t.Run()` for subtests
- Use `t.Helper()` in helper functions
- Use `t.TempDir()` for temporary directories

## Documentation

- Update `README.md` for user-facing changes
- Update relevant docs in `docs/` directory
- Add comments to exported types and functions
- Include examples where helpful

## Pull Request Process

1. **Ensure all tests pass:**
   ```bash
   make test
   ```

2. **Run the linter:**
   ```bash
   make lint
   ```

3. **Update documentation** if needed.

4. **Create the pull request** with:
   - Clear title describing the change
   - Description of what and why
   - Reference to any related issues

5. **Respond to feedback** and make requested changes.

### PR Checklist

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Code is formatted (`make fmt`)
- [ ] Documentation updated if needed
- [ ] Commit messages follow convention
- [ ] No unrelated changes included

## Reporting Issues

### Bug Reports

Include:
- proj version (`proj --version`)
- Go version (`go version`)
- Operating system and version
- Steps to reproduce
- Expected vs actual behavior
- Error messages or logs

### Feature Requests

Include:
- Clear description of the feature
- Use case / why it would be helpful
- Any implementation ideas (optional)

## Releases

Maintainers: See [RELEASE.md](RELEASE.md) for the release process and version management guidelines.

## Getting Help

- Open an issue for questions
- Check existing issues and documentation first
- Be patient and respectful

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to proj!
