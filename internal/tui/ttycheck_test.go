package tui

import "testing"

func TestPlaceholderWhenNoTTY(t *testing.T) {
	m := NewModel(DefaultUIConfig())
	m.width = 0
	m.height = 0
	m.checkTTY()
	if m.placeholder == "" {
		t.Fatalf("expected placeholder when no TTY")
	}
	view := m.View()
	if m.mode != ModeFull {
		m.mode = ModeFull
	}

	if view == "" {
		t.Fatalf("expected non-empty view with placeholder")
	}
}
