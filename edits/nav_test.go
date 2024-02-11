package edits

import (
	"testing"

	"git.sericyb.com.au/jotty/test"
	"github.com/stretchr/testify/assert"
)

func TestScanSect(t *testing.T) {
	isect = []int{0}
	cursor.pos[Sect] = 1
	document = nil
	scanSect()
	assert.Equal(t, counts{0, 0, 1, 1, 1}, total)

	document = []byte("Six words, two sentences.\nTwo paragraphs.")
	scanSect()
	assert.Equal(t, counts{41, 6, 2, 2, 1}, total)

	isect = append(isect, 42)
	document = append(document, []byte("\fAnother section.")...)
	scanSect()
	assert.Equal(t, counts{41, 6, 2, 2, 2}, total)

	isect = []int{0}
	document = []byte{'1', 255, '2'}
	scanSect()
	assert.Equal(t, counts{2, 1, 1, 1, 1}, total)
}

func TestLeftChar(t *testing.T) {
	isect = []int{0, 3}
	document = []byte("One\fTwo")
	cursor.pos[Sect] = 2
	cursor.pos[Char] = 1

	leftChar()
	assert.Equal(t, 2, cursor.pos[Sect])
	assert.Equal(t, 0, cursor.pos[Char])

	leftChar()
	assert.Equal(t, 1, cursor.pos[Sect])
	assert.Equal(t, 3, cursor.pos[Char])

	cursor.pos[Char] = 1
	leftChar()
	assert.Equal(t, 1, cursor.pos[Sect])
	assert.Equal(t, 0, cursor.pos[Char])

	leftChar()
	assert.Equal(t, 1, cursor.pos[Sect])
	assert.Equal(t, 0, cursor.pos[Char])
}

func TestRightChar(t *testing.T) {
	isect = []int{0, 3}
	document = []byte("One\fTwo")
	cursor.pos[Sect] = 1
	cursor.pos[Char] = 2
	total[Sect] = 2
	total[Char] = 3

	rightChar()
	assert.Equal(t, 1, cursor.pos[Sect])
	assert.Equal(t, 3, cursor.pos[Char])

	rightChar()
	assert.Equal(t, 2, cursor.pos[Sect])
	assert.Equal(t, 0, cursor.pos[Char])

	cursor.pos[Char] = 3
	rightChar()
	assert.Equal(t, 2, cursor.pos[Sect])
	assert.Equal(t, 3, cursor.pos[Char])
}

func TestLeftWord(t *testing.T) {
	document = []byte("1 23\f4")
	isect = []int{0, 5}
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
	isect = []int{0, 2}
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
	isect = []int{0, 6}
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
	isect = []int{0, 2}
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
	isect = []int{0, 5}
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
	isect = []int{0, 2}
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

func TestLeftSect(t *testing.T) {
	document = []byte("1\f23\f4")
	isect = []int{0, 3, 5}
	cursor.pos = counts{1, 1, 1, 1, 3}

	leftSect()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	leftSect()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)

	leftSect()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
}

func TestRightSect(t *testing.T) {
	document = []byte("1\f23\f4")
	isect = []int{0, 2, 5}
	total = counts{1, 1, 1, 1, 3}
	cursor.pos = counts{0, 0, 0, 0, 1}

	rightSect()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

	rightSect()
	updateCursorPos()
	assert.Equal(t, counts{0, 0, 0, 0, 3}, cursor.pos)

	rightSect()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor.pos)

	rightSect()
	updateCursorPos()
	assert.Equal(t, counts{1, 1, 1, 1, 3}, cursor.pos)
}

func TestLeft(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("1\f2\n3. 4 56")
		isect = []int{0, 2}
		cursor.pos = counts{9, 4, 3, 2, 2}
		scanSect()
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

		scope = Sect
		Left()
		assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
	})
}

func TestRight(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("12 3. 4\n5\f6")
		isect = []int{0, 10}
		cursor.pos = counts{0, 0, 0, 0, 1}
		scanSect()
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

		scope = Sect
		Right()
		assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)
	})
}

func TestHome(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("1\f2\f3\n4")
		isect = []int{0, 2, 4}
		osect = 0
		cursor.pos = counts{3, 2, 2, 2, 3}
		scanSect()
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
		assert.Equal(t, Sect, scope)

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
		scope = Sect
		Home()
		assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor.pos)

		cursor.pos = counts{0, 0, 0, 0, 2}
		Home()
		assert.Equal(t, counts{0, 0, 0, 0, 2}, cursor.pos)

		cursor.pos = counts{0, 0, 0, 0, 1}
		osect = 1
		ochar = 1
		Home()
		assert.Equal(t, counts{1, 1, 1, 1, 1}, cursor.pos)
	})
}

func TestEnd(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte("12\n3\f4\f5")
		isect = []int{0, 5, 7}
		osect = 0
		cursor.pos = counts{0, 0, 0, 0, 1}
		scanSect()
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
		assert.Equal(t, Sect, scope)

		End()
		assert.Equal(t, counts{0, 0, 0, 0, 1}, cursor.pos)
		assert.Equal(t, Char, scope)

		cursor.pos = counts{1, 1, 1, 1, 1}
		scanSect()
		End()
		assert.Equal(t, counts{2, 1, 1, 1, 1}, cursor.pos)

		scope = Char
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor.pos)

		scope = Char
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor.pos)

		scope = Sect
		End()
		assert.Equal(t, counts{4, 2, 2, 2, 1}, cursor.pos)

		cursor.pos = counts{0, 0, 0, 0, 3}
		scanSect()
		End()
		assert.Equal(t, counts{0, 0, 0, 0, 3}, cursor.pos)

		cursor.pos = counts{1, 1, 1, 1, 3}
		osect = 3
		End()
		assert.Equal(t, counts{0, 0, 0, 0, 3}, cursor.pos)
	})
}
