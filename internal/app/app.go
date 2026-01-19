package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/s33g/proj/internal/actions"
	"github.com/s33g/proj/internal/config"
	"github.com/s33g/proj/internal/git"
	"github.com/s33g/proj/internal/project"
	"github.com/s33g/proj/internal/tui"
	"github.com/s33g/proj/internal/tui/views"
	"github.com/s33g/proj/pkg/plugin"
)

// View represents the current view state
type View int

const (
	ViewLoading View = iota
	ViewProjects
	ViewActions
	ViewNewProject
	ViewExecuting
	ViewResult
	ViewBranches
	ViewConfirmStash
)

// Model is the main application model
type Model struct {
	config          *config.Config
	pluginRegistry  *plugin.Registry
	view            View
	projects        []*project.Project
	selectedProject *project.Project
	projectList     views.ProjectListModel
	actionMenu      views.ActionMenuModel
	submenuStack    []views.ActionMenuModel // Stack for nested submenus
	newProject      views.NewProjectModel
	resultViewport  viewport.Model
	resultTitle     string
	resultSuccess   bool
	branchList      list.Model
	targetBranch    string
	keys            tui.KeyMap
	width           int
	height          int
	err             error
	message         string
	ready           bool
	cdPath          string   // Path to change to on exit
	execCmd         []string // Command to exec on exit
}

// New creates a new application model
func New(cfg *config.Config) Model {
	// Setup plugin registry
	configDir, _ := config.ConfigDir()
	pluginsDir := filepath.Join(configDir, "plugins")

	// Also check for plugins in the project directory (for development)
	if cwd, err := os.Getwd(); err == nil {
		devPluginsDir := filepath.Join(cwd, "plugins")
		if stat, err := os.Stat(devPluginsDir); err == nil && stat.IsDir() {
			pluginsDir = devPluginsDir
		}
	}

	registry := plugin.NewRegistry(pluginsDir, configDir, cfg.Plugins.Enabled, cfg.Plugins.Config)

	// Load plugins (ignore errors for now)
	_ = registry.LoadAll()

	return Model{
		config:         cfg,
		pluginRegistry: registry,
		view:           ViewLoading,
		keys:           tui.DefaultKeyMap(),
	}
}

type errMsg error
type projectsLoadedMsg []*project.Project
type actionCompleteMsg struct {
	success     bool
	message     string
	actionLabel string
	cdPath      string
	execCmd     []string
}
type branchesLoadedMsg []string
type branchSwitchedMsg struct {
	success bool
	message string
	branch  string
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadProjects(m.config),
		tea.EnterAltScreen,
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.updateSizes()
		return m, nil

	case projectsLoadedMsg:
		m.projects = []*project.Project(msg)
		if len(m.projects) > 0 {
			m.projectList = views.NewProjectListModel(m.projects)
			m.view = ViewProjects
			m.updateSizes() // Call after setting view so size is applied
		} else {
			m.message = "No projects found"
			m.view = ViewProjects
		}
		return m, nil

	case actionCompleteMsg:
		if msg.cdPath != "" {
			m.cdPath = msg.cdPath
			return m, tea.Quit
		}
		if len(msg.execCmd) > 0 {
			m.execCmd = msg.execCmd
			return m, tea.Quit
		}
		// Show result in a dedicated view
		m.resultTitle = msg.actionLabel
		m.resultSuccess = msg.success
		m.resultViewport = viewport.New(m.width-4, m.height-10)
		m.resultViewport.SetContent(msg.message)
		m.view = ViewResult
		return m, nil

	case branchesLoadedMsg:
		// Create branch list
		items := make([]list.Item, len(msg))
		for i, branch := range msg {
			items[i] = branchItem(branch)
		}
		m.branchList = list.New(items, branchDelegate{}, m.width-4, m.height-10)
		m.branchList.Title = "Switch Branch"
		m.branchList.SetShowStatusBar(false)
		m.branchList.SetFilteringEnabled(true)
		m.branchList.Styles.Title = tui.TitleStyle
		m.view = ViewBranches
		return m, nil

	case branchSwitchedMsg:
		if msg.success {
			// Update project's branch info
			m.selectedProject.GitBranch = msg.branch
			m.selectedProject.GitDirty = false
		}
		m.resultTitle = "Switch Branch"
		m.resultSuccess = msg.success
		m.resultViewport = viewport.New(m.width-4, m.height-10)
		m.resultViewport.SetContent(msg.message)
		m.view = ViewResult
		return m, nil

	case errMsg:
		m.err = msg
		m.view = ViewProjects
		return m, nil
	}

	// Delegate to sub-views
	return m.updateView(msg)
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case ViewProjects:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.New):
			m.newProject = views.NewNewProjectModel()
			m.view = ViewNewProject
			m.updateSizes()
			return m, m.newProject.Init()
		case key.Matches(msg, m.keys.Enter):
			if m.selectedProject = m.projectList.SelectedProject(); m.selectedProject != nil {
				// Get built-in actions
				actions := views.DefaultActions(m.selectedProject, m.config.Actions.EnableGitOperations, m.config.Actions.EnableTestRunner)

				// Get plugin actions
				if m.pluginRegistry != nil {
					pluginActions := m.getPluginActions(m.selectedProject)
					// Insert plugin actions before the "back" action
					if len(actions) > 0 {
						// Find the back action
						backIdx := len(actions) - 1
						for i, a := range actions {
							if a.ID == "back" {
								backIdx = i
								break
							}
						}
						// Insert plugin actions before back
						actions = append(actions[:backIdx], append(pluginActions, actions[backIdx:]...)...)
					} else {
						actions = append(actions, pluginActions...)
					}
				}

				m.actionMenu = views.NewActionMenuModel(m.selectedProject, actions)
				m.view = ViewActions
				m.updateSizes()
			}
			return m, nil
		default:
			// Pass other keys (including arrows) to the list for navigation
			var cmd tea.Cmd
			m.projectList, cmd = m.projectList.Update(msg)
			return m, cmd
		}

	case ViewActions:
		switch {
		case key.Matches(msg, m.keys.Back):
			// Check if we're in a submenu
			if len(m.submenuStack) > 0 {
				// Pop back to parent menu
				m.actionMenu = m.submenuStack[len(m.submenuStack)-1]
				m.submenuStack = m.submenuStack[:len(m.submenuStack)-1]
				m.updateSizes()
			} else {
				// Back to project list
				m.view = ViewProjects
				m.selectedProject = nil
			}
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Enter):
			if action := m.actionMenu.SelectedAction(); action != nil {
				if action.ID == "back" {
					// Check if we're in a submenu
					if len(m.submenuStack) > 0 {
						// Pop back to parent menu
						m.actionMenu = m.submenuStack[len(m.submenuStack)-1]
						m.submenuStack = m.submenuStack[:len(m.submenuStack)-1]
						m.updateSizes()
					} else {
						// Back to project list
						m.view = ViewProjects
						m.selectedProject = nil
					}
					return m, nil
				}

				// Check if this is a submenu action
				if action.IsSubmenu {
					// Push current menu onto stack
					m.submenuStack = append(m.submenuStack, m.actionMenu)

					// Create new menu with submenu items
					submenuActions := append(action.Children, views.Action{
						ID:    "back",
						Label: "← Back",
						Desc:  "Return to previous menu",
						Icon:  "",
					})
					m.actionMenu = views.NewActionMenuModel(m.selectedProject, submenuActions)
					m.updateSizes()
					return m, nil
				}

				// Special handling for git-branch - show interactive picker
				if action.ID == "git-branch" {
					m.view = ViewExecuting
					m.message = "Loading branches..."
					return m, loadBranches(m.selectedProject.Path)
				}
				m.view = ViewExecuting
				m.message = fmt.Sprintf("Executing: %s...", action.Label)
				return m, executeAction(action.ID, action.Label, action.Command, m.selectedProject, m.config, m.pluginRegistry)
			}
			return m, nil
		default:
			// Pass other keys (including arrows) to the menu for navigation
			var cmd tea.Cmd
			m.actionMenu, cmd = m.actionMenu.Update(msg)
			return m, cmd
		}

	case ViewNewProject:
		switch {
		case key.Matches(msg, m.keys.Back):
			m.view = ViewProjects
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			projectName := m.newProject.Value()
			if projectName != "" {
				m.view = ViewExecuting
				m.message = fmt.Sprintf("Creating project: %s...", projectName)
				return m, createProject(projectName, m.config.ReposPath)
			}
			return m, nil
		}

	case ViewResult:
		switch {
		case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Quit):
			m.view = ViewActions
			return m, nil
		default:
			// Pass keys to viewport for scrolling
			var cmd tea.Cmd
			m.resultViewport, cmd = m.resultViewport.Update(msg)
			return m, cmd
		}

	case ViewBranches:
		switch {
		case key.Matches(msg, m.keys.Back):
			m.view = ViewActions
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Enter):
			if item, ok := m.branchList.SelectedItem().(branchItem); ok {
				m.targetBranch = string(item)
				// Check if repo is dirty
				dirty, _ := git.IsDirty(m.selectedProject.Path)
				if dirty {
					m.view = ViewConfirmStash
					return m, nil
				}
				// Not dirty, switch directly
				m.view = ViewExecuting
				m.message = fmt.Sprintf("Switching to %s...", m.targetBranch)
				return m, switchBranch(m.selectedProject.Path, m.targetBranch, false)
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.branchList, cmd = m.branchList.Update(msg)
			return m, cmd
		}

	case ViewConfirmStash:
		switch msg.String() {
		case "y", "Y":
			// Stash and switch
			m.view = ViewExecuting
			m.message = fmt.Sprintf("Stashing changes and switching to %s...", m.targetBranch)
			return m, switchBranch(m.selectedProject.Path, m.targetBranch, true)
		case "n", "N", "esc", "q":
			// Cancel, go back to branches
			m.view = ViewBranches
			return m, nil
		}
	}

	return m, nil
}

// updateView updates the current sub-view
func (m Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.view {
	case ViewProjects:
		if len(m.projects) > 0 {
			m.projectList, cmd = m.projectList.Update(msg)
		}
	case ViewActions:
		m.actionMenu, cmd = m.actionMenu.Update(msg)
	case ViewNewProject:
		m.newProject, cmd = m.newProject.Update(msg)
	}

	return m, cmd
}

// updateSizes updates sizes for all sub-views
func (m *Model) updateSizes() {
	// Don't update if we don't have window dimensions yet
	if m.width == 0 || m.height == 0 {
		return
	}

	contentHeight := m.height - 10 // Reserve space for header and help
	if contentHeight < 5 {
		contentHeight = 5 // Minimum height
	}

	if len(m.projects) > 0 {
		m.projectList.SetSize(m.width-4, contentHeight)
	}
	if m.view == ViewActions {
		m.actionMenu.SetSize(m.width-4, contentHeight)
	}
	if m.view == ViewNewProject {
		m.newProject.SetSize(m.width-4, contentHeight)
	}
}

// View renders the current view
func (m Model) View() string {
	if !m.ready {
		return ""
	}

	switch m.view {
	case ViewLoading:
		return tui.ContainerStyle.Render("Loading projects...")

	case ViewProjects:
		return m.renderProjectsView()

	case ViewActions:
		return m.renderActionsView()

	case ViewNewProject:
		return tui.ContainerStyle.Render(m.newProject.View())

	case ViewExecuting:
		return tui.ContainerStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				tui.TitleStyle.Render("⏳ Working..."),
				"",
				m.message,
			),
		)

	case ViewResult:
		return m.renderResultView()

	case ViewBranches:
		return m.renderBranchesView()

	case ViewConfirmStash:
		return m.renderConfirmStashView()
	}

	return ""
}

// renderProjectsView renders the projects list view
func (m Model) renderProjectsView() string {
	header := views.Header(m.config.ReposPath, len(m.projects))

	content := ""
	if len(m.projects) == 0 {
		content = tui.SubtitleStyle.Render("No projects found. Press 'n' to create a new one.")
	} else {
		content = m.projectList.View()
	}

	help := tui.HelpStyle.Render("↑/↓: navigate  •  enter: select  •  n: new project  •  q: quit")

	errorMsg := ""
	if m.err != nil {
		errorMsg = "\n" + tui.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	if m.message != "" {
		errorMsg = "\n" + tui.SuccessStyle.Render(m.message)
	}

	return tui.ContainerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			content,
			errorMsg,
			"",
			help,
		),
	)
}

// renderActionsView renders the actions menu view
func (m Model) renderActionsView() string {
	header := views.ActionHeader(
		m.selectedProject.Name,
		m.selectedProject.Language,
		m.selectedProject.GitBranch,
		m.selectedProject.GitDirty,
	)

	content := m.actionMenu.View()

	// Update help text based on whether we're in a submenu
	helpText := "↑/↓: navigate  •  enter: execute  •  esc: back  •  q: quit"
	if len(m.submenuStack) > 0 {
		helpText = "↑/↓: navigate  •  enter: select  •  esc: back to menu  •  q: quit"
	}
	help := tui.HelpStyle.Render(helpText)

	return tui.ContainerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			content,
			"",
			help,
		),
	)
}

// renderResultView renders the action result view
func (m Model) renderResultView() string {
	// Status icon and title
	icon := "✓"
	titleStyle := tui.SuccessStyle
	if !m.resultSuccess {
		icon = "✗"
		titleStyle = tui.ErrorStyle
	}

	title := titleStyle.Render(fmt.Sprintf("%s %s", icon, m.resultTitle))

	// Scroll indicator
	scrollInfo := ""
	if m.resultViewport.TotalLineCount() > m.resultViewport.Height {
		scrollInfo = tui.SubtitleStyle.Render(
			fmt.Sprintf(" (%d%%)", int(m.resultViewport.ScrollPercent()*100)),
		)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Left, title, scrollInfo)

	// Content in a bordered box
	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1).
		Width(m.width - 6).
		Render(m.resultViewport.View())

	help := tui.HelpStyle.Render("↑/↓: scroll  •  esc/q: close")

	return tui.ContainerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			content,
			"",
			help,
		),
	)
}

// renderBranchesView renders the branch selection view
func (m Model) renderBranchesView() string {
	header := views.ActionHeader(
		m.selectedProject.Name,
		m.selectedProject.Language,
		m.selectedProject.GitBranch,
		m.selectedProject.GitDirty,
	)

	content := m.branchList.View()
	help := tui.HelpStyle.Render("↑/↓: navigate  •  /: filter  •  enter: switch  •  esc: back")

	return tui.ContainerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			content,
			"",
			help,
		),
	)
}

// renderConfirmStashView renders the stash confirmation dialog
func (m Model) renderConfirmStashView() string {
	header := tui.TitleStyle.Render("⚠️  Uncommitted Changes")

	message := fmt.Sprintf(
		"You have uncommitted changes in %s.\n\n"+
			"Do you want to stash them before switching to '%s'?\n\n"+
			"  [Y] Yes, stash and switch\n"+
			"  [N] No, cancel",
		m.selectedProject.Name,
		m.targetBranch,
	)

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Render(message)

	help := tui.HelpStyle.Render("y: stash and switch  •  n/esc: cancel")

	return tui.ContainerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			content,
			"",
			help,
		),
	)
}

// GetCdPath returns the path to change to on exit
func (m Model) GetCdPath() string {
	return m.cdPath
}

// GetExecCmd returns the command to execute on exit
func (m Model) GetExecCmd() []string {
	return m.execCmd
}

// loadProjects loads projects from the repos path
func loadProjects(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		scanner := project.NewScanner(cfg)
		projects, err := scanner.Scan(cfg.ReposPath)
		if err != nil {
			return errMsg(err)
		}

		// Sort projects
		sortBy := project.SortBy(cfg.Display.SortBy)
		projects = project.Sort(projects, sortBy)

		return projectsLoadedMsg(projects)
	}
}

// executeAction executes an action
func executeAction(actionID string, actionLabel string, actionCommand string, proj *project.Project, cfg *config.Config, registry *plugin.Registry) tea.Cmd {
	return func() tea.Msg {
		// If action has a command, execute it directly
		if actionCommand != "" {
			executor := actions.NewExecutor(cfg)
			result := executor.ExecuteCommand(actionCommand, proj)
			return actionCompleteMsg{
				success:     result.Success,
				message:     result.Message,
				actionLabel: actionLabel,
				cdPath:      result.CdPath,
				execCmd:     result.ExecCmd,
			}
		}

		// Try plugin actions first
		if registry != nil {
			pluginProj := projectToPlugin(proj)
			result, err := registry.ExecuteAction(actionID, pluginProj)
			if err == nil && result != nil {
				return actionCompleteMsg{
					success:     result.Success,
					message:     result.Message,
					actionLabel: actionLabel,
					cdPath:      result.CdPath,
					execCmd:     result.ExecCmd,
				}
			}
		}

		// Fall back to built-in actions
		executor := actions.NewExecutor(cfg)
		result := executor.Execute(actionID, proj)

		return actionCompleteMsg{
			success:     result.Success,
			message:     result.Message,
			actionLabel: actionLabel,
			cdPath:      result.CdPath,
			execCmd:     result.ExecCmd,
		}
	}
}

// getPluginActions gets actions from plugins for a project
func (m Model) getPluginActions(proj *project.Project) []views.Action {
	if m.pluginRegistry == nil {
		return nil
	}

	pluginProj := projectToPlugin(proj)
	pluginActions := m.pluginRegistry.GetActions(pluginProj)

	// Convert plugin actions to view actions
	viewActions := make([]views.Action, len(pluginActions))
	for i, pa := range pluginActions {
		viewActions[i] = views.Action{
			ID:    pa.ID,
			Label: pa.Label,
			Desc:  pa.Description,
			Icon:  pa.Icon,
		}
	}

	return viewActions
}

// projectToPlugin converts a project.Project to a plugin.Project
func projectToPlugin(proj *project.Project) plugin.Project {
	return plugin.Project{
		Name:      proj.Name,
		Path:      proj.Path,
		Language:  proj.Language,
		GitBranch: proj.GitBranch,
		GitDirty:  proj.GitDirty,
		IsGitRepo: proj.IsGitRepo,
	}
}

// createProject creates a new project directory
func createProject(name string, reposPath string) tea.Cmd {
	return func() tea.Msg {
		expandedPath := config.ExpandPath(reposPath)
		projectPath := filepath.Join(expandedPath, name)

		if err := os.MkdirAll(projectPath, 0755); err != nil {
			return actionCompleteMsg{
				success: false,
				message: fmt.Sprintf("Failed to create project: %v", err),
			}
		}

		return actionCompleteMsg{
			success: true,
			message: fmt.Sprintf("Created project: %s", name),
		}
	}
}

// loadBranches loads git branches for a project
func loadBranches(projectPath string) tea.Cmd {
	return func() tea.Msg {
		branches, err := git.GetBranches(projectPath)
		if err != nil {
			return errMsg(err)
		}
		return branchesLoadedMsg(branches)
	}
}

// switchBranch switches to a different branch, optionally stashing first
func switchBranch(projectPath, branch string, stashFirst bool) tea.Cmd {
	return func() tea.Msg {
		var messages []string

		if stashFirst {
			stashOut, err := git.Stash(projectPath)
			if err != nil {
				return branchSwitchedMsg{
					success: false,
					message: fmt.Sprintf("Failed to stash changes: %v\n%s", err, stashOut),
					branch:  branch,
				}
			}
			messages = append(messages, "Stashed changes: "+stashOut)
		}

		checkoutOut, err := git.Checkout(projectPath, branch)
		if err != nil {
			return branchSwitchedMsg{
				success: false,
				message: fmt.Sprintf("Failed to switch branch: %v\n%s", err, checkoutOut),
				branch:  branch,
			}
		}
		messages = append(messages, fmt.Sprintf("Switched to branch '%s'", branch))
		if checkoutOut != "" {
			messages = append(messages, checkoutOut)
		}

		return branchSwitchedMsg{
			success: true,
			message: fmt.Sprintf("%s\n\nUse 'git stash pop' to restore stashed changes.",
				joinMessages(messages)),
			branch: branch,
		}
	}
}

func joinMessages(msgs []string) string {
	result := ""
	for i, msg := range msgs {
		if i > 0 {
			result += "\n"
		}
		result += msg
	}
	return result
}

// branchItem is a list item for the branch picker
type branchItem string

func (b branchItem) FilterValue() string { return string(b) }
func (b branchItem) Title() string       { return string(b) }
func (b branchItem) Description() string { return "" }

// branchDelegate is a simple delegate for branch items
type branchDelegate struct{}

func (d branchDelegate) Height() int                             { return 1 }
func (d branchDelegate) Spacing() int                            { return 0 }
func (d branchDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d branchDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	branch, ok := item.(branchItem)
	if !ok {
		return
	}

	str := string(branch)

	// Check if this is the current branch (marked with *)
	isCurrent := false
	if len(str) > 0 && str[0] == '*' {
		isCurrent = true
		str = str[1:]
		str = "  " + str + " (current)"
	} else {
		str = "  " + str
	}

	if index == m.Index() {
		str = tui.SelectedStyle.Render("> " + str[2:])
	} else if isCurrent {
		str = tui.SubtitleStyle.Render(str)
	}

	_, _ = fmt.Fprint(w, str)
}
