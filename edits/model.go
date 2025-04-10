package edits

import (
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xanni/jotty/i18n"
	ps "github.com/xanni/jotty/permascroll"
)

const syncDelay = 10 * time.Second

var dispatch = map[tea.KeyType]func(){
	tea.KeyEsc: _help,
	tea.KeyUp:  IncScope, tea.KeyDown: DecScope,
	tea.KeyLeft: Left, tea.KeyRight: Right,
	tea.KeyCtrlC: Copy,
	tea.KeyEnd:   End, tea.KeyCtrlD: End,
	tea.KeyCtrlE:     _export,
	tea.KeyBackspace: Backspace, tea.KeyCtrlH: Backspace,
	tea.KeyTab: Mark, tea.KeyShiftTab: ClearMarks,
	tea.KeyCtrlJ: Join,
	tea.KeyEnter: Enter, tea.KeySpace: Space,
	tea.KeyPgDown: NextCut, tea.KeyCtrlN: NextCut,
	tea.KeyPgUp: PrevCut, tea.KeyCtrlP: PrevCut,
	tea.KeyCtrlQ: _quit, tea.KeyCtrlW: _quit,
	tea.KeyHome: Home, tea.KeyCtrlU: Home,
	tea.KeyInsert: InsertCut, tea.KeyCtrlV: InsertCut,
	tea.KeyDelete: Delete, tea.KeyCtrlX: Delete,
	tea.KeyCtrlY: Redo, tea.KeyCtrlZ: Undo,
}

var exportDispatch = map[tea.KeyType]func(){
	tea.KeyEsc:  ClearMode,
	tea.KeyLeft: PromptLeft, tea.KeyRight: PromptRight,
	tea.KeyEnd: PromptEnd, tea.KeyCtrlD: PromptEnd,
	tea.KeyBackspace: PromptBackspace, tea.KeyCtrlH: PromptBackspace,
	tea.KeyHome: PromptHome, tea.KeyCtrlU: PromptHome,
}

var (
	exportMarkedPath, exportPath string // Paths for export of marked portations and entire document
	sx, sy                       int    // screen dimensions
)

type model struct{ timer *time.Timer }

var m model

func _export() {
	SetMode(PromptExport, IconExport)

	if len(mark) == 0 {
		PromptDefault(exportPath)
	} else {
		if len(mark) == 1 && cursor[Char] != mark[0] {
			updateSelections()
			drawPara(markPara)
		}

		PromptDefault(exportMarkedPath)
	}
}

func _help() { SetMode(Help, "") }
func _quit() { SetMode(ConfirmQuit, i18n.Text["confirm"]) }

// True if the window is sufficiently large.
func isSizeOK() bool { return sx > margin && sy > 2 }

func Run(version, path string) {
	name, exportPath = "Jotty "+version, path

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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) acceptKey(msg tea.KeyMsg) {
	m.timer.Reset(syncDelay)
	if f, ok := dispatch[msg.Type]; ok {
		f()
	} else if msg.Type == tea.KeyRunes && !msg.Alt {
		InsertRunes(msg.Runes)
	}
}

func (m model) cutsKey(key tea.KeyMsg) {
	switch key.Type {
	case tea.KeyEsc:
		ClearMode()
	case tea.KeyPgDown, tea.KeyCtrlN:
		NextCut()
	case tea.KeyPgUp, tea.KeyCtrlP:
		PrevCut()
	case tea.KeySpace, tea.KeyEnter, tea.KeyInsert, tea.KeyCtrlV:
		ClearMode()
		InsertCut()
	case tea.KeyRunes:
		if !key.Alt {
			m.timer.Reset(syncDelay)
			ClearMode()
			InsertRunes(key.Runes)
		}
	}
}

func (m model) exportKey(key tea.KeyMsg) {
	if f, ok := exportDispatch[key.Type]; ok {
		f()

		return
	}

	switch key.Type {
	case tea.KeyEnter:
		path := PromptResponse()
		if len(mark) > 0 {
			exportMarkedPath = path
		} else {
			exportPath = path
		}

		if f, err := os.Stat(path); err == nil && f.Size() > 0 {
			SetMode(ConfirmOverwrite, i18n.Text["overwrite"])
		} else {
			Export(path)
		}
	case tea.KeyRunes, tea.KeySpace:
		if !key.Alt {
			PromptInsertRunes(key.Runes)
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sx, sy = msg.Width, msg.Height
		ResizeScreen(msg.Width, msg.Height)
	case tea.KeyMsg:
		if !isSizeOK() {
			break
		}

		switch Mode {
		case Cuts:
			m.cutsKey(msg)
		case ConfirmOverwrite:
			switch msg.Type {
			case tea.KeyEsc:
				_export()
			case tea.KeyEnter, tea.KeyCtrlE:
				Export(exportPath)
			}
		case ConfirmQuit:
			switch msg.Type {
			case tea.KeyEsc:
				ClearMode()
			case tea.KeyEnter, tea.KeyCtrlQ, tea.KeyCtrlW:
				return m, tea.Quit
			}
		case Error:
			if msg.Type == tea.KeySpace || msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc {
				ClearMode()
			}
		case Help:
			if msg.Type == tea.KeyEsc {
				ClearMode()
			}
		case PromptExport:
			m.exportKey(msg)
		default:
			m.acceptKey(msg)
		}
	}

	return m, nil
}

func (m model) View() (s string) {
	if isSizeOK() {
		s = Screen()
	}

	return s
}
