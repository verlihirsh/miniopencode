package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) sendInput() tea.Cmd {
	text := m.currentInputText()
	if strings.TrimSpace(text) == "" {
		return nil
	}
	m.clearInput()
	return func() tea.Msg {
		if m.streamer == nil {
			return Chunk{Kind: ChunkRaw, Text: "streamer not set"}
		}
		ctx := context.Background()
		err := m.streamer.SendPrompt(ctx, m.sessionID, text, m.promptCfg)
		if err != nil {
			return err
		}
		return Chunk{Kind: ChunkAnswer, Text: text}
	}
}

func (m *Model) clearInput() {
	m.textinput.SetValue("")
	m.textarea.SetValue("")
}

func (m Model) currentInputText() string {
	if m.multiline {
		return m.textarea.Value()
	}
	return m.textinput.Value()
}
