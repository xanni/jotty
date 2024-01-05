package edits

import (
	"github.com/gdamore/tcell/v2"
)

var CursorRune = [...]rune{'_', '#', '$', '¶', '§'}

var cursorStyle = tcell.StyleDefault.Blink(true)

type Scope int

const (
	Char Scope = iota
	Word
	Sent // Sentence
	Para // Paragraph
	Sect // Section
)

var cursor struct {
	offset int
	x, y   int
}

var primedia []rune

var scope Scope

// AppendRune appends a rune to the Primedia scroll.
func AppendRune(screen tcell.Screen, r rune) {
	primedia = append(primedia, r)
	scope = Char
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
