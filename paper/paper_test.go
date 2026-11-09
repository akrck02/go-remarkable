package paper

import (
	"strings"
	"testing"
)

func TestLinesNeverPassTheWidthAndKeepEveryWord(test *testing.T) {
	text := strings.Repeat("a tide comes in and goes out ", 9)
	lines := WrapText(text, 12)
	for _, line := range lines {
		if len(line) > 12 {
			test.Errorf("line too long: %q", line)
		}
	}
	if strings.Join(lines, " ") != strings.TrimSpace(text) {
		test.Error("words were lost or changed")
	}
}
