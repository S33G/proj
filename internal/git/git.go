package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Status represents the git status of a project
type Status struct {
	IsRepo  bool
	Branch  string
	IsDirty bool
}

// GetStatus returns the git status for a project directory
func GetStatus(projectPath string) (*Status, error) {
	status := &Status{}

	// Check if .git directory exists
	gitDir := filepath.Join(projectPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil || !info.IsDir() {
		return status, nil
	}

	status.IsRepo = true

	// Get current branch
	branch, err := getCurrentBranch(projectPath)
	if err == nil {
		status.Branch = branch
	}

	// Check if dirty
	dirty, err := isDirty(projectPath)
	if err == nil {
		status.IsDirty = dirty
	}

	return status, nil
}

// getCurrentBranch returns the current git branch
func getCurrentBranch(projectPath string) (string, error) {
	cmd := exec.Command("git", "-C", projectPath, "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

// isDirty checks if there are uncommitted changes
func isDirty(projectPath string) (bool, error) {
	cmd := exec.Command("git", "-C", projectPath, "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return false, err
	}

	return len(strings.TrimSpace(out.String())) > 0, nil
}

// GetBranches returns all git branches
func GetBranches(projectPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", projectPath, "branch", "--format=%(refname:short)")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	branches := strings.Split(strings.TrimSpace(out.String()), "\n")
	result := make([]string, 0, len(branches))
	for _, branch := range branches {
		if branch = strings.TrimSpace(branch); branch != "" {
			result = append(result, branch)
		}
	}

	return result, nil
}

// Pull performs a git pull
func Pull(projectPath string) (string, error) {
	cmd := exec.Command("git", "-C", projectPath, "pull")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// Checkout switches to a different branch
func Checkout(projectPath, branch string) (string, error) {
	cmd := exec.Command("git", "-C", projectPath, "checkout", branch)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

// Log returns recent git log entries
func Log(projectPath string, limit int) ([]string, error) {
	cmd := exec.Command("git", "-C", projectPath, "log", "--oneline", "--graph", "--decorate", "-n", fmt.Sprintf("%d", limit))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

// IsInstalled checks if git is installed on the system
func IsInstalled() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}
