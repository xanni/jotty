package main

import (
	"log"
	"os"

	"git.sericyb.com.au/jotty/edits"
	"github.com/gdamore/tcell/v2"
)

const version = "0"

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

	edits.ID = "Jotty v" + version
	edits.Screen = s
	edits.DrawWindow()
	edits.DrawCursor()

	for {
		// Update screen
		s.Show()

		// Poll event
		ev := s.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			edits.DrawWindow()
			s.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				edits.AppendRune(ev.Rune())
			case tcell.KeyEsc:
				quit()
			case tcell.KeyCtrlQ:
				quit()
			case tcell.KeyCtrlW:
				quit()
			case tcell.KeyUp:
				edits.IncScope()
			case tcell.KeyDown:
				edits.DecScope()
			}
		}
	}
}
