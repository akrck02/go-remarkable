// Package canvas composes a frame on the processor and copies to the panel only what changed.
package canvas

import (
	"encoding/binary"
	"image"
	"math"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/opengl/overlay"
)

// The sizes of a pixel on the canvas and in the panel's memory.
const (
	BytesPerRGBPixel    = 3
	BytesPerScreenPixel = 2
	MaximumChannel      = 255
	halfAPixel          = 0.5 // from a pixel's corner to its centre, and how far an edge is softened
	roundingToNearest   = 0.5
)

// LinePoint is one end of a line, in canvas pixels, and how wide the line is there.
type LinePoint struct {
	X, Y          float32
	WidthInPixels float32
}

// Rectangle is an area of pixels; right and bottom are one past the last pixel.
type Rectangle struct {
	Left, Top, Right, Bottom int
}

// IsEmpty says whether a Rectangle holds no pixels.
func (area Rectangle) IsEmpty() bool {
	return area.Right <= area.Left || area.Bottom <= area.Top
}

// Joined is the smallest Rectangle that holds both.
func (area Rectangle) Joined(other Rectangle) Rectangle {
	if area.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return area
	}
	return Rectangle{
		Left: min(area.Left, other.Left), Top: min(area.Top, other.Top),
		Right: max(area.Right, other.Right), Bottom: max(area.Bottom, other.Bottom),
	}
}

// Within is the part of a Rectangle inside another.
func (area Rectangle) Within(limit Rectangle) Rectangle {
	return Rectangle{
		Left: max(area.Left, limit.Left), Top: max(area.Top, limit.Top),
		Right: min(area.Right, limit.Right), Bottom: min(area.Bottom, limit.Bottom),
	}
}

// Canvas is the frame being composed on the processor, always upright, as the interface sees it.
// It stays private to the program, so the panel only ever receives finished pixels and only the
// ones that changed; orientation is applied on the way there.
type Canvas struct {
	Width, Height           int
	PanelWidth, PanelHeight int
	Turn                    Orientation
	Pixels                  []byte // red, green and blue of each pixel
}

// New makes an upright canvas the size of a panel, in pixels.
func New(width, height int) *Canvas {
	return &Canvas{
		Width: width, Height: height, PanelWidth: width, PanelHeight: height,
		Pixels: make([]byte, width*height*BytesPerRGBPixel),
	}
}

// TurnTo changes the orientation; the canvas takes the size the panel has that way round.
func (target *Canvas) TurnTo(turn Orientation) {
	target.Turn = turn
	target.Width, target.Height = turn.Size(target.PanelWidth, target.PanelHeight)
}

// Whole is the area of the entire canvas.
func (target *Canvas) Whole() Rectangle {
	return Rectangle{Right: target.Width, Bottom: target.Height}
}

// Clear fills the canvas with one colour.
func (target *Canvas) Clear(colour draw2d.Color) {
	red, green, blue := toByte(colour[0]), toByte(colour[1]), toByte(colour[2])
	for offset := 0; offset < len(target.Pixels); offset += BytesPerRGBPixel {
		target.Pixels[offset], target.Pixels[offset+1], target.Pixels[offset+2] = red, green, blue
	}
}

// Paint blends one Rectangle of the painter's atlas onto the canvas: the atlas gives the coverage
// and the quad the colour. It is what a GPU's fragment shader does for the other backends. Pictures
// and frosted backdrops are left out, as paper interfaces have neither.
func (target *Canvas) Paint(quad overlay.Quad, atlas *image.Alpha) {
	if quad.Pictured || quad.Backdrop {
		return
	}
	area := target.covered(quad)
	horizontalRatio := (quad.Atlas[2] - quad.Atlas[0]) / (quad.Screen[2] - quad.Screen[0])
	verticalRatio := (quad.Atlas[3] - quad.Atlas[1]) / (quad.Screen[3] - quad.Screen[1])
	for y := area.Top; y < area.Bottom; y++ {
		atlasY := int(quad.Atlas[1] + (float32(y)+halfAPixel-quad.Screen[1])*verticalRatio)
		for x := area.Left; x < area.Right; x++ {
			atlasX := int(quad.Atlas[0] + (float32(x)+halfAPixel-quad.Screen[0])*horizontalRatio)
			coverage := float32(atlas.AlphaAt(atlasX, atlasY).A) / MaximumChannel
			target.blend(x, y, quad.Color, coverage*quad.Color[3])
		}
	}
}

// covered is the part of the canvas a quad touches.
func (target *Canvas) covered(quad overlay.Quad) Rectangle {
	return Rectangle{
		Left:   max(int(math.Floor(float64(quad.Screen[0]))), 0),
		Top:    max(int(math.Floor(float64(quad.Screen[1]))), 0),
		Right:  min(int(math.Ceil(float64(quad.Screen[2]))), target.Width),
		Bottom: min(int(math.Ceil(float64(quad.Screen[3]))), target.Height),
	}
}

// blend mixes a colour into one pixel, by a share from 0 (nothing) to 1 (all of it).
func (target *Canvas) blend(x, y int, colour [4]float32, share float32) {
	if share <= 0 {
		return
	}
	offset := (y*target.Width + x) * BytesPerRGBPixel
	for channel := range BytesPerRGBPixel {
		current := float32(target.Pixels[offset+channel]) / MaximumChannel
		target.Pixels[offset+channel] = toByte(current + (colour[channel]-current)*share)
	}
}

// CopyChanged writes to a panel's RGB565 memory only the pixels that differ from what it holds,
// and gives the Rectangle around them, in panel pixels. It says false when the panel already
// shows this frame.
func (target *Canvas) CopyChanged(memory []byte) (Rectangle, bool) {
	return target.CopyChangedIn(memory, target.Whole())
}

// CopyChangedIn is CopyChanged for one area of the canvas, which is what a pen stroke can afford.
func (target *Canvas) CopyChangedIn(memory []byte, area Rectangle) (Rectangle, bool) {
	area = area.Within(target.Whole())
	origin, stepX, stepY := target.Turn.panelSteps(target.PanelWidth, target.PanelHeight)
	changed := Rectangle{Left: target.Width, Top: target.Height}
	for y := area.Top; y < area.Bottom; y++ {
		for x := area.Left; x < area.Right; x++ {
			pixelOffset := (y*target.Width + x) * BytesPerRGBPixel
			wanted := toRGB565(target.Pixels[pixelOffset], target.Pixels[pixelOffset+1], target.Pixels[pixelOffset+2])
			memoryOffset := (origin + x*stepX + y*stepY) * BytesPerScreenPixel
			if binary.LittleEndian.Uint16(memory[memoryOffset:]) == wanted {
				continue
			}
			binary.LittleEndian.PutUint16(memory[memoryOffset:], wanted)
			changed.Left, changed.Top = min(changed.Left, x), min(changed.Top, y)
			changed.Right, changed.Bottom = max(changed.Right, x+1), max(changed.Bottom, y+1)
		}
	}
	if changed.Right == 0 {
		return Rectangle{}, false
	}
	return target.PanelArea(changed), true
}

// PanelArea is where an area of the canvas lies on the panel.
func (target *Canvas) PanelArea(area Rectangle) Rectangle {
	firstX, firstY := target.Turn.ToPanel(area.Left, area.Top, target.PanelWidth, target.PanelHeight)
	lastX, lastY := target.Turn.ToPanel(area.Right-1, area.Bottom-1, target.PanelWidth, target.PanelHeight)
	return Rectangle{
		Left: min(firstX, lastX), Top: min(firstY, lastY),
		Right: max(firstX, lastX) + 1, Bottom: max(firstY, lastY) + 1,
	}
}

// PaintLine draws a line with round ends, softened on its edges, kept inside limit. It gives the
// area it touched.
func (target *Canvas) PaintLine(from, to LinePoint, colour [4]float32, limit Rectangle) Rectangle {
	radius := max(from.WidthInPixels, to.WidthInPixels) / 2
	reach := int(math.Ceil(float64(radius))) + 1
	area := Rectangle{
		Left: int(min(from.X, to.X)) - reach, Top: int(min(from.Y, to.Y)) - reach,
		Right: int(max(from.X, to.X)) + reach + 1, Bottom: int(max(from.Y, to.Y)) + reach + 1,
	}.Within(limit).Within(target.Whole())
	for y := area.Top; y < area.Bottom; y++ {
		for x := area.Left; x < area.Right; x++ {
			distance := distanceToSegment(float32(x)+halfAPixel, float32(y)+halfAPixel, from, to)
			target.blend(x, y, colour, min(radius-distance+halfAPixel, 1))
		}
	}
	return area
}

// distanceToSegment is how far a point is from the nearest point of a line between two others.
func distanceToSegment(x, y float32, from, to LinePoint) float32 {
	alongX, alongY := to.X-from.X, to.Y-from.Y
	lengthSquared := alongX*alongX + alongY*alongY
	share := float32(0)
	if lengthSquared > 0 {
		share = min(max(((x-from.X)*alongX+(y-from.Y)*alongY)/lengthSquared, 0), 1)
	}
	return float32(math.Hypot(float64(x-from.X-share*alongX), float64(y-from.Y-share*alongY)))
}

// toByte turns a channel from 0–1 into 0–255.
func toByte(channel float32) byte {
	return byte(min(max(channel, 0), 1)*MaximumChannel + roundingToNearest)
}

// toRGB565 packs a colour into the screen's sixteen bits.
func toRGB565(red, green, blue byte) uint16 {
	return uint16(red>>3)<<11 | uint16(green>>2)<<5 | uint16(blue>>3)
}
