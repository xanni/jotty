package edits

import (
	"slices"
	"unicode"
	"unicode/utf8"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/rivo/uniseg"
)

// Implements miscellaneous actions

func appendParaBreak() {
	// TODO Split an existing paragraph
	pn := cursor[Para]
	t := doc.GetText(pn)
	i := len(t) - 1
	if i >= 0 && t[i] == ' ' {
		doc.SetText(pn, t[:i])
	}

	pn++
	cursor = counts{0, 0, 0, pn}
	doc.CreateParagraph(pn)
	cache = slices.Insert[[]para](cache, pn-1, para{})
	initialCap = true
	ocursor = counts{}
	scope = Para
	indexPara()
}

// Append runes to the document.
// TODO Implement insertion instead
func AppendRunes(runes []rune) {
	if initialCap && unicode.IsLower(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}

	t := string(runes)
	doc.AppendText(cursor[Para], t)
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
		appendParaBreak()
		return
	}

	t := []byte(doc.GetText(cursor[Para]))
	i := len(t) - 1
	if i < 0 {
		return
	}

	lb := t[i]
	if scope == Char {
		if lb != ' ' {
			AppendRunes([]rune(" "))
		}

		lr, _ := utf8.DecodeLastRune(t)
		if unicode.Is(unicode.Sentence_Terminal, lr) {
			initialCap = true
			scope = Sent
		} else {
			initialCap = false
			scope = Word
		}

		return
	}

	// scope == Word by elimination
	if lb != ' ' {
		AppendRunes([]rune(" "))
	} else {
		initialCap = true
		lr, _ := utf8.DecodeLastRune(t[:i])
		if unicode.In(lr, unicode.L, unicode.N) { // Alphanumeric
			t = slices.Insert(t, i, '.')
			doc.SetText(cursor[Para], string(t))
			cursor[Char]++
			ocursor = counts{}
		}
	}

	scope = Sent
}

func Enter() {
	appendParaBreak()
}
