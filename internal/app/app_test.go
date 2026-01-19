package app

import (
	"testing"

	"github.com/s33g/proj/internal/project"
	"github.com/s33g/proj/pkg/plugin"
)

func TestNextSortByCyclesThroughOptions(t *testing.T) {
	m := Model{currentSortBy: project.SortByName}

	if next := m.nextSortBy(); next != project.SortByLastModified {
		t.Fatalf("expected SortByLastModified after SortByName, got %v", next)
	}

	m.currentSortBy = project.SortByLastModified
	if next := m.nextSortBy(); next != project.SortByLanguage {
		t.Fatalf("expected SortByLanguage after SortByLastModified, got %v", next)
	}

	m.currentSortBy = project.SortByLanguage
	if next := m.nextSortBy(); next != project.SortByName {
		t.Fatalf("expected SortByName after SortByLanguage, got %v", next)
	}
}

func TestGetSortLabel(t *testing.T) {
	tests := []struct {
		sortBy   project.SortBy
		expected string
	}{
		{project.SortByName, "Alphabetical (A-Z)"},
		{project.SortByLastModified, "Last Modified"},
		{project.SortByLanguage, "Language"},
		{"unknown", "Unknown"},
	}

	for _, tt := range tests {
		m := Model{currentSortBy: tt.sortBy}
		if got := m.getSortLabel(); got != tt.expected {
			t.Errorf("getSortLabel(%v) = %q, want %q", tt.sortBy, got, tt.expected)
		}
	}
}

func TestProjectToPlugin(t *testing.T) {
	proj := &project.Project{
		Name:      "proj",
		Path:      "/tmp/proj",
		Language:  "Go",
		GitBranch: "main",
		GitDirty:  true,
		IsGitRepo: true,
	}

	converted := projectToPlugin(proj)

	expected := plugin.Project{
		Name:      "proj",
		Path:      "/tmp/proj",
		Language:  "Go",
		GitBranch: "main",
		GitDirty:  true,
		IsGitRepo: true,
	}

	if converted != expected {
		t.Fatalf("projectToPlugin mismatch: got %+v, want %+v", converted, expected)
	}
}

func TestJoinMessages(t *testing.T) {
	msg := joinMessages([]string{"first line", "second line", "third"})
	if msg != "first line\nsecond line\nthird" {
		t.Fatalf("unexpected joinMessages result: %q", msg)
	}
}
