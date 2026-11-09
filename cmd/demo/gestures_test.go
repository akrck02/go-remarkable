package main

import (
	"testing"

	"github.com/akrck02/go-remarkable/qtfb"

	"github.com/akrck02/go-renderer/input"
)

func finger(kind qtfb.EventKind, device, x, y int32) qtfb.Event {
	return qtfb.Event{Kind: kind, Device: device, X: x, Y: y}
}

func TestTwoFingersMovingApartReportASpreadAboveOne(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Gestures", target)
	send(test, shown,
		finger(qtfb.TouchPress, 0, 700, 1600),
		finger(qtfb.TouchPress, 1, 900, 1600),
		finger(qtfb.TouchMove, 1, 900, 1600),
		finger(qtfb.TouchMove, 1, 1100, 1600))
	if spread := shown.demo.gestures.spread; spread <= 1 {
		test.Fatalf("spread %v", spread)
	}
}

func TestTwoFingersTurningReportATwist(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Gestures", target)
	send(test, shown,
		finger(qtfb.TouchPress, 0, 700, 1600),
		finger(qtfb.TouchPress, 1, 900, 1600),
		finger(qtfb.TouchMove, 1, 900, 1600),
		finger(qtfb.TouchMove, 1, 900, 1500))
	if turn := shown.demo.gestures.turnDegrees; turn <= 0 {
		test.Fatalf("turn %v", turn)
	}
}

func TestASidewaysDragChangesThePage(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Gestures", target)
	x, y := middleOf(test, shown, cardsName)
	send(test, shown,
		finger(qtfb.TouchPress, 0, x+300, y),
		finger(qtfb.TouchMove, 0, x+200, y),
		finger(qtfb.TouchMove, 0, x, y),
		finger(qtfb.TouchMove, 0, x-200, y),
		finger(qtfb.TouchRelease, 0, x-200, y))
	if shown.demo.gestures.card != 1 {
		test.Fatalf("card %d", shown.demo.gestures.card)
	}
}

func TestAShortDragStaysOnTheSamePage(test *testing.T) {
	page := newGesturesPage(input.NewState())
	page.settle(-swipeThresholdInLayoutPixels / 2)
	if page.card != 0 {
		test.Fatalf("card %d", page.card)
	}
}

func TestTheFirstCardCannotGoFurtherBack(test *testing.T) {
	page := newGesturesPage(input.NewState())
	page.settle(swipeThresholdInLayoutPixels * 2)
	if page.card != 0 {
		test.Fatalf("card %d", page.card)
	}
}

func TestASwipeDownGoesToTheEndOfTheList(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Gestures", target)
	x, y := middleOf(test, shown, rowsName)
	send(test, shown,
		finger(qtfb.TouchPress, 0, x, y-200),
		finger(qtfb.TouchMove, 0, x, y-100),
		finger(qtfb.TouchMove, 0, x, y),
		finger(qtfb.TouchMove, 0, x, y+200),
		finger(qtfb.TouchRelease, 0, x, y+200))
	page := shown.demo.gestures
	if page.firstRow == 0 || page.firstRow != page.lastFirstRow() {
		test.Fatalf("first row %d, last %d", page.firstRow, page.lastFirstRow())
	}
}

func TestASwipeUpLeavesTheListWhereItWas(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Gestures", target)
	x, y := middleOf(test, shown, rowsName)
	send(test, shown,
		finger(qtfb.TouchPress, 0, x, y+200),
		finger(qtfb.TouchMove, 0, x, y+100),
		finger(qtfb.TouchMove, 0, x, y-200),
		finger(qtfb.TouchRelease, 0, x, y-200))
	if first := shown.demo.gestures.firstRow; first != 0 {
		test.Fatalf("first row %d", first)
	}
}

func TestASwipeDownMovesNothingWhenEveryRowFits(test *testing.T) {
	page := newGesturesPage(input.NewState())
	page.rowsArea.Height = (rowCount + 1) * rowHeightInLayoutPixels * 2
	page.goToEnd()
	if page.firstRow != 0 {
		test.Fatalf("first row %d", page.firstRow)
	}
}

func TestTwoFingersMovingUpScrollTheList(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Gestures", target)
	x, y := middleOf(test, shown, rowsName)
	send(test, shown,
		finger(qtfb.TouchPress, 0, x-100, y+100),
		finger(qtfb.TouchPress, 1, x+100, y+100),
		finger(qtfb.TouchMove, 1, x+100, y+100),
		finger(qtfb.TouchMove, 0, x-100, y-100),
		finger(qtfb.TouchMove, 1, x+100, y-100))
	page := shown.demo.gestures
	if page.firstRow == 0 || page.firstRow == page.lastFirstRow() {
		test.Fatalf("first row %d, last %d", page.firstRow, page.lastFirstRow())
	}
}

func TestTwoFingersOutsideTheListLeaveItAlone(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Gestures", target)
	x, y := middleOf(test, shown, cardsName)
	send(test, shown,
		finger(qtfb.TouchPress, 0, x-100, y+50),
		finger(qtfb.TouchPress, 1, x+100, y+50),
		finger(qtfb.TouchMove, 1, x+100, y+50),
		finger(qtfb.TouchMove, 0, x-100, y-50),
		finger(qtfb.TouchMove, 1, x+100, y-50))
	if first := shown.demo.gestures.firstRow; first != 0 {
		test.Fatalf("first row %d", first)
	}
}

func TestScrollingStopsAtBothEndsOfTheList(test *testing.T) {
	page := newGesturesPage(input.NewState())
	page.rowsArea.Height = rowHeightInLayoutPixels * 10
	page.scroll(-rowHeightInLayoutPixels * rowCount * 2)
	if page.firstRow != page.lastFirstRow() {
		test.Fatalf("past the end: first row %d", page.firstRow)
	}
	page.scroll(rowHeightInLayoutPixels * rowCount * 2)
	if page.firstRow != 0 {
		test.Fatalf("past the start: first row %d", page.firstRow)
	}
}
