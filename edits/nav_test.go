package edits

import (
	"testing"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/stretchr/testify/assert"
)

func setNewSection(sn int, t string) {
	doc.CreateSection(sn)
	doc.SetText(sn, 1, t)
}

func setNewParagraph(sn, pn int, t string) {
	doc.CreateParagraph(sn, pn)
	doc.SetText(sn, pn, t)
}

func resetIndex() {
	sections = nil
	indexSectn()
}

func TestIndexWord(t *testing.T) {
	resetIndex()

	indexWord(1, 1, 0)
	assert.Equal(t, []int{0}, sections[0].p[0].cword)

	indexWord(1, 1, 0)
	assert.Equal(t, []int{0}, sections[0].p[0].cword)

	indexWord(1, 1, 1)
	assert.Equal(t, []int{0, 1}, sections[0].p[0].cword)
}

func TestIndexSent(t *testing.T) {
	resetIndex()

	indexSent(1, 1, 0)
	assert.Equal(t, []int{0}, sections[0].p[0].csent)

	indexSent(1, 1, 0)
	assert.Equal(t, []int{0}, sections[0].p[0].csent)

	indexSent(1, 1, 1)
	assert.Equal(t, []int{0, 1}, sections[0].p[0].csent)
}

func TestIndexPara(t *testing.T) {
	resetIndex()

	indexPara(1)
	assert.Equal(t, 2, len(sections[0].p))
}

func TestIndexSectn(t *testing.T) {
	resetIndex()

	indexSectn()
	assert.Equal(t, 2, len(sections))
}

func TestLeftChar(t *testing.T) {
	doc.SetText(1, 1, "One")
	setNewSection(2, "Two")
	setNewParagraph(2, 2, "Three")
	sections = []isectn{{chars: 3, p: []ipara{{chars: 3}}}, {chars: 8, p: []ipara{{chars: 3}, {chars: 5}}}}
	cursor = counts{1, 0, 0, 2, 2}

	leftChar()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)

	leftChar()
	assert.Equal(t, counts{3, 0, 0, 1, 2}, cursor)

	cursor[Char] = 0
	leftChar()
	assert.Equal(t, counts{3, 0, 0, 1, 1}, cursor)

	cursor[Char] = 0
	leftChar()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	doc.CreateSection(3)
	cursor[Sectn] = 3
	leftChar()
	assert.Equal(t, counts{5, 0, 0, 2, 2}, cursor)

	doc.DeleteSection(3)
	doc.DeleteSection(2)
}

func TestRightChar(t *testing.T) {
	doc.SetText(1, 1, "One")
	setNewSection(2, "Two")
	setNewParagraph(2, 2, "Three")
	sections = []isectn{{chars: 3, p: []ipara{{chars: 3}}}, {chars: 8, p: []ipara{{chars: 3}, {chars: 5}}}}
	cursor = counts{2, 0, 0, 1, 1}

	rightChar()
	assert.Equal(t, counts{3, 0, 0, 1, 1}, cursor)

	rightChar()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	cursor[Char] = 3
	rightChar()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)

	cursor[Char] = 5
	rightChar()
	assert.Equal(t, counts{5, 0, 0, 2, 2}, cursor)

	doc.DeleteSection(2)
}

func TestLeftWord(t *testing.T) {
	doc.SetText(1, 1, "1 23")
	setNewSection(2, "4")
	setNewParagraph(2, 2, "5")
	sections = []isectn{
		{words: 2, chars: 4, p: []ipara{{chars: 4, cword: []int{0, 2}}}},
		{words: 2, chars: 2, p: []ipara{{chars: 1, cword: []int{0}}, {chars: 1, cword: []int{0}}}},
	}
	cursor = counts{1, 1, 1, 2, 2}

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 0, 1, 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

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

	doc.SetText(1, 1, "1")
	doc.CreateSection(2)
	doc.SetText(3, 1, "2")
	sections = []isectn{
		{chars: 1, p: []ipara{{chars: 1, cword: []int{0}}}},
		{p: []ipara{{}}},
		{chars: 1, p: []ipara{{chars: 1, cword: []int{0}}}},
	}
	cursor = counts{0, 0, 1, 1, 3}
	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	doc.DeleteSection(3)
	doc.DeleteSection(2)
}

func TestRightWord(t *testing.T) {
	doc.SetText(1, 1, "1")
	setNewSection(2, "23 4")
	sections = []isectn{
		{chars: 1, p: []ipara{{chars: 1, cword: []int{0}}}},
		{chars: 4, p: []ipara{{chars: 4, cword: []int{0, 3}}}},
	}
	cursor = counts{0, 0, 0, 1, 1}

	rightWord()
	updateCursorPos()
	cursor = counts{0, 0, 0, 1, 2}

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

	doc.DeleteSection(2)
}

func TestLeftSent(t *testing.T) {
	doc.SetText(1, 1, "1. 23")
	setNewSection(2, "4")
	doc.CreateParagraph(2, 2)
	setNewParagraph(2, 3, "5")
	sections = []isectn{
		{sents: 2, chars: 5, p: []ipara{{chars: 5, csent: []int{0, 3}}}},
		{sents: 2, chars: 2, p: []ipara{{chars: 1, csent: []int{0}}, {}, {chars: 1, csent: []int{0}}}},
	}
	cursor = counts{1, 1, 1, 3, 2}

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 3, 2}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	cursor = counts{4, 2, 2, 1, 1}
	leftSent()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])

	doc.DeleteSection(2)
}

func TestRightSent(t *testing.T) {
	doc.SetText(1, 1, "1")
	setNewSection(2, "23. 4")
	sections = []isectn{
		{chars: 1, p: []ipara{{chars: 1, csent: []int{0}}}},
		{chars: 5, p: []ipara{{chars: 5, csent: []int{0, 4}}}},
	}
	cursor = counts{0, 0, 0, 1, 1}

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

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

	doc.DeleteSection(2)
}

func TestLeftPara(t *testing.T) {
	doc.SetText(1, 1, "1")
	setNewParagraph(1, 2, "23")
	setNewSection(2, "4")
	sections = []isectn{
		{chars: 3, p: []ipara{{chars: 1}, {chars: 2}}},
		{chars: 1, p: []ipara{{chars: 1}}},
	}
	cursor = counts{1, 0, 0, 1, 2}

	leftPara()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	leftPara()
	assert.Equal(t, counts{0, 0, 0, 2, 1}, cursor)

	leftPara()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	leftPara()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	cursor = counts{3, 0, 0, 2, 1}
	leftPara()
	assert.Equal(t, counts{0, 0, 0, 2, 1}, cursor)

	doc.DeleteSection(2)
	doc.DeleteParagraph(1, 2)
}

func TestRightPara(t *testing.T) {
	doc.SetText(1, 1, "1")
	setNewSection(2, "23")
	setNewParagraph(2, 2, "4")
	sections = []isectn{
		{chars: 1, p: []ipara{{chars: 1}}},
		{chars: 3, p: []ipara{{chars: 2}, {chars: 1}}},
	}
	cursor = counts{0, 0, 0, 1, 1}

	rightPara()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	rightPara()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)

	rightPara()
	assert.Equal(t, counts{1, 0, 0, 2, 2}, cursor)

	rightPara()
	assert.Equal(t, counts{1, 0, 0, 2, 2}, cursor)

	cursor = counts{1, 0, 0, 1, 2}
	rightPara()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)

	cursor = counts{2, 0, 0, 1, 2}
	rightPara()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)

	doc.DeleteSection(2)
}

func TestLeftSectn(t *testing.T) {
	doc.SetText(1, 1, "1")
	setNewSection(2, "23")
	setNewSection(3, "4")
	sections = []isectn{
		{chars: 1, p: []ipara{{chars: 1}}},
		{chars: 2, p: []ipara{{chars: 2}}},
		{chars: 1, p: []ipara{{chars: 1}}},
	}
	cursor = counts{1, 0, 0, 1, 3}

	leftSectn()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	leftSectn()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	leftSectn()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	doc.DeleteSection(3)
	doc.DeleteSection(2)
}

func TestRightSectn(t *testing.T) {
	doc.SetText(1, 1, "1")
	setNewSection(2, "23")
	setNewSection(3, "4")
	sections = []isectn{
		{chars: 1, p: []ipara{{chars: 1}}},
		{chars: 2, p: []ipara{{chars: 2}}},
		{chars: 1, p: []ipara{{chars: 1}}},
	}
	cursor = counts{0, 0, 0, 1, 1}

	rightSectn()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	rightSectn()
	assert.Equal(t, counts{0, 0, 0, 1, 3}, cursor)

	rightSectn()
	assert.Equal(t, counts{1, 0, 0, 1, 3}, cursor)

	rightSectn()
	assert.Equal(t, counts{1, 0, 0, 1, 3}, cursor)

	doc.DeleteSection(3)
	doc.DeleteSection(2)
}

func TestLeft(t *testing.T) {
	doc.SetText(1, 1, "1")
	setNewSection(2, "2")
	setNewParagraph(2, 2, "3. 4 56")
	sections = []isectn{
		{1, 1, 1, []ipara{{[]int{0}, []int{0}, 1}}},
		{3, 4, 8, []ipara{{[]int{0}, []int{0}, 1}, {[]int{0, 3}, []int{0, 3, 5}, 1}}},
	}
	cursor = counts{7, 3, 2, 2, 2}
	ResizeScreen(margin+7, 4)

	scope = Char
	Left()
	updateCursorPos()
	assert.Equal(t, counts{6, 3, 2, 2, 2}, cursor)

	scope = Word
	Left()
	updateCursorPos()
	assert.Equal(t, counts{5, 2, 2, 2, 2}, cursor)

	scope = Sent
	Left()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 2, 2}, cursor)

	scope = Para
	Left()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)

	scope = Sectn
	Left()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	Left()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	doc.DeleteSection(2)
}

func TestRight(t *testing.T) {
	doc.SetText(1, 1, "12 3. 4")
	setNewParagraph(1, 2, "5")
	setNewSection(2, "6")
	sections = []isectn{
		{3, 4, 8, []ipara{{[]int{0, 6}, []int{0, 3, 6}, 7}, {[]int{0}, []int{0}, 1}}},
		{1, 1, 1, []ipara{{[]int{0}, []int{0}, 1}}},
	}
	cursor = counts{0, 0, 0, 1, 1}
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
	assert.Equal(t, counts{0, 0, 0, 2, 1}, cursor)

	scope = Sectn
	Right()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	doc.DeleteSection(2)
	doc.DeleteParagraph(1, 2)
}

func TestHome(t *testing.T) {
	doc.SetText(1, 1, "1")
	setNewSection(2, "2")
	setNewParagraph(2, 2, "3")
	sections = []isectn{{1, 1, 1, []ipara{{chars: 1}}}, {2, 2, 2, []ipara{{chars: 1}, {chars: 1}}}}

	ocursor = counts{}
	cursor = counts{1, 0, 0, 2, 2}
	ResizeScreen(margin+2, 4)

	scope = Char
	Home()
	assert.Equal(t, counts{0, 0, 0, 2, 2}, cursor)
	assert.Equal(t, Sent, scope)

	Home()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)
	assert.Equal(t, Para, scope)

	Home()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)
	assert.Equal(t, Sectn, scope)

	Home()
	assert.Equal(t, counts{1, 0, 0, 2, 2}, cursor)
	assert.Equal(t, Char, scope)

	cursor = counts{0, 0, 0, 1, 1}
	Home()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)
	assert.Equal(t, Sent, scope)

	Home()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)
	assert.Equal(t, Para, scope)

	cursor = counts{1, 1, 1, 1, 1}
	scope = Sectn
	Home()
	assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor)

	cursor = counts{0, 0, 0, 1, 2}
	Home()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)

	cursor = counts{0, 0, 0, 1, 1}
	ocursor = cursor
	Home()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)

	doc.DeleteSection(2)
}

func TestEnd(t *testing.T) {
	doc.SetText(1, 1, "12")
	setNewParagraph(1, 2, "3")
	setNewSection(2, "4")
	setNewSection(3, "5")
	setNewParagraph(3, 2, "6")
	sections = []isectn{
		{3, 2, 2, []ipara{{chars: 2}, {chars: 1}}},
		{1, 1, 1, []ipara{{chars: 1}}},
		{1, 1, 1, []ipara{{chars: 1}, {chars: 1}}},
	}

	ocursor = counts{}
	cursor = counts{0, 0, 0, 1, 1}
	ResizeScreen(margin+3, 4)

	scope = Char
	End()
	assert.Equal(t, counts{2, 0, 0, 1, 1}, cursor)
	assert.Equal(t, Sent, scope)

	End()
	assert.Equal(t, counts{1, 0, 0, 2, 1}, cursor)
	assert.Equal(t, Para, scope)

	End()
	assert.Equal(t, counts{1, 0, 0, 2, 3}, cursor)
	assert.Equal(t, Sectn, scope)

	End()
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)
	assert.Equal(t, Char, scope)

	cursor = counts{1, 1, 1, 1, 1}
	End()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)

	scope = Char
	End()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)

	scope = Sectn
	End()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)

	cursor = counts{0, 0, 1, 1, 3}
	ocursor = counts{0, 0, 0, 1, 1}
	End()
	assert.Equal(t, counts{0, 0, 1, 1, 3}, cursor)

	cursor = counts{0, 0, 1, 2, 3}
	End()
	assert.Equal(t, counts{0, 0, 1, 2, 3}, cursor)

	doc.DeleteSection(3)
	doc.DeleteSection(2)
	doc.DeleteParagraph(1, 2)
}
