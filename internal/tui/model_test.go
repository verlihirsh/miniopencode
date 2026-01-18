package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestResizeClamps(t *testing.T) {
	cfg := DefaultUIConfig()
	cfg.InputHeight = 5
	m := NewModel(cfg)
	m.width = 80
	m.height = 24
	m.applySizes()

	anyM, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	m = anyM.(Model)
	anyM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	m = anyM.(Model)
	if m.inputHeight <= 5 {
		t.Fatalf("expected inputHeight to grow, got %d", m.inputHeight)
	}

	for i := 0; i < 20; i++ {
		anyM, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
		m = anyM.(Model)
		anyM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
		m = anyM.(Model)
	}
	if m.inputHeight < minInputHeight {
		t.Fatalf("expected clamped inputHeight, got %d", m.inputHeight)
	}
}

func TestTypewriterBuffer(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)
	m.width = 80
	m.height = 24
	m.applySizes()

	chunk := Chunk{
		Kind:      ChunkAnswer,
		Text:      "Hello world",
		PartID:    "part-1",
		MessageID: "msg-1",
	}
	m = m.bufferChunk(chunk)

	if len(m.tw.buf) != 11 {
		t.Fatalf("expected 11 runes in buffer, got %d", len(m.tw.buf))
	}
	if m.tw.partID != "part-1" {
		t.Fatalf("expected partID 'part-1', got %s", m.tw.partID)
	}
}

func TestTypewriterTickDrains(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)
	m.width = 80
	m.height = 24
	m.applySizes()

	m.tw.buf = []rune("Hello")
	m.tw.partID = "part-1"
	m.tw.msgID = "msg-1"
	m.transcript.EnsureAssistantMessage("msg-1")

	var cmd tea.Cmd
	for len(m.tw.buf) > 0 {
		var anyM tea.Model
		anyM, cmd = m.handleTypewriterTick()
		m = anyM.(Model)
	}

	if len(m.tw.buf) != 0 {
		t.Fatalf("expected empty buffer after ticks, got %d", len(m.tw.buf))
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd when buffer empty")
	}
}

func TestThinkingChunkSkipsTypewriter(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)
	m.width = 80
	m.height = 24
	m.applySizes()

	chunk := Chunk{
		Kind:      ChunkThinking,
		Text:      "Thinking...",
		PartID:    "part-1",
		MessageID: "msg-1",
	}
	m = m.bufferChunk(chunk)

	if len(m.tw.buf) != 0 {
		t.Fatalf("thinking chunks should not buffer, got %d runes", len(m.tw.buf))
	}
}
