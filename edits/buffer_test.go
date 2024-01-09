package edits

import (
	"testing"
	"unicode/utf8"

	"git.sericyb.com.au/jotty/test"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestScreenRegionFill(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 10)
		sr := &ScreenRegion{s, 1, 2, 5, 5}
		sr.Fill('^', tcell.StyleDefault.Bold(true))
		s.Sync()

		test.AssertCellContents(t, s, [][]rune{
			[]rune("          "),
			[]rune("          "),
			[]rune(" ^^^^^    "),
			[]rune(" ^^^^^    "),
			[]rune(" ^^^^^    "),
			[]rune(" ^^^^^    "),
			[]rune(" ^^^^^    "),
			[]rune("          "),
			[]rune("          "),
			[]rune("          "),
		})
	})
}

func TestDrawStatusBar(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		ID = "Jotty v0"
		s.SetSize(0, 0)
		Screen = s
		assert.NotPanics(t, func() { DrawStatusBar() })

		s.SetSize(4, 2)
		DrawStatusBar()
		s.Sync()
		test.AssertCellContents(t, s, [][]rune{
			[]rune("    "),
			[]rune("c0/0"),
		})

		s.SetSize(15, 2)
		DrawStatusBar()
		s.Sync()
		test.AssertCellContents(t, s, [][]rune{
			[]rune("               "),
			[]rune("Jotty v0  c0/0 "),
		})
	})
}

func TestAppendRune(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(margin+1, 2)
		Screen = s
		cursor.pos = 0
		document = nil
		AppendRune('🇦')
		b := make([]byte, 4)
		utf8.EncodeRune(b, '🇦')
		assert.Equal(t, b, document)
		assert.Equal(t, 1, cursor.pos)
		assert.Equal(t, Char, scope)

		AppendRune('🇺')
		assert.Equal(t, []byte("🇦🇺"), document)
		assert.Equal(t, 1, cursor.pos)

		AppendRune(' ')
		assert.Equal(t, []byte("🇦🇺 "), document)
		assert.Equal(t, 2, cursor.pos)
	})
}

func TestDecScope(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		Screen = s
		cursor.x = 0
		cursor.y = 0

		DecScope()
		s.Sync()
		assert.Equal(t, Sect, scope)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Sect]}})

		DecScope()
		s.Sync()
		assert.Equal(t, Para, scope)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Para]}})
	})
}

func TestIncScope(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		Screen = s
		scope = Sect
		cursor.x = 0
		cursor.y = 0

		IncScope()
		s.Sync()
		assert.Equal(t, Char, scope)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Char]}})

		IncScope()
		s.Sync()
		assert.Equal(t, Word, scope)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Word]}})
	})
}

func TestDrawCursor(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		Screen = s
		scope = Char
		cursor.x = 0
		cursor.y = 0
		DrawCursor()
		s.Sync()
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Char]}})
	})
}

func TestDrawWindow(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(0, 0)
		Screen = s
		assert.NotPanics(t, func() { DrawWindow() })

		s.SetSize(4, 2)
		ResizeScreen()
		test.AssertCellContents(t, s, [][]rune{[]rune("<-->"), []rune("    ")})

		s.SetSize(6, 2)
		ID = "Jotty v0"
		cursor.pos = 1
		document = []byte("🇦🇺")
		scope = Char
		ResizeScreen()
		test.AssertCellContents(t, s, [][]rune{
			[]rune("🇦🇺_   "),
			[]rune("c1/1  "),
		})

		s.SetSize(24, 2)
		document = []byte("🇦🇺 Aussie")
		ResizeScreen()
		test.AssertCellContents(t, s, [][]rune{
			[]rune("🇦🇺_ Aussie              "),
			[]rune(ID + "  c1/8          "),
		})

		document = []byte("🇦🇺 Aussie, Aussie, Aussie")
		DrawWindow()
		Screen.Show()
		test.AssertCellContents(t, s, [][]rune{
			[]rune("🇦🇺_ Aussie, Aussie,     "),
			[]rune(ID + "  c1/18         "),
		})

		s.SetSize(30, 3)
		document = append(document, []byte("\nOi oi")...)
		cursor.pos = 30
		ResizeScreen()
		test.AssertCellContents(t, s, [][]rune{
			[]rune("🇦🇺 Aussie, Aussie, Aussie     "),
			[]rune("Oi oi_                        "),
			[]rune(ID + "  c30/30              "),
		})

		document = append(document, []byte(" oi!")...)
		cursor.pos = 34
		DrawWindow()
		Screen.Show()
		test.AssertCellContents(t, s, [][]rune{
			[]rune("🇦🇺 Aussie, Aussie, Aussie     "),
			[]rune("Oi oi oi!_                    "),
			[]rune(ID + "  c34/34              "),
		})
	})
}
