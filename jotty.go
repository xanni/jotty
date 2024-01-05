package main

import (
	"log"
	"os"
	"unicode/utf8"

	"git.sericyb.com.au/jotty/edits"
	"github.com/gdamore/tcell/v2"
)

const version = "0"

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

// DrawStatusBar draws a state bar on the last line of the screen.
func DrawStatusBar(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()
	if screenHeight == 0 {
		return
	}

	sr := &ScreenRegion{screen, 0, screenHeight - 1, screenWidth, 1}
	sr.Fill(' ', tcell.StyleDefault)
	drawStringNoWrap(sr, "Jotty v"+version, 0, 0, tcell.StyleDefault)
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

	DrawStatusBar(s)
	edits.DrawCursor(s)

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
			case tcell.KeyRune:
				edits.AppendRune(s, ev.Rune())
			case tcell.KeyEsc:
				quit()
			case tcell.KeyCtrlQ:
				quit()
			case tcell.KeyCtrlW:
				quit()
			case tcell.KeyUp:
				edits.IncScope(s)
			case tcell.KeyDown:
				edits.DecScope(s)
			}
		}
	}
}
