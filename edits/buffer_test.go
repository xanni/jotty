package edits

import (
	"strings"
	"testing"

	"git.sericyb.com.au/jotty/test"
	"github.com/stretchr/testify/assert"
	nc "github.com/vit1251/go-ncursesw"
)

func TestAppendRune(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = nil
		newSection(1)
		Sx = margin + 1
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		AppendRune([]byte("•"))
		assert.Equal(t, []byte("•"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 1, cursor.pos[Char])

		AppendRune([]byte("ü"))
		assert.Equal(t, []byte("•ü"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 2, cursor.pos[Char])

		initialCap = true
		AppendRune([]byte("å"))
		assert.Equal(t, []byte("•üÅ"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 3, cursor.pos[Char])
		assert.Equal(t, []int{1}, iword)
		assert.Equal(t, counts{3, 1, 1, 1, 1}, total)

		AppendRune([]byte{'\n'})
		assert.Equal(t, []byte("•üÅ\n"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 4, cursor.pos[Char])
		assert.Equal(t, []int{1}, iword)
		assert.Equal(t, counts{4, 1, 2, 2, 1}, total)

		AppendRune([]byte{'X'})
		assert.Equal(t, []byte("•üÅ\nX"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 5, cursor.pos[Char])
		assert.Equal(t, []int{1, 4}, iword)
		assert.Equal(t, counts{5, 2, 2, 2, 1}, total)
	})
}

func TestAppendRuneCluster(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = nil
		Sx = 3
		Sy = 1
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		AppendRune([]byte("🇦"))
		AppendRune([]byte("🇺"))
		assert.Equal(t, []byte("🇦🇺"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 1, cursor.pos[Char])

		AppendRune([]byte(" "))
		assert.Equal(t, []byte("🇦🇺 "), document)
		assert.Equal(t, 2, cursor.pos[Char])
	})
}

func TestDecScope(t *testing.T) {
	test.WithSimScreen(t, func() {
		Sx = margin + 1
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = nil
		ResizeScreen()

		initialCap = false
		scope = Char
		DecScope()
		assert.Equal(t, Sect, scope)
		assert.Equal(t, nc.Char(cursorChar[Sect])|nc.A_BLINK, win.MoveInChar(0, 0))

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
		Sx = margin + 1
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = nil
		ResizeScreen()

		initialCap = false
		scope = Sect
		IncScope()
		assert.Equal(t, Char, scope)
		assert.Equal(t, nc.Char(cursorChar[Char])|nc.A_BLINK, win.MoveInChar(0, 0))

		IncScope()
		assert.Equal(t, Word, scope)
		assert.Equal(t, nc.Char(cursorChar[Word])|nc.A_BLINK, win.MoveInChar(0, 0))
	})
}

func TestDrawCursor(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		Sx = margin + 1
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		initialCap = false
		scope = Char
		DrawCursor()
		assert.Equal(t, nc.Char(cursorChar[Char])|nc.A_BLINK, win.MoveInChar(0, 0))
	})
}

func TestDrawStatusBar(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		newSection(1)
		ID = "Jotty v0"
		initialCap = false

		Sx = margin + 1
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		for scope = Char; scope <= Sect; scope++ {
			DrawStatusBar()
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

func TestCursorRow(t *testing.T) {
	buffer = []line{
		{sect: 2, text: []byte("Two")},
		{},
		{sect: 2, chars: 4, text: []byte("Plus")},
		{},
		{sect: 3, text: []byte("Three")},
	}
	bufy = 4

	cursor.pos = counts{Sect: 2, Char: 3}
	assert.Equal(t, 0, cursorRow())

	cursor.pos = counts{Sect: 2, Char: 4}
	assert.Equal(t, 2, cursorRow())

	cursor.pos = counts{Sect: 3}
	assert.Equal(t, 4, cursorRow())
}

func TestIsCursorInBuffer(t *testing.T) {
	test.WithSimScreen(t, func() {
		Sx = 6
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		buffer = nil
		assert.False(t, isCursorInBuffer())

		buffer = []line{{sect: 2, chars: 5, text: []byte("two ")}, {sect: 2, chars: 9}}
		bufy = 1
		cursor.pos = counts{4, 0, 0, 0, 1}
		assert.False(t, isCursorInBuffer())

		cursor.pos[Sect] = 2
		assert.False(t, isCursorInBuffer())

		bufc = 9
		cursor.pos[Char] = 11
		cursor.y = 0
		scope = Sent
		assert.False(t, isCursorInBuffer())

		cursor.y = 1
		cursor.pos[Char] = 10
		assert.True(t, isCursorInBuffer())
		assert.Equal(t, 1, cursor.y)

		scope = Word
		assert.True(t, isCursorInBuffer())
		assert.Equal(t, 1, cursor.y)

		cursor.pos[Char] = 11
		assert.True(t, isCursorInBuffer())
		assert.Equal(t, 0, cursor.y)
		assert.Equal(t, line{}, buffer[1])

		cursor.pos[Sect] = 3
		assert.False(t, isCursorInBuffer())

		buffer[1].sect = 3
		cursor.pos[Char] = 0
		cursor.y = 1
		assert.True(t, isCursorInBuffer())

		cursor.pos[Sect] = 2
		cursor.pos[Char] = 9
		scope = Sent
		assert.True(t, isCursorInBuffer())
	})
}

func TestIsNewParagraph(t *testing.T) {
	assert.True(t, isNewParagraph(0))

	ipara = []index{{}, {2, 2}}
	cursor.pos[Para] = 1
	assert.False(t, isNewParagraph(1))
	assert.True(t, isNewParagraph(2))

	cursor.pos[Para] = 2
	assert.False(t, isNewParagraph(2))
}

func TestDrawWindow(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{1, 0, 1, 1, 1}
		document = []byte("🇦🇺")
		scope = Char
		newSection(1)
		ID = "J"
		Sx = 6
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺", string(buffer[0].text))
		assert.Equal(t, counts{1, 0, 1, 1, 1}, total)

		document = []byte("🇦🇺 Aussie")
		Sx = 24
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺 Aussie", string(buffer[0].text))
		assert.Equal(t, counts{8, 1, 1, 1, 1}, total)

		document = []byte("🇦🇺 Aussie, Aussie, Aussie")
		DrawWindow()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, ", string(buffer[0].text))
		assert.Equal(t, "Aussie", string(buffer[1].text))
		assert.Equal(t, counts{24, 3, 1, 1, 1}, total)
	})
}

func TestDrawWindowInvalidUTF8(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = []byte("1\xff2")
		ID = "J"
		Sx = 6
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, []byte("12"), buffer[0].text)
		assert.Equal(t, 2, total[Char])
	})
}

func TestDrawWindowLineBreak(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{6, 0, 0, 0, 1}
		iword = nil
		document = []byte("length")
		initialCap = false
		newSection(1)
		ID = "J"
		Sx = 6
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		cc := rune(cursorChar[Char] | nc.A_BLINK)
		test.AssertCellContents(t, [][]rune{
			[]rune("lengt" + string('-'|nc.A_REVERSE)),
			[]rune("h" + string(cc) + "    "),
			[]rune("@6/6  "),
		})
		assert.Equal(t, "lengt", string(buffer[0].text))
		assert.Equal(t, "h", string(buffer[1].text))
		assert.Equal(t, 1, total[Word])

		cursor.pos = counts{6, 0, 0, 0, 1}
		document = []byte("length +")
		Sx = 8
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune("length" + string(cc) + " "),
			[]rune("+       "),
			[]rune("@6/8    "),
		})
		assert.Equal(t, "length ", string(buffer[0].text))
		assert.Equal(t, "+", string(buffer[1].text))
		assert.Equal(t, 1, total[Word])
	})
}

func TestDrawWindowSentence(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = []byte("This is a sentence.")
		newSection(1)
		ID = "J"
		Sx = 30
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "This is a sentence.", string(buffer[0].text))
		assert.Equal(t, []index{{}}, isent)
		assert.Equal(t, counts{19, 4, 1, 1, 1}, total)

		document = append(document, []byte(" More")...)
		DrawWindow()
		assert.Equal(t, "This is a sentence. More", string(buffer[0].text))
		assert.Equal(t, []index{{}, {20, 20}}, isent)
		assert.Equal(t, counts{24, 5, 2, 1, 1}, total)

		DrawWindow()
		assert.Equal(t, []index{{}, {20, 20}}, isent)
	})
}

func TestDrawWindowParagraph(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = []byte("🇦🇺 Aussie, Aussie, Aussie\nOi oi oi!")
		scope = Char
		newSection(1)
		ID = "J"
		Sx = 30
		Sy = 4
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "Oi oi oi!", string(buffer[2].text))
		assert.Equal(t, []index{{}, {32, 25}}, ipara)
		assert.Equal(t, []int{2, 10, 18, 25, 28, 31}, iword)
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

		buffer = nil
		cursor.pos = counts{25, 3, 1, 1, 1}
		DrawWindow()
		assert.Equal(t, "Oi oi oi!", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "", string(buffer[2].text))
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

		cursor.pos = counts{5, 1, 1, 1, 1}
		DrawWindow()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "Oi oi oi!", string(buffer[2].text))
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

		buffer = nil
		cursor.pos = counts{26, 4, 2, 2, 1}
		DrawWindow()
		assert.Equal(t, "Oi oi oi!", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "", string(buffer[2].text))
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)
	})
}

func TestDrawWindowSection(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = []byte("🇦🇺 Aussie, Aussie, Aussie\fOi oi oi!")
		scope = Char
		newSection(1)
		ID = "J"
		Sx = 30
		Sy = 4
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, nc.Char(nc.ACS_HLINE), win.MoveInChar(1, 0))
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "Oi oi oi!", string(buffer[2].text))
		assert.Equal(t, counts{24, 3, 1, 1, 2}, total)

		cursor.pos = counts{9, 3, 1, 1, 2}
		newSection(2)
		DrawWindow()
		assert.Equal(t, counts{9, 3, 1, 1, 2}, total)
		assert.Equal(t, 2, cursor.y)

		document = append(document, '\f')
		DrawWindow()
		assert.Equal(t, counts{9, 3, 1, 1, 3}, total)

		cursor.pos = counts{0, 0, 0, 0, 1}
		newSection(1)
		DrawWindow()
		assert.Equal(t, counts{24, 3, 1, 1, 3}, total)
	})
}

func TestDrawWindowScroll(t *testing.T) {
	test.WithSimScreen(t, func() {
		ipara = []index{{}}
		cursor.pos = counts{7, 1, 1, 1, 1}
		document = []byte("Scroll test: ")
		ID = "J"
		Sx = 12
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		document = append(document, []byte("line 3")...)
		DrawWindow()
		assert.Equal(t, "Scroll ", string(buffer[0].text))
		assert.Equal(t, "test: ", string(buffer[1].text))

		cursor.pos = counts{13, 2, 1, 1, 1}
		DrawWindow()
		assert.Equal(t, "test: ", string(buffer[0].text))
		assert.Equal(t, "line 3", string(buffer[1].text))
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
		cursor.pos = counts{0, 0, 0, 0, 1}
		iword = nil
		document = []byte("Two words")
		scope = Char
		Sx = 10
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, 0, cursor.pos[Word])

		cursor.pos[Char]++
		DrawWindow()
		assert.Equal(t, 1, cursor.pos[Word])

		cursor.pos[Char] = 3
		DrawWindow()
		assert.Equal(t, 1, cursor.pos[Word])

		cursor.pos[Char]++
		DrawWindow()
		assert.Equal(t, 1, cursor.pos[Word])

		cursor.pos[Char]++
		DrawWindow()
		assert.Equal(t, 2, cursor.pos[Word])

		cursor.pos[Char] = 9
		DrawWindow()
		assert.Equal(t, 2, cursor.pos[Word])
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
		assert.Nil(t, buffer)
		test.AssertCellContents(t, [][]rune{{' '}})
	})
}

func TestAppendParaBreak(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = []byte(" ")
		isect = []int{0}
		Sx = margin + 1
		Sy = 4
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		appendParaBreak()
		assert.Equal(t, []byte("\n"), document)

		appendParaBreak()
		assert.Equal(t, []byte("\n\n"), document)
	})
}

func TestAppendSectBreak(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = []byte("\n")
		isect = []int{0}
		Sx = margin + 1
		Sy = 5
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		buffer = nil
		appendSectBreak()
		assert.Equal(t, []byte("\f"), document)

		appendSectBreak()
		assert.Equal(t, []byte("\f\f"), document)

		buffer = nil
		cursor.pos = counts{1, 0, 1, 1, 1}
		document = []byte("\n")
		isectn = []int{0}
		newSection(1)
		DrawWindow()
		appendSectnBreak()
		assert.Equal(t, []byte("\f"), document)
	})
}

func TestSpace(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = nil
		isect = []int{0}
		newSection(1)
		scope = Char
		Sx = margin + 1
		Sy = 4
		nc.ResizeTerm(Sy, Sx)
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
		assert.Equal(t, scope, Sect)

		Space()
		assert.Equal(t, "Test.\f", string(document))
		assert.Equal(t, scope, Sect)

		cursor.pos = counts{5, 1, 1, 1, 1}
		document = []byte("Test,")
		isect = []int{0}
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

		document = []byte("Test.")
		Space()
		assert.Equal(t, "Test.\f", string(document))
		assert.Equal(t, scope, Sect)

		buffer = nil
		ipara = []index{{}, {5, 5}}
		isect = []int{0}
		cursor.pos = counts{5, 1, 1, 1, 1}
		document = []byte("Test\n")
		scope = Para
		Space()
		assert.Equal(t, 0, cursor.y)
		assert.Equal(t, "Test\f", string(document))
		assert.Equal(t, scope, Sect)
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
		cursor.pos = counts{0, 0, 0, 0, 1}
		document = nil
		Sx = margin + 1
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		scope = Char
		assert.NotPanics(t, func() { Enter() })
		assert.Nil(t, document)

		document = []byte("Test.")
		Enter()
		assert.Equal(t, "Test.\n", string(document))
		assert.Equal(t, scope, Para)

		Enter()
		assert.Equal(t, "Test.\f", string(document))
		assert.Equal(t, scope, Sect)

		Enter()
		assert.Equal(t, "Test.\f", string(document))
		assert.Equal(t, scope, Sect)

		buffer = nil
		cursor.pos = counts{5, 1, 1, 1, 1}
		document = []byte("Test ")
		isect = []int{0}
		ipara = []index{{}}
		scope = Word
		Enter()
		assert.Equal(t, "Test\n", string(document))
		assert.Equal(t, scope, Para)

		Enter()
		assert.Equal(t, 0, cursor.y)
		assert.Equal(t, "Test\f", string(document))
		assert.Equal(t, scope, Sect)

		cursor.pos = counts{0, 0, 0, 0, 1}
		document = []byte("Test.")
		scope = Para
		Enter()
		assert.Equal(t, "Test.\f", string(document))
		assert.Equal(t, scope, Sect)
	})
}
