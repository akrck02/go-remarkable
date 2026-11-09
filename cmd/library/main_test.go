package main

import (
	"testing"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/apptest"
	"github.com/akrck02/go-remarkable/qtfb"
)

func startLibrary(test *testing.T) (*library, *app.Application) {
	test.Helper()
	var shown *library
	application, _ := apptest.Start(test, func(device app.Device) app.Program {
		shown = newLibrary(device)
		return programOf(shown)
	})
	return shown, application
}

func TestContinueOpensTheBookOnThePageItWasLeft(test *testing.T) {
	shown, application := startLibrary(test)
	shown.current.page = 1
	apptest.Tap(test, application, continueButtonName)
	if shown.screen != screenReader || shown.current.page != 1 {
		test.Fatalf("screen %d, page %d", shown.screen, shown.current.page)
	}
}

func TestChoosingABookFromTheShelfOpensThatBook(test *testing.T) {
	shown, application := startLibrary(test)
	apptest.Tap(test, application, shown.books[2].title)
	if shown.screen != screenReader || shown.current != shown.books[2] {
		test.Fatalf("screen %d, reading %q", shown.screen, shown.current.title)
	}
}

func TestTheLastPageHasNoNextOne(test *testing.T) {
	shown, application := startLibrary(test)
	shown.open(shown.books[2])
	apptest.Update(test, application)
	for range len(pagesOf(shown.books[2])) + 2 {
		apptest.Tap(test, application, nextButtonName)
	}
	if last := len(pagesOf(shown.books[2])) - 1; shown.current.page != last {
		test.Fatalf("page %d, last %d", shown.current.page, last)
	}
}

func TestEveryCoverHasTheShapeOfItsPlaceOnTheShelf(test *testing.T) {
	for _, entry := range newBooks() {
		size := entry.cover.Pixels.Bounds().Size()
		if size.X*heroCoverHeightInLayoutPixels != size.Y*heroCoverWidthInLayoutPixels {
			test.Errorf("%q has a cover of %v, which would be cropped", entry.title, size)
		}
	}
}

func TestEveryLineOfABookIsOnSomePage(test *testing.T) {
	for _, entry := range newBooks() {
		count := 0
		for _, page := range pagesOf(entry) {
			if len(page) > linesPerReadingPage {
				test.Errorf("%q has a page of %d lines", entry.title, len(page))
			}
			count += len(page)
		}
		if count == 0 {
			test.Errorf("%q is empty", entry.title)
		}
	}
}

func TestThePenWritesOnlyOnTheNotebook(test *testing.T) {
	shown, application := startLibrary(test)
	pen := func(kind qtfb.EventKind) qtfb.Event { return qtfb.Event{Kind: kind, X: 800, Y: 1000} }
	apptest.Send(test, application, pen(qtfb.PenPress), pen(qtfb.PenRelease))
	if shown.page.StrokeCount() != 0 {
		test.Fatal("the shelf took a stroke")
	}
	shown.navigate(screenNotebook)
	apptest.Update(test, application)
	apptest.Send(test, application, pen(qtfb.PenPress), pen(qtfb.PenRelease))
	if shown.page.StrokeCount() != 1 {
		test.Fatalf("%d strokes", shown.page.StrokeCount())
	}
}
