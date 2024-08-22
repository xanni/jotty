package edits

import (
	"os"
	"slices"
	"testing"

	ps "git.sericyb.com.au/jotty/permascroll"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := ps.OpenPermascroll(os.DevNull); err != nil {
		panic(err)
	}
}

func TestInsertParaBreak(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 8)

	ps.Init()
	ps.AppendText(1, "Test")
	cursor = counts{4, 0, 0, 1}
	drawWindow()
	insertParaBreak()
	drawWindow()
	assert.Equal(2, cursor[Para])
	assert.Equal(2, ps.Paragraphs())
	assert.Equal("Test", ps.GetText(1))

	ps.Init()
	ps.AppendText(1, "Test ")
	cache = nil
	cursPara = 1
	cursor = counts{5, 0, 0, 1}
	drawWindow()
	insertParaBreak()
	drawWindow()
	assert.Equal(2, cursor[Para])
	assert.Equal(2, ps.Paragraphs())
	assert.Equal("Test", ps.GetText(1))
	expect := []para{{4, []int{0}, []int{0}, []string{"Test"}}, {text: []string{string(cursorCharCap)}}}
	assert.Equal(expect, cache)

	insertParaBreak()
	drawWindow()
	assert.Equal(3, cursor[Para])
	assert.Equal(3, ps.Paragraphs())
	expect = slices.Insert(expect, 1, para{text: []string{""}})
	assert.Equal(expect, cache)

	cursor = counts{5, 0, 0, 1}
	drawWindow()
	insertParaBreak()
	assert.Equal(2, cursor[Para])
	assert.Equal(4, ps.Paragraphs())
}

func TestInsertRunes(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 3)
	drawWindow()
	initialCap = true
	InsertRunes([]rune("A"))
	assert.Equal("A", ps.GetText(1))
	assert.Equal(1, cursor[Char])

	initialCap = true
	InsertRunes([]rune("u"))
	assert.Equal("AU", ps.GetText(1))
	assert.Equal(2, cursor[Char])

	InsertRunes([]rune("🇦🇺"))
	assert.Equal("AU🇦🇺", ps.GetText(1))
	assert.Equal(3, cursor[Char])

	cursor[Char] = 2
	drawWindow()
	InsertRunes([]rune("="))
	assert.Equal("AU=🇦🇺", ps.GetText(1))
	assert.Equal(3, cursor[Char])
}

func TestDecScope(t *testing.T) {
	assert := assert.New(t)
	setupTest()

	DecScope()
	assert.Equal(Para, scope)

	initialCap = true
	DecScope()
	assert.Equal(Sent, scope)
	assert.True(initialCap)

	DecScope()
	assert.Equal(Word, scope)
	assert.False(initialCap)
}

func TestIncScope(t *testing.T) {
	assert := assert.New(t)
	setupTest()

	scope = Para
	IncScope()
	assert.Equal(Char, scope)

	IncScope()
	assert.Equal(Word, scope)
}

func TestSpace(t *testing.T) {
	assert := assert.New(t)
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
		t.Run(name, func(_ *testing.T) {
			ps.Init()
			ps.AppendText(1, test.text)
			cursor[Char] = test.cursor
			initialCap = false
			scope = test.scope

			drawWindow()
			Space()

			assert.Equal(test.expect, ps.GetText(1), "text")
			assert.Equal(test.newScope, scope, "scope")
			assert.Equal(test.initialCap, initialCap, "initialCap")
		})
	}

	ps.Init()
	ps.AppendText(1, "Test")
	cursor[Char] = 4
	scope = Sent
	ResizeScreen(margin+4, 3)

	drawWindow()
	Space()

	assert.Equal("Test", ps.GetText(1))
	assert.Equal(2, ps.Paragraphs())
	assert.Equal(Para, scope)
	assert.True(initialCap)
}

func TestEnter(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 3)

	ps.AppendText(1, "Test")
	drawWindow()
	scope = Sent
	Enter()
	assert.Equal(2, ps.Paragraphs())
	assert.Equal(Para, scope)
}

func TestBackspaceMerge(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 3)

	Backspace()
	assert.Equal(counts{0, 0, 0, 1}, cursor)

	tests := map[string]struct {
		p1, p2    string
		expect    string
		newCursor int
	}{
		"Empty previous": {"", "A", "A", 0},
		"Text in both":   {"A", "B", "A B", 1},
		"Trailing space": {"A ", "B", "A B", 2},
		"Empty current":  {"A B", "", "A B", 3},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			ps.Init()
			ps.SplitParagraph(1, 0)
			ps.AppendText(1, test.p1)
			ps.AppendText(2, test.p2)
			cursor = counts{0, 0, 0, 2}

			drawWindow()
			Backspace()

			assert.Equal(test.expect, ps.GetText(1), "text")
			assert.Equal(test.newCursor, cursor[Char], "cursor")
		})
	}
}

func TestGetWords(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("Test", getWords(2, "Test"))
}

func TestBackspace(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 3)

	ps.SplitParagraph(1, 0)
	ps.AppendText(2, "C D")
	cursor = counts{1, 0, 0, 2}
	scope = Para
	drawWindow()
	Backspace()
	assert.Equal(counts{0, 1, 1, 2}, cursor)
	assert.Equal(" D", ps.GetText(2))

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
		t.Run(name, func(_ *testing.T) {
			ps.Init()
			ps.AppendText(1, test.text)
			cursor[Char] = test.cursor
			scope = test.scope

			cache = nil
			drawWindow()
			Backspace()

			assert.Equal(test.expect, ps.GetText(1), "text")
			assert.Equal(test.newCursor, cursor[Char], "cursor")
		})
	}
}

func TestDeleteMerge(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 3)
	drawWindow()

	Delete()
	assert.Equal(counts{0, 0, 0, 1}, cursor)

	tests := map[string]struct {
		p1, p2    string
		expect    string
		newCursor int
	}{
		"Empty next":     {"A", "", "A", 1},
		"Text in both":   {"A", "B", "A B", 1},
		"Trailing space": {"A ", "B", "A B", 2},
		"Empty current":  {"", "A B", "A B", 0},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			ps.Init()
			ps.SplitParagraph(1, 0)
			ps.AppendText(1, test.p1)
			ps.AppendText(2, test.p2)
			cursor = counts{0, 0, 0, 2}
			drawWindow()
			cursor = counts{len(test.p1), 0, 0, 1}

			drawWindow()
			Delete()

			assert.Equal(test.expect, ps.GetText(1), "text")
			assert.Equal(test.newCursor, cursor[Char], "cursor")
		})
	}
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 3)

	tests := map[string]struct {
		text   string
		cursor int
		scope  Scope
		expect string
	}{
		"Char": {"A B", 1, Char, "AB"},
		"Word": {"A  B  C", 1, Word, "AC"},
		"Sent": {"A.  B?  C!", 2, Sent, "A.C!"},
		"Para": {"A B", 1, Para, "A"},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			ps.Init()
			ps.AppendText(1, test.text)
			cursor[Char] = test.cursor
			scope = test.scope

			drawWindow()
			Delete()

			assert.Equal(test.expect, ps.GetText(1), "text")
		})
	}
}

func TestUndoRedo(t *testing.T) {
	assert := assert.New(t)
	setupTest()

	Undo()
	assert.Equal(0, cursor[Char])

	Redo()
	assert.Equal(0, cursor[Char])

	ps.InsertText(1, 0, "Test")
	ps.Flush()
	ps.DeleteText(1, 1, 2)
	Undo()
	assert.Equal(2, cursor[Char])

	Redo()
	assert.Equal(1, cursor[Char])
}
