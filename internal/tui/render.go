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
	prefix := ""
	switch c.Kind {
	case ChunkThinking:
		prefix = "[Thinking]\n"
	case ChunkTool:
		prefix = "[Tool]\n"
	}
	return prefix + renderMarkdown(width, c.Text)
}
