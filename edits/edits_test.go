package edits

import (
	"slices"
	"testing"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/stretchr/testify/assert"
)

func resetCache() {
	cache = []para{{}}
	total = counts{0, 0, 0, 1}
}

func setupTest() {
	ID = "J"
	cursor = counts{Para: 1}
	firstPara, firstLine = 0, 0
	initialCap = false
	scope = Char
	doc.SetText(1, "")
	resetCache()
}

func TestIndexWord(t *testing.T) {
	resetCache()

	indexWord(1, 0)
	assert.Equal(t, []int{0}, cache[0].cword)

	indexWord(1, 1)
	assert.Equal(t, []int{0, 1}, cache[0].cword)
}

func TestIndexSent(t *testing.T) {
	resetCache()

	indexSent(1, 0)
	assert.Equal(t, []int{0}, cache[0].csent)

	indexSent(1, 1)
	assert.Equal(t, []int{0, 1}, cache[0].csent)
}

func TestCursorPos(t *testing.T) {
	resetCache()
	cursor = counts{Para: 1}
	assert.Equal(t, counts{0, 0, 0, 1}, cursorPos())

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
	ResizeScreen(margin+3, 2)
	setupTest()

	c := 0
	source := []byte{}
	state := -1
	assert.Equal(t, "_", drawLine(1, &c, &source, &state))
	assert.Equal(t, "", drawLine(2, &c, &source, &state))

	source = []byte("1\xff2")
	state = -1
	assert.Equal(t, "_1\xff2", drawLine(1, &c, &source, &state))
	assert.Equal(t, 3, c)

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
	assert.Equal(t, 2, lastSentence(1))

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
	ResizeScreen(margin+1, 3)
	setupTest()

	for scope = Char; scope < MaxScope; scope++ {
		c := 0
		source := []byte{}
		state := -1
		assert.Equal(t, string(cursorChar[scope]), drawLine(1, &c, &source, &state))
	}
}

func TestDrawPara(t *testing.T) {
	ResizeScreen(margin+4, 2)
	setupTest()
	drawPara(1)
	assert.Equal(t, para{text: []string{"_"}}, cache[0])

	doc.SetText(1, "Test")
	drawPara(1)
	assert.Equal(t, para{4, []int{0}, []int{0}, []string{"_Test"}}, cache[0])
	assert.Equal(t, counts{4, 1, 1, 1}, total)

	cursor[Char] = 4
	drawPara(1)
	assert.Equal(t, para{4, []int{0}, []int{0}, []string{"Test_"}}, cache[0])
	assert.Equal(t, 0, cursLine)
	assert.Equal(t, counts{4, 1, 1, 1}, total)

	doc.SetText(1, "One two")
	drawPara(1)
	assert.Equal(t, para{7, []int{0, 4}, []int{0}, []string{"One ", "_two"}}, cache[0])
	assert.Equal(t, 1, cursLine)
	assert.Equal(t, counts{7, 2, 1, 1}, total)

	cursor[Char] = 0
	drawPara(1)
	assert.Equal(t, para{7, []int{0, 4}, []int{0}, []string{"_One ", "two"}}, cache[0])
	assert.Equal(t, counts{7, 2, 1, 1}, total)

	doc.CreateParagraph(2, "")
	defer doc.DeleteParagraph(2)
	drawPara(2)
	assert.Equal(t, para{text: []string{""}}, cache[1])
}

func TestDrawWindow(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 3)
	drawWindow()
	assert.Equal(t, []para{{text: []string{"_"}}}, cache)

	doc.CreateParagraph(2, "Test")
	defer doc.DeleteParagraph(2)
	cursor = counts{4, 0, 0, 2}
	drawWindow()
	expect := []para{{text: []string{""}}, {4, []int{0}, []int{0}, []string{"Test_"}}}
	assert.Equal(t, expect, cache)

	insertParaBreak()
	defer doc.DeleteParagraph(3)
	initialCap = false
	scope = Char
	drawWindow()
	expect = []para{{text: []string{""}}, {4, []int{0}, []int{0}, []string{"Test"}}, {text: []string{"_"}}}
	assert.Equal(t, expect, cache)

	cache = nil
	cursor[Para] = 2
	drawWindow()
	expect = []para{{}, {4, []int{0}, []int{0}, []string{"_Test"}}}
	assert.Equal(t, expect, cache)

	cursor[Para] = 1
	drawWindow()
	expect = []para{{text: []string{"_"}}, {4, []int{0}, []int{0}, []string{"Test"}}}
	assert.Equal(t, expect, cache)

	firstLine = 1
	drawWindow()
	assert.Equal(t, 0, firstLine)

	ResizeScreen(margin+4, 12)
	drawWindow()
	expect = append(expect, para{text: []string{""}})
	assert.Equal(t, expect, cache)

	cache = slices.Delete(cache, 2, 3)
	doc.CreateParagraph(4, "")
	defer doc.DeleteParagraph(4)
	cursor[Para] = 4
	drawWindow()
	expect = []para{{}, {}, {}, {text: []string{"_"}}}
	assert.Equal(t, expect, cache)
}

func TestScreen(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 5)
	doc.SetText(1, "")
	assert.Equal(t, "_\n\n\n\n@0/0", Screen())

	doc.CreateParagraph(2, "")
	defer doc.DeleteParagraph(2)
	cursor[Para] = 2
	assert.Equal(t, "\n\n_\n\n@0/0", Screen())

	cursor[Para] = 1
	assert.Equal(t, "_\n\n\n\n@0/0", Screen())

	doc.CreateParagraph(3, "")
	defer doc.DeleteParagraph(3)
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
