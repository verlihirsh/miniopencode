package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

func renderMarkdown(width int, md string) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithAutoStyle(),
	)
	if err != nil {
		return md
	}
	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return strings.TrimSpace(out)
}
