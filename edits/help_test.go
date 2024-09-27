package edits

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xanni/jotty/i18n"
)

func TestDropParagraphs(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(6, 4)

	assert.Equal([]string{"One", "Two"}, dropParagraphs([]string{"One", "Two", "Three"}))
	assert.Equal([]string{"Three"}, dropParagraphs([]string{"One", "", "Two", "", "Three"}))

	ResizeScreen(6, 5)
	assert.Equal([]string{"Two", "", "Three"}, dropParagraphs([]string{"One", "", "Two", "", "Three"}))
}

func TestHelpWindow(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(6, 8)

	i18n.HelpText, i18n.HelpWidth = []string{}, 0
	assert.Equal([]string{"——————"}, helpWindow())

	i18n.HelpText, i18n.HelpWidth = []string{"One", "", "Two"}, 3
	assert.Equal([]string{" One", "", " Two", "——————"}, helpWindow())

	i18n.HelpText, i18n.HelpWidth = []string{"Testing", "", "More", "", "Text"}, 7
	assert.Equal([]string{"Testi-", "ng", "", "More", "", "Text", "——————"}, helpWindow())

	ResizeScreen(6, 5)
	assert.Equal([]string{"More", "", "Text", "——————"}, helpWindow())
}

func TestRewrap(t *testing.T) {
	assert := assert.New(t)
	ResizeScreen(6, 2)

	assert.Equal([]string{"Test", "Long", "test", "sente-", "nce"}, rewrap([]string{"Test", "Long test sentence"}))
	assert.Equal([]string{`"A",`, `"B" a-`, "nd", `"C"`}, rewrap([]string{`"A", "B" and "C"`}))
}
