package canvas

import (
	"encoding/binary"
	"image"
	"testing"

	"github.com/akrck02/go-renderer/opengl/overlay"
)

func TestOnlyThePixelsThatChangedAreCopiedAndReported(test *testing.T) {
	frame := New(10, 10)
	memory := make([]byte, 10*10*BytesPerScreenPixel)
	if _, any := frame.CopyChanged(memory); any {
		test.Fatal("a black frame on a black screen reported a change")
	}
	frame.blend(3, 4, [4]float32{1, 1, 1, 1}, 1)
	frame.blend(5, 6, [4]float32{1, 1, 1, 1}, 1)
	changed, any := frame.CopyChanged(memory)
	if !any || changed != (Rectangle{Left: 3, Top: 4, Right: 6, Bottom: 7}) {
		test.Fatalf("got %+v, %v", changed, any)
	}
	if binary.LittleEndian.Uint16(memory[(4*10+3)*BytesPerScreenPixel:]) != 0xFFFF {
		test.Error("the changed pixel did not reach the screen")
	}
}

func TestAQuadPaintsItsAtlasCoverageInItsColour(test *testing.T) {
	atlas := image.NewAlpha(image.Rect(0, 0, 2, 1))
	atlas.Pix[0], atlas.Pix[1] = 255, 0
	frame := New(2, 1)
	frame.Paint(overlay.Quad{Screen: [4]float32{0, 0, 2, 1}, Atlas: [4]float32{0, 0, 2, 1}, Color: [4]float32{1, 0, 0, 1}}, atlas)
	if frame.Pixels[0] != 255 {
		test.Errorf("covered pixel has red %d", frame.Pixels[0])
	}
	if frame.Pixels[3] != 0 {
		test.Errorf("uncovered pixel has red %d", frame.Pixels[3])
	}
}

func TestAQuadOutsideTheCanvasPaintsNothingAndDoesNotCrash(test *testing.T) {
	atlas := image.NewAlpha(image.Rect(0, 0, 4, 4))
	frame := New(2, 2)
	frame.Paint(overlay.Quad{Screen: [4]float32{-2, -2, 9, 9}, Atlas: [4]float32{0, 0, 4, 4}, Color: [4]float32{1, 1, 1, 1}}, atlas)
}

func TestCopyingOneAreaLeavesTheRestOfThePanelAlone(test *testing.T) {
	frame := New(10, 10)
	memory := make([]byte, 10*10*BytesPerScreenPixel)
	frame.blend(1, 1, [4]float32{1, 1, 1, 1}, 1)
	frame.blend(8, 8, [4]float32{1, 1, 1, 1}, 1)
	changed, found := frame.CopyChangedIn(memory, Rectangle{Left: 0, Top: 0, Right: 3, Bottom: 3})
	if !found || changed != (Rectangle{Left: 1, Top: 1, Right: 2, Bottom: 2}) {
		test.Fatalf("got %+v, %v", changed, found)
	}
	if memory[(8*10+8)*BytesPerScreenPixel] != 0 {
		test.Error("a pixel outside the area was copied")
	}
}
