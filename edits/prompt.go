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
	responseBefore, responseAfter       string // Portions of the user response before and after the prompt cursor
	responseBeforeLen, responseAfterLen int    // Number of grapheme clusters before and after the prompt cursor
	responseWidth                       int    // Display width of the response string
)

func PromptBackspace() {
	if responseBeforeLen > 0 {
		s := getChars(responseBeforeLen-1, responseBefore)
		responseWidth -= uniseg.StringWidth(responseBefore[len(s):])
		responseBefore = s
		responseBeforeLen--
	}
}

func PromptDefault(s string) {
	responseAfter, responseBefore = "", s
	responseBeforeLen, responseAfterLen = uniseg.GraphemeClusterCount(s), 0
	responseWidth = uniseg.StringWidth(s)
}

func PromptInsertRunes(runes []rune) {
	if responseWidth < ex-promptMargin {
		s := string(runes)
		responseBefore += s
		responseBeforeLen += uniseg.GraphemeClusterCount(s)
		responseWidth += uniseg.StringWidth(s)
	}
}

func PromptLeft() {
	if responseBeforeLen > 0 {
		s := getChars(responseBeforeLen-1, responseBefore)
		responseAfter = responseBefore[len(s):] + responseAfter
		responseAfterLen++
		responseBefore = s
		responseBeforeLen--
	}
}

func PromptRight() {
	if responseAfterLen > 0 {

		var gc string
		gc, responseAfter, _, _ = uniseg.FirstGraphemeClusterInString(responseAfter, -1)
		responseAfterLen--
		responseBefore += gc
		responseBeforeLen++
	}
}

func PromptHome() {
	responseAfter, responseAfterLen = responseBefore+responseAfter, responseBeforeLen+responseAfterLen
	responseBefore, responseBeforeLen = "", 0
}

func PromptEnd() {
	responseBefore, responseBeforeLen = responseBefore+responseAfter, responseBeforeLen+responseAfterLen
	responseAfter, responseAfterLen = "", 0
}

func promptLine() string {
	return promptStyle(message) + " " + responseStyle(responseBefore) + responseStyle(cursorStyle(responseCursor)) +
		responseStyle(responseAfter) // BUG when screen is resized smaller
}

func PromptResponse() string { return responseBefore + responseAfter }
