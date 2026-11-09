package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/akrck02/go-remarkable/qtfb"
)

func TestAnEventReachesTheCountAndTheLog(test *testing.T) {
	var written bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&written, nil)))
	defer slog.SetDefault(previous)
	page := newMonitor()
	page.record(qtfb.Event{Kind: qtfb.PenPress, X: 7, Y: 9})
	if page.shownCounts[eventName(qtfb.PenPress)] != 1 {
		test.Fatalf("counts %v", page.shownCounts)
	}
	if !strings.Contains(written.String(), "pen press") {
		test.Fatalf("log: %q", written.String())
	}
}

func TestAMovementIsCountedButNotLoggedOrShownAtOnce(test *testing.T) {
	var written bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&written, nil)))
	defer slog.SetDefault(previous)
	page := newMonitor()
	page.record(qtfb.Event{Kind: qtfb.PenMove})
	if written.Len() != 0 {
		test.Fatalf("log: %q", written.String())
	}
	if len(page.shownCounts) != 0 {
		test.Fatalf("shown %v", page.shownCounts)
	}
	page.record(qtfb.Event{Kind: qtfb.PenPress})
	if page.shownCounts[eventName(qtfb.PenMove)] != 1 {
		test.Fatalf("the movement was lost: %v", page.shownCounts)
	}
}

func TestOnlyTheLastEventsAreKept(test *testing.T) {
	page := newMonitor()
	for range monitorEventsShown + 5 {
		page.record(qtfb.Event{Kind: qtfb.PenPress})
	}
	if len(page.recent) != monitorEventsShown {
		test.Fatalf("kept %d", len(page.recent))
	}
}

func TestTheProgramRecordsEveryEventItReceives(test *testing.T) {
	shown, _ := newShownApplication(test)
	shown.Apply(qtfb.Event{Kind: qtfb.TouchPress, X: 5, Y: 5})
	if shown.demo.monitor.shownCounts[eventName(qtfb.TouchPress)] != 1 {
		test.Fatalf("counts %v", shown.demo.monitor.shownCounts)
	}
}
