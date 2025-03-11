package edits

import (
	"bytes"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	tt "github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/xanni/jotty/i18n"
)

func setupModel(t *testing.T) *tt.TestModel {
	setupTest()
	sx, sy = 15, 3
	ResizeScreen(sx, sy)
	m.timer = time.NewTimer(time.Minute)
	tm := tt.NewTestModel(t, m, tt.WithInitialTermSize(sx, sy))
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("@0/0")) })

	return tm
}

func TestIsSizeOK(t *testing.T) {
	assert := assert.New(t)
	sx, sy = 5, 3
	assert.False(isSizeOK())

	sx, sy = 6, 2
	assert.False(isSizeOK())

	sx, sy = 6, 3
	assert.True(isSizeOK())
}

func TestCuts(t *testing.T) {
	tm := setupModel(t)

	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Type("test")
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyDelete})
	tm.Send(tea.KeyMsg{Type: tea.KeyPgDown})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("test")) })

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("@0/0")) })

	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Type("more")
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyDelete})
	tm.Send(tea.KeyMsg{Type: tea.KeyPgDown})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("test")) })

	tm.Send(tea.KeyMsg{Type: tea.KeyPgDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("@4/4")) })

	tm.Send(tea.KeyMsg{Type: tea.KeyPgUp})
	tm.Send(tea.KeyMsg{Type: tea.KeyPgUp})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("test")) })

	tm.Type(".")
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("@5/5")) })
}

func TestHelp(t *testing.T) {
	tm := setupModel(t)
	i18n.HelpText, i18n.HelpWidth = []string{"Help text"}, 9

	tm.Type("test")
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("test_")) })

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("Help text")) })

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("test_")) })
}

func TestQuit(t *testing.T) {
	tm := setupModel(t)

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlQ})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("Confirm exit?")) })

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("@0/0")) })

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlQ})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	tm.WaitFinished(t, tt.WithFinalTimeout(time.Second))
}

func TestTooSmall(t *testing.T) {
	tm := setupModel(t)

	tm.Send(tea.WindowSizeMsg{Width: 1, Height: 1})
	tm.Type("x") // Should be ignored
	tm.Send(tea.WindowSizeMsg{Width: 15, Height: 3})
	tt.WaitFor(t, tm.Output(), func(bts []byte) bool { return bytes.Contains(bts, []byte("@0/0")) })
}
