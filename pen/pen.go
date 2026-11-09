// Package pen reads what a program cannot learn from AppLoad: the pen's pressure, tilt and eraser.
package pen

import (
	"encoding/binary"
	"log/slog"
	"math"
	"os"
	"sync/atomic"

	"github.com/akrck02/go-remarkable/failure"
)

// The Linux input event on a 64-bit system: sixteen bytes of time, then type, code and value. The
// codes are from linux/input-event-codes.h; the ranges were read from this tablet's pen.
const (
	// PaperProDevicePath is where a reMarkable Paper Pro's pen is.
	PaperProDevicePath = "/dev/input/event2"

	penEventBytes = 24

	penEventKey      uint16 = 1
	penEventAbsolute uint16 = 3

	penCodePressure   uint16 = 0x18
	penCodeTiltX      uint16 = 0x1a
	penCodeTiltY      uint16 = 0x1b
	penCodeToolPen    uint16 = 0x140
	penCodeToolEraser uint16 = 0x141
	penCodeTouch      uint16 = 0x14a

	penMaximumPressure = 4096
	penMaximumTilt     = 9000
)

// Sample is what the pen is doing now, beyond its position. Pressure and tilt run from 0 to 1;
// Readable says false when nothing could be learned about the pen.
type Sample struct {
	Readable bool
	Pressure float32
	Tilt     float32
	Erasing  bool
}

// Plain is the pen when its device cannot be read: all that is known is that it is there.
func Plain() Sample { return Sample{} }

// Device reads the pen's own input device, for what qtfb does not pass on. It only reads: the
// device is never grabbed, so the system's pen keeps working.
type Device struct {
	file     *os.File
	answered atomic.Bool // an event has arrived: opening it proves nothing while the system holds it
	stopped  atomic.Bool
	pressure atomic.Int32
	tiltX    atomic.Int32
	tiltY    atomic.Int32
	erasing  atomic.Bool
}

// Open opens the device at a path.
func Open(path string) (*Device, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, failure.Wrap(DeviceUnreadable, path, err)
	}
	return &Device{file: file}, nil
}

// Follow starts following the device at a path, and gives the function that asks it what
// the pen is doing. If the device cannot be opened, ink still works: it draws plain lines.
func Follow(path string) func() Sample {
	device, err := Open(path)
	if err != nil {
		slog.Warn("no pen details, ink draws plain lines", "error", err)
		return Plain
	}
	go device.follow()
	return device.Sample
}

// follow reads events until the device goes away.
func (device *Device) follow() {
	packet := make([]byte, penEventBytes)
	for {
		if _, err := device.file.Read(packet); err != nil {
			device.stopped.Store(true)
			slog.Warn("the pen device stopped", "error", err)
			return
		}
		kind, code, value := decodePenEvent(packet)
		device.applyEvent(kind, code, value)
	}
}

// decodePenEvent reads an event's type, code and value.
func decodePenEvent(packet []byte) (kind, code uint16, value int32) {
	return binary.LittleEndian.Uint16(packet[16:]), binary.LittleEndian.Uint16(packet[18:]), int32(binary.LittleEndian.Uint32(packet[20:]))
}

// applyEvent keeps what an event says. Touches and the pen coming into range are logged, to compare
// with what arrives through qtfb when input is lost.
func (device *Device) applyEvent(kind, code uint16, value int32) {
	device.answered.Store(true)
	switch {
	case kind == penEventAbsolute && code == penCodePressure:
		device.pressure.Store(value)
	case kind == penEventAbsolute && code == penCodeTiltX:
		device.tiltX.Store(value)
	case kind == penEventAbsolute && code == penCodeTiltY:
		device.tiltY.Store(value)
	case kind == penEventKey && code == penCodeToolEraser:
		device.erasing.Store(value != 0)
		slog.Info("pen device", "eraser in range", value != 0)
	case kind == penEventKey && code == penCodeToolPen:
		slog.Info("pen device", "pen in range", value != 0)
	case kind == penEventKey && code == penCodeTouch:
		slog.Info("pen device", "touching", value != 0)
	}
}

// Sample says what the pen is doing now. The lean is the two tilts together, whichever way it goes.
func (device *Device) Sample() Sample {
	lean := math.Hypot(float64(device.tiltX.Load()), float64(device.tiltY.Load()))
	return Sample{
		Readable: device.answered.Load() && !device.stopped.Load(),
		Pressure: float32(device.pressure.Load()) / penMaximumPressure,
		Tilt:     float32(min(lean/penMaximumTilt, 1)),
		Erasing:  device.erasing.Load(),
	}
}
