// Package qtfb talks to AppLoad, the system extension that gives a program a framebuffer on the
// reMarkable Paper Pro: pixels in shared memory, and a socket for refreshes and input.
package qtfb

import (
	"encoding/binary"
	"log/slog"
	"net"
	"os"
	"strconv"
	"syscall"

	"github.com/akrck02/go-remarkable/canvas"
	"github.com/akrck02/go-remarkable/failure"
)

// The qtfb messages are C unions, so these offsets include the compiler's padding. They match
// AppLoad v0.6.0's src/qtfb/common.h on 64-bit Linux, and were tried on a Paper Pro. This is an
// independent implementation of that protocol; see https://github.com/asivery/rm-appload (GPL-3.0).
const (
	socketPath       = "/tmp/qtfb.sock"
	sharedMemoryRoot = "/dev/shm/qtfb_"

	serverPayloadOffset = 8 // the type byte, then padding up to the union
	serverFieldBytes    = 4

	clientMessageBytes = 24
	serverMessageBytes = 32

	messageInitialize         = 0
	messageUpdate             = 1
	messageUserInput          = 4
	messageSetRefreshMode     = 5
	messageDeviceStateChanged = 7
	messageDeviceStateInitial = 8

	formatPaperProRGB565 = 3
	updatePartial        = 1

	stateChangedReasonRotation int32 = 0
)

// RefreshMode is how the panel updates. Faster modes leave more of the old picture behind.
type RefreshMode int32

// The refresh modes AppLoad accepts. It closes the connection on any higher number.
const (
	UltraFast RefreshMode = 0
	Fast      RefreshMode = 1
	Animate   RefreshMode = 2
	Content   RefreshMode = 3
	Interface RefreshMode = 4
)

// EventKind is what an event is about.
type EventKind int32

// The kinds of event a program receives.
const (
	TouchPress      EventKind = 0x10
	TouchRelease    EventKind = 0x11
	TouchMove       EventKind = 0x12
	PenPress        EventKind = 0x20
	PenRelease      EventKind = 0x21
	PenMove         EventKind = 0x22
	RotationChanged EventKind = 0x100 // not from qtfb's input: a rotation message, with its value in X
)

// Rotation is which way the system says it stands, as AppLoad reports it.
type Rotation int32

// The rotations AppLoad reports.
const (
	RotationNone         Rotation = 0
	RotationQuarterLeft  Rotation = 1
	RotationQuarterRight Rotation = 2
	RotationHalf         Rotation = 3
)

// Event is one pen, touch or rotation event from AppLoad.
type Event struct {
	Kind   EventKind
	Device int32
	X, Y   int32
}

// Connection is a framebuffer AppLoad gave the program: memory to paint and a socket to say what
// changed.
type Connection struct {
	socket net.Conn
	memory []byte
}

// Memory is the framebuffer's pixels, shared with AppLoad.
func (connection *Connection) Memory() []byte { return connection.memory }

// Open connects to AppLoad and maps the framebuffer named by key, which AppLoad gives a program in
// its QTFB_KEY variable.
func Open(key int32) (*Connection, error) {
	socket, err := net.Dial("unixpacket", socketPath)
	if err != nil {
		return nil, failure.Wrap(ServerNotFound, socketPath, err)
	}
	if _, err := socket.Write(initializeMessage(key, formatPaperProRGB565)); err != nil {
		return nil, failure.Wrap(InitializeFailed, "", err)
	}
	packet := make([]byte, serverMessageBytes)
	received, err := socket.Read(packet)
	if err != nil {
		return nil, failure.Wrap(InitializeFailed, "", err)
	}
	memoryKey, memoryBytes, err := parseInitializeAnswer(packet[:received])
	if err != nil {
		return nil, err
	}
	memory, err := mapSharedMemory(memoryKey, memoryBytes)
	if err != nil {
		return nil, err
	}
	return &Connection{socket: socket, memory: memory}, nil
}

// mapSharedMemory maps the pixels the server shares with us.
func mapSharedMemory(key int32, sizeInBytes uint64) ([]byte, error) {
	file, err := os.OpenFile(sharedMemoryRoot+strconv.Itoa(int(key)), os.O_RDWR, 0)
	if err != nil {
		return nil, failure.Wrap(SharedMemoryFailed, "", err)
	}
	defer file.Close() // the mapping outlives the file, and a failed close of a read cannot lose data
	memory, err := syscall.Mmap(int(file.Fd()), 0, int(sizeInBytes), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, failure.Wrap(SharedMemoryFailed, "", err)
	}
	return memory, nil
}

// SetRefreshMode chooses how the panel updates from now on.
func (connection *Connection) SetRefreshMode(mode RefreshMode) error {
	message := make([]byte, clientMessageBytes)
	message[0] = messageSetRefreshMode
	binary.LittleEndian.PutUint32(message[4:], uint32(mode))
	if _, err := connection.socket.Write(message); err != nil {
		return failure.Wrap(UpdateFailed, "", err)
	}
	return nil
}

// Refresh tells the server which Rectangle of the framebuffer changed.
func (connection *Connection) Refresh(area canvas.Rectangle) error {
	if _, err := connection.socket.Write(updateMessage(area)); err != nil {
		return failure.Wrap(UpdateFailed, "", err)
	}
	return nil
}

// ReadEvents sends every input event to events until the server hangs up, then closes it.
func (connection *Connection) ReadEvents(events chan<- Event) {
	defer close(events)
	packet := make([]byte, serverMessageBytes)
	for {
		received, err := connection.socket.Read(packet)
		if err != nil {
			slog.Info("server closed the connection", "cause", err) // how an app is told to leave
			return
		}
		event, found := parseServerEvent(packet[:received])
		if !found {
			slog.Info("server message", "type", packet[0], "bytes", received)
			continue
		}
		events <- event
	}
}

// initializeMessage asks to be connected to a framebuffer in a pixel format.
func initializeMessage(key int32, format uint8) []byte {
	message := make([]byte, clientMessageBytes)
	message[0] = messageInitialize
	binary.LittleEndian.PutUint32(message[4:], uint32(key))
	message[8] = format
	return message
}

// updateMessage says that a Rectangle of the framebuffer changed.
func updateMessage(area canvas.Rectangle) []byte {
	message := make([]byte, clientMessageBytes)
	message[0] = messageUpdate
	fields := []int{updatePartial, area.Left, area.Top, area.Right - area.Left, area.Bottom - area.Top}
	for index, value := range fields {
		binary.LittleEndian.PutUint32(message[4+4*index:], uint32(value))
	}
	return message
}

// parseInitializeAnswer reads the shared memory key and size from the server's first answer.
func parseInitializeAnswer(packet []byte) (key int32, sizeInBytes uint64, err error) {
	if len(packet) != serverMessageBytes || packet[0] != messageInitialize {
		return 0, 0, BadAnswer
	}
	return int32(binary.LittleEndian.Uint32(packet[8:])), binary.LittleEndian.Uint64(packet[16:]), nil
}

// parseServerEvent reads a message the program acts on, an input or a rotation.
func parseServerEvent(packet []byte) (Event, bool) {
	if event, found := parseRotation(packet); found {
		return event, true
	}
	return parseInput(packet)
}

// parseRotation reads a message saying which way the system stands. AppLoad sends one on connecting
// and another each time the system turns.
func parseRotation(packet []byte) (Event, bool) {
	if len(packet) != serverMessageBytes || (packet[0] != messageDeviceStateChanged && packet[0] != messageDeviceStateInitial) {
		return Event{}, false
	}
	if serverField(packet, 0) != stateChangedReasonRotation {
		return Event{}, false
	}
	return Event{Kind: RotationChanged, X: serverField(packet, 1)}, true
}

// parseInput reads an input event, and says false for any other message.
func parseInput(packet []byte) (Event, bool) {
	if len(packet) != serverMessageBytes || packet[0] != messageUserInput {
		return Event{}, false
	}
	return Event{
		Kind:   EventKind(serverField(packet, 0)),
		Device: serverField(packet, 1),
		X:      serverField(packet, 2),
		Y:      serverField(packet, 3),
	}, true
}

// serverField reads one of the four-byte numbers of a server message's contents.
func serverField(packet []byte, index int) int32 {
	return int32(binary.LittleEndian.Uint32(packet[serverPayloadOffset+index*serverFieldBytes:]))
}
