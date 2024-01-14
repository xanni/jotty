package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	nc "github.com/vit1251/go-ncursesw"
)

func TestWithSimScreen(t *testing.T) {
	assert.NotPanics(t, func() { WithSimScreen(t, func() {}) })
}

func TestAssertCellContents(t *testing.T) {
	WithSimScreen(t, func() {
		nc.ResizeTerm(1, 3)
		nc.StdScr().Clear()
		nc.StdScr().Print("😃!")
		nc.Update()
		AssertCellContents(t, [][]rune{{'😃', '😃', '!'}})
	})
}
