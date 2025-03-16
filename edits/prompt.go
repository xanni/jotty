package edits

import (
	"strings"

	"github.com/rivo/uniseg"
)

/*
Implements a single line of editable text used for prompting the user, for
example for an export file path or search string.
*/

const (
	IconExport     = "💾"
	promptMargin   = 5 // Character cell width of the prompt icon, space, and right margin
	responseCursor = "_"
)

const (
	segHidden = iota
	segBefore
	segAfter
	segments
)

type responseSegment struct {
	length, width int // Number of grapheme clusters and display width
	text          string
}

var (
	promptWidth int                       // Display width of the prompt
	response    [segments]responseSegment // Segments of the user response offscreen, before and after the prompt cursor
)

func segConcat(a, b responseSegment) responseSegment {
	return responseSegment{a.length + b.length, a.width + b.width, a.text + b.text}
}

func PromptBackspace() {
	if response[segBefore].length > 0 {
		s := getChars(response[segBefore].length-1, response[segBefore].text)
		response[segBefore].width -= uniseg.StringWidth(response[segBefore].text[len(s):])
		response[segBefore].text = s
		response[segBefore].length--
	}
}

func PromptDefault(s string) {
	promptWidth = uniseg.StringWidth(message) + 1
	response = [segments]responseSegment{{}, {uniseg.GraphemeClusterCount(s), uniseg.StringWidth(s), s}, {}}
}

func PromptInsertRunes(runes []rune) {
	s := string(runes)
	response[segBefore] = segConcat(response[segBefore],
		responseSegment{uniseg.GraphemeClusterCount(s), uniseg.StringWidth(s), s})
}

func PromptLeft() {
	seg := segHidden // Which segment is not empty

	if response[segBefore].length > 0 {
		seg = segBefore
	} else if response[segHidden].length == 0 {
		return
	}

	s := getChars(response[seg].length-1, response[seg].text)
	gc := response[seg].text[len(s):]
	width := uniseg.StringWidth(gc)

	response[seg].text = s
	response[seg].length--
	response[seg].width -= width
	response[segAfter] = segConcat(responseSegment{1, width, gc}, response[segAfter])
}

func PromptRight() {
	if response[segAfter].length > 0 {
		var gc string
		gc, response[segAfter].text, _, _ = uniseg.FirstGraphemeClusterInString(response[segAfter].text, -1)
		width := uniseg.StringWidth(gc)

		response[segAfter].length--
		response[segAfter].width -= width
		response[segBefore] = segConcat(response[segBefore], responseSegment{1, width, gc})
	}
}

func PromptHome() {
	response[segAfter] = segConcat(segConcat(response[segHidden], response[segBefore]), response[segAfter])
	response[segHidden] = responseSegment{}
	response[segBefore] = responseSegment{}
}

func PromptEnd() {
	response[segBefore] = segConcat(response[segBefore], response[segAfter])
	response[segAfter] = responseSegment{}
}

// Helper function
func walkString(s string, w int, callback func(string)) (string, int) {
	var gc string
	var total, width int

	for state := -1; w > 0; w -= width {
		gc, s, width, state = uniseg.FirstGraphemeClusterInString(s, state)
		callback(gc)
		total += width
	}

	return s, total
}

func promptLine() string {
	var r strings.Builder
	r.WriteString(promptStyle(message))
	r.WriteRune(' ')

	excess := promptWidth + response[segBefore].width + 2 - ex
	if response[segHidden].length > 0 {
		excess++
	}
	if response[segAfter].length > 0 {
		excess++
	}

	if excess > 0 && response[segBefore].length > 0 { // Scroll left
		if response[segHidden].length == 0 {
			excess++
		}

		s, total := walkString(response[segBefore].text, excess, func(gc string) {
			response[segHidden].text += gc
			response[segHidden].length += len(gc)
		})

		response[segHidden].width += total
		response[segBefore] = responseSegment{len(s), response[segBefore].width - total, s}
	}

	if response[segHidden].length > 0 {
		r.WriteString(truncatedStyle("…"))
	}

	if response[segBefore].length > 0 {
		r.WriteString(responseStyle(response[segBefore].text))
	}

	r.WriteString(responseStyle(cursorStyle(responseCursor)))

	if response[segAfter].length > 0 {
		excess += response[segAfter].width - 1

		if excess < 1 {
			r.WriteString(responseStyle(response[segAfter].text))
		} else { // Truncate
			walkString(response[segAfter].text, response[segAfter].width-excess-1,
				func(gc string) { r.WriteString(responseStyle(gc)) })
			r.WriteString(truncatedStyle("…"))
		}
	}

	return r.String()
}

func PromptResponse() string {
	return response[segHidden].text + response[segBefore].text + response[segAfter].text
}
