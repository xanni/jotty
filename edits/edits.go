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
	cursorCharCap  = '↑'       // Capitalisation indicator character
	margin         = 6         // Up to 4 edit marks, cursor and wrap indicator
	markChar       = '|'       // Visual representation of an edit mark
	moreChar       = '…'       // Continuation indicator character
	cutColor       = "#808080" // Cut text: ANSIBrightBlack
	markColor      = "#ffff00" // Edit mark: ANSIBrightYellow
	primaryColor   = "#ff0000" // Primary selection: ANSIBrightRed
	secondaryColor = "#ff00ff" // Secondary selection: ANSIBrightMagenta
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

type selection struct {
	cbegin, cend int // Character positions within the marked paragraph
	obegin, oend int // Byte offsets within the paragraph in the document
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
	primary, secondary   selection
	prevSelected         bool // Previous paragraph is currently selected
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
	return output.String(string(markChar)).Blink().Foreground(output.Color(markColor)).String()
}

func cutStyle(s string) string {
	return output.String(s).CrossOut().Foreground(output.Color(cutColor)).String()
}

func primaryStyle(s string) string {
	return output.String(s).Reverse().Foreground(output.Color(primaryColor)).String()
}

func secondaryStyle(s string) string {
	return output.String(s).Underline().Foreground(output.Color(secondaryColor)).String()
}

// Draw the status bar that appears on the last line of the screen.
func statusLine() string {
	const cutLabel = "   cut: "
	const minCut = 4       // Minimum amount of cut buffer to display
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

	cutBuffer := ps.GetCut()
	cutMax := ex - (t.Len() + len(cutLabel) + 2)
	cutLen := len(cutBuffer)
	if cutLen > 0 && cutMax >= minCut {
		t.WriteString(cutLabel)
		t.WriteString(cutStyle(cutBuffer[:min(cutLen, cutMax)]))
		if cutLen > cutMax {
			t.WriteRune(moreChar)
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

// Get the character position one scope unit backwards.
func preceding(end int) (begin int) {
	p := cache[cursor[Para]-1]

	switch scope {
	case Char:
		if end > 0 {
			begin = end - 1
		}
	case Word:
		i, _ := slices.BinarySearch[[]int](p.cword, end)
		if i > 0 {
			begin = p.cword[i-1]
		}
	case Sent:
		i, _ := slices.BinarySearch[[]int](p.csent, end)
		if i > 0 {
			begin = p.csent[i-1]
		}
	default: // Para
	}

	return begin
}

// Get the character position one scope unit forwards.
func following(begin int) (end int) {
	p := cache[cursor[Para]-1]
	end = p.chars

	switch scope {
	case Char:
		if begin < p.chars {
			end = begin + 1
		}
	case Word:
		i, found := slices.BinarySearch[[]int](p.cword, begin)
		if found {
			i++
		}
		if i < len(p.cword) {
			end = p.cword[i]
		}
	case Sent:
		i, found := slices.BinarySearch[[]int](p.csent, begin)
		if found {
			i++
		}
		if i < len(p.csent) {
			end = p.csent[i]
		}
	default: // Para
	}

	return end
}

// Set the selections based on the current edit marks and scope.
func updateSelections() {
	if len(mark) > 0 && cursor[Para] != markPara {
		mark = nil
	}

	selectPrevPara := scope == Para && markPara > 1 && len(mark) == 1 && mark[0] == 0
	if selectPrevPara != prevSelected {
		prevSelected = selectPrevPara
		drawPara(markPara - 1)
	}

	switch len(mark) {
	case 0:
		primary, secondary = selection{}, selection{}
	case 1:
		if cursor[Char] == mark[0] {
			primary, secondary = selection{cbegin: mark[0], cend: following(mark[0])},
				selection{cbegin: preceding(mark[0]), cend: mark[0]}
		} else {
			first := min(mark[0], cursor[Char])
			second := max(mark[0], cursor[Char])
			primary, secondary = selection{cbegin: first, cend: second}, selection{cbegin: second, cend: following(second)}
		}
	case 2:
		first := min(mark[0], mark[1])
		second := max(mark[0], mark[1])
		primary, secondary = selection{cbegin: first, cend: second}, selection{cbegin: second, cend: following(second)}
	default: // 3 or 4 marks
		sorted := make([]int, len(mark))
		copy(sorted, mark)
		slices.Sort(sorted)
		primary, secondary = selection{cbegin: sorted[0], cend: sorted[1]},
			selection{cbegin: sorted[len(sorted)-2], cend: sorted[len(sorted)-1]}
	}
}

type line struct {
	c      int             // Character count
	m      int             // Right margin
	pn     int             // Paragraph number
	source *[]byte         // Paragraph text being rendered
	state  int             // Unicode segmentation state
	t      strings.Builder // Text
	w      int             // Monospace width of current character
	x      int             // Current column position in the line
}

// Helper function.
func (l *line) updateBeforeAndAfter(g []byte) {
	if l.pn == markPara {
		offset := before.Len() + after.Len()
		switch l.c {
		case primary.cbegin:
			primary.obegin = offset
		case primary.cend:
			primary.oend = offset
		}
		switch l.c {
		case secondary.cbegin:
			secondary.obegin = offset
		case secondary.cend:
			secondary.oend = offset
		}
	}

	if l.c < cursor[Char] {
		before.Write(g)
	} else {
		after.Write(g)
	}
}

// Draw one character in the edit window with highlighting as required.
func (l *line) drawChar(g []byte) {
	isPrimary := l.pn == markPara && l.c >= primary.cbegin && l.c < primary.cend
	isSecondary := (l.pn == markPara && l.c >= secondary.cbegin && l.c < secondary.cend) ||
		(l.pn == markPara-1 && prevSelected)

	switch {
	case isPrimary:
		l.t.WriteString((primaryStyle(string(g))))
	case isSecondary:
		l.t.WriteString((secondaryStyle(string(g))))
	default:
		l.t.Write(g)
	}

	l.c++
	l.x += l.w
}

func (l *line) drawMarker(s string) {
	l.t.WriteString(s)
	l.m++
	l.x++
}

// Draw the cursor and any edit mark if at the current position.
func (l *line) drawAllMarkers() {
	if l.x == 0 || l.w > 0 {
		if l.pn == markPara {
			for _, mc := range mark {
				if l.c == mc {
					l.drawMarker(markString())
				}
			}
		}

		if l.pn == cursor[Para] && l.c == cursor[Char] {
			l.drawMarker(cursorString())
			updateCursorPos()
		}
	}
}

/*
Draw one line in the edit window.  Word wraps at the end of the line.

Returns the text of the line.  Consumes text from the document source and
updates the character count and uniseg state.
*/
func (l *line) drawLine() string {
	l.m, l.w, l.x = ex-margin-1, 0, 0
	l.t.Reset()
	var f int // Unicode boundary flags
	var r rune
	for {
		// Break loop at margin or mandatory break
		f &= uniseg.MaskLine
		if len(*l.source) > 0 &&
			(l.x > l.m || f == uniseg.LineMustBreak || (f == uniseg.LineCanBreak && l.x+nextSegWidth(*l.source) > l.m)) {
			break
		}

		l.drawAllMarkers()

		if len(*l.source) == 0 {
			break
		}

		var g []byte // Grapheme cluster
		g, *l.source, f, l.state = uniseg.Step(*l.source, l.state)
		r, _ = utf8.DecodeRune(g)

		if l.pn == cursor[Para] {
			l.updateBeforeAndAfter(g)
		}

		l.w = f >> uniseg.ShiftWidth
		if l.w > 0 {
			l.drawChar(g)
		}

		if f&uniseg.MaskWord != 0 && isAlphanumeric(*l.source) {
			indexWord(l.pn, l.c)
		}

		if f&uniseg.MaskSentence != 0 && len(*l.source) > 0 {
			indexSent(l.pn, l.c)
		}
	}

	if l.x > l.m && f == 0 && !unicode.Is(unicode.Z, r) { // Not a space character
		l.t.WriteString(strings.Repeat(" ", ex-l.x-1))
		l.t.WriteString(output.String("-").Reverse().String())
	}

	return l.t.String()
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

	l := line{pn: pn, source: &source, state: -1}
	p := &cache[pn-1]
	for {
		p.text = append(p.text, l.drawLine())
		if cursLine == -1 && (l.c > cursor[Char] || len(source) == 0) {
			cursLine = len(p.text) - 1
		}
		if len(source) == 0 {
			break
		}
	}

	// Update character counts
	p.chars = l.c
	total[Char] += l.c

	// Update selections
	if pn == markPara {
		offset := before.Len() + after.Len()
		if primary.cend == l.c {
			primary.oend = offset
		} else if secondary.cend == l.c {
			secondary.oend = offset
		}
	}
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
