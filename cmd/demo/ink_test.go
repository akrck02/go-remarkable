package main

import (
	"testing"
	"time"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/apptest"
	"github.com/akrck02/go-remarkable/canvas"
	"github.com/akrck02/go-remarkable/ink"
	"github.com/akrck02/go-remarkable/pen"
	"github.com/akrck02/go-remarkable/qtfb"
)

func openInk(test *testing.T) (*shownDemo, *apptest.Display, int32, int32) {
	test.Helper()
	shown, target := newShownApplication(test)
	openPage(test, shown, "Ink", target)
	x, y := middleOf(test, shown, ink.PaperName)
	return shown, target, x, y
}

func pixelIsDark(shown *shownDemo, x, y int32) bool {
	return shown.Canvas().Pixels[(int(y)*shown.Canvas().Width+int(x))*canvas.BytesPerRGBPixel] < canvas.MaximumChannel/2
}

func TestAStrokeRefreshesOnlyItsArea(test *testing.T) {
	shown, target, x, y := openInk(test)
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y}, qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y})
	if len(target.Refreshes) == 0 {
		test.Fatal("the stroke reached no refresh")
	}
	for _, area := range target.Refreshes {
		if area.Right-area.Left > 80 || area.Bottom-area.Top > 40 {
			test.Errorf("refreshed %+v for a short line", area)
		}
	}
	if !pixelIsDark(shown, x+20, y) {
		test.Error("the line's pixels were not painted")
	}
}

// aStroke is a short line from a press to a lift.
func aStroke(x, y int32) []qtfb.Event {
	return []qtfb.Event{
		{Kind: qtfb.PenPress, X: x, Y: y},
		{Kind: qtfb.PenMove, X: x + 40, Y: y},
		{Kind: qtfb.PenRelease, X: x + 40, Y: y},
	}
}

// controlTheClock makes the Ink screen read a clock the test moves.
func controlTheClock(shown *shownDemo) *time.Time {
	now := time.Unix(1000, 0)
	shown.demo.ink.surface.Clock = func() time.Time { return now }
	return &now
}

func answerNow(test *testing.T, shown *shownDemo) {
	test.Helper()
	if err := shown.Answer(); err != nil {
		test.Fatal(err)
	}
}

func TestStrokesAreCleanedOnceThePenHasRested(test *testing.T) {
	shown, target, x, y := openInk(test)
	now := controlTheClock(shown)
	send(test, shown, aStroke(x, y)...)
	if target.Modes[len(target.Modes)-1] != qtfb.UltraFast {
		test.Fatalf("modes %v", target.Modes)
	}
	*now = now.Add(ink.CleanDelay)
	answerNow(test, shown)
	if target.Modes[len(target.Modes)-1] != qtfb.Content {
		test.Fatalf("no clean refresh: modes %v", target.Modes)
	}
}

func TestWritingStaysInOneModeFromStrokeToStroke(test *testing.T) {
	shown, target, x, y := openInk(test)
	now := controlTheClock(shown)
	send(test, shown, aStroke(x, y)...)
	*now = now.Add(ink.CleanDelay / 2)
	send(test, shown, aStroke(x, y+60)...)
	if len(target.Modes) != 1 {
		test.Fatalf("modes changed while writing: %v", target.Modes)
	}
}

func TestThePanelGoesBackToTheWritingModeAfterCleaning(test *testing.T) {
	shown, target, x, y := openInk(test)
	now := controlTheClock(shown)
	send(test, shown, aStroke(x, y)...)
	*now = now.Add(ink.CleanDelay)
	answerNow(test, shown)
	*now = now.Add(app.InkRearmDelay)
	answerNow(test, shown)
	if target.Modes[len(target.Modes)-1] != qtfb.UltraFast {
		test.Fatalf("modes %v", target.Modes)
	}
}

func TestAStrokeBegunRightAfterALiftStillGetsTheFirstOneCleaned(test *testing.T) {
	shown, target, x, y := openInk(test)
	now := controlTheClock(shown)
	sendTogether(test, shown, append(aStroke(x, y), aStroke(x, y+200)...)...)
	target.Forget()
	*now = now.Add(ink.CleanDelay)
	answerNow(test, shown)
	if len(target.Refreshes) != 1 || target.Refreshes[0].Bottom-target.Refreshes[0].Top < 200 {
		test.Fatalf("clean refreshes %+v", target.Refreshes)
	}
}

func TestAPenReleaseThatNeverCameDoesNotKillTheFingers(test *testing.T) {
	shown, target, x, y := openInk(test)
	now := controlTheClock(shown)
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y})
	*now = now.Add(time.Hour)
	tap(test, shown, pagesTabsName, "Buttons")
	if shown.demo.selected != pageButtons || len(target.Refreshes) == 0 {
		test.Fatal("a tap did nothing after a lost pen release")
	}
}

func TestAPenPressOnATabAfterALostReleaseWorks(test *testing.T) {
	shown, _, x, y := openInk(test)
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y})
	tabX, tabY := middleOf(test, shown, pagesTabsName, "Buttons")
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: tabX, Y: tabY}, qtfb.Event{Kind: qtfb.PenRelease, X: tabX, Y: tabY})
	if shown.demo.selected != pageButtons {
		test.Fatal("the pen did not change the tab")
	}
}

func TestClearingAsThePenLandsDoesNotCrash(test *testing.T) {
	shown, _, x, y := openInk(test)
	clearX, clearY := middleOf(test, shown, clearButtonName)
	send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: clearX, Y: clearY})
	sendTogether(test, shown, qtfb.Event{Kind: qtfb.TouchRelease, X: clearX, Y: clearY}, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y})
	send(test, shown, qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y}, qtfb.Event{Kind: qtfb.PenRelease, X: x + 40, Y: y})
}

func TestChangingTabAsThePenLandsLeavesNoStrokeOpen(test *testing.T) {
	shown, _, x, y := openInk(test)
	tabX, tabY := middleOf(test, shown, pagesTabsName, "Buttons")
	send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: tabX, Y: tabY})
	sendTogether(test, shown, qtfb.Event{Kind: qtfb.TouchRelease, X: tabX, Y: tabY}, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y})
	if shown.demo.ink.surface.Writing() {
		test.Fatal("the stroke stayed open behind another screen")
	}
}

func TestStrokesStayWhereTheyAreOnThePanelWhenTheScreenTurns(test *testing.T) {
	shown, _, x, y := openInk(test)
	send(test, shown, aStroke(x, y)...)
	shown.demo.panel.turnAsked = true
	if err := shown.Update(); err != nil {
		test.Fatal(err)
	}
	wantedX, wantedY := canvas.TurnedRight.ToCanvas(x, y, app.PanelWidthInPixels, app.PanelHeightInPixels)
	if !pixelIsDark(shown, wantedX, wantedY) {
		test.Fatalf("no ink at %d,%d after the turn", wantedX, wantedY)
	}
}

func TestAStrokeDoesNotRebuildTheInterface(test *testing.T) {
	shown, _, x, y := openInk(test)
	before := shown.builds
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y}, qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y})
	if shown.builds != before {
		test.Fatal("a stroke asked to rebuild the interface")
	}
}

func TestATouchDuringAStrokeChangesNothingInIt(test *testing.T) {
	shown, target, x, y := openInk(test)
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y})
	target.Forget()
	send(test, shown,
		qtfb.Event{Kind: qtfb.TouchPress, X: x + 200, Y: y + 200, Device: 3},
		qtfb.Event{Kind: qtfb.TouchMove, X: x + 210, Y: y + 210, Device: 3},
		qtfb.Event{Kind: qtfb.TouchRelease, X: x + 210, Y: y + 210, Device: 3})
	if len(target.Refreshes) != 0 {
		test.Fatalf("a palm sent %d refreshes", len(target.Refreshes))
	}
	send(test, shown, qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y})
	if !pixelIsDark(shown, x+20, y) {
		test.Error("the stroke did not go on after the palm")
	}
}

func TestStrokesSurviveARepaintOfTheInterface(test *testing.T) {
	shown, target, x, y := openInk(test)
	send(test, shown,
		qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y},
		qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y},
		qtfb.Event{Kind: qtfb.PenRelease, X: x + 40, Y: y})
	openPage(test, shown, "Buttons", target)
	openPage(test, shown, "Ink", target)
	if !pixelIsDark(shown, x+20, y) {
		test.Error("the stroke was lost")
	}
}

func TestClearingRemovesTheStrokes(test *testing.T) {
	shown, _, x, y := openInk(test)
	send(test, shown,
		qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y},
		qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y},
		qtfb.Event{Kind: qtfb.PenRelease, X: x + 40, Y: y})
	tap(test, shown, clearButtonName)
	if pixelIsDark(shown, x+20, y) {
		test.Error("the stroke is still there")
	}
}

func TestTheEraserPaintsPaper(test *testing.T) {
	shown, _, x, y := openInk(test)
	send(test, shown,
		qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y},
		qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y},
		qtfb.Event{Kind: qtfb.PenRelease, X: x + 40, Y: y})
	shown.demo.ink.surface.Sample = func() pen.Sample { return pen.Sample{Readable: true, Erasing: true} }
	send(test, shown,
		qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y},
		qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y},
		qtfb.Event{Kind: qtfb.PenRelease, X: x + 40, Y: y})
	if pixelIsDark(shown, x+20, y) {
		test.Error("the eraser left the line")
	}
}

func TestTheEraserMarksItsPathWhileItTouchesAndNotAfter(test *testing.T) {
	shown, target, x, y := openInk(test)
	shown.demo.ink.surface.Sample = func() pen.Sample { return pen.Sample{Readable: true, Erasing: true} }
	send(test, shown,
		qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y},
		qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y})
	if !pixelIsDark(shown, x+20, y) {
		test.Error("the eraser shows nothing while it touches")
	}
	if len(target.Refreshes) == 0 {
		test.Error("the mark did not reach the panel")
	}
	send(test, shown, qtfb.Event{Kind: qtfb.PenRelease, X: x + 40, Y: y})
	if pixelIsDark(shown, x+20, y) {
		test.Error("the mark stayed after lifting")
	}
}

func TestTheEraserMarksExactlyTheWidthItRemoves(test *testing.T) {
	shown, _, x, y := openInk(test)
	shown.demo.ink.surface.Sample = func() pen.Sample { return pen.Sample{Readable: true, Erasing: true} }
	send(test, shown,
		qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y},
		qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y})
	inside, outside := int32(ink.EraserWidthInPixels/2-2), int32(ink.EraserWidthInPixels/2+2)
	if !pixelIsDark(shown, x+20, y+inside) {
		test.Error("the mark is narrower than what the eraser removes")
	}
	if pixelIsDark(shown, x+20, y+outside) {
		test.Error("the mark is wider than what the eraser removes")
	}
}

func TestAnUnreadableDeviceLeavesInkWorking(test *testing.T) {
	shown, _, x, y := openInk(test)
	shown.demo.ink.surface.Sample = pen.Plain
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y}, qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y})
	if !pixelIsDark(shown, x+20, y) {
		test.Fatal("no line without the device")
	}
}

func TestAPenOutsideThePaperStillPressesControls(test *testing.T) {
	shown, target, _, _ := openInk(test)
	x, y := middleOf(test, shown, pagesTabsName, "Buttons")
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y}, qtfb.Event{Kind: qtfb.PenRelease, X: x, Y: y})
	if len(target.Refreshes) == 0 || shown.demo.selected != pageButtons {
		test.Fatal("the pen did not change the tab")
	}
}

func TestAPalmOnAButtonDuringAStrokeDoesNotPressIt(test *testing.T) {
	shown, _, x, y := openInk(test)
	send(test, shown, qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y}, qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y})
	clearX, clearY := middleOf(test, shown, clearButtonName)
	send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: clearX, Y: clearY, Device: 3}, qtfb.Event{Kind: qtfb.TouchRelease, X: clearX, Y: clearY, Device: 3})
	if shown.demo.ink.surface.StrokeCount() != 1 {
		test.Fatal("a palm pressed Clear during a stroke")
	}
}

func TestOnlyTheStrokesOwnAreaIsWaitingForTheScreen(test *testing.T) {
	shown, _, x, y := openInk(test)
	shown.Apply(qtfb.Event{Kind: qtfb.PenPress, X: x, Y: y})
	shown.Apply(qtfb.Event{Kind: qtfb.PenMove, X: x + 40, Y: y})
	area := shown.demo.ink.surface.TakePending()
	if area.Right-area.Left > 80 || area.Bottom-area.Top > 40 {
		test.Fatalf("waiting area %+v for a short line", area)
	}
}

func TestAStrokeWaitingToBeCleanedIsNotSentOffThePanelAfterATurn(test *testing.T) {
	shown, target, x, y := openInk(test)
	now := controlTheClock(shown)
	send(test, shown, aStroke(x, y)...)
	shown.demo.panel.turnAsked = true
	answerUpdate(test, shown)
	target.Forget()
	*now = now.Add(ink.CleanDelay)
	answerNow(test, shown)
	if len(target.Refreshes) != 0 {
		test.Fatalf("refreshes %+v after a turn repainted everything", target.Refreshes)
	}
}

func answerUpdate(test *testing.T, shown *shownDemo) {
	test.Helper()
	if err := shown.Update(); err != nil {
		test.Fatal(err)
	}
}
