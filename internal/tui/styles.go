package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	Primary = lipgloss.AdaptiveColor{Light: "#00CED1", Dark: "#00CED1"}
	Accent  = lipgloss.AdaptiveColor{Light: "#32CD32", Dark: "#32CD32"}
	Error   = lipgloss.AdaptiveColor{Light: "#FF6347", Dark: "#FF6347"}
	Muted   = lipgloss.AdaptiveColor{Light: "#A0A0A0", Dark: "#707070"}
	Border  = lipgloss.AdaptiveColor{Light: "#D0D0D0", Dark: "#404040"}
)

// Styles
var (
	// Title style
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginBottom(1)

	// Selected item style
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			PaddingLeft(2)

	// Normal item style
	NormalStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1).
			MarginTop(1)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	// Badge style
	BadgeStyle = lipgloss.NewStyle().
			Background(Primary).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			MarginRight(1)

	// Git dirty badge
	DirtyBadgeStyle = lipgloss.NewStyle().
			Background(Error).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			MarginRight(1)

	// Language badge
	LanguageBadgeStyle = lipgloss.NewStyle().
				Foreground(Accent).
				Bold(true).
				MarginRight(1)

	// Container style
	ContainerStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Input style
	InputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1).
			MarginTop(1)

	// Focused input style
	FocusedInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(0, 1).
				MarginTop(1).
				Bold(true)
)

// UpdateTheme updates the color scheme from config
func UpdateTheme(primary, accent, errorColor string) {
	Primary = lipgloss.AdaptiveColor{Light: primary, Dark: primary}
	Accent = lipgloss.AdaptiveColor{Light: accent, Dark: accent}
	Error = lipgloss.AdaptiveColor{Light: errorColor, Dark: errorColor}

	// Re-apply colors to styles
	TitleStyle = TitleStyle.Foreground(Primary)
	SelectedStyle = SelectedStyle.Foreground(Primary)
	SuccessStyle = SuccessStyle.Foreground(Accent)
	ErrorStyle = ErrorStyle.Foreground(Error)
	BadgeStyle = BadgeStyle.Background(Primary)
	DirtyBadgeStyle = DirtyBadgeStyle.Background(Error)
	LanguageBadgeStyle = LanguageBadgeStyle.Foreground(Accent)
	InputStyle = InputStyle.BorderForeground(Primary)
	FocusedInputStyle = FocusedInputStyle.BorderForeground(Primary)
}
