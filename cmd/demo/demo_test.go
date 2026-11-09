package main

import (
	"testing"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/apptest"
	"github.com/akrck02/go-remarkable/qtfb"
	ui "github.com/akrck02/go-ui"
)

// shownDemo is the whole program, with only the connection to AppLoad replaced.
type shownDemo struct {
	*app.Application
	demo   *demo
	builds int // how many times the interface was built
}

func newShownApplication(test *testing.T) (*shownDemo, *apptest.Display) {
	test.Helper()
	shown := &shownDemo{}
	application, target := apptest.Start(test, func(device app.Device) app.Program {
		shown.demo = newDemo(device)
		program := programOf(shown.demo)
		draw := program.Draw
		program.Draw = func(context *ui.Context) {
			shown.builds++
			draw(context)
		}
		return program
	})
	shown.Application = application
	return shown, target
}

func send(test *testing.T, shown *shownDemo, events ...qtfb.Event) {
	test.Helper()
	apptest.Send(test, shown.Application, events...)
}

func sendTogether(test *testing.T, shown *shownDemo, events ...qtfb.Event) {
	test.Helper()
	apptest.SendTogether(test, shown.Application, events...)
}

func middleOf(test *testing.T, shown *shownDemo, path ...string) (x, y int32) {
	test.Helper()
	return apptest.MiddleOf(test, shown.Application, path...)
}

func tap(test *testing.T, shown *shownDemo, path ...string) {
	test.Helper()
	apptest.Tap(test, shown.Application, path...)
}

func openPage(test *testing.T, shown *shownDemo, name string, target *apptest.Display) {
	test.Helper()
	tap(test, shown, pagesTabsName, name)
	target.Forget()
}

func TestATapReachesTheScreenAsOneRefresh(test *testing.T) {
	shown, target := newShownApplication(test)
	tap(test, shown, "Next")
	if len(target.Refreshes) != 1 {
		test.Fatalf("a tap sent %d refreshes", len(target.Refreshes))
	}
}

func TestTappingAScreensNameShowsThatScreen(test *testing.T) {
	shown, target := newShownApplication(test)
	openPage(test, shown, "Buttons", target)
	if _, found := shown.Context().Area(tapButtonName); !found {
		test.Fatal("the Buttons screen is not showing")
	}
}

func TestPressingAButtonChangesNothingOnTheScreen(test *testing.T) {
	shown, target := newShownApplication(test)
	x, y := middleOf(test, shown, "Next")
	send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: x, Y: y})
	if len(target.Refreshes) != 0 {
		test.Fatalf("pressing sent %d refreshes", len(target.Refreshes))
	}
}

func TestNothingIsPaintedWhenNothingChanged(test *testing.T) {
	shown, target := newShownApplication(test)
	for _, name := range pageNames {
		openPage(test, shown, name, target)
		if err := shown.Update(); err != nil {
			test.Fatal(err)
		}
		if len(target.Refreshes) != 0 {
			test.Errorf("%s repainted with nothing new", name)
		}
	}
}

func TestEachScreenPicksItsOwnRefreshMode(test *testing.T) {
	shown, target := newShownApplication(test)
	x, y := middleOf(test, shown, pagesTabsName, "Gestures")
	send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: x, Y: y}, qtfb.Event{Kind: qtfb.TouchRelease, X: x, Y: y})
	if len(target.Modes) != 1 || target.Modes[0] != qtfb.Fast {
		test.Fatalf("got modes %v", target.Modes)
	}
}

func TestEveryKindOfControlCostsOneRefresh(test *testing.T) {
	cases := []struct {
		name string
		page string
		path []string
	}{
		{"a button that turns a page", "Reading", []string{"Next"}},
		{"a counting button", "Buttons", []string{tapButtonName}},
		{"a switch", "Buttons", []string{switchName}},
		{"a tab", "Reading", []string{pagesTabsName, "Monitor"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			shown, target := newShownApplication(test)
			openPage(test, shown, testCase.page, target)
			tap(test, shown, testCase.path...)
			if len(target.Refreshes) != 1 {
				test.Fatalf("got %d refreshes", len(target.Refreshes))
			}
		})
	}
}
