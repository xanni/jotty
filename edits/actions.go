package edits

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	doc "git.sericyb.com.au/jotty/document"
	"github.com/rivo/uniseg"
)

// Implements miscellaneous actions

func appendParaBreak() {
	sn := cursor[Sectn]
	if sectionChars(sn) == 0 {
		return
	}

	// TODO Split an existing paragraph
	pn := cursor[Para]
	p := doc.GetText(sn, pn)
	i := len(p) - 1
	if i >= 0 && p[i] == ' ' {
		doc.SetText(sn, pn, p[:i])
	}

	pn++
	cursor = counts{0, 0, 0, pn, sn}
	doc.CreateParagraph(sn, pn)
	initialCap = true
	ocursor = counts{}
	scope = Para
	indexPara(sn)
	buffer = slices.Insert(buffer, curs_buff+1, para{sn, pn, nil})
	for i := curs_buff + 2; i < len(buffer) && buffer[i].sn == sn; i++ {
		buffer[i].pn++
	}
}

func appendSectnBreak() {
	sn := cursor[Sectn]
	if sectionChars(sn) == 0 {
		return
	}

	// TODO Split an existing paragraph
	pn := cursor[Para]
	if paragraphChars(sn, pn) == 0 {
		doc.DeleteParagraph(sn, pn)
		sections[sn-1].p = slices.Delete[[]ipara](sections[sn-1].p, pn-1, pn)
		buffer = slices.Delete(buffer, curs_buff, curs_buff+1)
		if curs_buff > 0 {
			curs_buff--
			t := &buffer[curs_buff].text
			(*t)[len(*t)-1] = strings.Repeat("─", ex)
		}
	}

	sn++
	doc.CreateSection(sn)
	indexSectn()
	if doc.Paragraphs(sn-1) >= pn {
		doc.DeleteParagraph(sn, 1)
		doc.MoveParagraphs(sn-1, pn, sn, 1)
		sections[sn-1].p = nil
		moveParas(sn-1, pn, sn, 1)
	}

	for i := curs_buff + 1; i < len(buffer); i++ {
		buffer[i].sn++
		if buffer[i].sn == sn {
			buffer[i].pn -= (pn - 1)
		}
	}

	cursor = counts{0, 0, 0, 1, sn}
	initialCap = true
	ocursor = counts{}
	scope = Sectn
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
