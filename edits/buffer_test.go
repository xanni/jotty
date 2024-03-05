package edits

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTest() {
	ID = "J"
	cursor = counts{Sectn: 1}
	cursy = 0
	document = nil
	initialCap = false
	scope = Char
	resetIndex()
}

func TestScrollUp(t *testing.T) {
	setupTest()
	ResizeScreen(margin+4, 4)

	buffer[2].sectn = 2
	scrollUp(2)
	assert.Equal(t, 0, cursy)

	scrollUp(1)
	assert.Equal(t, 1, cursy)

	cursy = 2
	scrollUp(1)
	assert.Equal(t, 1, cursy)
}

func TestCursorRow(t *testing.T) {
	buffer = []line{{sectn: 2, beg_c: 1, end_c: 4}, {}, {sectn: 2, beg_c: 4, end_c: 5}, {}, {sectn: 3}}
	cursor = counts{Sectn: 2, Char: 3}
	assert.Equal(t, 0, cursorRow())

	cursor = counts{Sectn: 2, Char: 4}
	assert.Equal(t, 0, cursorRow())

	cursor = counts{Sectn: 2, Char: 5}
	assert.Equal(t, 2, cursorRow())

	cursor = counts{Sectn: 3}
	assert.Equal(t, 4, cursorRow())
}

func TestDrawstatusLine(t *testing.T) {
	setupTest()
	ID = "Jotty v0"
	initialCap = false
	ResizeScreen(26, 3)
	assert.Equal(t, "§1/1: ¶0/1 $0/1 #0/0 @0/0 ", statusLine())

	ResizeScreen(36, 3)
	assert.Equal(t, "Jotty v0  §1/1: ¶0/1 $0/1 #0/0 @0/0 ", statusLine())
}

func TestIsCursorInBuffer(t *testing.T) {
	setupTest()
	ResizeScreen(margin+1, 3)

	buffer = []line{{sectn: 2, beg_c: 5, end_c: 9}, {sectn: 2, beg_c: 9, end_c: 9}}
	cursor = counts{4, 0, 0, 0, 1}
	assert.False(t, isCursorInBuffer())

	cursor[Sectn] = 2
	assert.False(t, isCursorInBuffer())

	cursor[Char] = 11
	cursy = 0
	scope = Sent
	assert.False(t, isCursorInBuffer())

	cursy = 1
	cursor[Char] = 10
	assert.True(t, isCursorInBuffer())
	assert.Equal(t, 1, cursy)

	scope = Word
	assert.True(t, isCursorInBuffer())
	assert.Equal(t, 1, cursy)

	cursor[Char] = 11
	assert.True(t, isCursorInBuffer())
	assert.Equal(t, 0, cursy)
	assert.Equal(t, line{}, buffer[1])

	cursor[Sectn] = 3
	assert.False(t, isCursorInBuffer())

	buffer[1].sectn = 3
	cursor[Char] = 0
	cursy = 1
	assert.True(t, isCursorInBuffer())

	cursor[Sectn] = 2
	cursor[Char] = 9
	scope = Sent
	assert.True(t, isCursorInBuffer())
}

func TestIsNewParagraph(t *testing.T) {
	setupTest()
	assert.True(t, isNewParagraph(0))

	sections[0].bpara = []int{0, 2}
	sections[0].cpara = []int{0, 2}
	cursor[Para] = 1
	assert.False(t, isNewParagraph(1))
	assert.True(t, isNewParagraph(2))

	cursor[Para] = 2
	assert.False(t, isNewParagraph(2))
}

func TestCursorString(t *testing.T) {
	initialCap = false
	scope = Char
	assert.Equal(t, string(cursorChar[Char]), cursorString())

	initialCap = true
	assert.Equal(t, string(cursorCharCap), cursorString())
}

func TestDrawLine(t *testing.T) {
	setupTest()
	ResizeScreen(margin+3, 2)
	newBuffer()

	drawLine(0)
	assert.Equal(t, Char, buffer[0].brk)

	document = []byte("1\xff2")
	drawLine(0)
	assert.Equal(t, 2, buffer[0].end_c)

	document = []byte("\f")
	buffer[0] = line{sectn: 1}
	drawLine(0)
	assert.Equal(t, Sectn, buffer[0].brk)

	document = []byte("Test")
	buffer[0] = line{sectn: 1}
	drawLine(0)
	assert.Equal(t, "_Tes   -", buffer[0].text)

	document = []byte("12 3")
	buffer[0] = line{sectn: 1}
	drawLine(0)
	assert.Equal(t, "_12 ", buffer[0].text)
}

func TestDrawLineCursor(t *testing.T) {
	setupTest()
	initialCap = false
	ResizeScreen(margin+1, 3)

	for scope = Char; scope <= Sectn; scope++ {
		drawLine(0)
		assert.Equal(t, string(cursorChar[scope]), buffer[0].text)
	}
}

func TestAdvanceLine(t *testing.T) {
	setupTest()
	cursor[Char] = 1
	document = []byte("\n")
	ResizeScreen(margin+3, 3)
	drawWindow()
	assert.Equal(t, 0, buffer[0].sectn)
	assert.Equal(t, 1, cursy)

	cursor = counts{Sectn: 1}
	newBuffer()
	buffer[0].brk = Sectn
	document = []byte("\f")
	l := buffer[0]
	y := 1
	advanceLine(&y, &l)
	assert.Equal(t, 1, y)
	assert.Equal(t, 0, buffer[0].sectn)
	assert.Equal(t, strings.Repeat("─", ex), buffer[0].text)

	document = []byte("\f\f")
	buffer[1].brk = Sectn
	l = buffer[1]
	y = 2
	advanceLine(&y, &l)
	assert.Equal(t, 1, y)
	assert.Equal(t, 0, buffer[0].sectn)
	assert.Equal(t, strings.Repeat("─", ex), buffer[0].text)
}

func TestDrawWindow(t *testing.T) {
	setupTest()
	cursor = counts{1, 0, 1, 1, 1}
	document = []byte("🇦🇺")
	ResizeScreen(margin+2, 3)
	drawWindow()
	assert.Equal(t, 1, sections[0].chars)
	assert.Equal(t, []int(nil), sections[0].cword)

	document = []byte("🇦🇺 Aussie")
	ResizeScreen(24, 3)
	drawWindow()
	assert.Equal(t, 8, sections[0].chars)
	assert.Equal(t, []int{2}, sections[0].cword)

	document = []byte("🇦🇺 Aussie, Aussie, Aussie")
	drawWindow()
	assert.Equal(t, 24, sections[0].chars)
	assert.Equal(t, []int{2, 10, 18}, sections[0].cword)
}

func TestDrawWindowLineBreak(t *testing.T) {
	setupTest()
	cursor = counts{6, 0, 0, 0, 1}
	document = []byte("length")
	initialCap = false
	ResizeScreen(margin+5, 3)
	assert.Equal(t, "lengt    -\nh_\n@6/6", Screen())
	assert.Equal(t, []int{0}, sections[0].cword)

	document = []byte("length +")
	ResizeScreen(margin+8, 3)
	assert.Equal(t, "length_ \n+\n@6/8", Screen())
	assert.Equal(t, []int{0}, sections[0].cword)
}

func TestDrawWindowSentence(t *testing.T) {
	setupTest()
	document = []byte("This is a sentence.")
	ResizeScreen(margin+25, 3)
	drawWindow()
	assert.Equal(t, 19, sections[0].chars)
	assert.Equal(t, []int{0}, sections[0].csent)

	document = append(document, []byte(" More")...)
	drawWindow()
	assert.Equal(t, 24, sections[0].chars)
	assert.Equal(t, []int{0, 20}, sections[0].csent)
}

func TestDrawWindowParagraph(t *testing.T) {
	setupTest()
	document = []byte("🇦🇺 Aussie, Aussie, Aussie\nOi oi oi!")
	ResizeScreen(margin+30, 3)
	drawWindow()
	assert.Equal(t, 34, sections[0].chars)
	assert.Equal(t, []int{2, 10, 18, 25, 28, 31}, sections[0].cword)
	assert.Equal(t, []int{0, 25}, sections[0].cpara)

	cursor = counts{25, 3, 1, 1, 1}
	newBuffer()
	drawWindow()
	assert.Equal(t, 34, sections[0].chars)
	assert.Equal(t, []int{0, 25}, sections[0].cpara)

	cursor = counts{5, 1, 1, 1, 1}
	drawWindow()
	assert.Equal(t, 34, sections[0].chars)
	assert.Equal(t, []int{0, 25}, sections[0].cpara)

	cursor = counts{26, 4, 2, 2, 1}
	newBuffer()
	drawWindow()
	assert.Equal(t, 34, sections[0].chars)
	assert.Equal(t, []int{0, 25}, sections[0].cpara)
}

func TestDrawWindowSection(t *testing.T) {
	setupTest()
	document = []byte("🇦🇺 Aussie, Aussie, Aussie\fOi oi oi!")
	ResizeScreen(margin+30, 3)
	drawWindow()
	assert.Equal(t, strings.Repeat("─", ex), buffer[0].text)
	assert.Equal(t, 24, sections[0].chars)
	assert.Equal(t, 2, len(sections))

	cursor = counts{9, 3, 1, 1, 2}
	drawWindow()
	assert.Equal(t, 9, sections[1].chars)
	assert.Equal(t, 2, len(sections))
	assert.Equal(t, 1, cursy)

	document = append(document, '\f')
	drawWindow()
	assert.Equal(t, 9, sections[1].chars)
	assert.Equal(t, 3, len(sections))

	cursor = counts{Sectn: 1}
	drawWindow()
	assert.Equal(t, 24, sections[0].chars)
	assert.Equal(t, 3, len(sections))
}

func TestDrawWindowScroll(t *testing.T) {
	setupTest()
	cursor = counts{7, 1, 1, 1, 1}
	document = []byte("Scroll test: ")
	ResizeScreen(margin+12, 3)

	document = append(document, []byte("line 3")...)
	drawWindow()
	assert.Equal(t, []line{{0, 0, 7, 7, Char, 1, "Scroll "}, {7, 7, 13, 13, Char, 1, "_test: "}}, buffer)

	cursor = counts{14, 2, 1, 1, 1}
	drawWindow()
	assert.Equal(t, []line{{7, 7, 13, 13, Char, 1, "test: "}, {13, 13, 19, 19, Char, 1, "l_ine 3"}}, buffer)
}
