package main

import (
	"testing"

	"github.com/akrck02/go-remarkable/qtfb"
)

func TestEveryModeOnOfferIsOneTheServerTakes(test *testing.T) {
	for _, mode := range refreshModes {
		if mode.value > qtfb.Interface {
			test.Errorf("%s would close the connection", mode.name)
		}
	}
}

func TestEachModeButtonSendsItsModeBeforeItsRefresh(test *testing.T) {
	for _, mode := range refreshModes {
		test.Run(mode.name, func(test *testing.T) {
			shown, target := newShownApplication(test)
			openPage(test, shown, "Screen", target)
			x, y := middleOf(test, shown, mode.name)
			send(test, shown, qtfb.Event{Kind: qtfb.TouchPress, X: x, Y: y}, qtfb.Event{Kind: qtfb.TouchRelease, X: x, Y: y})
			if len(target.Refreshes) != 1 {
				test.Fatalf("got %d refreshes", len(target.Refreshes))
			}
			if mode.value != qtfb.Content {
				if len(target.Calls) < 2 || target.Calls[0] != "mode" || target.Modes[0] != mode.value {
					test.Fatalf("calls %v, modes %v", target.Calls, target.Modes)
				}
			}
		})
	}
}
