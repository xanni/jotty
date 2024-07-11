package edits

import (
	"slices"
	"testing"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/stretchr/testify/assert"
)

func TestAppendParaBreak(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 8)

	doc.SetText(1, "Test")
	cursor = counts{4, 0, 0, 1}
	drawWindow()
	appendParaBreak()
	drawWindow()
	assert.Equal(t, 2, cursor[Para])
	assert.Equal(t, 2, doc.Paragraphs())
	assert.Equal(t, "Test", doc.GetText(1))

	doc.DeleteParagraph(2)
	cache = slices.Delete(cache, 1, 2)
	paras = slices.Delete(paras, 1, 2)
	curs_para = 0
	doc.SetText(1, "Test ")
	cursor = counts{5, 0, 0, 1}
	appendParaBreak()
	defer doc.DeleteParagraph(2)
	drawWindow()
	assert.Equal(t, 2, cursor[Para])
	assert.Equal(t, 2, doc.Paragraphs())
	assert.Equal(t, "Test", doc.GetText(1))
	assert.Equal(t, []para{{1, []string{"Test", ""}}, {2, []string{string(cursorCharCap), ""}}}, cache)

	appendParaBreak()
	defer doc.DeleteParagraph(3)
	drawWindow()
	assert.Equal(t, 3, cursor[Para])
	assert.Equal(t, 3, doc.Paragraphs())
	expect := []para{{1, []string{"Test", ""}}, {2, []string{"", ""}}, {3, []string{string(cursorCharCap), ""}}}
	assert.Equal(t, expect, cache)

	cursor = counts{5, 0, 0, 1}
	drawWindow()
	appendParaBreak()
	defer doc.DeleteParagraph(4)
	assert.Equal(t, 2, cursor[Para])
	assert.Equal(t, 4, doc.Paragraphs())
}

func TestAppendRunes(t *testing.T) {
	setupTest()
	initialCap = true
	AppendRunes([]rune("A"))
	assert.Equal(t, "A", doc.GetText(1))
	assert.Equal(t, 1, cursor[Char])

	initialCap = true
	AppendRunes([]rune("u"))
	assert.Equal(t, "AU", doc.GetText(1))
	assert.Equal(t, 2, cursor[Char])

	AppendRunes([]rune("🇦🇺"))
	assert.Equal(t, "AU🇦🇺", doc.GetText(1))
	assert.Equal(t, 3, cursor[Char])
}

func TestDecScope(t *testing.T) {
	setupTest()

	DecScope()
	assert.Equal(t, Para, scope)

	initialCap = true
	DecScope()
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	DecScope()
	assert.Equal(t, Word, scope)
	assert.False(t, initialCap)
}

func TestIncScope(t *testing.T) {
	setupTest()

	scope = Para
	IncScope()
	assert.Equal(t, Char, scope)

	IncScope()
	assert.Equal(t, Word, scope)
}

func TestSpace(t *testing.T) {
	setupTest()

	scope = Char
	Space()
	assert.Equal(t, "", doc.GetText(1))
	assert.Equal(t, Char, scope)

	doc.SetText(1, "Test")
	Space()
	assert.Equal(t, "Test ", doc.GetText(1))
	assert.Equal(t, Word, scope)

	Space()
	assert.Equal(t, "Test. ", doc.GetText(1))
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	scope = Char
	Space()
	assert.Equal(t, "Test. ", doc.GetText(1))
	assert.Equal(t, Word, scope)
	assert.False(t, initialCap)

	doc.SetText(1, "Test.")
	scope = Char
	Space()
	assert.Equal(t, "Test. ", doc.GetText(1))
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	scope = Word
	Space()
	assert.Equal(t, "Test. ", doc.GetText(1))
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	doc.SetText(1, "Test")
	scope = Word
	Space()
	assert.Equal(t, "Test ", doc.GetText(1))
	assert.Equal(t, Sent, scope)
	assert.False(t, initialCap)

	Space()
	assert.Equal(t, "Test", doc.GetText(1))
	assert.Equal(t, 2, doc.Paragraphs())
	assert.Equal(t, Para, scope)
	assert.True(t, initialCap)

	doc.DeleteParagraph(2)
}

func TestEnter(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)

	doc.SetText(1, "Test")
	drawWindow()
	scope = Sent
	Enter()
	assert.Equal(t, 2, doc.Paragraphs())
	assert.Equal(t, Para, scope)

	doc.DeleteParagraph(2)
}
