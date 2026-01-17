package actions

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/s33g/proj/internal/config"
	"github.com/s33g/proj/internal/git"
	"github.com/s33g/proj/internal/project"
)

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
