package tui

import "testing"

func TestTruncateLines(t *testing.T) {
	lines := []string{"1", "2", "3", "4"}
	got := truncateLines(lines, 2)
	if len(got) != 2 || got[0] != "3" || got[1] != "4" {
		t.Fatalf("unexpected truncation: %v", got)
	}
	same := truncateLines(lines, 10)
	if len(same) != 4 {
		t.Fatalf("should keep all lines")
	}
}
