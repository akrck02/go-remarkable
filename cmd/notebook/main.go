// Command notebook is a page to write on with the pen, and a button to start again.
package main

import (
	"log/slog"
	"os"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/ink"
	"github.com/akrck02/go-remarkable/paper"
)

const (
	marginInLayoutPixels = 24
	gapInLayoutPixels    = 12
	clearButtonName      = "Start again"
)

// notebook is one page of paper.
type notebook struct {
	page *ink.Surface
}

func (shown *notebook) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(marginInLayoutPixels), Gap: gapInLayoutPixels}, func() {
		if paper.Button(context, clearButtonName) {
			shown.page.Clear()
		}
		shown.page.DrawPaper(context)
	})
}

func newProgram(device app.Device) app.Program {
	shown := &notebook{page: device.NewInk()}
	return app.Program{Draw: shown.draw, Ink: shown.page}
}

func main() {
	if err := app.Run(newProgram); err != nil {
		slog.Error("stopped", "error", err)
		os.Exit(1)
	}
}
