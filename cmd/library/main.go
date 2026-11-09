// Command library is a quiet place to read and write on the tablet: a shelf of books, a page that
// stays still until it is turned, and a notebook. It is the example of what an interface for paper
// can look like.
package main

import (
	"log/slog"
	"os"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/ink"
	"github.com/akrck02/go-remarkable/paper"
)

// The screens of the program.
const (
	screenShelf = iota
	screenReader
	screenNotebook
)

const (
	marginInLayoutPixels    = 40
	largeGapInLayoutPixels  = 28
	mediumGapInLayoutPixels = 14
	smallGapInLayoutPixels  = 6

	titleTextSize   = 40
	bookTitleSize   = 22
	captionTextSize = 15
	bigIconSize     = 26

	shelfName    = "Library"
	notebookName = "Notebook"
	clearName    = "Start again"
)

// library is the program's state.
type library struct {
	screen  int
	books   []*book
	current *book // the one being read
	page    *ink.Surface
}

func newLibrary(device app.Device) *library {
	books := newBooks()
	return &library{books: books, current: books[0], page: device.NewInk()}
}

func (shown *library) draw(context *ui.Context) {
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(marginInLayoutPixels), Gap: largeGapInLayoutPixels}
	context.Box(layout, func() {
		switch shown.screen {
		case screenShelf:
			shown.drawShelf(context)
		case screenReader:
			shown.drawReader(context)
		case screenNotebook:
			shown.drawNotebook(context)
		}
		if shown.screen != screenReader {
			shown.drawNavigation(context)
		}
	})
}

// open starts reading a book where it was left.
func (shown *library) open(chosen *book) {
	shown.current = chosen
	shown.screen = screenReader
}

// navigate goes to one of the screens in the bottom bar, and never leaves a stroke open.
func (shown *library) navigate(screen int) {
	shown.page.Stop()
	shown.screen = screen
}

func (shown *library) drawNavigation(context *ui.Context) {
	context.Separator()
	context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), AlignX: ui.Center, Gap: navigationGapInLayoutPixels}, func() {
		if navigationItem(context, shelfName, "library", shown.screen == screenShelf) {
			shown.navigate(screenShelf)
		}
		if navigationItem(context, notebookName, "pen-line", shown.screen == screenNotebook) {
			shown.navigate(screenNotebook)
		}
	})
}

// navigationItem is an icon over its name; the current one is in full ink.
func navigationItem(context *ui.Context, name, icon string, current bool) bool {
	colour := paper.QuietInk
	if current {
		colour = paper.Ink
	}
	look := ui.Look{}
	layout := ui.Layout{Direction: ui.Vertical, AlignX: ui.Center, Gap: smallGapInLayoutPixels, Padding: ui.PadXY(navigationPaddingInLayoutPixels, smallGapInLayoutPixels)}
	return context.Clickable(name, layout, ui.Looks{Normal: look, Hovered: look, Pressed: look}, func() {
		context.IconSized(icon, bigIconSize, colour)
		context.Text(name, ui.TextStyle{Size: captionTextSize, Color: colour})
	})
}

func (shown *library) drawNotebook(context *ui.Context) {
	context.Box(ui.Layout{Direction: ui.Horizontal, Width: ui.Grow(), AlignY: ui.Center}, func() {
		context.Text("Field notes", ui.TextStyle{Size: titleTextSize})
		context.Spacer()
		if paper.Button(context, clearName) {
			shown.page.Clear()
		}
	})
	shown.page.DrawPaper(context)
}

func newProgram(device app.Device) app.Program {
	return programOf(newLibrary(device))
}

func programOf(shown *library) app.Program {
	return app.Program{
		Draw:    shown.draw,
		Ink:     shown.page,
		InkOpen: func() bool { return shown.screen == screenNotebook },
	}
}

func main() {
	if err := app.Run(newProgram); err != nil {
		slog.Error("stopped", "error", err)
		os.Exit(1)
	}
}
