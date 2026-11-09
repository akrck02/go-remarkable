package main

import (
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/ink"
	"github.com/akrck02/go-remarkable/paper"
)

const clearButtonName = "Clear"

// inkPage is the Ink screen: a toolbar over the paper.
type inkPage struct {
	surface *ink.Surface
}

func newInkPage(surface *ink.Surface) *inkPage {
	return &inkPage{surface: surface}
}

// stop ends the stroke in progress, if there is one.
func (page *inkPage) stop() { page.surface.Stop() }

// draw builds the screen.
func (page *inkPage) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: mediumGapInLayoutPixels}, func() {
		context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), Gap: largeGapInLayoutPixels, AlignY: ui.Center}, func() {
			if paper.Button(context, clearButtonName) {
				page.surface.Clear()
			}
			context.Label(page.statusText())
		})
		page.surface.DrawPaper(context)
	})
}

// statusText says whether the pen's details are being read.
func (page *inkPage) statusText() string {
	if page.surface.ReadsPenDetails() {
		return "Pressure, tilt and eraser: on"
	}
	return "Plain lines: no details from the pen yet"
}
