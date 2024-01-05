package edits

import (
	"strconv"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var CursorRune = [...]rune{'_', '#', '$', '¶', '§'}

var cursorStyle = tcell.StyleDefault.Blink(true)

// ID is the program name and version.
var ID string

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
func AppendRune(screen tcell.Screen, r rune) {
	primedia = append(primedia, r)
	scope = Char
	cursor.pos++
	DrawStatusBar(screen)
	DrawWindow(screen)
}

func DecScope(screen tcell.Screen) {
	if scope == Char {
		scope = Sect
	} else {
		scope--
	}
	DrawCursor(screen)
}

func IncScope(screen tcell.Screen) {
	if scope == Sect {
		scope = Char
	} else {
		scope++
	}
	DrawCursor(screen)
}

// DrawCursor draws the cursor.
func DrawCursor(screen tcell.Screen) {
	screen.SetContent(cursor.x, cursor.y, CursorRune[scope], nil, cursorStyle)
}

func drawStringNoWrap(sr *ScreenRegion, s string, col int, row int, style tcell.Style) int {
	for i := 0; i < len(s); {
		r, rsize := utf8.DecodeRuneInString(s[i:])
		i += rsize
		if col+rsize > sr.width {
			break
		}
		sr.screen.SetContent(sr.x+col, sr.y+row, r, nil, style)
		col += rsize
	}

	return col
}

// DrawStatusBar draws a status bar on the last line of the screen.
func DrawStatusBar(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 {
		return
	}

	sr := &ScreenRegion{screen, 0, screenHeight - 1, screenWidth, 1}
	sr.Fill(' ', tcell.StyleDefault)
	chars := strconv.Itoa(cursor.pos)
	status := ID + "  c" + chars + "/" + chars
	if len(status) > screenWidth {
		status = status[len(ID)+2:]
	}
	drawStringNoWrap(sr, status, 0, 0, tcell.StyleDefault)
}

// DrawWindow draws the edit window.
func DrawWindow(screen tcell.Screen) {
	screenWidth, _ := screen.Size()
	start := max(0, len(primedia)-screenWidth+1)
	for i, r := range primedia[start:] {
		screen.SetContent(i, 0, r, nil, tcell.StyleDefault)
	}
	end := min(len(primedia), screenWidth-1)
	if cursor.x < end {
		cursor.x = end
		DrawCursor(screen)
	}
}
