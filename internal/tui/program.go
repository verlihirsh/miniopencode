package tui

import tea "github.com/charmbracelet/bubbletea"

func newProgram(m Model) *tea.Program {
	return tea.NewProgram(m, tea.WithAltScreen())
}
