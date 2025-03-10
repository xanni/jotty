package edits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptAppend(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	input := "🇦🇺"
	PromptAppend([]rune(input))
	assert.Equal(input, response)
	assert.Equal(1, responseLen)
	assert.Equal(1, responsePos)
	assert.Equal(2, responseWidth)
}

func TestPromptBackspace(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	PromptBackspace()
	assert.Equal("", response)

	PromptDefault("Test🇦🇺")
	PromptBackspace()
	assert.Equal("Test", response)
}

func TestPromptDefault(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("")
	assert.Empty(response)
	assert.Zero(responseLen)
	assert.Zero(responsePos)
	assert.Zero(responseWidth)

	input := "🇦🇺"
	PromptDefault(input)
	assert.Equal(input, response)
	assert.Equal(1, responseLen)
	assert.Equal(1, responsePos)
	assert.Equal(2, responseWidth)
}

func TestPromptLine(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("Test")
	message = "Prompt"
	assert.Equal("Prompt Test_", promptLine())
}

func TestPromptResponse(t *testing.T) {
	assert := assert.New(t)
	PromptDefault("Test")
	assert.Equal("Test", PromptResponse())
}
