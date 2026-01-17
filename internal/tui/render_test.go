package tui

import "testing"

func TestRenderMarkdownWraps(t *testing.T) {
	md := "# Title\n\nHello world"
	out := renderMarkdown(20, md)
	if len(out) == 0 {
		t.Fatalf("expected rendered output")
	}
}
