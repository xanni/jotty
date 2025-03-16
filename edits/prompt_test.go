package edits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptBackspace(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	PromptBackspace()
	assert.Equal(responseSegment{}, response[segBefore])

	PromptDefault("Test🇦🇺")
	PromptBackspace()
	assert.Equal(responseSegment{4, 4, "Test"}, response[segBefore])
}

func TestPromptDefault(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	assert.Equal(responseSegment{}, response[segBefore])

	PromptDefault("🇦🇺")
	assert.Equal(responseSegment{1, 2, "🇦🇺"}, response[segBefore])
}

func TestPromptInsertRunes(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	ex = promptMargin + 2
	PromptInsertRunes([]rune("🇦🇺"))
	assert.Equal(responseSegment{1, 2, "🇦🇺"}, response[segBefore])
}

func TestPromptLeft(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	PromptLeft()
	assert.Equal(responseSegment{}, response[segBefore])

	PromptDefault("Test🇦🇺")
	PromptLeft()
	assert.Equal([segments]responseSegment{{}, {4, 4, "Test"}, {1, 2, "🇦🇺"}}, response)

	response = [segments]responseSegment{{1, 2, "🇦🇺"}, {}, {4, 4, "Test"}}
	PromptLeft()
	assert.Equal([segments]responseSegment{{}, {}, {5, 6, "🇦🇺Test"}}, response)
}

func TestPromptRight(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("Test")
	PromptRight()
	assert.Equal(responseSegment{}, response[segAfter])

	response[segAfter] = responseSegment{2, 3, "🇦🇺!"}
	PromptRight()
	assert.Equal(responseSegment{5, 6, "Test🇦🇺"}, response[segBefore])
	assert.Equal(responseSegment{1, 1, "!"}, response[segAfter])
}

func TestPromptHome(t *testing.T) {
	assert := assert.New(t)
	response[segBefore], response[segAfter] = responseSegment{2, 2, "Te"}, responseSegment{2, 2, "st"}
	PromptHome()
	assert.Equal(responseSegment{}, response[segBefore])
	assert.Equal(responseSegment{4, 4, "Test"}, response[segAfter])
}

func TestPromptEnd(t *testing.T) {
	assert := assert.New(t)
	response[segBefore], response[segAfter] = responseSegment{2, 2, "Te"}, responseSegment{2, 2, "st"}
	PromptEnd()
	assert.Equal(responseSegment{4, 4, "Test"}, response[segBefore])
	assert.Equal(responseSegment{}, response[segAfter])
}

func TestPromptLine(t *testing.T) {
	assert := assert.New(t)
	ex, message = 15, "Prompt"
	PromptDefault("Test")
	assert.Equal("Prompt Test_", promptLine())

	response[segAfter] = responseSegment{2, 2, "ed"}
	assert.Equal("Prompt Test_ed", promptLine())

	ex = 14
	assert.Equal("Prompt Test_…", promptLine())

	response[segHidden], response[segBefore] = responseSegment{1, 1, "T"}, responseSegment{3, 3, "est"}
	assert.Equal("Prompt …est_…", promptLine())

	response[segHidden], response[segBefore] = responseSegment{2, 2, "Te"}, responseSegment{2, 2, "st"}
	assert.Equal("Prompt …st_ed", promptLine())

	response[segAfter] = responseSegment{3, 3, "ed."}
	assert.Equal("Prompt …st_e…", promptLine())

	ex = 12
	assert.Equal("Prompt …t_…", promptLine())

	response[segBefore], response[segAfter] = responseSegment{6, 6, "Tested"}, responseSegment{}
	assert.Equal("Prompt …ed_", promptLine())
}

func TestPromptResponse(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("Test")
	assert.Equal("Test", PromptResponse())

	response[segAfter] = responseSegment{2, 2, "ed"}
	assert.Equal("Tested", PromptResponse())
}
