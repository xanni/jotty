package edits

import (
	"os"
	"slices"
	"testing"

	ps "git.sericyb.com.au/jotty/permascroll"
	"github.com/muesli/termenv"
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
	initialCap, prevSelected, ShowHelp = false, false, false
	mark, markPara = nil, 0
	primary, secondary = selection{}, selection{}
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
	ps.AppendText(1, "Testing")
	ps.CopyText(1, 0, 7)
	ResizeScreen(20, 3)
	expect := "@0/0"
	assert.Equal(expect, statusLine())

	ResizeScreen(21, 3)
	expect = "¶1/1 $0/0 #0/0 " + expect + " "
	assert.Equal(expect, statusLine())

	ResizeScreen(31, 3)
	expect = ID + "  " + expect
	assert.Equal(expect, statusLine())

	ResizeScreen(46, 3)
	expect += "   cut: Testin" + string(moreChar)
	assert.Equal(expect, statusLine())

	ResizeScreen(47, 3)
	expect = "Jotty v0  ¶1/1 $0/0 #0/0 @0/0    cut: Testing"
	assert.Equal(expect, statusLine())

	ResizeScreen(57, 3)
	expect += "  ESC=Help"
	assert.Equal(expect, statusLine())
}

func TestNextSegWidth(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(0, nextSegWidth([]byte("")))
	assert.Equal(3, nextSegWidth([]byte("One two")))
	assert.Equal(4, nextSegWidth([]byte("One-two")))
}

func TestPreceding(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	cache = []para{{26, []int{0, 5, 12, 16}, []int{0, 12}, []string{"Four words. Two sentences."}}}

	tests := map[string]struct {
		scope      Scope
		end, begin int
	}{
		"Char": {Char, 18, 17}, "Word": {Word, 18, 16}, "Sent": {Sent, 18, 12}, "Para": {Para, 18, 0},
		"Word start": {Word, 0, 0}, "Sent start": {Sent, 0, 0},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			scope = test.scope
			assert.Equal(test.begin, preceding(test.end))
		})
	}
}

func TestFollowing(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	cache = []para{{26, []int{0, 5, 12, 16}, []int{0, 12}, []string{"Four words. Two sentences."}}}

	tests := map[string]struct {
		scope      Scope
		begin, end int
	}{
		"Char": {Char, 2, 3}, "Word": {Word, 2, 5}, "Sent": {Sent, 2, 12}, "Para": {Para, 2, 26},
		"Word boundary": {Word, 5, 12}, "Sent boundary": {Sent, 12, 26},
		"Last word": {Word, 20, 26}, "Last sent": {Sent, 18, 26},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			scope = test.scope
			assert.Equal(test.end, following(test.begin))
		})
	}
}

func TestUpdateSelections(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+3, 2)
	setupTest()
	cache = []para{{26, []int{0, 5, 12, 16}, []int{0, 12}, []string{"Four words. Two sentences."}}}

	tests := map[string]struct {
		cursor             int
		mark               []int
		primary, secondary selection
	}{
		"No marks":      {0, []int{}, selection{}, selection{}},
		"One mark":      {2, []int{2}, selection{2, 3, 0, 0}, selection{1, 2, 0, 0}},
		"Cursor before": {0, []int{2}, selection{0, 2, 0, 0}, selection{2, 3, 0, 0}},
		"Cursor after":  {4, []int{2}, selection{2, 4, 0, 0}, selection{4, 5, 0, 0}},
		"Two marks":     {4, []int{4, 2}, selection{2, 4, 0, 0}, selection{4, 5, 0, 0}},
		"Three marks":   {6, []int{6, 4, 2}, selection{2, 4, 0, 0}, selection{4, 6, 0, 0}},
		"Four marks":    {8, []int{6, 4, 8, 2}, selection{2, 4, 0, 0}, selection{6, 8, 0, 0}},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			cursor[Char] = test.cursor
			markPara, mark = 1, test.mark
			updateSelections()
			assert.Equal(test.primary, primary)
			assert.Equal(test.secondary, secondary)
		})
	}

	cache = append(cache, para{4, []int{0}, []int{0}, []string{"Test"}})
	mark = nil
	cursor[Para], markPara = 2, 2
	scope = Para
	updateSelections()
	assert.False(prevSelected)

	mark = []int{1}
	updateSelections()
	assert.False(prevSelected)

	mark = []int{0}
	updateSelections()
	assert.True(prevSelected)

	cursor[Para] = 1
	updateSelections()
	assert.False(prevSelected)
}

func setProfile(profile termenv.Profile) {
	output = termenv.NewOutput(os.Stdout, termenv.WithProfile(profile))
}

func TestDrawChar(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+3, 2)
	setupTest()
	setProfile(termenv.ANSI)
	defer setProfile(termenv.Ascii)

	prevSelected = true
	primary, secondary = selection{1, 2, 0, 0}, selection{2, 3, 0, 0}
	tests := map[string]struct {
		c, markPara int
		expect      string
	}{
		"Unmarked para": {0, 3, "T"},
		"Before mark":   {0, 1, "T"},
		"Primary":       {1, 1, primaryStyle("T")},
		"Secondary":     {2, 1, secondaryStyle("T")},
		"After mark":    {3, 1, "T"},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			l := line{c: test.c, pn: 1}
			markPara = test.markPara
			l.drawChar([]byte("T"))
			assert.Equal(test.expect, l.t.String())
		})
	}

	l := line{pn: 1}
	markPara = 2
	prevSelected = true
	l.drawChar([]byte("T"))
	assert.Equal(secondaryStyle("T"), l.t.String())

	prevSelected = false
	l.t.Reset()
	l.drawChar([]byte("T"))
	assert.Equal("T", l.t.String())
}

func TestDrawLine(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+3, 2)
	setupTest()

	source := []byte{}
	l := line{pn: 1, source: &source, state: -1}
	assert.Equal("_", l.drawLine())
	l.pn++
	assert.Equal("", l.drawLine())

	source = []byte("1\xff2")
	l = line{pn: 1, source: &source, state: -1}
	assert.Equal("_1\xff2", l.drawLine())
	assert.Equal(3, l.c)

	source = []byte("Test")
	l = line{pn: 1, source: &source, state: -1}
	assert.Equal("_Tes    -", l.drawLine())

	cursor[Char] = 3
	source = []byte("12 3")
	l = line{pn: 1, source: &source, state: -1}
	assert.Equal("12 ", l.drawLine())
	assert.Equal("_3", l.drawLine())

	source = []byte(". Test")
	l = line{pn: 1, source: &source, state: -1}
	assert.Equal(". ", l.drawLine())
	assert.Equal(2, lastSentence(1))

	source = []byte("1\n2")
	l = line{pn: 1, source: &source, state: -1}
	assert.Equal("1", l.drawLine())

	cursor[Char] = 1
	source = []byte("1\u200b2") // Zero-width space
	l = line{pn: 1, source: &source, state: -1}
	assert.Equal("1_2", l.drawLine())

	source = []byte("12  ")
	l = line{pn: 1, source: &source, state: -1}
	assert.Equal("1_2 ", l.drawLine())
}

func TestDrawLineCursor(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+1, 3)
	setupTest()

	for scope = Char; scope < MaxScope; scope++ {
		source := []byte{}
		l := line{pn: 1, source: &source, state: -1}
		assert.Equal(string(cursorChar[scope]), l.drawLine())
	}
}

func TestDrawLineMark(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(margin+2, 3)
	setupTest()

	markPara = 1
	mark = []int{1, 2, 0}
	source := []byte("Test")
	l := line{pn: 1, source: &source, state: -1}
	assert.Equal("|_T|e  -", l.drawLine())
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

	markPara, primary, secondary = 1, selection{cend: 4}, selection{cbegin: 4, cend: 7}
	drawPara(1)
	assert.Equal(selection{0, 4, 0, 4}, primary)
	assert.Equal(selection{4, 7, 4, 7}, secondary)
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

func TestTruncate(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("Test", truncate(4, "Test"))
	assert.Equal("T"+string(moreChar)+"t", truncate(3, "Test"))
}

func TestScreen(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(margin+4, 5)

	Message = "Confirm"
	assert.Equal("_\n\n\n\nConfirm", Screen())

	Message = "test"
	assert.Equal("_\n\n\n\nError: t"+string(moreChar)+"t", Screen())

	Message = ""
	assert.Equal("_\n\n\n\n@0/0", Screen())

	ps.SplitParagraph(1, 0)
	cursor[Para] = 2
	assert.Equal("\n\n_\n\n@0/0", Screen())

	cursor[Para] = 1
	assert.Equal("_\n\n\n\n@0/0", Screen())

	ps.SplitParagraph(2, 0)
	cursor[Para] = 3
	assert.Equal("\n\n_\n\n@0/0", Screen())

	cursor[Para] = 2
	assert.Equal("_\n\n\n\n@0/0", Screen())

	ps.AppendText(1, "A B C D")
	ps.AppendText(2, "1 2 3 4")
	cursor[Para] = 1
	assert.Equal("_A B \nC D\n\n1 2 \n@0/14", Screen())

	cursor[Para] = 3
	assert.Equal("1 2 \n3 4\n\n_\n@14/14", Screen())

	Help = []byte("Test")
	ShowHelp = true
	assert.Equal("   Test\n——————————\n\n_\n@14/14", Screen())

	ShowHelp = false
	assert.Equal("1 2 \n3 4\n\n_\n@14/14", Screen())
}
