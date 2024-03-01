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
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/muesli/termenv"
	"github.com/rivo/uniseg"
)

var ID string  // the program name and version
var ex, ey int // edit window dimensions

var counterChar = [...]rune{'@', '#', '$', '¶', '§'}
var cursorChar = [...]rune{'_', '#', '$', '¶', '§'}
var output = termenv.NewOutput(os.Stdout)

const margin = 5 // up to 3 edit marks, cursor and wrap indicator

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
	beg_b, beg_c int    // cumulative counts at start of line
	end_b, end_c int    // cumulative counts at end of line
	brk          Scope  // whether the line ends in a break
	sectn        int    // section that contains the line
	text         string // rendered line including cursor and marks
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
var cursy int                 // current row in the edit window
var document []byte
var initialCap = true // initial capital at the start of a sentence

func appendParaBreak() {
	i := len(document) - 1
	if document[i] == ' ' {
		document[i] = '\n'
	} else {
		document = append(document, '\n')
		cursor[Char]++
	}
	scope = Para
}

func appendSectnBreak() {
	i := len(document) - 1
	if document[i] == '\n' {
		document[i] = '\f'
		s := &sections[cursor[Sectn]-1]
		s.bpara = s.bpara[:len(s.bpara)-1]
		s.cpara = s.cpara[:len(s.cpara)-1]
	} else {
		document = append(document, '\f')
	}
	drawWindow()

	s := cursor[Sectn] + 1
	cursor = counts{Sectn: s}
	scope = Sectn
	indexSectn(len(document))
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

// Draw the status bar that appears on the last line of the screen
func StatusLine() string {
	var c [MaxScope]string // counters for each scope
	var w int              // width of counters

	s := &sections[cursor[Sectn]-1]
	total := counts{s.chars, len(s.cword), len(s.csent), len(s.cpara), len(sections)}
	for sc := Char; sc <= Sectn; sc++ {
		c[sc] = string(counterChar[sc]) + strconv.Itoa(cursor[sc]) + "/" + strconv.Itoa(total[sc])
		w += uniseg.StringWidth(c[sc])
	}

	var t strings.Builder
	if ex >= len(ID)+w+6 {
		t.WriteString(ID)
		t.WriteString("  ")
	}

	if ex < w+6 {
		t.WriteString(c[scope])
	} else {
		for sc := Sectn; sc >= Char; sc-- {
			if sc != scope {
				t.WriteString(c[sc])
			} else {
				t.WriteString(output.String(c[sc]).Bold().String())
			}
			if sc == Sectn {
				t.WriteByte(':')
			}
			t.WriteByte(' ')
		}
	}

	return t.String()
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

	if passed_end && last_row == len(buffer)-1 {
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
	cpara := sections[cursor[Sectn]-1].cpara
	return p < len(cpara) && c == cpara[p]
}

// Set up a fresh display buffer before redrawing the entire window
func newBuffer() {
	buffer = make([]line, ey)
	cursy = 0
	p := getPara()
	s := sections[cursor[Sectn]-1]
	if p > len(s.cpara)-1 || !isNewParagraph(cursor[Char]) {
		p--
	}
	buffer[0] = line{beg_b: s.bpara[p], beg_c: s.cpara[p], sectn: cursor[Sectn]}
}

// Get the monospace display width of the next breakable segment in source
func nextSegWidth(source []byte) int {
	seg, _, _, _ := uniseg.FirstLineSegment(source, -1) // next breakable segment
	return uniseg.StringWidth(string(seg))
}

// Scroll the edit window 1 or 2 lines and update the buffer and cursor row
func scrollUp(lines int) {
	// TODO implement scrolling region
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

// Append runes to the document.
// TODO implement insertion instead
func AppendRunes(runes []rune) {
	if initialCap && unicode.IsLower(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}

	document = append(document, []byte(string(runes))...)
	initialCap = false
	osectn = 0
	scope = Char

	if cursor[Sectn] != len(sections) {
		cursor[Sectn] = len(sections)
		cursor[Char] = uniseg.GraphemeClusterCount(string(document[sections[cursor[Sectn]-1].bsectn:]))
	} else {
		cursor[Char] = buffer[cursy].beg_c + uniseg.GraphemeClusterCount(string(document[buffer[cursy].beg_b:]))
		drawLine(cursy)
	}

	drawWindow()
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

	drawLine(cursy)
}

func IncScope() {
	if scope == Sectn {
		scope = Char
		initialCap = false
	} else {
		scope++
	}

	drawLine(cursy)
}

// The entire screen including the edits window and status line
func Screen() string {
	var t strings.Builder
	for i := 0; i < ey; i++ {
		t.WriteString(buffer[i].text)
		t.WriteByte('\n')
	}
	t.WriteString(StatusLine())
	return t.String()
}

func cursorString() string {
	cc := '↑'
	if !initialCap {
		cc = cursorChar[scope]
	}
	return output.String(string(cc)).Reverse().Blink().String()
}

// Draw a section or paragraph break, represented respectively by a horizontal
// rule and a blank line, and clear the buffer entry for that line
func drawBreak(y int, brk Scope) {
	buffer[y] = line{}
	if brk == Sectn {
		buffer[y].text = strings.Repeat("─", ex)
	}
}

// Draw one line in the edit window.  Word wraps at the end of the line.
func drawLine(y int) {
	b := buffer[y].beg_b
	c := buffer[y].beg_c
	s := buffer[y].sectn
	var t strings.Builder
	source := document[b:]

	// A new paragraph is often the start of a new word that might not have been recorded yet
	if isNewParagraph(c) && isAlphanumeric(source) {
		indexWord(s, c)
	}

	var f int            // Unicode boundary flags
	m := ex - margin - 1 // Right margin
	var r rune
	state := -1
	var x int // Column position in the line
	for {
		if s == cursor[Sectn] && c == cursor[Char] {
			t.WriteString(cursorString())
			cursy = y
			updateCursorPos()
			m++
			x++
		}

		if len(source) == 0 {
			break
		}

		var g []byte // grapheme cluster
		g, source, f, state = uniseg.Step(source, state)
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
		if r != '\f' && f&uniseg.MaskWord != 0 && isAN {
			indexWord(s, c)
		}

		if r == '\n' || (f&uniseg.MaskSentence != 0 && len(source) > 0) {
			indexSent(s, c)
		}

		if r == '\n' {
			indexPara(s, b, c)
		}

		if r == '\f' {
			indexSectn(b)

			if isAN {
				indexWord(s+1, 0)
			}
		}

		if w > 0 {
			t.Write(g)
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

	switch {
	case r == '\f':
		buffer[y].brk = Sectn
	case r == '\n':
		buffer[y].brk = Para
	case x > m && f != uniseg.LineCanBreak:
		t.WriteString(strings.Repeat(" ", ex-x-1))
		t.WriteString(output.String("-").Reverse().String())
		buffer[y].brk = Word
	default:
		buffer[y].brk = Char
	}

	buffer[y].end_b = b
	buffer[y].end_c = c
	buffer[y].text = t.String()
}

// Draw any paragraph or section break, scroll the window if required, and
// update the next line in the buffer.
func advanceLine(y *int, l *line) {
	if l.brk >= Para {
		if *y < ey {
			drawBreak(*y, l.brk)
		}
		*y++
	}

	if *y >= ey {
		lines := (*y + 1) - ey
		scrollUp(lines) // Always 1 or 2 lines
		*y -= lines
		if l.brk >= Para {
			drawBreak(*y-1, l.brk)
		}
	}

	if l.brk == Sectn {
		l.end_c = 0
		l.sectn++
	}

	if l.brk >= Para && l.sectn == cursor[Sectn] && l.end_c == cursor[Char] {
		cursy = *y
		buffer[*y].text = cursorString()
	}

	l.beg_b = l.end_b
	l.beg_c = l.end_c
	l.brk = Char
	l.text = ""
	buffer[*y] = *l
}

/*
Draw the edit window.

It word wraps, buffers and displays a portion of the document starting from
the line the cursor is on and ending at the last line of the edit window
or the end of the document, whichever comes first.  If the cursor is not
within the screen area, it moves the starting position to bring the cursor
back in view.  It also updates the navigation indexes.
*/
func drawWindow() {
	var y int // current screen coordinates

	// First find the character the cursor is located at on the screen, if possible
	if isCursorInBuffer() {
		y = min(cursy, cursorRow())
	} else {
		// Nothing has been drawn yet, or the cursor is outside the screen: redraw everything
		newBuffer()
	}

	l := buffer[y]
	for l.beg_b < len(document) {
		drawLine(y)
		l = buffer[y]
		s := &sections[l.sectn-1]
		s.chars = max(s.chars, l.end_c)
		y++
		if y >= ey && (l.sectn > cursor[Sectn] || l.end_c >= cursor[Char]) {
			break
		}

		advanceLine(&y, &l)
	}

	for i := y + 1; i < ey; i++ {
		buffer[i] = line{}
	}
}

func ResizeScreen(x, y int) {
	ex, ey = x, y-1
	newBuffer()
	drawLine(cursy)
	drawWindow()
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
			AppendRunes([]rune(" "))
		}
		if unicode.Is(unicode.Sentence_Terminal, lr) {
			scope = Sent
		} else {
			scope = Word
		}
	case Word:
		if lb != ' ' && lb != '\n' && lb != '\f' {
			AppendRunes([]rune(" "))
		}
		if lb == ' ' {
			lr, _ := utf8.DecodeLastRune(document[:i])
			if unicode.In(lr, unicode.L, unicode.N) { // alphanumeric
				document[i] = '.'
				AppendRunes([]rune(" "))
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
	drawWindow()
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
	drawWindow()
}
