package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type UIMode int

const (
	ModeFull UIMode = iota
	ModeOutput
	ModeInput
)

const (
	minInputHeight  = 3
	minOutputHeight = 5
)

type UIConfig struct {
	Mode           string
	Multiline      bool
	InputHeight    int
	ShowThinking   bool
	ShowTools      bool
	Wrap           bool
	MaxOutputLines int
}

func DefaultUIConfig() UIConfig {
	return UIConfig{
		Mode:           "full",
		Multiline:      false,
		InputHeight:    6,
		ShowThinking:   true,
		ShowTools:      true,
		Wrap:           true,
		MaxOutputLines: 4000,
	}
}

type Model struct {
	keys        KeyMap
	help        help.Model
	viewport    viewport.Model
	textinput   textinput.Model
	textarea    textarea.Model
	placeholder string

	multiline     bool
	mode          UIMode
	width         int
	height        int
	inputHeight   int
	showThinking  bool
	showTools     bool
	pendingResize bool

	streamer       *Streamer
	sessionID      string
	promptCfg      PromptConfig
	chunkCh        <-chan Chunk
	errCh          <-chan error
	maxOutputLines int
	transcript     []Chunk
}

func (m Model) appendChunk(c Chunk) Model {
	if c.Kind == ChunkThinking && !m.showThinking {
		return m
	}
	if c.Kind == ChunkTool && !m.showTools {
		return m
	}
	m.transcript = append(m.transcript, c)
	if m.maxOutputLines > 0 {
		var lines []string
		for _, chunk := range m.transcript {
			lines = append(lines, strings.Split(chunk.Text, "\n")...)
		}
		lines = truncateLines(lines, m.maxOutputLines)
		m.transcript = []Chunk{{Kind: ChunkAnswer, Text: strings.Join(lines, "\n")}}

	}
	m.viewport.SetContent(renderTranscript(m.transcript, m.width))
	return m
}

func NewModel(cfg UIConfig) Model {
	km := DefaultKeyMap()
	ti := textinput.New()
	ti.Prompt = "> "

	ta := textarea.New()
	ta.SetHeight(cfg.InputHeight)

	h := help.New()
	h.ShowAll = false

	m := Model{
		keys:         km,
		help:         h,
		textinput:    ti,
		textarea:     ta,
		multiline:    cfg.Multiline,
		showThinking: cfg.ShowThinking,
		showTools:    cfg.ShowTools,
		inputHeight:  cfg.InputHeight,
	}

	switch cfg.Mode {
	case "input":
		m.mode = ModeInput
	case "output":
		m.mode = ModeOutput
	default:
		m.mode = ModeFull
	}

	return m
}

func (m Model) Init() tea.Cmd {
	m.checkTTY()
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.applySizes()
	case tea.KeyMsg:
		return m.handleKey(msg)
	case Chunk:
		m = m.appendChunk(msg)
	case error:
		m.transcript = append(m.transcript, Chunk{Kind: ChunkRaw, Text: msg.Error()})
	}
	return m, nil
}

func (m Model) View() string {
	switch m.mode {
	case ModeInput:
		return m.viewInputOnly()
	case ModeOutput:
		return m.viewOutputOnly()
	default:
		return m.viewFull()
	}
}

func (m *Model) applySizes() {
	if m.width == 0 || m.height == 0 {
		return
	}
	if m.inputHeight < minInputHeight {
		m.inputHeight = minInputHeight
	}
	if m.inputHeight > m.height-minOutputHeight {
		m.inputHeight = m.height - minOutputHeight
	}
	if m.inputHeight < minInputHeight {
		m.inputHeight = minInputHeight
	}
	if m.mode == ModeOutput {
		m.viewport.Width = m.width
		m.viewport.Height = m.height - 2
		return
	}
	if m.mode == ModeInput {
		m.textinput.Width = m.width - 2
		m.textarea.SetWidth(m.width - 2)
		m.textarea.SetHeight(m.height - 2)
		return
	}
	// full
	m.viewport.Width = m.width
	m.viewport.Height = m.height - m.inputHeight - 1
	m.textinput.Width = m.width - 2
	m.textarea.SetWidth(m.width - 2)
	m.textarea.SetHeight(m.inputHeight)
}

func (m Model) viewInputOnly() string {
	if m.multiline {
		return m.textarea.View()
	}
	return m.textinput.View()
}

func (m Model) viewOutputOnly() string {
	return m.viewport.View()
}

func (m Model) viewFull() string {
	return fmt.Sprintf("%s\n%s", m.viewport.View(), m.inputView())
}

func (m Model) inputView() string {
	if m.placeholder != "" {
		return m.placeholder
	}
	if m.multiline {
		return m.textarea.View()
	}
	return m.textinput.View()
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.ModeInput):
		m.mode = ModeInput
	case key.Matches(msg, m.keys.ModeOutput):
		m.mode = ModeOutput
	case key.Matches(msg, m.keys.ModeFull):
		m.mode = ModeFull
	case msg.Type == tea.KeyCtrlM || key.Matches(msg, m.keys.ToggleMultiline):
		m.multiline = !m.multiline
	case key.Matches(msg, m.keys.SendSingle) && !m.multiline:
		return m, m.sendInput()
	case key.Matches(msg, m.keys.SendMultiline) && m.multiline:
		return m, m.sendInput()
	case msg.Type == tea.KeyCtrlW:
		m.pendingResize = true
		return m, nil
	case m.pendingResize && resizeIncrease(msg):
		m.inputHeight++
	case m.pendingResize && resizeDecrease(msg):
		m.inputHeight--
	}
	m.applySizes()
	m.pendingResize = false
	return m, nil
}

func resizeIncrease(msg tea.KeyMsg) bool {
	if msg.String() == "+" || msg.String() == "=" {
		return true
	}
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		return msg.Runes[0] == '+' || msg.Runes[0] == '='
	}
	return false
}

func resizeDecrease(msg tea.KeyMsg) bool {
	if msg.String() == "-" {
		return true
	}
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		return msg.Runes[0] == '-'
	}
	return false
}
