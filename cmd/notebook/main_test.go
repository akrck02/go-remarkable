package main

import (
	"testing"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/apptest"
	"github.com/akrck02/go-remarkable/ink"
	"github.com/akrck02/go-remarkable/qtfb"
)

func TestTheButtonStartsThePageAgain(test *testing.T) {
	shown := &notebook{}
	application, _ := apptest.Start(test, func(device app.Device) app.Program {
		shown.page = device.NewInk()
		return app.Program{Draw: shown.draw, Ink: shown.page}
	})
	x, y := apptest.MiddleOf(test, application, ink.PaperName)
	apptest.Send(test, application,
		qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y},
		qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y},
		qtfb.Event{Kind: qtfb.PenRelease, X: x + 40, Y: y})
	if shown.page.StrokeCount() != 1 {
		test.Fatalf("%d strokes", shown.page.StrokeCount())
	}
	apptest.Tap(test, application, clearButtonName)
	if shown.page.StrokeCount() != 0 {
		test.Fatalf("%d strokes after starting again", shown.page.StrokeCount())
	}
}
