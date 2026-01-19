package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/s33g/proj/internal/docker"
	"github.com/s33g/proj/internal/project"
	"github.com/s33g/proj/internal/scripts"
	"github.com/s33g/proj/internal/tui"
)

// Action styles
var (
	actionItemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	actionSelectedStyle = lipgloss.NewStyle().PaddingLeft(1).Foreground(tui.Primary).Bold(true)
)

// Action represents an action that can be performed on a project
type Action struct {
	ID       string
	Label    string
	Desc     string
	Icon     string
	Command  string   // For script actions, the command to run
	Source   string   // Source of the script (package.json, Makefile, etc)
	IsSubmenu bool    // Whether this action opens a submenu
	Children []Action // Submenu actions
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
	hasIcon := a.Icon != ""
	if hasIcon {
		label = fmt.Sprintf("%s  %s", a.Icon, label)
	}

	// Add arrow indicator for submenus
	if a.IsSubmenu {
		label = label + " →"
	}

	if isSelected {
		line.WriteString(actionSelectedStyle.Render("▸ " + label))
	} else {
		line.WriteString(actionItemStyle.Render("  " + label))
	}

	// Description on next line with proper indentation
	// Account for: "  " prefix (2) + emoji (2 display width) + "  " separator (2) = 6 total
	// Or just "  " prefix (2) if no icon
	line.WriteString("\n")
	if a.Desc != "" {
		descIndent := 4 // Base padding from style
		if hasIcon {
			descIndent += 4 // Add space for emoji (2 chars wide) + "  " separator
		}
		descStyle := lipgloss.NewStyle().Foreground(tui.Muted).PaddingLeft(descIndent)
		line.WriteString(descStyle.Render(a.Desc))
	}

	_, _ = fmt.Fprint(w, line.String())
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
			ID:    "open-editor",
			Label: "Open in Editor",
			Desc:  "Open project in configured editor",
			Icon:  "🚀",
		},
		{
			ID:    "cd",
			Label: "Change Directory",
			Desc:  "Navigate to project directory",
			Icon:  "📂",
		},
	}

	// Git actions (if enabled and is git repo)
	if gitEnabled && proj.IsGitRepo {
		actions = append(actions,
			Action{
				ID:    "git-log",
				Label: "View Git Log",
				Desc:  "Show recent commits",
				Icon:  "🔍",
			},
			Action{
				ID:    "git-pull",
				Label: "Git Pull",
				Desc:  "Pull latest changes",
				Icon:  "🔄",
			},
			Action{
				ID:    "git-branch",
				Label: "Switch Branch",
				Desc:  "Checkout a different branch",
				Icon:  "🌿",
			},
		)
	} else if gitEnabled && !proj.IsGitRepo {
		// Show git init for non-git directories
		actions = append(actions, Action{
			ID:    "git-init",
			Label: "Git Init",
			Desc:  "Initialize a new git repository",
			Icon:  "🌱",
		})
	}

	// Detect and add project scripts
	detectedScripts := scripts.Detect(proj.Path, proj.Language)
	if len(detectedScripts) > 0 {
		// Group scripts by source
		sourceGroups := make(map[string][]scripts.Script)
		for _, s := range detectedScripts {
			sourceGroups[s.Source] = append(sourceGroups[s.Source], s)
		}

		// Create submenus for each source with multiple scripts
		for source, scriptsInSource := range sourceGroups {
			if len(scriptsInSource) >= 3 {
				// Create submenu for sources with 3+ scripts
				icon := getScriptIcon(source)
				sourceName := getSourceDisplayName(source)

				children := make([]Action, len(scriptsInSource))
				for i, s := range scriptsInSource {
					desc := s.Desc
					if desc == "" {
						desc = s.Command
					}
					children[i] = Action{
						ID:      s.ID,
						Label:   s.Name,
						Desc:    desc,
						Icon:    "▸",
						Command: s.Command,
						Source:  s.Source,
					}
				}

				actions = append(actions, Action{
					ID:        "submenu-" + source,
					Label:     sourceName,
					Desc:      fmt.Sprintf("%d available commands", len(scriptsInSource)),
					Icon:      icon,
					IsSubmenu: true,
					Children:  children,
				})
			} else {
				// Add individual scripts for sources with < 3 scripts
				for _, s := range scriptsInSource {
					icon := getScriptIcon(s.Source)
					desc := s.Desc
					if desc == "" {
						desc = s.Command
					}
					actions = append(actions, Action{
						ID:      s.ID,
						Label:   s.Name,
						Desc:    desc,
						Icon:    icon,
						Command: s.Command,
						Source:  s.Source,
					})
				}
			}
		}
	}

	// Docker actions - group into submenu if present
	if proj.HasDockerfile || proj.HasCompose {
		dockerInfo, err := docker.Detect(proj.Path)
		if err == nil {
			dockerActions := docker.GetActionsForProject(dockerInfo)
			if len(dockerActions) > 0 {
				// Group Docker actions by type
				var dockerChildren []Action
				var composeChildren []Action

				for _, da := range dockerActions {
					action := Action{
						ID:    da.ID,
						Label: da.Name,
						Desc:  da.Description,
						Icon:  "",
					}

					if strings.HasPrefix(da.ID, "compose-") {
						composeChildren = append(composeChildren, action)
					} else {
						dockerChildren = append(dockerChildren, action)
					}
				}

				// Add Docker submenu if there are docker actions
				if len(dockerChildren) > 0 {
					actions = append(actions, Action{
						ID:        "submenu-docker",
						Label:     "Docker",
						Desc:      fmt.Sprintf("%d container actions", len(dockerChildren)),
						Icon:      "🐳",
						IsSubmenu: true,
						Children:  dockerChildren,
					})
				}

				// Add Compose submenu if there are compose actions
				if len(composeChildren) > 0 {
					actions = append(actions, Action{
						ID:        "submenu-compose",
						Label:     "Docker Compose",
						Desc:      fmt.Sprintf("%d service actions", len(composeChildren)),
						Icon:      "🐙",
						IsSubmenu: true,
						Children:  composeChildren,
					})
				}
			}
		}
	}

	// General actions
	actions = append(actions,
		Action{
			ID:    "install-deps",
			Label: "Install Dependencies",
			Desc:  "Run package manager install",
			Icon:  "📦",
		},
		Action{
			ID:    "clean",
			Label: "Clean Build Artifacts",
			Desc:  "Remove build directories",
			Icon:  "🗑️",
		},
		Action{
			ID:    "back",
			Label: "← Back",
			Desc:  "Return to project list",
			Icon:  "",
		},
	)

	return actions
}

// getScriptIcon returns an icon based on the script source
func getScriptIcon(source string) string {
	switch source {
	case "package.json":
		return "📜"
	case "Makefile":
		return "⚙️"
	case "justfile":
		return "📋"
	case "go":
		return "🔵"
	case "cargo":
		return "🦀"
	case "poetry", "pip", "python", "pytest":
		return "🐍"
	case "django":
		return "🎸"
	case "bundler", "rake", "rails":
		return "💎"
	default:
		if strings.HasSuffix(source, "/") {
			return "📄" // Shell scripts in directories
		}
		return "▶️"
	}
}

// getSourceDisplayName returns a friendly display name for a script source
func getSourceDisplayName(source string) string {
	switch source {
	case "package.json":
		return "npm Scripts"
	case "Makefile":
		return "Make"
	case "justfile":
		return "Just"
	case "go":
		return "Go Commands"
	case "cargo":
		return "Cargo"
	case "poetry":
		return "Poetry"
	case "pip":
		return "Pip"
	case "python":
		return "Python"
	case "pytest":
		return "Pytest"
	case "django":
		return "Django"
	case "bundler":
		return "Bundler"
	case "rake":
		return "Rake"
	case "rails":
		return "Rails"
	default:
		return source
	}
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
