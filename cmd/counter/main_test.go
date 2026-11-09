package main

import (
	"testing"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/apptest"
)

func TestEachPressCountsAndCostsOneRefresh(test *testing.T) {
	shown := &counter{}
	application, target := apptest.Start(test, func(app.Device) app.Program {
		return app.Program{Draw: shown.draw}
	})
	apptest.Tap(test, application, addButtonName)
	if shown.count != 1 {
		test.Fatalf("count %d", shown.count)
	}
	if len(target.Refreshes) != 1 {
		test.Fatalf("%d refreshes", len(target.Refreshes))
	}
}
