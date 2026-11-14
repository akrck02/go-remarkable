package canvas

import (
	"bytes"
	"math/rand/v2"
	"testing"
)

// fullCopy is the panel's memory as a copy of every pixel would leave it, to check the fast one.
func fullCopy(frame *Canvas) []byte {
	memory := make([]byte, frame.PanelWidth*frame.PanelHeight*BytesPerScreenPixel)
	fresh := *frame
	fresh.sent, fresh.sentIsWhole = nil, false
	fresh.CopyChanged(memory)
	return memory
}

func TestSkippingUnchangedRowsLeavesTheSameMemoryAsCopyingEverything(test *testing.T) {
	random := rand.New(rand.NewPCG(1, 2))
	for _, turn := range []Orientation{Upright, TurnedRight} {
		frame := New(40, 30)
		frame.TurnTo(turn)
		memory := make([]byte, 40*30*BytesPerScreenPixel)
		for step := range 200 {
			for range random.IntN(5) {
				frame.blend(random.IntN(frame.Width), random.IntN(frame.Height), [4]float32{random.Float32(), random.Float32(), random.Float32(), 1}, 1)
			}
			if step%3 == 0 {
				frame.CopyChangedIn(memory, Rectangle{Left: 5, Top: 5, Right: 20, Bottom: 15})
			}
			frame.CopyChanged(memory)
			if !bytes.Equal(memory, fullCopy(frame)) {
				test.Fatalf("turn %v, step %d: the memory differs from a full copy", turn, step)
			}
		}
	}
}

func TestAChangeInOneRowReportsOnlyThatRow(test *testing.T) {
	frame := New(10, 10)
	memory := make([]byte, 10*10*BytesPerScreenPixel)
	frame.CopyChanged(memory)
	frame.blend(7, 2, [4]float32{1, 1, 1, 1}, 1)
	if changed, found := frame.CopyChanged(memory); !found || changed != (Rectangle{Left: 7, Top: 2, Right: 8, Bottom: 3}) {
		test.Fatalf("got %+v, %v", changed, found)
	}
}

func TestATurnCopiesEveryPixelAgain(test *testing.T) {
	frame := New(4, 6)
	memory := make([]byte, 4*6*BytesPerScreenPixel)
	frame.blend(0, 0, [4]float32{1, 1, 1, 1}, 1)
	frame.CopyChanged(memory)
	frame.TurnTo(TurnedRight)
	frame.blend(0, 0, [4]float32{1, 1, 1, 1}, 1)
	frame.CopyChanged(memory)
	if !bytes.Equal(memory, fullCopy(frame)) {
		test.Fatal("pixels kept their old place after the turn")
	}
}

// BenchmarkCopyingAFrameWithOneChangedLine is a keystroke: one line of text changes on a panel the
// size of the Paper Pro's.
func BenchmarkCopyingAFrameWithOneChangedLine(benchmark *testing.B) {
	const width, height, lineHeight = 1620, 2160, 60
	frame := New(width, height)
	memory := make([]byte, width*height*BytesPerScreenPixel)
	frame.CopyChanged(memory)
	shade := float32(0)
	for benchmark.Loop() {
		shade = 1 - shade
		for y := 1000; y < 1000+lineHeight; y++ {
			frame.blend(width/2, y, [4]float32{shade, shade, shade, 1}, 1)
		}
		frame.CopyChanged(memory)
	}
}
