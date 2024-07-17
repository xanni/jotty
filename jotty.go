package main

import (
	"log"
	"os"

	"git.sericyb.com.au/jotty/edits"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "0"

var dispatch = map[tea.KeyType]func(){
	tea.KeyUp:        edits.IncScope,
	tea.KeyDown:      edits.DecScope,
	tea.KeyEnter:     edits.Enter,
	tea.KeySpace:     edits.Space,
	tea.KeyLeft:      edits.Left,
	tea.KeyRight:     edits.Right,
	tea.KeyHome:      edits.Home,
	tea.KeyCtrlU:     edits.Home,
	tea.KeyEnd:       edits.End,
	tea.KeyCtrlD:     edits.End,
	tea.KeyBackspace: edits.Backspace,
	tea.KeyCtrlH:     edits.Backspace,
}

var sx, sy int // screen dimensions

type model struct{}

// True if the window is sufficiently large
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

func main() {
	var m model
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("%+v", err)
		os.Exit(1)
	}
}
