package edits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptBackspace(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	PromptBackspace()
	assert.Zero(responseBeforeLen)

	PromptDefault("Test🇦🇺")
	PromptBackspace()
	assert.Equal("Test", responseBefore)
	assert.Equal(4, responseWidth)
}

func TestPromptDefault(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	assert.Empty(responseBefore)
	assert.Zero(responseBeforeLen)
	assert.Zero(responseWidth)

	input := "🇦🇺"
	PromptDefault(input)
	assert.Equal(input, responseBefore)
	assert.Equal(1, responseBeforeLen)
	assert.Equal(2, responseWidth)
}

func TestPromptInsertRunes(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	ex = promptMargin + 2
	PromptInsertRunes([]rune("🇦🇺"))
	assert.Equal("🇦🇺", responseBefore)
	assert.Equal(1, responseBeforeLen)
	assert.Equal(2, responseWidth)
}

func TestPromptLeft(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	PromptLeft()
	assert.Zero(responseBeforeLen)

	PromptDefault("Test🇦🇺")
	PromptLeft()
	assert.Equal("Test", responseBefore)
	assert.Equal("🇦🇺", responseAfter)
}

func TestPromptRight(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("Test")
	PromptRight()
	assert.Zero(responseAfterLen)

	responseAfter = "🇦🇺!"
	responseAfterLen = 2
	PromptRight()
	assert.Equal("Test🇦🇺", responseBefore)
	assert.Equal("!", responseAfter)
}

func TestPromptHome(t *testing.T) {
	assert := assert.New(t)
	responseBefore, responseBeforeLen, responseAfter, responseAfterLen = "Te", 2, "st", 2
	PromptHome()
	assert.Zero(responseBeforeLen)
	assert.Equal("Test", responseAfter)
}

func TestPromptEnd(t *testing.T) {
	assert := assert.New(t)
	responseBefore, responseBeforeLen, responseAfter, responseAfterLen = "Te", 2, "st", 2
	PromptEnd()
	assert.Equal("Test", responseBefore)
	assert.Zero(responseAfterLen)
}

func TestPromptLine(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("Test")
	message = "Prompt"
	assert.Equal("Prompt Test_", promptLine())

	responseAfter = "ed"
	assert.Equal("Prompt Test_ed", promptLine())
}

func TestPromptResponse(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("Test")
	assert.Equal("Test", PromptResponse())

	responseAfter = "ed"
	assert.Equal("Tested", PromptResponse())
}
