package main

import (
	"fmt"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/paper"
)

const (
	backButtonName     = "Back to the shelf"
	nextButtonName     = "Next page"
	previousButtonName = "Previous page"

	readingTextSize          = 22
	readingLineHeight        = 1.6
	readingColumnWidth       = 620
	charactersPerReadingLine = 54
	linesPerReadingPage      = 21
	readingColumnPadding     = 8
)

// pagesOf breaks a book into pages of lines, with an empty line between paragraphs.
func pagesOf(reading *book) [][]string {
	var pages [][]string
	var page []string
	for _, paragraph := range reading.paragraphs {
		if len(page) > 0 && len(page) < linesPerReadingPage {
			page = append(page, "")
		}
		for _, line := range paper.WrapText(paragraph, charactersPerReadingLine) {
			if len(page) == linesPerReadingPage {
				pages, page = append(pages, page), nil
			}
			page = append(page, line)
		}
	}
	return append(pages, page)
}

func (shown *library) drawReader(context *ui.Context) {
	pages := pagesOf(shown.current)
	shown.current.page = min(shown.current.page, len(pages)-1)
	context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), AlignY: ui.Center}, func() {
		if paper.Button(context, backButtonName) {
			shown.screen = screenShelf
		}
		context.Spacer()
		context.Text(shown.current.author, ui.TextStyle{Size: captionTextSize, Color: paper.SoftInk})
	})
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(readingColumnWidth), Height: ui.Grow(), Gap: largeGapInLayoutPixels, Padding: ui.PadXY(readingColumnPadding, 0)}, func() {
			context.Text(shown.current.title, ui.TextStyle{Size: titleTextSize})
			drawPageLines(context, pages[shown.current.page])
		})
	})
	shown.drawPageFooter(context, len(pages))
}

func drawPageLines(context *ui.Context, lines []string) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
		for _, line := range lines {
			context.Text(line+" ", ui.TextStyle{Size: readingTextSize, LineHeight: readingLineHeight}) // the space keeps an empty line tall
		}
	})
}

func (shown *library) drawPageFooter(context *ui.Context, pages int) {
	current := shown.current
	paper.ProgressBar(context, float32(current.page+1)/float32(pages))
	context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), AlignY: ui.Center, Gap: mediumGapInLayoutPixels}, func() {
		context.Text(fmt.Sprintf("Page %d of %d", current.page+1, pages), ui.TextStyle{Size: captionTextSize, Color: paper.SoftInk})
		context.Spacer()
		if paper.Button(context, previousButtonName) && current.page > 0 {
			current.page--
		}
		if paper.PrimaryButton(context, nextButtonName) && current.page < pages-1 {
			current.page++
		}
	})
}
