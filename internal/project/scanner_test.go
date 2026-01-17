package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cjennings/proj/internal/config"
)

func TestScanner_Scan(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create test projects
	projects := []string{"project1", "project2", "project3"}
	for _, name := range projects {
		projectPath := filepath.Join(tmpDir, name)
		os.Mkdir(projectPath, 0755)
		
		// Add a go.mod to make it detectable
		os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module test"), 0644)
	}

	// Create a hidden directory (should be excluded by default)
	hiddenPath := filepath.Join(tmpDir, ".hidden")
	os.Mkdir(hiddenPath, 0755)

	// Create an excluded directory
	excludedPath := filepath.Join(tmpDir, "node_modules")
	os.Mkdir(excludedPath, 0755)

	// Create scanner
	cfg := config.DefaultConfig()
	cfg.Display.ShowHiddenDirs = false
	scanner := NewScanner(cfg)

	// Scan directory
	found, err := scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("Expected 3 projects, found %d", len(found))
	}

	// Verify project details
	for _, p := range found {
		if p.Name == "" {
			t.Error("Project name should not be empty")
		}
		if p.Path == "" {
			t.Error("Project path should not be empty")
		}
		if p.Language != "Go" {
			t.Errorf("Expected Go language, got %s", p.Language)
		}
	}
}

func TestScanner_ScanWithHidden(t *testing.T) {
	tmpDir := t.TempDir()

	// Create visible and hidden directories
	os.Mkdir(filepath.Join(tmpDir, "visible"), 0755)
	os.Mkdir(filepath.Join(tmpDir, ".hidden"), 0755)

	// Scanner with hidden dirs disabled
	cfg := config.DefaultConfig()
	cfg.Display.ShowHiddenDirs = false
	scanner := NewScanner(cfg)

	found, err := scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(found) != 1 {
		t.Errorf("Expected 1 project (hidden disabled), found %d", len(found))
	}

	// Scanner with hidden dirs enabled
	cfg.Display.ShowHiddenDirs = true
	scanner = NewScanner(cfg)

	found, err = scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 projects (hidden enabled), found %d", len(found))
	}
}

func TestScanner_GitDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a git repo
	gitProject := filepath.Join(tmpDir, "gitproject")
	os.Mkdir(gitProject, 0755)

	// Initialize git
	cmd := exec.Command("git", "init")
	cmd.Dir = gitProject
	if err := cmd.Run(); err != nil {
		t.Skip("git not available, skipping git detection test")
	}

	// Configure git
	exec.Command("git", "-C", gitProject, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", gitProject, "config", "user.name", "Test User").Run()

	// Create initial commit
	initialFile := filepath.Join(gitProject, "init.txt")
	os.WriteFile(initialFile, []byte("init"), 0644)
	exec.Command("git", "-C", gitProject, "add", ".").Run()
	exec.Command("git", "-C", gitProject, "commit", "-m", "initial").Run()

	// Create a non-git project
	nonGitProject := filepath.Join(tmpDir, "nongit")
	os.Mkdir(nonGitProject, 0755)

	// Scan
	cfg := config.DefaultConfig()
	scanner := NewScanner(cfg)

	found, err := scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Find the git project
	var gitProj *Project
	var nonGitProj *Project
	for _, p := range found {
		if p.Name == "gitproject" {
			gitProj = p
		}
		if p.Name == "nongit" {
			nonGitProj = p
		}
	}

	if gitProj == nil {
		t.Fatal("Git project not found")
	}

	if !gitProj.IsGitRepo {
		t.Error("Git project should be detected as git repo")
	}

	if gitProj.GitBranch == "" {
		t.Error("Git project should have a branch")
	}

	if nonGitProj == nil {
		t.Fatal("Non-git project not found")
	}

	if nonGitProj.IsGitRepo {
		t.Error("Non-git project should not be detected as git repo")
	}
}

func TestSort(t *testing.T) {
	now := time.Now()
	projects := []*Project{
		{Name: "charlie", LastModified: now.Add(-2 * time.Hour)},
		{Name: "alice", LastModified: now},
		{Name: "bob", LastModified: now.Add(-1 * time.Hour)},
	}

	// Sort by name
	sorted := Sort(projects, SortByName)
	if sorted[0].Name != "alice" || sorted[1].Name != "bob" || sorted[2].Name != "charlie" {
		t.Error("Sort by name failed")
	}

	// Sort by last modified
	sorted = Sort(projects, SortByLastModified)
	if sorted[0].Name != "alice" || sorted[1].Name != "bob" || sorted[2].Name != "charlie" {
		t.Error("Sort by last modified failed")
	}
}

func TestFilter(t *testing.T) {
	projects := []*Project{
		{Name: "myapp", Language: "Go"},
		{Name: "webapp", Language: "TypeScript"},
		{Name: "api", Language: "Go", GitBranch: "feature"},
	}

	// Filter by name
	filtered := Filter(projects, "app")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 projects with 'app', got %d", len(filtered))
	}

	// Filter by language
	filtered = Filter(projects, "go")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 Go projects, got %d", len(filtered))
	}

	// Filter by branch
	filtered = Filter(projects, "feature")
	if len(filtered) != 1 {
		t.Errorf("Expected 1 project with 'feature' branch, got %d", len(filtered))
	}

	// Empty query returns all
	filtered = Filter(projects, "")
	if len(filtered) != 3 {
		t.Errorf("Expected all 3 projects with empty query, got %d", len(filtered))
	}
}

func TestScanner_IsExcluded(t *testing.T) {
	cfg := config.DefaultConfig()
	scanner := NewScanner(cfg)

	if !scanner.isExcluded("node_modules") {
		t.Error("node_modules should be excluded")
	}

	if !scanner.isExcluded(".git") {
		t.Error(".git should be excluded")
	}

	if scanner.isExcluded("myproject") {
		t.Error("myproject should not be excluded")
	}
}
