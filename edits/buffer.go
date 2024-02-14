package edits

/*
Implements the buffer that represents the visible user interface elements
of the edits window and status line.
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

const margin = 5 // Up to 3 edit marks, cursor and wrap indicator

var ID string  // The program name and version
var Sx, Sy int // Screen dimensions

type Scope int

const (
	Char Scope = iota
	Word
	Sent  // Sentence
	Para  // Paragraph
	Sectn // Section
	MaxScope
)

var scope Scope

type counts [MaxScope]int

type line struct {
	bytes, chars int // cumulative counts at start of line
	sectn        int // current section at start of line
}

var bufc, bufy int // last character position and last row in the buffer
var buffer []line
var cursor = counts{Sectn: 1} // current position in the section/document
var cursx, cursy int          // current position in the edit window
var document []byte
var initialCap = true // initial capital at the start of a sentence
var total counts
var win *nc.Window

func appendParaBreak() {
	i := len(document) - 1
	if document[i] != ' ' {
		AppendRune([]byte{'\n'})
	} else {
		document[i] = '\n'
		DrawWindow()
	}
	scope = Para
}

func appendSectnBreak() {
	s := cursor[Sectn] + 1
	cursor = counts{Sectn: s}

	i := len(document) - 1
	if document[i] != '\n' {
		document = append(document, '\f')
		i++
	} else {
		document[i] = '\f'
	}
	i++
	indexSectn(i)
	newSection(s)

	if len(buffer) > 0 {
		if cursy >= Sy-2 {
			scrollUp(2)
		}
		if cursy > 0 {
			win.HLine(cursy-1, 0, nc.ACS_HLINE, Sx-1)
		}
		buffer[cursy] = line{bytes: i, sectn: s}
	}

	scope = Sectn
	DrawWindow()
}

// Find which screen row contains the character position of the cursor
func cursorRow() (y int) {
	for y = bufy; y > 0; y-- {
		if buffer[y].sectn == cursor[Sectn] && buffer[y].chars <= cursor[Char] {
			break
		}
	}

	return y
}

// Draw the cursor
// TODO erase the old cursor when moving the cursor to a new position
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
	cur_s, cur_c := cursor[Sectn], cursor[Char]

	if len(buffer) == 0 ||
		cur_s < buffer[0].sectn ||
		(cur_s == buffer[0].sectn && cur_c < buffer[0].chars) ||
		cur_s > buffer[bufy].sectn ||
		(scope >= Sent && cur_s == buffer[bufy].sectn && cur_c > bufc+1) {
		return false
	}

	if scope < Sent && cursy == Sy-1 && cur_c > bufc+1 {
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
	bufy = 0
	cursx = 0
	cursy = 0
	p := getPara()
	if p > len(ipara)-1 || !isNewParagraph(cursor[Char]) {
		p--
	}
	buffer[0] = line{bytes: ipara[p].b, chars: ipara[p].c, sectn: cursor[Sectn]}
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
		cursy = 0
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

	if len(buffer) == 0 {
		cursor[Char] = uniseg.GraphemeClusterCount(string(document))
	} else {
		cursor[Char] = buffer[cursy].chars + uniseg.GraphemeClusterCount(string(document[buffer[cursy].bytes:]))
		DrawWindow()
	}
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

	var l line   // current line with cumulative counts
	var x, y int // current screen coordinates

	// First find the character the cursor is located at on the screen, if possible
	if isCursorInBuffer() {
		y = cursorRow()
	} else {
		// Nothing has been drawn yet, or the cursor is outside the screen: redraw everything
		newBuffer()
	}

	l = buffer[y]
	source := document[l.bytes:]
	state := -1

	// A new paragraph is often the start of a new word that might not have been recorded yet
	if isNewParagraph(l.chars) && isAlphanumeric(source) {
		indexWord(l.chars)
	}

	for y < Sy-1 {
		if l.sectn == cursor[Sectn] && l.chars == cursor[Char] {
			cursx = x
			cursy = y
			updateCursorPos()
			x++
		}

		if len(source) == 0 {
			break
		}

		var f int    // Unicode boundary flags
		var g []byte // grapheme cluster

		g, source, f, state = uniseg.Step(source, state)
		l.bytes += len(g)
		r, _ := utf8.DecodeRune(g)
		if r == utf8.RuneError {
			continue
		}

		w := f >> uniseg.ShiftWidth // monospace width of character
		if w > 0 || r == '\n' {
			l.chars++
		}

		isAN := isAlphanumeric(source)
		if f&uniseg.MaskWord != 0 && isAN {
			indexWord(l.chars)
		}

		if r == '\n' || (f&uniseg.MaskSentence != 0 && len(source) > 0) {
			indexSent(l.bytes, l.chars)
		}

		if r == '\n' {
			indexPara(l.bytes, l.chars)
		}

		if r == '\f' {
			indexSectn(l.bytes)
			l.chars = 0
			l.sectn++
			newSection(l.sectn)

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
		if x < Sx-1 &&
			(f != uniseg.LineCanBreak || x+w+nextSegWidth(source) < Sx-1) &&
			(f != uniseg.LineMustBreak || (len(source) == 0 && !uniseg.HasTrailingLineBreak(document))) {
			continue
		}

		if x >= Sx-1 && f != uniseg.LineCanBreak {
			win.MoveAddChar(y, x, '-'|nc.A_REVERSE)
		} else {
			win.Move(y, x)
			win.ClearToEOL()
		}

		x = 0
		y++

		if y < Sy-1 {
			if r == '\n' {
				win.Move(y, x)
				win.ClearToEOL()
			}
			if r == '\f' {
				win.HLine(y, 0, nc.ACS_HLINE, Sx-1)
			}
		}

		if r == '\n' || r == '\f' {
			y++
		}

		// At the last line of the window but haven't passed the cursor yet
		if y >= Sy-1 && l.sectn <= cursor[Sectn] && l.chars <= cursor[Char] {
			lines := (y + 2) - Sy
			scrollUp(lines)
			y -= lines
		}

		if y < Sy-1 {
			buffer[y] = l
		}
	}

	bufc = l.chars

	if y >= Sy-1 {
		bufy = Sy - 2
	} else {
		win.Move(y, x)
		win.ClearToBottom()
		if y > bufy {
			bufy = y
		}
	}

	if l.sectn == cursor[Sectn] {
		total = counts{l.chars, len(iword), len(isent), len(ipara), len(isectn)}
	} else {
		scanSectn()
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
