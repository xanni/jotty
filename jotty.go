package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"git.sericyb.com.au/jotty/edits"
	nc "github.com/vit1251/go-ncursesw"
	"golang.org/x/term"
)

const version = "0"

const CTRL_Q, CTRL_W = '\x11', '\x17'

func resize(s *nc.Screen) {
	s.End() // The old screen must be closed to re-initialise ncurses

	var err error
	edits.Sx, edits.Sy, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("%+v", err)
	}

	edits.ResizeScreen()
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGWINCH)

	// Initialize screen
	s, err := nc.NewTerm("", os.Stdout, os.Stdin)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	win := nc.StdScr()

	nc.Raw(true)                            // turn on raw "uncooked" input
	nc.Echo(false)                          // turn echoing of typed characters off
	nc.StartColor()                         // enable color if available, ignore error if not
	nc.InitPair(1, nc.C_YELLOW, nc.C_BLACK) // errors are black on yellow like caution tape
	win.Keypad(true)                        // enable function keys

	quit := func() {
		// Shutdown ncurses and restore the terminal before printing any diagnostics
		maybePanic := recover()
		s.End()
		s.Delete()
		if maybePanic != nil {
			panic(maybePanic)
		}
		os.Exit(0)
	}
	defer quit()

	edits.ID = "Jotty v" + version
	edits.Screen = *s
	resize(s)

	for {
		// Update screen
		nc.Update()

		select {
		case <-sigs:
			resize(s)
		default:
			// Process input
			key := win.GetChar()
			switch {
			case key == nc.KEY_RESIZE:
				resize(s)
			case key == nc.KEY_ESC || key == CTRL_Q || key == CTRL_W:
				quit()
			case key == nc.KEY_UP:
				edits.IncScope()
			case key == nc.KEY_DOWN:
				edits.DecScope()
			case key >= 32 && key <= 255:
				edits.AppendByte(byte(key))
			}
		}
	}
}
