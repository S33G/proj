package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/s33g/proj/internal/language"
	"github.com/s33g/proj/internal/project"
	"github.com/s33g/proj/internal/tui"
)

// Item styles for the list
var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(tui.Primary).Bold(true)
	langStyle         = lipgloss.NewStyle().Foreground(tui.Accent)
	branchStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	dirtyStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6347")).Bold(true)
)

// ProjectListItem implements list.Item for projects
type ProjectListItem struct {
	Project *project.Project
}

func (i ProjectListItem) FilterValue() string {
	return i.Project.Name
}

func (i ProjectListItem) Title() string {
	return i.Project.Name
}

func (i ProjectListItem) Description() string {
	parts := []string{}

	// Add language icon and name
	if i.Project.Language != "" && i.Project.Language != "Unknown" {
		icon := language.GetIcon(i.Project.Language)
		parts = append(parts, fmt.Sprintf("%s %s", icon, i.Project.Language))
	}

	// Add git branch
	if i.Project.GitBranch != "" {
		branch := fmt.Sprintf(" %s", i.Project.GitBranch)
		if i.Project.GitDirty {
			branch += "*"
		}
		parts = append(parts, branch)
	}

	return strings.Join(parts, "  ")
}

// itemDelegate is a custom delegate for rendering project items
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ProjectListItem)
	if !ok {
		return
	}

	p := i.Project
	isSelected := index == m.Index()
	isGroup := p.IsGroup
	hasChildren := p.SubProjectCount > 0

	// Build the line
	var line strings.Builder

	// Prefix for selection
	prefix := "  "
	if isSelected {
		prefix = "▸ "
	}

	// Project/group name
	name := p.Name
	if hasChildren {
		name = fmt.Sprintf("%s (%d)", name, p.SubProjectCount)
	}

	// Style based on selection and type
	if isGroup {
		// Pure group (folder containing projects)
		groupStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
		if isSelected {
			line.WriteString(selectedItemStyle.Render(prefix + "📁 " + name))
		} else {
			line.WriteString(itemStyle.Render(prefix + groupStyle.Render("📁 " + name)))
		}
	} else if hasChildren {
		// Monorepo (project with sub-projects)
		if isSelected {
			line.WriteString(selectedItemStyle.Render(prefix + "📦 " + name))
		} else {
			line.WriteString(itemStyle.Render(prefix + "📦 " + name))
		}
	} else if isSelected {
		line.WriteString(selectedItemStyle.Render(prefix + name))
	} else {
		line.WriteString(itemStyle.Render(prefix + name))
	}

	// Calculate padding for alignment
	visibleLen := len(prefix) + len(p.Name)
	if hasChildren {
		visibleLen += len(fmt.Sprintf(" (%d)", p.SubProjectCount))
	}
	if isGroup || hasChildren {
		visibleLen += 3 // icon
	}
	padding := 35 - visibleLen
	if padding < 2 {
		padding = 2
	}
	line.WriteString(strings.Repeat(" ", padding))

	// Only show details for actual projects (not pure groups)
	if !p.IsGroup {
		// Language
		if p.Language != "" && p.Language != "Unknown" {
			icon := language.GetIcon(p.Language)
			line.WriteString(langStyle.Render(fmt.Sprintf("%s %-10s", icon, p.Language)))
		} else {
			line.WriteString(strings.Repeat(" ", 12))
		}

		// Git branch
		if p.GitBranch != "" {
			branch := p.GitBranch
			if len(branch) > 20 {
				branch = branch[:17] + "..."
			}
			line.WriteString(branchStyle.Render(branch))
			if p.GitDirty {
				line.WriteString(dirtyStyle.Render("*"))
			}
		}

		// Docker indicators
		if p.HasCompose {
			line.WriteString("  🐙")
		} else if p.HasDockerfile {
			line.WriteString("  🐳")
		}
	}

	_, _ = fmt.Fprint(w, line.String())
}

// ProjectListModel is the model for the project list view
type ProjectListModel struct {
	list       list.Model
	projects   []*project.Project
	width      int
	height     int
	showAll    bool // When true, show all projects regardless of depth (for group views)
}

// NewProjectListModel creates a new project list model
func NewProjectListModel(projects []*project.Project) ProjectListModel {
	// Use our custom compact delegate
	delegate := itemDelegate{}

	// Initialize with reasonable default size (will be updated on WindowSizeMsg)
	l := list.New([]list.Item{}, delegate, 80, 24)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(true)
	l.Styles.Title = lipgloss.NewStyle()
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(tui.Muted)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(tui.Muted)

	m := ProjectListModel{
		list:     l,
		projects: projects,
		width:    80,
		height:   24,
		showAll:  false, // Main list only shows top-level
	}
	
	// Build the visible items list
	m.RebuildList()
	
	return m
}

// NewGroupListModel creates a project list model for showing group contents (shows all items)
func NewGroupListModel(projects []*project.Project) ProjectListModel {
	// Use our custom compact delegate
	delegate := itemDelegate{}

	// Initialize with reasonable default size
	l := list.New([]list.Item{}, delegate, 80, 24)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(true)
	l.Styles.Title = lipgloss.NewStyle()
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(tui.Muted)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(tui.Muted)

	m := ProjectListModel{
		list:     l,
		projects: projects,
		width:    80,
		height:   24,
		showAll:  true, // Group list shows all items
	}
	
	// Build the visible items list
	m.RebuildList()
	
	return m
}

func (m ProjectListModel) Init() tea.Cmd {
	return nil
}

func (m ProjectListModel) Update(msg tea.Msg) (ProjectListModel, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ProjectListModel) View() string {
	return m.list.View()
}

// SetSize sets the size of the list
func (m *ProjectListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// SelectedProject returns the currently selected project
func (m ProjectListModel) SelectedProject() *project.Project {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	return item.(ProjectListItem).Project
}

// RebuildList rebuilds the list items
func (m *ProjectListModel) RebuildList() {
	visibleProjects := make([]*project.Project, 0)
	for _, p := range m.projects {
		// Show all items if showAll is true, otherwise only top-level
		if m.showAll || p.Depth == 0 {
			visibleProjects = append(visibleProjects, p)
		}
	}

	items := make([]list.Item, len(visibleProjects))
	for i, p := range visibleProjects {
		items[i] = ProjectListItem{Project: p}
	}
	m.list.SetItems(items)
}

// ProjectList renders a simple project list
func ProjectList(projects []*project.Project, cursor int) string {
	if len(projects) == 0 {
		return tui.SubtitleStyle.Render("No projects found")
	}

	var b strings.Builder
	for i, p := range projects {
		// Cursor indicator
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}

		// Add indentation based on depth
		indent := strings.Repeat("  ", p.Depth)

		// Project name with nesting indicator
		name := p.Name
		if p.SubProjectCount > 0 {
			name = name + fmt.Sprintf(" (%d)", p.SubProjectCount)
		}
		if i == cursor {
			name = tui.SelectedStyle.Render(indent + name)
		} else {
			name = indent + name
		}

		// Language icon
		langIcon := ""
		if p.Language != "" && p.Language != "Unknown" {
			icon := language.GetIcon(p.Language)
			langIcon = tui.LanguageBadgeStyle.Render(fmt.Sprintf("%s %s", icon, p.Language))
		}

		// Git info
		gitInfo := ""
		if p.GitBranch != "" {
			branch := p.GitBranch
			if p.GitDirty {
				branch += " *"
			}
			gitInfo = tui.BadgeStyle.Render(fmt.Sprintf(" %s ", branch))
		}

		// Docker info
		dockerInfo := ""
		if p.HasCompose {
			dockerInfo = " 🐙"
		} else if p.HasDockerfile {
			dockerInfo = " 🐳"
		}

		line := fmt.Sprintf("%s%s  %s %s%s", prefix, name, langIcon, gitInfo, dockerInfo)
		b.WriteString(line + "\n")
	}

	return b.String()
}

// Help renders the help text
func Help(keys ...string) string {
	helpText := strings.Join(keys, "  •  ")
	return lipgloss.NewStyle().
		Foreground(tui.Muted).
		Render(helpText)
}
