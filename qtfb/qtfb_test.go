package qtfb

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/akrck02/go-remarkable/canvas"
)

func TestAnUpdateMessageCarriesItsRectangleAsPositionAndSize(test *testing.T) {
	message := updateMessage(canvas.Rectangle{Left: 10, Top: 20, Right: 40, Bottom: 60})
	for index, wanted := range []uint32{updatePartial, 10, 20, 30, 40} {
		if value := binary.LittleEndian.Uint32(message[4+4*index:]); value != wanted {
			test.Errorf("field %d = %d, want %d", index, value, wanted)
		}
	}
}

func TestAnAnswerOfAnotherSizeIsRefusedWithItsCode(test *testing.T) {
	if _, _, err := parseInitializeAnswer(make([]byte, 3)); !errors.Is(err, BadAnswer) {
		test.Fatalf("got %v", err)
	}
}

func serverPacket(kind byte, first, second int32) []byte {
	packet := make([]byte, serverMessageBytes)
	packet[0] = kind
	binary.LittleEndian.PutUint32(packet[8:], uint32(first))
	binary.LittleEndian.PutUint32(packet[12:], uint32(second))
	return packet
}

func TestARotationMessageBecomesARotationEvent(test *testing.T) {
	for _, kind := range []byte{messageDeviceStateChanged, messageDeviceStateInitial} {
		event, found := parseRotation(serverPacket(kind, stateChangedReasonRotation, int32(RotationHalf)))
		if !found || event.Kind != RotationChanged || event.X != int32(RotationHalf) {
			test.Fatalf("message %d gave %+v, %v", kind, event, found)
		}
	}
}

func TestOtherServerMessagesAreNotRotations(test *testing.T) {
	cases := [][]byte{
		serverPacket(messageUserInput, int32(PenPress), 0),
		serverPacket(messageDeviceStateChanged, stateChangedReasonRotation+1, int32(RotationHalf)),
		make([]byte, 3),
	}
	for _, packet := range cases {
		if _, found := parseRotation(packet); found {
			test.Errorf("%v was read as a rotation", packet)
		}
	}
}

func TestAServerMessageIsReadAsARotationOrAnInput(test *testing.T) {
	cases := []struct {
		name   string
		packet []byte
		wanted EventKind
	}{
		{"a rotation", serverPacket(messageDeviceStateChanged, stateChangedReasonRotation, int32(RotationHalf)), RotationChanged},
		{"an input", serverPacket(messageUserInput, int32(PenPress), 0), PenPress},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			event, found := parseServerEvent(testCase.packet)
			if !found || event.Kind != testCase.wanted {
				test.Fatalf("got %+v, %v", event, found)
			}
		})
	}
}
