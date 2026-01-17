package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/s33g/proj/internal/actions"
	"github.com/s33g/proj/internal/config"
	"github.com/s33g/proj/internal/project"
	"github.com/s33g/proj/internal/tui"
	"github.com/s33g/proj/internal/tui/views"
	"github.com/s33g/proj/pkg/plugin"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View represents the current view state
type View int

const (
	ViewLoading View = iota
	ViewProjects
	ViewActions
	ViewNewProject
	ViewExecuting
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
	newProject      views.NewProjectModel
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
	success bool
	message string
	cdPath  string
	execCmd []string
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
		m.message = msg.message
		if msg.success {
			m.view = ViewActions
		} else {
			m.view = ViewActions
		}
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
			m.view = ViewProjects
			m.selectedProject = nil
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Enter):
			if action := m.actionMenu.SelectedAction(); action != nil {
				if action.ID == "back" {
					m.view = ViewProjects
					m.selectedProject = nil
					return m, nil
				}
				m.view = ViewExecuting
				m.message = fmt.Sprintf("Executing: %s...", action.Label)
				return m, executeAction(action.ID, m.selectedProject, m.config, m.pluginRegistry)
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
	help := tui.HelpStyle.Render("↑/↓: navigate  •  enter: execute  •  esc: back  •  q: quit")

	errorMsg := ""
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
func executeAction(actionID string, proj *project.Project, cfg *config.Config, registry *plugin.Registry) tea.Cmd {
	return func() tea.Msg {
		// Try plugin actions first
		if registry != nil {
			pluginProj := projectToPlugin(proj)
			result, err := registry.ExecuteAction(actionID, pluginProj)
			if err == nil && result != nil {
				return actionCompleteMsg{
					success: result.Success,
					message: result.Message,
					cdPath:  result.CdPath,
					execCmd: result.ExecCmd,
				}
			}
		}
		
		// Fall back to built-in actions
		executor := actions.NewExecutor(cfg)
		result := executor.Execute(actionID, proj)
		
		return actionCompleteMsg{
			success: result.Success,
			message: result.Message,
			cdPath:  result.CdPath,
			execCmd: result.ExecCmd,
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
