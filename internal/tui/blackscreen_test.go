package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialViewNotBlank(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()

	view := m.View()

	if strings.TrimSpace(view) == "" {
		t.Errorf("initial view should not be blank; got empty view")
	}
}

func TestInputDelegationToWidgets(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()

	m.textinput.Focus()

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	updatedM, _ := m.Update(keyMsg)
	model := updatedM.(Model)

	inputText := model.currentInputText()
	if !strings.Contains(inputText, "h") {
		t.Errorf("expected input to contain 'h', got %q", inputText)
	}
}

func TestWindowSizeInitializesViewport(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updatedM, _ := m.Update(sizeMsg)
	model := updatedM.(Model)

	if model.viewport.Width == 0 {
		t.Error("viewport width should be initialized after WindowSizeMsg")
	}
	if model.viewport.Height == 0 {
		t.Error("viewport height should be initialized after WindowSizeMsg")
	}
}

func TestInitReturnsChunkListener(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	chunkCh := make(chan Chunk, 1)
	errCh := make(chan error, 1)
	m.chunkCh = chunkCh
	m.errCh = errCh

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a command to listen for chunks, got nil")
	}
}

func TestStatusBarVisible(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 100
	m.height = 30
	m.sessionID = "test-session"
	m.serverHost = "localhost"
	m.serverPort = 4096
	m.applySizes()

	view := m.View()

	if !strings.Contains(view, "miniopencode") {
		t.Error("status bar should contain 'miniopencode' title")
	}
	if !strings.Contains(view, "test-session") {
		t.Error("status bar should contain session ID")
	}
	if !strings.Contains(view, "localhost:4096") {
		t.Error("status bar should contain server info")
	}
}

func TestBordersVisible(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.applySizes()

	view := m.View()

	if !strings.Contains(view, "╭") && !strings.Contains(view, "┌") {
		t.Error("view should contain border characters")
	}
}

func TestOutputModeScrollable(t *testing.T) {
	cfg := DefaultUIConfig()
	m := NewModel(cfg)

	m.width = 80
	m.height = 24
	m.mode = ModeOutput
	m.applySizes()

	longContent := ""
	for i := 0; i < 50; i++ {
		longContent += fmt.Sprintf("Line %d\n", i+1)
	}
	m.viewport.SetContent(longContent)

	initialOffset := m.viewport.YOffset

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedM, _ := m.Update(downMsg)
	model := updatedM.(Model)

	if model.viewport.YOffset == initialOffset && model.viewport.Height > 0 {
		t.Errorf("viewport should scroll in output mode when content exceeds height, offset stayed at %d", initialOffset)
	}
}
