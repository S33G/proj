package language

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{
			name:     "Go project",
			files:    []string{"go.mod", "main.go"},
			expected: "Go",
		},
		{
			name:     "Rust project",
			files:    []string{"Cargo.toml", "src/main.rs"},
			expected: "Rust",
		},
		{
			name:     "TypeScript project",
			files:    []string{"tsconfig.json", "index.ts"},
			expected: "TypeScript",
		},
		{
			name:     "JavaScript project",
			files:    []string{"package.json", "index.js"},
			expected: "JavaScript",
		},
		{
			name:     "Python project",
			files:    []string{"requirements.txt", "main.py"},
			expected: "Python",
		},
		{
			name:     "Java project",
			files:    []string{"pom.xml", "Main.java"},
			expected: "Java",
		},
		{
			name:     "Ruby project",
			files:    []string{"Gemfile", "app.rb"},
			expected: "Ruby",
		},
		{
			name:     "C++ project",
			files:    []string{"CMakeLists.txt", "main.cpp"},
			expected: "C/C++",
		},
		{
			name:     "Git repo only",
			files:    []string{".git"},
			expected: "Git Repo",
		},
		{
			name:     "Unknown project",
			files:    []string{"README.md"},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Create test files
			for _, file := range tt.files {
				filePath := filepath.Join(tmpDir, file)
				dir := filepath.Dir(filePath)
				
				// Create directory if needed
				if dir != tmpDir {
					os.MkdirAll(dir, 0755)
				}

				// Create file or directory
				if file == ".git" {
					os.Mkdir(filePath, 0755)
				} else {
					os.WriteFile(filePath, []byte("test"), 0644)
				}
			}

			// Detect language
			result, err := Detect(tmpDir)
			if err != nil {
				t.Fatalf("Detect failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetIcon(t *testing.T) {
	tests := []struct {
		language string
		expected string
	}{
		{"Go", "🐹"},
		{"Rust", "🦀"},
		{"TypeScript", "⚡"},
		{"JavaScript", "📜"},
		{"Python", "🐍"},
		{"Java", "☕"},
		{"Unknown", "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			icon := GetIcon(tt.language)
			if icon != tt.expected {
				t.Errorf("Expected icon %s for %s, got %s", tt.expected, tt.language, icon)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !contains(slice, "b") {
		t.Error("Should contain 'b'")
	}

	if contains(slice, "d") {
		t.Error("Should not contain 'd'")
	}
}

func TestHasExtension(t *testing.T) {
	files := []string{"main.go", "test.txt", "README.md"}

	if !hasExtension(files, ".go") {
		t.Error("Should have .go extension")
	}

	if !hasExtension(files, ".md", ".txt") {
		t.Error("Should have .md or .txt extension")
	}

	if hasExtension(files, ".rs") {
		t.Error("Should not have .rs extension")
	}
}

func TestAddDetector(t *testing.T) {
	initialCount := len(GetDetectors())

	customDetector := Detector{
		Language: "Test",
		Priority: 99,
		Check: func(files []string) bool {
			return contains(files, "test.txt")
		},
	}

	AddDetector(customDetector)

	newCount := len(GetDetectors())
	if newCount != initialCount+1 {
		t.Errorf("Expected %d detectors, got %d", initialCount+1, newCount)
	}
}

func TestPriorityOrder(t *testing.T) {
	// Create a directory with both package.json and tsconfig.json
	// TypeScript should be detected before JavaScript due to priority
	tmpDir := t.TempDir()
	
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte("{}"), 0644)

	result, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if result != "TypeScript" {
		t.Errorf("Expected TypeScript to be detected with higher priority, got %s", result)
	}
}
