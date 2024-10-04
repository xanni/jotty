package edits

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	ps "github.com/xanni/jotty/permascroll"
)

func TestDrawCut(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(4, 2)
	assert.Equal("T…t", drawCut(false, "Test", time.Time{}))

	ResizeScreen(5, 2)
	assert.Equal("Test", drawCut(false, "Test", time.Time{}))

	ResizeScreen(26, 2)
	assert.Equal("                    Test", drawCut(false, "Test", time.Time{}))
	ts := time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC)
	expect := "2020-01-02 03:04:05 Test"
	assert.Equal(expect, drawCut(false, "Test", ts), "unselected")
	assert.Equal(expect, drawCut(true, "Test", ts), "selected")
}

func TestCutsWindow(t *testing.T) {
	assert := assert.New(t)
	setupTest()
	ResizeScreen(5, 2)
	ps.Init("I1,0:Test\nC1,0+4\n")
	currentCut = 1
	assert.Equal([]string{"—————", "Test"}, cutsWindow())

	currentCut = ps.CopyText(1, 0, 1)
	ps.CopyText(1, 2, 4)
	assert.Equal([]string{"—————", "Test", "T", "st"}, cutsWindow())
}

func TestPrevCut(t *testing.T) {
	assert := assert.New(t)
	setupTest()

	PrevCut()
	assert.Equal(None, Mode)
	assert.Equal(0, currentCut)

	ps.Init("I1,0:Test\nC1,0+4\n")
	PrevCut()
	assert.Equal(Cuts, Mode)
	assert.Equal(1, currentCut)

	currentCut = ps.CopyText(1, 1, 3)
	PrevCut()
	assert.Equal(Cuts, Mode)
	assert.Equal(1, currentCut)
}

func TestNextCut(t *testing.T) {
	assert := assert.New(t)
	setupTest()

	NextCut()
	assert.Equal(None, Mode)
	assert.Equal(0, currentCut)

	ps.Init("I1,0:Test\nC1,0+4\n")
	NextCut()
	assert.Equal(Cuts, Mode)
	assert.Equal(1, currentCut)

	currentCut = ps.CopyText(1, 1, 3)
	NextCut()
	assert.Equal(Cuts, Mode)
	assert.Equal(1, currentCut)
}
