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

	tests := map[string]struct {
		text       string
		cursor     int
		scope      Scope
		expect     string
		newScope   Scope
		initialCap bool
	}{
		"Char new paragraph":  {"", 0, Char, "", Char, false},
		"Char end of word":    {"Test", 4, Char, "Test ", Word, false},
		"Char after period":   {"Test.", 5, Char, "Test. ", Sent, true},
		"Char after sentence": {"Test. ", 6, Char, "Test. ", Word, false},
		"Word end of word":    {"Test", 4, Word, "Test ", Sent, false},
		"Word after space":    {"Test ", 5, Word, "Test. ", Sent, true},
		"Word after sentence": {"Test. ", 6, Word, "Test. ", Sent, true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			doc.SetText(1, test.text)
			cursor[Char] = test.cursor
			initialCap = false
			scope = test.scope

			drawWindow()
			Space()

			assert.Equal(t, test.expect, doc.GetText(1), "text")
			assert.Equal(t, test.newScope, scope, "scope")
			assert.Equal(t, test.initialCap, initialCap, "initialCap")
		})
	}

	doc.SetText(1, "Test")
	cursor[Char] = 4
	scope = Sent
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

func TestBackspaceMerge(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)

	Backspace()
	assert.Equal(t, counts{0, 0, 0, 1}, cursor)

	tests := map[string]struct {
		p1, p2    string
		expect    string
		newCursor int
	}{
		"Empty previous": {"", "A", "A", 0},
		"Text in both":   {"A", "B", "A B", 1},
		"Empty current":  {"A B", "", "A B", 3},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			doc.CreateParagraph(2)
			doc.SetText(1, test.p1)
			doc.SetText(2, test.p2)
			cursor = counts{0, 0, 0, 2}

			drawWindow()
			Backspace()

			assert.Equal(t, test.expect, doc.GetText(1), "text")
			assert.Equal(t, test.newCursor, cursor[Char], "cursor")
		})
	}
}

func TestBackspace(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)

	doc.CreateParagraph(2)
	defer doc.DeleteParagraph(2)
	doc.SetText(2, "C D")
	cursor = counts{1, 0, 0, 2}
	scope = Para
	drawWindow()
	Backspace()
	assert.Equal(t, counts{0, 1, 1, 2}, cursor)
	assert.Equal(t, " D", doc.GetText(2))

	tests := map[string]struct {
		text      string
		cursor    int
		scope     Scope
		expect    string
		newCursor int
	}{
		"Char":                  {"A B", 2, Char, "AB", 1},
		"Word without space":    {"A B C", 3, Word, "A  C", 2},
		"Word":                  {"A B C", 4, Word, "A C", 2},
		"Word at paragraph end": {"A B C ", 6, Word, "A B ", 4},
		"Sent without space":    {"A. B? C!", 5, Sent, "A.  C!", 3},
		"Sent":                  {"A. B? C!", 6, Sent, "A. C!", 3},
		"Sent at paragraph end": {"A. B? C! ", 9, Sent, "A. B? ", 6},
	}

	cursor[Para] = 1
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			doc.SetText(1, test.text)
			cursor[Char] = test.cursor
			scope = test.scope

			drawWindow()
			Backspace()

			assert.Equal(t, test.expect, doc.GetText(1), "text")
			assert.Equal(t, test.newCursor, cursor[Char], "cursor")
		})
	}
}
