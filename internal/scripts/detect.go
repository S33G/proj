package scripts

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Script represents a runnable script/command
type Script struct {
	ID      string // Unique identifier for the script
	Name    string // Display name
	Command string // The command to run
	Desc    string // Description
	Source  string // Where the script came from (package.json, Makefile, etc)
}

// Detect detects available scripts for a project
func Detect(projectPath, language string) []Script {
	var scripts []Script

	// Language-specific detection
	switch language {
	case "JavaScript", "TypeScript":
		scripts = append(scripts, detectPackageJSON(projectPath)...)
	case "Go":
		scripts = append(scripts, detectGoScripts(projectPath)...)
	case "Rust":
		scripts = append(scripts, detectCargoScripts(projectPath)...)
	case "Python":
		scripts = append(scripts, detectPythonScripts(projectPath)...)
	case "Ruby":
		scripts = append(scripts, detectRubyScripts(projectPath)...)
	}

	// Universal detection (Makefile, shell scripts)
	scripts = append(scripts, detectMakefile(projectPath)...)
	scripts = append(scripts, detectShellScripts(projectPath)...)
	scripts = append(scripts, detectJustfile(projectPath)...)

	return scripts
}

// detectPackageJSON extracts scripts from package.json
func detectPackageJSON(projectPath string) []Script {
	pkgPath := filepath.Join(projectPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	// Detect package manager
	pm := detectPackageManager(projectPath)

	var scripts []Script
	for name, cmd := range pkg.Scripts {
		// Skip some common internal scripts
		if shouldSkipNpmScript(name) {
			continue
		}

		scripts = append(scripts, Script{
			ID:      "npm-" + name,
			Name:    name,
			Command: pm + " run " + name,
			Desc:    truncateCommand(cmd, 50),
			Source:  "package.json",
		})
	}

	return scripts
}

// detectPackageManager determines which package manager to use
func detectPackageManager(projectPath string) string {
	if _, err := os.Stat(filepath.Join(projectPath, "bun.lockb")); err == nil {
		return "bun"
	}
	if _, err := os.Stat(filepath.Join(projectPath, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(projectPath, "yarn.lock")); err == nil {
		return "yarn"
	}
	return "npm"
}

// shouldSkipNpmScript returns true for scripts that shouldn't be shown
func shouldSkipNpmScript(name string) bool {
	skip := []string{
		"preinstall", "postinstall", "prepublish", "prepublishOnly",
		"prepack", "postpack", "prepare", "preshrinkwrap", "shrinkwrap",
		"postshrinkwrap", "preversion", "postversion", "preuninstall",
		"postuninstall", "prestop", "poststop", "prestart", "poststart",
		"prerestart", "postrestart", "pretest", "posttest",
	}
	for _, s := range skip {
		if name == s {
			return true
		}
	}
	return false
}

// detectMakefile extracts targets from Makefile
func detectMakefile(projectPath string) []Script {
	makefilePath := filepath.Join(projectPath, "Makefile")
	file, err := os.Open(makefilePath)
	if err != nil {
		// Try GNUmakefile
		makefilePath = filepath.Join(projectPath, "GNUmakefile")
		file, err = os.Open(makefilePath)
		if err != nil {
			return nil
		}
	}
	defer func() { _ = file.Close() }()

	// Match target definitions like "target:" but not ".PHONY:" or variable assignments
	targetRegex := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)
	// Match comment above target for description
	commentRegex := regexp.MustCompile(`^#\s*(.+)`)

	var scripts []Script
	var lastComment string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for comment
		if matches := commentRegex.FindStringSubmatch(line); matches != nil {
			lastComment = matches[1]
			continue
		}

		// Check for target
		if matches := targetRegex.FindStringSubmatch(line); matches != nil {
			target := matches[1]

			// Skip internal targets (starting with _ or .)
			if strings.HasPrefix(target, "_") || strings.HasPrefix(target, ".") {
				lastComment = ""
				continue
			}

			// Skip some common meta-targets
			if target == "all" || target == "default" {
				lastComment = ""
				continue
			}

			scripts = append(scripts, Script{
				ID:      "make-" + target,
				Name:    target,
				Command: "make " + target,
				Desc:    lastComment,
				Source:  "Makefile",
			})
			lastComment = ""
		} else if line != "" && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
			// Non-empty line that's not a recipe, reset comment
			lastComment = ""
		}
	}

	return scripts
}

// detectJustfile extracts recipes from justfile
func detectJustfile(projectPath string) []Script {
	justfilePath := filepath.Join(projectPath, "justfile")
	file, err := os.Open(justfilePath)
	if err != nil {
		// Try Justfile (capitalized)
		justfilePath = filepath.Join(projectPath, "Justfile")
		file, err = os.Open(justfilePath)
		if err != nil {
			return nil
		}
	}
	defer func() { _ = file.Close() }()

	// Match recipe definitions
	recipeRegex := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)
	commentRegex := regexp.MustCompile(`^#\s*(.+)`)

	var scripts []Script
	var lastComment string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := commentRegex.FindStringSubmatch(line); matches != nil {
			lastComment = matches[1]
			continue
		}

		if matches := recipeRegex.FindStringSubmatch(line); matches != nil {
			recipe := matches[1]
			if !strings.HasPrefix(recipe, "_") {
				scripts = append(scripts, Script{
					ID:      "just-" + recipe,
					Name:    recipe,
					Command: "just " + recipe,
					Desc:    lastComment,
					Source:  "justfile",
				})
			}
			lastComment = ""
		} else if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			lastComment = ""
		}
	}

	return scripts
}

// detectShellScripts finds executable shell scripts in common locations
func detectShellScripts(projectPath string) []Script {
	var scripts []Script

	// Check common script directories
	scriptDirs := []string{
		"scripts",
		"script",
		"bin",
		".scripts",
	}

	for _, dir := range scriptDirs {
		dirPath := filepath.Join(projectPath, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			// Only include shell scripts
			if !isShellScript(name) {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Check if executable
			if info.Mode()&0111 == 0 {
				continue
			}

			scriptPath := filepath.Join(dir, name)
			baseName := strings.TrimSuffix(name, filepath.Ext(name))

			scripts = append(scripts, Script{
				ID:      "script-" + baseName,
				Name:    baseName,
				Command: "./" + scriptPath,
				Desc:    "Shell script",
				Source:  dir + "/",
			})
		}
	}

	// Also check for common root-level scripts
	rootScripts := []string{"run.sh", "build.sh", "deploy.sh", "setup.sh", "start.sh", "test.sh"}
	for _, name := range rootScripts {
		scriptPath := filepath.Join(projectPath, name)
		info, err := os.Stat(scriptPath)
		if err != nil {
			continue
		}
		if info.Mode()&0111 == 0 {
			continue
		}

		baseName := strings.TrimSuffix(name, ".sh")
		scripts = append(scripts, Script{
			ID:      "script-" + baseName,
			Name:    baseName,
			Command: "./" + name,
			Desc:    "Shell script",
			Source:  "root",
		})
	}

	return scripts
}

// isShellScript checks if a filename looks like a shell script
func isShellScript(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	shellExts := []string{".sh", ".bash", ".zsh"}
	for _, e := range shellExts {
		if ext == e {
			return true
		}
	}
	// No extension could also be a shell script
	return ext == ""
}

// detectGoScripts detects Go-specific scripts
func detectGoScripts(projectPath string) []Script {
	var scripts []Script

	// Check for go.mod to confirm it's a Go project
	if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); err != nil {
		return nil
	}

	// Common Go commands
	scripts = append(scripts,
		Script{
			ID:      "go-build",
			Name:    "build",
			Command: "go build ./...",
			Desc:    "Build all packages",
			Source:  "go",
		},
		Script{
			ID:      "go-test",
			Name:    "test",
			Command: "go test ./...",
			Desc:    "Run all tests",
			Source:  "go",
		},
		Script{
			ID:      "go-vet",
			Name:    "vet",
			Command: "go vet ./...",
			Desc:    "Run go vet",
			Source:  "go",
		},
		Script{
			ID:      "go-fmt",
			Name:    "fmt",
			Command: "go fmt ./...",
			Desc:    "Format code",
			Source:  "go",
		},
	)

	// Check for main.go to add run command
	if _, err := os.Stat(filepath.Join(projectPath, "main.go")); err == nil {
		scripts = append(scripts, Script{
			ID:      "go-run",
			Name:    "run",
			Command: "go run .",
			Desc:    "Run main package",
			Source:  "go",
		})
	} else if _, err := os.Stat(filepath.Join(projectPath, "cmd")); err == nil {
		scripts = append(scripts, Script{
			ID:      "go-run",
			Name:    "run",
			Command: "go run ./cmd/...",
			Desc:    "Run cmd packages",
			Source:  "go",
		})
	}

	return scripts
}

// detectCargoScripts detects Rust/Cargo scripts
func detectCargoScripts(projectPath string) []Script {
	if _, err := os.Stat(filepath.Join(projectPath, "Cargo.toml")); err != nil {
		return nil
	}

	return []Script{
		{
			ID:      "cargo-build",
			Name:    "build",
			Command: "cargo build",
			Desc:    "Build the project",
			Source:  "cargo",
		},
		{
			ID:      "cargo-run",
			Name:    "run",
			Command: "cargo run",
			Desc:    "Run the project",
			Source:  "cargo",
		},
		{
			ID:      "cargo-test",
			Name:    "test",
			Command: "cargo test",
			Desc:    "Run tests",
			Source:  "cargo",
		},
		{
			ID:      "cargo-check",
			Name:    "check",
			Command: "cargo check",
			Desc:    "Check for errors",
			Source:  "cargo",
		},
		{
			ID:      "cargo-clippy",
			Name:    "clippy",
			Command: "cargo clippy",
			Desc:    "Run linter",
			Source:  "cargo",
		},
		{
			ID:      "cargo-fmt",
			Name:    "fmt",
			Command: "cargo fmt",
			Desc:    "Format code",
			Source:  "cargo",
		},
	}
}

// detectPythonScripts detects Python-specific scripts
func detectPythonScripts(projectPath string) []Script {
	var scripts []Script

	// Check for pyproject.toml (modern Python)
	if _, err := os.Stat(filepath.Join(projectPath, "pyproject.toml")); err == nil {
		// Check for poetry
		if _, err := os.Stat(filepath.Join(projectPath, "poetry.lock")); err == nil {
			scripts = append(scripts,
				Script{ID: "poetry-install", Name: "install", Command: "poetry install", Desc: "Install dependencies", Source: "poetry"},
				Script{ID: "poetry-run", Name: "run", Command: "poetry run python", Desc: "Run with poetry", Source: "poetry"},
				Script{ID: "poetry-test", Name: "test", Command: "poetry run pytest", Desc: "Run tests", Source: "poetry"},
			)
		}
	}

	// Check for setup.py
	if _, err := os.Stat(filepath.Join(projectPath, "setup.py")); err == nil {
		scripts = append(scripts,
			Script{ID: "py-install", Name: "install", Command: "pip install -e .", Desc: "Install in dev mode", Source: "setup.py"},
		)
	}

	// Check for requirements.txt
	if _, err := os.Stat(filepath.Join(projectPath, "requirements.txt")); err == nil {
		scripts = append(scripts,
			Script{ID: "pip-install", Name: "install-deps", Command: "pip install -r requirements.txt", Desc: "Install requirements", Source: "pip"},
		)
	}

	// Common Python scripts
	if _, err := os.Stat(filepath.Join(projectPath, "manage.py")); err == nil {
		// Django project
		scripts = append(scripts,
			Script{ID: "django-runserver", Name: "runserver", Command: "python manage.py runserver", Desc: "Start Django server", Source: "django"},
			Script{ID: "django-migrate", Name: "migrate", Command: "python manage.py migrate", Desc: "Run migrations", Source: "django"},
			Script{ID: "django-shell", Name: "shell", Command: "python manage.py shell", Desc: "Django shell", Source: "django"},
		)
	}

	// Check for pytest
	if _, err := os.Stat(filepath.Join(projectPath, "pytest.ini")); err == nil {
		scripts = append(scripts,
			Script{ID: "pytest", Name: "test", Command: "pytest", Desc: "Run pytest", Source: "pytest"},
		)
	} else if _, err := os.Stat(filepath.Join(projectPath, "tests")); err == nil {
		scripts = append(scripts,
			Script{ID: "pytest", Name: "test", Command: "pytest", Desc: "Run tests", Source: "python"},
		)
	}

	return scripts
}

// detectRubyScripts detects Ruby-specific scripts
func detectRubyScripts(projectPath string) []Script {
	var scripts []Script

	// Check for Gemfile (Bundler)
	if _, err := os.Stat(filepath.Join(projectPath, "Gemfile")); err == nil {
		scripts = append(scripts,
			Script{ID: "bundle-install", Name: "install", Command: "bundle install", Desc: "Install gems", Source: "bundler"},
		)
	}

	// Check for Rakefile
	if _, err := os.Stat(filepath.Join(projectPath, "Rakefile")); err == nil {
		scripts = append(scripts, detectRakeTasks(projectPath)...)
	}

	// Check for Rails
	if _, err := os.Stat(filepath.Join(projectPath, "config", "application.rb")); err == nil {
		scripts = append(scripts,
			Script{ID: "rails-server", Name: "server", Command: "rails server", Desc: "Start Rails server", Source: "rails"},
			Script{ID: "rails-console", Name: "console", Command: "rails console", Desc: "Rails console", Source: "rails"},
			Script{ID: "rails-migrate", Name: "migrate", Command: "rails db:migrate", Desc: "Run migrations", Source: "rails"},
		)
	}

	return scripts
}

// detectRakeTasks extracts common rake tasks
func detectRakeTasks(projectPath string) []Script {
	// Just return common rake tasks without parsing
	return []Script{
		{ID: "rake-test", Name: "test", Command: "rake test", Desc: "Run tests", Source: "rake"},
		{ID: "rake-default", Name: "default", Command: "rake", Desc: "Run default task", Source: "rake"},
	}
}

// truncateCommand truncates a command string for display
func truncateCommand(cmd string, maxLen int) string {
	// Remove newlines
	cmd = strings.ReplaceAll(cmd, "\n", " ")
	cmd = strings.ReplaceAll(cmd, "\r", "")

	if len(cmd) <= maxLen {
		return cmd
	}
	return cmd[:maxLen-3] + "..."
}
