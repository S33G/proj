package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDockerfile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"Dockerfile", true},
		{"Dockerfile.dev", true},
		{"Dockerfile.prod", true},
		{"Dockerfile.test", true},
		{"dockerfile", false},
		{"MyDockerfile", false},
		{"Dockerfile-backup", false},
		{"docker-compose.yml", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isDockerfile(tt.filename)
			if result != tt.expected {
				t.Errorf("isDockerfile(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsComposeFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"docker-compose.yml", true},
		{"docker-compose.yaml", true},
		{"compose.yml", true},
		{"compose.yaml", true},
		{"docker-compose.override.yml", false},
		{"Dockerfile", false},
		{"compose.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isComposeFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isComposeFile(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "docker-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test 1: Empty directory
	t.Run("empty directory", func(t *testing.T) {
		info, err := Detect(tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if info.HasDockerfile {
			t.Error("Expected HasDockerfile to be false for empty directory")
		}
		if info.HasCompose {
			t.Error("Expected HasCompose to be false for empty directory")
		}
	})

	// Test 2: Directory with Dockerfile
	t.Run("with Dockerfile", func(t *testing.T) {
		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte("FROM alpine"), 0644); err != nil {
			t.Fatal(err)
		}

		info, err := Detect(tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if !info.HasDockerfile {
			t.Error("Expected HasDockerfile to be true")
		}
		if len(info.Dockerfiles) != 1 || info.Dockerfiles[0] != "Dockerfile" {
			t.Errorf("Expected Dockerfiles to contain 'Dockerfile', got %v", info.Dockerfiles)
		}

		os.Remove(dockerfilePath)
	})

	// Test 3: Directory with Compose file
	t.Run("with compose file", func(t *testing.T) {
		composePath := filepath.Join(tmpDir, "docker-compose.yml")
		if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
			t.Fatal(err)
		}

		info, err := Detect(tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if !info.HasCompose {
			t.Error("Expected HasCompose to be true")
		}
		if len(info.ComposeFiles) != 1 || info.ComposeFiles[0] != "docker-compose.yml" {
			t.Errorf("Expected ComposeFiles to contain 'docker-compose.yml', got %v", info.ComposeFiles)
		}

		os.Remove(composePath)
	})

	// Test 4: Directory with both Dockerfile and Compose
	t.Run("with both", func(t *testing.T) {
		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		composePath := filepath.Join(tmpDir, "compose.yml")

		if err := os.WriteFile(dockerfilePath, []byte("FROM alpine"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(composePath, []byte("version: '3'"), 0644); err != nil {
			t.Fatal(err)
		}

		info, err := Detect(tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if !info.HasDockerfile {
			t.Error("Expected HasDockerfile to be true")
		}
		if !info.HasCompose {
			t.Error("Expected HasCompose to be true")
		}

		os.Remove(dockerfilePath)
		os.Remove(composePath)
	})

	// Test 5: Directory with multiple Dockerfile variants
	t.Run("with multiple Dockerfiles", func(t *testing.T) {
		files := []string{"Dockerfile", "Dockerfile.dev", "Dockerfile.prod"}
		for _, f := range files {
			path := filepath.Join(tmpDir, f)
			if err := os.WriteFile(path, []byte("FROM alpine"), 0644); err != nil {
				t.Fatal(err)
			}
		}

		info, err := Detect(tmpDir)
		if err != nil {
			t.Fatalf("Detect() error = %v", err)
		}
		if !info.HasDockerfile {
			t.Error("Expected HasDockerfile to be true")
		}
		if len(info.Dockerfiles) != 3 {
			t.Errorf("Expected 3 Dockerfiles, got %d", len(info.Dockerfiles))
		}

		for _, f := range files {
			os.Remove(filepath.Join(tmpDir, f))
		}
	})
}

func TestGetPrimaryDockerfile(t *testing.T) {
	tests := []struct {
		name        string
		dockerfiles []string
		expected    string
	}{
		{
			name:        "prefer plain Dockerfile",
			dockerfiles: []string{"Dockerfile.dev", "Dockerfile", "Dockerfile.prod"},
			expected:    "Dockerfile",
		},
		{
			name:        "first variant if no plain Dockerfile",
			dockerfiles: []string{"Dockerfile.dev", "Dockerfile.prod"},
			expected:    "Dockerfile.dev",
		},
		{
			name:        "empty list",
			dockerfiles: []string{},
			expected:    "Dockerfile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPrimaryDockerfile(tt.dockerfiles)
			if result != tt.expected {
				t.Errorf("GetPrimaryDockerfile(%v) = %q, want %q", tt.dockerfiles, result, tt.expected)
			}
		})
	}
}

func TestGetPrimaryComposeFile(t *testing.T) {
	tests := []struct {
		name         string
		composeFiles []string
		expected     string
	}{
		{
			name:         "prefer compose.yml",
			composeFiles: []string{"docker-compose.yml", "compose.yml"},
			expected:     "compose.yml",
		},
		{
			name:         "fallback to docker-compose.yml",
			composeFiles: []string{"docker-compose.yml", "docker-compose.yaml"},
			expected:     "docker-compose.yml",
		},
		{
			name:         "empty list",
			composeFiles: []string{},
			expected:     "docker-compose.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPrimaryComposeFile(tt.composeFiles)
			if result != tt.expected {
				t.Errorf("GetPrimaryComposeFile(%v) = %q, want %q", tt.composeFiles, result, tt.expected)
			}
		})
	}
}

func TestHasDevContainer(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "devcontainer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test without devcontainer
	t.Run("without devcontainer", func(t *testing.T) {
		if HasDevContainer(tmpDir) {
			t.Error("Expected HasDevContainer to be false")
		}
	})

	// Test with devcontainer
	t.Run("with devcontainer", func(t *testing.T) {
		devcontainerDir := filepath.Join(tmpDir, ".devcontainer")
		if err := os.Mkdir(devcontainerDir, 0755); err != nil {
			t.Fatal(err)
		}

		devcontainerFile := filepath.Join(devcontainerDir, "devcontainer.json")
		if err := os.WriteFile(devcontainerFile, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		if !HasDevContainer(tmpDir) {
			t.Error("Expected HasDevContainer to be true")
		}
	})
}
