package main

import (
	"testing"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/canvas"
	"github.com/akrck02/go-remarkable/qtfb"
	"github.com/akrck02/go-renderer/input"
)

func TestTheTurnButtonTurnsTheScreenWithOneRefresh(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Screen", target)
	tap(test, shown, turnButtonName)
	if shown.Canvas().Turn != canvas.TurnedRight || shown.Canvas().Width != app.PanelHeightInPixels {
		test.Fatalf("canvas.Orientation %v, width %d", shown.Canvas().Turn, shown.Canvas().Width)
	}
	if len(target.Refreshes) != 1 {
		test.Fatalf("turning sent %d refreshes", len(target.Refreshes))
	}
}

func TestATapLandsOnTheControlWhenTheScreenIsTurned(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Screen", target)
	tap(test, shown, turnButtonName)
	canvasX, canvasY := middleOf(test, shown, pagesTabsName, "Buttons")
	// turned right, a canvas point sits at panel (width-1-y, x)
	panelX, panelY := int32(app.PanelWidthInPixels)-1-canvasY, canvasX
	send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: panelX, Y: panelY}, qtfb.Event{Kind: qtfb.TouchRelease, X: panelX, Y: panelY})
	if _, found := shown.Context().Area(tapButtonName); !found {
		test.Fatal("the tap missed the Buttons tab")
	}
}

func TestARotationFromTheServerTurnsTheCanvasToTheMatchingOrientation(test *testing.T) {
	cases := []struct {
		name     string
		rotation qtfb.Rotation
		wanted   canvas.Orientation
	}{
		{"quarter turn one way", qtfb.RotationQuarterLeft, canvas.TurnedRight},
		{"quarter turn the other way", qtfb.RotationQuarterRight, canvas.TurnedLeft},
		{"half turn", qtfb.RotationHalf, canvas.UpsideDown},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			shown, target := newShownApplication(test)
			send(test, shown, qtfb.Event{Kind: qtfb.RotationChanged, X: int32(testCase.rotation)})
			if shown.Canvas().Turn != testCase.wanted {
				test.Fatalf("canvas.Orientation %v", shown.Canvas().Turn)
			}
			if len(target.Refreshes) != 1 {
				test.Fatalf("turning sent %d refreshes", len(target.Refreshes))
			}
		})
	}
}

func TestTheRotationWeAlreadyHavePaintsNothing(test *testing.T) {
	shown, target := newShownApplication(test)
	send(test, shown, qtfb.Event{Kind: qtfb.RotationChanged, X: int32(qtfb.RotationNone)})
	if len(target.Refreshes) != 0 {
		test.Fatalf("sent %d refreshes", len(target.Refreshes))
	}
}

func TestAnUnknownRotationIsIgnored(test *testing.T) {
	shown, _ := newShownApplication(test)
	send(test, shown, qtfb.Event{Kind: qtfb.RotationChanged, X: int32(qtfb.RotationHalf)})
	send(test, shown, qtfb.Event{Kind: qtfb.RotationChanged, X: 9})
	if shown.Canvas().Turn != canvas.UpsideDown {
		test.Fatalf("canvas.Orientation %v", shown.Canvas().Turn)
	}
}

func TestTheTurnButtonGoesOnFromARotationTheServerAskedFor(test *testing.T) {
	shown, target := newShownApplication(test)
	send(test, shown, qtfb.Event{Kind: qtfb.RotationChanged, X: int32(qtfb.RotationHalf)})
	shown.demo.selected = pagePanelLab
	if err := shown.Update(); err != nil {
		test.Fatal(err)
	}
	target.Forget()
	shown.demo.panel.turnAsked = true
	if err := shown.Update(); err != nil {
		test.Fatal(err)
	}
	if shown.Canvas().Turn != canvas.TurnedLeft {
		test.Fatalf("canvas.Orientation %v", shown.Canvas().Turn)
	}
}

func TestTurningTheTabletWinsOverTheButtonInTheSameUpdate(test *testing.T) {
	shown, _ := newShownApplication(test)
	shown.demo.panel.turnAsked = true
	send(test, shown, qtfb.Event{Kind: qtfb.RotationChanged, X: int32(qtfb.RotationHalf)})
	if shown.Canvas().Turn != canvas.UpsideDown || shown.demo.panel.turnAsked {
		test.Fatalf("canvas.Orientation %v, button still asked: %v", shown.Canvas().Turn, shown.demo.panel.turnAsked)
	}
}

func TestAFingerHeldWhileTheTabletTurnsIsLetGo(test *testing.T) {
	shown, _ := newShownApplication(test)
	send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: 400, Y: 400})
	send(test, shown, qtfb.Event{Kind: qtfb.RotationChanged, X: int32(qtfb.RotationHalf)})
	if shown.Input().Button(input.MouseLeft) {
		test.Fatal("a press survived the turn")
	}
}
