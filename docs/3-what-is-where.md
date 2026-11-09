# What is where

The packages of Go for reMarkable, the programs and tools it comes with, what you need, and what was
checked on a real tablet.

## What is where

| Piece | What it does |
|---|---|
| `app` | The program loop: `app.Run` opens the screen and answers events until AppLoad closes the program. A program is an `app.Program`, and only `Draw` is required. Between events it sleeps. |
| `qtfb` | The connection to AppLoad: pixels in shared memory, a socket for refreshes, input events, the refresh modes and the tablet's rotation. |
| `canvas` | The frame, composed on the processor. It blends Go Renderer's shapes, turns for each orientation, and copies to the panel only the pixels that changed. |
| `ink` | `ink.Surface`, paper to write on: lines that follow the pen with no interface in between, an eraser, and a clean refresh when the pen rests. |
| `pen` | The pen's pressure, tilt and eraser, read from its own device, which is never taken from the system. If it cannot be read, ink still works with plain lines. |
| `pointer` | Turns the tablet's pen and fingers into the single pointer an interface expects, and copes with a resting palm and with presses whose release never came. |
| `paper` | A calm look for e-ink: a theme of soft ink and thin lines on pure white, and buttons, a switch, cards and a progress bar that cost no extra refresh when touched. |
| `apptest` | Runs a program with no tablet: a stand-in for AppLoad that remembers what it was asked, and events delivered by hand. |
| `failure` | Errors with a stable name, such as `qtfb.server-not-found`, to check with `errors.Is`. |
| `cmd/hello`, `cmd/counter`, `cmd/notebook` | The small programs of [the quest](2-the-quest.md). |
| `cmd/library` | The example of a finished interface: a shelf of books, a reader and a notebook. |
| `cmd/demo` | The test bench: Reading, Buttons, Ink, Gestures, Screen and Monitor. Each screen logs what it cost. |
| `tools/build.sh` | Builds `./cmd/<program>` for the tablet into `out/<program>`, with its manifest. |
| `tools/install.sh` | Copies `out/<program>` into AppLoad's folder over SSH. It writes only inside the program's own folder. `REMARKABLE_HOST` changes where it goes. |

## How a frame gets to the panel

```
your Draw → Go UI → Go Renderer's painter → canvas → only what changed → AppLoad → panel
```

Ink takes a short cut, because a pen cannot wait for an interface:

```
the pen → ink.Surface → canvas → that small area → AppLoad → panel
```

The first picture is built twice per update, because a press changes what the next build shows, and
it reaches the panel only if the frame looks different from the last one.

## What you need

- Go 1.26 or later. The programs are pure Go, so `tools/build.sh` needs nothing else.
- A reMarkable Paper Pro in developer mode, reachable over SSH.
- [xovi](https://github.com/asivery/xovi) and [AppLoad](https://github.com/asivery/rm-appload) on it,
  started as in [the quest](2-the-quest.md#before-you-start). xovi's table of names must be built for
  your exact version of the tablet's software.

`go.mod` pins Go UI and Go Renderer at commits of theirs, until they release what this uses.

## What was checked

On a Paper Pro with software 3.29.1.0 and a build of AppLoad's `main` branch, the demo was tried by
hand, before it was divided into these packages:

| What | Result |
|---|---|
| Reading and Buttons | Good. A tap costs one frame of about 150 ms up to sending the refresh. |
| Ink | Follows the pen and comes out clean; pressure, tilt and eraser were felt to work; a palm breaks nothing. |
| Gestures, Monitor | All respond. |
| Screen | The turn button keeps taps in the right place; `Fast` sometimes looks greyer. |
| Turning the tablet | All four positions, in full screen. |
| AppLoad's own gestures | The swipe from the top and five fingers leave the program working. |
| At rest | All six screens stayed at 0 % processor with no frames, for 30 seconds each. |

**Not run on a tablet:** `cmd/hello`, `cmd/counter`, `cmd/notebook`, `cmd/library`, `tools/build.sh` and
`tools/install.sh` as they stand here, nor the demo since it became packages. The tests run the same
code with a stand-in for AppLoad. The demo's two-finger list and its swipe to the end, and the
eraser's band at its final width, were installed but not reported on.

## Before writing a larger program

- **Changing the refresh mode costs a second.** AppLoad reads nothing more from a program for a second
  after each change. Keep one mode per screen. The mode you set is read later, when the panel repaints,
  so do not set it back straight after a refresh.
- **Modes above `qtfb.Interface` close the connection.** AppLoad takes five, and hangs up on the rest.
- **The refresh is not the end.** The times logged stop at sending it; AppLoad never says when the panel
  has finished.
- **Some gestures are AppLoad's.** A swipe from the top edge opens its bar, and five fingers force a full
  refresh.
- **Input can be messy.** AppLoad v0.6.0 reports a finger that did not move as another press; its
  `main` branch fixes that. A system menu can keep a finger's release for itself. `pointer` copes
  with both.
- **Follow the tablet's conventions.** On the rest of the tablet two fingers scroll, and a swipe down
  goes to the end of the page. `cmd/demo/gestures.go` shows how.
- **In a window, nothing turns by itself.** See [level 4](2-the-quest.md#level-4--turn-the-tablet).

## Not yet

- Only the Paper Pro. The panel's size and pixel format are written into `app` and `qtfb`.
- `canvas` paints no frosted backdrops, so Go UI's glass does not show.
- A program cannot save what is written yet: strokes live in memory.
- No text entry: nothing brings up a keyboard.
- No desktop runner in this repository: besides the tablet, there is only `apptest`.
- Once, before these packages, the tablet's input stopped reaching a program for minutes. It did not
  come back in later tries, and its cause is unknown.
