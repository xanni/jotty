package edits

/*
Implements the buffer that represents the visible user interface elements
of the edits window and status line.
*/

import (
	"sort"
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

// ID is the program name and version.
var ID string
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
	bytes, chars, words int // cumulative counts at start of line
	text                []byte
}

var buffer []line
var cursor struct {
	pos  counts // current position in the section/document
	x, y int    // current position in the edit window
}
var document []byte
var initialCap = true // initial capital at the start of a sentence
var para = []int{0}   // byte index of each paragraph in the section
var sent = []int{0}   // byte index of each sentence in the section
var sect = []int{0}   // byte index of each section in the document
var total counts
var win *nc.Window

// AppendRune appends a UTF-8 encoded rune to the document.
func AppendRune(rb []byte) {
	r, _ := utf8.DecodeRune(rb)
	if initialCap && unicode.IsLower(r) {
		rb = []byte(string(unicode.ToUpper(r)))
	}

	document = append(document, rb...)
	initialCap = false
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

// DrawCursor draws the cursor.
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

// DrawStatusBar draws a status bar on the last line of the screen.
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
DrawWindow draws the edit window.

It word wraps, buffers and displays a portion of the document starting from
the line the cursor is on and ending at the last line of the edit window
or the end of the document, whichever comes first.
*/
func DrawWindow() {
	if Sx <= margin || Sy <= 1 {
		return
	}

	var l line   // current line with cumulative counts
	var x, y int // current screen coordinates

	if len(buffer) == 0 { // redraw everything
		buffer = make([]line, Sy-1)
	} else {
		y = cursor.y
		l = line{bytes: buffer[y].bytes, chars: buffer[y].chars, words: buffer[y].words}
	}
	source := document[l.bytes:]
	state := -1

	for {
		if cursor.pos[Char] == l.chars {
			cursor.pos[Word] = l.words
			cursor.pos[Sent] = sort.Search(len(sent), func(i int) bool { return sent[i] >= l.bytes })
			cursor.pos[Para] = sort.Search(len(para), func(i int) bool { return para[i] >= l.bytes })
			cursor.x = x
			cursor.y = y
			DrawCursor()
			x++
		}

		if len(source) == 0 {
			buffer[y].text = l.text
			break
		}

		var c []byte // grapheme cluster
		var f int    // Unicode boundary flags

		c, source, f, state = uniseg.Step(source, state)
		r, _ := utf8.DecodeRune(c)
		if r == utf8.RuneError {
			continue
		}

		w := f >> uniseg.ShiftWidth // monospace width of character
		l.bytes += len(c)
		if w > 0 || c[0] == '\n' {
			l.chars++
		}

		// Is the first rune in the grapheme cluster alphanumeric?
		if f&uniseg.MaskWord != 0 && unicode.In(r, unicode.L, unicode.N) {
			l.words++
		}

		if len(source) > 0 && f&uniseg.MaskSentence != 0 && l.bytes > sent[len(sent)-1] {
			sent = append(sent, l.bytes)
		}

		if c[0] == '\n' && l.bytes > para[len(para)-1] {
			para = append(para, l.bytes)
		}

		if c[0] == '\f' {
			if l.bytes > sect[len(sect)-1] {
				sect = append(sect, l.bytes)
			}
			l.chars = 0
			l.words = 0
			sent = []int{0}
			para = []int{0}
			cursor.pos = counts{0, 0, 0, 0, cursor.pos[Sect] + 1}
		}

		if w > 0 {
			l.text = append(l.text, c...)
			win.MovePrint(y, x, string(c))
			x += w
		}

		seg, _, _, _ := uniseg.FirstLineSegment(source, -1) // next breakable segment
		nw := uniseg.StringWidth(string(seg))               // width of next breakable segment

		// Break if at margin or mandatory break that is not just end of source
		f &= uniseg.MaskLine
		br := x >= Sx-1 ||
			(f == uniseg.LineCanBreak && x+w+nw >= Sx-1) ||
			(f == uniseg.LineMustBreak && (len(source) > 0 || uniseg.HasTrailingLineBreak(document)))

		if br {
			if x >= Sx-1 && f != uniseg.LineCanBreak {
				win.MoveAddChar(y, x, '-'|nc.A_REVERSE)
			} else {
				win.Move(y, x)
				win.ClearToEOL()
			}

			buffer[y].text = l.text
			x = 0
			y++
			if c[0] == '\f' && y < Sy-1 {
				win.HLine(y, 0, nc.ACS_HLINE, Sx-1)
			}
			if c[0] == '\n' || c[0] == '\f' {
				y++
			}
			if y >= Sy-1 { // last line of the window
				lines := (y + 2) - Sy
				scrollUp(lines)
				y -= lines
			}

			buffer[y] = line{bytes: l.bytes, chars: l.chars, words: l.words}
			l.text = nil
		}
	}

	total = counts{l.chars, l.words, len(sent), len(para), len(sect)}
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
	cursor.pos[Sect] = 1
	win = nc.StdScr()
	win.Clear()

	if Sx > margin && Sy > 1 {
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
	i := len(document) - 1
	if document[i] != '\n' {
		AppendRune([]byte{'\f'})
	} else {
		document[i] = '\f'
		if cursor.y > 1 {
			cursor.y -= 2
		}
		DrawWindow()
	}
	scope = Sect
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
	DrawCursor()
	DrawStatusBar()
}
