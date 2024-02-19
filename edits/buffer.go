package edits

/*
Implements the buffer that represents the visible user interface elements of
the edits window and status line.

In the olden days of ASCII, this would all be simple, but in the modern
Unicode world we try to represent things that are a lot more complicated.
Therefore this file maps between the storage representation of the text as a
UTF-8 encoded stream of bytes and the visual presentation of a window
displaying "characters" (Unicode "grapheme clusters") grouped into "words",
"sentences", "paragraphs" and "sections".

Characters can be made up of one or more Go "runes" and can display as glyphs
occuping one or more monospace grid cells in the terminal window.  "Words" are
defined as strings of characters starting with a Unicode alphanumeric
character (class "Letter" or "Number") and ending at a word boundary as
defined in Unicode Standard Annex #29: Text Segmentation, implemented in the
"uniseg" module. Likewise, "Sentences" are defined as strings of characters
between sentence boundaries as defined in the same Annex and module.

"Paragraphs" are strings of characters between newline '\n' or formfeed '\f'
characters, and are visually represented with a single blank line on the
terminal between each paragraph.  "Sections" represent conceptually distinct
portions of the text, and therefore provide a convenient boundary for demand
paging so that the entire document and corresponding internal indexes do not
have to be memory resident at all times.  The total character, word, sentence
and paragraph counts and cursor position are all expressed as offsets relative
to the start of the current section, so positions in the text are uniquely
identified by the section number starting from 1 and the character count
starting from 0.  There is no section 0.

Within paragraphs, word wrapping and line breaking is performed according to
Unicode Standard Annex #14: Line Breaking Algorithm, also implemented in the
"uniseg" module.  Multiple sections can potentially be on screen at the same
time, and are visually represented by a single terminal row containing a
horizontal rule.  Selections cannot cross section boundaries.
*/

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
	nc "github.com/vit1251/go-ncursesw"
)

var counterChar = [...]rune{'@', '#', '$', '¶', '§'}
var cursorChar = [...]rune{'_', '#', '$', '¶', '§'}
var cursorStyle = nc.A_BLINK
var errorStyle = nc.ColorPair(1) | nc.A_REVERSE

const margin = 5 // up to 3 edit marks, cursor and wrap indicator

var ID string  // the program name and version
var Sx, Sy int // screen dimensions

type Scope int

const (
	Char Scope = iota
	Word
	Sent  // sentence
	Para  // paragraph
	Sectn // section
	MaxScope
)

var scope Scope

type counts [MaxScope]int

// Cache for a single line in the terminal window
type line struct {
	beg_b, beg_c int  // cumulative counts at start of line
	end_b, end_c int  // cumulative counts at end of line
	r            rune // the last rune on the line
	sectn        int  // section that contains the line
}

/*
The "buffer" variable caches the starting and ending byte offsets within the
document, the starting and ending character offsets within the section, the
last rune on the line (to record whether it's a paragraph or section break)
and the current section for each line displayed in the terminal to speed up
rendering by avoiding constantly redrawing the entire terminal window.

The blank lines between paragraphs and sections are represented by all zero
entries and should be skipped.
*/

var buffer []line
var cursor = counts{Sectn: 1} // current position in the section/document
var cursx, cursy int          // current position in the edit window
var document []byte
var initialCap = true // initial capital at the start of a sentence
var total counts
var win *nc.Window

func appendParaBreak() {
	i := len(document) - 1
	if document[i] == ' ' {
		document[i] = '\n'
	} else {
		document = append(document, '\n')
		cursor[Char]++
	}
	scope = Para
	DrawWindow()
}

func appendSectnBreak() {
	i := len(document) - 1
	if document[i] == '\n' {
		document[i] = '\f'
		scanSectn()
	} else {
		document = append(document, '\f')
	}
	DrawWindow()

	s := cursor[Sectn] + 1
	cursor = counts{Sectn: s}
	scope = Sectn
	newSection(s)
	DrawWindow()
}

// Find which screen row contains the character position of the cursor
func cursorRow() (y int) {
	for y = 0; y < len(buffer)-1; y++ {
		if buffer[y].sectn == cursor[Sectn] && buffer[y].end_c >= cursor[Char] {
			break
		}
	}

	return y
}

// Draw the cursor
func drawCursor() {
	cc := '↑'
	if !initialCap {
		cc = cursorChar[scope]
	}

	win.AttrSet(nc.Char(cursorStyle))
	win.MovePrint(cursy, cursx, string(cc))
	win.AttrSet(nc.A_NORMAL)
	win.Move(cursy, cursx)
	win.NoutRefresh()
}

func drawResizeRequest() {
	if Sx < 2 || Sy < 1 {
		return
	}

	win.AttrSet(errorStyle)
	win.MovePrint((Sy-1)/2, 0, "<"+strings.Repeat("-", Sx-2)+">")
	win.AttrSet(nc.A_NORMAL)
	win.NoutRefresh()
}

// Draw a status bar on the last line of the screen, then the cursor
func drawStatusBar() {
	win.Move(Sy-1, 0)
	win.ClearToBottom()

	var c [MaxScope]string // counters for each scope
	var w int              // width of counters
	for s := Char; s <= Sectn; s++ {
		c[s] = string(counterChar[s]) + strconv.Itoa(cursor[s]) + "/" + strconv.Itoa(total[s])
		w += utf8.RuneCountInString(c[s])
	}

	win.Move(Sy-1, 0)
	if Sx >= len(ID)+w+6 {
		win.Print(ID + "  ")
	}

	if Sx < w+6 {
		win.Print(c[scope])
	} else {
		for s := Sectn; s >= Char; s-- {
			if s != scope {
				win.Print(c[s])
			} else {
				win.AttrOn(nc.A_BOLD)
				win.Print(c[s])
				win.AttrOff(nc.A_BOLD)
			}
			if s == Sectn {
				win.AddChar(':')
			}
			win.AddChar(' ')
		}
	}

	drawCursor()
}

// True if the first rune in source is a Unicode letter or number
func isAlphanumeric(source []byte) bool {
	r, _ := utf8.DecodeRune(source)
	return unicode.In(r, unicode.L, unicode.N)
}

// Check if the cursor character position is within the buffered rows, and as
// a special case if the cursor is one character or word below the screen then
// scroll up one line to put the cursor back on screen.
func isCursorInBuffer() bool {
	if len(buffer) == 0 {
		return false
	}

	cur_s, cur_c := cursor[Sectn], cursor[Char]
	var first_row, last_row int
	if buffer[first_row].sectn == 0 {
		first_row++
	}
	for last_row = len(buffer) - 1; last_row > 0 && buffer[last_row].sectn == 0; last_row-- {
	}
	passed_end := cur_s == buffer[last_row].sectn && cur_c > buffer[last_row].end_c+1

	if cur_s < buffer[first_row].sectn ||
		(cur_s == buffer[first_row].sectn && cur_c < buffer[first_row].beg_c) ||
		cur_s > buffer[last_row].sectn ||
		(passed_end && scope >= Sent) {
		return false
	}

	if passed_end {
		scrollUp(1)
	}

	return true
}

// Check if the cursor is at the start of a new paragraph
func isNewParagraph(c int) bool {
	if c == 0 {
		return true
	}

	p := cursor[Para]
	return p < len(ipara) && c == ipara[p].c
}

// Set up a fresh display buffer before redrawing the entire window
func newBuffer() {
	buffer = make([]line, Sy-1)
	cursx = 0
	cursy = 0
	p := getPara()
	if p > len(ipara)-1 || !isNewParagraph(cursor[Char]) {
		p--
	}
	buffer[0] = line{beg_b: ipara[p].b, beg_c: ipara[p].c, sectn: cursor[Sectn]}
}

// Get the monospace display width of the next breakable segment in source
func nextSegWidth(source []byte) int {
	seg, _, _, _ := uniseg.FirstLineSegment(source, -1) // next breakable segment
	return uniseg.StringWidth(string(seg))
}

func scrollUp(lines int) {
	if lines > Sy-1 {
		lines = Sy - 1
	}

	win.Move(Sy-1, 0)
	win.ClearToBottom() // Erase the status line before scrolling
	win.Scroll(lines)
	buffer = append(buffer[lines:], make([]line, lines)...)
	cursy -= lines
	if cursy < 0 {
		if buffer[0].sectn > 0 {
			cursy = 0
		} else {
			cursy = 1
		}
	}
}

// Append a UTF-8 encoded rune to the document.
// TODO implement insertion instead
func AppendRune(rb []byte) {
	r, _ := utf8.DecodeRune(rb)
	if initialCap && unicode.IsLower(r) {
		rb = []byte(string(unicode.ToUpper(r)))
	}

	document = append(document, rb...)
	initialCap = false
	osectn = 0
	scope = Char

	if len(buffer) == 0 || cursor[Sectn] != len(isectn) {
		cursor[Sectn] = len(isectn)
		cursor[Char] = uniseg.GraphemeClusterCount(string(document[isectn[cursor[Sectn]-1]:]))
	} else {
		cursor[Char] = buffer[cursy].beg_c + uniseg.GraphemeClusterCount(string(document[buffer[cursy].beg_b:]))
		state := -1
		drawLine(cursy, &state)
	}

	DrawWindow()
}

func DecScope() {
	if scope == Char {
		scope = Sectn
	} else {
		scope--
	}

	if scope < Sent {
		initialCap = false
	}

	drawStatusBar()
}

func IncScope() {
	if scope == Sectn {
		scope = Char
		initialCap = false
	} else {
		scope++
	}

	drawStatusBar()
}

// Draw a paragraph or section break, represented respectively by a blank line
// and a horizontal rule
func drawBreak(y int, r rune) {
	buffer[y] = line{}
	if r == '\n' {
		win.Move(y, 0)
		win.ClearToEOL()
	} else if r == '\f' {
		win.HLine(y, 0, nc.ACS_HLINE, Sx-1)
	}
}

// Draw one line in the edit window.  Word wraps at the end of the line.
func drawLine(y int, state *int) {
	b := buffer[y].beg_b
	c := buffer[y].beg_c
	s := buffer[y].sectn
	source := document[b:]

	// A new paragraph is often the start of a new word that might not have been recorded yet
	if isNewParagraph(c) && isAlphanumeric(source) {
		indexWord(c)
	}

	var f int            // Unicode boundary flags
	m := Sx - margin - 1 // Right margin
	var r rune
	var x int // Column position in the line
	for {
		if s == cursor[Sectn] && c == cursor[Char] {
			cursx = x
			cursy = y
			updateCursorPos()
			m++
			x++
		}

		if len(source) == 0 {
			break
		}

		var g []byte // grapheme cluster
		g, source, f, *state = uniseg.Step(source, *state)
		b += len(g)
		r, _ = utf8.DecodeRune(g)
		if r == utf8.RuneError {
			continue
		}

		w := f >> uniseg.ShiftWidth // monospace width of character
		if w > 0 || r == '\n' {
			c++
		}

		isAN := isAlphanumeric(source)
		if f&uniseg.MaskWord != 0 && isAN {
			indexWord(c)
		}

		if r == '\n' || (f&uniseg.MaskSentence != 0 && len(source) > 0) {
			indexSent(b, c)
		}

		if r == '\n' {
			indexPara(b, c)
		}

		if r == '\f' {
			indexSectn(b)
			newSection(s + 1)

			if isAN {
				iword = []int{0}
			}
		}

		if w > 0 {
			win.MovePrint(y, x, string(g))
			x += w
		}

		// Break if at margin or mandatory break that is not just end of source
		f &= uniseg.MaskLine
		if x > m ||
			(f == uniseg.LineMustBreak && (len(source) > 0 || uniseg.HasTrailingLineBreak(document))) ||
			(f == uniseg.LineCanBreak && x+w+nextSegWidth(source) > m) {
			break
		}
	}

	if x > m && f != uniseg.LineCanBreak {
		win.MoveAddChar(y, Sx-1, '-'|nc.A_REVERSE)
	} else {
		win.Move(y, x)
		win.ClearToEOL()
	}

	buffer[y].end_b = b
	buffer[y].end_c = c
	buffer[y].r = r
}

// Draw any paragraph or section break, scroll the window if required, and
// update the next line in the buffer.
func advanceLine(y *int, l *line) {
	if l.r == '\n' || l.r == '\f' {
		if *y < Sy-1 {
			drawBreak(*y, l.r)
		}
		*y++
	}

	if *y >= Sy-1 {
		lines := (*y + 2) - Sy
		scrollUp(lines)
		*y -= lines
		drawBreak(*y, l.r)
	}

	if l.r == '\f' {
		l.end_c = 0
		l.sectn++
	}

	if (l.r == '\n' || l.r == '\f') && l.sectn == cursor[Sectn] && l.end_c == cursor[Char] {
		cursx = 0
		cursy = *y
	}

	l.beg_b = l.end_b
	l.beg_c = l.end_c
	l.r = 0
	buffer[*y] = *l
}

/*
Draw the edit window.

It word wraps, buffers and displays a portion of the document starting from
the line the cursor is on and ending at the last line of the edit window
or the end of the document, whichever comes first.  If the cursor is not
within the screen area, it moves the starting position to bring the cursor
back in view.  It also updates the navigation indexes and totals counters.
*/
func DrawWindow() {
	if Sx <= margin || Sy <= 1 {
		return
	}

	var y int // current screen coordinates

	// First find the character the cursor is located at on the screen, if possible
	if isCursorInBuffer() {
		y = min(cursy, cursorRow())
	} else {
		// Nothing has been drawn yet, or the cursor is outside the screen: redraw everything
		newBuffer()
	}

	l := buffer[y]
	state := -1
	for l.beg_b < len(document) {
		drawLine(y, &state)
		l = buffer[y]
		y++
		if y >= Sy-1 && (l.sectn > cursor[Sectn] || l.end_c >= cursor[Char]) {
			break
		}

		advanceLine(&y, &l)
	}

	if y < Sy-1 {
		win.Move(y, 0)
		win.ClearToBottom()
	}

	if l.r == '\f' || l.sectn != cursor[Sectn] {
		scanSectn()
	} else {
		total = counts{max(total[Char], l.end_c), len(iword), len(isent), len(ipara), len(isectn)}
	}

	drawStatusBar()
}

func ResizeScreen() {
	buffer = nil
	win = nc.StdScr()
	win.Clear()

	if Sx > margin && Sy > 2 {
		DrawWindow()
	} else {
		drawResizeRequest()
	}
}

func Space() {
	i := len(document) - 1
	if scope == Sectn || i < 0 {
		return
	}

	lb := document[i]
	switch scope {
	case Char:
		lr, _ := utf8.DecodeLastRune(document)
		if lb != ' ' && lb != '\n' && lb != '\f' {
			AppendRune([]byte{' '})
		}
		if unicode.Is(unicode.Sentence_Terminal, lr) {
			scope = Sent
		} else {
			scope = Word
		}
	case Word:
		if lb != ' ' && lb != '\n' && lb != '\f' {
			AppendRune([]byte{' '})
		}
		if lb == ' ' {
			lr, _ := utf8.DecodeLastRune(document[:i])
			if unicode.In(lr, unicode.L, unicode.N) { // alphanumeric
				document[i] = '.'
				AppendRune([]byte{' '})
			}
		}
		scope = Sent
	case Sent:
		appendParaBreak()
	default: // Para because Sectn has already been excluded above
		appendSectnBreak()
	}

	initialCap = scope >= Sent
	osectn = 0
	drawStatusBar()
}

func Enter() {
	if scope == Sectn || len(document) == 0 {
		return
	}

	if scope <= Sent {
		appendParaBreak()
	} else { // scope == Para
		appendSectnBreak()
	}

	initialCap = true
	osectn = 0
	drawStatusBar()
}
