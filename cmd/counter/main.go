// Command counter is a button that counts: each press costs one refresh of the panel, and between
// presses the program sleeps.
package main

import (
	"fmt"
	"log/slog"
	"os"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/paper"
)

const (
	marginInLayoutPixels = 24
	gapInLayoutPixels    = 16
	addButtonName        = "Add one"
)

// counter is the program's whole state.
type counter struct {
	count int
}

func (shown *counter) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(marginInLayoutPixels), Gap: gapInLayoutPixels}, func() {
		context.Heading(fmt.Sprintf("Pressed %d times", shown.count))
		if paper.Button(context, addButtonName) {
			shown.count++
		}
	})
}

func newProgram(app.Device) app.Program {
	return app.Program{Draw: (&counter{}).draw}
}

func main() {
	if err := app.Run(newProgram); err != nil {
		slog.Error("stopped", "error", err)
		os.Exit(1)
	}
}
