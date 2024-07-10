package edits

/*
Implements the buffer that represents the visible user interface elements of the
edits window and status line.

In the olden days of ASCII, this would all be simple, but in the modern Unicode
world we try to represent things that are a lot more complicated. Therefore this
file maps between the storage representation of the text as a UTF-8 encoded
stream of bytes and the visual presentation of a window displaying "characters"
(Unicode "grapheme clusters") grouped into "words", "sentences", and
"paragraphs".

Characters can be made up of one or more Go "runes" and can display as glyphs
occuping one or more monospace grid cells in the terminal window.  "Words" are
defined as strings of characters starting with a Unicode alphanumeric character
(class "Letter" or "Number") and ending at a word boundary as defined in Unicode
Standard Annex #29: Text Segmentation, implemented in the "uniseg" module.
Likewise, "Sentences" are defined as strings of characters between sentence
boundaries as defined in the same Annex and module.

"Paragraphs" are strings of characters between newline '\n' characters, and are
visually represented with a single blank line on the terminal between each
paragraph.  Within paragraphs, word wrapping and line breaking is performed
according to Unicode Standard Annex #14: Line Breaking Algorithm, also
implemented in the "uniseg" module.
*/

import (
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/muesli/termenv"
	"github.com/rivo/uniseg"
)

var ID string // The program name and version
var counterChar = [...]rune{'@', '#', '$', '¶'}
var cursorChar = [...]rune{'_', '#', '$', '¶'}
var curs_buff, curs_line int   // Buffer and line containing the cursor
var ex, ey int                 // Edit window dimensions
var first_buff, first_line int // Buffer and line at top of edit window
var output = termenv.NewOutput(os.Stdout)

const cursorCharCap = '↑'
const margin = 5 // Up to 3 edit marks, cursor and wrap indicator

type Scope int

const (
	Char Scope = iota // Character
	Word
	Sent // Sentence
	Para // Paragraph
	MaxScope
)

type counts [MaxScope]int

// Information about a single paragraph in the terminal window
type para struct {
	pn   int      // Paragraph number in the document
	text []string // Rendered lines including cursor and marks
}

/*
The "buffer" variable caches the paragraphs displayed in the terminal in order
to speed up rendering by avoiding constantly redrawing the entire window.

The "cursor" variable contains the current cursor position for navigation
purposes.
*/

var buffer []para
var cursor = counts{Para: 1} // Current cursor position
var initialCap = true        // Initial capital at the start of a sentence
var scope Scope

// The current cursor position within the whole document, not just the paragraph
func cursorPos() (c counts) {
	c = cursor
	for i := 0; i < c[Para]-1; i++ {
		p := paras[i]
		c[Sent] += len(p.csent)
		c[Word] += len(p.cword)
		c[Char] += p.chars
	}

	return c
}

// The current cursor character and terminal attributes
func cursorString() string {
	cc := cursorCharCap
	if !initialCap {
		cc = cursorChar[scope]
	}
	return output.String(string(cc)).Reverse().Blink().String()
}

// Draw the status bar that appears on the last line of the screen
func statusLine() string {
	const separators = 4   // One space after each scope
	var c [MaxScope]string // Counters for each scope
	var w int              // Total width of counters in character cells

	current := cursorPos()
	total[Para] = len(paras)
	for sc := Char; sc <= Para; sc++ {
		c[sc] = string(counterChar[sc]) + strconv.Itoa(current[sc]) + "/" + strconv.Itoa(total[sc])
		w += uniseg.StringWidth(c[sc])
	}

	var t strings.Builder
	if ex >= len(ID)+w+separators {
		t.WriteString(ID)
		t.WriteString("  ")
	}

	if ex < w+separators {
		t.WriteString(c[scope])
	} else {
		for sc := Para; sc >= Char; sc-- {
			if sc == scope {
				t.WriteString(output.String(c[sc]).Bold().String())
			} else {
				t.WriteString(c[sc])
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

// Get the monospace display width of the next breakable segment in source
func nextSegWidth(source []byte) (width int) {
	var f int    // Unicode boundary flags
	var g []byte // Grapheme cluster
	state := -1
	for len(source) > 0 && f&uniseg.MaskLine == 0 {
		g, source, f, state = uniseg.Step(source, state)
		width += f >> uniseg.ShiftWidth
	}

	// Don't count any trailing whitespace character
	r, _ := utf8.DecodeRune(g)
	if unicode.Is(unicode.Z, r) {
		width -= f >> uniseg.ShiftWidth
	}

	return width
}

/*
Draw one line in the edit window.  Word wraps at the end of the line.

Takes the current paragraph number, character count, document source text and
uniseg state and returns the text of the line.  Consumes text from the document
source and updates the character count and uniseg state.
*/
func drawLine(pn int, c *int, source *[]byte, state *int) string {
	var f int            // Unicode boundary flags
	m := ex - margin - 1 // Right margin
	var r rune
	var t strings.Builder
	var w int // Monospace width of character
	var x int // Column position in the line
	for {
		// Break loop at margin or mandatory break
		f &= uniseg.MaskLine
		if len(*source) > 0 &&
			(x > m || f == uniseg.LineMustBreak || (f == uniseg.LineCanBreak && x+nextSegWidth(*source) > m)) {
			break
		}

		if pn == cursor[Para] && *c == cursor[Char] && (x == 0 || w > 0) {
			t.WriteString(cursorString())
			updateCursorPos()
			m++
			x++
		}

		if len(*source) == 0 {
			break
		}

		var g []byte // Grapheme cluster
		g, *source, f, *state = uniseg.Step(*source, *state)
		r, _ = utf8.DecodeRune(g)
		if r == utf8.RuneError {
			continue
		}

		if f&uniseg.MaskWord != 0 && isAlphanumeric(*source) {
			indexWord(pn, *c)
		}

		if f&uniseg.MaskSentence != 0 && len(*source) > 0 {
			indexSent(pn, *c)
		}

		w = f >> uniseg.ShiftWidth
		if w > 0 {
			*c++
			t.Write(g)
			x += w
		}
	}

	if x > m && f == 0 && !unicode.Is(unicode.Z, r) { // Not a space character
		t.WriteString(strings.Repeat(" ", ex-x-1))
		t.WriteString(output.String("-").Reverse().String())
	}

	return t.String()
}

// Draw one paragraph in the edit window
func drawPara(pn int) (text []string) {
	// Reset sentence, word and character indexes for this paragraph
	p := &paras[pn-1]
	total[Char] -= p.chars
	total[Word] -= len(p.cword)
	total[Sent] -= len(p.csent)
	*p = ipara{}

	if pn == cursor[Para] {
		curs_line = -1
	}

	source := []byte(doc.GetText(pn))
	if len(source) > 0 {
		indexSent(pn, 0)
	}
	if isAlphanumeric(source) {
		indexWord(pn, 0)
	}

	var c int
	state := -1
	for {
		text = append(text, drawLine(pn, &c, &source, &state))
		if curs_line == -1 && (c > cursor[Char] || len(source) == 0) {
			curs_line = len(text) - 1
		}
		if len(source) == 0 {
			break
		}
	}

	text = append(text, "")

	// Update character counts
	p.chars = c
	total[Char] += c

	return text
}

/*
Draw the edit window.

It word wraps, renders and caches the paragraph that the cursor is in.
If the cursor is not within the currently displayed paragraphs, it moves
the starting position to bring the cursor back in view.  The navigation
indexes are updated during this process.
*/
func drawWindow() {
	pn := cursor[Para]

	if len(buffer) > 0 {
		// Erase the old cursor position
		buf := &buffer[curs_buff]
		if cursor[Para] != buf.pn {
			buf.text = drawPara(buf.pn)
		}

		// Optimise for common cases:
		lastbuf := buffer[len(buffer)-1]
		if pn == lastbuf.pn+1 {
			// The cursor is one paragraph below, draw it
			buffer = append(buffer, para{pn, drawPara(pn)})
			curs_buff = len(buffer) - 1
			return
		}

		if pn+1 == buffer[0].pn {
			// The cursor is one paragraph above, draw it
			buffer = slices.Insert[[]para](buffer, 0, para{pn, drawPara(pn)})
			curs_buff, first_buff, first_line = 0, 0, curs_line
			return
		}

		// TODO Draw a newly inserted paragraph

		for i := range buffer {
			if buffer[i].pn == pn {
				// Redraw current paragraph
				buffer[i] = para{pn, drawPara(pn)}
				curs_buff = i
				if curs_buff < first_buff || curs_line < first_line {
					first_buff, first_line = curs_buff, curs_line // Scroll backwards
				}
				return
			}
		}

		// The cursor is outside the screen: redraw everything
		buffer = nil
	}

	curs_buff, first_buff, first_line = 0, 0, 0
	var rows int // Number of rows drawn
	for {
		text := drawPara(pn)
		buffer = append(buffer, para{pn, text})
		rows += len(text)
		if rows >= ey || pn >= doc.Paragraphs() {
			break
		}

		pn++
	}
}

func ResizeScreen(x, y int) {
	buffer = nil
	ex, ey = x, y-1
	first_buff, first_line = 0, 0
}

// The entire screen including the edits window and status line
func Screen() string {
	drawWindow()
	b, l := first_buff, first_line
	var t []string
	for i := 0; i < ey && b < len(buffer); i++ {
		t = append(t, buffer[b].text[l])
		l++
		if l >= len(buffer[b].text) {
			b++
			l = 0
		}
	}

	// Scroll forwards until the cursor has been drawn
	for b < len(buffer) {
		buf := buffer[b]
		if buf.pn > cursor[Para] || (buf.pn == cursor[Para] && l > curs_line) {
			break
		}

		// Scroll forwards
		t = slices.Delete(t, 0, 1)
		first_line++
		if first_line >= len(buffer[first_buff].text) {
			first_buff++
			first_line = 0
		}

		t = append(t, buffer[b].text[l])
		l++
		if l >= len(buffer[b].text) {
			b++
			l = 0
		}
	}

	for len(t) < ey {
		t = append(t, "")
	}
	t = append(t, statusLine())
	return strings.Join(t, "\n")
}
