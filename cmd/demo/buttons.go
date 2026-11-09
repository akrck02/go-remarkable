package main

import (
	"fmt"
	"time"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/paper"
)

const (
	tapButtonName   = "Tap, tap twice or hold"
	switchName      = "A switch"
	doubleTapWindow = 400 * time.Millisecond
)

// buttonsPage is the Buttons screen: each way of pressing something, counted.
type buttonsPage struct {
	taps, doubleTaps, longPresses int
	switched                      bool
	clock                         func() time.Time
	lastTap                       time.Time
}

// newButtonsPage makes the page with every count at zero.
func newButtonsPage() *buttonsPage {
	return &buttonsPage{clock: time.Now}
}

// noteTap counts a tap, and a double tap when it comes soon after the one before. Go UI has no
// double tap of its own. The tap that completes a double does not start another.
func (page *buttonsPage) noteTap() {
	now := page.clock()
	page.taps++
	if !page.lastTap.IsZero() && now.Sub(page.lastTap) < doubleTapWindow {
		page.doubleTaps++
		page.lastTap = time.Time{}
		return
	}
	page.lastTap = now
}

// draw builds the screen.
func (page *buttonsPage) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: screenMarginInLayoutPixels}, func() {
		context.Heading("Buttons")
		if paper.Button(context, tapButtonName) {
			page.noteTap()
		}
		if context.LongPressed() {
			page.longPresses++
		}
		paper.Switch(context, switchName, &page.switched)
		context.Label(fmt.Sprintf("Taps: %d", page.taps))
		context.Label(fmt.Sprintf("Double taps: %d", page.doubleTaps))
		context.Label(fmt.Sprintf("Long presses: %d", page.longPresses))
	})
}
