package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withSimScreen(t *testing.T, f func(tcell.SimulationScreen)) {
	s := tcell.NewSimulationScreen("")
	require.NotNil(t, s)
	err := s.Init()
	require.NoError(t, err)
	defer s.Fini()
	f(s)
}

func assertCellContents(t *testing.T, s tcell.SimulationScreen, expectedChars [][]rune) {
	cells, width, height := s.GetContents()
	require.Equal(t, len(expectedChars), height)
	require.Equal(t, len(expectedChars[0]), width)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			actualChar := cells[x+y*width].Runes[0]
			expectedChar := expectedChars[y][x]
			assert.Equal(t, expectedChar, actualChar, "Wrong character at (%d, %d), expected '%c' but got '%c'", x, y, expectedChar, actualChar)
		}
	}
}

func TestScreenRegionFill(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 10)
		sr := &ScreenRegion{s, 1, 2, 5, 5}
		sr.Fill('^', tcell.StyleDefault.Bold(true))
		s.Sync()

		assertCellContents(t, s, [][]rune{
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

func assertNoPanic(t *testing.T, f func()) {
	defer func() {
		assert.Nil(t, recover())
	}()
	f()
}

func TestDrawStatusBarZeroHeight(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 0)
		assertNoPanic(t, func() { DrawStatusBar(s) })
	})
}

func TestDrawStatusBar(t *testing.T) {
	withSimScreen(t, func(s tcell.SimulationScreen) {
		s.SetSize(10, 2)
		DrawStatusBar(s)
		s.Sync()
		assertCellContents(t, s, [][]rune{
			[]rune("          "),
			[]rune("Jotty v0  "),
		})
	})
}
