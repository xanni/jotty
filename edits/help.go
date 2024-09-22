package edits

import (
	"slices"
	"strings"

	"github.com/rivo/uniseg"
)

var HelpText []byte

func dropParagraphs(w []string) []string {
	i := len(w) - ey
	for i < len(w) && len(w[i]) > 0 {
		i++
	}

	if i < len(w) {
		return w[i+1:]
	}

	return w[:ey-1]
}

// The help window.
func helpWindow() (w []string) {
	w = strings.Split(string(HelpText), "\n")
	if len(w[len(w)-1]) == 0 { // Trim final blank line
		w = slices.Delete(w, len(w)-1, len(w))
	}

	var longest int
	for _, l := range w {
		longest = max(longest, uniseg.StringWidth(l))
	}

	if longest > ex {
		w = rewrap(w)
	} else {
		padding := strings.Repeat(" ", (ex-longest)/2)
		for i, l := range w {
			if len(l) > 0 {
				w[i] = padding + l
			}
		}
	}

	if len(w) > ey-1 {
		w = dropParagraphs(w)
	}

	w = append(w, strings.Repeat("—", ex))

	for i, l := range w {
		w[i] = helpStyle(l)
	}

	return w
}

func rewrap(o []string) (w []string) {
	var p string // Pending line

	for _, l := range o {
		if len(l) == 0 { // Paragraph break
			w, p = append(w, p, ""), ""

			continue
		}

		if len(p) > 0 {
			p += " "
		}
		p += l
		for size := uniseg.StringWidth(p); size >= ex; size = uniseg.StringWidth(p) {
			s := size // Split point
			for s >= ex || (s > 0 && p[s-1] == '"') {
				s = strings.LastIndexByte(p[:s-1], ' ')
			}

			if s < 0 {
				s = ex - 1
				w = append(w, p[:s]+"-")
				p = p[s:]
			} else {
				w = append(w, p[:s])
				p = p[s+1:]
			}
		}
	}

	w = append(w, p)

	return w
}
