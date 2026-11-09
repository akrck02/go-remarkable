// Package app runs a Go UI program on a reMarkable Paper Pro, opened from AppLoad. Go UI builds the
// interface, Go Renderer's painter gives rectangles, the processor blends them into a private
// canvas, and only the pixels that changed reach the panel.
//
// The program sleeps between events, so a still screen costs no processor and no panel refresh.
package app

import (
	"hash/maphash"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/opengl/overlay"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"

	"github.com/akrck02/go-remarkable/canvas"
	"github.com/akrck02/go-remarkable/failure"
	"github.com/akrck02/go-remarkable/ink"
	"github.com/akrck02/go-remarkable/paper"
	"github.com/akrck02/go-remarkable/pen"
	"github.com/akrck02/go-remarkable/pointer"
	"github.com/akrck02/go-remarkable/qtfb"
)

// The panel of a reMarkable Paper Pro, and how many of its pixels make one pixel of a layout.
const (
	PanelWidthInPixels   = 1620
	PanelHeightInPixels  = 2160
	PixelsPerLayoutPixel = 2
)

const (
	framebufferKeyVariable = "QTFB_KEY"
	queuedEvents           = 64
	noRefreshModeSentYet   = qtfb.RefreshMode(-1)

	// initialStateWait is how long to wait at the start for the system's orientation, so a tablet
	// held sideways does not paint its whole first frame twice.
	initialStateWait = 200 * time.Millisecond

	// InkRearmDelay is how long after a clean refresh the panel goes back to the writing mode. The
	// clean refresh must be on its way first, and the next stroke then pays no second for the change.
	InkRearmDelay = 2 * time.Second
)

// Device is what a program is given to work with.
type Device struct {
	Canvas *canvas.Canvas    // the frame; ink paints on it
	Input  *input.State      // what the fingers and the pen are doing, in Go Renderer's terms
	Pen    func() pen.Sample // the pen's pressure, tilt and eraser
}

// NewInk makes paper to write on, for a Program's Ink.
func (device Device) NewInk() *ink.Surface {
	return ink.New(device.Canvas, device.Pen, PixelsPerLayoutPixel)
}

// Program is a program for the tablet. Only Draw is needed.
type Program struct {
	Draw func(context *ui.Context) // builds the interface

	RefreshMode func() qtfb.RefreshMode // how the frame about to be painted reaches the panel; qtfb.Content by default
	Ink         *ink.Surface            // the program's paper, if it has one
	InkOpen     func() bool             // whether the paper is on screen now; always by default
	TakeTurn    func() bool             // says once that the interface should turn a quarter
	OnEvent     func(event qtfb.Event)  // sees every event first, to count or log them
}

// withDefaults fills in what the program left out, so that the loop asks nothing twice.
func (program Program) withDefaults(device Device) Program {
	if program.RefreshMode == nil {
		program.RefreshMode = func() qtfb.RefreshMode { return qtfb.Content }
	}
	if program.InkOpen == nil {
		hasInk := program.Ink != nil
		program.InkOpen = func() bool { return hasInk }
	}
	if program.Ink == nil {
		program.Ink = device.NewInk() // never open, and it saves asking whether there is one
	}
	if program.TakeTurn == nil {
		program.TakeTurn = func() bool { return false }
	}
	if program.OnEvent == nil {
		program.OnEvent = func(qtfb.Event) {}
	}
	return program
}

// Display is what a program asks of AppLoad. A qtfb.Connection is one; apptest has a stand-in.
type Display interface {
	SetRefreshMode(mode qtfb.RefreshMode) error
	Refresh(area canvas.Rectangle) error
}

// Run opens the screen AppLoad prepared, shows the first frame and answers events until AppLoad
// closes the program.
func Run(newProgram func(device Device) Program) error {
	key, err := strconv.Atoi(os.Getenv(framebufferKeyVariable))
	if err != nil {
		return failure.Wrap(NoFramebufferKey, framebufferKeyVariable, err)
	}
	connection, err := qtfb.Open(int32(key))
	if err != nil {
		return err
	}
	application, err := New(connection, connection.Memory(), pen.Follow(pen.PaperProDevicePath), newProgram)
	if err != nil {
		return err
	}
	events := make(chan qtfb.Event, queuedEvents)
	go connection.ReadEvents(events)
	return application.Loop(events)
}

// Application joins a program, its input and the screen.
type Application struct {
	target     Display
	memory     []byte
	context    *ui.Context
	state      *input.State
	router     *pointer.Router
	painter    *overlay.Painter
	frame      *canvas.Canvas
	program    Program
	seed       maphash.Seed
	painted    uint64
	mode       qtfb.RefreshMode
	needsBuild bool
	liftedAt   time.Time // when the last press ended, until a frame answers it
	rearmAt    time.Time // when to go back to the writing mode; zero if nothing is planned

	systemTurn        canvas.Orientation // the way the system last said it stands
	systemTurnWaiting bool               // systemTurn has not been applied yet
}

// New prepares a program on a display, with no frame drawn yet. Run does this for the tablet; tests
// call it with a stand-in.
func New(target Display, memory []byte, penDetails func() pen.Sample, newProgram func(device Device) Program) (*Application, error) {
	context, err := ui.New()
	if err != nil {
		return nil, failure.Wrap(InterfaceFailed, "", err)
	}
	context.Theme = paper.Theme()
	context.LessMotion = true
	painter, err := overlay.NewPainter(overlay.Fonts{})
	if err != nil {
		return nil, failure.Wrap(InterfaceFailed, "", err)
	}
	state := input.NewState()
	frame := canvas.New(PanelWidthInPixels, PanelHeightInPixels)
	device := Device{Canvas: frame, Input: state, Pen: penDetails}
	return &Application{
		target:  target,
		memory:  memory,
		context: context,
		state:   state,
		router:  pointer.New(PixelsPerLayoutPixel),
		painter: painter,
		frame:   frame,
		program: newProgram(device).withDefaults(device),
		seed:    maphash.MakeSeed(),
		mode:    noRefreshModeSentYet,
	}, nil
}

// Context is the interface, for looking at what the last frame built.
func (application *Application) Context() *ui.Context { return application.context }

// Canvas is the frame as last composed.
func (application *Application) Canvas() *canvas.Canvas { return application.frame }

// Input is what the program is told the fingers and the pen are doing.
func (application *Application) Input() *input.State { return application.state }

// Loop draws once, then sleeps until there is an event or a frame the interface asked for.
func (application *Application) Loop(events <-chan qtfb.Event) error {
	application.applyInitialState(events)
	if err := application.Update(); err != nil {
		return err
	}
	for {
		var timer <-chan time.Time
		if delay, wanted := application.nextWake(); wanted {
			timer = time.After(delay)
		}
		select {
		case event, open := <-events:
			if !open {
				return nil
			}
			application.Apply(event)
			if !application.applyWaiting(events) {
				return nil
			}
		case <-timer:
			application.needsBuild = true
		}
		if err := application.Answer(); err != nil {
			return err
		}
	}
}

// applyInitialState takes the first message from the server, which says how the system stands.
func (application *Application) applyInitialState(events <-chan qtfb.Event) {
	select {
	case event, open := <-events:
		if open {
			application.Apply(event)
		}
	case <-time.After(initialStateWait):
	}
}

// nextWake is how long to sleep if no event comes, and false when only an event should wake us.
func (application *Application) nextWake() (time.Duration, bool) {
	delay, wanted := application.state.Feed().TakeFrameRequest()
	if wait, due := application.program.Ink.TimeUntilClean(); due && (!wanted || wait < delay) {
		delay, wanted = wait, true
	}
	if !application.rearmAt.IsZero() {
		if wait := max(application.rearmAt.Sub(application.program.Ink.Clock()), 0); !wanted || wait < delay {
			delay, wanted = wait, true
		}
	}
	return delay, wanted
}

// applyWaiting takes in every event already queued, so one frame answers them all: a pen sends
// far more positions than an e-ink panel can show. It says false when the server has gone.
func (application *Application) applyWaiting(events <-chan qtfb.Event) bool {
	for {
		select {
		case event, open := <-events:
			if !open {
				return false
			}
			application.Apply(event)
		default:
			return true
		}
	}
}

// Apply records one event and gives it to the Ink screen or the interface.
func (application *Application) Apply(event qtfb.Event) {
	application.program.OnEvent(event)
	if event.Kind == qtfb.RotationChanged {
		application.noteSystemRotation(qtfb.Rotation(event.X))
		return
	}
	if isPointerEvent(event.Kind) {
		event.X, event.Y = application.frame.Turn.ToCanvas(event.X, event.Y, application.frame.PanelWidth, application.frame.PanelHeight)
	}
	if application.program.InkOpen() && application.applyToInk(event) {
		return
	}
	if event.Kind == qtfb.PenRelease || event.Kind == qtfb.TouchRelease {
		application.liftedAt = time.Now()
	}
	application.router.Apply(application.state, event)
	application.needsBuild = true
}

// noteSystemRotation remembers which way the system says it stands, for the next update.
func (application *Application) noteSystemRotation(rotation qtfb.Rotation) {
	turn, known := orientationFromSystem(rotation)
	if !known {
		slog.Warn("unknown rotation from the server", "rotation", rotation)
		return
	}
	application.systemTurn, application.systemTurnWaiting, application.needsBuild = turn, true, true
}

// applyToInk says whether the Ink screen took the event. A stroke that begins lets go of whatever
// the interface held, such as a palm that landed first.
func (application *Application) applyToInk(event qtfb.Event) bool {
	wasWriting := application.program.Ink.Writing()
	if !application.program.Ink.Receive(event) {
		return false
	}
	if !wasWriting {
		application.router.LetGoOfEverything(application.state)
	}
	return true
}

func isPointerEvent(kind qtfb.EventKind) bool {
	switch kind {
	case qtfb.TouchPress, qtfb.TouchRelease, qtfb.TouchMove, qtfb.PenPress, qtfb.PenRelease, qtfb.PenMove:
		return true
	}
	return false
}

// Answer shows what the events changed: ink at once, and the interface only if it was touched.
func (application *Application) Answer() error {
	if err := application.flushInk(); err != nil {
		return err
	}
	if !application.needsBuild {
		return nil
	}
	application.needsBuild = false
	return application.Update()
}

// flushInk sends new ink to the panel in the writing mode, and cleans it once the pen has rested.
func (application *Application) flushInk() error {
	surface := application.program.Ink
	if pending := surface.TakePending(); !pending.IsEmpty() {
		if changed, found := application.frame.CopyChangedIn(application.memory, pending); found {
			if err := application.sendFrame(changed, qtfb.UltraFast); err != nil {
				return err
			}
		}
	}
	if finished := surface.TakeFinished(); !finished.IsEmpty() {
		application.planRearm()
		return application.sendFrame(application.frame.PanelArea(finished), qtfb.Content)
	}
	return application.rearmIfDue()
}

// planRearm asks to go back to the writing mode a little later.
func (application *Application) planRearm() {
	application.rearmAt = application.program.Ink.Clock().Add(InkRearmDelay)
}

// rearmIfDue puts the panel back in the writing mode while nobody is writing, so the second that
// costs is not paid by the first stroke.
func (application *Application) rearmIfDue() error {
	if application.rearmAt.IsZero() || application.program.Ink.Clock().Before(application.rearmAt) {
		return nil
	}
	application.rearmAt = time.Time{}
	if !application.program.InkOpen() || application.program.Ink.Writing() {
		return nil
	}
	return application.useMode(qtfb.UltraFast)
}

// useMode changes the panel's refresh mode if it is not that one already. A change is expensive:
// AppLoad reads nothing more from this program for a second afterwards.
func (application *Application) useMode(mode qtfb.RefreshMode) error {
	if mode == application.mode {
		return nil
	}
	application.mode = mode
	return application.target.SetRefreshMode(mode)
}

// sendFrame refreshes an area of the panel in a mode. The mode is left set afterwards: the system
// reads it later, when it repaints, so restoring it could change this very refresh.
func (application *Application) sendFrame(area canvas.Rectangle, mode qtfb.RefreshMode) error {
	if err := application.useMode(mode); err != nil {
		return err
	}
	return application.target.Refresh(area)
}

// Update builds the interface twice, since a press changes what the next build shows, and sends
// the result to the screen if it looks different.
func (application *Application) Update() error {
	application.build()
	application.build()
	if turn := application.takeWantedTurn(); turn != application.frame.Turn {
		application.turnCanvas(turn)
		application.build()
		application.build()
	}
	return application.show()
}

// takeWantedTurn is the orientation the screen should have: what the system said, or else the next
// one if the button was pressed.
func (application *Application) takeWantedTurn() canvas.Orientation {
	if application.systemTurnWaiting {
		application.program.TakeTurn() // the tablet was turned, which outranks the button
		application.systemTurnWaiting = false
		return application.systemTurn
	}
	if application.program.TakeTurn() {
		return application.frame.Turn.Next()
	}
	return application.frame.Turn
}

// turnCanvas changes the orientation, keeping the strokes where they are on the panel.
func (application *Application) turnCanvas(turn canvas.Orientation) {
	from := application.frame.Turn
	application.program.Ink.Stop()
	application.router.LetGoOfEverything(application.state) // positions after the turn would make the pointer jump
	application.frame.TurnTo(turn)
	application.program.Ink.Reorient(from, turn)
	application.painted = 0
}

// build runs one step of the interface.
func (application *Application) build() {
	state := application.state
	state.WindowWidth, state.WindowHeight = application.frame.Width/PixelsPerLayoutPixel, application.frame.Height/PixelsPerLayoutPixel
	state.FramebufferWidth, state.FramebufferHeight = application.frame.Width, application.frame.Height
	application.context.Begin(state)
	application.program.Draw(application.context)
	application.context.End()
	state.Feed().EndFrame()
}

// show paints the frame on the processor and refreshes only what changed, if anything did. It logs
// what the frame cost, which is what this program is here to measure.
func (application *Application) show() error {
	drawing := application.context.Drawing()
	if application.program.Ink.TakeCleared() {
		application.painted = 0
	}
	fingerprint := drawing.Fingerprint(application.seed)
	if fingerprint == application.painted {
		application.liftedAt = time.Time{} // that press changed nothing, so no later frame answers it
		return nil
	}
	application.painted = fingerprint
	start := time.Now()
	quadCount := application.paint(drawing)
	painted := time.Since(start)
	changed, found := application.frame.CopyChanged(application.memory)
	mode := application.program.RefreshMode()
	if found {
		if err := application.sendFrame(changed, mode); err != nil {
			return err
		}
	}
	if application.program.InkOpen() {
		application.planRearm()
	}
	slog.Info("frame", "mode", mode, "quads", quadCount, "paint", painted,
		"total", time.Since(start), "changed", changed, "sinceLift", application.takeTimeSinceLift())
	return nil
}

// paint composes the whole frame on the canvas, and says how many rectangles it took.
func (application *Application) paint(drawing *draw2d.List) int {
	quads := application.painter.Paint(&scene.Overlay{Drawing: drawing}, nil, overlay.Frame{
		Width: application.frame.Width, Height: application.frame.Height, Scale: PixelsPerLayoutPixel,
	})
	application.frame.Clear(application.context.Theme.Background)
	for _, quad := range quads {
		application.frame.Paint(quad, application.painter.Atlas().Image)
	}
	if application.program.InkOpen() {
		application.program.Ink.PaintAll()
	}
	return len(quads)
}

// takeTimeSinceLift is how long ago the last press ended, once, or zero if none is waiting. It
// stops at sending the refresh: qtfb never says when the panel has finished.
func (application *Application) takeTimeSinceLift() time.Duration {
	if application.liftedAt.IsZero() {
		return 0
	}
	elapsed := time.Since(application.liftedAt)
	application.liftedAt = time.Time{}
	return elapsed
}

// orientationFromSystem is the orientation that matches a rotation AppLoad reports. It says false
// for a value it does not know. Left becomes right on purpose: AppLoad turns the picture one way to
// keep it fixed to the panel (FBController::paint), so the program has to turn its content the other.
func orientationFromSystem(rotation qtfb.Rotation) (canvas.Orientation, bool) {
	switch rotation {
	case qtfb.RotationNone:
		return canvas.Upright, true
	case qtfb.RotationQuarterLeft:
		return canvas.TurnedRight, true
	case qtfb.RotationQuarterRight:
		return canvas.TurnedLeft, true
	case qtfb.RotationHalf:
		return canvas.UpsideDown, true
	}
	return canvas.Upright, false
}
