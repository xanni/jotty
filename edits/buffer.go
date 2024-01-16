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

var CursorChar = [...]nc.Char{'_', '#', '$', '¶', '§'}
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
)

var scope Scope

type line struct {
	bytes, chars, words int // cumulative counts at start of line
	text                []byte
}

var buffer []line
var cursor struct {
	char, word, sent, para int // current position in the document
	x, y                   int // current position in the edit window
}
var document []byte
var para = []int{0} // byte index of each paragraph in the document
var sent = []int{0} // byte index of each sentence in the document
var total struct {
	chars, words, sents, paras int
}
var win *nc.Window

// AppendRune appends a UTF-8 encoded rune to the document.
func AppendRune(r []byte) {
	document = append(document, r...)
	scope = Char

	if len(buffer) == 0 {
		cursor.char = uniseg.GraphemeClusterCount(string(document))
	} else {
		cursor.char = buffer[cursor.y].chars + uniseg.GraphemeClusterCount(string(document[buffer[cursor.y].bytes:]))
		DrawWindow()
	}
}

func DecScope() {
	if scope == Char {
		scope = Sect
	} else {
		scope--
	}
	DrawCursor()
}

func IncScope() {
	if scope == Sect {
		scope = Char
	} else {
		scope++
	}
	DrawCursor()
}

// DrawCursor draws the cursor.
func DrawCursor() {
	win.AttrSet(nc.Char(cursorStyle))
	win.MovePrint(cursor.y, cursor.x, string(rune(CursorChar[scope])))
	win.AttrSet(nc.A_NORMAL)
	win.Move(cursor.y, cursor.x)
	win.NoutRefresh()
}

// DrawStatusBar draws a status bar on the last line of the screen.
func DrawStatusBar() {
	win.Move(Sy-1, 0)
	win.ClearToBottom()
	chars := "@" + strconv.Itoa(cursor.char) + "/" + strconv.Itoa(total.chars)
	words := string(rune(CursorChar[Word])) + strconv.Itoa(cursor.word) + "/" + strconv.Itoa(total.words)
	sents := string(rune(CursorChar[Sent])) + strconv.Itoa(cursor.sent) + "/" + strconv.Itoa(total.sents)
	paras := string(rune(CursorChar[Para])) + strconv.Itoa(cursor.para) + "/" + strconv.Itoa(total.paras)
	status := paras + " " + sents + " " + words + " " + chars

	var x int
	if Sx >= len(ID)+2+len(status) {
		win.MovePrint(Sy-1, 0, ID)
		x = len(ID) + 2
	} else if Sx < len(status) {
		switch scope {
		case Char:
			status = chars
		case Word:
			status = words
		case Sent:
			status = sents
		case Para:
			status = paras
		}
	}

	win.MovePrint(Sy-1, x, status)
	win.Move(cursor.y, cursor.x)
	win.NoutRefresh()
}

/*
DrawWindow draws the edit window.

It word wraps, buffers and displays a portion of the document starting from
the line the cursor is on and ending at the last line of the edit window
or the end of the document, whichever comes first.
*/
func DrawWindow() {
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
		if cursor.char == l.chars {
			cursor.word = l.words
			cursor.sent = sort.Search(len(sent), func(i int) bool { return sent[i] >= l.bytes })
			cursor.para = sort.Search(len(para), func(i int) bool { return para[i] >= l.bytes })
			cursor.x = x
			cursor.y = y
			DrawCursor()
			x++
		}

		if len(source) == 0 {
			buffer[y].text = l.text
			break
		}

		var c []byte   // grapheme cluster
		var f int      // Unicode boundary flags
		var seg []byte // next breakable segment

		c, source, f, state = uniseg.Step(source, state)
		r, _ := utf8.DecodeRune(c)
		if r == utf8.RuneError {
			continue
		}

		if len(source) > 0 && f&uniseg.MaskSentence != 0 && l.bytes > sent[len(sent)-1] {
			sent = append(sent, l.bytes)
		}

		if c[0] == '\n' && l.bytes > para[len(para)-1] {
			para = append(para, l.bytes)
		}

		w := f >> uniseg.ShiftWidth // monospace width of character
		l.bytes += len(c)
		if w > 0 || c[0] == '\n' {
			l.chars++
		}

		// Is the first rune in the grapheme cluster alphanumeric?
		if f&uniseg.MaskWord != 0 && (unicode.IsLetter(r) || unicode.IsNumber(r)) {
			l.words++
		}

		if w > 0 {
			l.text = append(l.text, c...)
			win.MovePrint(y, x, string(c))
			x += w
		}

		seg, _, _, _ = uniseg.FirstLineSegment(source, -1)
		nw := uniseg.StringWidth(string(seg)) // width of next breakable segment

		// Break if at margin or mandatory break that is not just end of source
		f &= uniseg.MaskLine
		br := x >= Sx-1 ||
			(f == uniseg.LineCanBreak && x+w+nw >= Sx-1) ||
			(f == uniseg.LineMustBreak && (len(source) > 0 || uniseg.HasTrailingLineBreak(document)))

		if br {
			if x == Sx-1 && f != uniseg.LineCanBreak {
				win.MoveAddChar(y, x, '-'|nc.A_REVERSE)
			} else {
				win.Move(y, x)
				win.ClearToEOL()
			}

			buffer[y].text = l.text
			x = 0
			y++
			if c[0] == '\n' {
				y++
			}
			if y >= Sy-1 { // last line of the window
				break // TODO scroll
			}

			buffer[y] = line{bytes: l.bytes, chars: l.chars, words: l.words}
			l.text = nil
		}
	}

	total.chars = l.chars
	total.words = l.words
	total.sents = len(sent)
	total.paras = len(para)
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

	if Sx > margin && Sy > 1 {
		DrawWindow()
	} else {
		drawResizeRequest()
	}
}
