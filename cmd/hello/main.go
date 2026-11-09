// Command hello is the smallest program for the tablet: some words on paper.
package main

import (
	"log/slog"
	"os"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/app"
)

const marginInLayoutPixels = 24

func newProgram(app.Device) app.Program {
	return app.Program{Draw: draw}
}

func draw(context *ui.Context) {
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(marginInLayoutPixels)}, func() {
		context.Heading("Hello, paper")
	})
}

func main() {
	if err := app.Run(newProgram); err != nil {
		slog.Error("stopped", "error", err)
		os.Exit(1)
	}
}
