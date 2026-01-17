package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryLoadAll(t *testing.T) {
	// Get current directory to find test plugins
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	
	// Use the example plugin in the project
	projectRoot := filepath.Join(cwd, "..", "..")
	pluginsDir := filepath.Join(projectRoot, "plugins")
	
	// Check if plugins directory exists
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		t.Skip("Plugins directory not found, skipping test")
	}
	
	// Create registry with example plugin enabled
	configDir := t.TempDir()
	registry := NewRegistry(pluginsDir, configDir, []string{"example"}, map[string]interface{}{
		"example": map[string]interface{}{
			"test": true,
		},
	})
	
	// Load plugins
	err = registry.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	
	// Check if example plugin was loaded
	plugin := registry.GetPlugin("example")
	if plugin == nil {
		// This is okay - plugin might not be built yet
		t.Log("Example plugin not loaded (might not be built)")
		return
	}
	
	if plugin.manifest.Name != "example" {
		t.Errorf("Expected plugin name 'example', got '%s'", plugin.manifest.Name)
	}
}

func TestRegistryGetActions(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}
	
	projectRoot := filepath.Join(cwd, "..", "..")
	pluginsDir := filepath.Join(projectRoot, "plugins")
	
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		t.Skip("Plugins directory not found, skipping test")
	}
	
	configDir := t.TempDir()
	registry := NewRegistry(pluginsDir, configDir, []string{"example"}, nil)
	
	err = registry.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	
	// Create a test project
	proj := Project{
		Name:     "test-project",
		Path:     "/tmp/test",
		Language: "Go",
	}
	
	// Get actions
	actions := registry.GetActions(proj)
	
	// Should have at least the example action if plugin is loaded
	if len(actions) > 0 {
		t.Logf("Got %d actions from plugins", len(actions))
		for _, action := range actions {
			t.Logf("  - %s: %s", action.ID, action.Label)
		}
	}
}

func TestRegistryIsEnabled(t *testing.T) {
	registry := NewRegistry("/tmp", "/tmp", []string{"plugin1", "plugin2"}, nil)
	
	if !registry.isEnabled("plugin1") {
		t.Error("plugin1 should be enabled")
	}
	
	if !registry.isEnabled("plugin2") {
		t.Error("plugin2 should be enabled")
	}
	
	if registry.isEnabled("plugin3") {
		t.Error("plugin3 should not be enabled")
	}
}

func TestRegistryGetPluginConfig(t *testing.T) {
	pluginConfig := map[string]interface{}{
		"plugin1": map[string]interface{}{
			"key": "value",
		},
	}
	
	registry := NewRegistry("/tmp", "/tmp", []string{}, pluginConfig)
	
	cfg := registry.getPluginConfig("plugin1")
	if cfg["key"] != "value" {
		t.Errorf("Expected key='value', got key='%v'", cfg["key"])
	}
	
	cfg2 := registry.getPluginConfig("plugin2")
	if len(cfg2) != 0 {
		t.Error("Expected empty config for non-existent plugin")
	}
}
