package views

import (
	"testing"

	"github.com/s33g/proj/internal/project"
)

func TestActionSubmenu(t *testing.T) {
	tests := []struct {
		name     string
		action   Action
		expected bool
	}{
		{
			name: "regular action",
			action: Action{
				ID:        "open-editor",
				Label:     "Open in Editor",
				IsSubmenu: false,
			},
			expected: false,
		},
		{
			name: "submenu action",
			action: Action{
				ID:        "submenu-docker",
				Label:     "Docker",
				IsSubmenu: true,
				Children: []Action{
					{ID: "docker-build", Label: "Build Image"},
					{ID: "docker-run", Label: "Run Container"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action.IsSubmenu != tt.expected {
				t.Errorf("IsSubmenu = %v, want %v", tt.action.IsSubmenu, tt.expected)
			}
			if tt.action.IsSubmenu && len(tt.action.Children) == 0 {
				t.Error("Submenu action should have children")
			}
		})
	}
}

func TestDefaultActionsStructure(t *testing.T) {
	// Create a test project
	proj := &project.Project{
		Name:          "test-project",
		Path:          "/tmp/test",
		Language:      "Go",
		IsGitRepo:     true,
		GitBranch:     "main",
		HasDockerfile: false,
		HasCompose:    false,
	}

	actions := DefaultActions(proj, true, true)

	// Should always have base actions
	if len(actions) < 2 {
		t.Errorf("Expected at least 2 base actions, got %d", len(actions))
	}

	// Check for required base actions
	hasOpenEditor := false
	hasChangeDir := false
	hasBack := false

	for _, a := range actions {
		switch a.ID {
		case "open-editor":
			hasOpenEditor = true
		case "cd":
			hasChangeDir = true
		case "back":
			hasBack = true
		}
	}

	if !hasOpenEditor {
		t.Error("Missing 'open-editor' action")
	}
	if !hasChangeDir {
		t.Error("Missing 'cd' action")
	}
	if !hasBack {
		t.Error("Missing 'back' action")
	}
}

func TestDefaultActionsWithDocker(t *testing.T) {
	// This test would require mocking the docker.Detect function
	// For now, we test the structure without actual docker files
	proj := &project.Project{
		Name:          "docker-project",
		Path:          "/tmp/docker-test",
		Language:      "Go",
		IsGitRepo:     true,
		GitBranch:     "main",
		HasDockerfile: true,
		HasCompose:    true,
	}

	actions := DefaultActions(proj, true, true)

	// Should have base actions
	if len(actions) < 2 {
		t.Errorf("Expected at least 2 actions, got %d", len(actions))
	}

	// Note: Docker actions won't actually be added without docker files
	// This is a limitation of the current test - we'd need to create temp docker files
	// or refactor to allow dependency injection for testing
}

func TestGetSourceDisplayName(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"package.json", "npm Scripts"},
		{"Makefile", "Make"},
		{"justfile", "Just"},
		{"go", "Go Commands"},
		{"cargo", "Cargo"},
		{"poetry", "Poetry"},
		{"unknown-source", "unknown-source"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := getSourceDisplayName(tt.source)
			if result != tt.expected {
				t.Errorf("getSourceDisplayName(%q) = %q, want %q", tt.source, result, tt.expected)
			}
		})
	}
}

func TestGetScriptIcon(t *testing.T) {
	tests := []struct {
		source       string
		expectedIcon string
	}{
		{"package.json", "📜"},
		{"Makefile", "⚙️"},
		{"justfile", "📋"},
		{"go", "🔵"},
		{"cargo", "🦀"},
		{"poetry", "🐍"},
		{"pip", "🐍"},
		{"python", "🐍"},
		{"pytest", "🐍"},
		{"django", "🎸"},
		{"bundler", "💎"},
		{"rake", "💎"},
		{"rails", "💎"},
		{"unknown", "▶️"},
		{"scripts/", "📄"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := getScriptIcon(tt.source)
			if result != tt.expectedIcon {
				t.Errorf("getScriptIcon(%q) = %q, want %q", tt.source, result, tt.expectedIcon)
			}
		})
	}
}

func TestActionFilterValue(t *testing.T) {
	action := Action{
		ID:    "test-action",
		Label: "Test Action",
		Desc:  "This is a test",
	}

	filterValue := action.FilterValue()
	if filterValue != "Test Action" {
		t.Errorf("FilterValue() = %q, want %q", filterValue, "Test Action")
	}
}

func TestActionTitle(t *testing.T) {
	tests := []struct {
		name     string
		action   Action
		expected string
	}{
		{
			name: "action with icon",
			action: Action{
				Icon:  "🚀",
				Label: "Open Editor",
			},
			expected: "🚀  Open Editor",
		},
		{
			name: "action without icon",
			action: Action{
				Icon:  "",
				Label: "Back",
			},
			expected: "Back",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action.Title()
			if result != tt.expected {
				t.Errorf("Title() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestActionDescription(t *testing.T) {
	action := Action{
		ID:    "test",
		Label: "Test",
		Desc:  "Test description",
	}

	desc := action.Description()
	if desc != "Test description" {
		t.Errorf("Description() = %q, want %q", desc, "Test description")
	}
}

func TestNewActionMenuModel(t *testing.T) {
	proj := &project.Project{
		Name:     "test",
		Path:     "/tmp/test",
		Language: "Go",
	}

	actions := []Action{
		{ID: "action1", Label: "Action 1"},
		{ID: "action2", Label: "Action 2"},
	}

	model := NewActionMenuModel(proj, actions)

	if model.project == nil {
		t.Error("Expected project to be set")
	}

	if model.project.Name != "test" {
		t.Errorf("Project name = %q, want %q", model.project.Name, "test")
	}
}
