package main

import (
	"fmt"
	"math"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/paper"
)

const (
	cardsName                    = "Cards"
	rowsName                     = "Rows"
	cardCount                    = 4
	rowCount                     = 40
	swipeThresholdInLayoutPixels = 60
	cardHeightInLayoutPixels     = 160
	rowHeightInLayoutPixels      = 36
	degreesPerRadian             = 180 / math.Pi
)

// gesturesPage is the Gestures screen: Go UI's swipe, and Go Renderer's two fingers, each with what
// it measured. The list follows the rest of the tablet: two fingers scroll it, and a swipe down
// goes to its end. So it moves whole rows by itself rather than using a scroll area, which one
// finger would drag.
type gesturesPage struct {
	state       *input.State
	follower    input.TwoFingers
	card        int
	spread      float64
	turnDegrees float64
	firstRow    int              // the row at the top of the list, counting from zero
	rowsArea    draw2d.Rectangle // where the list was in the last frame
	rowsDrag    float64          // how far two fingers moved the list, short of a whole row
}

// newGesturesPage makes the page, reading fingers from a state.
func newGesturesPage(state *input.State) *gesturesPage {
	return &gesturesPage{state: state, spread: 1}
}

// settle moves to the next or previous card once a swipe ends, if it went far enough.
func (page *gesturesPage) settle(offset float32) {
	if offset <= -swipeThresholdInLayoutPixels && page.card < cardCount-1 {
		page.card++
	}
	if offset >= swipeThresholdInLayoutPixels && page.card > 0 {
		page.card--
	}
}

// followFingers adds up what two fingers did since the last build.
func (page *gesturesPage) followFingers() {
	gesture, following := page.follower.Step(page.state)
	if !following {
		return
	}
	page.spread *= gesture.Spread
	page.turnDegrees += gesture.Turn * degreesPerRadian
	if page.rowsArea.Contains(float32(gesture.CenterX), float32(gesture.CenterY)) {
		page.scroll(gesture.PanY)
	}
}

// visibleRows is how many whole rows fit in the list.
func (page *gesturesPage) visibleRows() int {
	return max(int(page.rowsArea.Height)-2*mediumGapInLayoutPixels, 0) / rowHeightInLayoutPixels
}

// lastFirstRow is the top row when the list shows its end; zero when every row fits.
func (page *gesturesPage) lastFirstRow() int {
	return max(rowCount-page.visibleRows(), 0)
}

// goToEnd shows the end of the list. When every row fits there is nowhere to go, and nothing moves.
func (page *gesturesPage) goToEnd() {
	page.firstRow = page.lastFirstRow()
}

// scroll moves the list with the fingers, by whole rows: fingers going up bring later rows.
func (page *gesturesPage) scroll(fingersMovedInLayoutPixels float64) {
	page.rowsDrag += fingersMovedInLayoutPixels
	rows := int(page.rowsDrag / rowHeightInLayoutPixels)
	page.rowsDrag -= float64(rows * rowHeightInLayoutPixels)
	page.firstRow = min(max(page.firstRow-rows, 0), page.lastFirstRow())
}

// draw builds the screen.
func (page *gesturesPage) draw(context *ui.Context) {
	page.followFingers()
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: largeGapInLayoutPixels}, func() {
		context.Heading("Swipe sideways")
		page.drawCards(context)
		context.Heading("Two fingers scroll, a swipe down goes to the end")
		page.drawRows(context)
		context.Heading("Pinch and twist anywhere")
		context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), Gap: largeGapInLayoutPixels, AlignY: ui.Center}, func() {
			context.Label(fmt.Sprintf("Spread: %.2f  Twist: %.0f degrees", page.spread, page.turnDegrees))
			if paper.Button(context, "Reset") {
				page.spread, page.turnDegrees = 1, 0
			}
		})
	})
}

func (page *gesturesPage) drawCards(context *ui.Context) {
	look := ui.Look{Border: context.Theme.ControlLine, BorderWidth: lineWidthInLayoutPixels}
	context.SwipeBox(cardsName, ui.Layout{Width: ui.Grow(), Height: ui.Fixed(cardHeightInLayoutPixels), Padding: ui.Pad(lineWidthInLayoutPixels)}, ui.Horizontal, func(swipe ui.Swipe) {
		if swipe.Released {
			page.settle(swipe.Offset)
		}
		context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center, AlignY: ui.Center}, look, func() {
			context.Heading(fmt.Sprintf("Card %d of %d", page.card+1, cardCount))
		})
	})
}

func (page *gesturesPage) drawRows(context *ui.Context) {
	look := ui.Look{Border: context.Theme.ControlLine, BorderWidth: lineWidthInLayoutPixels}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(mediumGapInLayoutPixels)}
	context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, look, func() {
		context.SwipeBox(rowsName, layout, ui.Vertical, func(swipe ui.Swipe) {
			page.rowsArea = swipe.Area
			page.firstRow = min(page.firstRow, page.lastFirstRow()) // the list may have grown with a turn
			if swipe.Released && swipe.Offset >= swipeThresholdInLayoutPixels {
				page.goToEnd()
			}
			page.drawVisibleRows(context)
		})
	})
}

func (page *gesturesPage) drawVisibleRows(context *ui.Context) {
	for number := page.firstRow + 1; number <= min(page.firstRow+page.visibleRows(), rowCount); number++ {
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(rowHeightInLayoutPixels), AlignY: ui.Center}, func() {
			context.Label(fmt.Sprintf("Row %d", number))
		})
	}
}
