package edits

import (
	"testing"

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
		s.SetSize(1, 1)
		Screen = s
		AppendRune('!')
		assert.Equal(t, []rune{'!'}, primedia)
		assert.Equal(t, scope, Char)
	})
}

func TestDecScope(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		Screen = s

		DecScope()
		s.Sync()
		assert.Equal(t, scope, Sect)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Sect]}})

		DecScope()
		s.Sync()
		assert.Equal(t, scope, Para)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Para]}})
	})
}

func TestIncScope(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		Screen = s
		scope = Sect

		IncScope()
		s.Sync()
		assert.Equal(t, scope, Char)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Char]}})

		IncScope()
		s.Sync()
		assert.Equal(t, scope, Word)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Word]}})
	})
}

func TestDrawCursor(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		Screen = s
		scope = Char
		DrawCursor()
		s.Sync()
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Char]}})
	})
}

func TestDrawWindow(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(2, 2)
		Screen = s
		cursor.pos = 0
		primedia = []rune{'x'}
		scope = Char
		DrawWindow()
		s.Sync()
		test.AssertCellContents(t, s, [][]rune{{'x', CursorRune[Char]}, {'c', '0'}})
	})
}
