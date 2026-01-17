package views

import (
	"fmt"

	"github.com/cjennings/proj/internal/tui"
	"github.com/charmbracelet/lipgloss"
)

// Header renders the application header
func Header(reposPath string, projectCount int) string {
	title := tui.TitleStyle.Render("📂 proj - Project Navigator")
	subtitle := tui.SubtitleStyle.Render(fmt.Sprintf("Path: %s  •  Projects: %d", reposPath, projectCount))
	
	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
}

// ActionHeader renders the header for the action menu
func ActionHeader(projectName, language string, gitBranch string, gitDirty bool) string {
	title := tui.TitleStyle.Render(fmt.Sprintf("🚀 %s", projectName))
	
	badges := ""
	if language != "" && language != "Unknown" {
		badges += tui.LanguageBadgeStyle.Render(language)
	}
	if gitBranch != "" {
		branchBadge := tui.BadgeStyle.Render(fmt.Sprintf(" %s ", gitBranch))
		badges += branchBadge
		if gitDirty {
			badges += tui.DirtyBadgeStyle.Render(" * ")
		}
	}
	
	if badges != "" {
		badges = "\n" + badges
	}
	
	return title + badges
}
