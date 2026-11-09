package main

import (
	"testing"
	"time"

	"github.com/akrck02/go-remarkable/qtfb"
)

func TestTwoTapsCloseTogetherAreADoubleTap(test *testing.T) {
	now := time.Unix(1000, 0)
	page := newButtonsPage()
	page.clock = func() time.Time { return now }
	page.noteTap()
	now = now.Add(doubleTapWindow / 2)
	page.noteTap()
	if page.taps != 2 || page.doubleTaps != 1 {
		test.Fatalf("taps %d, double taps %d", page.taps, page.doubleTaps)
	}
}

func TestTwoTapsFarApartAreNotADoubleTap(test *testing.T) {
	now := time.Unix(1000, 0)
	page := newButtonsPage()
	page.clock = func() time.Time { return now }
	page.noteTap()
	now = now.Add(2 * doubleTapWindow)
	page.noteTap()
	if page.doubleTaps != 0 {
		test.Fatalf("double taps %d", page.doubleTaps)
	}
}

func TestAThirdTapDoesNotMakeASecondDoubleTap(test *testing.T) {
	now := time.Unix(1000, 0)
	page := newButtonsPage()
	page.clock = func() time.Time { return now }
	page.noteTap()
	page.noteTap()
	page.noteTap()
	if page.doubleTaps != 1 {
		test.Fatalf("double taps %d", page.doubleTaps)
	}
}

func TestHoldingAButtonIsALongPress(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Buttons", target)
	now := time.Unix(1000, 0)
	shown.Input().Now = func() time.Time { return now }
	x, y := middleOf(test, shown, tapButtonName)
	send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: x, Y: y})
	now = now.Add(time.Second)
	if err := shown.Update(); err != nil {
		test.Fatal(err)
	}
	if shown.demo.buttons.longPresses != 1 {
		test.Fatalf("long presses %d", shown.demo.buttons.longPresses)
	}
}
