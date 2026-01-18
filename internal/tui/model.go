package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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
	InputHeight    int
	ShowThinking   bool
	ShowTools      bool
	Wrap           bool
	MaxOutputLines int
}

func DefaultUIConfig() UIConfig {
	return UIConfig{
		Mode:           "full",
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
	spinner     spinner.Model
	placeholder string

	mode          UIMode
	width         int
	height        int
	inputHeight   int
	showThinking  bool
	showTools     bool
	pendingResize bool
	sending       bool
	followOutput  bool

	streamer       *Streamer
	sessionID      string
	promptCfg      PromptConfig
	chunkCh        <-chan Chunk
	errCh          <-chan error
	maxOutputLines int

	transcript Transcript

	lastPartID    string
	lastMessageID string

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

	m.transcript.AppendAssistantChunk(c.MessageID, c.PartID, c.Kind, c.Text)
	m.viewport.SetContent(m.transcript.Render(m.showThinking, m.showTools, m.spinner.View(), m.sending))
	if m.followOutput {
		m.viewport.GotoBottom()
	}
	return m
}

func NewModel(cfg UIConfig) Model {
	km := DefaultKeyMap()
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Focus()

	h := help.New()
	h.ShowAll = false

	vp := viewport.New(0, 0)
	vp.SetContent(welcomeMessage())

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := Model{
		keys:         km,
		help:         h,
		viewport:     vp,
		textinput:    ti,
		spinner:      sp,
		showThinking: cfg.ShowThinking,
		showTools:    cfg.ShowTools,
		inputHeight:  cfg.InputHeight,
		followOutput: true,
	}

	m.textinput.Focus()

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
	cmds := []tea.Cmd{m.spinner.Tick}

	if m.mode != ModeOutput {
		cmds = append(cmds, textinput.Blink)
	}

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
	var cmd tea.Cmd
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
	case sendComplete:
		m = m.handleSendComplete()
	case error:
		m = m.clearInput()
		m.sending = false
		if msg != nil {
			log.Printf("tui: error displayed: %v", msg)
			m.transcript.AddAssistantSystemLine("[Error] " + msg.Error())
			m.viewport.SetContent(m.transcript.Render(m.showThinking, m.showTools, m.spinner.View(), m.sending))

		}
		if m.errCh != nil {
			return m, tea.Batch(waitForChunk(m.chunkCh), waitForError(m.errCh))
		}
		return m, waitForChunk(m.chunkCh)
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		if m.sending {
			m.viewport.SetContent(m.transcript.Render(m.showThinking, m.showTools, m.spinner.View(), m.sending))
			if m.followOutput {
				m.viewport.GotoBottom()
			}
		}
		return m, cmd
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
	m.checkTTY()
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
		m.viewport.Width = m.width - 4
		m.viewport.Height = m.height - 4
		return
	}
	if m.mode == ModeInput {
		m.textinput.Width = m.width - 2
		return
	}
	m.viewport.Width = m.width - 4
	m.viewport.Height = m.height - m.inputHeight - 4
	m.textinput.Width = m.width - 2
}

func (m Model) viewInputOnly() string {
	content := m.textinput.View()
	status := m.renderStatus()
	return fmt.Sprintf("%s\n%s", status, renderWithBorder(content, inputBorderStyle, m.width, m.height-1))
}

func (m Model) viewOutputOnly() string {
	status := m.renderStatus()
	outputBox := renderWithBorder(m.viewport.View(), outputBorderStyle, m.width, m.height-1)
	return fmt.Sprintf("%s\n%s", status, outputBox)
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
	sendingIndicator := ""
	if m.sending {
		sendingIndicator = fmt.Sprintf(" %s thinking...", m.spinner.View())
	}

	left := titleStyle.Render(fmt.Sprintf("miniopencode"))
	middle := statusStyle.Render(fmt.Sprintf("session=%s | mode=%s%s%s", m.sessionID, mode, multilineIndicator, sendingIndicator))
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
	return m.textinput.View()
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.SendSingle):
		return m.sendInput()
	case msg.Type == tea.KeyCtrlW:
		m.pendingResize = true
		return m, nil
	case m.pendingResize && resizeIncrease(msg):
		m.inputHeight++
	case m.pendingResize && resizeDecrease(msg):
		m.inputHeight--
	case isScrollKey(msg):
		if msg.Type == tea.KeyEnd {
			m.followOutput = true
		} else {
			m.followOutput = false
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		if m.viewport.AtBottom() {
			m.followOutput = true
		}
		return m, cmd
	default:
		var cmd tea.Cmd
		if m.mode == ModeOutput {
			m.viewport, cmd = m.viewport.Update(msg)
		} else {
			m.textinput, cmd = m.textinput.Update(msg)
		}
		return m, cmd
	}
	m.applySizes()
	m.pendingResize = false
	return m, nil
}

func isScrollKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown, tea.KeyHome, tea.KeyEnd:
		return true
	case tea.KeyCtrlU, tea.KeyCtrlD:
		return true
	}
	return false
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
