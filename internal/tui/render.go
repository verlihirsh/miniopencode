package tui

import (
	"strings"
)

func renderTranscript(chunks []Chunk, width int) string {
	var b strings.Builder
	for i, c := range chunks {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(renderChunk(c, width))
	}
	return b.String()
}

func renderChunk(c Chunk, width int) string {
	md := c.Text
	switch c.Kind {
	case ChunkThinking:
		md = "### Thinking\n" + md
	case ChunkTool:
		md = "### Tool\n" + md
	case ChunkAnswer:
		md = "### Answer\n" + md
	default:
	}
	return renderMarkdown(width, md)
}
