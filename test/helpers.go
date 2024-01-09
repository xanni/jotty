package test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func WithSimScreen(t *testing.T, f func(tcell.SimulationScreen)) {
	s := tcell.NewSimulationScreen("")
	require.NotNil(t, s)
	err := s.Init()
	require.NoError(t, err)
	defer s.Fini()
	f(s)
}

func AssertCellContents(t *testing.T, s tcell.SimulationScreen, expectedChars [][]rune) {
	cells, width, height := s.GetContents()
	require.Equal(t, height, len(expectedChars))
	require.Equal(t, width, len(expectedChars[0]))
	for y := 0; y < height; y++ {
		for x := 0; x < width; {
			actualChar := cells[x+y*width].Runes
			expectedChar := expectedChars[y][x : x+len(actualChar)]
			assert.Equal(t, expectedChar, actualChar, "Wrong character at (%d, %d), expected '%c' but got '%c'", x, y, expectedChar, actualChar)
			x += len(actualChar)
		}
	}
}
