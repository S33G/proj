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
	cdPath          string // Path to change to on exit
	execCmd         []string // Command to exec on exit
}

// New creates a new application model
func New(cfg *config.Config) Model {
	return Model{
		config: cfg,
		view:   ViewLoading,
		keys:   tui.DefaultKeyMap(),
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
			m.updateSizes()
			m.view = ViewProjects
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
				actions := views.DefaultActions(m.selectedProject, m.config.Actions.EnableGitOperations, m.config.Actions.EnableTestRunner)
				m.actionMenu = views.NewActionMenuModel(m.selectedProject, actions)
				m.view = ViewActions
				m.updateSizes()
			}
			return m, nil
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
				return m, executeAction(action.ID, m.selectedProject, m.config)
			}
			return m, nil
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
	contentHeight := m.height - 10 // Reserve space for header and help

	if m.view == ViewProjects && len(m.projects) > 0 {
		m.projectList.SetSize(m.width-4, contentHeight)
	} else if m.view == ViewActions {
		m.actionMenu.SetSize(m.width-4, contentHeight)
	} else if m.view == ViewNewProject {
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
func executeAction(actionID string, proj *project.Project, cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
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
