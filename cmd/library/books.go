package main

import "github.com/akrck02/go-renderer/draw2d"

// book is one thing to read, with its cover and the page it was left on.
type book struct {
	title      string
	author     string
	cover      *draw2d.Image
	paragraphs []string
	page       int
}

// newBooks is the shelf. Every book and author here is made up.
func newBooks() []*book {
	return []*book{
		{title: "The keeper's year", author: "Ida Marsh", cover: mustDecodeCover(lighthouseCoverFile), page: 1, paragraphs: keepersYear},
		{title: "Reading the tide", author: "Tomas Reed", cover: mustDecodeCover(whirlpoolCoverFile), paragraphs: readingTheTide},
		{title: "Twelve useful knots", author: "Ana Lowe", cover: mustDecodeCover(knotCoverFile), paragraphs: usefulKnots},
	}
}

var keepersYear = []string{
	"The lamp is lit an hour before the light goes, and it stays lit until the morning has made it useless.",
	"That is the whole of the job, and it takes the whole of the year. Everything else a keeper does, the polishing and the painting and the long entries in the log, is only a way of being ready for that one hour.",
	"In winter the hour comes early. The sea and the sky turn the same grey, the gulls go quiet, and you climb the hundred and twelve steps with a can of oil in one hand and the other on a rail worn smooth by people you will never meet.",
	"Nobody on the ships out there knows your name. They know the light: three flashes, a long dark, three flashes again. They count it the way you might count a friend's footsteps on the stairs, and they turn a little to the north, and they go home.",
	"I have been asked if it is lonely. It is quiet, which is a different thing. There is the wind, which is never the same twice, and the weather glass, and the small business of the tide going out across the rocks and coming back as if it had forgotten something.",
	"By the spring the days have grown so long that the lamp feels almost like a courtesy. The terns return. The rocks get their green back. You open every window in the tower and the whole building breathes out the winter in one long draught of paraffin and salt.",
	"Summer is for paint. A tower that stands in the sea is always being taken apart, very slowly, a grain at a time, and a keeper's real opponent is not the storm but the patient, ordinary damp.",
	"And then one evening, when you were not paying attention, the hour has moved again. You look up from the brush and the light is going. You put it down, wipe your hands, and climb.",
	"I kept that light for nineteen years. I wrote down the weather every day of them, and it is the only book I ever finished until this one.",
	"What I would like to tell you is what it was like, and I find I can only tell you what I did. So this is a book of what I did, season by season, in the order that the sea allowed.",
	"It begins, as everything out there begins, with checking the oil.",
	"The can holds enough for three nights, and you should never let it go below one. Fill it at noon, when your hands are steady and the light is good, and wipe the spout before you carry it up.",
	"Next comes the wick. A wick that has been trimmed badly smokes, and smoke on the glass costs you a mile of light on a night when somebody may need that mile. There is a small brass tool for the job, and it was old when I arrived.",
	"The glass itself you clean with a soft cloth and a little vinegar, working from the top, and you do not stop to admire it until you are done. Vanity has broken more lenses than weather has.",
	"After the glass, the log. Wind, sea, sky and the state of the oil, in that order and in ink. The man before me wrote his in a hand so small that I took it at first for a kind of stitching.",
	"I have read all of his entries now. In eleven years he allowed himself one remark that was not about the weather. It says: fine morning, saw a whale, did not tell anyone.",
	"I think about that entry more than I would have guessed. There is a great deal out there that nobody is told about, and most of it is perfectly happy that way.",
	"You will want to know about the storms, and I will come to them. But a storm is only a few bad nights, and a year has three hundred and sixty-five, and most of them are like this one: the oil, the wick, the glass, the log, and the slow climb.",
}

var readingTheTide = []string{
	"The sea keeps two appointments a day, and it is never quite on time for either of them.",
	"Learn why, and a beach stops being a place and becomes a clock. The moon pulls, the sun pulls less, and the shape of the bay does the rest.",
	"Walk out at low water and look for the line of weed. It tells you where the last high tide stopped, and so how far you can trust the sand behind you.",
}

var usefulKnots = []string{
	"A good knot does three things: it holds, it does not harm the rope, and it comes undone when you ask it to.",
	"Start with the reef knot, and learn at once what it is not for. Then the bowline, which makes a loop that will not slip, and which you should be able to tie with your eyes shut.",
}
