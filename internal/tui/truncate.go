package tui

func truncateLines(lines []string, max int) []string {
	if max <= 0 {
		return lines
	}
	if len(lines) <= max {
		return lines
	}
	return lines[len(lines)-max:]
}
