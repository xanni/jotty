package edits

import (
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

var CursorRune = [...]rune{'_', '#', '$', '¶', '§'}
var cursorStyle = tcell.StyleDefault.Blink(true)
var errorStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Reverse(true).Blink(true)

const margin = 5 // Up to 3 edit marks, cursor and wrap indicator

// ID is the program name and version.
var ID string
var Screen tcell.Screen

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
	char, word int // current character and word in the document
	x, y       int // current position in the edit window
}
var document []byte
var total struct {
	chars, words int
}

// ScreenRegion defines a rectangular region of a screen.
type ScreenRegion struct {
	screen              tcell.Screen
	x, y, width, height int
}

// Fill fills a rectangular region of the screen with a character.
func (sr *ScreenRegion) Fill(r rune, style tcell.Style) {
	for x, y := 0, 0; y < sr.height; {
		sr.screen.SetContent(sr.x+x, sr.y+y, r, nil, style)
		x++
		if x >= sr.width {
			x = 0
			y++
		}
	}
}

// AppendRune appends a rune to the document.
func AppendRune(r rune) {
	document = utf8.AppendRune(document, r)
	scope = Char
	if len(buffer) == 0 {
		cursor.char += uniseg.GraphemeClusterCount(string(document))
	} else if uniseg.GraphemeClusterCount(string(document[buffer[cursor.y].bytes:])) >
		uniseg.GraphemeClusterCount(string(buffer[cursor.y].text)) {
		cursor.char++
	}
	DrawWindow()
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
	Screen.SetContent(cursor.x, cursor.y, CursorRune[scope], nil, cursorStyle)
	Screen.ShowCursor(cursor.x, cursor.y)
}

func drawStringNoWrap(sr *ScreenRegion, s string, col int, row int, style tcell.Style) int {
	for _, r := range s {
		if col >= sr.width {
			break
		}
		sr.screen.SetContent(sr.x+col, sr.y+row, r, nil, style)
		col++
	}

	return col
}

// DrawStatusBar draws a status bar on the last line of the screen.
func DrawStatusBar() {
	screenWidth, screenHeight := Screen.Size()
	if screenHeight == 0 {
		return
	}

	sr := &ScreenRegion{Screen, 0, screenHeight - 1, screenWidth, 1}
	sr.Fill(' ', tcell.StyleDefault)
	chars := "c" + strconv.Itoa(cursor.char) + "/" + strconv.Itoa(total.chars)
	words := "#" + strconv.Itoa(cursor.word) + "/" + strconv.Itoa(total.words)
	status := words + " " + chars
	var x int
	if screenWidth >= len(ID)+2+len(status) {
		drawStringNoWrap(sr, ID, 0, 0, tcell.StyleDefault)
		x = len(ID) + 2
	} else if screenWidth < len(status) {
		status = chars
	}
	drawStringNoWrap(sr, status, x, 0, tcell.StyleDefault)
}

func drawResizeRequest() {
	screenWidth, screenHeight := Screen.Size()
	if screenWidth < 1 || screenHeight < 1 {
		return
	}

	Screen.Clear()
	row := (screenHeight - 1) / 2
	Screen.SetContent(0, row, '<', nil, errorStyle)
	for x := 1; x < screenWidth-1; x++ {
		Screen.SetContent(x, row, '-', nil, errorStyle)
	}
	Screen.SetContent(screenWidth-1, row, '>', nil, errorStyle)
}

/*
DrawWindow draws the edit window.

It word wraps, buffers and displays a portion of the document starting from
the line the cursor is on and ending at the last line of the edit window
or the end of the document, whichever comes first.
*/
func DrawWindow() {
	screenWidth, screenHeight := Screen.Size()
	if screenWidth < margin+1 || screenHeight < 2 {
		drawResizeRequest()
		return
	}

	var l line   // current line with cumulative counts
	var x, y int // screen coordinates

	if len(buffer) == 0 {
		buffer = make([]line, screenHeight-1) // redraw everything
	} else {
		y = cursor.y
		l = line{bytes: buffer[y].bytes, chars: buffer[y].chars, words: buffer[y].words}
	}
	source := document[l.bytes:]
	state := -1

	for {
		if cursor.char == l.chars {
			cursor.word = l.words
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
		w := f >> uniseg.ShiftWidth // monospace width of character
		l.bytes += len(c)
		if w > 0 || c[0] == '\n' {
			l.chars++
		}

		if f&uniseg.MaskWord != 0 {
			// Is the first rune in the grapheme cluster alphanumeric?
			r, _ := utf8.DecodeRune(c)
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				l.words++
			}
		}

		if w > 0 {
			l.text = append(l.text, c...)
			r1, s1 := utf8.DecodeRune(c) // first rune and size
			var cr []rune                // combining runes
			for i := s1; i < len(c); {
				r, s := utf8.DecodeRune(c[i:])
				cr = append(cr, r)
				i += s
			}
			Screen.SetContent(x, y, r1, cr, tcell.StyleDefault)
			x += w
		}

		seg, _, _, _ = uniseg.FirstLineSegment(source, -1)
		nw := uniseg.StringWidth(string(seg)) // width of next breakable segment

		// Break if at margin or mandatory break that is not just end of source
		f &= uniseg.MaskLine
		br := x >= screenWidth-1 ||
			(f == uniseg.LineCanBreak && x+w+nw >= screenWidth-1) ||
			(f == uniseg.LineMustBreak && (len(source) > 0 || uniseg.HasTrailingLineBreak(document)))

		if br {
			if x == screenWidth-1 && f != uniseg.LineCanBreak {
				Screen.SetContent(x, y, '-', nil, tcell.StyleDefault.Reverse(true))
			} else {
				for x < screenWidth {
					Screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
					x++
				}
			}

			buffer[y].text = l.text
			if y >= screenHeight-2 { // last line of the window
				break // TODO scroll
			}

			x = 0
			y++
			buffer[y] = line{bytes: l.bytes, chars: l.chars, words: l.words}
			l.text = nil
		}
	}

	total.chars = l.chars
	total.words = l.words
	DrawStatusBar()
}

func ResizeScreen() {
	Screen.Clear()
	buffer = nil
	DrawWindow()
	Screen.Sync()
}
