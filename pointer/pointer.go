// Package pointer turns the tablet's pen and fingers into the single pointer an interface expects.
package pointer

import (
	"time"

	"github.com/akrck02/go-renderer/input"

	"github.com/akrck02/go-remarkable/qtfb"
)

// LostPressSilence is how long with no events means a press that was still open has been lost.
const LostPressSilence = 1500 * time.Millisecond

// penFinger is the number the pen goes by among the fingers; a finger's is its device plus
// firstFinger, so the two never meet.
const (
	penFinger   = int64(1)
	firstFinger = int64(2)
)

// Router turns pen and touch events into go-ui's single pointer.
//
// It stands between AppLoad and input.TouchPointer because the tablet's events are messier than a
// phone's. AppLoad reports a touch point that did not change as another press, and a palm rests on
// the screen while the pen writes. TouchPointer counts presses, so each repeated one would leave it
// believing more fingers are down, and the pointer would never come back.
type Router struct {
	pointer              input.TouchPointer
	pixelsPerLayoutPixel float32
	pressed              map[int64]bool // the fingers down, each counted once
	penDown              bool
	clock                func() time.Time
	last                 time.Time // when the last event came
}

// New makes a router with nothing touching the screen.
func New(pixelsPerLayoutPixel float32) *Router {
	return &Router{pixelsPerLayoutPixel: pixelsPerLayoutPixel, pressed: map[int64]bool{}, clock: time.Now}
}

// forgetLostPresses lets go of everything when a press comes after a long silence. A menu of the
// system's can take a finger's release for itself, and we would believe that finger was still down.
func (router *Router) forgetLostPresses(state *input.State, now time.Time) {
	if now.Sub(router.last) < LostPressSilence || (!router.penDown && len(router.pressed) == 0) {
		return
	}
	router.LetGoOfEverything(state)
}

// LetGoOfEverything forgets every press, for when something else has taken the screen.
func (router *Router) LetGoOfEverything(state *input.State) {
	router.pointer.Cancel(state)
	clear(router.pressed)
	router.penDown = false
}

// Apply feeds one event to the interface. The pen wins: while it is down, touches are a resting
// palm and are ignored.
func (router *Router) Apply(state *input.State, event qtfb.Event) {
	now := router.clock()
	if event.Kind == qtfb.PenPress || event.Kind == qtfb.TouchPress {
		router.forgetLostPresses(state, now)
	}
	router.last = now
	x, y := float32(event.X), float32(event.Y)
	switch event.Kind {
	case qtfb.PenPress:
		router.pointer.Cancel(state)
		clear(router.pressed)
		router.penDown = true
		router.pointer.Begin(state, penFinger, x, y, router.pixelsPerLayoutPixel)
	case qtfb.PenMove:
		if router.penDown {
			router.pointer.Move(state, penFinger, x, y, router.pixelsPerLayoutPixel)
		}
	case qtfb.PenRelease:
		if router.penDown {
			router.penDown = false
			router.pointer.End(state, penFinger, x, y, router.pixelsPerLayoutPixel)
		}
	case qtfb.TouchPress, qtfb.TouchMove, qtfb.TouchRelease:
		if !router.penDown {
			router.applyTouch(state, event.Kind, int64(event.Device)+firstFinger, x, y)
		}
	}
}

// applyTouch follows one finger. A press of a finger already down is only a position.
func (router *Router) applyTouch(state *input.State, kind qtfb.EventKind, finger int64, x, y float32) {
	switch {
	case kind == qtfb.TouchPress && !router.pressed[finger]:
		router.pressed[finger] = true
		router.pointer.Begin(state, finger, x, y, router.pixelsPerLayoutPixel)
	case kind == qtfb.TouchRelease && router.pressed[finger]:
		delete(router.pressed, finger)
		router.pointer.End(state, finger, x, y, router.pixelsPerLayoutPixel)
	case kind != qtfb.TouchRelease && router.pressed[finger]:
		router.pointer.Move(state, finger, x, y, router.pixelsPerLayoutPixel)
	}
}
