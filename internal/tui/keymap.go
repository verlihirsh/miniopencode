package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit            key.Binding
	Help            key.Binding
	ModeInput       key.Binding
	ModeOutput      key.Binding
	ModeFull        key.Binding
	ToggleMultiline key.Binding
	SendSingle      key.Binding
	SendMultiline   key.Binding
	ResizeUp        key.Binding
	ResizeDown      key.Binding
	ToggleThinking  key.Binding
	ToggleTools     key.Binding
	// scrolling
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	HalfUp   key.Binding
	HalfDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:            key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Help:            key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		ModeInput:       key.NewBinding(key.WithKeys("g", "i"), key.WithHelp("g i", "input mode")),
		ModeOutput:      key.NewBinding(key.WithKeys("g", "o"), key.WithHelp("g o", "output mode")),
		ModeFull:        key.NewBinding(key.WithKeys("g", "f"), key.WithHelp("g f", "full mode")),
		ToggleMultiline: key.NewBinding(key.WithKeys("ctrl+m"), key.WithHelp("ctrl+m", "toggle multiline")),
		SendSingle:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "send")),
		SendMultiline:   key.NewBinding(key.WithKeys("ctrl+enter", "ctrl+j"), key.WithHelp("ctrl+enter", "send")),
		ResizeUp:        key.NewBinding(key.WithKeys("ctrl+w", "+"), key.WithHelp("ctrl+w +", "input taller")),
		ResizeDown:      key.NewBinding(key.WithKeys("ctrl+w", "-"), key.WithHelp("ctrl+w -", "input shorter")),
		ToggleThinking:  key.NewBinding(key.WithKeys("t", "t"), key.WithHelp("t t", "toggle thinking")),
		ToggleTools:     key.NewBinding(key.WithKeys("t", "o"), key.WithHelp("t o", "toggle tools")),
		Up:              key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
		Down:            key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
		PageUp:          key.NewBinding(key.WithKeys("b", "pgup"), key.WithHelp("b", "page up")),
		PageDown:        key.NewBinding(key.WithKeys("f", "pgdown", " "), key.WithHelp("f", "page down")),
		HalfUp:          key.NewBinding(key.WithKeys("u", "ctrl+u"), key.WithHelp("u", "half up")),
		HalfDown:        key.NewBinding(key.WithKeys("d", "ctrl+d"), key.WithHelp("d", "half down")),
		Top:             key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g", "top")),
		Bottom:          key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "bottom")),
	}
}
