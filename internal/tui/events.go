package tui

import tea "github.com/charmbracelet/bubbletea"

func waitForChunk(ch <-chan Chunk) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

func waitForError(ch <-chan error) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}
