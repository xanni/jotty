package edits

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
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

var cursor struct {
	pos  int
	x, y int
}
var primedia []rune
var scope Scope

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

// AppendRune appends a rune to the Primedia scroll.
func AppendRune(r rune) {
	primedia = append(primedia, r)
	scope = Char
	cursor.pos++
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
	chars := strconv.Itoa(cursor.pos)
	status := ID + "  c" + chars + "/" + chars
	if len(status) > screenWidth {
		status = status[len(ID)+2:]
	}
	drawStringNoWrap(sr, status, 0, 0, tcell.StyleDefault)
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

// DrawWindow draws the edit window.
func DrawWindow() {
	screenWidth, _ := Screen.Size()
	if screenWidth < margin+1 {
		drawResizeRequest()
		return
	}

	start := max(0, len(primedia)-screenWidth+1)
	for i, r := range primedia[start:] {
		Screen.SetContent(i, 0, r, nil, tcell.StyleDefault)
	}
	end := min(len(primedia), screenWidth-1)
	if cursor.x < end {
		cursor.x = end
		DrawCursor()
	}
	DrawStatusBar()
}
