package main

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/paper"
)

const (
	continueButtonName = "Continue"

	navigationGapInLayoutPixels     = 72
	navigationPaddingInLayoutPixels = 24

	heroCoverWidthInLayoutPixels   = 168
	heroCoverHeightInLayoutPixels  = 224
	shelfColumns                   = 3
	shelfCoverHeightInLayoutPixels = 306
)

func (shown *library) drawShelf(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Gap: smallGapInLayoutPixels}, func() {
		context.Text("Library", ui.TextStyle{Size: titleTextSize, Weight: draw2d.Bold})
		context.Text(fmt.Sprintf("%d books, and one you are in the middle of", len(shown.books)), ui.TextStyle{Size: captionTextSize + 2, Color: paper.SoftInk})
	})
	shown.drawContinue(context)
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: mediumGapInLayoutPixels}, func() {
		paper.SectionLabel(context, "On the shelf")
		context.Box(ui.Layout{Columns: shelfColumns, Width: ui.Grow(), Gap: mediumGapInLayoutPixels + smallGapInLayoutPixels}, func() {
			for _, entry := range shown.books {
				shown.drawShelfEntry(context, entry)
			}
		})
	})
	context.Spacer()
}

// drawContinue is the one thing on the shelf that asks for attention.
func (shown *library) drawContinue(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: mediumGapInLayoutPixels}, func() {
		paper.SectionLabel(context, "Continue reading")
		paper.Card(context, ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), Padding: ui.Pad(largeGapInLayoutPixels), Gap: largeGapInLayoutPixels}, func() {
			drawCover(context, ui.Layout{Width: ui.Fixed(heroCoverWidthInLayoutPixels), Height: ui.Fixed(heroCoverHeightInLayoutPixels)}, shown.current.cover)
			shown.drawContinueDetails(context)
		})
	})
}

func (shown *library) drawContinueDetails(context *ui.Context) {
	current := shown.current
	pages := len(pagesOf(current))
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(heroCoverHeightInLayoutPixels), Gap: smallGapInLayoutPixels}, func() {
		context.Text(current.title, ui.TextStyle{Size: bookTitleSize + 4, Weight: draw2d.Bold})
		context.Text(current.author, ui.TextStyle{Color: paper.SoftInk})
		context.Spacer()
		paper.ProgressBar(context, float32(current.page+1)/float32(pages))
		context.Text(fmt.Sprintf("Page %d of %d", current.page+1, pages), ui.TextStyle{Size: captionTextSize, Color: paper.SoftInk})
		context.Spacer()
		if paper.PrimaryButton(context, continueButtonName) {
			shown.open(current)
		}
	})
}

func (shown *library) drawShelfEntry(context *ui.Context, entry *book) {
	look := ui.Look{}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: smallGapInLayoutPixels}
	if context.Clickable(entry.title, layout, ui.Looks{Normal: look, Hovered: look, Pressed: look}, func() {
		drawCover(context, ui.Layout{Width: ui.Grow(), Height: ui.Fixed(shelfCoverHeightInLayoutPixels)}, entry.cover)
		context.Text(entry.title, ui.TextStyle{Weight: draw2d.Medium})
		context.Text(entry.author, ui.TextStyle{Size: captionTextSize, Color: paper.SoftInk})
	}) {
		shown.open(entry)
	}
}
