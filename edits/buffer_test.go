package edits

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Setup() {
	ID = "J"
	cursor = counts{Sectn: 1}
	cursy = 0
	document = nil
	initialCap = false
	isectn = []int{0}
	newSection(1)
	scope = Char
}

func TestAppendParaBreak(t *testing.T) {
	Setup()
	document = []byte(" ")
	cursor[Char]++
	ResizeScreen(margin+1, 4)

	appendParaBreak()
	drawWindow()
	assert.Equal(t, 2, cursy)
	assert.Equal(t, []byte("\n"), document)

	appendParaBreak()
	assert.Equal(t, []byte("\n\n"), document)
}

func TestAppendSectnBreak(t *testing.T) {
	Setup()
	document = []byte("\n")
	ResizeScreen(margin+1, 5)

	newBuffer()
	appendSectnBreak()
	assert.Equal(t, []byte("\f"), document)

	appendSectnBreak()
	assert.Equal(t, []byte("\f\f"), document)

	cursor = counts{1, 0, 1, 1, 1}
	document = []byte("\n")
	isectn = []int{0}
	newSection(1)
	newBuffer()
	drawWindow()
	appendSectnBreak()
	assert.Equal(t, []byte("\f"), document)
}

func TestScrollUp(t *testing.T) {
	Setup()
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

func TestAppendRune(t *testing.T) {
	Setup()
	ResizeScreen(margin+4, 4)

	AppendRunes([]rune("•"))
	assert.Equal(t, []byte("•"), document)
	assert.Equal(t, Char, scope)
	assert.Equal(t, 1, cursor[Char])

	AppendRunes([]rune("ü"))
	assert.Equal(t, []byte("•ü"), document)
	assert.Equal(t, 2, cursor[Char])

	initialCap = true
	AppendRunes([]rune("å"))
	assert.Equal(t, []byte("•üÅ"), document)
	assert.Equal(t, 3, cursor[Char])
	assert.Equal(t, []int{1}, iword)
	assert.Equal(t, counts{3, 1, 1, 1, 1}, total)

	AppendRunes([]rune{'\n'})
	assert.Equal(t, []byte("•üÅ\n"), document)
	assert.Equal(t, Char, scope)
	assert.Equal(t, 4, cursor[Char])
	assert.Equal(t, []int{1}, iword)
	assert.Equal(t, counts{4, 1, 2, 2, 1}, total)

	AppendRunes([]rune{'X'})
	assert.Equal(t, []byte("•üÅ\nX"), document)
	assert.Equal(t, 5, cursor[Char])
	assert.Equal(t, []int{1, 4}, iword)
	assert.Equal(t, counts{5, 2, 2, 2, 1}, total)

	appendSectnBreak()
	drawWindow()
	AppendRunes([]rune{'Y'})
	assert.Equal(t, []byte("•üÅ\nX\fY"), document)
	assert.Equal(t, Char, scope)
	assert.Equal(t, 1, cursor[Char])
	assert.Equal(t, []int{0}, iword)
	assert.Equal(t, counts{1, 1, 1, 1, 2}, total)

	cursor = counts{Sectn: 1}
	AppendRunes([]rune{'Z'})
	assert.Equal(t, []byte("•üÅ\nX\fYZ"), document)
	assert.Equal(t, 2, cursor[Char])
	assert.Equal(t, []int{0}, iword)
	assert.Equal(t, counts{2, 1, 1, 1, 2}, total)
}

func TestAppendRuneCluster(t *testing.T) {
	Setup()
	ResizeScreen(margin+3, 3)

	AppendRunes([]rune("🇦"))
	AppendRunes([]rune("🇺"))
	assert.Equal(t, []byte("🇦🇺"), document)
	assert.Equal(t, Char, scope)
	assert.Equal(t, 1, cursor[Char])

	AppendRunes([]rune(" "))
	assert.Equal(t, []byte("🇦🇺 "), document)
	assert.Equal(t, 2, cursor[Char])
}

func TestDecScope(t *testing.T) {
	Setup()
	ResizeScreen(margin+1, 3)

	initialCap = false
	DecScope()
	assert.Equal(t, Sectn, scope)
	assert.Equal(t, string(cursorChar[Sectn]), buffer[0].text)

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
	Setup()
	ResizeScreen(margin+1, 3)

	initialCap = false
	scope = Sectn
	IncScope()
	assert.Equal(t, Char, scope)
	assert.Equal(t, string(cursorChar[Char]), buffer[0].text)

	IncScope()
	assert.Equal(t, Word, scope)
	assert.Equal(t, string(cursorChar[Word]), buffer[0].text)
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

func TestDrawStatusBar(t *testing.T) {
	Setup()
	ID = "Jotty v0"
	initialCap = false
	ResizeScreen(margin+1, 3)

	for scope = Char; scope <= Sectn; scope++ {
		drawStatusBar()
		assert.Equal(t, string(cursorChar[scope]), buffer[0].text)
	}

	scope = Char
	ResizeScreen(26, 3)
	assert.Equal(t, "§1/1: ¶0/1 $0/1 #0/0 @0/0 ", StatusLine)

	ResizeScreen(36, 3)
	assert.Equal(t, "Jotty v0  §1/1: ¶0/1 $0/1 #0/0 @0/0 ", StatusLine)
}

func TestIsCursorInBuffer(t *testing.T) {
	Setup()
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
	assert.True(t, isNewParagraph(0))

	ipara = []index{{}, {2, 2}}
	cursor[Para] = 1
	assert.False(t, isNewParagraph(1))
	assert.True(t, isNewParagraph(2))

	cursor[Para] = 2
	assert.False(t, isNewParagraph(2))
}

func TestDrawLine(t *testing.T) {
	Setup()
	ResizeScreen(margin+3, 2)
	newBuffer()

	state := -1
	drawLine(0, &state)
	assert.Equal(t, rune(0), buffer[0].r)

	document = []byte("1\xff2")
	drawLine(0, &state)
	assert.Equal(t, 2, buffer[0].end_c)

	document = []byte("\f")
	buffer[0] = line{sectn: 1}
	state = -1
	drawLine(0, &state)
	assert.Equal(t, '\f', buffer[0].r)

	document = []byte("Test")
	buffer[0] = line{sectn: 1}
	state = -1
	drawLine(0, &state)
	assert.Equal(t, "_Tes   -", buffer[0].text)

	document = []byte("12 3")
	buffer[0] = line{sectn: 1}
	state = -1
	drawLine(0, &state)
	assert.Equal(t, "_12 ", buffer[0].text)
}

func TestAdvanceLine(t *testing.T) {
	Setup()
	cursor[Char] = 1
	document = []byte("\n")
	ResizeScreen(margin+3, 3)
	assert.Equal(t, 0, buffer[0].sectn)
	assert.Equal(t, 1, cursy)

	cursor = counts{Sectn: 1}
	newBuffer()
	buffer[0].r = '\f'
	document = []byte("\f")
	l := buffer[0]
	y := 1
	advanceLine(&y, &l)
	assert.Equal(t, 1, y)
	assert.Equal(t, 0, buffer[0].sectn)
	assert.Equal(t, strings.Repeat("─", ex), buffer[0].text)

	document = []byte("\f\f")
	buffer[1].r = '\f'
	l = buffer[1]
	y = 2
	advanceLine(&y, &l)
	assert.Equal(t, 1, y)
	assert.Equal(t, 0, buffer[0].sectn)
	assert.Equal(t, strings.Repeat("─", ex), buffer[0].text)
}

func TestDrawWindow(t *testing.T) {
	Setup()
	cursor = counts{1, 0, 1, 1, 1}
	document = []byte("🇦🇺")
	ResizeScreen(margin+2, 3)
	assert.Equal(t, counts{1, 0, 1, 1, 1}, total)

	document = []byte("🇦🇺 Aussie")
	ResizeScreen(24, 3)
	assert.Equal(t, counts{8, 1, 1, 1, 1}, total)

	document = []byte("🇦🇺 Aussie, Aussie, Aussie")
	drawWindow()
	assert.Equal(t, counts{24, 3, 1, 1, 1}, total)
}

func TestDrawWindowLineBreak(t *testing.T) {
	Setup()
	cursor = counts{6, 0, 0, 0, 1}
	document = []byte("length")
	initialCap = false
	ResizeScreen(margin+5, 3)
	assert.Equal(t, "lengt    -\nh_\n@6/6", Screen())
	assert.Equal(t, 1, total[Word])

	document = []byte("length +")
	ResizeScreen(margin+8, 3)
	assert.Equal(t, "length_ \n+\n@6/8", Screen())
	assert.Equal(t, 1, total[Word])
}

func TestDrawWindowSentence(t *testing.T) {
	Setup()
	document = []byte("This is a sentence.")
	ResizeScreen(margin+25, 3)
	assert.Equal(t, []index{{}}, isent)
	assert.Equal(t, counts{19, 4, 1, 1, 1}, total)

	document = append(document, []byte(" More")...)
	drawWindow()
	assert.Equal(t, []index{{}, {20, 20}}, isent)
	assert.Equal(t, counts{24, 5, 2, 1, 1}, total)

	drawWindow()
	assert.Equal(t, []index{{}, {20, 20}}, isent)
}

func TestDrawWindowParagraph(t *testing.T) {
	Setup()
	document = []byte("🇦🇺 Aussie, Aussie, Aussie\nOi oi oi!")
	ResizeScreen(margin+30, 3)
	assert.Equal(t, []index{{}, {32, 25}}, ipara)
	assert.Equal(t, []int{2, 10, 18, 25, 28, 31}, iword)
	assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

	cursor = counts{25, 3, 1, 1, 1}
	newBuffer()
	drawWindow()
	assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

	cursor = counts{5, 1, 1, 1, 1}
	drawWindow()
	assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

	cursor = counts{26, 4, 2, 2, 1}
	newBuffer()
	drawWindow()
	assert.Equal(t, counts{34, 6, 2, 2, 1}, total)
}

func TestDrawWindowSection(t *testing.T) {
	Setup()
	document = []byte("🇦🇺 Aussie, Aussie, Aussie\fOi oi oi!")
	ResizeScreen(margin+30, 3)
	assert.Equal(t, strings.Repeat("─", ex), buffer[0].text)
	assert.Equal(t, counts{24, 3, 1, 1, 2}, total)

	cursor = counts{9, 3, 1, 1, 2}
	newSection(2)
	drawWindow()
	assert.Equal(t, counts{9, 3, 1, 1, 2}, total)
	assert.Equal(t, 1, cursy)

	document = append(document, '\f')
	drawWindow()
	assert.Equal(t, counts{9, 3, 1, 1, 3}, total)

	cursor = counts{Sectn: 1}
	newSection(1)
	drawWindow()
	assert.Equal(t, counts{24, 3, 1, 1, 3}, total)
}

func TestDrawWindowScroll(t *testing.T) {
	Setup()
	ipara = []index{{}}
	cursor = counts{7, 1, 1, 1, 1}
	document = []byte("Scroll test: ")
	ResizeScreen(margin+12, 3)

	document = append(document, []byte("line 3")...)
	drawWindow()
	assert.Equal(t, []line{{0, 0, 7, 7, ' ', 1, "Scroll "}, {7, 7, 13, 13, ' ', 1, "_test: "}}, buffer)

	cursor = counts{14, 2, 1, 1, 1}
	drawWindow()
	assert.Equal(t, []line{{7, 7, 13, 13, ' ', 1, "test: "}, {13, 13, 19, 19, '3', 1, "l_ine 3"}}, buffer)
}

func TestDrawWindowWordCount(t *testing.T) {
	Setup()
	document = []byte("Two words")
	ResizeScreen(margin+5, 3)
	assert.Equal(t, 0, cursor[Word])

	cursor[Char]++
	drawWindow()
	assert.Equal(t, 1, cursor[Word])

	cursor[Char] = 3
	drawWindow()
	assert.Equal(t, 1, cursor[Word])

	cursor[Char]++
	drawWindow()
	assert.Equal(t, 1, cursor[Word])

	cursor[Char]++
	drawWindow()
	assert.Equal(t, 2, cursor[Word])

	cursor[Char] = 9
	drawWindow()
	assert.Equal(t, 2, cursor[Word])
}

func TestSpace(t *testing.T) {
	Setup()
	ResizeScreen(margin+6, 4)
	assert.NotPanics(t, func() { Space() })
	assert.Nil(t, document)

	document = []byte("Test")
	Space()
	assert.Equal(t, "Test ", string(document))
	assert.Equal(t, scope, Word)

	Space()
	assert.Equal(t, "Test. ", string(document))
	assert.Equal(t, scope, Sent)

	Space()
	assert.Equal(t, "Test.\n", string(document))
	assert.Equal(t, scope, Para)

	Space()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	Space()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	cursor = counts{5, 1, 1, 1, 1}
	document = []byte("Test,")
	isectn = []int{0}
	scope = Word
	Space()
	assert.Equal(t, "Test, ", string(document))
	assert.Equal(t, scope, Sent)

	document = []byte("Test.")
	scope = Char
	Space()
	assert.Equal(t, "Test. ", string(document))
	assert.Equal(t, scope, Sent)

	Space()
	assert.Equal(t, "Test.\n", string(document))
	assert.Equal(t, scope, Para)

	cursor = counts{5, 1, 1, 1, 1}
	document = []byte("Test.")
	ipara = []index{{}}
	newBuffer()
	Space()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	cursor = counts{5, 1, 1, 1, 1}
	document = []byte("Test\n")
	ipara = []index{{}, {5, 5}}
	isectn = []int{0}
	newBuffer()
	scope = Para
	Space()
	assert.Equal(t, 2, cursy)
	assert.Equal(t, "Test\f", string(document))
	assert.Equal(t, scope, Sectn)
}

func TestSpaceAfterSpace(t *testing.T) {
	Setup()
	ResizeScreen(margin+1, 3)

	document = []byte("Test, ")
	scope = Char
	Space()
	assert.Equal(t, "Test, ", string(document))
	assert.Equal(t, scope, Word)

	Space()
	assert.Equal(t, "Test, ", string(document))
	assert.Equal(t, scope, Sent)

	document = []byte("Test\n")
	scope = Char
	Space()
	assert.Equal(t, "Test\n", string(document))
	assert.Equal(t, scope, Word)

	Space()
	assert.Equal(t, "Test\n", string(document))
	assert.Equal(t, scope, Sent)

	document = []byte("Test\f")
	scope = Char
	Space()
	assert.Equal(t, "Test\f", string(document))
	assert.Equal(t, scope, Word)

	Space()
	assert.Equal(t, "Test\f", string(document))
	assert.Equal(t, scope, Sent)
}

func TestEnter(t *testing.T) {
	Setup()
	ResizeScreen(margin+6, 3)

	assert.NotPanics(t, func() { Enter() })
	assert.Nil(t, document)

	document = []byte("Test.")
	Enter()
	assert.Equal(t, "Test.\n", string(document))
	assert.Equal(t, scope, Para)

	Enter()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	Enter()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)

	cursor = counts{5, 1, 1, 1, 1}
	document = []byte("Test ")
	isectn = []int{0}
	ipara = []index{{}}
	scope = Word
	newBuffer()
	Enter()
	assert.Equal(t, "Test\n", string(document))
	assert.Equal(t, scope, Para)

	Enter()
	assert.Equal(t, "Test\f", string(document))
	assert.Equal(t, scope, Sectn)

	cursor = counts{Sectn: 1}
	document = []byte("Test.")
	scope = Para
	Enter()
	assert.Equal(t, "Test.\f", string(document))
	assert.Equal(t, scope, Sectn)
}
