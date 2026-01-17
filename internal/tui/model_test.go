package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestToggleMultiline(t *testing.T) {
	m := NewModel(DefaultUIConfig())
	m.textinput.SetValue("hello")

	anyM, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlM})
	m = anyM.(Model)
	if !m.multiline {
		t.Fatalf("expected multiline true")
	}
	if m.textarea.Value() != "hello" {
		t.Fatalf("expected text to transfer, got %q", m.textarea.Value())
	}

	anyM, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlM})
	m = anyM.(Model)
	if m.multiline {
		t.Fatalf("expected multiline false")
	}
	if m.textinput.Value() != "hello" {
		t.Fatalf("expected text to transfer back, got %q", m.textinput.Value())
	}
}

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
