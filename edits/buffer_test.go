package edits

import (
	"slices"
	"testing"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/stretchr/testify/assert"
)

func setupTest() {
	ID = "J"
	cursor = counts{Para: 1}
	first_para, first_line = 0, 0
	initialCap = false
	scope = Char
	doc.SetText(1, "")
	resetIndex()
}

func TestCursorPos(t *testing.T) {
	resetIndex()
	cursor = counts{Para: 1}
	assert.Equal(t, counts{0, 0, 0, 1}, cursorPos())

	indexPara()
	indexSent(1, 0)
	indexWord(1, 0)
	cursor[Para] = 2
	assert.Equal(t, counts{0, 1, 1, 2}, cursorPos())
}

func TestCursorString(t *testing.T) {
	initialCap = false
	scope = Char
	assert.Equal(t, string(cursorChar[Char]), cursorString())

	initialCap = true
	assert.Equal(t, string(cursorCharCap), cursorString())
}

func TestStatusLine(t *testing.T) {
	setupTest()
	resetIndex()
	ID = "Jotty v0"
	initialCap = false
	ResizeScreen(6, 3)
	assert.Equal(t, "@0/0", statusLine())

	ResizeScreen(20, 3)
	assert.Equal(t, "¶1/1 $0/0 #0/0 @0/0 ", statusLine())

	ResizeScreen(30, 3)
	assert.Equal(t, "Jotty v0  ¶1/1 $0/0 #0/0 @0/0 ", statusLine())
}

func TestNextSegWidth(t *testing.T) {
	assert.Equal(t, 0, nextSegWidth([]byte("")))
	assert.Equal(t, 3, nextSegWidth([]byte("One two")))
	assert.Equal(t, 4, nextSegWidth([]byte("One-two")))
}

func TestDrawLine(t *testing.T) {
	setupTest()
	ResizeScreen(margin+3, 2)

	c := 0
	source := []byte{}
	state := -1
	assert.Equal(t, "_", drawLine(1, &c, &source, &state))
	assert.Equal(t, "", drawLine(2, &c, &source, &state))

	source = []byte("1\xff2")
	state = -1
	assert.Equal(t, "_12", drawLine(1, &c, &source, &state))
	assert.Equal(t, 2, c)

	source = []byte("Test")
	state = -1
	assert.Equal(t, "Tes    -", drawLine(1, &c, &source, &state))

	c = 0
	cursor[Char] = 3
	source = []byte("12 3")
	state = -1
	assert.Equal(t, "12 ", drawLine(1, &c, &source, &state))
	assert.Equal(t, "_3", drawLine(1, &c, &source, &state))

	c = 0
	source = []byte(". Test")
	state = -1
	assert.Equal(t, ". ", drawLine(1, &c, &source, &state))
	assert.Equal(t, 1, lastSentence(1))

	c = 0
	source = []byte("1\n2")
	state = -1
	assert.Equal(t, "1", drawLine(1, &c, &source, &state))

	c = 0
	cursor[Char] = 1
	source = []byte("1\u200b2") // Zero-width space
	state = -1
	assert.Equal(t, "1_2", drawLine(1, &c, &source, &state))

	source = []byte("12  ")
	state = -1
	assert.Equal(t, "12 ", drawLine(1, &c, &source, &state))
}

func TestDrawLineCursor(t *testing.T) {
	setupTest()
	ResizeScreen(margin+1, 3)

	for scope = Char; scope < MaxScope; scope++ {
		c := 0
		source := []byte{}
		state := -1
		assert.Equal(t, string(cursorChar[scope]), drawLine(1, &c, &source, &state))
	}
}

func TestDrawPara(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 2)
	assert.Equal(t, []string{"_", ""}, drawPara(1))

	doc.SetText(1, "Test")
	assert.Equal(t, []string{"_Test", ""}, drawPara(1))
	assert.Equal(t, 0, lastSentence(1))
	assert.Equal(t, 0, lastWord(1))
	assert.Equal(t, counts{4, 1, 1, 1}, total)

	cursor[Char] = 4
	assert.Equal(t, []string{"Test_", ""}, drawPara(1))
	assert.Equal(t, 0, curs_line)
	assert.Equal(t, 0, lastSentence(1))
	assert.Equal(t, 0, lastWord(1))
	assert.Equal(t, counts{4, 1, 1, 1}, total)

	doc.SetText(1, "One two")
	assert.Equal(t, []string{"One ", "_two", ""}, drawPara(1))
	assert.Equal(t, 1, curs_line)
	assert.Equal(t, 0, lastSentence(1))
	assert.Equal(t, 3, lastWord(1))
	assert.Equal(t, counts{7, 2, 1, 1}, total)

	cursor[Char] = 0
	assert.Equal(t, []string{"_One ", "two", ""}, drawPara(1))
	assert.Equal(t, 0, lastSentence(1))
	assert.Equal(t, 3, lastWord(1))
	assert.Equal(t, counts{7, 2, 1, 1}, total)

	doc.CreateParagraph(2)
	indexPara()
	assert.Equal(t, []string{"", ""}, drawPara(2))

	doc.DeleteParagraph(2)
}

func TestDrawWindow(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)
	doc.SetText(1, "")
	drawWindow()
	assert.Equal(t, []para{{[]string{"_", ""}}}, cache)
	assert.Equal(t, 1, curs_para)

	doc.CreateParagraph(2)
	defer doc.DeleteParagraph(2)
	indexPara()
	doc.SetText(2, "Test")
	cursor[Para] = 2
	drawWindow()
	expect := []para{{[]string{"", ""}}, {[]string{"_Test", ""}}}
	assert.Equal(t, expect, cache)

	appendParaBreak()
	defer doc.DeleteParagraph(3)
	initialCap = false
	scope = Char
	drawWindow()
	expect[1].text[0] = "Test"
	expect = append(expect, para{[]string{"_", ""}})
	assert.Equal(t, expect, cache)

	drawWindow()
	assert.Equal(t, expect, cache)

	cache = nil
	cursor[Para] = 2
	drawWindow()
	assert.Equal(t, []para{{}, {[]string{"_Test", ""}}}, cache)

	cursor[Para] = 1
	drawWindow()
	expect = []para{{[]string{"_", ""}}, {[]string{"Test", ""}}}
	assert.Equal(t, expect, cache)

	drawWindow()
	assert.Equal(t, 1, curs_para)

	first_line = 1
	drawWindow()
	assert.Equal(t, 0, first_line)

	ResizeScreen(margin+4, 12)
	drawWindow()
	expect = append(expect, para{[]string{"", ""}})
	assert.Equal(t, expect, cache)

	cache = slices.Delete(cache, 2, 3)
	doc.CreateParagraph(4)
	defer doc.DeleteParagraph(4)
	indexPara()
	cursor[Para] = 4
	drawWindow()
	expect = []para{{}, {}, {}, {[]string{"_", ""}}}
	assert.Equal(t, expect, cache)

	// TODO Test that inserting a new paragraph still draws the text below
}

func TestScreen(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 5)
	doc.SetText(1, "")
	assert.Equal(t, "_\n\n\n\n@0/0", Screen())

	doc.CreateParagraph(2)
	defer doc.DeleteParagraph(2)
	indexPara()
	cursor[Para] = 2
	assert.Equal(t, "\n\n_\n\n@0/0", Screen())

	cursor[Para] = 1
	assert.Equal(t, "_\n\n\n\n@0/0", Screen())

	doc.CreateParagraph(3)
	defer doc.DeleteParagraph(3)
	indexPara()
	cursor[Para] = 3
	assert.Equal(t, "\n\n\n_\n@0/0", Screen())

	cursor[Para] = 2
	assert.Equal(t, "_\n\n\n\n@0/0", Screen())

	doc.SetText(1, "A B C D")
	doc.SetText(2, "1 2 3 4")
	cursor[Para] = 1
	assert.Equal(t, "_A B \nC D\n\n1 2 \n@0/14", Screen())

	cursor[Para] = 3
	assert.Equal(t, "1 2 \n3 4\n\n_\n@14/14", Screen())
}
