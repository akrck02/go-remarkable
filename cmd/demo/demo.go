// Command demo is a test bench for the tablet: one screen for each way of using it, reading,
// buttons, ink, gestures, the panel's refresh modes and everything that arrives.
package main

import (
	"log/slog"
	"os"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/qtfb"
)

// pagesTabsName is the name of the row of tabs. The sizes are shared by every screen.
const (
	pagesTabsName = "Screens"

	lineGapInLayoutPixels      = 4
	smallGapInLayoutPixels     = 8
	mediumGapInLayoutPixels    = 12
	largeGapInLayoutPixels     = 16
	screenMarginInLayoutPixels = 24
	lineWidthInLayoutPixels    = 1
)

// The pages, in the order of their tabs.
const (
	pageReading = iota
	pageButtons
	pageInk
	pageGestures
	pagePanelLab
	pageMonitor
)

// pageNames are the tabs, in the order of the page numbers above.
var pageNames = []string{"Reading", "Buttons", "Ink", "Gestures", "Screen", "Monitor"}

// pageRefreshModes is how each page's frames reach the panel. Nearly all share one mode, because
// AppLoad stops reading from a program for a whole second after each change of mode.
var pageRefreshModes = []qtfb.RefreshMode{
	qtfb.Content, qtfb.Content, qtfb.Content, qtfb.Fast, qtfb.Content, qtfb.Content,
}

// demo is the program's interface: a row of tabs and the screen each one opens.
type demo struct {
	selected int
	reading  *library
	buttons  *buttonsPage
	ink      *inkPage
	gestures *gesturesPage
	panel    *panelLab
	monitor  *monitor
}

// newDemo makes every screen.
func newDemo(device app.Device) *demo {
	return &demo{
		reading:  newLibrary(),
		buttons:  newButtonsPage(),
		ink:      newInkPage(device.NewInk()),
		gestures: newGesturesPage(device.Input),
		panel:    newPanelLab(),
		monitor:  newMonitor(),
	}
}

// draw builds the tabs and the screen that is open.
func (shown *demo) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(screenMarginInLayoutPixels), Gap: largeGapInLayoutPixels}, func() {
		if context.Tabs(pagesTabsName, &shown.selected, pageNames...) {
			shown.ink.stop()
		}
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
			shown.drawPage(context)
		})
	})
}

func (shown *demo) drawPage(context *ui.Context) {
	switch shown.selected {
	case pageReading:
		shown.reading.draw(context)
	case pageButtons:
		shown.buttons.draw(context)
	case pageInk:
		shown.ink.draw(context)
	case pageGestures:
		shown.gestures.draw(context)
	case pagePanelLab:
		shown.panel.draw(context)
	case pageMonitor:
		shown.monitor.draw(context)
	}
}

// inkShowing says whether the Ink screen is open.
func (shown *demo) inkShowing() bool { return shown.selected == pageInk }

// takeFrameMode gives the refresh mode for the frame about to be painted, using up a mode the Screen
// page asked for.
func (shown *demo) takeFrameMode() qtfb.RefreshMode {
	if mode, asked := shown.panel.takeMode(); asked {
		return mode
	}
	return pageRefreshModes[shown.selected]
}

// newProgram joins the demo's screens to the tablet.
func newProgram(device app.Device) app.Program {
	return programOf(newDemo(device))
}

func programOf(shown *demo) app.Program {
	return app.Program{
		Draw:        shown.draw,
		RefreshMode: shown.takeFrameMode,
		Ink:         shown.ink.surface,
		InkOpen:     shown.inkShowing,
		TakeTurn:    shown.panel.takeTurn,
		OnEvent:     shown.monitor.record,
	}
}

func main() {
	if err := app.Run(newProgram); err != nil {
		slog.Error("stopped", "error", err)
		os.Exit(1)
	}
}
