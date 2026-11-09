// Package apptest runs a program with no tablet: AppLoad is replaced by a stand-in that remembers
// what it was asked, and events are delivered by hand.
package apptest

import (
	"testing"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/canvas"
	"github.com/akrck02/go-remarkable/pen"
	"github.com/akrck02/go-remarkable/qtfb"
)

// highestModeAppLoadTakes is the last refresh mode AppLoad v0.6.0 accepts; it closes the connection
// on any higher one.
const highestModeAppLoadTakes = qtfb.Interface

// Display takes the place of AppLoad's connection and remembers what it was asked, in order.
type Display struct {
	Calls     []string
	Modes     []qtfb.RefreshMode
	Refreshes []canvas.Rectangle
}

// SetRefreshMode remembers a mode, and refuses one AppLoad would not take.
func (target *Display) SetRefreshMode(mode qtfb.RefreshMode) error {
	if mode > highestModeAppLoadTakes {
		return qtfb.UpdateFailed
	}
	target.Calls = append(target.Calls, "mode")
	target.Modes = append(target.Modes, mode)
	return nil
}

// Refresh remembers an area.
func (target *Display) Refresh(area canvas.Rectangle) error {
	target.Calls = append(target.Calls, "refresh")
	target.Refreshes = append(target.Refreshes, area)
	return nil
}

// Forget empties the memory of what was asked.
func (target *Display) Forget() {
	target.Calls, target.Modes, target.Refreshes = nil, nil, nil
}

// Start runs a program up to its first frame, and forgets what that frame asked for.
func Start(test testing.TB, newProgram func(device app.Device) app.Program) (*app.Application, *Display) {
	test.Helper()
	return StartWithPen(test, pen.Plain, newProgram)
}

// StartWithPen is Start with a pen that answers as told.
func StartWithPen(test testing.TB, details func() pen.Sample, newProgram func(device app.Device) app.Program) (*app.Application, *Display) {
	test.Helper()
	target := &Display{}
	memory := make([]byte, app.PanelWidthInPixels*app.PanelHeightInPixels*canvas.BytesPerScreenPixel)
	application, err := app.New(target, memory, details, newProgram)
	if err != nil {
		test.Fatal(err)
	}
	if err := application.Update(); err != nil {
		test.Fatal(err)
	}
	target.Forget()
	return application, target
}

// Send delivers events one at a time, each answered before the next, as they reach the tablet.
func Send(test testing.TB, application *app.Application, events ...qtfb.Event) {
	test.Helper()
	for _, event := range events {
		application.Apply(event)
		Answer(test, application)
	}
}

// SendTogether delivers events in one batch, as the program's loop takes in whatever is queued.
func SendTogether(test testing.TB, application *app.Application, events ...qtfb.Event) {
	test.Helper()
	for _, event := range events {
		application.Apply(event)
	}
	Answer(test, application)
}

// Answer lets the program show what the events changed.
func Answer(test testing.TB, application *app.Application) {
	test.Helper()
	if err := application.Answer(); err != nil {
		test.Fatal(err)
	}
}

// Update builds the interface again, as a timer would.
func Update(test testing.TB, application *app.Application) {
	test.Helper()
	if err := application.Update(); err != nil {
		test.Fatal(err)
	}
}

// MiddleOf is the centre of a control, in the canvas pixels events arrive in.
func MiddleOf(test testing.TB, application *app.Application, path ...string) (x, y int32) {
	test.Helper()
	area, found := application.Context().Area(path...)
	if !found {
		test.Fatalf("no control at %v", path)
	}
	return int32((area.X + area.Width/2) * app.PixelsPerLayoutPixel), int32((area.Y + area.Height/2) * app.PixelsPerLayoutPixel)
}

// Tap presses and lifts a finger in the middle of a control.
func Tap(test testing.TB, application *app.Application, path ...string) {
	test.Helper()
	x, y := MiddleOf(test, application, path...)
	Send(test, application, qtfb.Event{Kind: qtfb.TouchPress, X: x, Y: y}, qtfb.Event{Kind: qtfb.TouchRelease, X: x, Y: y})
}
