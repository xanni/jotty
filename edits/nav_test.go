package edits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetIndex() {
	sections = []index{{bpara: []int{0}, cpara: []int{0}, csent: []int{0}}}
}

func TestScanSection(t *testing.T) {
	resetIndex()
	document = nil
	ScanSection(1)
	s := &sections[0]
	assert.Equal(t, 0, s.chars)
	assert.Equal(t, 0, len(s.cword))
	assert.Equal(t, 1, len(s.csent))
	assert.Equal(t, 1, len(s.cpara))

	document = []byte("Six words, two sentences.\nTwo paragraphs.")
	ScanSection(1)
	assert.Equal(t, 41, s.chars)
	assert.Equal(t, 6, len(s.cword))
	assert.Equal(t, 2, len(s.csent))
	assert.Equal(t, 2, len(s.cpara))

	document = append(document, []byte("\fAnother section.")...)
	indexSectn(42)
	ScanSection(1)
	assert.Equal(t, 41, s.chars)
	assert.Equal(t, 6, len(s.cword))
	assert.Equal(t, 2, len(s.csent))
	assert.Equal(t, 2, len(s.cpara))

	resetIndex()
	document = []byte{'1', 255, '2'}
	ScanSection(1)
	s = &sections[0]
	assert.Equal(t, 2, s.chars)
	assert.Equal(t, 1, len(s.cword))
	assert.Equal(t, 1, len(s.csent))
	assert.Equal(t, 1, len(s.cpara))
}

func TestLeftChar(t *testing.T) {
	document = []byte("One\fTwo")
	sections = []index{{chars: 3}, {chars: 3}}
	cursor[Sectn] = 2
	cursor[Char] = 1

	leftChar()
	assert.Equal(t, 2, cursor[Sectn])
	assert.Equal(t, 0, cursor[Char])

	leftChar()
	assert.Equal(t, 1, cursor[Sectn])
	assert.Equal(t, 3, cursor[Char])

	cursor[Char] = 1
	leftChar()
	assert.Equal(t, 1, cursor[Sectn])
	assert.Equal(t, 0, cursor[Char])

	leftChar()
	assert.Equal(t, 1, cursor[Sectn])
	assert.Equal(t, 0, cursor[Char])
}

func TestRightChar(t *testing.T) {
	document = []byte("One\fTwo")
	sections = []index{{chars: 3}, {chars: 3}}
	cursor[Sectn] = 1
	cursor[Char] = 2

	rightChar()
	assert.Equal(t, 1, cursor[Sectn])
	assert.Equal(t, 3, cursor[Char])

	rightChar()
	assert.Equal(t, 2, cursor[Sectn])
	assert.Equal(t, 0, cursor[Char])

	cursor[Char] = 3
	rightChar()
	assert.Equal(t, 2, cursor[Sectn])
	assert.Equal(t, 3, cursor[Char])
}

func TestLeftWord(t *testing.T) {
	document = []byte("1 23\f4")
	sections = []index{{chars: 5, cword: []int{0, 2}}, {bsectn: 5, chars: 1, cword: []int{0}}}
	cursor = counts{1, 1, 1, 1, 2}

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, 1, cursor[Word])

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	cursor = counts{4, 2, 1, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, 1, cursor[Word])

	cursor = counts{3, 2, 1, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, 1, cursor[Word])

	document = []byte("1\f\f2")
	sections = []index{
		{chars: 1, cword: []int{0}},
		{bsectn: 2, chars: 0, cword: []int(nil)},
		{bsectn: 3, chars: 1, cword: []int{0}},
	}
	cursor = counts{0, 0, 1, 1, 3}
	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)
}

func TestRightWord(t *testing.T) {
	document = []byte("1\f23 4")
	sections = []index{{chars: 1, cword: []int{0}}, {bsectn: 2, chars: 4, cword: []int{0, 3}}}
	cursor = counts{Sectn: 1}

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	rightWord()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Word])

	rightWord()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 2, cursor[Word])

	rightWord()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 2, cursor[Word])

	cursor = counts{1, 1, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Word])

	cursor = counts{2, 1, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Word])
}

func TestLeftSent(t *testing.T) {
	document = []byte("1. 23\f4")
	sections = []index{{chars: 5, csent: []int{0, 3}}, {bsectn: 6, chars: 1, csent: []int{0}}}
	cursor = counts{1, 1, 1, 1, 2}

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	cursor = counts{4, 2, 2, 1, 1}
	leftSent()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])
}

func TestRightSent(t *testing.T) {
	document = []byte("1\f23. 4")
	sections = []index{{chars: 1, csent: []int{0}}, {bsectn: 2, chars: 5, csent: []int{0, 4}}}
	cursor = counts{Sectn: 1}

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	rightSent()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])

	rightSent()
	updateCursorPos()
	assert.Equal(t, 5, cursor[Char])
	assert.Equal(t, 2, cursor[Sent])

	rightSent()
	updateCursorPos()
	assert.Equal(t, 5, cursor[Char])
	assert.Equal(t, 2, cursor[Sent])

	cursor = counts{1, 1, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])

	cursor = counts{2, 1, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])
}

func TestLeftPara(t *testing.T) {
	document = []byte("1\n23\f4")
	sections = []index{
		{chars: 4, bpara: []int{0, 2}, cpara: []int{0, 2}},
		{bsectn: 5, chars: 1, bpara: []int{5}, cpara: []int{0}},
	}
	cursor = counts{1, 1, 1, 1, 2}

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	leftPara()
	updateCursorPos()
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, 1, cursor[Para])

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	cursor = counts{3, 2, 2, 2, 1}
	leftPara()
	updateCursorPos()
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, 1, cursor[Para])
}

func TestRightPara(t *testing.T) {
	document = []byte("1\f23\n4")
	sections = []index{
		{chars: 1, bpara: []int{0}, cpara: []int{0}},
		{bsectn: 2, chars: 4, bpara: []int{2, 5}, cpara: []int{0, 3}},
	}
	cursor = counts{Sectn: 1}

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	rightPara()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Para])

	rightPara()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 2, cursor[Para])

	rightPara()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 2, cursor[Para])

	cursor = counts{1, 1, 1, 1, 2}
	rightPara()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Para])

	cursor = counts{2, 1, 1, 1, 2}
	rightPara()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Para])
}

func TestLeftSectn(t *testing.T) {
	document = []byte("1\f23\f4")
	sections = []index{{chars: 1}, {bsectn: 2, chars: 2}, {bsectn: 5, chars: 1}}
	cursor = counts{1, 1, 1, 1, 3}

	leftSectn()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	leftSectn()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	leftSectn()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)
}

func TestRightSectn(t *testing.T) {
	document = []byte("1\f23\f4")
	sections = []index{{chars: 1}, {bsectn: 2, chars: 2}, {bsectn: 5, chars: 1}}
	cursor = counts{Sectn: 1}

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 3}, cursor)

	rightSectn()
	updateCursorPos()
	assert.Equal(t, 1, cursor[Char])
	assert.Equal(t, 3, cursor[Sectn])

	rightSectn()
	updateCursorPos()
	assert.Equal(t, 1, cursor[Char])
	assert.Equal(t, 3, cursor[Sectn])
}

func TestLeft(t *testing.T) {
	document = []byte("1\f2\n3. 4 56")
	sections = []index{
		{chars: 1, bpara: []int{0}, cpara: []int{0}, csent: []int{0}, cword: []int{0}},
		{bsectn: 2, chars: 9, bpara: []int{2, 4}, cpara: []int{0, 2}, csent: []int{0, 2, 5}, cword: []int{0, 2, 5, 7}},
	}
	cursor = counts{9, 4, 3, 2, 2}
	ResizeScreen(margin+7, 4)

	scope = Char
	Left()
	updateCursorPos()
	assert.Equal(t, counts{8, 4, 3, 2, 2}, cursor)

	scope = Word
	Left()
	updateCursorPos()
	assert.Equal(t, counts{7, 3, 3, 2, 2}, cursor)

	scope = Sent
	Left()
	updateCursorPos()
	assert.Equal(t, counts{5, 2, 2, 2, 2}, cursor)

	scope = Para
	Left()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 2}, cursor)

	scope = Sectn
	Left()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)
}

func TestRight(t *testing.T) {
	document = []byte("12 3. 4\n5\f6")
	sections = []index{
		{chars: 9, bpara: []int{0, 8}, cpara: []int{0, 8}, csent: []int{0, 6, 8}, cword: []int{0, 3, 6, 8}},
		{bsectn: 10, chars: 1, bpara: []int{10}, cpara: []int{0}, csent: []int{0}, cword: []int{0}},
	}
	cursor = counts{Sectn: 1}
	ResizeScreen(margin+9, 5)

	scope = Char
	Right()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor)

	scope = Word
	Right()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 1}, cursor)

	scope = Sent
	Right()
	updateCursorPos()
	assert.Equal(t, counts{6, 2, 1, 1, 1}, cursor)

	scope = Para
	Right()
	updateCursorPos()
	assert.Equal(t, counts{8, 3, 2, 1, 1}, cursor)

	scope = Sectn
	Right()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)
}

func TestHome(t *testing.T) {
	document = []byte("1\f2\f3\n4")
	sections = []index{
		{chars: 1, bpara: []int{0}, cpara: []int{0}, csent: []int{0}, cword: []int{0}},
		{bsectn: 2, chars: 1, bpara: []int{2}, cpara: []int{0}, csent: []int{0}, cword: []int{0}},
		{bsectn: 4, chars: 3, bpara: []int{4, 6}, cpara: []int{0, 2}, csent: []int{0, 2}, cword: []int{0, 2}},
	}
	osectn = 0
	cursor = counts{3, 2, 2, 2, 3}
	ResizeScreen(margin+2, 4)

	scope = Char
	Home()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 3}, cursor)
	assert.Equal(t, Sent, scope)

	Home()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)
	assert.Equal(t, Para, scope)

	Home()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)
	assert.Equal(t, Sectn, scope)

	Home()
	updateCursorPos()
	assert.Equal(t, counts{3, 2, 2, 2, 3}, cursor)
	assert.Equal(t, Char, scope)

	cursor = counts{Sectn: 1}
	Home()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)
	assert.Equal(t, Sent, scope)

	Home()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)
	assert.Equal(t, Para, scope)

	cursor = counts{1, 1, 1, 1, 1}
	scope = Sectn
	Home()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor)

	cursor = counts{Sectn: 2}
	Home()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	cursor = counts{Sectn: 1}
	osectn = 1
	ochar = 1
	Home()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor)
}

func TestEnd(t *testing.T) {
	document = []byte("12\n3\f4\f5")
	sections = []index{
		{chars: 4, bpara: []int{0, 3}, cpara: []int{0, 3}, csent: []int{0, 3}, cword: []int{0, 3}},
		{bsectn: 5, chars: 1, bpara: []int{5}, cpara: []int{0}, csent: []int{0}, cword: []int{0}},
		{bsectn: 7, chars: 1, bpara: []int{7}, cpara: []int{0}, csent: []int{0}, cword: []int{0}},
	}
	osectn = 0
	cursor = counts{Sectn: 1}
	ResizeScreen(margin+3, 4)

	scope = Char
	End()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)
	assert.Equal(t, Sent, scope)

	End()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor)
	assert.Equal(t, Para, scope)

	End()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor)
	assert.Equal(t, Sectn, scope)

	End()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)
	assert.Equal(t, Char, scope)

	cursor = counts{1, 1, 1, 1, 1}
	End()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)

	scope = Char
	End()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor)

	scope = Char
	End()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor)

	scope = Sectn
	End()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor)

	cursor = counts{Sectn: 3}
	End()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 3}, cursor)

	cursor = counts{1, 1, 1, 1, 3}
	osectn = 3
	End()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 3}, cursor)
}
