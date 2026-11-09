package pointer

import (
	"testing"
	"time"

	"github.com/akrck02/go-remarkable/qtfb"

	"github.com/akrck02/go-renderer/input"
)

func touch(kind qtfb.EventKind, device int32) qtfb.Event {
	return qtfb.Event{Kind: kind, Device: device, X: 100, Y: 100}
}

func penEvent(kind qtfb.EventKind) qtfb.Event {
	return qtfb.Event{Kind: kind, X: 100, Y: 100}
}

func TestATouchReportedAsPressedManyTimesDoesNotKillThePointer(test *testing.T) {
	state := input.NewState()
	router := New(1)
	for _, event := range []qtfb.Event{touch(qtfb.TouchPress, 0), touch(qtfb.TouchPress, 0), touch(qtfb.TouchPress, 0), touch(qtfb.TouchRelease, 0)} {
		router.Apply(state, event)
	}
	if state.Button(input.MouseLeft) {
		test.Fatal("the button stayed down after the finger lifted")
	}
	router.Apply(state, touch(qtfb.TouchPress, 0))
	if !state.Button(input.MouseLeft) {
		test.Fatal("a new touch did not press: the pointer is dead")
	}
}

func TestAPalmRestingBeforeThePenDoesNotStopThePen(test *testing.T) {
	state := input.NewState()
	router := New(1)
	router.Apply(state, touch(qtfb.TouchPress, 3))
	router.Apply(state, penEvent(qtfb.PenPress))
	if !state.Button(input.MouseLeft) {
		test.Fatal("the pen did not press with a palm on the screen")
	}
	router.Apply(state, penEvent(qtfb.PenRelease))
	if state.Button(input.MouseLeft) {
		test.Fatal("the button stayed down after the pen lifted")
	}
}

func TestAPalmLandingWhileThePenIsDownIsIgnored(test *testing.T) {
	state := input.NewState()
	router := New(1)
	router.Apply(state, penEvent(qtfb.PenPress))
	router.Apply(state, touch(qtfb.TouchPress, 3))
	if !state.Button(input.MouseLeft) {
		test.Fatal("a palm let go of the pen's press")
	}
}

func TestThePointerWorksAfterAPenAndPalmSequence(test *testing.T) {
	state := input.NewState()
	router := New(1)
	for _, event := range []qtfb.Event{touch(qtfb.TouchPress, 3), penEvent(qtfb.PenPress), penEvent(qtfb.PenRelease), touch(qtfb.TouchRelease, 3)} {
		router.Apply(state, event)
	}
	router.Apply(state, touch(qtfb.TouchPress, 0))
	if !state.Button(input.MouseLeft) {
		test.Fatal("a touch after the sequence did not press")
	}
}

func TestAPressWhoseReleaseWasLostDoesNotKillThePointer(test *testing.T) {
	state := input.NewState()
	router := New(1)
	now := time.Unix(1000, 0)
	router.clock = func() time.Time { return now }
	router.Apply(state, touch(qtfb.TouchPress, 0))
	now = now.Add(5 * time.Second)
	router.Apply(state, touch(qtfb.TouchPress, 7))
	router.Apply(state, touch(qtfb.TouchRelease, 7))
	if state.Button(input.MouseLeft) {
		test.Fatal("the button stayed down after the new finger lifted")
	}
	now = now.Add(5 * time.Second)
	router.Apply(state, touch(qtfb.TouchPress, 9))
	if !state.Button(input.MouseLeft) {
		test.Fatal("a later touch did not press: the pointer is dead")
	}
}

func TestPenMovementWithoutAPressDoesNothing(test *testing.T) {
	state := input.NewState()
	New(1).Apply(state, penEvent(qtfb.PenMove))
	if state.Button(input.MouseLeft) {
		test.Fatal("a hovering pen pressed the button")
	}
}
