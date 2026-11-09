package main

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/qtfb"
)

const monitorEventsShown = 16

// monitor is the Monitor screen, and the program's record of everything that arrives. After a hang
// its log says whether events were still reaching the program.
type monitor struct {
	clock       func() time.Time
	counts      map[string]int
	shownCounts map[string]int // the counts as of the last event that was not a movement
	recent      []string
	lastEvent   time.Time
}

// newMonitor makes a monitor that has seen nothing.
func newMonitor() *monitor {
	return &monitor{clock: time.Now, counts: map[string]int{}, shownCounts: map[string]int{}}
}

// eventName says what kind of event a code is.
func eventName(kind qtfb.EventKind) string {
	switch kind {
	case qtfb.TouchPress:
		return "touch press"
	case qtfb.TouchRelease:
		return "touch release"
	case qtfb.TouchMove:
		return "touch move"
	case qtfb.PenPress:
		return "pen press"
	case qtfb.PenRelease:
		return "pen release"
	case qtfb.PenMove:
		return "pen move"
	case qtfb.RotationChanged:
		return "rotation"
	}
	return fmt.Sprintf("other %#x", kind)
}

func isMovement(kind qtfb.EventKind) bool {
	return kind == qtfb.PenMove || kind == qtfb.TouchMove
}

// record counts an event. Movements are only counted: a pen sends hundreds a second, and showing
// each would refresh the panel for no reason.
func (page *monitor) record(event qtfb.Event) {
	name := eventName(event.Kind)
	page.counts[name]++
	if isMovement(event.Kind) {
		return
	}
	slog.Info("event", "kind", name, "device", event.Device, "x", event.X, "y", event.Y)
	page.lastEvent = page.clock()
	page.shownCounts = maps.Clone(page.counts)
	page.recent = append(page.recent, fmt.Sprintf("%s  %s  device %d  at %d, %d",
		page.lastEvent.Format(time.TimeOnly), name, event.Device, event.X, event.Y))
	page.recent = page.recent[max(len(page.recent)-monitorEventsShown, 0):]
}

// draw builds the screen. It shows a clock time, not "seconds ago", so an idle screen need not tick.
func (page *monitor) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: smallGapInLayoutPixels}, func() {
		context.Heading("Counts")
		for _, name := range slices.Sorted(maps.Keys(page.shownCounts)) {
			context.Label(fmt.Sprintf("%s: %d", name, page.shownCounts[name]))
		}
		context.Heading("Last events")
		for _, line := range page.recent {
			context.Label(line)
		}
	})
}
