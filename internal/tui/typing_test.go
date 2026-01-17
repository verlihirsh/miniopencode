package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTypingInFullMode(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.mode = ModeFull
	m.applySizes()

	m.textinput.Focus()

	keys := []rune{'h', 'e', 'l', 'l', 'o'}
	for _, r := range keys {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedM, _ := m.Update(keyMsg)
		m = updatedM.(Model)
	}

	inputText := m.currentInputText()
	if !strings.Contains(inputText, "hello") {
		t.Errorf("expected input to contain 'hello', got %q", inputText)
	}

	view := m.View()
	if !strings.Contains(view, "hello") {
		t.Errorf("expected view to show 'hello', got:\n%s", view)
	}
}

func TestTypingInInputMode(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.mode = ModeInput
	m.applySizes()

	m.textinput.Focus()

	keys := []rune{'t', 'e', 's', 't'}
	for _, r := range keys {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedM, _ := m.Update(keyMsg)
		m = updatedM.(Model)
	}

	inputText := m.currentInputText()
	if !strings.Contains(inputText, "test") {
		t.Errorf("expected input to contain 'test', got %q", inputText)
	}

	view := m.View()
	if !strings.Contains(view, "test") {
		t.Errorf("expected view to show 'test', got:\n%s", view)
	}
}

func TestTypingInMultilineMode(t *testing.T) {
	cfg := DefaultUIConfig()
	cfg.Multiline = true
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.mode = ModeFull
	m.applySizes()

	m.textarea.Focus()

	keys := []rune{'m', 'u', 'l', 't', 'i'}
	for _, r := range keys {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedM, _ := m.Update(keyMsg)
		m = updatedM.(Model)
	}

	inputText := m.currentInputText()
	if !strings.Contains(inputText, "multi") {
		t.Errorf("expected input to contain 'multi', got %q", inputText)
	}

	view := m.View()
	if !strings.Contains(view, "multi") {
		t.Errorf("expected view to show 'multi', got:\n%s", view)
	}
}

func TestTypingAfterInit(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()

	m.textinput.Focus()

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedM, _ := m.Update(keyMsg)
	m = updatedM.(Model)

	inputText := m.currentInputText()
	if !strings.Contains(inputText, "a") {
		t.Errorf("expected input to contain 'a' with focused input, got %q", inputText)
	}
}

func TestBackspaceWorks(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()

	m.textinput.Focus()

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedM, _ := m.Update(keyMsg)
	m = updatedM.(Model)

	if !strings.Contains(m.currentInputText(), "x") {
		t.Error("expected 'x' to be typed")
	}

	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedM, _ = m.Update(backspaceMsg)
	m = updatedM.(Model)

	inputText := m.currentInputText()
	if strings.Contains(inputText, "x") {
		t.Errorf("expected 'x' to be deleted by backspace, got %q", inputText)
	}
}

func TestSpaceWorks(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()

	m.textinput.Focus()

	keys := []rune{'h', 'i', ' ', 't', 'h', 'e', 'r', 'e'}
	for _, r := range keys {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedM, _ := m.Update(keyMsg)
		m = updatedM.(Model)
	}

	inputText := m.currentInputText()
	if !strings.Contains(inputText, "hi there") {
		t.Errorf("expected input to contain 'hi there', got %q", inputText)
	}
}
