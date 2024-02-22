package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"git.sericyb.com.au/jotty/edits"
	nc "github.com/vit1251/go-ncursesw"
	"golang.org/x/term"
)

const version = "0"

const CTRL_D, CTRL_Q, CTRL_U, CTRL_W = '\x04', '\x11', '\x15', '\x17'

var dispatch = map[nc.Key]func(){
	nc.KEY_UP:     edits.IncScope,
	nc.KEY_DOWN:   edits.DecScope,
	nc.KEY_RETURN: edits.Enter,
	nc.KEY_ENTER:  edits.Enter,
	' ':           edits.Space,
	nc.KEY_LEFT:   edits.Left,
	nc.KEY_RIGHT:  edits.Right,
	nc.KEY_HOME:   edits.Home,
	CTRL_U:        edits.Home,
	nc.KEY_END:    edits.End,
	CTRL_D:        edits.End,
}

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

	nc.Raw(true)                            // enable raw "uncooked" input
	nc.Echo(false)                          // disable echoing of input characters
	nc.StartColor()                         // enable color if available, ignore error if not
	nc.InitPair(1, nc.C_YELLOW, nc.C_BLACK) // errors are black on yellow like caution tape
	win.Keypad(true)                        // enable function keys
	win.ScrollOk(true)
	win.Timeout(100) // 100ms timeout to ensure SIGWINCH gets processed

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
	edits.Sy, edits.Sx = win.MaxYX()
	edits.ResizeScreen()

	var crune []byte // current rune being appended

	for {
		// Update screen
		nc.Update()

		select {
		case <-sigs:
			resize(s)
		default:
			// Process input
			key := win.GetChar()
			if key == nc.KEY_RESIZE {
				resize(s)
			} else if key == nc.KEY_ESC || key == CTRL_Q || key == CTRL_W {
				quit()
			}

			if !edits.IsSizeOK() {
				continue
			}

			if key <= 32 || key > 255 {
				crune = nil
				if f, ok := dispatch[key]; ok {
					f()
				}
			} else {
				crune = append(crune, byte(key))
				r, _ := utf8.DecodeLastRune(crune)
				if r != utf8.RuneError {
					edits.AppendRune(crune)
					crune = nil
				}
			}
		}
	}
}
