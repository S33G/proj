package actions

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/s33g/proj/internal/config"
	"github.com/s33g/proj/internal/git"
	"github.com/s33g/proj/internal/project"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func withTempCommand(t *testing.T, name string, fn func()) {
	t.Helper()

	dir := t.TempDir()
	cmdPath := filepath.Join(dir, name)
	writeFile(t, cmdPath, "#!/bin/sh\nexit 0")
	if err := os.Chmod(cmdPath, 0o755); err != nil {
		t.Fatalf("failed to chmod %s: %v", cmdPath, err)
	}

	origPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+origPath); err != nil {
		t.Fatalf("failed to set PATH: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("PATH", origPath)
	})

	fn()
}

func TestNewExecutor(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}

	if executor.config != cfg {
		t.Error("Executor config not set correctly")
	}
}

func TestCommandExists(t *testing.T) {
	// Test with a command that should exist
	if !commandExists("ls") {
		t.Error("ls should exist on Unix systems")
	}

	// Test with a command that shouldn't exist
	if commandExists("this-command-definitely-does-not-exist-12345") {
		t.Error("Non-existent command reported as existing")
	}
}

func TestDetectBuildArtifacts(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	tests := []struct {
		language string
		expected []string
	}{
		{"Go", []string{".DS_Store", "Thumbs.db", "bin", "dist", "vendor"}},
		{"Rust", []string{".DS_Store", "Thumbs.db", "target"}},
		{"JavaScript", []string{".DS_Store", "Thumbs.db", "node_modules", "dist", "build", ".next", "out", ".turbo", ".nuxt"}},
		{"Python", []string{".DS_Store", "Thumbs.db", "__pycache__", "*.pyc", ".pytest_cache", "dist", "build", "*.egg-info", ".venv", "venv"}},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			proj := &project.Project{Language: tt.language}
			artifacts := executor.detectBuildArtifacts(proj)

			if len(artifacts) != len(tt.expected) {
				t.Errorf("Expected %d artifacts, got %d", len(tt.expected), len(artifacts))
			}

			// Check if all expected artifacts are present
			for _, expected := range tt.expected {
				found := false
				for _, artifact := range artifacts {
					if artifact == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected artifact '%s' not found in results", expected)
				}
			}
		})
	}
}

func TestClean(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	// Create temp directory with some artifacts
	tmpDir := t.TempDir()

	// Create some build artifacts
	os.Mkdir(filepath.Join(tmpDir, "node_modules"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "dist"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dist", "test.js"), []byte("test"), 0644)

	proj := &project.Project{
		Path:     tmpDir,
		Language: "JavaScript",
	}

	result := executor.clean(proj)

	if !result.Success {
		t.Errorf("Clean failed: %s", result.Message)
	}

	// Verify artifacts were removed
	if _, err := os.Stat(filepath.Join(tmpDir, "node_modules")); !os.IsNotExist(err) {
		t.Error("node_modules should have been removed")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "dist")); !os.IsNotExist(err) {
		t.Error("dist should have been removed")
	}
}

func TestClean_NoArtifacts(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	tmpDir := t.TempDir()
	proj := &project.Project{
		Path:     tmpDir,
		Language: "Go",
	}

	result := executor.clean(proj)

	if !result.Success {
		t.Errorf("Clean should succeed even with no artifacts: %s", result.Message)
	}

	if result.Message != "No build artifacts found to clean" {
		t.Errorf("Unexpected message: %s", result.Message)
	}
}

func TestCdAction(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	proj := &project.Project{
		Path: "/test/path",
	}

	result := executor.Execute("cd", proj)

	if !result.Success {
		t.Error("cd action should succeed")
	}

	if result.CdPath != proj.Path {
		t.Errorf("Expected CdPath %s, got %s", proj.Path, result.CdPath)
	}
}

func TestUnknownAction(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	proj := &project.Project{}
	result := executor.Execute("unknown-action", proj)

	if result.Success {
		t.Error("Unknown action should fail")
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"echo hello world", []string{"echo", "hello", "world"}},
		{`cmd "arg with spaces" 'and more'`, []string{"cmd", "arg with spaces", "and more"}},
		{"single", []string{"single"}},
		{"", nil},
	}

	for _, tt := range tests {
		if got := parseCommand(tt.input); !reflect.DeepEqual(got, tt.expected) {
			t.Fatalf("parseCommand(%q) = %#v, want %#v", tt.input, got, tt.expected)
		}
	}
}

func TestGitLog_NonGitRepo(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	proj := &project.Project{
		IsGitRepo: false,
	}

	result := executor.gitLog(proj)

	if result.Success {
		t.Error("gitLog should fail on non-git repo")
	}
}

func TestDetectNodeTestCommandPrefersAvailableManagers(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	dir := t.TempDir()
	proj := &project.Project{Path: dir, Language: "JavaScript"}

	// Bun lock takes priority
	writeFile(t, filepath.Join(dir, "bun.lockb"), "")
	if cmd := executor.detectNodeTestCommand(proj); cmd.Args[0] != "bun" {
		t.Fatalf("expected bun test command, got %v", cmd.Args)
	}
	os.Remove(filepath.Join(dir, "bun.lockb"))

	// pnpm next
	writeFile(t, filepath.Join(dir, "pnpm-lock.yaml"), "")
	if cmd := executor.detectNodeTestCommand(proj); cmd.Args[0] != "pnpm" {
		t.Fatalf("expected pnpm test command, got %v", cmd.Args)
	}
	os.Remove(filepath.Join(dir, "pnpm-lock.yaml"))

	// yarn next
	writeFile(t, filepath.Join(dir, "yarn.lock"), "")
	if cmd := executor.detectNodeTestCommand(proj); cmd.Args[0] != "yarn" {
		t.Fatalf("expected yarn test command, got %v", cmd.Args)
	}
	os.Remove(filepath.Join(dir, "yarn.lock"))

	// fall back to npm when package.json present
	writeFile(t, filepath.Join(dir, "package.json"), `{}`)
	if cmd := executor.detectNodeTestCommand(proj); cmd.Args[0] != "npm" {
		t.Fatalf("expected npm test command, got %v", cmd.Args)
	}
}

func TestDetectNodeInstallCommandMatchesLocks(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)
	dir := t.TempDir()
	proj := &project.Project{Path: dir, Language: "JavaScript"}

	writeFile(t, filepath.Join(dir, "bun.lockb"), "")
	if cmd := executor.detectNodeInstallCommand(proj); cmd.Args[0] != "bun" {
		t.Fatalf("expected bun install, got %v", cmd.Args)
	}
	os.Remove(filepath.Join(dir, "bun.lockb"))

	writeFile(t, filepath.Join(dir, "pnpm-lock.yaml"), "")
	if cmd := executor.detectNodeInstallCommand(proj); cmd.Args[0] != "pnpm" {
		t.Fatalf("expected pnpm install, got %v", cmd.Args)
	}
	os.Remove(filepath.Join(dir, "pnpm-lock.yaml"))

	writeFile(t, filepath.Join(dir, "yarn.lock"), "")
	if cmd := executor.detectNodeInstallCommand(proj); cmd.Args[0] != "yarn" {
		t.Fatalf("expected yarn install, got %v", cmd.Args)
	}
	os.Remove(filepath.Join(dir, "yarn.lock"))

	if cmd := executor.detectNodeInstallCommand(proj); cmd.Args[0] != "npm" {
		t.Fatalf("expected npm install fallback, got %v", cmd.Args)
	}
}

func TestDetectPythonCommands(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)
	dir := t.TempDir()
	proj := &project.Project{Path: dir, Language: "Python"}

	// Prefer pytest when available in PATH
	withTempCommand(t, "pytest", func() {
		cmd := executor.detectPythonTestCommand(proj)
		if cmd.Args[0] != "pytest" {
			t.Fatalf("expected pytest command, got %v", cmd.Args)
		}
	})

	// Install command priorities
	writeFile(t, filepath.Join(dir, "Pipfile"), "")
	if cmd := executor.detectPythonInstallCommand(proj); cmd.Args[0] != "pipenv" {
		t.Fatalf("expected pipenv install, got %v", cmd.Args)
	}
	os.Remove(filepath.Join(dir, "Pipfile"))

	writeFile(t, filepath.Join(dir, "poetry.lock"), "")
	if cmd := executor.detectPythonInstallCommand(proj); cmd.Args[0] != "poetry" {
		t.Fatalf("expected poetry install, got %v", cmd.Args)
	}
	os.Remove(filepath.Join(dir, "poetry.lock"))

	writeFile(t, filepath.Join(dir, "requirements.txt"), "")
	cmd := executor.detectPythonInstallCommand(proj)
	if cmd.Args[0] != "pip" {
		t.Fatalf("expected pip install, got %v", cmd.Args)
	}
	expectedArgs := []string{"pip", "install", "-r", "requirements.txt"}
	if !reflect.DeepEqual(cmd.Args, expectedArgs) {
		t.Fatalf("expected %v, got %v", expectedArgs, cmd.Args)
	}
}

func TestGitLog(t *testing.T) {
	if !git.IsInstalled() {
		t.Skip("git not installed")
	}

	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	// Create a git repo
	tmpDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("Failed to init git repo")
	}

	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create a commit
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	if err := exec.Command("git", "-C", tmpDir, "commit", "-m", "test commit").Run(); err != nil {
		t.Skip("Failed to create commit")
	}

	proj := &project.Project{
		Path:      tmpDir,
		IsGitRepo: true,
	}

	result := executor.gitLog(proj)

	if !result.Success {
		t.Errorf("gitLog failed: %s", result.Message)
	}

	if result.Message == "" {
		t.Error("gitLog should return a message")
	}
}

func TestDetectNodeTestCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"bun", []string{"bun.lockb", "package.json"}, "bun"},
		{"pnpm", []string{"pnpm-lock.yaml", "package.json"}, "pnpm"},
		{"yarn", []string{"yarn.lock", "package.json"}, "yarn"},
		{"npm", []string{"package.json"}, "npm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for _, file := range tt.files {
				os.WriteFile(filepath.Join(tmpDir, file), []byte("{}"), 0644)
			}

			proj := &project.Project{Path: tmpDir}
			cmd := executor.detectNodeTestCommand(proj)

			if cmd == nil {
				t.Fatal("Command should not be nil")
			}

			if cmd.Args[0] != tt.expected {
				t.Errorf("Expected command %s, got %s", tt.expected, cmd.Args[0])
			}
		})
	}
}

func TestDetectNodeInstallCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	executor := NewExecutor(cfg)

	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"bun", []string{"bun.lockb"}, "bun"},
		{"pnpm", []string{"pnpm-lock.yaml"}, "pnpm"},
		{"yarn", []string{"yarn.lock"}, "yarn"},
		{"npm", []string{}, "npm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for _, file := range tt.files {
				os.WriteFile(filepath.Join(tmpDir, file), []byte(""), 0644)
			}

			proj := &project.Project{Path: tmpDir}
			cmd := executor.detectNodeInstallCommand(proj)

			if cmd == nil {
				t.Fatal("Command should not be nil")
			}

			if cmd.Args[0] != tt.expected {
				t.Errorf("Expected command %s, got %s", tt.expected, cmd.Args[0])
			}
		})
	}
}
