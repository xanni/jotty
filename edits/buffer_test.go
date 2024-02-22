package edits

import (
	"strings"
	"testing"

	"git.sericyb.com.au/jotty/test"
	"github.com/stretchr/testify/assert"
	nc "github.com/vit1251/go-ncursesw"
)

func Setup(x, y int) {
	ID = "J"
	Sx = margin + x
	Sy = y
	nc.ResizeTerm(Sy, Sx)

	cursor = counts{Sectn: 1}
	cursy = 0
	document = nil
	isectn = []int{0}
	newSection(1)
	scope = Char
}

func TestAppendParaBreak(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(1, 4)
		document = []byte(" ")
		cursor[Char]++
		ResizeScreen()

		appendParaBreak()
		assert.Equal(t, 2, cursy)
		assert.Equal(t, []byte("\n"), document)

		appendParaBreak()
		assert.Equal(t, []byte("\n\n"), document)
	})
}

func TestAppendSectnBreak(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(1, 5)
		document = []byte("\n")
		ResizeScreen()

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
		DrawWindow()
		appendSectnBreak()
		assert.Equal(t, []byte("\f"), document)
	})
}

func TestScrollUp(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(4, 4)
		ResizeScreen()

		buffer[2].sectn = 2
		scrollUp(2)
		assert.Equal(t, 0, cursy)

		scrollUp(1)
		assert.Equal(t, 1, cursy)

		cursy = 2
		scrollUp(1)
		assert.Equal(t, 1, cursy)
	})
}

func TestAppendRune(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(4, 4)
		ResizeScreen()

		AppendRune([]byte("•"))
		assert.Equal(t, []byte("•"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 1, cursor[Char])

		AppendRune([]byte("ü"))
		assert.Equal(t, []byte("•ü"), document)
		assert.Equal(t, 2, cursor[Char])

		initialCap = true
		AppendRune([]byte("å"))
		assert.Equal(t, []byte("•üÅ"), document)
		assert.Equal(t, 3, cursor[Char])
		assert.Equal(t, []int{1}, iword)
		assert.Equal(t, counts{3, 1, 1, 1, 1}, total)

		AppendRune([]byte{'\n'})
		assert.Equal(t, []byte("•üÅ\n"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 4, cursor[Char])
		assert.Equal(t, []int{1}, iword)
		assert.Equal(t, counts{4, 1, 2, 2, 1}, total)

		AppendRune([]byte{'X'})
		assert.Equal(t, []byte("•üÅ\nX"), document)
		assert.Equal(t, 5, cursor[Char])
		assert.Equal(t, []int{1, 4}, iword)
		assert.Equal(t, counts{5, 2, 2, 2, 1}, total)

		appendSectnBreak()
		AppendRune([]byte{'Y'})
		assert.Equal(t, []byte("•üÅ\nX\fY"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 1, cursor[Char])
		assert.Equal(t, []int{0}, iword)
		assert.Equal(t, counts{1, 1, 1, 1, 2}, total)

		cursor = counts{Sectn: 1}
		AppendRune([]byte{'Z'})
		assert.Equal(t, []byte("•üÅ\nX\fYZ"), document)
		assert.Equal(t, 2, cursor[Char])
		assert.Equal(t, []int{0}, iword)
		assert.Equal(t, counts{2, 1, 1, 1, 2}, total)
	})
}

func TestAppendRuneCluster(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(3, 3)
		ResizeScreen()

		AppendRune([]byte("🇦"))
		AppendRune([]byte("🇺"))
		assert.Equal(t, []byte("🇦🇺"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 1, cursor[Char])

		AppendRune([]byte(" "))
		assert.Equal(t, []byte("🇦🇺 "), document)
		assert.Equal(t, 2, cursor[Char])
	})
}

func TestDecScope(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(1, 3)
		ResizeScreen()

		initialCap = false
		DecScope()
		assert.Equal(t, Sectn, scope)
		assert.Equal(t, nc.Char(cursorChar[Sectn])|nc.A_BLINK, win.MoveInChar(0, 0))

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
	})
}

func TestIncScope(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(1, 3)
		ResizeScreen()

		initialCap = false
		scope = Sectn
		IncScope()
		assert.Equal(t, Char, scope)
		assert.Equal(t, nc.Char(cursorChar[Char])|nc.A_BLINK, win.MoveInChar(0, 0))

		IncScope()
		assert.Equal(t, Word, scope)
		assert.Equal(t, nc.Char(cursorChar[Word])|nc.A_BLINK, win.MoveInChar(0, 0))
	})
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

func TestDrawCursor(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(1, 3)
		ResizeScreen()

		initialCap = false
		drawCursor()
		assert.Equal(t, nc.Char(cursorChar[Char])|nc.A_BLINK, win.MoveInChar(0, 0))
	})
}

func TestDrawStatusBar(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(1, 3)
		ID = "Jotty v0"
		initialCap = false
		ResizeScreen()

		for scope = Char; scope <= Sectn; scope++ {
			drawStatusBar()
			assert.Equal(t, nc.Char(counterChar[scope]), win.MoveInChar(Sy-1, 0))
		}

		cc := rune(cursorChar[Char] | nc.A_BLINK)
		scope = Char
		Sx = 26
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune(string(cc) + strings.Repeat(" ", 25)),
			[]rune(strings.Repeat(" ", 26)),
			append([]rune("§1/1: ¶0/1 $0/1 #0/0 "), '@'|nc.A_BOLD, '0'|nc.A_BOLD, '/'|nc.A_BOLD, '0'|nc.A_BOLD, ' '),
		})

		Sx = 36
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune(string(cc) + strings.Repeat(" ", 35)),
			[]rune(strings.Repeat(" ", 36)),
			append([]rune("Jotty v0  §1/1: ¶0/1 $0/1 #0/0 "),
				'@'|nc.A_BOLD, '0'|nc.A_BOLD, '/'|nc.A_BOLD, '0'|nc.A_BOLD, ' '),
		})
	})
}

func TestIsCursorInBuffer(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(1, 3)
		ResizeScreen()

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
	})
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
	test.WithSimScreen(t, func() {
		Setup(3, 2)
		win = nc.StdScr()
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
		test.AssertCellContents(t, [][]rune{[]rune(" Tes   " + string('-'|nc.A_REVERSE)), []rune("        ")})

		document = []byte("12 3")
		buffer[0] = line{sectn: 1}
		state = -1
		drawLine(0, &state)
		test.AssertCellContents(t, [][]rune{[]rune(" 12     "), []rune("        ")})
	})
}

func TestAdvanceLine(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(3, 3)
		cursor[Char] = 1
		document = []byte("\n")
		ResizeScreen()
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
		assert.Equal(t, nc.Char(nc.ACS_HLINE), win.MoveInChar(0, 0))

		document = []byte("\f\f")
		buffer[1].r = '\f'
		l = buffer[1]
		y = 2
		advanceLine(&y, &l)
		assert.Equal(t, 1, y)
		assert.Equal(t, 0, buffer[0].sectn)
		assert.Equal(t, nc.Char(nc.ACS_HLINE), win.MoveInChar(0, 0))
	})
}

func TestDrawWindow(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(2, 3)
		cursor = counts{1, 0, 1, 1, 1}
		document = []byte("🇦🇺")
		ResizeScreen()
		assert.Equal(t, counts{1, 0, 1, 1, 1}, total)

		document = []byte("🇦🇺 Aussie")
		Sx = 24
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, counts{8, 1, 1, 1, 1}, total)

		document = []byte("🇦🇺 Aussie, Aussie, Aussie")
		DrawWindow()
		assert.Equal(t, counts{24, 3, 1, 1, 1}, total)
	})
}

func TestDrawWindowLineBreak(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(5, 3)
		cursor = counts{6, 0, 0, 0, 1}
		document = []byte("length")
		initialCap = false
		ResizeScreen()
		cc := rune(cursorChar[Char] | nc.A_BLINK)
		test.AssertCellContents(t, [][]rune{
			[]rune("lengt    " + string('-'|nc.A_REVERSE)),
			[]rune("h" + string(cc) + "        "),
			[]rune("@6/6      "),
		})
		assert.Equal(t, 1, total[Word])

		cursor = counts{6, 0, 0, 0, 1}
		document = []byte("length +")
		Sx = margin + 8
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune("length" + string(cc) + "      "),
			[]rune("+            "),
			[]rune("@6/8         "),
		})
		assert.Equal(t, 1, total[Word])
	})
}

func TestDrawWindowSentence(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(25, 3)
		document = []byte("This is a sentence.")
		ResizeScreen()
		assert.Equal(t, []index{{}}, isent)
		assert.Equal(t, counts{19, 4, 1, 1, 1}, total)

		document = append(document, []byte(" More")...)
		DrawWindow()
		assert.Equal(t, []index{{}, {20, 20}}, isent)
		assert.Equal(t, counts{24, 5, 2, 1, 1}, total)

		DrawWindow()
		assert.Equal(t, []index{{}, {20, 20}}, isent)
	})
}

func TestDrawWindowParagraph(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(30, 3)
		document = []byte("🇦🇺 Aussie, Aussie, Aussie\nOi oi oi!")
		ResizeScreen()
		assert.Equal(t, []index{{}, {32, 25}}, ipara)
		assert.Equal(t, []int{2, 10, 18, 25, 28, 31}, iword)
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

		cursor = counts{25, 3, 1, 1, 1}
		newBuffer()
		DrawWindow()
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

		cursor = counts{5, 1, 1, 1, 1}
		DrawWindow()
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

		cursor = counts{26, 4, 2, 2, 1}
		newBuffer()
		DrawWindow()
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)
	})
}

func TestDrawWindowSection(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(30, 3)
		document = []byte("🇦🇺 Aussie, Aussie, Aussie\fOi oi oi!")
		ResizeScreen()
		assert.Equal(t, nc.Char(nc.ACS_HLINE), win.MoveInChar(0, 0))
		assert.Equal(t, counts{24, 3, 1, 1, 2}, total)

		cursor = counts{9, 3, 1, 1, 2}
		newSection(2)
		DrawWindow()
		assert.Equal(t, counts{9, 3, 1, 1, 2}, total)
		assert.Equal(t, 1, cursy)

		document = append(document, '\f')
		DrawWindow()
		assert.Equal(t, counts{9, 3, 1, 1, 3}, total)

		cursor = counts{Sectn: 1}
		newSection(1)
		DrawWindow()
		assert.Equal(t, counts{24, 3, 1, 1, 3}, total)
	})
}

func TestDrawWindowScroll(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(12, 3)
		ipara = []index{{}}
		cursor = counts{7, 1, 1, 1, 1}
		document = []byte("Scroll test: ")
		ResizeScreen()

		document = append(document, []byte("line 3")...)
		DrawWindow()
		assert.Equal(t, []line{{0, 0, 7, 7, ' ', 1}, {7, 7, 13, 13, ' ', 1}}, buffer)

		cursor = counts{14, 2, 1, 1, 1}
		DrawWindow()
		assert.Equal(t, []line{{7, 7, 13, 13, ' ', 1}, {13, 13, 19, 19, '3', 1}}, buffer)
	})
}

func TestDrawWindowTooSmall(t *testing.T) {
	test.WithSimScreen(t, func() {
		Sx = 2
		Sy = 1
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{{rune('<' | errorStyle), rune('>' | errorStyle)}})

		assert.NotPanics(t, func() { DrawWindow() })

		Sx = 6
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			{rune('<' | errorStyle),
				rune('-' | errorStyle),
				rune('-' | errorStyle),
				rune('-' | errorStyle),
				rune('-' | errorStyle),
				rune('>' | errorStyle)},
		})

		assert.NotPanics(t, func() { DrawWindow() })
	})
}

func TestDrawWindowWordCount(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(5, 3)
		document = []byte("Two words")
		ResizeScreen()
		assert.Equal(t, 0, cursor[Word])

		cursor[Char]++
		DrawWindow()
		assert.Equal(t, 1, cursor[Word])

		cursor[Char] = 3
		DrawWindow()
		assert.Equal(t, 1, cursor[Word])

		cursor[Char]++
		DrawWindow()
		assert.Equal(t, 1, cursor[Word])

		cursor[Char]++
		DrawWindow()
		assert.Equal(t, 2, cursor[Word])

		cursor[Char] = 9
		DrawWindow()
		assert.Equal(t, 2, cursor[Word])
	})
}

func TestResizeScreen(t *testing.T) {
	test.WithSimScreen(t, func() {
		Sx = 2
		Sy = 0
		nc.ResizeTerm(Sy, Sx)
		assert.NotPanics(t, func() { ResizeScreen() })

		Sx = 1
		Sy = 1
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{{' '}})
	})
}

func TestSpace(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(6, 4)
		ResizeScreen()
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
	})
}

func TestSpaceAfterSpace(t *testing.T) {
	test.WithSimScreen(t, func() {
		Sx = 1
		Sy = 1
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

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

	})
}

func TestEnter(t *testing.T) {
	test.WithSimScreen(t, func() {
		Setup(6, 3)
		ResizeScreen()

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
	})
}
