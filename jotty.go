package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.sericyb.com.au/jotty/edits"
	"git.sericyb.com.au/jotty/i18n"
	ps "git.sericyb.com.au/jotty/permascroll"
	tea "github.com/charmbracelet/bubbletea"
)

//go:generate sh -c "printf %s $(git describe --always --tags) > version.txt"
//go:embed version.txt
var version string

const (
	defaultName = "jotty.jot"
	syncDelay   = 10 * time.Second
)

var dispatch = map[tea.KeyType]func(){
	tea.KeyEsc: help,
	tea.KeyUp:  edits.IncScope, tea.KeyDown: edits.DecScope,
	tea.KeyLeft: edits.Left, tea.KeyRight: edits.Right,
	tea.KeyCtrlC: edits.Copy,
	tea.KeyEnd:   edits.End, tea.KeyCtrlD: edits.End,
	tea.KeyCtrlE:     export,
	tea.KeyBackspace: edits.Backspace, tea.KeyCtrlH: edits.Backspace,
	tea.KeyTab: edits.Mark, tea.KeyShiftTab: edits.ClearMarks,
	tea.KeyCtrlJ: edits.Join,
	tea.KeyEnter: edits.Enter, tea.KeySpace: edits.Space,
	tea.KeyCtrlQ: confirmExit, tea.KeyCtrlW: confirmExit,
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

func confirmExit() { edits.SetMode(edits.Quit, i18n.Text["confirm"]) }
func export()      { edits.Export(exportPath) }
func help()        { edits.SetMode(edits.Help, "") }

// True if the window is sufficiently large.
func isSizeOK() bool {
	return sx > 5 && sy > 2
}

func (m model) Init() tea.Cmd {
	edits.ID = "Jotty " + version

	return nil
}

func acceptKey(m *model, msg tea.KeyMsg) {
	if isSizeOK() {
		m.timer.Reset(syncDelay)
		if f, ok := dispatch[msg.Type]; ok {
			f()
		} else if msg.Type == tea.KeyRunes && !msg.Alt {
			edits.InsertRunes(msg.Runes)
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sx, sy = msg.Width, msg.Height
		edits.ResizeScreen(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch edits.Mode {
		case edits.Error:
			if msg.Type == tea.KeySpace || msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc {
				edits.ClearMode()
			}
		case edits.Help:
			if msg.Type == tea.KeyEsc {
				edits.ClearMode()
			}
		case edits.Quit:
			switch msg.Type {
			case tea.KeyEsc:
				edits.ClearMode()
			case tea.KeySpace, tea.KeyEnter, tea.KeyCtrlQ, tea.KeyCtrlW:
				return m, tea.Quit
			}
		default:
			acceptKey(&m, msg)
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

func usage() {
	fmt.Println("https://github.com/xanni/jotty  ⓒ 2024 Andrew Pam <xanni@xanadu.net>")
	fmt.Printf("\n"+i18n.Text["usage"]+"\n", filepath.Base(os.Args[0]), defaultName)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	vFlag := flag.Bool("version", false, i18n.Text["version"])
	flag.Parse()
	if *vFlag {
		println(filepath.Base(os.Args[0]) + " " + version)
		os.Exit(0)
	}

	path := defaultName
	if len(os.Args) > 1 {
		exportPath, path = flag.Arg(0), flag.Arg(0)
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
