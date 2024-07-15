package edits

import (
	"slices"
	"unicode"
	"unicode/utf8"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/rivo/uniseg"
)

// Implements miscellaneous actions

func insertParaBreak() {
	pn := cursor[Para]
	t := before.String()
	i := len(t) - 1
	if i >= 0 && t[i] == ' ' {
		doc.SetText(pn, t[:i])
	} else {
		doc.SetText(pn, t)
	}

	pn++
	cursor = counts{0, 0, 0, pn}
	doc.CreateParagraph(pn)
	doc.SetText(pn, after.String())
	cache = slices.Insert[[]para](cache, pn-1, para{})
	initialCap = true
	ocursor = counts{}
	scope = Para
}

// Insert runes into the document
func InsertRunes(runes []rune) {
	if initialCap && unicode.IsLower(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}

	t := string(runes)
	if cursor[Char] >= cache[cursor[Para]-1].chars {
		doc.AppendText(cursor[Para], t)
	} else {
		doc.SetText(cursor[Para], before.String()+t+after.String())
	}
	cursor[Char] += uniseg.GraphemeClusterCount(t)
	initialCap = false
	ocursor = counts{}
	scope = Char
}

func DecScope() {
	if scope == Char {
		scope = Para
	} else {
		scope--
	}

	if scope < Sent {
		initialCap = false
	}
}

func IncScope() {
	if scope == Para {
		scope = Char
		initialCap = false
	} else {
		scope++
	}
}

func Space() {
	if scope >= Sent {
		insertParaBreak()
		return
	}

	t := []byte(before.String())
	i := len(t) - 1
	if i < 0 {
		return
	}

	oscope := scope // Original scope before InsertRunes, which sets it to Char
	lb := t[i]      // Last byte before the cursor: space is always a single byte
	if lb != ' ' {
		InsertRunes([]rune(" "))
	}
	scope = Sent

	if oscope == Char {
		lr, _ := utf8.DecodeLastRune(t)
		if unicode.Is(unicode.Sentence_Terminal, lr) {
			initialCap = true
		} else {
			initialCap = false
			scope = Word
		}

		return
	}

	// oscope == Word by elimination
	if lb == ' ' {
		initialCap = true
		lr, _ := utf8.DecodeLastRune(t[:i])       // Last rune before the space
		if unicode.In(lr, unicode.L, unicode.N) { // Alphanumeric
			t = slices.Insert(t, i, '.')
			doc.SetText(cursor[Para], string(t)+after.String())
			cursor[Char]++
			ocursor = counts{}
		}
	}
}

func Enter() {
	insertParaBreak()
}
