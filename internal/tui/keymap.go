package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit       key.Binding
	Help       key.Binding
	SendSingle key.Binding
	ResizeUp   key.Binding
	ResizeDown key.Binding
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	HalfUp     key.Binding
	HalfDown   key.Binding
	Top        key.Binding
	Bottom     key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:       key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Help:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		SendSingle: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "send")),
		ResizeUp:   key.NewBinding(key.WithKeys("ctrl+w", "+"), key.WithHelp("ctrl+w +", "input taller")),
		ResizeDown: key.NewBinding(key.WithKeys("ctrl+w", "-"), key.WithHelp("ctrl+w -", "input shorter")),
		Up:         key.NewBinding(key.WithKeys("up"), key.WithHelp("up", "up")),
		Down:       key.NewBinding(key.WithKeys("down"), key.WithHelp("down", "down")),
		PageUp:     key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown:   key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdown", "page down")),
		HalfUp:     key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "half up")),
		HalfDown:   key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "half down")),
		Top:        key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "top")),
		Bottom:     key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "bottom")),
	}
}
