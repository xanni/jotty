package edits

import (
	"testing"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/stretchr/testify/assert"
)

func setNewParagraph(pn int, t string) {
	doc.CreateParagraph(pn)
	doc.SetText(pn, t)
}

func resetIndex() {
	paras = nil
	total = counts{0, 0, 0, 1}
	indexPara()
}

func TestIndexWord(t *testing.T) {
	resetIndex()

	indexWord(1, 0)
	assert.Equal(t, []int{0}, paras[0].cword)

	indexWord(1, 1)
	assert.Equal(t, []int{0, 1}, paras[0].cword)
}

func TestIndexSent(t *testing.T) {
	resetIndex()

	indexSent(1, 0)
	assert.Equal(t, []int{0}, paras[0].csent)

	indexSent(1, 1)
	assert.Equal(t, []int{0, 1}, paras[0].csent)
}

func TestIndexPara(t *testing.T) {
	resetIndex()

	indexPara()
	assert.Equal(t, 2, len(paras))
}

func TestLeftChar(t *testing.T) {
	doc.SetText(1, "One")
	setNewParagraph(2, "Two")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{chars: 3}, {chars: 3}}
	cursor = counts{1, 0, 0, 2}

	leftChar()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)

	leftChar()
	assert.Equal(t, counts{3, 0, 0, 1}, cursor)

	cursor[Char] = 0
	leftChar()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)
}

func TestRightChar(t *testing.T) {
	doc.SetText(1, "One")
	setNewParagraph(2, "Two")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{chars: 3}, {chars: 3}}
	cursor = counts{2, 0, 0, 1}

	rightChar()
	assert.Equal(t, counts{3, 0, 0, 1}, cursor)

	rightChar()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)

	cursor[Char] = 3
	rightChar()
	assert.Equal(t, counts{3, 0, 0, 2}, cursor)
}

func TestLeftWord(t *testing.T) {
	doc.SetText(1, "1 23")
	doc.CreateParagraph(2)
	defer doc.DeleteParagraph(2)
	setNewParagraph(3, "4")
	defer doc.DeleteParagraph(3)
	paras = []ipara{{chars: 4, cword: []int{0, 2}}, {}, {chars: 1, cword: []int{0}}}
	cursor = counts{1, 1, 1, 3}

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 3}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 0, 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)

	cursor = counts{4, 2, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, 1, cursor[Word])

	cursor = counts{3, 2, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, 1, cursor[Word])
}

func TestRightWord(t *testing.T) {
	doc.SetText(1, "1")
	setNewParagraph(2, "23 4")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{chars: 1, cword: []int{0}}, {chars: 4, cword: []int{0, 3}}}
	cursor = counts{0, 0, 0, 1}

	rightWord()
	updateCursorPos()
	cursor = counts{0, 0, 0, 2}

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

	cursor = counts{1, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Word])

	cursor = counts{2, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Word])
}

func TestLeftSent(t *testing.T) {
	doc.SetText(1, "1. 23")
	doc.CreateParagraph(2)
	defer doc.DeleteParagraph(2)
	setNewParagraph(3, "4")
	defer doc.DeleteParagraph(3)
	paras = []ipara{{chars: 5, csent: []int{0, 3}}, {}, {chars: 1, csent: []int{0}}}
	cursor = counts{1, 1, 1, 3}

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 3}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)

	cursor = counts{4, 2, 2, 1}
	leftSent()
	updateCursorPos()
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])
}

func TestRightSent(t *testing.T) {
	doc.SetText(1, "1")
	setNewParagraph(2, "23. 4")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{chars: 1, csent: []int{0}}, {chars: 5, csent: []int{0, 4}}}
	cursor = counts{0, 0, 0, 1}

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)

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

	cursor = counts{1, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])

	cursor = counts{2, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 1, cursor[Sent])
}

func TestLeftPara(t *testing.T) {
	doc.SetText(1, "1")
	setNewParagraph(2, "23")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{chars: 1}, {chars: 2}}
	cursor = counts{1, 0, 0, 2}

	leftPara()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)

	leftPara()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)

	leftPara()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)

	cursor = counts{2, 0, 0, 2}
	leftPara()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)
}

func TestRightPara(t *testing.T) {
	doc.SetText(1, "1")
	setNewParagraph(2, "23")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{chars: 1}, {chars: 2}}
	cursor = counts{0, 0, 0, 1}

	rightPara()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)

	rightPara()
	assert.Equal(t, counts{2, 0, 0, 2}, cursor)

	rightPara()
	assert.Equal(t, counts{2, 0, 0, 2}, cursor)

	cursor = counts{1, 0, 0, 1}
	rightPara()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)
}

func TestLeft(t *testing.T) {
	doc.SetText(1, "1")
	setNewParagraph(2, "2. 3 45")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{[]int{0}, []int{0}, 1}, {[]int{0, 3}, []int{0, 3, 5}, 1}}
	cursor = counts{7, 3, 2, 2}
	ResizeScreen(margin+7, 4)

	scope = Char
	Left()
	updateCursorPos()
	assert.Equal(t, counts{6, 3, 2, 2}, cursor)

	scope = Word
	Left()
	updateCursorPos()
	assert.Equal(t, counts{5, 2, 2, 2}, cursor)

	scope = Sent
	Left()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 2}, cursor)

	scope = Para
	Left()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)

	Left()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)
}

func TestRight(t *testing.T) {
	doc.SetText(1, "12 3. 4")
	setNewParagraph(2, "5")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{[]int{0, 6}, []int{0, 3, 6}, 7}, {[]int{0}, []int{0}, 1}}
	cursor = counts{0, 0, 0, 1}
	ResizeScreen(margin+9, 5)

	scope = Char
	Right()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1}, cursor)

	scope = Word
	Right()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1}, cursor)

	scope = Sent
	Right()
	updateCursorPos()
	assert.Equal(t, counts{6, 2, 1, 1}, cursor)

	scope = Para
	Right()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)
}

func TestHome(t *testing.T) {
	doc.SetText(1, "1")
	setNewParagraph(2, "2")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{chars: 1}, {chars: 1}}

	ocursor = counts{}
	cursor = counts{1, 0, 0, 2}
	ResizeScreen(margin+2, 4)

	scope = Char
	Home()
	assert.Equal(t, counts{0, 0, 0, 2}, cursor)
	assert.Equal(t, Sent, scope)

	Home()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)
	assert.Equal(t, Para, scope)

	Home()
	assert.Equal(t, counts{1, 0, 0, 2}, cursor)
	assert.Equal(t, Char, scope)

	cursor = counts{0, 0, 0, 1}
	Home()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)
	assert.Equal(t, Sent, scope)

	Home()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)
	assert.Equal(t, Para, scope)

	cursor = counts{0, 0, 0, 1}
	ocursor = cursor
	Home()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)
}

func TestEnd(t *testing.T) {
	doc.SetText(1, "12")
	setNewParagraph(2, "3")
	defer doc.DeleteParagraph(2)
	paras = []ipara{{chars: 2}, {chars: 1}}

	ocursor = counts{}
	cursor = counts{0, 0, 0, 1}
	ResizeScreen(margin+3, 4)

	scope = Char
	End()
	assert.Equal(t, counts{2, 0, 0, 1}, cursor)
	assert.Equal(t, Sent, scope)

	End()
	assert.Equal(t, counts{1, 0, 0, 2}, cursor)
	assert.Equal(t, Para, scope)

	scope = Sent
	End()
	assert.Equal(t, counts{1, 0, 0, 2}, cursor)
	assert.Equal(t, Para, scope)

	End()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)
	assert.Equal(t, Char, scope)

	cursor = counts{1, 1, 1, 1}
	End()
	assert.Equal(t, counts{2, 1, 1, 1}, cursor)

	scope = Char
	End()
	assert.Equal(t, counts{2, 1, 1, 1}, cursor)
}
