package edits

import (
	"slices"

	doc "git.sericyb.com.au/jotty/document"
)

/*
Implements navigation through the document.  The logical cursor position is
a combination of the current section 1..Sections(), the current paragraph
1..Paragraphs(cursor[Sectn]) and the current character
0..sections[cursor[Sectn]-1].p[cursor[Para]-1].chars

The cursor screen x, y coordinates and current word and sentence are computed
by drawWindow() when called from Screen()

The user navigates via scope units (characters, words, sentences, paragraphs
and sections) but the document is stored as UTF-8 encoded Unicode strings.
For character navigation we can just decode and scan through the grapheme
clusters sequentially, but for simplicity and performance we cache
the character indexes of the words and sentences in the current paragraph.
*/

// Paragraph index
type ipara struct {
	csent []int // character index of each sentence in the paragraph
	cword []int // character index of each word in the paragraph
	chars int   // total number of characters in the paragraph
}

// Section index
type isectn struct {
	sents, words, chars int     // Total number of sentences, words and characters in the section
	p                   []ipara // Paragraphs in the section
}

var sections = []isectn{{p: []ipara{{}}}}
var ocursor counts // Original cursor position

// Add a word to the index if not already present
func indexWord(sn, pn, c int) {
	s := &sections[sn-1]
	p := &s.p[pn-1]
	if len(p.cword) == 0 || c > p.cword[len(p.cword)-1] {
		s.words++
		p.cword = append(p.cword, c)
	}
}

// Add a sentence to the index if not already present
func indexSent(sn, pn, c int) {
	s := &sections[sn-1]
	p := &s.p[pn-1]
	if len(p.csent) == 0 || c > p.csent[len(p.csent)-1] {
		s.sents++
		p.csent = append(p.csent, c)
	}
}

// Add a paragraph to the index
func indexPara(sn int) {
	s := &sections[sn-1]
	s.p = append(s.p, ipara{})
}

// Add a section to the index
func indexSectn() {
	sections = append(sections, isectn{p: []ipara{{}}}) // Empty paragraph
}

// Last sentence in the paragraph
func lastSentence(sn, pn int) int {
	csent := sections[sn-1].p[pn-1].csent
	if len(csent) > 0 {
		return csent[len(csent)-1]
	}

	return 0
}

// Last word in the paragraph
func lastWord(sn, pn int) int {
	cword := sections[sn-1].p[pn-1].cword
	if len(cword) > 0 {
		return cword[len(cword)-1]
	}

	return 0
}

// Characters in the paragraph
func paragraphChars(s, p int) int {
	return sections[s-1].p[p-1].chars
}

// Characters in the section
func sectionChars(s int) int {
	return sections[s-1].chars
}

// Find the current word and sentence in the indexes
func updateCursorPos() {
	p := sections[cursor[Sectn]-1].p[cursor[Para]-1]
	cursor[Sent], _ = slices.BinarySearch[[]int](p.csent, cursor[Char])
	cursor[Word], _ = slices.BinarySearch[[]int](p.cword, cursor[Char])
}

func leftChar() {
	switch {
	case cursor[Char] > 0:
		cursor[Char]--
	case cursor[Para] > 1:
		cursor[Para]--
		cursor[Char] = paragraphChars(cursor[Sectn], cursor[Para])
	case cursor[Sectn] > 1:
		cursor[Sectn]--
		cursor[Para] = doc.Paragraphs(cursor[Sectn])
		cursor[Char] = paragraphChars(cursor[Sectn], cursor[Para])
	}
}

func rightChar() {
	switch {
	case cursor[Char] < paragraphChars(cursor[Sectn], cursor[Para]):
		cursor[Char]++
	case cursor[Para] < doc.Paragraphs(cursor[Sectn]):
		cursor[Para]++
		cursor[Char] = 0
	case cursor[Sectn] < doc.Sections():
		cursor[Sectn]++
		cursor[Para] = 1
		cursor[Char] = 0
	}
}

func leftWord() {
	switch {
	case cursor[Word] > 0:
		cursor[Char] = sections[cursor[Sectn]-1].p[cursor[Para]-1].cword[cursor[Word]-1]
	case cursor[Para] > 1:
		cursor[Para]--
		cursor[Char] = lastWord(cursor[Sectn], cursor[Para])
	case cursor[Sectn] > 1:
		cursor[Sectn]--
		cursor[Para] = doc.Paragraphs(cursor[Sectn])
		cursor[Char] = lastWord(cursor[Sectn], cursor[Para])
	default:
		cursor[Char] = 0
	}
}

func rightWord() {
	cword := sections[cursor[Sectn]-1].p[cursor[Para]-1].cword
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
		cursor[Char] = sections[cursor[Sectn]-1].p[cursor[Para]-1].csent[cursor[Sent]-1]
	case cursor[Para] > 1:
		cursor[Para]--
		cursor[Char] = lastSentence(cursor[Sectn], cursor[Para])
	case cursor[Sectn] > 1:
		cursor[Sectn]--
		cursor[Para] = doc.Paragraphs(cursor[Sectn])
		cursor[Char] = lastSentence(cursor[Sectn], cursor[Para])
	default:
		cursor[Char] = 0
	}
}

func rightSent() {
	csent := sections[cursor[Sectn]-1].p[cursor[Para]-1].csent
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
	switch {
	case cursor[Char] > 0:
		cursor[Char] = 0
	case cursor[Para] > 1:
		cursor[Para]--
		cursor[Char] = 0
	default:
		leftSectn()
	}
}

func rightPara() {
	if cursor[Para] < doc.Paragraphs(cursor[Sectn]) {
		cursor[Para]++
		cursor[Char] = 0
	} else {
		rightSectn()
	}
}

func leftSectn() {
	if cursor[Para] > 1 {
		cursor[Para] = 1
	} else if cursor[Sectn] > 1 {
		cursor[Sectn]--
		cursor[Para] = doc.Paragraphs(cursor[Sectn])
	}
	cursor[Char] = 0
}

func rightSectn() {
	if cursor[Sectn] < doc.Sections() {
		cursor[Sectn]++
		cursor[Para] = 1
		cursor[Char] = 0
	} else {
		cursor[Para] = doc.Paragraphs(cursor[Sectn])
		cursor[Char] = paragraphChars(cursor[Sectn], cursor[Para])
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
	case Para:
		leftPara()
	default: // Sectn
		leftSectn()
	}
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
	case Para:
		rightPara()
	default: // Sectn
		rightSectn()
	}
}

func Home() {
	if ocursor[Sectn] == 0 {
		ocursor = cursor
	}

	switch {
	case scope < Sent:
		cursor[Char] = 0
		scope = Sent
	case scope == Sent:
		cursor[Para] = 1
		cursor[Char] = 0
		scope = Para
	case scope == Para:
		cursor = counts{0, 0, 0, 1, 1}
		scope = Sectn
	case cursor == (counts{0, 0, 0, 1, 1}): // scope == Sectn
		cursor = ocursor
		scope = Char
	}
}

func End() {
	if ocursor[Sectn] == 0 {
		ocursor = cursor
	}

	ts := doc.Sections()
	switch {
	case scope < Sent:
		cursor[Char] = paragraphChars(cursor[Sectn], cursor[Para])
		scope = Sent
	case scope == Sent:
		cursor[Para] = doc.Paragraphs(cursor[Sectn])
		cursor[Char] = paragraphChars(cursor[Sectn], cursor[Para])
		scope = Para
	case scope == Para:
		cursor[Sectn] = ts
		cursor[Para] = doc.Paragraphs(cursor[Sectn])
		cursor[Char] = paragraphChars(cursor[Sectn], cursor[Para])
		scope = Sectn
	case cursor[Sectn] == ts &&
		cursor[Para] == doc.Paragraphs(ts) &&
		cursor[Char] == paragraphChars(ts, doc.Paragraphs(ts)): // scope == Sectn
		cursor = ocursor
		scope = Char
	}
}
