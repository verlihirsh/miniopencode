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
	"github.com/charmbracelet/lipgloss"
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

	serverHost string
	serverPort int
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
	m.viewport.SetContent(welcomeMessage())
	cmds := []tea.Cmd{m.textinput.Focus()}
	if m.chunkCh != nil {
		cmds = append(cmds, waitForChunk(m.chunkCh))
	}
	if m.errCh != nil {
		cmds = append(cmds, waitForError(m.errCh))
	}
	return tea.Batch(cmds...)
}

func welcomeMessage() string {
	return `Welcome to miniopencode!

Type your message and press Enter to send.
Press ? for help, q to quit.

Ready to chat...`
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
		if m.chunkCh != nil {
			return m, waitForChunk(m.chunkCh)
		}
	case error:
		m.transcript = append(m.transcript, Chunk{Kind: ChunkRaw, Text: msg.Error()})
		if m.errCh != nil {
			return m, waitForError(m.errCh)
		}
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
	var content string
	if m.multiline {
		content = m.textarea.View()
	} else {
		content = m.textinput.View()
	}
	status := m.renderStatus()
	return fmt.Sprintf("%s\n%s", status, renderWithBorder(content, inputBorderStyle, m.width, m.height-1))
}

func (m Model) viewOutputOnly() string {
	status := m.renderStatus()
	return fmt.Sprintf("%s\n%s", status, renderWithBorder(m.viewport.View(), outputBorderStyle, m.width, m.height-1))
}

func (m Model) viewFull() string {
	status := m.renderStatus()
	outputBox := renderWithBorder(m.viewport.View(), outputBorderStyle, m.width, m.height-m.inputHeight-2)
	inputBox := renderWithBorder(m.inputView(), inputBorderStyle, m.width, m.inputHeight)
	return fmt.Sprintf("%s\n%s\n%s", status, outputBox, inputBox)
}

func (m Model) renderStatus() string {
	mode := "full"
	switch m.mode {
	case ModeInput:
		mode = "input"
	case ModeOutput:
		mode = "output"
	}
	multilineIndicator := ""
	if m.multiline {
		multilineIndicator = " [multiline]"
	}

	left := titleStyle.Render(fmt.Sprintf("miniopencode"))
	middle := statusStyle.Render(fmt.Sprintf("session=%s | mode=%s%s", m.sessionID, mode, multilineIndicator))
	right := statusStyle.Render(fmt.Sprintf("%s:%d", m.serverHost, m.serverPort))

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(middle) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap/2), middle, strings.Repeat(" ", gap-gap/2), right)
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
		return m, nil
	case key.Matches(msg, m.keys.ModeOutput):
		m.mode = ModeOutput
		return m, nil
	case key.Matches(msg, m.keys.ModeFull):
		m.mode = ModeFull
		return m, nil
	case msg.Type == tea.KeyCtrlM:
		m.multiline = !m.multiline
		return m, nil
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
	default:
		var cmd tea.Cmd
		if m.mode == ModeOutput {
			m.viewport, cmd = m.viewport.Update(msg)
		} else if m.multiline {
			m.textarea, cmd = m.textarea.Update(msg)
		} else {
			m.textinput, cmd = m.textinput.Update(msg)
		}
		return m, cmd
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
