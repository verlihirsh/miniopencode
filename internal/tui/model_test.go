package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModeSwitching(t *testing.T) {
	m := NewModel(DefaultUIConfig())
	mAny, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = mAny.(Model)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = mAny.(Model)
	if m.mode != ModeOutput {
		t.Fatalf("expected output mode, got %v", m.mode)
	}
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = mAny.(Model)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = mAny.(Model)
	if m.mode != ModeInput {
		t.Fatalf("expected input mode, got %v", m.mode)
	}
}

func TestToggleMultiline(t *testing.T) {
	m := NewModel(DefaultUIConfig())
	anyM, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlM})
	m = anyM.(Model)
	if !m.multiline {
		t.Fatalf("expected multiline true")
	}
	anyM, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlM})
	m = anyM.(Model)
	if m.multiline {
		t.Fatalf("expected multiline false")
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
