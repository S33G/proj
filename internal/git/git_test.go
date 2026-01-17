package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsInstalled(t *testing.T) {
	// This test assumes git is installed on the system
	if !IsInstalled() {
		t.Skip("git is not installed, skipping git tests")
	}
}

func TestGetStatus(t *testing.T) {
	if !IsInstalled() {
		t.Skip("git not installed")
	}

	// Create a temporary git repo for testing
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for testing
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit so we have a HEAD
	testFile := filepath.Join(tmpDir, "initial.txt")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial commit").Run()

	// Get status
	status, err := GetStatus(tmpDir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.IsRepo {
		t.Error("Should be detected as git repo")
	}

	// Branch might be "master" or "main" depending on git config
	if status.Branch == "" {
		t.Error("Branch should not be empty")
	}

	// Create a file to make repo dirty
	testFile = filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Check if dirty
	status, err = GetStatus(tmpDir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.IsDirty {
		t.Error("Repo should be dirty after adding file")
	}

	// Commit the file
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "test commit").Run()

	// Check if clean
	status, err = GetStatus(tmpDir)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.IsDirty {
		t.Error("Repo should be clean after commit")
	}
}

func TestGetStatus_NonGitDir(t *testing.T) {
	tmpDir := t.TempDir()

	status, err := GetStatus(tmpDir)
	if err != nil {
		t.Fatalf("GetStatus should not error on non-git dir: %v", err)
	}

	if status.IsRepo {
		t.Error("Non-git directory should not be detected as repo")
	}

	if status.Branch != "" {
		t.Error("Non-git directory should have empty branch")
	}

	if status.IsDirty {
		t.Error("Non-git directory should not be dirty")
	}
}

func TestGetBranches(t *testing.T) {
	if !IsInstalled() {
		t.Skip("git not installed")
	}

	// Create a temporary git repo for testing
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Get branches
	branches, err := GetBranches(tmpDir)
	if err != nil {
		t.Fatalf("GetBranches failed: %v", err)
	}

	if len(branches) == 0 {
		t.Error("Should have at least one branch")
	}

	// Create a new branch
	exec.Command("git", "-C", tmpDir, "branch", "feature").Run()

	branches, err = GetBranches(tmpDir)
	if err != nil {
		t.Fatalf("GetBranches failed: %v", err)
	}

	if len(branches) != 2 {
		t.Errorf("Expected 2 branches, got %d", len(branches))
	}
}

func TestLog(t *testing.T) {
	if !IsInstalled() {
		t.Skip("git not installed")
	}

	// Create a temporary git repo for testing
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	// Create commits
	for i := 1; i <= 3; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.txt", i))
		os.WriteFile(testFile, []byte(fmt.Sprintf("test %d", i)), 0644)
		exec.Command("git", "-C", tmpDir, "add", ".").Run()
		exec.Command("git", "-C", tmpDir, "commit", "-m", fmt.Sprintf("commit %d", i)).Run()
	}

	// Get log
	logs, err := Log(tmpDir, 10)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(logs))
	}
}
