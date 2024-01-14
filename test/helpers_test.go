package test

import (
	"testing"

	nc "github.com/vit1251/go-ncursesw"
)

func TestAssertCellContents(t *testing.T) {
	WithSimScreen(t, func() {
		nc.ResizeTerm(1, 3)
		nc.StdScr().Clear()
		nc.StdScr().Print("😃!")
		nc.Update()
		AssertCellContents(t, [][]rune{{'😃', '😃', '!'}})
	})
}
