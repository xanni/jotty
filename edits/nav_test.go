package edits

import (
	"testing"

	"git.sericyb.com.au/jotty/test"
	"github.com/stretchr/testify/assert"
)

func TestScanSectn(t *testing.T) {
	isectn = []int{0}
	cursor[Sectn] = 1
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
	isectn = []int{0, 3}
	document = []byte("One\fTwo")
	cursor[Sectn] = 1
	cursor[Char] = 2
	total[Sectn] = 2
	total[Char] = 3

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
	isectn = []int{0, 5}
	iword = []int{0}
	cursor = counts{1, 1, 1, 1, 2}

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	cursor = counts{4, 2, 1, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)

	cursor = counts{3, 2, 1, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)
}

func TestRightWord(t *testing.T) {
	document = []byte("1\f23 4")
	isectn = []int{0, 2}
	iword = []int{0}
	total = counts{1, 1, 1, 1, 2}
	cursor = counts{Sectn: 1}

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor)

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 1, 1, 2}, cursor)

	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 1, 1, 2}, cursor)

	cursor = counts{1, 1, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor)

	cursor = counts{2, 1, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor)
}

func TestLeftSent(t *testing.T) {
	document = []byte("1. 23\f4")
	isectn = []int{0, 6}
	isent = []index{{}, {3, 3}}
	cursor = counts{1, 1, 1, 1, 2}

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 1}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	cursor = counts{4, 2, 2, 1, 1}
	leftSent()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 1}, cursor)
}

func TestRightSent(t *testing.T) {
	document = []byte("1\f23. 4")
	isectn = []int{0, 2}
	isent = []index{{}}
	total = counts{1, 1, 1, 1, 2}
	cursor = counts{Sectn: 1}

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{4, 1, 1, 1, 2}, cursor)

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{5, 2, 2, 1, 2}, cursor)

	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{5, 2, 2, 1, 2}, cursor)

	cursor = counts{1, 1, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{4, 1, 1, 1, 2}, cursor)

	cursor = counts{2, 1, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(t, counts{4, 1, 1, 1, 2}, cursor)
}

func TestLeftPara(t *testing.T) {
	document = []byte("1\n23\f4")
	isectn = []int{0, 5}
	ipara = []index{{}, {1, 1}}
	cursor = counts{1, 1, 1, 1, 2}

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 1}, cursor)

	cursor = counts{3, 2, 2, 2, 1}
	leftPara()
	updateCursorPos()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)
}

func TestRightPara(t *testing.T) {
	document = []byte("1\f23\n4")
	isectn = []int{0, 2}
	ipara = []index{{}}
	total = counts{1, 1, 1, 1, 2}
	cursor = counts{Sectn: 1}

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor)

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 2, 2, 2}, cursor)

	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{4, 2, 2, 2, 2}, cursor)

	cursor = counts{1, 1, 1, 1, 2}
	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor)

	cursor = counts{2, 1, 1, 1, 2}
	rightPara()
	updateCursorPos()
	assert.Equal(t, counts{3, 1, 1, 1, 2}, cursor)
}

func TestLeftSectn(t *testing.T) {
	document = []byte("1\f23\f4")
	isectn = []int{0, 3, 5}
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
	isectn = []int{0, 2, 5}
	total = counts{1, 1, 1, 1, 3}
	cursor = counts{Sectn: 1}

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 2}, cursor)

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{Sectn: 3}, cursor)

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor)

	rightSectn()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor)
}

func TestLeft(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("1\f2\n3. 4 56")
		isectn = []int{0, 2}
		cursor = counts{9, 4, 3, 2, 2}
		scanSectn()
		Sx = margin + 7
		Sy = 4
		ResizeScreen()

		scope = Char
		Left()
		assert.Equal(t, counts{8, 4, 3, 2, 2}, cursor)

		scope = Word
		Left()
		assert.Equal(t, counts{7, 3, 3, 2, 2}, cursor)

		scope = Sent
		Left()
		assert.Equal(t, counts{5, 2, 2, 2, 2}, cursor)

		scope = Para
		Left()
		assert.Equal(t, counts{2, 1, 1, 1, 2}, cursor)

		scope = Sectn
		Left()
		assert.Equal(t, counts{Sectn: 1}, cursor)
	})
}

func TestRight(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("12 3. 4\n5\f6")
		isectn = []int{0, 10}
		cursor = counts{Sectn: 1}
		scanSectn()
		Sx = margin + 9
		Sy = 5
		ResizeScreen()

		scope = Char
		Right()
		assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor)

		scope = Word
		Right()
		assert.Equal(t, counts{3, 1, 1, 1, 1}, cursor)

		scope = Sent
		Right()
		assert.Equal(t, counts{6, 2, 1, 1, 1}, cursor)

		scope = Para
		Right()
		assert.Equal(t, counts{8, 3, 2, 1, 1}, cursor)

		scope = Sectn
		Right()
		assert.Equal(t, counts{Sectn: 2}, cursor)
	})
}

func TestHome(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("1\f2\f3\n4")
		isectn = []int{0, 2, 4}
		osectn = 0
		cursor = counts{3, 2, 2, 2, 3}
		scanSectn()
		Sx = margin + 2
		Sy = 4
		ResizeScreen()

		scope = Char
		Home()
		assert.Equal(t, counts{2, 1, 1, 1, 3}, cursor)
		assert.Equal(t, Sent, scope)

		Home()
		assert.Equal(t, counts{Sectn: 2}, cursor)
		assert.Equal(t, Para, scope)

		Home()
		assert.Equal(t, counts{Sectn: 1}, cursor)
		assert.Equal(t, Sectn, scope)

		Home()
		assert.Equal(t, counts{3, 2, 2, 2, 3}, cursor)
		assert.Equal(t, Char, scope)

		cursor = counts{Sectn: 1}
		Home()
		assert.Equal(t, counts{Sectn: 1}, cursor)
		assert.Equal(t, Sent, scope)

		Home()
		assert.Equal(t, counts{Sectn: 1}, cursor)
		assert.Equal(t, Para, scope)

		cursor = counts{1, 1, 1, 1, 1}
		scope = Sectn
		Home()
		assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor)

		cursor = counts{Sectn: 2}
		Home()
		assert.Equal(t, counts{Sectn: 2}, cursor)

		cursor = counts{Sectn: 1}
		osectn = 1
		ochar = 1
		Home()
		assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor)
	})
}

func TestEnd(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("12\n3\f4\f5")
		isectn = []int{0, 5, 7}
		osectn = 0
		cursor = counts{Sectn: 1}
		scanSectn()
		Sx = margin + 3
		Sy = 4
		ResizeScreen()

		scope = Char
		End()
		assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)
		assert.Equal(t, Sent, scope)

		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor)
		assert.Equal(t, Para, scope)

		End()
		assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor)
		assert.Equal(t, Sectn, scope)

		End()
		assert.Equal(t, counts{Sectn: 1}, cursor)
		assert.Equal(t, Char, scope)

		cursor = counts{1, 1, 1, 1, 1}
		scanSectn()
		End()
		assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor)

		scope = Char
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor)

		scope = Char
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor)

		scope = Sectn
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor)

		cursor = counts{Sectn: 3}
		scanSectn()
		End()
		assert.Equal(t, counts{Sectn: 3}, cursor)

		cursor = counts{1, 1, 1, 1, 3}
		osectn = 3
		End()
		assert.Equal(t, counts{Sectn: 3}, cursor)
	})
}
