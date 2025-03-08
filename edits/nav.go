package edits

import (
	"slices"

	ps "github.com/xanni/jotty/permascroll"
)

/*
Implements navigation through the document.  The logical cursor position is a
combination of the current paragraph 1..Paragraphs() and the current character
0..cache[cursor[Para]-1].chars

The cursor screen x, y coordinates and current word and sentence are computed by
drawWindow() when called from Screen()

The user navigates via scope units (characters, words, sentences, and
paragraphs) but the document is stored as UTF-8 encoded Unicode strings. For
character navigation we can just decode and scan through the grapheme clusters
sequentially, but for simplicity and performance we cache the character indexes
of the words and sentences in the current paragraph.
*/

var (
	ocursor counts // Original cursor position
	total   counts // Total scope unit counts in the current paragraph
)

// Last word in the paragraph.
func lastWord(pn int) (c int) {
	cword := cache[pn-1].cword
	if len(cword) > 0 {
		c = cword[len(cword)-1]
	}

	return c
}

// Last sentence in the paragraph.
func lastSentence(pn int) (c int) {
	csent := cache[pn-1].csent
	if len(csent) > 0 {
		c = csent[len(csent)-1]
	}

	return c
}

// Characters in the paragraph.
func paragraphChars(pn int) int { return cache[pn-1].chars }

// Find the current word and sentence in the index.
func updateCursorPos() {
	p := cache[cursor[Para]-1]
	cursor[Word], _ = slices.BinarySearch(p.cword, cursor[Char])
	cursor[Sent], _ = slices.BinarySearch(p.csent, cursor[Char])
}

func leftChar() {
	if cursor[Char] > 0 {
		cursor[Char]--
	} else if cursor[Para] > 1 {
		cursor[Para]--
		cursor[Char] = paragraphChars(cursor[Para])
	}
}

func rightChar() {
	if cursor[Char] < paragraphChars(cursor[Para]) {
		cursor[Char]++
	} else if cursor[Para] < ps.Paragraphs() {
		cursor[Para]++
		cursor[Char] = 0
	}
}

func leftWord() {
	switch {
	case cursor[Word] > 0:
		cursor[Char] = cache[cursor[Para]-1].cword[cursor[Word]-1]
	case cursor[Para] > 1:
		cursor[Para]--
		cursor[Char] = lastWord(cursor[Para])
	default:
		cursor[Char] = 0
	}
}

func rightWord() {
	cword := cache[cursor[Para]-1].cword
	w := cursor[Word]
	if w < len(cword) && cursor[Char] == cword[w] {
		w++
	}

	if w < len(cword) {
		cursor[Char] = cword[w]
	} else {
		rightPara()
	}
}

func leftSent() {
	switch {
	case cursor[Sent] > 0:
		cursor[Char] = cache[cursor[Para]-1].csent[cursor[Sent]-1]
	case cursor[Para] > 1:
		cursor[Para]--
		cursor[Char] = lastSentence(cursor[Para])
	default:
		cursor[Char] = 0
	}
}

func rightSent() {
	csent := cache[cursor[Para]-1].csent
	s := cursor[Sent]
	if s < len(csent) && cursor[Char] == csent[s] {
		s++
	}

	if s < len(csent) {
		cursor[Char] = csent[s]
	} else {
		rightPara()
	}
}

func leftPara() {
	if cursor[Char] == 0 && cursor[Para] > 1 {
		cursor[Para]--
	}
	cursor[Char] = 0
}

func rightPara() {
	if cursor[Para] < ps.Paragraphs() {
		cursor[Para]++
		cursor[Char] = 0
	} else {
		cursor[Char] = paragraphChars(cursor[Para])
	}
}

func Left() {
	ocursor = counts{}
	switch scope {
	case Char:
		leftChar()
	case Word:
		leftWord()
	case Sent:
		leftSent()
	default: // Para
		leftPara()
	}
	updateSelections()
}

func Right() {
	ocursor = counts{}
	switch scope {
	case Char:
		rightChar()
	case Word:
		rightWord()
	case Sent:
		rightSent()
	default: // Para
		rightPara()
	}
	updateSelections()
}

func Home() {
	if ocursor[Para] == 0 {
		ocursor = cursor
	}

	switch {
	case cursor[Char] > 0 || scope < Sent:
		cursor[Char] = 0
		scope = Sent
	case cursor[Para] > 1 || scope == Sent:
		cursor = counts{0, 0, 0, 1}
		scope = Para
	default: // scope == Para
		cursor = ocursor
		scope = Char
	}
	updateSelections()
}

func End() {
	if ocursor[Para] == 0 {
		ocursor = cursor
	}

	lastPara := ps.Paragraphs()
	paraEnd := paragraphChars(cursor[Para])
	switch {
	case cursor[Char] < paraEnd || scope < Sent:
		cursor[Char] = paraEnd
		scope = Sent
	case cursor[Para] < lastPara || scope == Sent:
		cursor[Para] = lastPara
		cursor[Char] = paragraphChars(lastPara)
		scope = Para
	default: // scope == Para
		cursor = ocursor
		scope = Char
	}
	updateSelections()
}
