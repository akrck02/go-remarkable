package app

import (
	"testing"
	"time"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/canvas"
	"github.com/akrck02/go-remarkable/pen"
	"github.com/akrck02/go-remarkable/qtfb"
)

// standInDisplay counts refreshes; apptest has a fuller one, but it needs this package.
type standInDisplay struct {
	refreshes   int
	mode        qtfb.RefreshMode
	modes       []qtfb.RefreshMode // the mode of each refresh, in order
	areas       []canvas.Rectangle
	modeChanges int
}

func (target *standInDisplay) SetRefreshMode(mode qtfb.RefreshMode) error {
	target.mode = mode
	target.modeChanges++
	return nil
}

func (target *standInDisplay) Refresh(area canvas.Rectangle) error {
	target.refreshes++
	target.modes = append(target.modes, target.mode)
	target.areas = append(target.areas, area)
	return nil
}

func emptyProgram(Device) Program {
	return Program{Draw: func(*ui.Context) {}}
}

func newApplicationForTest(test *testing.T) (*Application, *standInDisplay) {
	test.Helper()
	target := &standInDisplay{}
	memory := make([]byte, PanelWidthInPixels*PanelHeightInPixels*canvas.BytesPerScreenPixel)
	application, err := New(target, memory, pen.Plain, emptyProgram)
	if err != nil {
		test.Fatal(err)
	}
	return application, target
}

func TestAPressThatChangesNothingIsNotTimedByALaterFrame(test *testing.T) {
	application, _ := newApplicationForTest(test)
	application.Apply(qtfb.Event{Kind: qtfb.TouchPress, X: 40, Y: 40})
	application.Apply(qtfb.Event{Kind: qtfb.TouchRelease, X: 40, Y: 40})
	if err := application.Update(); err != nil {
		test.Fatal(err)
	}
	if err := application.Update(); err != nil {
		test.Fatal(err)
	}
	if !application.liftedAt.IsZero() {
		test.Fatal("the lift is still waiting for a frame to report it")
	}
}

func TestTheFirstFrameIsAlreadyTheRightWayRound(test *testing.T) {
	application, target := newApplicationForTest(test)
	events := make(chan qtfb.Event, 1)
	events <- qtfb.Event{Kind: qtfb.RotationChanged, X: int32(qtfb.RotationHalf)}
	application.applyInitialState(events)
	if err := application.Update(); err != nil {
		test.Fatal(err)
	}
	if application.frame.Turn != canvas.UpsideDown || target.refreshes != 1 {
		test.Fatalf("orientation %v, %d refreshes", application.frame.Turn, target.refreshes)
	}
}

func TestStartingDoesNotWaitLongWhenTheSystemSaysNothing(test *testing.T) {
	application, _ := newApplicationForTest(test)
	start := time.Now()
	application.applyInitialState(make(chan qtfb.Event))
	if time.Since(start) > 2*initialStateWait {
		test.Fatal("starting waited too long")
	}
}

func TestAProgramNeedsOnlyToDraw(test *testing.T) {
	application, _ := newApplicationForTest(test)
	application.Apply(qtfb.Event{Kind: qtfb.PenPress, X: 40, Y: 40})
	if err := application.Answer(); err != nil {
		test.Fatal(err)
	}
}
