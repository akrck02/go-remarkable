package pen

import (
	"encoding/binary"
	"errors"
	"testing"
)

func penDeviceBytes(kind, code uint16, value int32) []byte {
	packet := make([]byte, penEventBytes)
	binary.LittleEndian.PutUint16(packet[16:], kind)
	binary.LittleEndian.PutUint16(packet[18:], code)
	binary.LittleEndian.PutUint32(packet[20:], uint32(value))
	return packet
}

func TestAPenEventIsDecodedFromItsBytes(test *testing.T) {
	kind, code, value := decodePenEvent(penDeviceBytes(penEventAbsolute, penCodeTiltX, -4500))
	if kind != penEventAbsolute || code != penCodeTiltX || value != -4500 {
		test.Fatalf("got %d, %d, %d", kind, code, value)
	}
}

func TestPenEventsFillTheSample(test *testing.T) {
	cases := []struct {
		name  string
		kind  uint16
		code  uint16
		value int32
		check func(Sample) bool
	}{
		{"pressure", penEventAbsolute, penCodePressure, penMaximumPressure / 2, func(sample Sample) bool { return sample.Pressure == 0.5 }},
		{"tilt sideways", penEventAbsolute, penCodeTiltX, penMaximumTilt, func(sample Sample) bool { return sample.Tilt == 1 }},
		{"tilt forwards", penEventAbsolute, penCodeTiltY, -penMaximumTilt, func(sample Sample) bool { return sample.Tilt == 1 }},
		{"eraser", penEventKey, penCodeToolEraser, 1, func(sample Sample) bool { return sample.Erasing }},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			device := &Device{}
			device.applyEvent(testCase.kind, testCase.code, testCase.value)
			sample := device.Sample()
			if !sample.Readable || !testCase.check(sample) {
				test.Fatalf("got %+v", sample)
			}
		})
	}
}

func TestADeviceIsNotReadableBeforeItsFirstEventOrAfterItStops(test *testing.T) {
	device := &Device{}
	if device.Sample().Readable {
		test.Fatal("readable before any event")
	}
	device.applyEvent(penEventAbsolute, penCodePressure, 1)
	if !device.Sample().Readable {
		test.Fatal("not readable after an event")
	}
	device.stopped.Store(true)
	if device.Sample().Readable {
		test.Fatal("readable after stopping")
	}
}

func TestAMissingPenDeviceGivesItsCode(test *testing.T) {
	if _, err := Open(test.TempDir() + "/missing"); !errors.Is(err, DeviceUnreadable) {
		test.Fatalf("got %v", err)
	}
}

func TestAPlainPenSaysItIsNotReadable(test *testing.T) {
	if Plain().Readable {
		test.Fatal("a plain pen claims details")
	}
}
