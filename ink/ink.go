// Package ink is paper to write on with the pen: lines that follow it with no wait, and a clean
// refresh once it rests.
package ink

import (
	"log/slog"
	"time"

	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/canvas"
	"github.com/akrck02/go-remarkable/pen"
	"github.com/akrck02/go-remarkable/pointer"
	"github.com/akrck02/go-remarkable/qtfb"
)

const (
	// PaperName is the name of the control that is the paper.
	PaperName = "Paper"

	paperBorderInLayoutPixels = 1

	plainStrokeWidthInPixels    = 4
	lightestStrokeWidthInPixels = 2
	pressureExtraWidthInPixels  = 8
	tiltExtraWidthInPixels      = 6
	EraserWidthInPixels         = 40

	// CleanDelay is how long the pen rests before a clean refresh. It is longer than the second
	// a change of mode costs, so writing is never held up by its own clean-up.
	CleanDelay = 2 * time.Second
)

var (
	inkColour   = [4]float32{0, 0, 0, 1}
	paperColour = [4]float32{1, 1, 1, 1}

	// eraserMarkColour is dark enough to stay visible even if the fastest refresh mode has no greys.
	eraserMarkColour = [4]float32{0.35, 0.35, 0.35, 1}
)

// stroke is one line drawn without lifting the pen. An eraser's stroke is a stroke of paper.
type stroke struct {
	points []canvas.LinePoint
	erases bool
}

// Surface is paper to write on. It takes the pen from the interface: rebuilding the interface for
// each point would be far too slow for handwriting.
type Surface struct {
	frame                *canvas.Canvas
	pixelsPerLayoutPixel float32
	Sample               func() pen.Sample // what the pen is doing; tests replace it
	strokes              []stroke
	paper                canvas.Rectangle // where ink can go, in canvas pixels
	writing              bool
	pending              canvas.Rectangle // painted on the canvas, not yet on the panel
	finished             canvas.Rectangle // a stroke just ended here and wants its clean refresh
	cleared              bool
	Clock                func() time.Time // tests replace it
	lastPen              time.Time        // when the pen last sent anything during a stroke
	liftedAt             time.Time        // when the last stroke ended
}

// New makes an empty surface that paints on a canvas.
func New(frame *canvas.Canvas, sample func() pen.Sample, pixelsPerLayoutPixel float32) *Surface {
	return &Surface{frame: frame, Sample: sample, Clock: time.Now, pixelsPerLayoutPixel: pixelsPerLayoutPixel}
}

// strokeWidth is how wide the pen draws: wider when pressed harder and when leaned over.
func strokeWidth(sample pen.Sample) float32 {
	if !sample.Readable {
		return plainStrokeWidthInPixels
	}
	if sample.Erasing {
		return EraserWidthInPixels
	}
	return lightestStrokeWidthInPixels + sample.Pressure*pressureExtraWidthInPixels + sample.Tilt*tiltExtraWidthInPixels
}

// DrawPaper lays out the paper: where the pen writes. It grows to fill the space it is given.
func (surface *Surface) DrawPaper(context *ui.Context) {
	look := ui.Look{Border: context.Theme.ControlLine, BorderWidth: paperBorderInLayoutPixels}
	context.Clickable(PaperName, ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, ui.Looks{Normal: look, Hovered: look, Pressed: look}, func() {})
	if area, found := context.Area(PaperName); found {
		surface.paper = canvas.Rectangle{
			Left: int(area.X * surface.pixelsPerLayoutPixel), Top: int(area.Y * surface.pixelsPerLayoutPixel),
			Right: int((area.X + area.Width) * surface.pixelsPerLayoutPixel), Bottom: int((area.Y + area.Height) * surface.pixelsPerLayoutPixel),
		}
	}
}

// Clear removes every stroke.
func (surface *Surface) Clear() {
	surface.Stop()
	surface.strokes, surface.cleared, surface.finished = nil, true, canvas.Rectangle{}
}

// ReadsPenDetails says whether pressure, tilt and the eraser are reaching the strokes.
func (surface *Surface) ReadsPenDetails() bool { return surface.Sample().Readable }

// Writing says whether a stroke is being made now.
func (surface *Surface) Writing() bool { return surface.writing }

// StrokeCount says how many strokes there are.
func (surface *Surface) StrokeCount() int { return len(surface.strokes) }

// Receive handles an event if it is the pen writing on the paper, or a palm while it writes, and
// says whether it did.
func (surface *Surface) Receive(event qtfb.Event) bool {
	switch event.Kind {
	case qtfb.PenPress:
		surface.Stop() // a release that never came
		if !surface.onPaper(event) {
			return false
		}
		surface.begin(event)
	case qtfb.PenMove:
		if !surface.writing {
			return false
		}
		surface.extend(event)
	case qtfb.PenRelease:
		if !surface.writing {
			return false
		}
		surface.end()
	case qtfb.TouchPress, qtfb.TouchMove, qtfb.TouchRelease:
		if surface.Clock().Sub(surface.lastPen) >= pointer.LostPressSilence {
			surface.Stop() // the pen's release was lost, and fingers must not stay dead for it
		}
		return surface.writing
	default:
		return false
	}
	return true
}

func (surface *Surface) onPaper(event qtfb.Event) bool {
	x, y := int(event.X), int(event.Y)
	return x >= surface.paper.Left && x < surface.paper.Right && y >= surface.paper.Top && y < surface.paper.Bottom
}

func (surface *Surface) begin(event qtfb.Event) {
	sample := surface.Sample()
	surface.writing = true
	surface.lastPen = surface.Clock()
	point := canvas.LinePoint{X: float32(event.X), Y: float32(event.Y), WidthInPixels: strokeWidth(sample)}
	surface.strokes = append(surface.strokes, stroke{points: []canvas.LinePoint{point}, erases: sample.Erasing})
	surface.paint(point, point)
}

func (surface *Surface) extend(event qtfb.Event) {
	surface.lastPen = surface.Clock()
	stroke := &surface.strokes[len(surface.strokes)-1]
	previous := stroke.points[len(stroke.points)-1]
	point := canvas.LinePoint{X: float32(event.X), Y: float32(event.Y), WidthInPixels: strokeWidth(surface.Sample())}
	stroke.points = append(stroke.points, point)
	surface.paint(previous, point)
}

func (surface *Surface) end() {
	surface.writing = false
	surface.liftedAt = surface.Clock()
	stroke := surface.strokes[len(surface.strokes)-1]
	if stroke.erases {
		surface.remember(surface.paintStroke(stroke))
	}
	slog.Info("stroke", "points", len(stroke.points), "erases", stroke.erases, "sample", surface.Sample())
}

// Stop ends the stroke in progress, if there is one.
func (surface *Surface) Stop() {
	if surface.writing {
		surface.end()
	}
}

// Reorient moves every stroke to where it stays on the panel once the canvas is turned.
func (surface *Surface) Reorient(from, to canvas.Orientation) {
	width, height := surface.frame.PanelWidth, surface.frame.PanelHeight
	surface.pending, surface.finished = canvas.Rectangle{}, canvas.Rectangle{} // the turn repaints everything anyway
	for _, stroke := range surface.strokes {
		for index := range stroke.points {
			point := &stroke.points[index]
			panelX, panelY := from.ToPanel(int(point.X), int(point.Y), width, height)
			x, y := to.ToCanvas(int32(panelX), int32(panelY), width, height)
			point.X, point.Y = float32(x), float32(y)
		}
	}
}

// paint draws one segment of the current stroke and remembers the area for the next refresh. An
// eraser marks the whole width it is about to remove, and removes the ink when it lifts: painting
// paper as it went would wipe its own mark with each step.
func (surface *Surface) paint(from, to canvas.LinePoint) {
	colour := inkColour
	if surface.strokes[len(surface.strokes)-1].erases {
		colour = eraserMarkColour
	}
	surface.remember(surface.frame.PaintLine(from, to, colour, surface.paper))
}

// remember notes an area painted on the canvas, for the next refresh and for the clean one.
func (surface *Surface) remember(area canvas.Rectangle) {
	surface.pending = surface.pending.Joined(area)
	surface.finished = surface.finished.Joined(area)
}

func colourOf(stroke stroke) [4]float32 {
	if stroke.erases {
		return paperColour
	}
	return inkColour
}

// PaintAll draws every stroke again, over a freshly painted interface.
func (surface *Surface) PaintAll() {
	for _, stroke := range surface.strokes {
		surface.paintStroke(stroke)
	}
}

// paintStroke draws a whole stroke and says which area it covered.
func (surface *Surface) paintStroke(stroke stroke) canvas.Rectangle {
	var area canvas.Rectangle
	for index := range stroke.points {
		from := stroke.points[max(index-1, 0)]
		area = area.Joined(surface.frame.PaintLine(from, stroke.points[index], colourOf(stroke), surface.paper))
	}
	return area
}

// TakePending gives the area painted since the last refresh, and forgets it.
func (surface *Surface) TakePending() canvas.Rectangle {
	area := surface.pending
	surface.pending = canvas.Rectangle{}
	return area
}

// TimeUntilClean says how long until the strokes want their clean refresh, if they do.
func (surface *Surface) TimeUntilClean() (time.Duration, bool) {
	if surface.writing || surface.finished.IsEmpty() {
		return 0, false
	}
	return max(CleanDelay-surface.Clock().Sub(surface.liftedAt), 0), true
}

// TakeFinished gives, once, the area written since the last clean refresh, when the pen has rested.
func (surface *Surface) TakeFinished() canvas.Rectangle {
	if wait, due := surface.TimeUntilClean(); !due || wait > 0 {
		return canvas.Rectangle{}
	}
	area := surface.finished
	surface.finished = canvas.Rectangle{}
	return area
}

// TakeCleared says once that the strokes were removed, as the interface alone cannot tell.
func (surface *Surface) TakeCleared() bool {
	cleared := surface.cleared
	surface.cleared = false
	return cleared
}
