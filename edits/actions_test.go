package edits

import (
	"slices"
	"testing"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/stretchr/testify/assert"
)

func TestInsertParaBreak(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 8)

	doc.SetText(1, "Test")
	cursor = counts{4, 0, 0, 1}
	drawWindow()
	insertParaBreak()
	drawWindow()
	assert.Equal(t, 2, cursor[Para])
	assert.Equal(t, 2, doc.Paragraphs())
	assert.Equal(t, "Test", doc.GetText(1))

	doc.DeleteParagraph(2)
	cache = slices.Delete(cache, 1, 2)
	curs_para = 1
	doc.SetText(1, "Test ")
	cursor = counts{5, 0, 0, 1}
	drawWindow()
	insertParaBreak()
	defer doc.DeleteParagraph(2)
	drawWindow()
	assert.Equal(t, 2, cursor[Para])
	assert.Equal(t, 2, doc.Paragraphs())
	assert.Equal(t, "Test", doc.GetText(1))
	assert.Equal(t, []para{{4, []int{0}, []int{0}, []string{"Test"}}, {text: []string{string(cursorCharCap)}}}, cache)

	insertParaBreak()
	defer doc.DeleteParagraph(3)
	drawWindow()
	assert.Equal(t, 3, cursor[Para])
	assert.Equal(t, 3, doc.Paragraphs())
	expect := []para{{4, []int{0}, []int{0}, []string{"Test"}}, {text: []string{""}}, {text: []string{string(cursorCharCap)}}}
	assert.Equal(t, expect, cache)

	cursor = counts{5, 0, 0, 1}
	drawWindow()
	insertParaBreak()
	defer doc.DeleteParagraph(4)
	assert.Equal(t, 2, cursor[Para])
	assert.Equal(t, 4, doc.Paragraphs())
}

func TestInsertRunes(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)
	drawWindow()
	initialCap = true
	InsertRunes([]rune("A"))
	assert.Equal(t, "A", doc.GetText(1))
	assert.Equal(t, 1, cursor[Char])

	initialCap = true
	InsertRunes([]rune("u"))
	assert.Equal(t, "AU", doc.GetText(1))
	assert.Equal(t, 2, cursor[Char])

	InsertRunes([]rune("🇦🇺"))
	assert.Equal(t, "AU🇦🇺", doc.GetText(1))
	assert.Equal(t, 3, cursor[Char])

	cursor[Char] = 2
	drawWindow()
	InsertRunes([]rune("="))
	assert.Equal(t, "AU=🇦🇺", doc.GetText(1))
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
	ResizeScreen(margin+4, 3)
	drawWindow()

	scope = Char
	Space()
	assert.Equal(t, "", doc.GetText(1))
	assert.Equal(t, Char, scope)

	doc.SetText(1, "Test")
	cursor[Char] = 4
	drawWindow()
	Space()
	assert.Equal(t, "Test ", doc.GetText(1))
	assert.Equal(t, Word, scope)

	drawWindow()
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
	cursor[Char] = 5
	drawWindow()
	scope = Char
	Space()
	assert.Equal(t, "Test. ", doc.GetText(1))
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	drawWindow()
	scope = Word
	Space()
	assert.Equal(t, "Test. ", doc.GetText(1))
	assert.Equal(t, Sent, scope)
	assert.True(t, initialCap)

	doc.SetText(1, "Test")
	cursor[Char] = 4
	drawWindow()
	scope = Word
	Space()
	assert.Equal(t, "Test ", doc.GetText(1))
	assert.Equal(t, Sent, scope)
	assert.False(t, initialCap)

	ResizeScreen(margin+4, 3)
	drawWindow()
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

func TestBackspace(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)

	Backspace()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)

	doc.CreateParagraph(2)
	doc.SetText(2, "A")
	cursor[Para] = 2
	drawWindow()
	Backspace()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)
	assert.Equal(t, "A", doc.GetText(1))

	doc.CreateParagraph(2)
	doc.SetText(2, "B")
	cursor[Para] = 2
	drawWindow()
	Backspace()
	assert.Equal(t, counts{1, 0, 0, 1}, cursor)
	assert.Equal(t, "A B", doc.GetText(1))

	doc.CreateParagraph(2)
	cursor = counts{0, 0, 0, 2}
	drawWindow()
	Backspace()
	assert.Equal(t, counts{3, 0, 0, 1}, cursor)
	assert.Equal(t, "A B", doc.GetText(1))

	doc.CreateParagraph(2)
	defer doc.DeleteParagraph(2)
	doc.SetText(2, "C D")
	cursor = counts{1, 0, 0, 2}
	scope = Para
	drawWindow()
	Backspace()
	assert.Equal(t, counts{0, 1, 1, 2}, cursor)
	assert.Equal(t, " D", doc.GetText(2))

	cursor = counts{2, 0, 0, 1}
	scope = Char
	drawWindow()
	Backspace()
	assert.Equal(t, 1, cursor[Char])
	assert.Equal(t, "AB", doc.GetText(1))

	doc.SetText(1, "A B C D")
	cursor[Char] = 5
	scope = Word
	drawWindow()
	Backspace()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, "A B  D", doc.GetText(1))

	doc.SetText(1, "A B C D")
	cursor[Char] = 6
	scope = Word
	drawWindow()
	Backspace()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, "A B D", doc.GetText(1))

	doc.SetText(1, "A B C ")
	cursor[Char] = 6
	scope = Word
	drawWindow()
	Backspace()
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, "A B ", doc.GetText(1))

	doc.SetText(1, "A. B? C! D")
	cursor[Char] = 8
	scope = Sent
	drawWindow()
	Backspace()
	assert.Equal(t, 6, cursor[Char])
	assert.Equal(t, "A. B?  D", doc.GetText(1))

	doc.SetText(1, "A. B? C! D")
	cursor[Char] = 9
	scope = Sent
	drawWindow()
	Backspace()
	assert.Equal(t, 6, cursor[Char])
	assert.Equal(t, "A. B? D", doc.GetText(1))

	doc.SetText(1, "A. B? C! ")
	cursor[Char] = 9
	scope = Sent
	drawWindow()
	Backspace()
	assert.Equal(t, 6, cursor[Char])
	assert.Equal(t, "A. B? ", doc.GetText(1))
}
