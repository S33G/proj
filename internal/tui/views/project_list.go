package views

import (
	"fmt"
	"strings"

	"github.com/cjennings/proj/internal/language"
	"github.com/cjennings/proj/internal/project"
	"github.com/cjennings/proj/internal/tui"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
			branch += " *"
		}
		parts = append(parts, branch)
	}
	
	return strings.Join(parts, "  •  ")
}

// ProjectListModel is the model for the project list view
type ProjectListModel struct {
	list     list.Model
	projects []*project.Project
	width    int
	height   int
}

// NewProjectListModel creates a new project list model
func NewProjectListModel(projects []*project.Project) ProjectListModel {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = ProjectListItem{Project: p}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = tui.SelectedStyle
	delegate.Styles.SelectedDesc = tui.SelectedStyle.Foreground(tui.Muted)

	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return ProjectListModel{
		list:     l,
		projects: projects,
	}
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

		// Project name
		name := p.Name
		if i == cursor {
			name = tui.SelectedStyle.Render(name)
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

		line := fmt.Sprintf("%s%s  %s %s", prefix, name, langIcon, gitInfo)
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
