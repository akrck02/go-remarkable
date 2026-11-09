package main

import (
	"bytes"
	_ "embed" // the covers travel inside the program

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

const coverRadiusInLayoutPixels = 8

// The covers are prints in the public domain, cropped to the shape of a cover. The books are made
// up; the prints are real, and THIRD_PARTY_NOTICES.md says who made each one.
var (
	//go:embed covers/lighthouse.jpg
	lighthouseCoverFile []byte
	//go:embed covers/whirlpool.jpg
	whirlpoolCoverFile []byte
	//go:embed covers/knot.jpg
	knotCoverFile []byte
)

// mustDecodeCover reads a cover built into the program. One that cannot be read is a broken build,
// not something that can happen to a person, so it stops the program.
func mustDecodeCover(file []byte) *draw2d.Image {
	picture, err := draw2d.DecodeImage(bytes.NewReader(file))
	if err != nil {
		panic(err)
	}
	return picture
}

// drawCover shows a cover with soft corners.
func drawCover(context *ui.Context, layout ui.Layout, cover *draw2d.Image) {
	context.Image(cover, layout, coverRadiusInLayoutPixels)
}
