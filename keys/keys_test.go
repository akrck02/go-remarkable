package keys

import (
	"testing"

	"github.com/akrck02/go-remarkable/qtfb"
)

func press(code int) qtfb.Event   { return qtfb.Event{Kind: qtfb.KeyPress, X: int32(code)} }
func release(code int) qtfb.Event { return qtfb.Event{Kind: qtfb.KeyRelease, X: int32(code)} }

// typedText feeds events to a fresh Keyboard and joins the text they typed.
func typedText(events ...qtfb.Event) string {
	var keyboard Keyboard
	text := ""
	for _, event := range events {
		if key, typed := keyboard.Read(event); typed {
			text += key.Text
		}
	}
	return text
}

func TestLettersComeOutAsTypedWithShiftAndCapsLock(test *testing.T) {
	cases := []struct {
		name   string
		events []qtfb.Event
		want   string
	}{
		{"plain letters are small", []qtfb.Event{press('H'), press('I')}, "hi"},
		{"shift makes a capital while held", []qtfb.Event{press(qtKeyShift), press('H'), release(qtKeyShift), press('I')}, "Hi"},
		{"caps lock makes capitals until pressed again", []qtfb.Event{press(qtKeyCapsLock), press('A'), press(qtKeyCapsLock), press('B')}, "Ab"},
		{"shift with caps lock makes a small letter", []qtfb.Event{press(qtKeyCapsLock), press(qtKeyShift), press('A')}, "a"},
		{"digits and signs come as their character", []qtfb.Event{press('1'), press('!'), press(' '), press('#')}, "1! #"},
		{"latin-1 letters such as ñ keep their case", []qtfb.Event{press(0xd1), press(qtKeyShift), press(0xd1)}, "ñÑ"},
	}
	for _, current := range cases {
		if got := typedText(current.events...); got != current.want {
			test.Errorf("%s: typed %q, want %q", current.name, got, current.want)
		}
	}
}

func TestAnAccentKeyWaitsForItsLetter(test *testing.T) {
	cases := []struct {
		name   string
		events []qtfb.Event
		want   string
	}{
		{"acute and e make é", []qtfb.Event{press(qtKeyDeadAcute), press('E')}, "é"},
		{"diaeresis and a capital U make Ü", []qtfb.Event{press(qtKeyDeadDiaeresis), press(qtKeyShift), press('U')}, "Ü"},
		{"an accent that fits no letter is typed before it", []qtfb.Event{press(qtKeyDeadAcute), press('T')}, "´t"},
		{"an editing key drops the waiting accent", []qtfb.Event{press(qtKeyDeadAcute), press(qtKeyLeft), press('E')}, "e"},
	}
	for _, current := range cases {
		if got := typedText(current.events...); got != current.want {
			test.Errorf("%s: typed %q, want %q", current.name, got, current.want)
		}
	}
}

func TestEditingKeysAreSpecialAndReleasesTypeNothing(test *testing.T) {
	var keyboard Keyboard
	if key, typed := keyboard.Read(press(qtKeyBackspace)); !typed || key.Special != Backspace {
		test.Errorf("backspace gave %+v, %v", key, typed)
	}
	if key, typed := keyboard.Read(press(qtKeyReturn)); !typed || key.Special != Enter {
		test.Errorf("return gave %+v, %v", key, typed)
	}
	if _, typed := keyboard.Read(release('A')); typed {
		test.Error("a release typed something")
	}
	if _, typed := keyboard.Read(qtfb.Event{Kind: qtfb.PenPress}); typed {
		test.Error("a pen event typed something")
	}
}

func TestAKeyWithoutAMeaningComesWithItsCode(test *testing.T) {
	var keyboard Keyboard
	const qtKeyF13 = 0x0100003c
	if key, typed := keyboard.Read(press(qtKeyF13)); !typed || key.Special != Other || key.Code != qtKeyF13 {
		test.Errorf("F13 gave %+v, %v", key, typed)
	}
}

func TestControlIsReportedForShortcuts(test *testing.T) {
	var keyboard Keyboard
	keyboard.Read(press(qtKeyControl))
	if key, _ := keyboard.Read(press('S')); !key.Control || key.Text != "s" {
		test.Errorf("control and s gave %+v", key)
	}
	keyboard.Read(release(qtKeyControl))
	if key, _ := keyboard.Read(press('S')); key.Control {
		test.Errorf("control stayed held after its release: %+v", key)
	}
}
