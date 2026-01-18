package tui

import (
	"context"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) sendInput() (Model, tea.Cmd) {
	text := m.currentInputText()
	if strings.TrimSpace(text) == "" {
		return m, nil
	}

	if m.sending {
		return m, nil
	}

	m.transcript.AddUserMessage(text)
	m.transcript.EnsureAssistantMessage("")
	m.viewport.SetContent(m.transcript.Render(m.showThinking, m.showTools, m.spinner.View(), m.sending))
	m.followOutput = true
	m.viewport.GotoBottom()

	m.sending = true

	cmd := func() tea.Msg {
		if m.streamer == nil {
			return sendComplete{}
		}
		log.Printf("tui: send prompt start session=%s len=%d", m.sessionID, len(text))
		ctx := context.Background()
		err := m.streamer.SendPrompt(ctx, m.sessionID, text, m.promptCfg)
		if err != nil {
			log.Printf("tui: send prompt error session=%s err=%v", m.sessionID, err)
			return err
		}
		log.Printf("tui: send prompt accepted session=%s", m.sessionID)
		return sendComplete{}
	}
	return m, cmd
}

type sendComplete struct{}

func (m Model) handleSendComplete() Model {
	m = m.clearInput()
	m.sending = false
	return m
}

func (m Model) clearInput() Model {
	m.textinput.SetValue("")
	return m
}

func (m Model) currentInputText() string {
	return m.textinput.Value()
}
