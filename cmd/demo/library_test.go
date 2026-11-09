package main

import (
	"strings"
	"testing"
)

func TestLinesNeverPassTheWidthAndKeepEveryWord(test *testing.T) {
	text := strings.Repeat("a tide comes in and goes out ", 9)
	lines := wrapText(text, 12)
	for _, line := range lines {
		if len(line) > 12 {
			test.Errorf("line too long: %q", line)
		}
	}
	if strings.Join(lines, " ") != strings.TrimSpace(text) {
		test.Error("words were lost or changed")
	}
}

func TestPagesCoverEveryLineOnce(test *testing.T) {
	lines := make([]string, linesPerPage*2+1)
	if pages := pageCount(lines); pages != 3 {
		test.Fatalf("got %d pages", pages)
	}
	if last := pageLines(lines, 2); len(last) != 1 {
		test.Fatalf("last page has %d lines", len(last))
	}
}
