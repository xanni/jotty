package edits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendParaBreak(t *testing.T) {
	setupTest()
	document = []byte(" ")
	cursor[Char]++
	ResizeScreen(margin+1, 4)

	appendParaBreak()
	drawWindow()
	assert.Equal(t, 2, cursy)
	assert.Equal(t, []byte("\n"), document)

	appendParaBreak()
	assert.Equal(t, []byte("\n\n"), document)
}

func TestAppendSectnBreak(t *testing.T) {
	setupTest()
	document = []byte("\n")
	cursor[Char] = 1
	ResizeScreen(margin+1, 5)

	appendSectnBreak()
	assert.Equal(t, []byte("\f"), document)

	appendSectnBreak()
	assert.Equal(t, []byte("\f\f"), document)

	cursor = counts{1, 0, 1, 1, 1}
	document = []byte("\n")
	resetIndex()
	newBuffer()
	drawWindow()
	appendSectnBreak()
	assert.Equal(t, []byte("\f"), document)
}

func TestAppendRunes(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 4)

	AppendRunes([]rune("•"))
	assert.Equal(t, []byte("•"), document)
	assert.Equal(t, Char, scope)
	assert.Equal(t, 1, cursor[Char])

	AppendRunes([]rune("ü"))
	assert.Equal(t, []byte("•ü"), document)
	assert.Equal(t, 2, cursor[Char])

	initialCap = true
	AppendRunes([]rune("å"))
	drawWindow()
	assert.Equal(t, []byte("•üÅ"), document)
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, 3, sections[0].chars)
	assert.Equal(t, []int{1}, sections[0].cword)

	AppendRunes([]rune{'\n'})
	drawWindow()
	assert.Equal(t, []byte("•üÅ\n"), document)
	assert.Equal(t, Char, scope)
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, 4, sections[0].chars)
	assert.Equal(t, []int{1}, sections[0].cword)
	assert.Equal(t, []int{0, 4}, sections[0].csent)
	assert.Equal(t, []int{0, 4}, sections[0].cpara)

	AppendRunes([]rune{'X'})
	drawWindow()
	assert.Equal(t, []byte("•üÅ\nX"), document)
	assert.Equal(t, 5, cursor[Char])
	assert.Equal(t, 5, sections[0].chars)
	assert.Equal(t, []int{1, 4}, sections[0].cword)

	appendSectnBreak()
	drawWindow()
	AppendRunes([]rune{'Y'})
	drawWindow()
	assert.Equal(t, []byte("•üÅ\nX\fY"), document)
	assert.Equal(t, Char, scope)
	assert.Equal(t, 1, cursor[Char])
	assert.Equal(t, 1, sections[1].chars)
	assert.Equal(t, []int{0}, sections[1].cword)
	assert.Equal(t, []int{0}, sections[1].csent)
	assert.Equal(t, []int{0}, sections[1].cpara)
	assert.Equal(t, 2, len(sections))

	cursor = counts{Sectn: 1}
	AppendRunes([]rune{'Z'})
	drawWindow()
	assert.Equal(t, []byte("•üÅ\nX\fYZ"), document)
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, 2, sections[1].chars)
}

func TestAppendRuneCluster(t *testing.T) {
	setupTest()
	ResizeScreen(margin+3, 3)

	AppendRunes([]rune("🇦"))
	AppendRunes([]rune("🇺"))
	assert.Equal(t, []byte("🇦🇺"), document)
	assert.Equal(t, Char, scope)
	assert.Equal(t, 1, cursor[Char])

	AppendRunes([]rune(" "))
	assert.Equal(t, []byte("🇦🇺 "), document)
	assert.Equal(t, 2, cursor[Char])
}

func TestDecScope(t *testing.T) {
	setupTest()
	ResizeScreen(margin+1, 3)

	initialCap = false
	DecScope()
	assert.Equal(t, Sectn, scope)

	initialCap = true
	DecScope()
	assert.Equal(t, Para, scope)
	assert.True(t, initialCap)

	DecScope()
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	DecScope()
	assert.Equal(t, Word, scope)
	assert.False(t, initialCap)
}

func TestIncScope(t *testing.T) {
	setupTest()
	ResizeScreen(margin+1, 3)

	initialCap = false
	scope = Sectn
	IncScope()
	assert.Equal(t, Char, scope)

	IncScope()
	assert.Equal(t, Word, scope)
}

func TestSpace(t *testing.T) {
	setupTest()
	ResizeScreen(margin+6, 4)
	assert.NotPanics(t, func() { Space() })
	assert.Nil(t, document)

	document = []byte("Test")
	Space()
	assert.Equal(t, "Test ", string(document))
	assert.Equal(t, scope, Word)

	Space()
	assert.Equal(t, "Test. ", string(document))
	assert.Equal(t, scope, Sent)

	Space()
	assert.Equal(t, "Test.\n", string(document))
	assert.Equal(t, scope, Para)

	Space()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	Space()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	cursor = counts{5, 1, 1, 1, 1}
	document = []byte("Test,")
	resetIndex()
	scope = Word
	Space()
	assert.Equal(t, "Test, ", string(document))
	assert.Equal(t, scope, Sent)

	document = []byte("Test.")
	scope = Char
	Space()
	assert.Equal(t, "Test. ", string(document))
	assert.Equal(t, scope, Sent)

	Space()
	assert.Equal(t, "Test.\n", string(document))
	assert.Equal(t, scope, Para)

	cursor = counts{5, 1, 1, 1, 1}
	document = []byte("Test.")
	resetIndex()
	newBuffer()
	Space()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	cursor = counts{5, 1, 1, 1, 1}
	document = []byte("Test\n")
	resetIndex()
	newBuffer()
	sections[0].bpara = []int{0, 5}
	sections[0].cpara = []int{0, 5}
	sections[0].csent = []int{0, 5}
	scope = Para
	Space()
	assert.Equal(t, "Test\f", string(document))
	assert.Equal(t, scope, Sectn)
}

func TestSpaceAfterSpace(t *testing.T) {
	setupTest()
	ResizeScreen(margin+1, 3)

	document = []byte("Test, ")
	scope = Char
	Space()
	assert.Equal(t, "Test, ", string(document))
	assert.Equal(t, scope, Word)

	Space()
	assert.Equal(t, "Test, ", string(document))
	assert.Equal(t, scope, Sent)

	document = []byte("Test\n")
	scope = Char
	Space()
	assert.Equal(t, "Test\n", string(document))
	assert.Equal(t, scope, Word)

	Space()
	assert.Equal(t, "Test\n", string(document))
	assert.Equal(t, scope, Sent)

	document = []byte("Test\f")
	scope = Char
	Space()
	assert.Equal(t, "Test\f", string(document))
	assert.Equal(t, scope, Word)

	Space()
	assert.Equal(t, "Test\f", string(document))
	assert.Equal(t, scope, Sent)
}

func TestEnter(t *testing.T) {
	setupTest()
	ResizeScreen(margin+6, 4)

	assert.NotPanics(t, func() { Enter() })
	assert.Nil(t, document)

	document = []byte("Test.")
	cursor[Char] = 5
	Enter()
	assert.Equal(t, "Test.\n", string(document))
	assert.Equal(t, scope, Para)

	Enter()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	Enter()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	cursor = counts{5, 1, 1, 1, 1}
	document = []byte("Test ")
	resetIndex()
	scope = Word
	newBuffer()
	Enter()
	assert.Equal(t, "Test\n", string(document))
	assert.Equal(t, scope, Para)

	Enter()
	assert.Equal(t, "Test\f", string(document))
	assert.Equal(t, scope, Sectn)

	cursor = counts{Sectn: 1}
	document = []byte("Test.")
	scope = Para
	Enter()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)
}
