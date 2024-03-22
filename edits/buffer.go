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
var counterChar = [...]rune{'@', '#', '$', '¶', '§'}
var cursorChar = [...]rune{'_', '#', '$', '¶', '§'}
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
	Sent  // Sentence
	Para  // Paragraph
	Sectn // Section
	MaxScope
)

type counts [MaxScope]int

// Information about a single paragraph in the terminal window
type para struct {
	sn, pn int      // Section and paragraph number in the document
	text   []string // Rendered lines including cursor and marks
}

/*
The "buffer" variable caches the paragraphs displayed in the terminal in order
to speed up rendering by avoiding constantly redrawing the entire window.

The "cursor" variable contains the current cursor position for navigation
purposes, with the character, word and sentence values relative to the current
paragraph.
*/

var buffer []para
var cursor = counts{Sectn: 1, Para: 1} // Current cursor position
var initialCap = true                  // Initial capital at the start of a sentence
var scope Scope

// The current cursor position within the section
func cursorPos() (c counts) {
	c = cursor
	s := sections[cursor[Sectn]-1]
	for i := 0; i < c[Para]-1; i++ {
		p := s.p[i]
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
	var c [MaxScope]string // Counters for each scope
	var w int              // Width of counters

	current := cursorPos()
	s := &sections[cursor[Sectn]-1]
	total := counts{s.chars, s.words, s.sents, len(s.p), len(sections)}
	for sc := Char; sc <= Sectn; sc++ {
		c[sc] = string(counterChar[sc]) + strconv.Itoa(current[sc]) + "/" + strconv.Itoa(total[sc])
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

// Draw one line in the edit window.  Word wraps at the end of the line.
func drawLine(sn, pn int, c *int, source *[]byte, state *int) string {
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

		if sn == cursor[Sectn] && pn == cursor[Para] && *c == cursor[Char] && (x == 0 || w > 0) {
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

		isAN := isAlphanumeric(*source)
		if f&uniseg.MaskWord != 0 && isAN {
			indexWord(sn, pn, *c)
		}

		if f&uniseg.MaskSentence != 0 && len(*source) > 0 {
			indexSent(sn, pn, *c)
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
func drawPara(sn, pn int) (text []string) {
	// Reset sentence, word and character counts for this paragraph
	s := &sections[sn-1]
	p := &s.p[pn-1]
	s.sents -= len(p.csent)
	s.words -= len(p.cword)
	s.chars -= p.chars
	*p = ipara{}

	if sn == cursor[Sectn] && pn == cursor[Para] {
		curs_line = -1
	}

	source := []byte(doc.GetText(sn, pn))
	if len(source) > 0 {
		indexSent(sn, pn, 0)
	}
	if isAlphanumeric(source) {
		indexWord(sn, pn, 0)
	}

	var c int
	state := -1
	for {
		text = append(text, drawLine(sn, pn, &c, &source, &state))
		if curs_line == -1 && (c > cursor[Char] || len(source) == 0) {
			curs_line = len(text) - 1
		}
		if len(source) == 0 {
			break
		}
	}

	if sn < doc.Sections() && pn == doc.Paragraphs(sn) {
		text = append(text, strings.Repeat("─", ex))
	} else {
		text = append(text, "")
	}

	// Update character counts
	s.chars += c
	p.chars = c

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
	sn := cursor[Sectn]

	if len(buffer) > 0 {
		// Erase the old cursor position
		buf := &buffer[curs_buff]
		if cursor[Sectn] != buf.sn || cursor[Para] != buf.pn {
			buf.text = drawPara(buf.sn, buf.pn)
		}

		// Optimise for common cases:
		lastbuf := buffer[len(buffer)-1]
		if (sn == lastbuf.sn && pn == lastbuf.pn+1) ||
			(lastbuf.pn == doc.Paragraphs(lastbuf.sn) && sn == lastbuf.sn+1) {
			// The cursor is one paragraph below, draw it
			buffer = append(buffer, para{sn, pn, drawPara(sn, pn)})
			curs_buff = len(buffer) - 1
			return
		}

		if (sn == buffer[0].sn && pn+1 == buffer[0].pn) ||
			(pn == doc.Paragraphs(sn) && sn+1 == buffer[0].sn) {
			// The cursor is one paragraph above, draw it
			buffer = slices.Insert[[]para](buffer, 0, para{sn, pn, drawPara(sn, pn)})
			curs_buff, first_buff, first_line = 0, 0, curs_line
			return
		}

		for i := range buffer {
			if buffer[i].sn == sn && buffer[i].pn == pn {
				// Redraw current paragraph
				buffer[i] = para{sn, pn, drawPara(sn, pn)}
				curs_buff = i
				if curs_line < first_line {
					first_buff, first_line = curs_buff, curs_line // Scroll backwards
				}
				return
			}
		}
	}

	// The cursor is outside the screen: redraw everything
	buffer = nil
	curs_buff, first_buff, first_line = 0, 0, 0
	var rows int // Number of rows drawn
	for {
		text := drawPara(sn, pn)
		buffer = append(buffer, para{sn, pn, text})
		rows += len(text)
		if rows >= ey {
			break
		}

		switch {
		case pn < doc.Paragraphs(sn):
			pn++
		case sn < doc.Sections():
			pn = 1
			sn++
		default:
			return
		}
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
		if buf.sn > cursor[Sectn] ||
			(buf.sn == cursor[Sectn] && buf.pn > cursor[Para]) ||
			(buf.sn == cursor[Sectn] && buf.pn == cursor[Para] && l > curs_line) {
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
