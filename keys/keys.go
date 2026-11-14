// Package keys turns the key events AppLoad forwards from a physical keyboard, such as the Type
// Folio, into text and editing keys.
//
// AppLoad sends each key as a Qt key code, the same code for "a" and "A". So a Keyboard keeps
// track of Shift, Caps Lock and the accent keys itself, and gives back the letter that was meant.
package keys

import (
	"unicode"

	"github.com/akrck02/go-remarkable/qtfb"
)

// Special is a key that edits or moves instead of typing.
type Special int

// The keys that do not type. None is the zero value: the key typed Text.
const (
	None Special = iota
	Backspace
	Delete
	Enter
	Tab
	Escape
	Left
	Right
	Up
	Down
	Home
	End
	PageUp
	PageDown
	Other // a key that neither types nor edits, such as a keyboard's own function key; see Code
)

// Key is one key someone pressed.
type Key struct {
	Text    string  // what it types, when Special is None
	Special Special // the editing key, when it is one
	Control bool    // whether Control or Meta was held, as in a shortcut
	Code    int     // the Qt key code, for an Other key: a program may give it a use
}

// Keyboard remembers the modifiers and the accent waiting for its letter.
type Keyboard struct {
	shift, capsLock, control bool
	waitingAccent            rune // a combining accent, or 0
}

// Read takes one event and says which Key it was, or false when it typed nothing: a release, a
// modifier, an accent waiting for its letter, or an event that is not a key.
func (keyboard *Keyboard) Read(event qtfb.Event) (Key, bool) {
	code := int(event.X)
	switch event.Kind {
	case qtfb.KeyRelease:
		keyboard.setModifier(code, false)
		return Key{}, false
	case qtfb.KeyPress:
	default:
		return Key{}, false
	}
	if keyboard.setModifier(code, true) {
		return Key{}, false
	}
	if code == qtKeyCapsLock {
		keyboard.capsLock = !keyboard.capsLock
		return Key{}, false
	}
	if accent, isAccent := deadKeys[code]; isAccent {
		keyboard.waitingAccent = accent
		return Key{}, false
	}
	if special, isSpecial := specialKeys[code]; isSpecial {
		keyboard.waitingAccent = 0
		return Key{Special: special, Control: keyboard.control}, true
	}
	if code >= firstQtSpecial {
		keyboard.waitingAccent = 0
		return Key{Special: Other, Code: code, Control: keyboard.control}, true
	}
	return keyboard.typed(code)
}

// setModifier records Shift or Control going down or up, and says whether code was one of them.
func (keyboard *Keyboard) setModifier(code int, down bool) bool {
	switch code {
	case qtKeyShift:
		keyboard.shift = down
	case qtKeyControl, qtKeyMeta:
		keyboard.control = down
	case qtKeyAlt, qtKeyAltGraph:
	default:
		return false
	}
	return true
}

// typed is the text of a printable key, with its case and any waiting accent.
func (keyboard *Keyboard) typed(code int) (Key, bool) {
	letter, printable := printableRune(code)
	if !printable {
		keyboard.waitingAccent = 0
		return Key{}, false
	}
	if unicode.IsLetter(letter) && keyboard.shift == keyboard.capsLock {
		letter = unicode.ToLower(letter)
	}
	text := string(letter)
	if keyboard.waitingAccent != 0 {
		text = withAccent(letter, keyboard.waitingAccent)
		keyboard.waitingAccent = 0
	}
	return Key{Text: text, Control: keyboard.control}, true
}

// printableRune is the character of a Qt key code. Qt numbers printable keys by their Latin-1
// character, and letters by their capital.
func printableRune(code int) (rune, bool) {
	if code < firstPrintable || code > lastLatin1 || (code > lastAscii && code < firstLatin1) {
		return 0, false
	}
	return rune(code), true
}

// withAccent puts the accent on the letter when such a character exists ("e" and an acute make
// "é"). Otherwise it types the accent on its own and then the letter, as a desktop does.
func withAccent(letter, accent rune) string {
	if composed, found := accented[string([]rune{letter, accent})]; found {
		return composed
	}
	return string(spacingAccents[accent]) + string(letter)
}

// accentedLetters lists, for each accent, the letters it goes on and what they become.
var accentedLetters = map[rune][2]string{
	acute:      {"aeiouyAEIOUY", "áéíóúýÁÉÍÓÚÝ"},
	grave:      {"aeiouAEIOU", "àèìòùÀÈÌÒÙ"},
	circumflex: {"aeiouAEIOU", "âêîôûÂÊÎÔÛ"},
	diaeresis:  {"aeiouyAEIOUY", "äëïöüÿÄËÏÖÜŸ"},
	tilde:      {"anoANO", "ãñõÃÑÕ"},
}

// accented maps a letter followed by its accent to the composed character, built from
// accentedLetters once.
var accented = buildAccented()

func buildAccented() map[string]string {
	composed := map[string]string{}
	for accent, letters := range accentedLetters {
		plain, marked := []rune(letters[0]), []rune(letters[1])
		for index, letter := range plain {
			composed[string([]rune{letter, accent})] = string(marked[index])
		}
	}
	return composed
}
