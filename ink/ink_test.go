package ink

import (
	"testing"

	"github.com/akrck02/go-remarkable/pen"
)

func TestHarderPressureDrawsAWiderLine(test *testing.T) {
	light := strokeWidth(pen.Sample{Readable: true, Pressure: 0.1})
	firm := strokeWidth(pen.Sample{Readable: true, Pressure: 0.9})
	if firm <= light {
		test.Fatalf("light %v, firm %v", light, firm)
	}
}
