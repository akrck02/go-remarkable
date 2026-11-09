// Package paper is how an interface should look on e-ink: high contrast, nothing that moves or
// changes by itself, and no extra refresh for pressing a button.
package paper

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

const (
	textSize                    = 18
	headingSize                 = 24
	controlHeightInLayoutPixels = 44
	buttonPaddingInLayoutPixels = 20
	lineWidthInLayoutPixels     = 1
)

// Theme is high contrast with no shadows: colour is muted on this panel, so dark against light has
// to carry everything.
func Theme() ui.Theme {
	theme := ui.LightTheme()
	black, white := draw2d.MustHex("#000000"), draw2d.MustHex("#ffffff")
	theme.Background, theme.Surface, theme.Sunken = white, white, white
	theme.Ink, theme.InkSoft, theme.InkQuiet = black, black, black
	theme.Line, theme.ControlLine, theme.FieldLine = black, black, black
	theme.Primary, theme.OnPrimary, theme.PrimaryTint = black, white, white
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
