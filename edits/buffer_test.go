package edits

import (
	"testing"

	"git.sericyb.com.au/jotty/test"
	"github.com/stretchr/testify/assert"
	nc "github.com/vit1251/go-ncursesw"
)

func TestAppendByte(t *testing.T) {
	test.WithSimScreen(t, func() {
		document = nil
		Sx = 1
		Sy = 1
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		AppendByte('!')
		assert.Equal(t, []byte("!"), document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 1, cursor.char)

		document = nil
		Sx = margin + 1
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()

		ch := []byte("🇦🇺")
		for _, b := range ch {
			AppendByte(b)
		}
		assert.Equal(t, ch, document)
		assert.Equal(t, Char, scope)
		assert.Equal(t, 1, cursor.char)

		AppendByte(' ')
		assert.Equal(t, []byte("🇦🇺 "), document)
		assert.Equal(t, 2, cursor.char)
	})
}

func TestDecScope(t *testing.T) {
	test.WithSimScreen(t, func() {
		Sx = margin + 1
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		cursor.char = 0
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
		cursor.char = 0
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

		Sx = margin + 1
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		cc := rune(CursorChar[Char] | nc.A_BLINK)
		test.AssertCellContents(t, [][]rune{
			[]rune(string(cc) + "     "),
			[]rune("c0/0  "),
		})

		Sx = 15
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune(string(cc) + "              "),
			[]rune("#0/0 c0/0      "),
		})

		Sx = 20
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		test.AssertCellContents(t, [][]rune{
			[]rune(string(cc) + "                   "),
			[]rune("Jotty v0  #0/0 c0/0 "),
		})
	})
}

func TestDrawWindow(t *testing.T) {
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

		ID = "Jotty v0"
		cursor.char = 1
		document = []byte("🇦🇺")
		scope = Char
		Sx = 6
		Sy = 2
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺", string(buffer[0].text))
		assert.Equal(t, 1, total.chars)
		assert.Equal(t, 0, total.words)

		document = []byte("🇦🇺 Aussie")
		Sx = 24
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺 Aussie", string(buffer[0].text))
		assert.Equal(t, 8, total.chars)
		assert.Equal(t, 1, total.words)

		document = []byte("🇦🇺 Aussie, Aussie, Aussie")
		DrawWindow()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, ", string(buffer[0].text))
		assert.Equal(t, 18, total.chars)
		assert.Equal(t, 2, total.words)

		document = append(document, []byte("\nOi oi")...)
		cursor.char = 30
		Sx = 30
		Sy = 3
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "Oi oi", string(buffer[1].text))
		assert.Equal(t, 30, total.chars)
		assert.Equal(t, 5, total.words)

		document = append(document, []byte(" oi!")...)
		cursor.char = 34
		DrawWindow()
		assert.Equal(t, "🇦🇺 Aussie, Aussie, Aussie", string(buffer[0].text))
		assert.Equal(t, "Oi oi oi!", string(buffer[1].text))
		assert.Equal(t, 34, total.chars)
		assert.Equal(t, 6, total.words)

		cursor.char = 6
		document = []byte("length")
		Sx = 6
		nc.ResizeTerm(Sy, Sx)
		ResizeScreen()
		cc := rune(CursorChar[Char] | nc.A_BLINK)
		test.AssertCellContents(t, [][]rune{
			[]rune("lengt" + string('-'|nc.A_REVERSE)),
			[]rune("h" + string(cc) + "    "),
			[]rune("c6/6  "),
		})

	})
}
