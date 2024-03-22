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
	if sectionChars(cursor[Sectn]) == 0 {
		return
	}

	p := doc.GetText(cursor[Sectn], cursor[Para])
	i := len(p) - 1
	if p[i] == ' ' {
		doc.SetText(cursor[Sectn], cursor[Para], p[:i])
	}

	cursor[Para]++
	cursor[Char] = 0
	doc.CreateParagraph(cursor[Sectn], cursor[Para])
	initialCap = true
	ocursor = counts{}
	scope = Para
	indexPara(cursor[Sectn])
}

func appendSectnBreak() {
	if sectionChars(cursor[Sectn]) == 0 {
		return
	}

	sn := cursor[Sectn]
	pn := cursor[Para]
	if paragraphChars(sn, pn) == 0 {
		doc.DeleteParagraph(sn, pn)
		sections[sn-1].p = slices.Delete[[]ipara](sections[sn-1].p, pn-1, pn)
		buffer = slices.Delete(buffer, curs_buff, curs_buff+1)
		curs_buff = max(0, curs_buff-1)
	}
	cursor = counts{0, 0, 0, 1, sn + 1}
	doc.CreateSection(cursor[Sectn])
	initialCap = true
	ocursor = counts{}
	scope = Sectn
	indexSectn()
}

// Append runes to the document.
// TODO Implement insertion instead
func AppendRunes(runes []rune) {
	if initialCap && unicode.IsLower(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}

	t := string(runes)
	doc.AppendText(cursor[Sectn], cursor[Para], t)
	cursor[Char] += uniseg.GraphemeClusterCount(t)
	initialCap = false
	ocursor = counts{}
	scope = Char
}

func DecScope() {
	if scope == Char {
		scope = Sectn
	} else {
		scope--
	}

	if scope < Sent {
		initialCap = false
	}
}

func IncScope() {
	if scope == Sectn {
		scope = Char
		initialCap = false
	} else {
		scope++
	}
}

func Space() {
	switch scope {
	case Sectn:
		return
	case Para:
		appendSectnBreak()
		return
	case Sent:
		appendParaBreak()
		return
	}

	p := []byte(doc.GetText(cursor[Sectn], cursor[Para]))
	i := len(p) - 1
	if i < 0 {
		return
	}

	lb := p[i]
	if scope == Char {
		if lb != ' ' {
			AppendRunes([]rune(" "))
		}

		lr, _ := utf8.DecodeLastRune(p)
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
		lr, _ := utf8.DecodeLastRune(p[:i])
		if unicode.In(lr, unicode.L, unicode.N) { // Alphanumeric
			p = slices.Insert(p, i, '.')
			doc.SetText(cursor[Sectn], cursor[Para], string(p))
			cursor[Char]++
			ocursor = counts{}
		}
	}

	scope = Sent
}

func Enter() {
	if scope == Sectn {
		return
	}

	if scope <= Sent {
		appendParaBreak()
	} else { // scope == Para
		appendSectnBreak()
	}
}
