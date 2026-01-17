package views

import (
	"fmt"

	"github.com/cjennings/proj/internal/tui"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NewProjectModel is the model for creating a new project
type NewProjectModel struct {
	textInput textinput.Model
	err       error
	width     int
	height    int
}

// NewNewProjectModel creates a new project creation model
func NewNewProjectModel() NewProjectModel {
	ti := textinput.New()
	ti.Placeholder = "my-new-project"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return NewProjectModel{
		textInput: ti,
	}
}

func (m NewProjectModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m NewProjectModel) Update(msg tea.Msg) (NewProjectModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m NewProjectModel) View() string {
	title := tui.TitleStyle.Render("📝 Create New Project")
	
	prompt := lipgloss.NewStyle().
		Foreground(tui.Muted).
		Render("Enter project name:")
	
	input := m.textInput.View()
	
	help := tui.HelpStyle.Render("enter: create  •  esc: cancel")
	
	errMsg := ""
	if m.err != nil {
		errMsg = "\n" + tui.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		prompt,
		input,
		errMsg,
		"",
		help,
	)
}

// Value returns the current input value
func (m NewProjectModel) Value() string {
	return m.textInput.Value()
}

// SetError sets an error message
func (m *NewProjectModel) SetError(err error) {
	m.err = err
}

// SetSize sets the size of the view
func (m *NewProjectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
