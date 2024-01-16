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

		DecScope()
		assert.Equal(t, Sect, scope)
		assert.Equal(t, CursorChar[Sect]|nc.A_BLINK, win.MoveInChar(0, 0))

		DecScope()
		assert.Equal(t, Para, scope)
		assert.Equal(t, CursorChar[Para]|nc.A_BLINK, win.MoveInChar(0, 0))
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
		scope = Sect

		IncScope()
		assert.Equal(t, Char, scope)
		assert.Equal(t, CursorChar[Char]|nc.A_BLINK, win.MoveInChar(0, 0))

		IncScope()
		assert.Equal(t, Word, scope)
		assert.Equal(t, CursorChar[Word]|nc.A_BLINK, win.MoveInChar(0, 0))
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
		assert.Equal(t, CursorChar[Char]|nc.A_BLINK, win.MoveInChar(0, 0))
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
		cc := rune(CursorChar[Char] | nc.A_BLINK)
		assert.Equal(t, nc.Char('@'), win.MoveInChar(Sy-1, 0))

		for scope = Word; scope <= Sect; scope++ {
			DrawStatusBar()
			assert.Equal(t, CursorChar[scope], win.MoveInChar(Sy-1, 0))
		}

		scope = Char
		Sx = 26
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune(string(cc) + strings.Repeat(" ", 25)),
			[]rune("§1/1: ¶0/1 $0/1 #0/0 @0/0 "),
		})

		Sx = 36
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune(string(cc) + strings.Repeat(" ", 35)),
			[]rune("Jotty v0  §1/1: ¶0/1 $0/1 #0/0 @0/0 "),
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
		Sy = 2
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
		assert.Equal(t, 18, total[Char])
		assert.Equal(t, 2, total[Word])
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
		cc := rune(CursorChar[Char] | nc.A_BLINK)
		test.AssertCellContents(t, [][]rune{
			[]rune("lengt" + string('-'|nc.A_REVERSE)),
			[]rune("h" + string(cc) + "    "),
			[]rune("@6/6  "),
		})
		assert.Equal(t, "lengt", string(buffer[0].text))
		assert.Equal(t, "h", string(buffer[1].text))
		assert.Equal(t, 1, total[Word])
	})
}

func TestDrawWindowParagraph(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos[Char] = 30
		document = []byte("🇦🇺 Aussie, Aussie, Aussie\nOi oi oi!")
		scope = Char
		ID = "Jotty v0"
		Sx = 30
		Sy = 4
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "Oi oi oi!", string(buffer[2].text))
		assert.Equal(t, counts{34, 6, 2, 2, 1}, total)
	})
}

func TestDrawWindowSection(t *testing.T) {
	test.WithSimScreen(t, func() {
		cursor.pos[Char] = 30
		document = []byte("🇦🇺 Aussie, Aussie, Aussie\fOi oi oi!")
		scope = Char
		ID = "Jotty v0"
		Sx = 30
		Sy = 4
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "", string(buffer[1].text))
		assert.Equal(t, "Oi oi oi!", string(buffer[2].text))
		assert.Equal(t, counts{9, 3, 1, 1, 2}, total)
	})
}

func TestDrawWindowTooSmall(t *testing.T) {
	test.WithSimScreen(t, func() {
		Sx = 2
		Sy = 1
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{{rune('<' | errorStyle), rune('>' | errorStyle)}})

		Sx = 4
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			{rune('<' | errorStyle),
				rune('-' | errorStyle),
				rune('-' | errorStyle),
				rune('>' | errorStyle)},
		})
	})
}

func TestResizeScreen(t *testing.T) {
	test.WithSimScreen(t, func() {
		Sx = 1
		Sy = 1
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Nil(t, buffer)
		test.AssertCellContents(t, [][]rune{{' '}})
	})
}
