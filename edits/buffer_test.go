package edits

import (
	"testing"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/stretchr/testify/assert"
)

func setupTest() {
	ID = "J"
	cursor = counts{Para: 1, Sectn: 1}
	first_buff, first_line = 0, 0
	initialCap = false
	scope = Char
	doc.SetText(1, 1, "")
	resetIndex()
}

func TestCursorPos(t *testing.T) {
	resetIndex()
	cursor = counts{Para: 1, Sectn: 1}
	assert.Equal(t, counts{0, 0, 0, 1, 1}, cursorPos())

	indexPara(1)
	indexSent(1, 1, 0)
	indexWord(1, 1, 0)
	cursor[Para] = 2
	assert.Equal(t, counts{0, 1, 1, 2, 1}, cursorPos())
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
	ID = "Jotty v0"
	initialCap = false
	ResizeScreen(6, 3)
	assert.Equal(t, "@0/0", statusLine())

	ResizeScreen(26, 3)
	assert.Equal(t, "§1/1: ¶1/1 $0/0 #0/0 @0/0 ", statusLine())

	ResizeScreen(36, 3)
	assert.Equal(t, "Jotty v0  §1/1: ¶1/1 $0/0 #0/0 @0/0 ", statusLine())
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
	assert.Equal(t, "_", drawLine(1, 1, &c, &source, &state))
	assert.Equal(t, "", drawLine(1, 2, &c, &source, &state))
	assert.Equal(t, "", drawLine(2, 1, &c, &source, &state))

	source = []byte("1\xff2")
	state = -1
	assert.Equal(t, "_12", drawLine(1, 1, &c, &source, &state))
	assert.Equal(t, 2, c)

	source = []byte("Test")
	state = -1
	assert.Equal(t, "Tes    -", drawLine(1, 1, &c, &source, &state))

	c = 0
	cursor[Char] = 3
	source = []byte("12 3")
	state = -1
	assert.Equal(t, "12 ", drawLine(1, 1, &c, &source, &state))
	assert.Equal(t, "_3", drawLine(1, 1, &c, &source, &state))

	c = 0
	source = []byte(". Test")
	state = -1
	assert.Equal(t, ". ", drawLine(1, 1, &c, &source, &state))
	assert.Equal(t, 1, lastSentence(1, 1))

	c = 0
	source = []byte("1\n2")
	state = -1
	assert.Equal(t, "1", drawLine(1, 1, &c, &source, &state))

	c = 0
	cursor[Char] = 1
	source = []byte("1\u200b2") // Zero-width space
	state = -1
	assert.Equal(t, "1_2", drawLine(1, 1, &c, &source, &state))

	source = []byte("12  ")
	state = -1
	assert.Equal(t, "12 ", drawLine(1, 1, &c, &source, &state))
}

func TestDrawLineCursor(t *testing.T) {
	setupTest()
	ResizeScreen(margin+1, 3)

	for scope = Char; scope <= Sectn; scope++ {
		c := 0
		source := []byte{}
		state := -1
		assert.Equal(t, string(cursorChar[scope]), drawLine(1, 1, &c, &source, &state))
	}
}

func TestDrawPara(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 2)
	assert.Equal(t, []string{"_", ""}, drawPara(1, 1))

	doc.SetText(1, 1, "Test")
	assert.Equal(t, []string{"_Test", ""}, drawPara(1, 1))
	assert.Equal(t, 0, lastSentence(1, 1))
	assert.Equal(t, 0, lastWord(1, 1))
	assert.Equal(t, 1, sections[0].sents)
	assert.Equal(t, 1, sections[0].words)
	assert.Equal(t, 4, sectionChars(1))

	cursor[Char] = 4
	assert.Equal(t, []string{"Test_", ""}, drawPara(1, 1))
	assert.Equal(t, 0, curs_line)
	assert.Equal(t, 0, lastSentence(1, 1))
	assert.Equal(t, 0, lastWord(1, 1))
	assert.Equal(t, 1, sections[0].sents)
	assert.Equal(t, 1, sections[0].words)
	assert.Equal(t, 4, sectionChars(1))

	doc.SetText(1, 1, "One two")
	assert.Equal(t, []string{"One ", "_two", ""}, drawPara(1, 1))
	assert.Equal(t, 1, curs_line)
	assert.Equal(t, 0, lastSentence(1, 1))
	assert.Equal(t, 3, lastWord(1, 1))
	assert.Equal(t, 1, sections[0].sents)
	assert.Equal(t, 2, sections[0].words)
	assert.Equal(t, 7, sectionChars(1))

	cursor[Char] = 0
	assert.Equal(t, []string{"_One ", "two", ""}, drawPara(1, 1))
	assert.Equal(t, 0, lastSentence(1, 1))
	assert.Equal(t, 3, lastWord(1, 1))
	assert.Equal(t, 1, sections[0].sents)
	assert.Equal(t, 2, sections[0].words)
	assert.Equal(t, 7, sectionChars(1))

	doc.CreateSection(2)
	indexSectn()
	assert.Equal(t, []string{"_One ", "two", "─────────"}, drawPara(1, 1))
	assert.Equal(t, []string{"", ""}, drawPara(2, 1))
	assert.Equal(t, 0, lastSentence(2, 1))
	assert.Equal(t, 0, lastWord(2, 1))
	assert.Equal(t, 0, sections[1].sents)
	assert.Equal(t, 0, sections[1].words)
	assert.Equal(t, 0, sectionChars(2))

	doc.CreateParagraph(2, 2)
	indexPara(2)
	cursor = counts{0, 0, 0, 1, 2}
	assert.Equal(t, []string{"", ""}, drawPara(2, 2))

	doc.CreateSection(3)
	indexSectn()
	assert.Equal(t, []string{"_", ""}, drawPara(2, 1))

	doc.DeleteSection(3)
	doc.DeleteSection(2)
}

func TestDrawWindow(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)
	doc.SetText(1, 1, "")
	drawWindow()
	assert.Equal(t, []para{{1, 1, []string{"_", ""}}}, buffer)

	doc.CreateSection(2)
	indexSectn()
	doc.SetText(2, 1, "Test")
	cursor[Sectn] = 2
	drawWindow()
	expect := []para{{1, 1, []string{"", "─────────"}}, {2, 1, []string{"_Test", ""}}}
	assert.Equal(t, expect, buffer)

	doc.CreateParagraph(2, 2)
	indexPara(2)
	cursor[Para] = 2
	drawWindow()
	expect[1].text[0] = "Test"
	expect = append(expect, para{sn: 2, pn: 2, text: []string{"_", ""}})
	assert.Equal(t, expect, buffer)

	drawWindow()
	assert.Equal(t, expect, buffer)

	doc.CreateParagraph(2, 3)
	indexPara(2)
	doc.CreateParagraph(2, 4)
	indexPara(2)
	cursor[Para] = 4
	drawWindow()
	assert.Equal(t, []para{{2, 4, []string{"_", ""}}}, buffer)

	cursor[Para] = 3
	drawWindow()
	assert.Equal(t, []para{{2, 3, []string{"_", ""}}, {2, 4, []string{"", ""}}}, buffer)

	cursor[Para] = 1
	drawWindow()
	cursor = counts{0, 0, 0, 1, 1}
	drawWindow()
	expect = []para{{1, 1, []string{"_", "─────────"}}, {2, 1, []string{"Test", ""}}}
	assert.Equal(t, expect, buffer)

	drawWindow()
	assert.Equal(t, 0, curs_buff)

	first_line = 1
	drawWindow()
	assert.Equal(t, 0, first_line)

	ResizeScreen(margin+4, 12)
	drawWindow()
	for pn := 2; pn <= 4; pn++ {
		expect = append(expect, para{2, pn, []string{"", ""}})
	}
	assert.Equal(t, expect, buffer)

	doc.DeleteSection(2)
}

func TestScreen(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 5)
	doc.SetText(1, 1, "")
	assert.Equal(t, "_\n\n\n\n@0/0", Screen())

	doc.CreateSection(2)
	indexSectn()
	cursor[Sectn] = 2
	assert.Equal(t, "\n─────────\n_\n\n@0/0", Screen())

	doc.CreateParagraph(2, 2)
	indexPara(2)
	cursor[Para] = 2
	assert.Equal(t, "─────────\n\n\n_\n@0/0", Screen())

	cursor = counts{0, 0, 0, 1, 1}
	assert.Equal(t, "_\n─────────\n\n\n@0/0", Screen())

	cursor[Sectn] = 2
	assert.Equal(t, "\n─────────\n_\n\n@0/0", Screen())

	doc.CreateSection(3)
	indexSectn()
	cursor[Sectn] = 3
	assert.Equal(t, "\n\n\n_\n@0/0", Screen())

	cursor[Sectn] = 2
	assert.Equal(t, "_\n\n\n\n@0/0", Screen())

	doc.SetText(2, 1, "One two 3 4 5 6")
	cursor[Para] = 1
	drawWindow()

	cursor[Para] = 2
	assert.Equal(t, "3 4 \n5 6\n\n_\n@15/15", Screen())

	doc.DeleteSection(3)
	doc.DeleteSection(2)
}
