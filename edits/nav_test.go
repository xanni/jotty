package edits

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	ps "github.com/xanni/jotty/permascroll"
)

func init() {
	if err := ps.OpenPermascroll(os.DevNull); err != nil {
		panic(err)
	}
}

func TestLeftChar(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.AppendText(1, "One")
	ps.AppendText(2, "Two")
	cache = []para{{chars: 3}, {chars: 3}}
	cursor = counts{1, 0, 0, 2}

	leftChar()
	assert.Equal(counts{0, 0, 0, 2}, cursor)

	leftChar()
	assert.Equal(counts{3, 0, 0, 1}, cursor)

	cursor[Char] = 0
	leftChar()
	assert.Equal(counts{0, 0, 0, 1}, cursor)
}

func TestRightChar(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.AppendText(1, "One")
	ps.AppendText(2, "Two")
	cache = []para{{chars: 3}, {chars: 3}}
	cursor = counts{2, 0, 0, 1}

	rightChar()
	assert.Equal(counts{3, 0, 0, 1}, cursor)

	rightChar()
	assert.Equal(counts{0, 0, 0, 2}, cursor)

	cursor[Char] = 3
	rightChar()
	assert.Equal(counts{3, 0, 0, 2}, cursor)
}

func TestLeftWord(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.SplitParagraph(2, 0)
	ps.AppendText(1, "1 23")
	ps.AppendText(3, "4")
	cache = []para{{chars: 4, cword: []int{0, 2}}, {}, {chars: 1, cword: []int{0}}}
	cursor = counts{1, 1, 1, 3}

	leftWord()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 3}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 2}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(counts{2, 1, 0, 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 1}, cursor)

	leftWord()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 1}, cursor)

	cursor = counts{4, 2, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(2, cursor[Char])
	assert.Equal(1, cursor[Word])

	cursor = counts{3, 2, 1, 1}
	leftWord()
	updateCursorPos()
	assert.Equal(2, cursor[Char])
	assert.Equal(1, cursor[Word])
}

func TestRightWord(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.AppendText(1, "1")
	ps.AppendText(2, "23 4")
	cache = []para{{chars: 1, cword: []int{0}}, {chars: 4, cword: []int{0, 3}}}
	cursor = counts{0, 0, 0, 1}

	rightWord()
	updateCursorPos()
	cursor = counts{0, 0, 0, 2}

	rightWord()
	updateCursorPos()
	assert.Equal(3, cursor[Char])
	assert.Equal(1, cursor[Word])

	rightWord()
	updateCursorPos()
	assert.Equal(4, cursor[Char])
	assert.Equal(2, cursor[Word])

	rightWord()
	updateCursorPos()
	assert.Equal(4, cursor[Char])
	assert.Equal(2, cursor[Word])

	cursor = counts{1, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(3, cursor[Char])
	assert.Equal(1, cursor[Word])

	cursor = counts{2, 1, 1, 2}
	rightWord()
	updateCursorPos()
	assert.Equal(3, cursor[Char])
	assert.Equal(1, cursor[Word])
}

func TestLeftSent(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.SplitParagraph(2, 0)
	ps.AppendText(1, "1. 23")
	ps.AppendText(3, "4")
	cache = []para{{chars: 5, csent: []int{0, 3}}, {}, {chars: 1, csent: []int{0}}}
	cursor = counts{1, 1, 1, 3}

	leftSent()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 3}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 2}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(3, cursor[Char])
	assert.Equal(1, cursor[Sent])

	leftSent()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 1}, cursor)

	leftSent()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 1}, cursor)

	cursor = counts{4, 2, 2, 1}
	leftSent()
	updateCursorPos()
	assert.Equal(3, cursor[Char])
	assert.Equal(1, cursor[Sent])
}

func TestRightSent(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.AppendText(1, "1")
	ps.AppendText(2, "23. 4")
	cache = []para{{chars: 1, csent: []int{0}}, {chars: 5, csent: []int{0, 4}}}
	cursor = counts{0, 0, 0, 1}

	rightSent()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 2}, cursor)

	rightSent()
	updateCursorPos()
	assert.Equal(4, cursor[Char])
	assert.Equal(1, cursor[Sent])

	rightSent()
	updateCursorPos()
	assert.Equal(5, cursor[Char])
	assert.Equal(2, cursor[Sent])

	rightSent()
	updateCursorPos()
	assert.Equal(5, cursor[Char])
	assert.Equal(2, cursor[Sent])

	cursor = counts{1, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(4, cursor[Char])
	assert.Equal(1, cursor[Sent])

	cursor = counts{2, 1, 1, 2}
	rightSent()
	updateCursorPos()
	assert.Equal(4, cursor[Char])
	assert.Equal(1, cursor[Sent])
}

func TestLeftPara(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.AppendText(1, "1")
	ps.AppendText(2, "23")
	cache = []para{{chars: 1}, {chars: 2}}
	cursor = counts{1, 0, 0, 2}

	leftPara()
	assert.Equal(counts{0, 0, 0, 2}, cursor)

	leftPara()
	assert.Equal(counts{0, 0, 0, 1}, cursor)

	leftPara()
	assert.Equal(counts{0, 0, 0, 1}, cursor)

	cursor = counts{2, 0, 0, 2}
	leftPara()
	assert.Equal(counts{0, 0, 0, 2}, cursor)
}

func TestRightPara(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.AppendText(1, "1")
	ps.AppendText(2, "23")
	cache = []para{{chars: 1}, {chars: 2}}
	cursor = counts{0, 0, 0, 1}

	rightPara()
	assert.Equal(counts{0, 0, 0, 2}, cursor)

	rightPara()
	assert.Equal(counts{2, 0, 0, 2}, cursor)

	rightPara()
	assert.Equal(counts{2, 0, 0, 2}, cursor)

	cursor = counts{1, 0, 0, 1}
	rightPara()
	assert.Equal(counts{0, 0, 0, 2}, cursor)
}

func TestLeft(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+7, 4)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.AppendText(1, "1")
	ps.AppendText(2, "2. 3 45")
	cache = []para{{1, []int{0}, []int{0}, nil}, {1, []int{0, 3, 5}, []int{0, 3}, nil}}
	cursor = counts{7, 3, 2, 2}

	scope = Char
	Left()
	updateCursorPos()
	assert.Equal(counts{6, 3, 2, 2}, cursor)

	scope = Word
	Left()
	updateCursorPos()
	assert.Equal(counts{5, 2, 2, 2}, cursor)

	scope = Sent
	Left()
	updateCursorPos()
	assert.Equal(counts{3, 1, 1, 2}, cursor)

	scope = Para
	Left()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 2}, cursor)

	Left()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 1}, cursor)
}

func TestRight(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+9, 5)
	ps.Init()
	ps.SplitParagraph(1, 0)
	ps.AppendText(1, "12 3. 4")
	ps.AppendText(2, "5")
	cache = []para{{7, []int{0, 3, 6}, []int{0, 6}, nil}, {1, []int{0}, []int{0}, nil}}
	cursor = counts{0, 0, 0, 1}

	scope = Char
	Right()
	updateCursorPos()
	assert.Equal(counts{1, 1, 1, 1}, cursor)

	scope = Word
	Right()
	updateCursorPos()
	assert.Equal(counts{3, 1, 1, 1}, cursor)

	scope = Sent
	Right()
	updateCursorPos()
	assert.Equal(counts{6, 2, 1, 1}, cursor)

	scope = Para
	Right()
	updateCursorPos()
	assert.Equal(counts{0, 0, 0, 2}, cursor)
}

func TestHome(t *testing.T) {
	assert := assert.New(t)
	ps.Init()
	ps.AppendText(1, "12")
	ps.SplitParagraph(1, 1)

	ocursor = counts{}
	cursor = counts{1, 0, 0, 2}
	ResizeScreen(margin+2, 4)

	scope = Char
	Home()
	assert.Equal(counts{0, 0, 0, 2}, cursor)
	assert.Equal(Sent, scope)

	Home()
	assert.Equal(counts{0, 0, 0, 1}, cursor)
	assert.Equal(Para, scope)

	Home()
	assert.Equal(counts{1, 0, 0, 2}, cursor)
	assert.Equal(Char, scope)

	cursor = counts{0, 0, 0, 1}
	Home()
	assert.Equal(counts{0, 0, 0, 1}, cursor)
	assert.Equal(Sent, scope)

	Home()
	assert.Equal(counts{0, 0, 0, 1}, cursor)
	assert.Equal(Para, scope)

	cursor = counts{0, 0, 0, 1}
	ocursor = cursor
	Home()
	assert.Equal(counts{0, 0, 0, 1}, cursor)
}

func TestEnd(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+3, 4)
	ps.Init()
	ps.AppendText(1, "123")
	ps.SplitParagraph(1, 2)
	cache = []para{{chars: 2}, {chars: 1}}

	ocursor = counts{}
	cursor = counts{0, 0, 0, 1}

	scope = Char
	End()
	assert.Equal(counts{2, 0, 0, 1}, cursor)
	assert.Equal(Sent, scope)

	End()
	assert.Equal(counts{1, 0, 0, 2}, cursor)
	assert.Equal(Para, scope)

	scope = Sent
	End()
	assert.Equal(counts{1, 0, 0, 2}, cursor)
	assert.Equal(Para, scope)

	End()
	assert.Equal(counts{0, 0, 0, 1}, cursor)
	assert.Equal(Char, scope)

	cursor = counts{1, 1, 1, 1}
	End()
	assert.Equal(counts{2, 1, 1, 1}, cursor)

	scope = Char
	End()
	assert.Equal(counts{2, 1, 1, 1}, cursor)
}
