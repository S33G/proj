package plugin

// Plugin represents a plugin that can extend proj
type Plugin interface {
	// Metadata
	Name() string
	Version() string
	Description() string

	// Lifecycle
	Init(config PluginConfig) error
	Shutdown() error

	// Extension points
	Actions(project Project) ([]Action, error)
	Languages() ([]LanguageDetector, error)
	OnAction(action string, project Project) (*ActionResult, error)
}

// PluginConfig holds configuration for a plugin
type PluginConfig struct {
	ConfigDir  string
	PluginData map[string]interface{}
}

// Project represents a project for plugins
type Project struct {
	Name      string
	Path      string
	Language  string
	GitBranch string
	GitDirty  bool
	IsGitRepo bool
}

// Action represents a custom action
type Action struct {
	ID          string
	Label       string
	Description string
	Icon        string
	Priority    int
}

// ActionResult represents the result of an action
type ActionResult struct {
	Success bool
	Message string
	CdPath  string
	ExecCmd []string
}

// LanguageDetector represents a language detector
type LanguageDetector struct {
	Language string
	Priority int
	Files    []string // Files to check for (e.g., "Cargo.toml")
	Pattern  string   // Regex pattern for files
}
