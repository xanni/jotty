package test

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	nc "github.com/vit1251/go-ncursesw"
)

var file *os.File

func init() {
	var err error
	file, err = os.OpenFile("/dev/null", os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("open", err)
	}
}

func WithSimScreen(t *testing.T, f func()) {
	s, err := nc.NewTerm("", file, file)
	require.NoError(t, err)
	defer s.Delete()
	defer s.End()
	f()
}

func AssertCellContents(t *testing.T, expectedChars [][]rune) {
	my, mx := nc.StdScr().MaxYX()
	require.Equal(t, my, len(expectedChars))
	require.Equal(t, mx, len(expectedChars[0]))
	for y := 0; y < my; y++ {
		for x := 0; x < mx; x++ {
			actualChar := nc.StdScr().MoveInChar(y, x)
			expectedChar := nc.Char(expectedChars[y][x])
			assert.Equal(t, expectedChar, actualChar, "Wrong character at (%d, %d), expected '%c' but got '%c'", x, y, expectedChar, actualChar)
		}
	}
}
