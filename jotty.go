package main

import (
	"log"
	"os"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

const version = "0"

var CursorStyle = tcell.StyleDefault.Blink(true)

type Scope int

const (
	Char Scope = iota
	Word
	Sent // Sentence
	Para // Paragraph
	Sect // Section
)

var CursorRune = [...]rune{'_', '#', '$', '¶', '§'}

type Cursor struct {
	x, y int
}

// State defines the current editor state
type State struct {
	screen tcell.Screen
	cursor Cursor
	scope  Scope
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

// DrawCursor draws the cursor.
func (state *State) DrawCursor() {
	state.screen.SetContent(state.cursor.x, state.cursor.y, CursorRune[state.scope], nil, CursorStyle)
}

// DrawStatusBar draws a state bar on the last line of the screen.
func (state *State) DrawStatusBar() {
	screenWidth, screenHeight := state.screen.Size()
	if screenHeight == 0 {
		return
	}

	sr := &ScreenRegion{state.screen, 0, screenHeight - 1, screenWidth, 1}
	sr.Fill(' ', tcell.StyleDefault)
	drawStringNoWrap(sr, "Jotty v"+version, 0, 0, tcell.StyleDefault)
}

func (state *State) DecScope() {
	if state.scope == Char {
		state.scope = Sect
	} else {
		state.scope--
	}
	state.DrawCursor()
}

func (state *State) IncScope() {
	if state.scope == Sect {
		state.scope = Char
	} else {
		state.scope++
	}
	state.DrawCursor()
}

func main() {
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	s.SetStyle(defStyle)
	s.Clear()

	quit := func() {
		// Shutdown tcell and restore the terminal before printing any diagnostics
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
		os.Exit(0)
	}
	defer quit()

	state := State{screen: s}
	state.DrawStatusBar()
	state.DrawCursor()

	for {
		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEsc:
				quit()
			case tcell.KeyCtrlQ:
				quit()
			case tcell.KeyCtrlW:
				quit()
			case tcell.KeyUp:
				state.IncScope()
			case tcell.KeyDown:
				state.DecScope()
			}
		}
	}
}
