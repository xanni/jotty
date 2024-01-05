package test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestWithSimScreen(t *testing.T) {
	WithSimScreen(t, func(s tcell.SimulationScreen) { assert.NotNil(t, s) })
}

func TestAssertCellContents(t *testing.T) {
	WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(1, 1)
		s.SetContent(0, 0, '!', nil, tcell.StyleDefault)
		s.Sync()
		AssertCellContents(t, s, [][]rune{{'!'}})
	})
}
