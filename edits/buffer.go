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
	Sent // Sentence
	Para // Paragraph
	Sect // Section
	MaxScope
)

var scope Scope

type counts [MaxScope]int

type line struct {
	bytes, chars int // cumulative counts at start of line
	sect         int // current section at start of line
	text         []byte
}

var bufc, bufy int // last character position and last row in the buffer
var buffer []line
var cursor struct {
	pos  counts // current position in the section/document
	x, y int    // current position in the edit window
}
var document []byte
var initialCap = true // initial capital at the start of a sentence
var total counts
var win *nc.Window

func init() {
	cursor.pos[Sect] = 1
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
	osect = 0
	scope = Char

	if len(buffer) == 0 {
		cursor.pos[Char] = uniseg.GraphemeClusterCount(string(document))
	} else {
		cursor.pos[Char] = buffer[cursor.y].chars + uniseg.GraphemeClusterCount(string(document[buffer[cursor.y].bytes:]))
		DrawWindow()
	}
}

func DecScope() {
	if scope == Char {
		scope = Sect
	} else {
		scope--
	}

	if scope < Sent {
		initialCap = false
	}

	DrawCursor()
	DrawStatusBar()
}

func IncScope() {
	if scope == Sect {
		scope = Char
		initialCap = false
	} else {
		scope++
	}

	DrawCursor()
	DrawStatusBar()
}

// Draw the cursor
// TODO erase the old cursor when moving the cursor to a new position
func DrawCursor() {
	cc := '↑'
	if !initialCap {
		cc = cursorChar[scope]
	}

	win.AttrSet(nc.Char(cursorStyle))
	win.MovePrint(cursor.y, cursor.x, string(cc))
	win.AttrSet(nc.A_NORMAL)
	win.NoutRefresh()
}

// Draw a status bar on the last line of the screen
func DrawStatusBar() {
	win.Move(Sy-1, 0)
	win.ClearToBottom()

	var c [MaxScope]string // counters for each scope
	var w int              // width of counters
	for s := Char; s <= Sect; s++ {
		c[s] = string(counterChar[s]) + strconv.Itoa(cursor.pos[s]) + "/" + strconv.Itoa(total[s])
		w += utf8.RuneCountInString(c[s])
	}

	win.Move(Sy-1, 0)
	if Sx >= len(ID)+w+6 {
		win.Print(ID + "  ")
	}

	if Sx < w+6 {
		win.Print(c[scope])
	} else {
		for s := Sect; s >= Char; s-- {
			if s != scope {
				win.Print(c[s])
			} else {
				win.AttrOn(nc.A_BOLD)
				win.Print(c[s])
				win.AttrOff(nc.A_BOLD)
			}
			if s == Sect {
				win.AddChar(':')
			}
			win.AddChar(' ')
		}
	}

	win.Move(cursor.y, cursor.x)
	win.NoutRefresh()
}

// Find which screen row contains the character position of the cursor
func cursorRow() (y int) {
	for y = bufy; y > 0; y-- {
		if buffer[y].sect == cursor.pos[Sect] && buffer[y].chars <= cursor.pos[Char] {
			break
		}
	}

	return y
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
	cur_s, cur_c := cursor.pos[Sect], cursor.pos[Char]

	if len(buffer) == 0 ||
		cur_s < buffer[0].sect ||
		(cur_s == buffer[0].sect && cur_c < buffer[0].chars) ||
		cur_s > buffer[bufy].sect ||
		(scope >= Sent && cur_s == buffer[bufy].sect && cur_c > bufc+1) {
		return false
	}

	if scope < Sent && cursor.y == Sy-1 && cur_c > bufc+1 {
		scrollUp(1)
	}

	return true
}

// Check if the cursor is at the start of a new paragraph
func isNewParagraph(c int) bool {
	if c == 0 {
		return true
	}

	p := cursor.pos[Para]
	return p < len(ipara) && c == ipara[p].c
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
	cursor.y -= lines
	if cursor.y < 0 {
		cursor.y = 0
	}
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
		buffer[y].text = nil
	} else {
		// Nothing has been drawn yet, or the cursor is outside the screen: redraw everything
		buffer = make([]line, Sy-1)
		bufy = 0
		cursor.x = 0
		cursor.y = 0
		buffer[0] = line{sect: cursor.pos[Sect]}
		p := getPara()
		if p > len(ipara)-1 || !isNewParagraph(cursor.pos[Char]) {
			p--
		}
		buffer[0].bytes = ipara[p].b
		buffer[0].chars = ipara[p].c
	}

	l = buffer[y]
	source := document[l.bytes:]
	state := -1

	// A new paragraph is often the start of a new word that might not have been recorded yet
	if isNewParagraph(l.chars) && isAlphanumeric(source) {
		indexWord(l.chars)
	}

	for y < Sy-1 {
		if l.sect == cursor.pos[Sect] && l.chars == cursor.pos[Char] {
			cursor.x = x
			cursor.y = y
			updateCursorPos()
			DrawCursor()
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
			indexSect(l.bytes)
			l.chars = 0
			l.sect++
			newSection(l.sect)

			if isAN {
				iword = []int{0}
			}
		}

		if w > 0 {
			l.text = append(l.text, g...)
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

		buffer[y].text = l.text
		l.text = nil
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
		if y >= Sy-1 && l.sect <= cursor.pos[Sect] && l.chars <= cursor.pos[Char] {
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
		buffer[y].text = l.text
		win.Move(y, x)
		win.ClearToBottom()
		if y > bufy {
			bufy = y
		}
	}

	if l.sect == cursor.pos[Sect] {
		total = counts{l.chars, len(iword), len(isent), len(ipara), len(isect)}
	} else {
		scanSect()
	}

	DrawStatusBar()
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

func appendSectBreak() {
	s := cursor.pos[Sect] + 1
	cursor.pos = counts{0, 0, 0, 0, s}

	i := len(document) - 1
	if document[i] != '\n' {
		document = append(document, '\f')
		i++
	} else {
		document[i] = '\f'
	}
	i++
	indexSect(i)
	newSection(s)

	if len(buffer) > 0 {
		if cursor.y >= Sy-2 {
			scrollUp(2)
		}
		if cursor.y > 0 {
			win.HLine(cursor.y-1, 0, nc.ACS_HLINE, Sx-1)
		}
		buffer[cursor.y] = line{bytes: i, sect: s}
	}

	scope = Sect
	DrawWindow()
}

func Space() {
	i := len(document) - 1
	if scope == Sect || i < 0 {
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
	default: // Para because Sect has already been excluded above
		appendSectBreak()
	}

	initialCap = scope >= Sent
	osect = 0
	DrawCursor()
	DrawStatusBar()
}

func Enter() {
	if scope == Sect || len(document) == 0 {
		return
	}

	if scope <= Sent {
		appendParaBreak()
	} else { // scope == Para
		appendSectBreak()
	}

	initialCap = true
	osect = 0
	DrawCursor()
	DrawStatusBar()
}
