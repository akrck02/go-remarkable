package main

import (
	"testing"

	"github.com/akrck02/go-remarkable/app"
	"github.com/akrck02/go-remarkable/apptest"
	"github.com/akrck02/go-remarkable/canvas"
)

func darkPixels(application *app.Application) int {
	frame := application.Canvas()
	count := 0
	for offset := 0; offset < len(frame.Pixels); offset += canvas.BytesPerRGBPixel {
		if frame.Pixels[offset] < canvas.MaximumChannel/2 {
			count++
		}
	}
	return count
}

func TestTheFirstFrameHasWordsOnIt(test *testing.T) {
	application, _ := apptest.Start(test, newProgram)
	if darkPixels(application) == 0 {
		test.Fatal("the page is blank")
	}
}
