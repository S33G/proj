package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest represents a plugin manifest file
type Manifest struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Executable   string                 `json:"executable"`
	Capabilities []string               `json:"capabilities"`
	Config       map[string]interface{} `json:"config"`
}

// Registry manages loaded plugins
type Registry struct {
	plugins      map[string]*ExternalPlugin
	pluginsDir   string
	configDir    string
	enabledList  []string
	pluginConfig map[string]interface{}
}

// NewRegistry creates a new plugin registry
func NewRegistry(pluginsDir, configDir string, enabledList []string, pluginConfig map[string]interface{}) *Registry {
	return &Registry{
		plugins:      make(map[string]*ExternalPlugin),
		pluginsDir:   pluginsDir,
		configDir:    configDir,
		enabledList:  enabledList,
		pluginConfig: pluginConfig,
	}
}

// LoadAll loads all enabled plugins
func (r *Registry) LoadAll() error {
	// Ensure plugins directory exists
	if err := os.MkdirAll(r.pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Read plugins directory
	entries, err := os.ReadDir(r.pluginsDir)
	if err != nil {
		return fmt.Errorf("failed to read plugins directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginName := entry.Name()

		// Check if plugin is enabled
		if !r.isEnabled(pluginName) {
			continue
		}

		// Load manifest
		manifestPath := filepath.Join(r.pluginsDir, pluginName, "plugin.json")
		manifest, err := r.loadManifest(manifestPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugin %s: %v\n", pluginName, err)
			continue
		}

		// Create external plugin
		execPath := filepath.Join(r.pluginsDir, pluginName, manifest.Executable)
		plugin := NewExternalPlugin(execPath, manifest, r.getPluginConfig(pluginName))

		// Initialize plugin
		if err := plugin.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize plugin %s: %v\n", pluginName, err)
			continue
		}

		r.plugins[pluginName] = plugin
	}

	return nil
}

// loadManifest loads a plugin manifest
func (r *Registry) loadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// isEnabled checks if a plugin is enabled
func (r *Registry) isEnabled(name string) bool {
	for _, enabled := range r.enabledList {
		if enabled == name {
			return true
		}
	}
	return false
}

// getPluginConfig gets configuration for a plugin
func (r *Registry) getPluginConfig(name string) map[string]interface{} {
	if r.pluginConfig == nil {
		return make(map[string]interface{})
	}
	if cfg, ok := r.pluginConfig[name].(map[string]interface{}); ok {
		return cfg
	}
	return make(map[string]interface{})
}

// GetPlugin gets a plugin by name
func (r *Registry) GetPlugin(name string) *ExternalPlugin {
	return r.plugins[name]
}

// GetAllPlugins returns all loaded plugins
func (r *Registry) GetAllPlugins() []*ExternalPlugin {
	plugins := make([]*ExternalPlugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// GetActions gets all actions from all plugins for a project
func (r *Registry) GetActions(proj Project) []Action {
	actions := []Action{}

	for _, plugin := range r.plugins {
		pluginActions, err := plugin.GetActions(proj)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: plugin %s failed to get actions: %v\n", plugin.manifest.Name, err)
			continue
		}
		actions = append(actions, pluginActions...)
	}

	return actions
}

// ExecuteAction executes a plugin action
func (r *Registry) ExecuteAction(actionID string, proj Project) (*ActionResult, error) {
	// Try each plugin
	for _, plugin := range r.plugins {
		result, err := plugin.ExecuteAction(actionID, proj)
		if err != nil {
			continue
		}
		if result != nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("no plugin handled action: %s", actionID)
}

// Shutdown shuts down all plugins
func (r *Registry) Shutdown() {
	for _, plugin := range r.plugins {
		if err := plugin.Shutdown(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to shutdown plugin %s: %v\n", plugin.manifest.Name, err)
		}
	}
}
