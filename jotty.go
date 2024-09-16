package main

import (
	"log"
	"os"
	"strings"
	"time"

	"git.sericyb.com.au/jotty/edits"
	ps "git.sericyb.com.au/jotty/permascroll"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultName = "jotty.jot"
	syncDelay   = 10 * time.Second
	version     = "0"
)

var dispatch = map[tea.KeyType]func(){
	tea.KeyUp: edits.IncScope, tea.KeyDown: edits.DecScope,
	tea.KeyLeft: edits.Left, tea.KeyRight: edits.Right,
	tea.KeyCtrlC: edits.Copy,
	tea.KeyEnd:   edits.End, tea.KeyCtrlD: edits.End,
	tea.KeyCtrlE:     export,
	tea.KeyBackspace: edits.Backspace, tea.KeyCtrlH: edits.Backspace,
	tea.KeyTab: edits.Mark, tea.KeyShiftTab: edits.ClearMarks,
	tea.KeyEnter: edits.Enter, tea.KeySpace: edits.Space,
	tea.KeyHome: edits.Home, tea.KeyCtrlU: edits.Home,
	tea.KeyInsert: edits.InsertCut, tea.KeyCtrlV: edits.InsertCut,
	tea.KeyDelete: edits.Delete, tea.KeyCtrlX: edits.Delete,
	tea.KeyCtrlY: edits.Redo, tea.KeyCtrlZ: edits.Undo,
}

var (
	exportPath = "jotty.txt"
	sx, sy     int // screen dimensions
)

type model struct{ timer *time.Timer }

func export() { edits.Export(exportPath) }

// True if the window is sufficiently large.
func isSizeOK() bool {
	return sx > 5 && sy > 2
}

func (m model) Init() tea.Cmd {
	edits.ID = "Jotty v" + version

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sx, sy = msg.Width, msg.Height
		edits.ResizeScreen(msg.Width, msg.Height)
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlQ || msg.Type == tea.KeyCtrlW {
			return m, tea.Quit
		}
		if isSizeOK() {
			m.timer.Reset(syncDelay)
			if f, ok := dispatch[msg.Type]; ok {
				f()
			} else if msg.Type == tea.KeyRunes && !msg.Alt {
				edits.InsertRunes(msg.Runes)
			}
		}
	}

	return m, nil
}

func (m model) View() (s string) {
	if isSizeOK() {
		s = edits.Screen()
	}

	return s
}

func cleanup() {
	ps.Flush()
	if err := ps.ClosePermascroll(); err != nil {
		log.Printf("%+v", err)
	}
}

func main() {
	path := defaultName
	if len(os.Args) > 1 {
		exportPath, path = os.Args[1], os.Args[1]
		if i := strings.LastIndex(exportPath, ".jot"); i >= 0 {
			exportPath = exportPath[:i]
		}
		exportPath += ".txt"
	}

	if err := ps.OpenPermascroll(path); err != nil {
		log.Fatalf("%+v", err)
	}
	defer cleanup()

	var m model
	m.timer = time.AfterFunc(syncDelay, func() {
		if err := ps.SyncPermascroll(); err != nil {
			log.Printf("%+v", err)
		}
	})
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("%+v", err)
	}
	m.timer.Stop()
}
