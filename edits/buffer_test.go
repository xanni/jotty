package edits

import (
	"testing"

	"git.sericyb.com.au/jotty/test"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestAppendRune(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		AppendRune(s, '!')
		assert.Equal(t, []rune{'!'}, primedia)
		assert.Equal(t, scope, Char)
	})
}

func TestDecScope(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)

		DecScope(s)
		s.Sync()
		assert.Equal(t, scope, Sect)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Sect]}})

		DecScope(s)
		s.Sync()
		assert.Equal(t, scope, Para)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Para]}})
	})
}

func TestIncScope(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		scope = Sect

		IncScope(s)
		s.Sync()
		assert.Equal(t, scope, Char)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Char]}})

		IncScope(s)
		s.Sync()
		assert.Equal(t, scope, Word)
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Word]}})
	})
}

func TestDrawCursor(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		scope = Char
		DrawCursor(s)
		s.Sync()
		test.AssertCellContents(t, s, [][]rune{{CursorRune[Char]}})
	})
}

func TestDrawWindow(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(2, 1)
		primedia = []rune{'x'}
		scope = Char
		DrawWindow(s)
		s.Sync()
		test.AssertCellContents(t, s, [][]rune{{'x', CursorRune[Char]}})
	})
}
