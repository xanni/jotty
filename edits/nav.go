package edits

import (
	"sort"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

/*
Implements navigation through the document.  The logical cursor position is
a combination of the current section 1..len(sections) and the current
character 0..sections[cursor[Sectn]-1].chars

The cursor screen x, y coordinates and current word, sentence and paragraph
are computed by drawWindow() when called from Screen()

The user navigates via scope units (characters, words, sentences, paragraphs
and sections) but the document is stored as a UTF-8 encoded Unicode string.
For character navigation we can just decode and scan through the grapheme
clusters sequentially, but for simplicity and performance we cache
the character indexes of the words, sentences and paragraphs in the current
section and the byte indexes of all the sections that have been parsed so far.

Note that character indexes are relative to the start of the current section
while byte indexes are absolute positions within the document.
*/

type index struct {
	bsectn       int   // byte index of the section in the document
	chars        int   // total number of characters in the section
	bpara, cpara []int // byte and character indexes of each paragraph in the section
	csent        []int // character index of each sentence in the section
	cword        []int // character index of each word in the section
}

var sections = []index{{bpara: []int{0}, cpara: []int{0}, csent: []int{0}}}
var osectn, ochar int // original cursor position

// Find the current paragraph
func getPara() int {
	cpara := sections[cursor[Sectn]-1].cpara
	return sort.Search(len(cpara), func(i int) bool { return cpara[i] >= cursor[Char] })
}

// Add a word to the index if not already present
func indexWord(sn, c int) {
	s := &sections[sn-1]
	if len(s.cword) == 0 || c > s.cword[len(s.cword)-1] {
		s.cword = append(s.cword, c)
	}
}

// Add a sentence to the index if not already present
func indexSent(sn, c int) {
	s := &sections[sn-1]
	if c > s.csent[len(s.csent)-1] {
		s.csent = append(s.csent, c)
	}
}

// Add a paragraph to the index if not already present
func indexPara(sn, b, c int) {
	s := &sections[sn-1]
	if b > s.bpara[len(s.bpara)-1] {
		s.bpara = append(s.bpara, b)
		s.cpara = append(s.cpara, c)
	}
}

// Add a section to the index if not already present
func indexSectn(b int) {
	if b > sections[len(sections)-1].bsectn {
		sections = append(sections,
			index{bsectn: b, bpara: []int{b}, cpara: []int{0}, csent: []int{0}})
	}
}

// Count the characters and words in a section and update the sentence and paragraph indexes
func ScanSection(sn int) {
	var source []byte

	b := sections[sn-1].bsectn
	if sn < len(sections) {
		source = document[b:sections[sn].bsectn]
	} else {
		source = document[b:]
	}

	var c []byte // grapheme cluster
	var chars int
	var f int // Unicode boundary flags
	s := &sections[sn-1]
	state := -1

	if isAlphanumeric(source) {
		s.cword = []int{0}
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
				s.cword = append(s.cword, chars)
			}

			if f&uniseg.MaskSentence != 0 {
				s.csent = append(s.csent, chars)
			}
		}

		if r == '\n' {
			s.bpara = append(s.bpara, b)
			s.cpara = append(s.cpara, chars)
		}
	}

	s.chars = chars
}

// Find the current word, sentence and paragraph in the indexes
func updateCursorPos() {
	c := cursor[Char]
	s := sections[cursor[Sectn]-1]
	cursor[Word] = sort.Search(len(s.cword), func(i int) bool { return s.cword[i] >= c })
	cursor[Sent] = sort.Search(len(s.csent), func(i int) bool { return s.csent[i] >= c })
	cursor[Para] = getPara()
}

func leftChar() {
	if cursor[Char] > 0 {
		cursor[Char]--
	} else if cursor[Sectn] > 1 {
		cursor[Sectn]--
		cursor[Char] = sections[cursor[Sectn]-1].chars
	}
}

func rightChar() {
	if cursor[Char] < sections[cursor[Sectn]-1].chars {
		cursor[Char]++
	} else if cursor[Sectn] < len(sections) {
		cursor[Sectn]++
		cursor[Char] = 0
	}
}

func leftWord() {
	cursor[Char] = 0

	if cursor[Word] > 0 {
		cursor[Char] = sections[cursor[Sectn]-1].cword[cursor[Word]-1]
	} else if cursor[Sectn] > 1 {
		cursor[Sectn]--
		cword := sections[cursor[Sectn]-1].cword
		if len(cword) > 0 {
			cursor[Char] = cword[len(cword)-1]
		}
	}
}

func rightWord() {
	cword := sections[cursor[Sectn]-1].cword
	w := cursor[Word]
	if w < len(cword) && cursor[Char] == cword[w] {
		w++
	}

	if w < len(cword) {
		cursor[Char] = cword[w]
	} else {
		rightSectn()
	}
}

func leftSent() {
	if cursor[Sent] > 0 {
		cursor[Char] = sections[cursor[Sectn]-1].csent[cursor[Sent]-1]
	} else if cursor[Sectn] > 1 {
		cursor[Sectn]--
		csent := sections[cursor[Sectn]-1].csent
		cursor[Char] = csent[len(csent)-1]
	} else {
		cursor[Char] = 0
	}
}

func rightSent() {
	csent := sections[cursor[Sectn]-1].csent
	s := cursor[Sent]
	if s < len(csent) && cursor[Char] == csent[s] {
		s++
	}

	if s < len(csent) {
		cursor[Char] = csent[s]
	} else {
		rightSectn()
	}
}

func leftPara() {
	if cursor[Para] > 0 {
		cursor[Char] = sections[cursor[Sectn]-1].cpara[cursor[Para]-1]
	} else if cursor[Sectn] > 1 {
		cursor[Sectn]--
		cpara := sections[cursor[Sectn]-1].cpara
		cursor[Char] = cpara[len(cpara)-1]
	} else {
		cursor[Char] = 0
	}
}

func rightPara() {
	cpara := sections[cursor[Sectn]-1].cpara
	p := cursor[Para]
	if p < len(cpara) && cursor[Char] == cpara[p] {
		p++
	}

	if p < len(cpara) {
		cursor[Char] = cpara[p]
	} else {
		rightSectn()
	}
}

func leftSectn() {
	if cursor[Sectn] > 1 {
		cursor[Sectn]--
	}
	cursor[Char] = 0
}

func rightSectn() {
	if cursor[Sectn] < len(sections) {
		cursor[Sectn]++
		cursor[Char] = 0
	} else {
		cursor[Char] = sections[cursor[Sectn]-1].chars
	}
}

func Left() {
	osectn = 0
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
	osectn = 0
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
	if osectn == 0 {
		osectn = cursor[Sectn]
		ochar = cursor[Char]
	}

	if scope < Sent {
		p := max(0, cursor[Para]-1)
		cursor[Char] = sections[cursor[Sectn]-1].cpara[p]
		scope = Sent
	} else if scope == Sent {
		if cursor[Sectn] > 1 {
			cursor[Sectn]--
		}
		cursor[Char] = 0
		scope = Para
	} else if scope == Para {
		cursor[Sectn] = 1
		cursor[Char] = 0
		scope = Sectn
	} else if cursor[Sectn] == 1 && cursor[Char] == 0 { // Sectn
		cursor[Sectn] = osectn
		cursor[Char] = ochar
		scope = Char
	}
}

func End() {
	if osectn == 0 {
		osectn = cursor[Sectn]
		ochar = cursor[Char]
	}

	s := &sections[cursor[Sectn]-1]
	if scope < Sent {
		p := cursor[Para]
		if p == 0 || (p < len(s.cpara) && cursor[Char] == s.cpara[p]-1) {
			p++
		}
		if p < len(s.cpara) {
			cursor[Char] = s.cpara[p] - 1
		} else {
			cursor[Char] = s.chars
		}
		scope = Sent
	} else if scope == Sent {
		cursor[Char] = s.chars
		scope = Para
	} else if scope == Para {
		cursor[Sectn] = len(sections)
		cursor[Char] = sections[len(sections)-1].chars
		scope = Sectn
	} else if cursor[Sectn] == len(sections) && cursor[Char] == s.chars { // Sectn
		cursor[Sectn] = osectn
		cursor[Char] = ochar
		scope = Char
	}
}
