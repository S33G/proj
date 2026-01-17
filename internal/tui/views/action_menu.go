package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/s33g/proj/internal/project"
	"github.com/s33g/proj/internal/tui"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Action styles
var (
	actionItemStyle    = lipgloss.NewStyle().PaddingLeft(2)
	actionSelectedStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(tui.Primary).Bold(true)
	actionDescStyle    = lipgloss.NewStyle().Foreground(tui.Muted).PaddingLeft(4)
)

// Action represents an action that can be performed on a project
type Action struct {
	ID    string
	Label string
	Desc  string
	Icon  string
}

// FilterValue implements list.Item
func (a Action) FilterValue() string {
	return a.Label
}

// Title implements list.Item
func (a Action) Title() string {
	if a.Icon != "" {
		return fmt.Sprintf("%s  %s", a.Icon, a.Label)
	}
	return a.Label
}

// Description implements list.Item
func (a Action) Description() string {
	return a.Desc
}

// actionDelegate is a custom delegate for rendering action items
type actionDelegate struct{}

func (d actionDelegate) Height() int                             { return 2 }
func (d actionDelegate) Spacing() int                            { return 0 }
func (d actionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d actionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	a, ok := listItem.(Action)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Build the line
	var line strings.Builder

	// Icon and label
	label := a.Label
	if a.Icon != "" {
		label = fmt.Sprintf("%s  %s", a.Icon, label)
	}

	if isSelected {
		line.WriteString(actionSelectedStyle.Render("▸ " + label))
	} else {
		line.WriteString(actionItemStyle.Render("  " + label))
	}

	// Description on next line
	line.WriteString("\n")
	if a.Desc != "" {
		line.WriteString(actionDescStyle.Render(a.Desc))
	}

	fmt.Fprint(w, line.String())
}

// ActionMenuModel is the model for the action menu
type ActionMenuModel struct {
	list    list.Model
	project *project.Project
	width   int
	height  int
}

// NewActionMenuModel creates a new action menu model
func NewActionMenuModel(proj *project.Project, actions []Action) ActionMenuModel {
	items := make([]list.Item, len(actions))
	for i, a := range actions {
		items[i] = a
	}

	delegate := actionDelegate{}

	l := list.New(items, delegate, 80, 20)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.Title = lipgloss.NewStyle()

	return ActionMenuModel{
		list:    l,
		project: proj,
		width:   80,
		height:  20,
	}
}

func (m ActionMenuModel) Init() tea.Cmd {
	return nil
}

func (m ActionMenuModel) Update(msg tea.Msg) (ActionMenuModel, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ActionMenuModel) View() string {
	return m.list.View()
}

// SetSize sets the size of the list
func (m *ActionMenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height)
}

// SelectedAction returns the currently selected action
func (m ActionMenuModel) SelectedAction() *Action {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	action := item.(Action)
	return &action
}

// DefaultActions returns the default set of actions
func DefaultActions(proj *project.Project, gitEnabled, testsEnabled bool) []Action {
	actions := []Action{
		{
			ID:   "open-editor",
			Label:       "Open in Editor",
			Desc: "Open project in configured editor",
			Icon:        "🚀",
		},
		{
			ID:   "cd",
			Label:       "Change Directory",
			Desc: "Navigate to project directory",
			Icon:        "📂",
		},
	}

	// Git actions (if enabled and is git repo)
	if gitEnabled && proj.IsGitRepo {
		actions = append(actions,
			Action{
				ID:   "git-log",
				Label:       "View Git Log",
				Desc: "Show recent commits",
				Icon:        "🔍",
			},
			Action{
				ID:   "git-pull",
				Label:       "Git Pull",
				Desc: "Pull latest changes",
				Icon:        "🔄",
			},
			Action{
				ID:   "git-branch",
				Label:       "Switch Branch",
				Desc: "Checkout a different branch",
				Icon:        "🌿",
			},
		)
	}

	// Test action
	if testsEnabled {
		actions = append(actions, Action{
			ID:   "run-tests",
			Label:       "Run Tests",
			Desc: "Execute test suite",
			Icon:        "🧪",
		})
	}

	// General actions
	actions = append(actions,
		Action{
			ID:   "install-deps",
			Label:       "Install Dependencies",
			Desc: "Run package manager install",
			Icon:        "📦",
		},
		Action{
			ID:   "clean",
			Label:       "Clean Build Artifacts",
			Desc: "Remove build directories",
			Icon:        "🗑️",
		},
		Action{
			ID:   "back",
			Label:       "← Back",
			Desc: "Return to project list",
			Icon:        "",
		},
	)

	return actions
}

// ActionMenu renders a simple action menu
func ActionMenu(actions []Action, cursor int) string {
	if len(actions) == 0 {
		return tui.SubtitleStyle.Render("No actions available")
	}

	var b strings.Builder
	for i, a := range actions {
		// Cursor indicator
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}

		// Action label
		label := a.Label
		if a.Icon != "" {
			label = fmt.Sprintf("%s  %s", a.Icon, label)
		}
		if i == cursor {
			label = tui.SelectedStyle.Render(label)
		}

		// Description
		desc := ""
		if a.Desc != "" {
			desc = tui.SubtitleStyle.Render(fmt.Sprintf("   %s", a.Desc))
		}

		b.WriteString(fmt.Sprintf("%s%s\n%s\n", prefix, label, desc))
	}

	return b.String()
}
