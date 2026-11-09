package canvas

import (
	"testing"
)

func TestACornerOfTheScreenMapsToTheRightLayoutPointInEachOrientation(test *testing.T) {
	const panelWidth, panelHeight = 30, 40
	cases := []struct {
		name                   string
		turn                   Orientation
		panelX, panelY         int32
		canvasX, canvasY       int32
		canvasWide, canvasTall int
	}{
		{"Upright", Upright, 0, 0, 0, 0, 30, 40},
		{"turned right", TurnedRight, 29, 0, 0, 0, 40, 30},
		{"upside down", UpsideDown, 29, 39, 0, 0, 30, 40},
		{"turned left", TurnedLeft, 0, 39, 0, 0, 40, 30},
		{"turned right, far corner", TurnedRight, 0, 39, 39, 29, 40, 30},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			x, y := testCase.turn.ToCanvas(testCase.panelX, testCase.panelY, panelWidth, panelHeight)
			if x != testCase.canvasX || y != testCase.canvasY {
				test.Errorf("got %d,%d", x, y)
			}
			wide, tall := testCase.turn.Size(panelWidth, panelHeight)
			if wide != testCase.canvasWide || tall != testCase.canvasTall {
				test.Errorf("canvas is %dx%d", wide, tall)
			}
		})
	}
}

func TestAPixelReachesThePanelWhereTheOrientationSays(test *testing.T) {
	const panelWidth, panelHeight = 3, 4
	for _, turn := range []Orientation{Upright, TurnedRight, UpsideDown, TurnedLeft} {
		frame := New(panelWidth, panelHeight)
		frame.TurnTo(turn)
		memory := make([]byte, panelWidth*panelHeight*BytesPerScreenPixel)
		frame.blend(1, 2, [4]float32{1, 1, 1, 1}, 1)
		changed, found := frame.CopyChanged(memory)
		if !found || changed.Right-changed.Left != 1 || changed.Bottom-changed.Top != 1 {
			test.Fatalf("%v: changed %+v", turn, changed)
		}
		x, y := turn.ToCanvas(int32(changed.Left), int32(changed.Top), panelWidth, panelHeight)
		if x != 1 || y != 2 {
			test.Errorf("%v: the pixel came back at %d,%d", turn, x, y)
		}
		offset := (changed.Top*panelWidth + changed.Left) * BytesPerScreenPixel
		if memory[offset] != 0xFF {
			test.Errorf("%v: the pixel is not in the panel's memory", turn)
		}
	}
}

const (
	panelWidthForTests  = 1620
	panelHeightForTests = 2160
)
