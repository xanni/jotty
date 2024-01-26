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
		document = nil
		Sx = margin + 1
		Sy = 2
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
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		cursor.pos[Char] = 0
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
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		cursor.pos[Char] = 0
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
		Sx = margin + 1
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		scope = Char
		DrawCursor()
		assert.Equal(t, nc.Char(cursorChar[Char])|nc.A_BLINK, win.MoveInChar(0, 0))
	})
}

func TestDrawStatusBar(t *testing.T) {
	test.WithSimScreen(t, func() {
		ID = "Jotty v0"
		sent = []int{0}

		Sx = margin + 1
		Sy = 2
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
			append([]rune("§1/1: ¶0/1 $0/1 #0/0 "), '@'|nc.A_BOLD, '0'|nc.A_BOLD, '/'|nc.A_BOLD, '0'|nc.A_BOLD, ' '),
		})

		Sx = 36
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune(string(cc) + strings.Repeat(" ", 35)),
			append([]rune("Jotty v0  §1/1: ¶0/1 $0/1 #0/0 "),
				'@'|nc.A_BOLD, '0'|nc.A_BOLD, '/'|nc.A_BOLD, '0'|nc.A_BOLD, ' '),
		})
	})
}

func TestDrawWindow(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos[Char] = 1
		document = []byte("🇦🇺")
		scope = Char
		ID = "J"
		Sx = 6
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺", string(buffer[0].text))
		assert.Equal(t, 1, total[Char])
		assert.Equal(t, 0, total[Word])

		document = []byte("🇦🇺 Aussie")
		Sx = 24
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺 Aussie", string(buffer[0].text))
		assert.Equal(t, 8, total[Char])
		assert.Equal(t, 1, total[Word])

		document = []byte("🇦🇺 Aussie, Aussie, Aussie")
		DrawWindow()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, ", string(buffer[0].text))
		assert.Equal(t, "Aussie", string(buffer[1].text))
		assert.Equal(t, 24, total[Char])
		assert.Equal(t, 3, total[Word])
	})
}

func TestDrawWindowInvalidUTF8(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = []byte{'1', 255, '2'}
		ID = "J"
		Sx = 6
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, []byte("12"), buffer[0].text)
		assert.Equal(t, 2, total[Char])
	})
}

func TestDrawWindowLineBreak(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos[Char] = 6
		document = []byte("length")
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
		cursor.pos[Char] = 0
		document = []byte("This is a sentence.")
		ID = "J"
		Sx = 30
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "This is a sentence.", string(buffer[0].text))
		assert.Equal(t, []int{0}, sent)
		assert.Equal(t, counts{19, 4, 1, 1, 1}, total)

		document = append(document, []byte(" More")...)
		DrawWindow()
		assert.Equal(t, "This is a sentence. More", string(buffer[0].text))
		assert.Equal(t, []int{0, 20}, sent)
		assert.Equal(t, counts{24, 5, 2, 1, 1}, total)

		DrawWindow()
		assert.Equal(t, []int{0, 20}, sent)
	})
}

func TestDrawWindowParagraph(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos[Char] = 0
		document = []byte("🇦🇺 Aussie, Aussie, Aussie\nOi oi oi!")
		scope = Char
		sent = []int{0}
		ID = "J"
		Sx = 30
		Sy = 4
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "Oi oi oi!", string(buffer[2].text))
		assert.Equal(t, []int{0, 32}, para)
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)

		DrawWindow()
		assert.Equal(t, []int{0, 32}, para)
	})
}

func TestDrawWindowSection(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos[Char] = 30
		document = []byte("🇦🇺 Aussie, Aussie, Aussie\fOi oi oi!")
		scope = Char
		ID = "J"
		Sx = 30
		Sy = 4
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, nc.Char(nc.ACS_HLINE), win.MoveInChar(1, 0))
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "Oi oi oi!", string(buffer[2].text))
		assert.Equal(t, counts{9, 3, 1, 1, 2}, total)

		document = append(document, '\f')
		DrawWindow()
		assert.Equal(t, counts{0, 0, 1, 1, 3}, total)
	})
}

func TestDrawWindowScroll(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos[Char] = 13
		document = []byte("Scroll test: ")
		ID = "J"
		Sx = 12
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		document = append(document, []byte("line 3")...)
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

func TestSpace(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos[Char] = 0
		document = nil
		Sx = margin + 1
		Sy = 2
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

		document = []byte("Test,")
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
		document = []byte("Test\n")
		scope = Para
		Sy = 3
		Space()
		assert.Equal(t, 1, cursor.y)
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
		cursor.pos[Char] = 0
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
		document = []byte("Test ")
		scope = Word
		Enter()
		assert.Equal(t, "Test\n", string(document))
		assert.Equal(t, scope, Para)

		Sy = 4
		nc.ResizeTerm(Sy, Sx)
		cursor.pos[Char] = 5
		ResizeScreen()
		assert.Equal(t, 2, cursor.y)

		Enter()
		assert.Equal(t, 2, cursor.y)
		assert.Equal(t, "Test\f", string(document))
		assert.Equal(t, scope, Sect)

		cursor.pos[Char] = 0
		cursor.pos[Sect] = 1
		document = []byte("Test.")
		scope = Para
		Enter()
		assert.Equal(t, "Test.\f", string(document))
		assert.Equal(t, scope, Sect)
	})
}
