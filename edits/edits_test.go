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

func resetCache() {
	cache = []para{{}}
	cursPara = 1
	total = counts{0, 0, 0, 1}
}

func setupTest() {
	ID = "J"
	cursor = counts{Para: 1}
	firstPara, firstLine = 0, 0
	initialCap = false
	mark = nil
	scope = Char
	ps.Init()
	resetCache()
}

func TestIndexWord(t *testing.T) {
	assert := assert.New(t)
	resetCache()

	indexWord(1, 0)
	assert.Equal([]int{0}, cache[0].cword)

	indexWord(1, 1)
	assert.Equal([]int{0, 1}, cache[0].cword)
}

func TestIndexSent(t *testing.T) {
	assert := assert.New(t)
	resetCache()

	indexSent(1, 0)
	assert.Equal([]int{0}, cache[0].csent)

	indexSent(1, 1)
	assert.Equal([]int{0, 1}, cache[0].csent)
}

func TestCursorPos(t *testing.T) {
	assert := assert.New(t)
	resetCache()

	cursor = counts{Para: 1}
	assert.Equal(counts{0, 0, 0, 1}, cursorPos())

	indexSent(1, 0)
	indexWord(1, 0)
	cursor[Para] = 2
	assert.Equal(counts{0, 1, 1, 2}, cursorPos())
}

func TestCursorString(t *testing.T) {
	assert := assert.New(t)
	initialCap = false
	scope = Char
	assert.Equal(string(cursorChar[Char]), cursorString())

	initialCap = true
	assert.Equal(string(cursorCharCap), cursorString())
}

func TestStatusLine(t *testing.T) {
	assert := assert.New(t)
	setupTest()

	ID = "Jotty v0"
	initialCap = false
	ResizeScreen(6, 3)
	assert.Equal("@0/0", statusLine())

	ResizeScreen(20, 3)
	assert.Equal("¶1/1 $0/0 #0/0 @0/0 ", statusLine())

	ResizeScreen(30, 3)
	assert.Equal("Jotty v0  ¶1/1 $0/0 #0/0 @0/0 ", statusLine())
}

func TestNextSegWidth(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, nextSegWidth([]byte("")))
	assert.Equal(3, nextSegWidth([]byte("One two")))
	assert.Equal(4, nextSegWidth([]byte("One-two")))
}

func TestDrawLine(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+3, 2)
	setupTest()

	source := []byte{}
	l := line{source: &source, state: -1}
	assert.Equal("_", l.drawLine(1))
	assert.Equal("", l.drawLine(2))

	source = []byte("1\xff2")
	l = line{source: &source, state: -1}
	assert.Equal("_1\xff2", l.drawLine(1))
	assert.Equal(3, l.c)

	source = []byte("Test")
	l = line{source: &source, state: -1}
	assert.Equal("_Tes   -", l.drawLine(1))

	cursor[Char] = 3
	source = []byte("12 3")
	l = line{source: &source, state: -1}
	assert.Equal("12 ", l.drawLine(1))
	assert.Equal("_3", l.drawLine(1))

	source = []byte(". Test")
	l = line{source: &source, state: -1}
	assert.Equal(". ", l.drawLine(1))
	assert.Equal(2, lastSentence(1))

	source = []byte("1\n2")
	l = line{source: &source, state: -1}
	assert.Equal("1", l.drawLine(1))

	cursor[Char] = 1
	source = []byte("1\u200b2") // Zero-width space
	l = line{source: &source, state: -1}
	assert.Equal("1_2", l.drawLine(1))

	source = []byte("12  ")
	l = line{source: &source, state: -1}
	assert.Equal("1_2 ", l.drawLine(1))
}

func TestDrawLineCursor(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+1, 3)
	setupTest()

	for scope = Char; scope < MaxScope; scope++ {
		source := []byte{}
		l := line{source: &source, state: -1}
		assert.Equal(string(cursorChar[scope]), l.drawLine(1))
	}
}

func TestDrawLineMark(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+2, 3)
	setupTest()

	markPara = 1
	mark = []int{1, 2, 0}
	source := []byte("Test")
	l := line{source: &source, state: -1}
	assert.Equal("|_T|e -", l.drawLine(1))
}

func TestDrawPara(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+4, 2)
	setupTest()

	drawPara(1)
	assert.Equal(para{text: []string{"_"}}, cache[0])

	ps.AppendText(1, "Test")
	drawPara(1)
	assert.Equal(para{4, []int{0}, []int{0}, []string{"_Test"}}, cache[0])
	assert.Equal(counts{4, 1, 1, 1}, total)

	cursor[Char] = 4
	drawPara(1)
	assert.Equal(para{4, []int{0}, []int{0}, []string{"Test_"}}, cache[0])
	assert.Equal(0, cursLine)
	assert.Equal(counts{4, 1, 1, 1}, total)

	ps.Init()
	ps.AppendText(1, "One two")
	drawPara(1)
	assert.Equal(para{7, []int{0, 4}, []int{0}, []string{"One ", "_two"}}, cache[0])
	assert.Equal(1, cursLine)
	assert.Equal(counts{7, 2, 1, 1}, total)

	cursor[Char] = 0
	drawPara(1)
	assert.Equal(para{7, []int{0, 4}, []int{0}, []string{"_One ", "two"}}, cache[0])
	assert.Equal(counts{7, 2, 1, 1}, total)

	ps.SplitParagraph(1, 7)
	drawPara(2)
	assert.Equal(para{text: []string{""}}, cache[1])
}

func TestDrawWindow(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 3)

	drawWindow()
	assert.Equal([]para{{text: []string{"_"}}}, cache)

	ps.SplitParagraph(1, 0)
	ps.AppendText(2, "Test")
	cursor = counts{4, 0, 0, 2}
	drawWindow()
	expect := []para{{text: []string{""}}, {4, []int{0}, []int{0}, []string{"Test_"}}}
	assert.Equal(expect, cache)

	insertParaBreak()
	initialCap = false
	scope = Char
	drawWindow()
	expect = []para{{text: []string{""}}, {4, []int{0}, []int{0}, []string{"Test"}}, {text: []string{"_"}}}
	assert.Equal(expect, cache)

	cache = nil
	cursor[Para] = 2
	drawWindow()
	expect = []para{{text: []string{""}}, {4, []int{0}, []int{0}, []string{"_Test"}}}
	assert.Equal(expect, cache)

	cursor[Para] = 1
	drawWindow()
	expect = []para{{text: []string{"_"}}, {4, []int{0}, []int{0}, []string{"Test"}}}
	assert.Equal(expect, cache)

	firstLine = 1
	drawWindow()
	assert.Equal(0, firstLine)

	ResizeScreen(margin+4, 12)
	drawWindow()
	expect = append(expect, para{text: []string{""}})
	assert.Equal(expect, cache)

	cache = slices.Delete(cache, 2, 3)
	ps.SplitParagraph(3, 0)
	cursor[Para] = 4
	drawWindow()
	expect[0].text = []string{""}
	expect = append(expect, para{text: []string{"_"}})
	assert.Equal(expect, cache)

	cursor[Para] = 3
	cache = nil
	firstPara = 4
	drawWindow()
	expect[0] = para{}
	expect[2].text = []string{"_"}
	expect[3].text = []string{""}
	assert.Equal(expect, cache)
	assert.Equal(2, firstPara)
}

func TestScreen(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 5)

	assert.Equal("_\n\n\n\n@0/0", Screen())

	ps.SplitParagraph(1, 0)
	cursor[Para] = 2
	assert.Equal("\n\n_\n\n@0/0", Screen())

	cursor[Para] = 1
	assert.Equal("_\n\n\n\n@0/0", Screen())

	ps.SplitParagraph(2, 0)
	cursor[Para] = 3
	assert.Equal("\n\n\n_\n@0/0", Screen())

	cursor[Para] = 2
	assert.Equal("_\n\n\n\n@0/0", Screen())

	ps.AppendText(1, "A B C D")
	ps.AppendText(2, "1 2 3 4")
	cursor[Para] = 1
	assert.Equal("_A B \nC D\n\n1 2 \n@0/14", Screen())

	cursor[Para] = 3
	assert.Equal("1 2 \n3 4\n\n_\n@14/14", Screen())
}
