package app

import (
	"testing"
	"time"

	"github.com/akrck02/go-remarkable/canvas"
	"github.com/akrck02/go-remarkable/pen"
	"github.com/akrck02/go-remarkable/qtfb"
	ui "github.com/akrck02/go-ui"
)

// fastApplication runs a program that draws a word in UltraFast, on a clock the test moves.
func fastApplication(test *testing.T, word *string) (*Application, *standInDisplay, *time.Time) {
	test.Helper()
	target := &standInDisplay{}
	memory := make([]byte, PanelWidthInPixels*PanelHeightInPixels*canvas.BytesPerScreenPixel)
	application, err := New(target, memory, pen.Plain, func(Device) Program {
		return Program{
			Draw:        func(context *ui.Context) { context.Text(*word, ui.TextStyle{}) },
			RefreshMode: func() qtfb.RefreshMode { return qtfb.UltraFast },
		}
	})
	if err != nil {
		test.Fatal(err)
	}
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	application.now = func() time.Time { return clock }
	return application, target, &clock
}

func TestUltraFastFramesAreRefreshedCleanlyOnceTheyRest(test *testing.T) {
	word := "a"
	application, target, clock := fastApplication(test, &word)
	if err := application.Update(); err != nil {
		test.Fatal(err)
	}
	word = "ab"
	if err := application.Update(); err != nil {
		test.Fatal(err)
	}
	fastFrames := target.refreshes
	*clock = clock.Add(FastCleanDelay / 2)
	if err := application.Answer(); err != nil || target.refreshes != fastFrames {
		test.Fatalf("cleaned before the pause: %d refreshes, %v", target.refreshes, err)
	}
	if delay, wanted := application.nextWake(); !wanted || delay != FastCleanDelay/2 {
		test.Fatalf("the loop would not wake for the clean refresh: %v, %v", delay, wanted)
	}
	*clock = clock.Add(FastCleanDelay / 2)
	if err := application.Answer(); err != nil {
		test.Fatal(err)
	}
	last := len(target.modes) - 1
	if target.refreshes != fastFrames+1 || target.modes[last] != qtfb.Content {
		test.Fatalf("no clean refresh after the pause: %d refreshes, modes %v", target.refreshes, target.modes)
	}
	if target.areas[last] != target.areas[0].Joined(target.areas[1]) {
		test.Fatalf("the clean refresh covered %+v, not what the fast frames touched", target.areas[last])
	}
	if err := application.Answer(); err != nil || target.refreshes != fastFrames+1 {
		test.Fatal("the area was cleaned twice")
	}
	*clock = clock.Add(InkRearmDelay)
	if err := application.Answer(); err != nil || target.mode != qtfb.UltraFast || target.refreshes != fastFrames+1 {
		test.Fatalf("the panel did not go back to UltraFast after the clean refresh: mode %v, %v", target.mode, err)
	}
	changes := target.modeChanges
	word = "abc"
	if err := application.Update(); err != nil || target.modeChanges != changes {
		test.Fatalf("the next key paid a change of mode: %d changes, %v", target.modeChanges-changes, err)
	}
}
