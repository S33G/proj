package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	ReposPath       string        `json:"reposPath" mapstructure:"reposPath"`
	Editor          EditorConfig  `json:"editor" mapstructure:"editor"`
	Shell           string        `json:"shell" mapstructure:"shell"`
	Theme           ThemeConfig   `json:"theme" mapstructure:"theme"`
	Display         DisplayConfig `json:"display" mapstructure:"display"`
	ExcludePatterns []string      `json:"excludePatterns" mapstructure:"excludePatterns"`
	Actions         ActionsConfig `json:"actions" mapstructure:"actions"`
	Plugins         PluginsConfig `json:"plugins" mapstructure:"plugins"`
}

// EditorConfig holds editor settings
type EditorConfig struct {
	Default string              `json:"default" mapstructure:"default"`
	Aliases map[string][]string `json:"aliases" mapstructure:"aliases"`
}

// ThemeConfig holds color theme settings
type ThemeConfig struct {
	PrimaryColor string `json:"primaryColor" mapstructure:"primaryColor"`
	AccentColor  string `json:"accentColor" mapstructure:"accentColor"`
	ErrorColor   string `json:"errorColor" mapstructure:"errorColor"`
}

// DisplayConfig holds display preferences
type DisplayConfig struct {
	ShowHiddenDirs bool   `json:"showHiddenDirs" mapstructure:"showHiddenDirs"`
	SortBy         string `json:"sortBy" mapstructure:"sortBy"` // "name" or "lastModified"
	ShowGitStatus  bool   `json:"showGitStatus" mapstructure:"showGitStatus"`
	ShowLanguage   bool   `json:"showLanguage" mapstructure:"showLanguage"`
}

// ActionsConfig holds action-related settings
type ActionsConfig struct {
	EnableGitOperations bool `json:"enableGitOperations" mapstructure:"enableGitOperations"`
	EnableTestRunner    bool `json:"enableTestRunner" mapstructure:"enableTestRunner"`
}

// PluginsConfig holds plugin settings
type PluginsConfig struct {
	Enabled []string               `json:"enabled" mapstructure:"enabled"`
	Config  map[string]interface{} `json:"config" mapstructure:"config"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()

	return &Config{
		ReposPath: filepath.Join(home, "code"),
		Editor: EditorConfig{
			Default: "code",
			Aliases: map[string][]string{
				"code":     {"code", "--goto"},
				"nvim":     {"nvim"},
				"vim":      {"vim"},
				"emacs":    {"emacsclient", "-n"},
				"idea":     {"idea"},
				"goland":   {"goland"},
				"pycharm":  {"pycharm"},
				"webstorm": {"webstorm"},
				"clion":    {"clion"},
				"rubymine": {"rubymine"},
				"phpstorm": {"phpstorm"},
				"zed":      {"zed"},
				"subl":     {"subl"},
				"hx":       {"hx"},
				"cursor":   {"cursor"},
			},
		},
		Shell: "/bin/bash",
		Theme: ThemeConfig{
			PrimaryColor: "#00CED1",
			AccentColor:  "#32CD32",
			ErrorColor:   "#FF6347",
		},
		Display: DisplayConfig{
			ShowHiddenDirs: false,
			SortBy:         "lastModified",
			ShowGitStatus:  true,
			ShowLanguage:   true,
		},
		ExcludePatterns: []string{".git", "node_modules", ".DS_Store", "__pycache__", "vendor"},
		Actions: ActionsConfig{
			EnableGitOperations: true,
			EnableTestRunner:    true,
		},
		Plugins: PluginsConfig{
			Enabled: []string{},
			Config:  make(map[string]interface{}),
		},
	}
}

// ConfigDir returns the configuration directory path
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "proj"), nil
}

// ConfigPath returns the full path to the config file
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load loads the configuration from the config file
func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// Set up viper
	viper.SetConfigFile(configPath)
	viper.SetConfigType("json")

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			return DefaultConfig(), nil
		}
		return nil, err
	}

	// Unmarshal into config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves the configuration to the config file
func Save(cfg *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configPath, data, 0644)
}

// Init initializes the configuration directory and file with defaults
func Init() error {
	cfg := DefaultConfig()
	return Save(cfg)
}

// setDefaults sets default values in viper
func setDefaults() {
	home, _ := os.UserHomeDir()

	viper.SetDefault("reposPath", filepath.Join(home, "code"))
	viper.SetDefault("editor.default", "code")
	viper.SetDefault("shell", "/bin/bash")
	viper.SetDefault("theme.primaryColor", "#00CED1")
	viper.SetDefault("theme.accentColor", "#32CD32")
	viper.SetDefault("theme.errorColor", "#FF6347")
	viper.SetDefault("display.showHiddenDirs", false)
	viper.SetDefault("display.sortBy", "lastModified")
	viper.SetDefault("display.showGitStatus", true)
	viper.SetDefault("display.showLanguage", true)
	viper.SetDefault("excludePatterns", []string{".git", "node_modules", ".DS_Store", "__pycache__", "vendor"})
	viper.SetDefault("actions.enableGitOperations", true)
	viper.SetDefault("actions.enableTestRunner", true)
}

// ExpandPath expands ~ to the user's home directory
func ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
