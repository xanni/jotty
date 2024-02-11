package edits

import (
	"sort"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

/*
Implements navigation through the document.  The logical cursor position is
a combination of the current section 1..total[Sect] and the current character
0..total[Char]

The cursor screen x, y coordinates and current word, sentence and paragraph
are computed by DrawWindow()

The user navigates via scope units (characters, words, sentences, paragraphs
and sections) but the document is stored as a UTF-8 encoded Unicode string.
For character navigation we can just decode and scan through the grapheme
clusters sequentially, but for simplicity and performance we cache
the character indexes of the words, sentences and paragraphs in the current
section and the byte indexes of all the sections that have been parsed so far.

Note that character indexes are relative to the start of the current section
while byte indexes are absolute positions within the document.
*/

type index struct{ b, c int } // byte and character indexes

var iword = []int{0}    // character index of each word in the section
var isent = []index{{}} // index of each sentence in the section
var ipara = []index{{}} // index of each paragraph in the section
var isect = []int{0}    // byte index of each section in the document
var osect, ochar int    // original cursor position

// Find the current paragraph
func getPara() int {
	return sort.Search(len(ipara), func(i int) bool { return ipara[i].c >= cursor.pos[Char] })
}

// Add a word to the index if not already present
func indexWord(c int) {
	if len(iword) == 0 || c > iword[len(iword)-1] {
		iword = append(iword, c)
	}
}

// Add a sentence to the index if not already present
func indexSent(b, c int) {
	if b > isent[len(isent)-1].b {
		isent = append(isent, index{b, c})
	}
}

// Add a paragraph to the index if not already present
func indexPara(b, c int) {
	if b > ipara[len(ipara)-1].b {
		ipara = append(ipara, index{b, c})
	}
}

// Add a section to the index if not already present
func indexSect(b int) {
	if b > isect[len(isect)-1] {
		isect = append(isect, b)
	}
}

// Reset the sentence and paragraph indexes
func newSection(s int) {
	iword = nil
	isent = []index{{isect[s-1], 0}}
	ipara = []index{{isect[s-1], 0}}
}

// Count the characters and words in a section and update the sentence and paragraph indexes
func scanSect() {
	p := cursor.pos[Sect]
	var source []byte

	b := isect[p-1]
	if p < len(isect) {
		source = document[b:isect[p]]
	} else {
		source = document[b:]
	}

	var c []byte // grapheme cluster
	var chars int
	var f int // Unicode boundary flags
	newSection(p)
	state := -1

	if isAlphanumeric(source) {
		iword = []int{0}
	}

	for len(source) > 0 {
		c, source, f, state = uniseg.Step(source, state)
		b += len(c)
		r, _ := utf8.DecodeRune(c)
		if r == utf8.RuneError {
			continue
		}

		if (f>>uniseg.ShiftWidth) > 0 || r == '\n' {
			chars++
		}

		if len(source) > 0 {
			if f&uniseg.MaskWord != 0 && isAlphanumeric(source) {
				iword = append(iword, chars)
			}

			if f&uniseg.MaskSentence != 0 {
				isent = append(isent, index{b, chars})
			}
		}

		if r == '\n' {
			ipara = append(ipara, index{b, chars})
		}
	}

	total = counts{chars, len(iword), len(isent), len(ipara), len(isect)}
}

// Find the current word, sentence and paragraph in the indexes
func updateCursorPos() {
	c := cursor.pos[Char]
	cursor.pos[Word] = sort.Search(len(iword), func(i int) bool { return iword[i] >= c })
	cursor.pos[Sent] = sort.Search(len(isent), func(i int) bool { return isent[i].c >= c })
	cursor.pos[Para] = getPara()
}

func leftChar() {
	if cursor.pos[Char] > 0 {
		cursor.pos[Char]--
	} else if cursor.pos[Sect] > 1 {
		cursor.pos[Sect]--
		scanSect()
		cursor.pos[Char] = total[Char]
	}
}

func rightChar() {
	if cursor.pos[Char] < total[Char] {
		cursor.pos[Char]++
	} else if cursor.pos[Sect] < total[Sect] {
		cursor.pos[Sect]++
		scanSect()
		cursor.pos[Char] = 0
	}
}

func leftWord() {
	if cursor.pos[Word] > 0 {
		cursor.pos[Char] = iword[cursor.pos[Word]-1]
	} else if cursor.pos[Sect] > 1 {
		cursor.pos[Sect]--
		scanSect()
		cursor.pos[Char] = iword[len(iword)-1]
	} else {
		cursor.pos[Char] = 0
	}
}

func rightWord() {
	w := cursor.pos[Word]
	if w < len(iword) && cursor.pos[Char] == iword[w] {
		w++
	}

	if w < len(iword) {
		cursor.pos[Char] = iword[w]
	} else {
		rightSect()
	}
}

func leftSent() {
	if cursor.pos[Sent] > 0 {
		cursor.pos[Char] = isent[cursor.pos[Sent]-1].c
	} else if cursor.pos[Sect] > 1 {
		cursor.pos[Sect]--
		scanSect()
		cursor.pos[Char] = isent[len(isent)-1].c
	} else {
		cursor.pos[Char] = 0
	}
}

func rightSent() {
	s := cursor.pos[Sent]
	if s < len(isent) && cursor.pos[Char] == isent[s].c {
		s++
	}

	if s < len(isent) {
		cursor.pos[Char] = isent[s].c
	} else {
		rightSect()
	}
}

func leftPara() {
	if cursor.pos[Para] > 0 {
		cursor.pos[Char] = ipara[cursor.pos[Para]-1].c
	} else if cursor.pos[Sect] > 1 {
		cursor.pos[Sect]--
		scanSect()
		cursor.pos[Char] = ipara[len(ipara)-1].c
	} else {
		cursor.pos[Char] = 0
	}
}

func rightPara() {
	p := cursor.pos[Para]
	if p < len(ipara) && cursor.pos[Char] == ipara[p].c {
		p++
	}

	if p < len(ipara) {
		cursor.pos[Char] = ipara[p].c
	} else {
		rightSect()
	}
}

func leftSect() {
	if cursor.pos[Sect] > 1 {
		cursor.pos[Sect]--
		scanSect()
	}
	cursor.pos[Char] = 0
}

func rightSect() {
	if cursor.pos[Sect] < total[Sect] {
		cursor.pos[Sect]++
		scanSect()
		cursor.pos[Char] = 0
	} else {
		cursor.pos[Char] = total[Char]
	}
}

func Left() {
	osect = 0
	switch scope {
	case Char:
		leftChar()
	case Word:
		leftWord()
	case Sent:
		leftSent()
	case Para:
		leftPara()
	default: // Sect
		leftSect()
	}
	DrawWindow()
}

func Right() {
	osect = 0
	switch scope {
	case Char:
		rightChar()
	case Word:
		rightWord()
	case Sent:
		rightSent()
	case Para:
		rightPara()
	default: // Sect
		rightSect()
	}
	DrawWindow()
}

func Home() {
	if osect == 0 {
		osect = cursor.pos[Sect]
		ochar = cursor.pos[Char]
	}

	if scope < Sent {
		p := cursor.pos[Para]
		if p > 0 {
			p--
		}
		cursor.pos[Char] = ipara[p].c
		scope = Sent
	} else if scope == Sent {
		if cursor.pos[Sect] > 1 {
			cursor.pos[Sect]--
			scanSect()
		}
		cursor.pos[Char] = 0
		scope = Para
	} else if scope == Para {
		cursor.pos[Sect] = 1
		cursor.pos[Char] = 0
		scanSect()
		scope = Sect
	} else if cursor.pos[Sect] == 1 && cursor.pos[Char] == 0 { // Sect
		cursor.pos[Sect] = osect
		cursor.pos[Char] = ochar
		if osect > 1 {
			scanSect()
		}
		scope = Char
	}
	DrawWindow()
}

func End() {
	if osect == 0 {
		osect = cursor.pos[Sect]
		ochar = cursor.pos[Char]
	}

	if scope < Sent {
		p := cursor.pos[Para]
		if p == 0 || (p < total[Para] && cursor.pos[Char] == ipara[p].c-1) {
			p++
		}
		if p < total[Para] {
			cursor.pos[Char] = ipara[p].c - 1
		} else {
			cursor.pos[Char] = total[Char]
		}
		scope = Sent
	} else if scope == Sent {
		cursor.pos[Char] = total[Char]
		scope = Para
	} else if scope == Para {
		cursor.pos[Sect] = total[Sect]
		scanSect()
		cursor.pos[Char] = total[Char]
		scope = Sect
	} else if cursor.pos[Sect] == total[Sect] && cursor.pos[Char] == total[Char] { // Sect
		cursor.pos[Sect] = osect
		cursor.pos[Char] = ochar
		if osect < total[Sect] {
			scanSect()
		}
		scope = Char
	}
	DrawWindow()
}
