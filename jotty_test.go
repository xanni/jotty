package main

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

func TestDrawStatusBarZeroHeight(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 0)
		assert.NotPanics(t, func() { DrawStatusBar(s) })
	})
}

func TestDrawStatusBar(t *testing.T) {
	test.WithSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 2)
		DrawStatusBar(s)
		s.Sync()
		test.AssertCellContents(t, s, [][]rune{
			[]rune("          "),
			[]rune("Jotty v0  "),
		})
	})
}
