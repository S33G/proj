package scripts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectPackageManagerPriority(t *testing.T) {
	dir := t.TempDir()

	if got := detectPackageManager(dir); got != "npm" {
		t.Fatalf("expected npm fallback, got %s", got)
	}

	files := []string{"bun.lockb", "pnpm-lock.yaml", "yarn.lock"}
	want := []string{"bun", "pnpm", "yarn"}

	for i, file := range files {
		path := filepath.Join(dir, file)
		if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
			t.Fatalf("failed to create %s: %v", file, err)
		}
		if got := detectPackageManager(dir); got != want[i] {
			t.Fatalf("detectPackageManager with %s = %s, want %s", file, got, want[i])
		}
		if err := os.Remove(path); err != nil {
			t.Fatalf("failed to remove %s: %v", file, err)
		}
	}
}

func TestShouldSkipNpmScript(t *testing.T) {
	if !shouldSkipNpmScript("preinstall") {
		t.Fatalf("expected preinstall to be skipped")
	}
	if shouldSkipNpmScript("start") {
		t.Fatalf("expected start to be kept")
	}
}

func TestDetectMakefileParsesTargets(t *testing.T) {
	dir := t.TempDir()
	makefile := `# Build it
build:
	echo building

# Internal target
_private:
	echo skip

all:
	echo skip all

# Run tests
test:
	echo testing
`

	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte(makefile), 0o644); err != nil {
		t.Fatalf("failed to write Makefile: %v", err)
	}

	scripts := detectMakefile(dir)
	if len(scripts) != 2 {
		t.Fatalf("expected 2 scripts, got %d", len(scripts))
	}

	if scripts[0].Name != "build" || scripts[0].Desc != "Build it" {
		t.Fatalf("unexpected first script: %+v", scripts[0])
	}
	if scripts[1].Name != "test" || scripts[1].Desc != "Run tests" {
		t.Fatalf("unexpected second script: %+v", scripts[1])
	}
}

func TestDetectShellScriptsRequiresExecutable(t *testing.T) {
	dir := t.TempDir()
	scriptsDir := filepath.Join(dir, "scripts")
	if err := os.Mkdir(scriptsDir, 0o755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	execScript := filepath.Join(scriptsDir, "deploy.sh")
	nonExecScript := filepath.Join(scriptsDir, "notes.sh")
	rootScript := filepath.Join(dir, "run.sh")

	if err := os.WriteFile(execScript, []byte("#!/bin/sh\necho ok"), 0o755); err != nil {
		t.Fatalf("failed to write exec script: %v", err)
	}
	if err := os.WriteFile(nonExecScript, []byte("echo nope"), 0o644); err != nil {
		t.Fatalf("failed to write non-exec script: %v", err)
	}
	if err := os.WriteFile(rootScript, []byte("#!/bin/sh\necho root"), 0o755); err != nil {
		t.Fatalf("failed to write root script: %v", err)
	}

	found := detectShellScripts(dir)
	if len(found) != 2 {
		t.Fatalf("expected 2 executable scripts, got %d", len(found))
	}

	names := map[string]bool{}
	for _, script := range found {
		names[script.Name] = true
	}

	if !names["deploy"] || !names["run"] {
		t.Fatalf("expected deploy and run scripts, got %v", names)
	}
}

func TestDetectGoScriptsAddsCommonCommands(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test"), 0o644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "cmd"), 0o755); err != nil {
		t.Fatalf("failed to create cmd dir: %v", err)
	}

	scripts := detectGoScripts(dir)
	if len(scripts) != 5 {
		t.Fatalf("expected 5 go scripts, got %d", len(scripts))
	}

	hasRun := false
	for _, script := range scripts {
		if script.ID == "go-run" && script.Command == "go run ./cmd/..." {
			hasRun = true
		}
	}
	if !hasRun {
		t.Fatalf("expected go-run command pointing to ./cmd/...")
	}
}

func TestTruncateCommand(t *testing.T) {
	long := strings.Repeat("a", 60)
	shortened := truncateCommand(long, 10)
	if len(shortened) != 10 {
		t.Fatalf("expected truncated length 10, got %d", len(shortened))
	}

	original := "npm test"
	if got := truncateCommand(original, 20); got != original {
		t.Fatalf("expected command to remain unchanged, got %q", got)
	}
}
