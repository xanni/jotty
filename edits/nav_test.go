package edits

import (
	"testing"

	"git.sericyb.com.au/jotty/test"
	"github.com/stretchr/testify/assert"
)

func TestScanSectn(t *testing.T) {
	isectn = []int{0}
	cursor.pos[Sectn] = 1
	document = nil
	scanSectn()
	assert.Equal(t, counts{0, 0, 1, 1, 1}, total)

	document = []byte("Six words, two sentences.\nTwo paragraphs.")
	scanSectn()
	assert.Equal(t, counts{41, 6, 2, 2, 1}, total)

	isectn = append(isectn, 42)
	document = append(document, []byte("\fAnother section.")...)
	scanSectn()
	assert.Equal(t, counts{41, 6, 2, 2, 2}, total)

	isectn = []int{0}
	document = []byte{'1', 255, '2'}
	scanSectn()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, total)
}

func TestLeftChar(t *testing.T) {
	isectn = []int{0, 3}
	document = []byte("One\fTwo")
	cursor.pos[Sectn] = 2
	cursor.pos[Char] = 1

	leftChar()
	assert.Equal(t, 2, cursor.pos[Sectn])
	assert.Equal(t, 0, cursor.pos[Char])

	leftChar()
	assert.Equal(t, 1, cursor.pos[Sectn])
	assert.Equal(t, 3, cursor.pos[Char])

	cursor.pos[Char] = 1
	leftChar()
	assert.Equal(t, 1, cursor.pos[Sectn])
	assert.Equal(t, 0, cursor.pos[Char])

	leftChar()
	assert.Equal(t, 1, cursor.pos[Sectn])
	assert.Equal(t, 0, cursor.pos[Char])
}

func TestRightChar(t *testing.T) {
	isectn = []int{0, 3}
	document = []byte("One\fTwo")
	cursor.pos[Sectn] = 1
	cursor.pos[Char] = 2
	total[Sectn] = 2
	total[Char] = 3

	rightChar()
	assert.Equal(t, 1, cursor.pos[Sectn])
	assert.Equal(t, 3, cursor.pos[Char])

	rightChar()
	assert.Equal(t, 2, cursor.pos[Sectn])
	assert.Equal(t, 0, cursor.pos[Char])

	cursor.pos[Char] = 3
	rightChar()
	assert.Equal(t, 2, cursor.pos[Sectn])
	assert.Equal(t, 3, cursor.pos[Char])
}

func TestLeftWord(t *testing.T) {
	document = []byte("1 23\f4")
	isectn = []int{0, 5}
	iword = []int{0}
	cursor.pos = counts{1, 1, 1, 1, 2}

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor.pos)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)

	cursor.pos = counts{4, 2, 1, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor.pos)

	cursor.pos = counts{3, 2, 1, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor.pos)
}

func TestRightWord(t *testing.T) {
	document = []byte("1\f23 4")
	isectn = []int{0, 2}
	iword = []int{0}
	total = counts{1, 1, 1, 1, 2}
	cursor.pos = counts{0, 0, 0, 0, 1}

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor.pos)

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 1, 1, 2}, cursor.pos)

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 1, 1, 2}, cursor.pos)

	cursor.pos = counts{1, 1, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor.pos)

	cursor.pos = counts{2, 1, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor.pos)
}

func TestLeftSent(t *testing.T) {
	document = []byte("1. 23\f4")
	isectn = []int{0, 6}
	isent = []index{{}, {3, 3}}
	cursor.pos = counts{1, 1, 1, 1, 2}

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 1}, cursor.pos)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)

	cursor.pos = counts{4, 2, 2, 1, 1}
	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 1}, cursor.pos)
}

func TestRightSent(t *testing.T) {
	document = []byte("1\f23. 4")
	isectn = []int{0, 2}
	isent = []index{{}}
	total = counts{1, 1, 1, 1, 2}
	cursor.pos = counts{0, 0, 0, 0, 1}

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{4, 1, 1, 1, 2}, cursor.pos)

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{5, 2, 2, 1, 2}, cursor.pos)

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{5, 2, 2, 1, 2}, cursor.pos)

	cursor.pos = counts{1, 1, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{4, 1, 1, 1, 2}, cursor.pos)

	cursor.pos = counts{2, 1, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{4, 1, 1, 1, 2}, cursor.pos)
}

func TestLeftPara(t *testing.T) {
	document = []byte("1\n23\f4")
	isectn = []int{0, 5}
	ipara = []index{{}, {1, 1}}
	cursor.pos = counts{1, 1, 1, 1, 2}

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor.pos)

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)

	cursor.pos = counts{3, 2, 2, 2, 1}
	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor.pos)
}

func TestRightPara(t *testing.T) {
	document = []byte("1\f23\n4")
	isectn = []int{0, 2}
	ipara = []index{{}}
	total = counts{1, 1, 1, 1, 2}
	cursor.pos = counts{0, 0, 0, 0, 1}

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor.pos)

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 2, 2, 2}, cursor.pos)

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 2, 2, 2}, cursor.pos)

	cursor.pos = counts{1, 1, 1, 1, 2}
	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor.pos)

	cursor.pos = counts{2, 1, 1, 1, 2}
	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor.pos)
}

func TestLeftSectn(t *testing.T) {
	document = []byte("1\f23\f4")
	isectn = []int{0, 3, 5}
	cursor.pos = counts{1, 1, 1, 1, 3}

	leftSectn()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	leftSectn()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)

	leftSectn()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
}

func TestRightSectn(t *testing.T) {
	document = []byte("1\f23\f4")
	isectn = []int{0, 2, 5}
	total = counts{1, 1, 1, 1, 3}
	cursor.pos = counts{0, 0, 0, 0, 1}

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 3}, cursor.pos)

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor.pos)

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor.pos)
}

func TestLeft(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("1\f2\n3. 4 56")
		isectn = []int{0, 2}
		cursor.pos = counts{9, 4, 3, 2, 2}
		scanSectn()
		Sx = margin + 1
		Sy = 4
		ResizeScreen()

		scope = Char
		Left()
		assert.Equal(t, counts{8, 4, 3, 2, 2}, cursor.pos)

		scope = Word
		Left()
		assert.Equal(t, counts{7, 3, 3, 2, 2}, cursor.pos)

		scope = Sent
		Left()
		assert.Equal(t, counts{5, 2, 2, 2, 2}, cursor.pos)

		scope = Para
		Left()
		assert.Equal(t, counts{2, 1, 1, 1, 2}, cursor.pos)

		scope = Sectn
		Left()
		assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
	})
}

func TestRight(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("12 3. 4\n5\f6")
		isectn = []int{0, 10}
		cursor.pos = counts{0, 0, 0, 0, 1}
		scanSectn()
		Sx = margin + 1
		Sy = 4
		ResizeScreen()

		scope = Char
		Right()
		assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor.pos)

		scope = Word
		Right()
		assert.Equal(t, counts{3, 1, 1, 1, 1}, cursor.pos)

		scope = Sent
		Right()
		assert.Equal(t, counts{6, 2, 1, 1, 1}, cursor.pos)

		scope = Para
		Right()
		assert.Equal(t, counts{8, 3, 2, 1, 1}, cursor.pos)

		scope = Sectn
		Right()
		assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)
	})
}

func TestHome(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("1\f2\f3\n4")
		isectn = []int{0, 2, 4}
		osectn = 0
		cursor.pos = counts{3, 2, 2, 2, 3}
		scanSectn()
		Sx = margin + 1
		Sy = 4
		ResizeScreen()

		scope = Char
		Home()
		assert.Equal(t, counts{2, 1, 1, 1, 3}, cursor.pos)
		assert.Equal(t, Sent, scope)

		Home()
		assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)
		assert.Equal(t, Para, scope)

		Home()
		assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
		assert.Equal(t, Sectn, scope)

		Home()
		assert.Equal(t, counts{3, 2, 2, 2, 3}, cursor.pos)
		assert.Equal(t, Char, scope)

		cursor.pos = counts{0, 0, 0, 0, 1}
		Home()
		assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
		assert.Equal(t, Sent, scope)

		Home()
		assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
		assert.Equal(t, Para, scope)

		cursor.pos = counts{1, 1, 1, 1, 1}
		scope = Sectn
		Home()
		assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor.pos)

		cursor.pos = counts{0, 0, 0, 0, 2}
		Home()
		assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

		cursor.pos = counts{0, 0, 0, 0, 1}
		osectn = 1
		ochar = 1
		Home()
		assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor.pos)
	})
}

func TestEnd(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("12\n3\f4\f5")
		isectn = []int{0, 5, 7}
		osectn = 0
		cursor.pos = counts{0, 0, 0, 0, 1}
		scanSectn()
		Sx = margin + 1
		Sy = 4
		ResizeScreen()

		scope = Char
		End()
		assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor.pos)
		assert.Equal(t, Sent, scope)

		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor.pos)
		assert.Equal(t, Para, scope)

		End()
		assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor.pos)
		assert.Equal(t, Sectn, scope)

		End()
		assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
		assert.Equal(t, Char, scope)

		cursor.pos = counts{1, 1, 1, 1, 1}
		scanSectn()
		End()
		assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor.pos)

		scope = Char
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor.pos)

		scope = Char
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor.pos)

		scope = Sectn
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor.pos)

		cursor.pos = counts{0, 0, 0, 0, 3}
		scanSectn()
		End()
		assert.Equal(t, counts{0, 0, 0, 0, 3}, cursor.pos)

		cursor.pos = counts{1, 1, 1, 1, 3}
		osectn = 3
		End()
		assert.Equal(t, counts{0, 0, 0, 0, 3}, cursor.pos)
	})
}
