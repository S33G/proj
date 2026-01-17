package project

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/s33g/proj/internal/config"
	"github.com/s33g/proj/internal/git"
	"github.com/s33g/proj/internal/language"
)

// Project represents a code project
type Project struct {
	Name         string
	Path         string
	Language     string
	GitBranch    string
	GitDirty     bool
	IsGitRepo    bool
	LastModified time.Time
}

// Scanner scans directories for projects
type Scanner struct {
	excludePatterns []string
	showHidden      bool
}

// NewScanner creates a new project scanner
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		excludePatterns: cfg.ExcludePatterns,
		showHidden:      cfg.Display.ShowHiddenDirs,
	}
}

// Scan scans a directory for projects
func (s *Scanner) Scan(reposPath string) ([]*Project, error) {
	// Expand path
	expandedPath := config.ExpandPath(reposPath)

	// Read directory entries
	entries, err := os.ReadDir(expandedPath)
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, 0)

	for _, entry := range entries {
		// Skip files, only process directories
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories unless enabled
		if !s.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip excluded patterns
		if s.isExcluded(entry.Name()) {
			continue
		}

		projectPath := filepath.Join(expandedPath, entry.Name())

		// Get project metadata
		project, err := s.scanProject(entry.Name(), projectPath)
		if err != nil {
			// Skip projects we can't read
			continue
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// scanProject scans a single project directory
func (s *Scanner) scanProject(name, path string) (*Project, error) {
	project := &Project{
		Name: name,
		Path: path,
	}

	// Get file info for last modified time
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	project.LastModified = info.ModTime()

	// Detect language
	lang, err := language.Detect(path)
	if err == nil {
		project.Language = lang
	} else {
		project.Language = "Unknown"
	}

	// Get git status
	gitStatus, err := git.GetStatus(path)
	if err == nil {
		project.IsGitRepo = gitStatus.IsRepo
		project.GitBranch = gitStatus.Branch
		project.GitDirty = gitStatus.IsDirty
	}

	return project, nil
}

// isExcluded checks if a directory name should be excluded
func (s *Scanner) isExcluded(name string) bool {
	for _, pattern := range s.excludePatterns {
		if name == pattern {
			return true
		}
	}
	return false
}

// SortBy defines sort orders for projects
type SortBy string

const (
	SortByName         SortBy = "name"
	SortByLastModified SortBy = "lastModified"
)

// Sort sorts projects by the specified criteria
func Sort(projects []*Project, by SortBy) []*Project {
	sorted := make([]*Project, len(projects))
	copy(sorted, projects)

	switch by {
	case SortByName:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})
	case SortByLastModified:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].LastModified.After(sorted[j].LastModified)
		})
	}

	return sorted
}

// Filter filters projects by a search query
func Filter(projects []*Project, query string) []*Project {
	if query == "" {
		return projects
	}

	lowerQuery := strings.ToLower(query)
	filtered := make([]*Project, 0)

	for _, project := range projects {
		if strings.Contains(strings.ToLower(project.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(project.Language), lowerQuery) ||
			strings.Contains(strings.ToLower(project.GitBranch), lowerQuery) {
			filtered = append(filtered, project)
		}
	}

	return filtered
}
