package main

import (
	"fmt"
	"strings"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/paper"
)

const (
	listWidthInLayoutPixels = 230
	readingTextSize         = 20
	charactersPerLine       = 44
	linesPerPage            = 36
)

// note is one document of the example library.
type note struct {
	title string
	lines []string
}

// library is the Reading screen: a list of notes and one note read page by page.
type library struct {
	notes    []note
	selected int
	page     int
}

// newLibrary fills the library with made-up notes, wrapped for the reading column.
func newLibrary() *library {
	return &library{notes: []note{
		{"Lighthouses", paper.WrapText(paragraphs(40), charactersPerLine)},
		{"Tides", paper.WrapText(paragraphs(14), charactersPerLine)},
		{"Knots", paper.WrapText(paragraphs(2), charactersPerLine)},
	}}
}

// paragraphs makes example text. Each paragraph is numbered, so no two pages look alike.
func paragraphs(count int) string {
	var text strings.Builder
	for number := 1; number <= count; number++ {
		fmt.Fprintf(&text, "%d. A note is only a text file in a folder. This screen stays completely still until "+
			"the pen touches it, because an e-ink panel keeps its picture for free and pays for every change. ", number)
	}
	return text.String()
}

// pageCount is how many pages a note takes, at least one.
func pageCount(lines []string) int {
	return max((len(lines)+linesPerPage-1)/linesPerPage, 1)
}

// pageLines gives the lines of one page.
func pageLines(lines []string, page int) []string {
	start := min(page*linesPerPage, len(lines))
	return lines[start:min(start+linesPerPage, len(lines))]
}

// draw builds the screen.
func (shown *library) draw(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), Height: ui.Grow(), Gap: screenMarginInLayoutPixels}, func() {
		shown.drawList(context)
		shown.drawNote(context)
	})
}

// drawList shows the notes; choosing one opens its first page.
func (shown *library) drawList(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(listWidthInLayoutPixels), Height: ui.Grow(), Gap: mediumGapInLayoutPixels}, func() {
		context.Heading("Notes")
		for index, entry := range shown.notes {
			if paper.Button(context, entry.title) {
				shown.selected, shown.page = index, 0
			}
		}
	})
}

// drawNote shows the open page and the buttons that turn it.
func (shown *library) drawNote(context *ui.Context) {
	current := shown.notes[shown.selected]
	pages := pageCount(current.lines)
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: lineGapInLayoutPixels}, func() {
		context.Heading(current.title)
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
			for _, line := range pageLines(current.lines, shown.page) {
				context.Text(line, ui.TextStyle{Size: readingTextSize})
			}
		})
		context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), Gap: mediumGapInLayoutPixels, AlignY: ui.Center}, func() {
			if paper.Button(context, "Previous") && shown.page > 0 {
				shown.page--
			}
			if paper.Button(context, "Next") && shown.page < pages-1 {
				shown.page++
			}
			context.Label(fmt.Sprintf("Page %d of %d", shown.page+1, pages))
		})
	})
}
