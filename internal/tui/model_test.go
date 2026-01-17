package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModeSwitching(t *testing.T) {
	m := NewModel(DefaultUIConfig())
	// Switch to output mode with alt+o
	mAny, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}, Alt: true})
	m = mAny.(Model)
	if m.mode != ModeOutput {
		t.Fatalf("expected output mode, got %v", m.mode)
	}
	// Switch to input mode with alt+i
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}, Alt: true})
	m = mAny.(Model)
	if m.mode != ModeInput {
		t.Fatalf("expected input mode, got %v", m.mode)
	}
	// Switch to full mode with alt+f
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}, Alt: true})
	m = mAny.(Model)
	if m.mode != ModeFull {
		t.Fatalf("expected full mode, got %v", m.mode)
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
