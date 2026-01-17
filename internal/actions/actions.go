package actions

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/s33g/proj/internal/config"
	"github.com/s33g/proj/internal/git"
	"github.com/s33g/proj/internal/project"
)

// Result represents the result of an action
type Result struct {
	Success bool
	Message string
	CdPath  string
	ExecCmd []string
}

// Executor executes actions on projects
type Executor struct {
	config *config.Config
}

// NewExecutor creates a new action executor
func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{config: cfg}
}

// Execute executes an action on a project
func (e *Executor) Execute(actionID string, proj *project.Project) Result {
	switch actionID {
	case "open-editor":
		return e.openEditor(proj)
	case "cd":
		return Result{Success: true, CdPath: proj.Path}
	case "git-log":
		return e.gitLog(proj)
	case "git-pull":
		return e.gitPull(proj)
	case "git-branch":
		return e.gitBranch(proj)
	case "git-init":
		return e.gitInit(proj)
	case "run-tests":
		return e.runTests(proj)
	case "install-deps":
		return e.installDeps(proj)
	case "clean":
		return e.clean(proj)
	default:
		return Result{Success: false, Message: "Unknown action: " + actionID}
	}
}

// openEditor opens the project in the configured editor
func (e *Executor) openEditor(proj *project.Project) Result {
	editorCmd := e.config.Editor.Default
	
	// Get editor command and args from aliases
	cmdArgs, ok := e.config.Editor.Aliases[editorCmd]
	if !ok || len(cmdArgs) == 0 {
		// Fallback to just the editor name
		cmdArgs = []string{editorCmd}
	}

	// Check if editor exists
	if !commandExists(cmdArgs[0]) {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Editor '%s' not found in PATH", cmdArgs[0]),
		}
	}

	// Prepare full command with project path
	fullArgs := append(cmdArgs[1:], proj.Path)
	
	// Execute in background
	cmd := exec.Command(cmdArgs[0], fullArgs...)
	if err := cmd.Start(); err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Failed to open editor: %v", err),
		}
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("Opened in %s", editorCmd),
	}
}

// gitLog shows git log for the project
func (e *Executor) gitLog(proj *project.Project) Result {
	if !proj.IsGitRepo {
		return Result{Success: false, Message: "Not a git repository"}
	}

	logs, err := git.Log(proj.Path, 20)
	if err != nil {
		return Result{Success: false, Message: fmt.Sprintf("Failed to get git log: %v", err)}
	}

	if len(logs) == 0 {
		return Result{Success: true, Message: "No commits yet"}
	}

	message := "Recent commits:\n\n" + strings.Join(logs, "\n")
	return Result{Success: true, Message: message}
}

// gitPull pulls latest changes
func (e *Executor) gitPull(proj *project.Project) Result {
	if !proj.IsGitRepo {
		return Result{Success: false, Message: "Not a git repository"}
	}

	output, err := git.Pull(proj.Path)
	if err != nil {
		return Result{Success: false, Message: fmt.Sprintf("Pull failed: %v\n%s", err, output)}
	}

	return Result{Success: true, Message: output}
}

// gitBranch switches git branch
func (e *Executor) gitBranch(proj *project.Project) Result {
	if !proj.IsGitRepo {
		return Result{Success: false, Message: "Not a git repository"}
	}

	branches, err := git.GetBranches(proj.Path)
	if err != nil {
		return Result{Success: false, Message: fmt.Sprintf("Failed to get branches: %v", err)}
	}

	if len(branches) == 0 {
		return Result{Success: false, Message: "No branches found"}
	}

	// For now, just show available branches
	// TODO: Add interactive branch selection in TUI
	message := "Available branches:\n\n" + strings.Join(branches, "\n")
	return Result{Success: true, Message: message}
}

// gitInit initializes a new git repository
func (e *Executor) gitInit(proj *project.Project) Result {
	if proj.IsGitRepo {
		return Result{Success: false, Message: "Already a git repository"}
	}

	output, err := git.Init(proj.Path)
	if err != nil {
		return Result{Success: false, Message: fmt.Sprintf("Git init failed: %v\n%s", err, output)}
	}

	// Update the project's git status
	proj.IsGitRepo = true

	return Result{Success: true, Message: output}
}

// runTests runs project tests
func (e *Executor) runTests(proj *project.Project) Result {
	// Detect test command based on language
	var cmd *exec.Cmd

	switch proj.Language {
	case "Go":
		cmd = exec.Command("go", "test", "./...")
	case "Rust":
		cmd = exec.Command("cargo", "test")
	case "JavaScript", "TypeScript":
		cmd = e.detectNodeTestCommand(proj)
	case "Python":
		cmd = e.detectPythonTestCommand(proj)
	default:
		return Result{
			Success: false,
			Message: fmt.Sprintf("Don't know how to run tests for %s projects", proj.Language),
		}
	}

	if cmd == nil {
		return Result{Success: false, Message: "No test command found"}
	}

	cmd.Dir = proj.Path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Tests failed:\n%s", output),
		}
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("Tests passed:\n%s", output),
	}
}

// installDeps installs project dependencies
func (e *Executor) installDeps(proj *project.Project) Result {
	var cmd *exec.Cmd

	switch proj.Language {
	case "Go":
		cmd = exec.Command("go", "mod", "download")
	case "Rust":
		cmd = exec.Command("cargo", "fetch")
	case "JavaScript", "TypeScript":
		cmd = e.detectNodeInstallCommand(proj)
	case "Python":
		cmd = e.detectPythonInstallCommand(proj)
	case "Ruby":
		cmd = exec.Command("bundle", "install")
	case "PHP":
		cmd = exec.Command("composer", "install")
	default:
		return Result{
			Success: false,
			Message: fmt.Sprintf("Don't know how to install dependencies for %s projects", proj.Language),
		}
	}

	if cmd == nil {
		return Result{Success: false, Message: "No install command found"}
	}

	cmd.Dir = proj.Path
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()

	if err != nil {
		return Result{
			Success: false,
			Message: fmt.Sprintf("Install failed:\n%s", output),
		}
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("Dependencies installed:\n%s", output),
	}
}

// clean removes build artifacts
func (e *Executor) clean(proj *project.Project) Result {
	artifacts := e.detectBuildArtifacts(proj)
	
	removed := []string{}
	for _, artifact := range artifacts {
		artifactPath := filepath.Join(proj.Path, artifact)
		if _, err := os.Stat(artifactPath); err == nil {
			if err := os.RemoveAll(artifactPath); err == nil {
				removed = append(removed, artifact)
			}
		}
	}

	if len(removed) == 0 {
		return Result{Success: true, Message: "No build artifacts found to clean"}
	}

	return Result{
		Success: true,
		Message: fmt.Sprintf("Removed: %s", strings.Join(removed, ", ")),
	}
}

// detectNodeTestCommand detects the appropriate test command for Node projects
func (e *Executor) detectNodeTestCommand(proj *project.Project) *exec.Cmd {
	packageJSON := filepath.Join(proj.Path, "package.json")
	
	// Check for package managers
	if _, err := os.Stat(filepath.Join(proj.Path, "bun.lockb")); err == nil {
		return exec.Command("bun", "test")
	}
	if _, err := os.Stat(filepath.Join(proj.Path, "pnpm-lock.yaml")); err == nil {
		return exec.Command("pnpm", "test")
	}
	if _, err := os.Stat(filepath.Join(proj.Path, "yarn.lock")); err == nil {
		return exec.Command("yarn", "test")
	}
	if _, err := os.Stat(packageJSON); err == nil {
		return exec.Command("npm", "test")
	}

	return nil
}

// detectPythonTestCommand detects the appropriate test command for Python projects
func (e *Executor) detectPythonTestCommand(proj *project.Project) *exec.Cmd {
	// Check for pytest
	if commandExists("pytest") {
		return exec.Command("pytest")
	}
	// Fallback to unittest
	return exec.Command("python", "-m", "unittest", "discover")
}

// detectNodeInstallCommand detects the appropriate install command for Node projects
func (e *Executor) detectNodeInstallCommand(proj *project.Project) *exec.Cmd {
	if _, err := os.Stat(filepath.Join(proj.Path, "bun.lockb")); err == nil {
		return exec.Command("bun", "install")
	}
	if _, err := os.Stat(filepath.Join(proj.Path, "pnpm-lock.yaml")); err == nil {
		return exec.Command("pnpm", "install")
	}
	if _, err := os.Stat(filepath.Join(proj.Path, "yarn.lock")); err == nil {
		return exec.Command("yarn", "install")
	}
	return exec.Command("npm", "install")
}

// detectPythonInstallCommand detects the appropriate install command for Python projects
func (e *Executor) detectPythonInstallCommand(proj *project.Project) *exec.Cmd {
	if _, err := os.Stat(filepath.Join(proj.Path, "Pipfile")); err == nil {
		return exec.Command("pipenv", "install")
	}
	if _, err := os.Stat(filepath.Join(proj.Path, "poetry.lock")); err == nil {
		return exec.Command("poetry", "install")
	}
	if _, err := os.Stat(filepath.Join(proj.Path, "requirements.txt")); err == nil {
		return exec.Command("pip", "install", "-r", "requirements.txt")
	}
	return nil
}

// detectBuildArtifacts returns common build artifacts for the project language
func (e *Executor) detectBuildArtifacts(proj *project.Project) []string {
	common := []string{".DS_Store", "Thumbs.db"}
	
	switch proj.Language {
	case "Go":
		return append(common, "bin", "dist", "vendor")
	case "Rust":
		return append(common, "target")
	case "JavaScript", "TypeScript":
		return append(common, "node_modules", "dist", "build", ".next", "out", ".turbo", ".nuxt")
	case "Python":
		return append(common, "__pycache__", "*.pyc", ".pytest_cache", "dist", "build", "*.egg-info", ".venv", "venv")
	case "Java":
		return append(common, "target", "build", ".gradle")
	case "C#":
		return append(common, "bin", "obj")
	case "Ruby":
		return append(common, "vendor/bundle")
	case "PHP":
		return append(common, "vendor")
	default:
		return common
	}
}

// commandExists checks if a command exists in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// EditorExists checks if the configured editor exists
func (e *Executor) EditorExists() bool {
	editorCmd := e.config.Editor.Default
	cmdArgs, ok := e.config.Editor.Aliases[editorCmd]
	if !ok || len(cmdArgs) == 0 {
		return commandExists(editorCmd)
	}
	return commandExists(cmdArgs[0])
}
