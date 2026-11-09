package main

import (
	"testing"
)

func TestPagesCoverEveryLineOnce(test *testing.T) {
	lines := make([]string, linesPerPage*2+1)
	if pages := pageCount(lines); pages != 3 {
		test.Fatalf("got %d pages", pages)
	}
	if last := pageLines(lines, 2); len(last) != 1 {
		test.Fatalf("last page has %d lines", len(last))
	}
}
