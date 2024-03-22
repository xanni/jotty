package edits

import (
	"testing"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/stretchr/testify/assert"
)

func TestAppendParaBreak(t *testing.T) {
	setupTest()
	appendParaBreak()
	assert.Equal(t, 1, doc.Paragraphs(1))

	doc.SetText(1, 1, "Test")
	sections[0].chars = 4
	cursor = counts{4, 0, 0, 1, 1}
	appendParaBreak()
	assert.Equal(t, 2, cursor[Para])
	assert.Equal(t, 2, doc.Paragraphs(1))
	assert.Equal(t, "Test", doc.GetText(1, 1))

	doc.DeleteParagraph(1, 2)
	doc.SetText(1, 1, "Test ")
	sections[0].chars = 5
	cursor = counts{5, 0, 0, 1, 1}
	appendParaBreak()
	assert.Equal(t, 2, cursor[Para])
	assert.Equal(t, 2, doc.Paragraphs(1))
	assert.Equal(t, "Test", doc.GetText(1, 1))

	doc.DeleteParagraph(1, 2)
}

func TestAppendSectnBreak(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)
	appendSectnBreak()
	assert.Equal(t, 1, doc.Sections())

	doc.SetText(1, 1, "Test")
	sections[0].chars = 4
	sections[0].p[0].chars = 4
	cursor = counts{4, 0, 0, 1, 1}
	appendSectnBreak()
	assert.Equal(t, 2, cursor[Sectn])
	assert.Equal(t, 2, doc.Sections())

	doc.DeleteSection(2)
	doc.CreateParagraph(1, 2)
	indexPara(1)
	cursor = counts{4, 0, 0, 2, 1}
	drawWindow()
	appendSectnBreak()
	assert.Equal(t, counts{0, 0, 0, 1, 2}, cursor)
	assert.Equal(t, 2, doc.Sections())
	assert.Equal(t, 1, doc.Paragraphs(1))

	doc.DeleteSection(2)
}

func TestAppendRunes(t *testing.T) {
	setupTest()
	initialCap = true
	AppendRunes([]rune("A"))
	assert.Equal(t, "A", doc.GetText(1, 1))
	assert.Equal(t, 1, cursor[Char])

	initialCap = true
	AppendRunes([]rune("u"))
	assert.Equal(t, "AU", doc.GetText(1, 1))
	assert.Equal(t, 2, cursor[Char])

	AppendRunes([]rune("🇦🇺"))
	assert.Equal(t, "AU🇦🇺", doc.GetText(1, 1))
	assert.Equal(t, 3, cursor[Char])
}

func TestDecScope(t *testing.T) {
	setupTest()

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

	scope = Sectn
	IncScope()
	assert.Equal(t, Char, scope)

	IncScope()
	assert.Equal(t, Word, scope)
}

func TestSpace(t *testing.T) {
	setupTest()

	for scope = Char; scope < MaxScope; scope++ {
		Space()
		assert.Equal(t, counts{0, 0, 0, 1, 1}, cursor)
	}

	scope = Char
	doc.SetText(1, 1, "Test")
	Space()
	assert.Equal(t, "Test ", doc.GetText(1, 1))
	assert.Equal(t, Word, scope)

	Space()
	assert.Equal(t, "Test. ", doc.GetText(1, 1))
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	scope = Char
	Space()
	assert.Equal(t, "Test. ", doc.GetText(1, 1))
	assert.Equal(t, Word, scope)
	assert.False(t, initialCap)

	doc.SetText(1, 1, "Test.")
	scope = Char
	Space()
	assert.Equal(t, "Test. ", doc.GetText(1, 1))
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	scope = Word
	Space()
	assert.Equal(t, "Test. ", doc.GetText(1, 1))
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	doc.SetText(1, 1, "Test")
	scope = Word
	Space()
	assert.Equal(t, "Test ", doc.GetText(1, 1))
	assert.Equal(t, Sent, scope)
	assert.False(t, initialCap)
}

func TestEnter(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)

	doc.SetText(1, 1, "Test")
	drawWindow()
	scope = Sent
	Enter()
	assert.Equal(t, 2, doc.Paragraphs(1))
	assert.Equal(t, Para, scope)

	Enter()
	assert.Equal(t, 2, doc.Sections())
	assert.Equal(t, Sectn, scope)

	Enter()
	assert.Equal(t, 2, doc.Sections())
	assert.Equal(t, Sectn, scope)

	doc.DeleteSection(2)
}
