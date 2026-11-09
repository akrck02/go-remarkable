package canvas

// Orientation is which way round the interface stands on the panel.
type Orientation int

// The four ways round.
const (
	Upright     Orientation = iota
	TurnedRight             // the canvas's top edge runs down the panel's right side
	UpsideDown
	TurnedLeft // the canvas's top edge runs up the panel's left side
	orientationCount
)

// Next is the orientation a quarter turn on.
func (turn Orientation) Next() Orientation {
	return (turn + 1) % orientationCount
}

// Size is how wide and tall the canvas is, for a panel of this size.
func (turn Orientation) Size(PanelWidth, PanelHeight int) (width, height int) {
	if turn == TurnedRight || turn == TurnedLeft {
		return PanelHeight, PanelWidth
	}
	return PanelWidth, PanelHeight
}

// ToCanvas turns a point of the panel, as input events give it, into a point of the canvas.
func (turn Orientation) ToCanvas(x, y int32, PanelWidth, PanelHeight int) (canvasX, canvasY int32) {
	switch turn {
	case TurnedRight:
		return y, int32(PanelWidth) - 1 - x
	case UpsideDown:
		return int32(PanelWidth) - 1 - x, int32(PanelHeight) - 1 - y
	case TurnedLeft:
		return int32(PanelHeight) - 1 - y, x
	}
	return x, y
}

// ToPanel turns a point of the canvas into a point of the panel.
func (turn Orientation) ToPanel(x, y, PanelWidth, PanelHeight int) (panelX, panelY int) {
	switch turn {
	case TurnedRight:
		return PanelWidth - 1 - y, x
	case UpsideDown:
		return PanelWidth - 1 - x, PanelHeight - 1 - y
	case TurnedLeft:
		return y, PanelHeight - 1 - x
	}
	return x, y
}

// panelSteps says where canvas pixels live in the panel's memory, counted in pixels: pixel (x, y)
// is at origin + x*stepX + y*stepY. One sum per pixel keeps the copy of a full frame cheap.
func (turn Orientation) panelSteps(PanelWidth, PanelHeight int) (origin, stepX, stepY int) {
	switch turn {
	case TurnedRight:
		return PanelWidth - 1, PanelWidth, -1
	case UpsideDown:
		return PanelHeight*PanelWidth - 1, -1, -PanelWidth
	case TurnedLeft:
		return (PanelHeight - 1) * PanelWidth, -PanelWidth, 1
	}
	return 0, 1, PanelWidth
}
