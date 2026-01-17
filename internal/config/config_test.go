package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if cfg.ReposPath == "" {
		t.Error("ReposPath should not be empty")
	}

	if cfg.Editor.Default == "" {
		t.Error("Editor.Default should not be empty")
	}

	if len(cfg.Editor.Aliases) == 0 {
		t.Error("Editor.Aliases should not be empty")
	}

	if cfg.Display.SortBy != "lastModified" && cfg.Display.SortBy != "name" {
		t.Errorf("Invalid SortBy value: %s", cfg.Display.SortBy)
	}

	if len(cfg.ExcludePatterns) == 0 {
		t.Error("ExcludePatterns should not be empty")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override config path for testing
	oldHome := os.Getenv("HOME")
	testHome := tmpDir
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", oldHome)

	// Create config
	cfg := DefaultConfig()
	cfg.ReposPath = "/test/path"
	cfg.Editor.Default = "nvim"
	cfg.Display.SortBy = "name"

	// Save config
	err := Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config matches saved config
	if loaded.ReposPath != cfg.ReposPath {
		t.Errorf("ReposPath mismatch: got %s, want %s", loaded.ReposPath, cfg.ReposPath)
	}

	if loaded.Editor.Default != cfg.Editor.Default {
		t.Errorf("Editor.Default mismatch: got %s, want %s", loaded.Editor.Default, cfg.Editor.Default)
	}

	if loaded.Display.SortBy != cfg.Display.SortBy {
		t.Errorf("Display.SortBy mismatch: got %s, want %s", loaded.Display.SortBy, cfg.Display.SortBy)
	}
}

func TestInit(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override home for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Initialize config
	err := Init()
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Verify config file exists
	configPath, _ := ConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load initialized config: %v", err)
	}

	if cfg == nil {
		t.Fatal("Loaded config is nil")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/code", filepath.Join(home, "code")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~", home},
	}

	for _, tt := range tests {
		result := ExpandPath(tt.input)
		if result != tt.expected {
			t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir failed: %v", err)
	}

	if dir == "" {
		t.Error("ConfigDir returned empty string")
	}

	if !filepath.IsAbs(dir) {
		t.Error("ConfigDir should return absolute path")
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}

	if path == "" {
		t.Error("ConfigPath returned empty string")
	}

	if !filepath.IsAbs(path) {
		t.Error("ConfigPath should return absolute path")
	}

	if filepath.Ext(path) != ".json" {
		t.Error("ConfigPath should have .json extension")
	}
}
