package tui

import "github.com/charmbracelet/lipgloss"

var (
	outputBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#89b4fa")).
				Padding(0, 1)

	inputBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#a6e3a1")).
				Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086")).
			Bold(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cba6f7")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086"))

	thinkingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f9e2af")).
			Bold(true)

	toolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94e2d5")).
			Bold(true)

	answerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))
)

func renderWithBorder(content string, style lipgloss.Style, width, height int) string {
	contentWidth := max(0, width-4)
	contentHeight := max(0, height-2)
	return style.Width(contentWidth).Height(contentHeight).Render(content)
}
