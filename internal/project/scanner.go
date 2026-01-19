package project

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/s33g/proj/internal/config"
	"github.com/s33g/proj/internal/docker"
	"github.com/s33g/proj/internal/git"
	"github.com/s33g/proj/internal/language"
)

// Project represents a code project
type Project struct {
	Name            string
	Path            string
	ParentPath      string // Path to parent group (empty if top-level)
	Language        string
	GitBranch       string
	GitDirty        bool
	IsGitRepo       bool
	LastModified    time.Time
	HasDockerfile   bool
	HasCompose      bool
	Depth           int
	SubProjectCount int
	IsGroup         bool // True if this is a folder containing projects (not a project itself)
	Expanded        bool // True if group is expanded to show children
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

// Scan scans a directory for projects (1 level deep for groups)
func (s *Scanner) Scan(reposPath string) ([]*Project, error) {
	expandedPath := config.ExpandPath(reposPath)
	return s.scanWithGroups(expandedPath)
}

// scanWithGroups scans top level and detects groups vs projects
func (s *Scanner) scanWithGroups(basePath string) ([]*Project, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if !s.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if s.isExcluded(entry.Name()) {
			continue
		}

		dirPath := filepath.Join(basePath, entry.Name())
		isProject := isProjectRoot(dirPath)

		// Check if this directory contains projects (making it a group)
		childProjects := s.findChildProjects(dirPath)

		if isProject {
			// It's a project (and possibly also contains sub-projects - monorepo)
			proj, err := s.scanProject(entry.Name(), dirPath, 0)
			if err != nil {
				continue
			}
			proj.SubProjectCount = len(childProjects)
			proj.Expanded = true // Projects with children are expanded by default
			projects = append(projects, proj)

			// Add child projects
			for _, child := range childProjects {
				child.ParentPath = dirPath
				child.Depth = 1
				projects = append(projects, child)
			}
		} else if len(childProjects) > 0 {
			// It's a group folder (not a project itself, but contains projects)
			group := &Project{
				Name:            entry.Name(),
				Path:            dirPath,
				Depth:           0,
				IsGroup:         true,
				SubProjectCount: len(childProjects),
				Expanded:        true,
			}

			// Get last modified from directory
			info, _ := os.Stat(dirPath)
			if info != nil {
				group.LastModified = info.ModTime()
			}

			projects = append(projects, group)

			// Add child projects
			for _, child := range childProjects {
				child.ParentPath = dirPath
				child.Depth = 1
				projects = append(projects, child)
			}
		}
		// If neither a project nor a group, skip it
	}

	return projects, nil
}

// findChildProjects finds all projects in a directory (1 level only)
func (s *Scanner) findChildProjects(dirPath string) []*Project {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil
	}

	projects := make([]*Project, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if !s.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if s.isExcluded(entry.Name()) {
			continue
		}

		childPath := filepath.Join(dirPath, entry.Name())

		if isProjectRoot(childPath) {
			proj, err := s.scanProject(entry.Name(), childPath, 1)
			if err != nil {
				continue
			}
			projects = append(projects, proj)
		}
	}

	return projects
}

// scanProject scans a single project directory
func (s *Scanner) scanProject(name, path string, depth int) (*Project, error) {
	project := &Project{
		Name:  name,
		Path:  path,
		Depth: depth,
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

	// Detect Docker
	dockerInfo, err := docker.Detect(path)
	if err == nil {
		project.HasDockerfile = dockerInfo.HasDockerfile
		project.HasCompose = dockerInfo.HasCompose
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

// isProjectRoot checks if a directory appears to be a project root
// by looking for common project indicators
func isProjectRoot(path string) bool {
	// Check for common project indicators
	indicators := []string{
		".git",           // Git repository
		"go.mod",         // Go project
		"package.json",   // Node.js/JavaScript project
		"Cargo.toml",     // Rust project
		"requirements.txt", // Python project
		"setup.py",       // Python project
		"pyproject.toml", // Python project
		"Gemfile",        // Ruby project
		"pom.xml",        // Java/Maven project
		"build.gradle",   // Java/Gradle project
		"composer.json",  // PHP project
		"CMakeLists.txt", // C/C++ project
		"Makefile",       // General project with Makefile
		".project",       // Eclipse project
		"README.md",      // Common project file
		"README",         // Common project file
	}

	for _, indicator := range indicators {
		indicatorPath := filepath.Join(path, indicator)
		if _, err := os.Stat(indicatorPath); err == nil {
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
	SortByLanguage     SortBy = "language"
)

// Sort sorts projects by the specified criteria while preserving parent-child relationships.
// Top-level items (Depth 0) are sorted, and children stay with their parent.
func Sort(projects []*Project, by SortBy) []*Project {
	// Separate top-level items and their children
	type groupedProject struct {
		parent   *Project
		children []*Project
	}

	groups := make([]*groupedProject, 0)
	groupMap := make(map[string]*groupedProject) // path -> group

	// First pass: identify parents and standalone projects
	for _, p := range projects {
		if p.Depth == 0 {
			g := &groupedProject{parent: p, children: make([]*Project, 0)}
			groups = append(groups, g)
			groupMap[p.Path] = g
		}
	}

	// Second pass: attach children to their parents
	for _, p := range projects {
		if p.Depth > 0 && p.ParentPath != "" {
			if g, ok := groupMap[p.ParentPath]; ok {
				g.children = append(g.children, p)
			}
		}
	}

	// Sort the groups by their parent
	sortFunc := func(i, j int) bool {
		pi := groups[i].parent
		pj := groups[j].parent
		switch by {
		case SortByName:
			return pi.Name < pj.Name
		case SortByLastModified:
			return pi.LastModified.After(pj.LastModified)
		case SortByLanguage:
			// Groups without language sort after projects with language
			langI := pi.Language
			langJ := pj.Language
			if pi.IsGroup {
				langI = "zzz" // Sort groups to end when sorting by language
			}
			if pj.IsGroup {
				langJ = "zzz"
			}
			if langI == langJ {
				return pi.Name < pj.Name
			}
			return langI < langJ
		default:
			return pi.Name < pj.Name
		}
	}
	sort.Slice(groups, sortFunc)

	// Sort children within each group
	for _, g := range groups {
		sort.Slice(g.children, func(i, j int) bool {
			ci := g.children[i]
			cj := g.children[j]
			switch by {
			case SortByName:
				return ci.Name < cj.Name
			case SortByLastModified:
				return ci.LastModified.After(cj.LastModified)
			case SortByLanguage:
				if ci.Language == cj.Language {
					return ci.Name < cj.Name
				}
				return ci.Language < cj.Language
			default:
				return ci.Name < cj.Name
			}
		})
	}

	// Flatten back into a single list
	result := make([]*Project, 0, len(projects))
	for _, g := range groups {
		result = append(result, g.parent)
		result = append(result, g.children...)
	}

	return result
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
