package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTypingInInputMode(t *testing.T) {
	cfg := DefaultUIConfig()
	cfg.Mode = "input"
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()
	m.textinput.Focus()

	mAny, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = mAny.(Model)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = mAny.(Model)

	if m.textinput.Value() != "hi" {
		t.Fatalf("expected 'hi', got %q", m.textinput.Value())
	}
}

func TestTypingInFullMode(t *testing.T) {
	cfg := DefaultUIConfig()
	cfg.Mode = "full"
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()
	m.textinput.Focus()

	mAny, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = mAny.(Model)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = mAny.(Model)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = mAny.(Model)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = mAny.(Model)

	if m.textinput.Value() != "test" {
		t.Fatalf("expected 'test', got %q", m.textinput.Value())
	}
}

func TestNoTypingInOutputMode(t *testing.T) {
	cfg := DefaultUIConfig()
	cfg.Mode = "output"
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()

	mAny, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = mAny.(Model)

	if m.textinput.Value() != "" {
		t.Fatalf("expected empty in output mode, got %q", m.textinput.Value())
	}
}

func TestMultilineTypingInInputMode(t *testing.T) {
	cfg := DefaultUIConfig()
	cfg.Mode = "input"
	cfg.Multiline = true
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()
	m.textarea.Focus()

	mAny, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = mAny.(Model)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = mAny.(Model)

	if m.textarea.Value() != "mu" {
		t.Fatalf("expected 'mu', got %q", m.textarea.Value())
	}
}

func TestMultilineTypingInFullMode(t *testing.T) {
	cfg := DefaultUIConfig()
	cfg.Mode = "full"
	cfg.Multiline = true
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()
	m.textarea.Focus()

	mAny, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = mAny.(Model)
	mAny, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = mAny.(Model)

	if m.textarea.Value() != "ok" {
		t.Fatalf("expected 'ok', got %q", m.textarea.Value())
	}
}
