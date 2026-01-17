package tui

func (m *Model) checkTTY() {
	if m.width == 0 && m.height == 0 {
		m.placeholder = "[no TTY detected]"
	}
}
