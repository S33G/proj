package plugin

import (
	"encoding/json"
	"fmt"
)

// ExternalPlugin represents a plugin loaded from an external executable
type ExternalPlugin struct {
	client   *RPCClient
	manifest *Manifest
	config   map[string]interface{}
	execPath string
}

// NewExternalPlugin creates a new external plugin
func NewExternalPlugin(execPath string, manifest *Manifest, config map[string]interface{}) *ExternalPlugin {
	return &ExternalPlugin{
		manifest: manifest,
		config:   config,
		execPath: execPath,
	}
}

// Init initializes the plugin
func (p *ExternalPlugin) Init() error {
	client, err := NewRPCClient(p.execPath)
	if err != nil {
		return err
	}
	p.client = client

	// Call init method
	params := map[string]interface{}{
		"config": p.config,
	}

	_, err = p.client.Call("init", params)
	return err
}

// GetActions gets actions from the plugin
func (p *ExternalPlugin) GetActions(proj Project) ([]Action, error) {
	if !p.hasCapability("actions") {
		return nil, nil
	}

	result, err := p.client.Call("actions", proj)
	if err != nil {
		return nil, err
	}

	var actions []Action
	if err := json.Unmarshal(result, &actions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
	}

	return actions, nil
}

// ExecuteAction executes an action
func (p *ExternalPlugin) ExecuteAction(actionID string, proj Project) (*ActionResult, error) {
	params := map[string]interface{}{
		"action":  actionID,
		"project": proj,
	}

	result, err := p.client.Call("executeAction", params)
	if err != nil {
		return nil, err
	}

	var actionResult ActionResult
	if err := json.Unmarshal(result, &actionResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal action result: %w", err)
	}

	return &actionResult, nil
}

// GetLanguages gets language detectors from the plugin
func (p *ExternalPlugin) GetLanguages() ([]LanguageDetector, error) {
	if !p.hasCapability("languages") {
		return nil, nil
	}

	result, err := p.client.Call("languages", nil)
	if err != nil {
		return nil, err
	}

	var languages []LanguageDetector
	if err := json.Unmarshal(result, &languages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal languages: %w", err)
	}

	return languages, nil
}

// Shutdown shuts down the plugin
func (p *ExternalPlugin) Shutdown() error {
	if p.client == nil {
		return nil
	}

	// Call shutdown method (ignore errors)
	_, _ = p.client.Call("shutdown", nil)

	// Close client
	return p.client.Close()
}

// hasCapability checks if the plugin has a capability
func (p *ExternalPlugin) hasCapability(capability string) bool {
	for _, cap := range p.manifest.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// Name returns the plugin name
func (p *ExternalPlugin) Name() string {
	return p.manifest.Name
}

// Version returns the plugin version
func (p *ExternalPlugin) Version() string {
	return p.manifest.Version
}

// Description returns the plugin description
func (p *ExternalPlugin) Description() string {
	return p.manifest.Description
}
