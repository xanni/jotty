package edits

import (
	"strings"
	"time"

	ps "github.com/xanni/jotty/permascroll"
)

const (
	layout = time.DateTime
	minCut = 5
)

// Select previous cut.
func PrevCut() {
	if ps.Cuts() > 0 {
		Mode = Cuts
		currentCut--
		if currentCut < 1 {
			currentCut = ps.Cuts()
		}
	}
}

// Select next cut.
func NextCut() {
	if ps.Cuts() > 0 {
		Mode = Cuts
		currentCut++
		if currentCut > ps.Cuts() {
			currentCut = 1
		}
	}
}

func drawCut(current bool, text string, ts time.Time) (s string) {
	maxLen := ex - len(layout) - 2
	if maxLen < minCut {
		maxLen = ex - 1
	} else {
		switch {
		case ts.IsZero():
			s = strings.Repeat(" ", len(layout)+1)
		case current:
			s = cutCurStyle(ts.Format(layout)) + " "
		default:
			s = cutTimeStyle(ts.Format(layout)) + " "
		}
	}

	if current {
		s += truncate(maxLen, text)
	} else {
		s += cutStyle(truncate(maxLen, text))
	}

	return s
}

// The preceding, current and following cuts.
func cutsWindow() (w []string) {
	w = []string{cutWinStyle(strings.Repeat("—", ex))}

	if currentCut > 1 {
		text, ts := ps.GetCut(currentCut - 1)
		w = append(w, drawCut(false, text, ts))
	}

	text, ts := ps.GetCut(currentCut)
	w = append(w, drawCut(true, text, ts))

	if currentCut < ps.Cuts() {
		text, ts := ps.GetCut(currentCut + 1)
		w = append(w, drawCut(false, text, ts))
	}

	return w
}
