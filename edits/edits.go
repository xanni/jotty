package edits

import (
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	ps "git.sericyb.com.au/jotty/permascroll"
	"github.com/muesli/termenv"
	"github.com/rivo/uniseg"
)

/*
Implements the visible user interface elements of the edits window and status
line.

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

var (
	ID          string // The program name and version
	counterChar = [...]rune{'@', '#', '$', '¶'}
	cursorChar  = [...]rune{'_', '#', '$', '¶'}
)

const (
	cursorCharCap = '↑'       // Capitalisation indicator character
	margin        = 5         // Up to 3 edit marks, cursor and wrap indicator
	markChar      = "|"       // Visual representation of an edit mark
	markColor     = "#ffff00" // ANSIBrightYellow
)

type Scope int

const (
	Char Scope = iota // Character
	Word              // Word
	Sent              // Sentence
	Para              // Paragraph
	MaxScope
)

type counts [MaxScope]int

// Information about a single paragraph in the terminal window.
type para struct {
	chars int      // Total number of characters in the paragraph
	cword []int    // Character index of each word in the paragraph
	csent []int    // Character index of each sentence in the paragraph
	text  []string // Rendered lines including cursor and marks
}

/*
The "cache" variable caches the paragraphs displayed in the terminal in order
to avoid constantly redrawing the entire window.

The "cursor" variable contains the current cursor position for navigation
purposes.
*/

var (
	after, before        strings.Builder // Text of the current paragraph after and before the cursor position
	cache                []para
	cursor               = counts{Para: 1} // Current cursor position
	cursPara, cursLine   int               // Paragraph and line containing the cursor
	ex, ey               int               // Edit window dimensions
	firstPara, firstLine int               // Paragraph and line at top of edit window
	initialCap           = true            // Initial capital at the start of a sentence
	markPara             int               // Paragraph containing the mark(s), if any
	mark                 []int             // Character positions of the marks
	output               = termenv.NewOutput(os.Stdout)
	scope                Scope
)

// Add a word to the index if not already present.
func indexWord(pn, c int) {
	p := &cache[pn-1]
	if len(p.cword) == 0 || c > p.cword[len(p.cword)-1] {
		p.cword = append(p.cword, c)
		total[Word]++
	}
}

// Add a sentence to the index if not already present.
func indexSent(pn, c int) {
	p := &cache[pn-1]
	if len(p.csent) == 0 || c > p.csent[len(p.csent)-1] {
		p.csent = append(p.csent, c)
		total[Sent]++
	}
}

// The current cursor position within the whole document, not just the paragraph.
func cursorPos() (c counts) {
	c = cursor
	for i := range c[Para] - 1 {
		p := cache[i]
		c[Sent] += len(p.csent)
		c[Word] += len(p.cword)
		c[Char] += p.chars
	}

	return c
}

// The current cursor character and terminal attributes.
func cursorString() string {
	cc := cursorCharCap
	if !initialCap {
		cc = cursorChar[scope]
	}

	return output.String(string(cc)).Reverse().Blink().String()
}

func markString() string {
	return output.String(markChar).Blink().Foreground(output.Color(markColor)).String()
}

// Draw the status bar that appears on the last line of the screen.
func statusLine() string {
	const separators = 4   // One space after each scope
	var c [MaxScope]string // Counters for each scope
	var w int              // Total width of counters in character cells

	current := cursorPos()
	total[Para] = ps.Paragraphs()
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

// True if the first rune in source is a Unicode letter or number.
func isAlphanumeric(source []byte) bool {
	r, _ := utf8.DecodeRune(source)

	return unicode.In(r, unicode.L, unicode.N)
}

// Get the monospace display width of the next breakable segment in source.
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

// Helper function.
func updateBeforeAndAfter(c int, g []byte) {
	if c < cursor[Char] {
		before.Write(g)
	} else {
		after.Write(g)
	}
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

		if x == 0 || w > 0 {
			if pn == markPara {
				for _, mc := range mark {
					if *c == mc {
						t.WriteString(markString())
						m++
						x++
					}
				}
			}

			if pn == cursor[Para] && *c == cursor[Char] {
				t.WriteString(cursorString())
				updateCursorPos()
				m++
				x++
			}
		}

		if len(*source) == 0 {
			break
		}

		var g []byte // Grapheme cluster
		g, *source, f, *state = uniseg.Step(*source, *state)
		r, _ = utf8.DecodeRune(g)

		if pn == cursor[Para] {
			updateBeforeAndAfter(*c, g)
		}

		w = f >> uniseg.ShiftWidth
		if w > 0 {
			*c++
			t.Write(g)
			x += w
		}

		if f&uniseg.MaskWord != 0 && isAlphanumeric(*source) {
			indexWord(pn, *c)
		}

		if f&uniseg.MaskSentence != 0 && len(*source) > 0 {
			indexSent(pn, *c)
		}
	}

	if x > m && f == 0 && !unicode.Is(unicode.Z, r) { // Not a space character
		t.WriteString(strings.Repeat(" ", ex-x-1))
		t.WriteString(output.String("-").Reverse().String())
	}

	return t.String()
}

// Reset character, word and sentence indexes for a paragraph.
func clearCache(pn int) {
	p := &cache[pn-1]
	total[Char] -= p.chars
	total[Word] -= len(p.cword)
	total[Sent] -= len(p.csent)
	*p = para{}
}

// Draw one paragraph in the edit window.
func drawPara(pn int) {
	if pn <= len(cache) {
		clearCache(pn)
	} else {
		cache = append(cache, para{})
	}

	if pn == cursor[Para] {
		before.Reset()
		after.Reset()
		cursLine = -1
	}

	source := []byte(ps.GetText(pn))
	if len(source) > 0 {
		indexSent(pn, 0)
	}
	if isAlphanumeric(source) {
		indexWord(pn, 0)
	}

	var c int
	state := -1
	p := &cache[pn-1]
	for {
		p.text = append(p.text, drawLine(pn, &c, &source, &state))
		if cursLine == -1 && (c > cursor[Char] || len(source) == 0) {
			cursLine = len(p.text) - 1
		}
		if len(source) == 0 {
			break
		}
	}

	// Update character counts
	p.chars = c
	total[Char] += c
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

	if len(cache) > 0 {
		if pn != cursPara {
			drawPara(cursPara) // Erase the old cursor position
		}

		// Optimise for common cases:
		if pn <= len(cache)+1 {
			drawPara(pn) // Redraw current paragraph
			cursPara = pn
			if cursPara < firstPara || cursLine < firstLine {
				firstPara, firstLine = cursPara, cursLine // Scroll backwards
			}

			return
		}
	}

	// Redraw everything
	cache = make([]para, pn-1, pn)
	cursPara = pn
	if firstPara < 1 || firstPara > cursPara { // Ensure cursor is onscreen
		firstPara, firstLine = max(pn-1, 1), 0 // Provide one preceding paragraph of context if possible
	}

	pn = firstPara
	for rows := 0; (pn <= cursor[Para] || rows < ey) && pn <= ps.Paragraphs(); pn++ {
		drawPara(pn)
		rows += len(cache[pn-1].text) + 1
	}
}

func ResizeScreen(x, y int) {
	cache = nil
	total = counts{0, 0, 0, 1}

	if x != ex {
		firstLine = 0
	}

	ex, ey = x, y-1
}

// Render one line to the screen.
func screenLine(pn, ln *int) (line string) {
	if *ln < len(cache[*pn-1].text) {
		line = cache[*pn-1].text[*ln]
		*ln++
	} else {
		*ln = 0
		*pn++
		line = ""
	}

	return line
}

// The entire screen including the edits window and status line.
func Screen() string {
	drawWindow()
	pn, ln := firstPara, firstLine
	var t []string

	for i := 0; i < ey && pn <= len(cache); i++ {
		t = append(t, screenLine(&pn, &ln))
	}

	// Scroll forwards until the cursor has been drawn
	for pn <= len(cache) {
		if pn > cursor[Para] || (pn == cursor[Para] && ln > cursLine) {
			break
		}

		// Scroll forwards
		t = slices.Delete(t, 0, 1)
		firstLine++
		if firstLine >= len(cache[firstPara].text) {
			firstPara++
			firstLine = 0
		}

		t = append(t, screenLine(&pn, &ln))
	}

	for len(t) < ey {
		t = append(t, "")
	}
	t = append(t, statusLine())

	return strings.Join(t, "\n")
}
