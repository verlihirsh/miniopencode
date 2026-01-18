package tui

import tea "github.com/charmbracelet/bubbletea"

// streamClosed signals that the SSE stream has been closed
type streamClosed struct{}

func waitForChunk(ch <-chan Chunk) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return streamClosed{}
		}
		return chunk
	}
}

func waitForError(ch <-chan error) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		err, ok := <-ch
		if !ok {
			return streamClosed{}
		}
		return err
	}
}
