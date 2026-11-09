// Package paper is how an interface should look on e-ink: calm, with soft ink and thin lines,
// nothing that moves or changes by itself, and no extra refresh for pressing a button.
package paper

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

const (
	textSize                             = 18
	headingSize                          = 24
	controlHeightInLayoutPixels          = 44
	buttonPaddingInLayoutPixels          = 20
	lineWidthInLayoutPixels              = 1
	sectionLabelSize                     = 15
	cardRadiusInLayoutPixels             = 12
	progressHeightInLayoutPixels         = 4
	switchWidthInLayoutPixels            = 44
	switchHeightInLayoutPixels   float32 = 24
	switchPaddingInLayoutPixels  float32 = 3
)

// The colours of paper. The background stays pure white: a tinted one would be dithered on this
// panel and look grainy, so calm comes from soft ink and thin lines instead.
var (
	Ink      = draw2d.MustHex("#1b1a18")
	SoftInk  = draw2d.MustHex("#5f5b53")
	QuietInk = draw2d.MustHex("#8a857a")
	Line     = draw2d.MustHex("#d9d4c8")
	White    = draw2d.MustHex("#ffffff")

	// Accent is the one colour of an interface: muted, as this panel shows colour, and used small.
	Accent = draw2d.MustHex("#b8862f")
)

// Theme is soft ink on white with no shadows: nothing that a slow panel would have to fade.
func Theme() ui.Theme {
	theme := ui.LightTheme()
	theme.Background, theme.Surface, theme.Sunken = White, White, White
	// Go UI's controls change between these inks when touched, which on paper is an extra refresh,
	// so its theme has only one. SoftInk and QuietInk are for text a program colours itself.
	theme.Ink, theme.InkSoft, theme.InkQuiet = Ink, Ink, Ink
	theme.Line, theme.ControlLine, theme.FieldLine = Line, Line, Line
	theme.Primary, theme.OnPrimary, theme.PrimaryTint = Ink, White, White
	theme.Shadow = draw2d.Color{}
	theme.ShadowBlur, theme.ShadowDrop = 0, 0
	theme.TextSize, theme.HeadingSize, theme.ControlHeight = textSize, headingSize, controlHeightInLayoutPixels
	return theme
}

// Button is a button with one look. Go UI's own turns dark while held, which on e-ink is a whole
// extra refresh before the action happens. It says whether it was pressed.
func Button(context *ui.Context, label string) bool {
	look := ui.Look{Border: context.Theme.ControlLine, BorderWidth: lineWidthInLayoutPixels, Radius: context.Theme.ControlRadius}
	layout := ui.Layout{Height: ui.Fixed(context.Theme.ControlHeight), Padding: ui.PadXY(buttonPaddingInLayoutPixels, 0), AlignX: ui.Center, AlignY: ui.Center}
	return context.Clickable(label, layout, ui.Looks{Normal: look, Hovered: look, Pressed: look}, func() {
		context.Label(label)
	})
}

// PrimaryButton is the one button of a screen that matters most: filled, and still with one look.
func PrimaryButton(context *ui.Context, label string) bool {
	look := ui.Look{Fill: Ink, Radius: context.Theme.ControlRadius}
	layout := ui.Layout{Height: ui.Fixed(context.Theme.ControlHeight), Padding: ui.PadXY(buttonPaddingInLayoutPixels, 0), AlignX: ui.Center, AlignY: ui.Center}
	return context.Clickable(label, layout, ui.Looks{Normal: look, Hovered: look, Pressed: look}, func() {
		context.Text(label, ui.TextStyle{Color: White})
	})
}

// Switch is a label and an on-off switch, with one look: Go UI's own darkens while it is touched,
// which on paper is an extra refresh. It says whether it was flipped.
func Switch(context *ui.Context, label string, value *bool) bool {
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(context.Theme.ControlHeight), AlignY: ui.Center}
	none := ui.Look{}
	clicked := context.Clickable(label, layout, ui.Looks{Normal: none, Hovered: none, Pressed: none}, func() {
		context.Label(label)
		context.Spacer()
		drawSwitchTrack(context, *value)
	})
	if clicked {
		*value = !*value
	}
	return clicked
}

func drawSwitchTrack(context *ui.Context, on bool) {
	fill, align := Line, ui.Start
	if on {
		fill, align = Ink, ui.End
	}
	layout := ui.Layout{Width: ui.Fixed(switchWidthInLayoutPixels), Height: ui.Fixed(switchHeightInLayoutPixels), Padding: ui.Pad(switchPaddingInLayoutPixels), AlignX: align, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: fill, Radius: switchHeightInLayoutPixels / 2}, func() {
		knob := switchHeightInLayoutPixels - 2*switchPaddingInLayoutPixels
		context.Frame(ui.Layout{Width: ui.Fixed(knob), Height: ui.Fixed(knob)}, ui.Look{Fill: White, Radius: knob / 2}, func() {})
	})
}

// SectionLabel is a small soft label above a group of things.
func SectionLabel(context *ui.Context, text string) {
	context.Text(text, ui.TextStyle{Size: sectionLabelSize, Color: SoftInk})
}

// Card holds related things inside one thin line.
func Card(context *ui.Context, layout ui.Layout, content func()) {
	look := ui.Look{Border: Line, BorderWidth: lineWidthInLayoutPixels, Radius: cardRadiusInLayoutPixels}
	context.Frame(layout, look, content)
}

// ProgressBar shows how far something has gone, from 0 to 1, in one thin line.
func ProgressBar(context *ui.Context, share float32) {
	share = min(max(share, 0), 1)
	context.Paint(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(progressHeightInLayoutPixels)}, func(area draw2d.Rectangle, drawing *draw2d.List) {
		radius := area.Height / 2
		drawing.Rectangle(area, radius, Line, draw2d.Transparent, 0)
		done := area
		done.Width = area.Width * share
		drawing.Rectangle(done, radius, Accent, draw2d.Transparent, 0)
	})
}

// WrapText breaks text into lines of at most width characters, at the spaces.
func WrapText(text string, width int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(text) {
		if line != "" && len(line)+1+len(word) > width {
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
