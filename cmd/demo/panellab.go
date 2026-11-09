package main

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/paper"
	"github.com/akrck02/go-remarkable/qtfb"
)

const (
	turnButtonName = "Turn the screen"
	noModeAsked    = qtfb.RefreshMode(-1)
	greySteps      = 8

	patchHeightInLayoutPixels = 180
	patchTextSize             = 24
	stepsHeightInLayoutPixels = 80
)

// refreshMode is one way the panel can update, with the name its button shows.
type refreshMode struct {
	name  string
	value qtfb.RefreshMode
}

var (
	refreshModes = []refreshMode{
		{"Ultra fast", qtfb.UltraFast},
		{"Fast", qtfb.Fast},
		{"Animate", qtfb.Animate},
		{"Content", qtfb.Content},
		{"Interface", qtfb.Interface},
	}
	colourSteps = []draw2d.Color{
		{1, 0, 0, 1}, {1, 1, 0, 1}, {0, 1, 0, 1}, {0, 1, 1, 1}, {0, 0, 1, 1}, {1, 0, 1, 1},
	}
)

// panelLab is the Screen screen: it turns the interface, and repaints one patch in each refresh
// mode to compare them by eye.
type panelLab struct {
	turnAsked bool
	modeAsked qtfb.RefreshMode
	patchDark bool
}

// newPanelLab makes the lab with nothing asked for.
func newPanelLab() *panelLab {
	return &panelLab{modeAsked: noModeAsked}
}

// takeTurn says once that the screen should be turned.
func (lab *panelLab) takeTurn() bool {
	asked := lab.turnAsked
	lab.turnAsked = false
	return asked
}

// takeMode gives once the refresh mode the next frame was asked to use.
func (lab *panelLab) takeMode() (qtfb.RefreshMode, bool) {
	mode := lab.modeAsked
	lab.modeAsked = noModeAsked
	return mode, mode != noModeAsked
}

// draw builds the screen.
func (lab *panelLab) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: largeGapInLayoutPixels}, func() {
		if paper.Button(context, turnButtonName) {
			lab.turnAsked = true
		}
		context.Heading("Repaint the patch in each mode")
		context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), Gap: mediumGapInLayoutPixels, Wrap: true}, func() {
			for _, mode := range refreshModes {
				if paper.Button(context, mode.name) {
					lab.modeAsked, lab.patchDark = mode.value, !lab.patchDark
				}
			}
		})
		lab.drawPatch(context)
		context.Heading("Greys")
		drawSteps(context, greys())
		context.Heading("Colours")
		drawSteps(context, colourSteps)
	})
}

func (lab *panelLab) drawPatch(context *ui.Context) {
	fill, ink := draw2d.Color{1, 1, 1, 1}, draw2d.Color{0, 0, 0, 1}
	if lab.patchDark {
		fill, ink = ink, fill
	}
	look := ui.Look{Fill: fill, Border: context.Theme.ControlLine, BorderWidth: lineWidthInLayoutPixels}
	context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(patchHeightInLayoutPixels), AlignX: ui.Center, AlignY: ui.Center}, look, func() {
		context.Text("The quick patch jumps over the lazy panel", ui.TextStyle{Size: patchTextSize, Color: ink})
	})
}

// greys goes from black to white in even steps.
func greys() []draw2d.Color {
	steps := make([]draw2d.Color, greySteps)
	for index := range steps {
		level := float32(index) / (greySteps - 1)
		steps[index] = draw2d.Color{level, level, level, 1}
	}
	return steps
}

func drawSteps(context *ui.Context, colours []draw2d.Color) {
	context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), Height: ui.Fixed(stepsHeightInLayoutPixels)}, func() {
		for _, colour := range colours {
			context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, ui.Look{Fill: colour}, func() {})
		}
	})
}
