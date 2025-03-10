package edits

import "github.com/rivo/uniseg"

/*
Implements a single line of editable text used for prompting the user, for
example for an export file path or search string.
*/

const (
	IconExport     = "💾"
	promptMargin   = 5 // Character cell width of the prompt icon, space, and right margin
	responseCursor = "_"
)

var (
	response      string // A description of the requested input and the user response
	responseLen   int    // Number of grapheme clusters in the response string
	responsePos   int    // Cursor position in the response string
	responseWidth int    // Display width of the response string
)

func PromptAppend(runes []rune) {
	if responseWidth >= ex-promptMargin {
		return
	}

	s := string(runes)
	gc := uniseg.GraphemeClusterCount(s)
	response += s
	responseLen += gc
	responsePos += gc
	responseWidth += uniseg.StringWidth(s)
}

func PromptBackspace() {
	if responseLen > 0 {
		PromptDefault(getChars(responseLen-1, response))
	}
}

func PromptDefault(s string) {
	response = s
	responseLen = uniseg.GraphemeClusterCount(s)
	responsePos = responseLen
	responseWidth = uniseg.StringWidth(s)
}

func promptLine() string {
	return promptStyle(message) + " " + responseStyle(response+cursorStyle(responseCursor))
}

func PromptResponse() string { return response }
