package keys

// Qt's key codes for the keys this package treats apart (qt.io/doc/qt-6/qt.html#Key-enum).
const (
	qtKeyEscape    = 0x01000000
	qtKeyTab       = 0x01000001
	qtKeyBacktab   = 0x01000002 // Shift and Tab
	qtKeyBackspace = 0x01000003
	qtKeyReturn    = 0x01000004
	qtKeyEnter     = 0x01000005 // the number pad's
	qtKeyDelete    = 0x01000007
	qtKeyHome      = 0x01000010
	qtKeyEnd       = 0x01000011
	qtKeyLeft      = 0x01000012
	qtKeyUp        = 0x01000013
	qtKeyRight     = 0x01000014
	qtKeyDown      = 0x01000015
	qtKeyPageUp    = 0x01000016
	qtKeyPageDown  = 0x01000017
	qtKeyShift     = 0x01000020
	qtKeyControl   = 0x01000021
	qtKeyMeta      = 0x01000022
	qtKeyAlt       = 0x01000023
	qtKeyCapsLock  = 0x01000024
	qtKeyAltGraph  = 0x01001103

	qtKeyDeadGrave      = 0x01001250
	qtKeyDeadAcute      = 0x01001251
	qtKeyDeadCircumflex = 0x01001252
	qtKeyDeadTilde      = 0x01001253
	qtKeyDeadDiaeresis  = 0x01001257
)

// firstQtSpecial is where Qt's codes for keys that type no character begin.
const firstQtSpecial = 0x01000000

// The printable range: Qt numbers these keys by their character in Latin-1.
const (
	firstPrintable = 0x20 // space
	lastAscii      = 0x7e // tilde
	firstLatin1    = 0xa0 // no-break space
	lastLatin1     = 0xff
)

// The combining accents, as Unicode writes them after a letter.
const (
	grave      = '̀'
	acute      = '́'
	circumflex = '̂'
	tilde      = '̃'
	diaeresis  = '̈'
)

// specialKeys maps the editing keys' codes to what they do.
var specialKeys = map[int]Special{
	qtKeyEscape:    Escape,
	qtKeyTab:       Tab,
	qtKeyBacktab:   Tab,
	qtKeyBackspace: Backspace,
	qtKeyReturn:    Enter,
	qtKeyEnter:     Enter,
	qtKeyDelete:    Delete,
	qtKeyHome:      Home,
	qtKeyEnd:       End,
	qtKeyLeft:      Left,
	qtKeyUp:        Up,
	qtKeyRight:     Right,
	qtKeyDown:      Down,
	qtKeyPageUp:    PageUp,
	qtKeyPageDown:  PageDown,
}

// deadKeys maps the accent keys, which type nothing until the next letter, to their accent.
var deadKeys = map[int]rune{
	qtKeyDeadGrave:      grave,
	qtKeyDeadAcute:      acute,
	qtKeyDeadCircumflex: circumflex,
	qtKeyDeadTilde:      tilde,
	qtKeyDeadDiaeresis:  diaeresis,
}

// spacingAccents is how each accent looks on its own, typed when no letter takes it.
var spacingAccents = map[rune]rune{
	grave:      '`',
	acute:      '´',
	circumflex: '^',
	tilde:      '~',
	diaeresis:  '¨',
}
