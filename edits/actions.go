package edits

import (
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

// Implements miscellaneous actions

func appendParaBreak() {
	i := len(document) - 1
	if document[i] == ' ' {
		document[i] = '\n'
	} else {
		document = append(document, '\n')
		cursor[Char]++
	}
	scope = Para
}

func appendSectnBreak() {
	i := len(document) - 1
	if document[i] != '\n' {
		document = append(document, '\f')
	} else {
		document[i] = '\f'
		s := &sections[cursor[Sectn]-1]
		s.chars--
		s.bpara = s.bpara[:len(s.bpara)-1]
		s.cpara = s.cpara[:len(s.cpara)-1]
		s.csent = s.csent[:len(s.csent)-1]
		y := cursy - 2
		if y >= 0 {
			l := buffer[y]
			l.brk = Sectn
			y++
			advanceLine(&y, &l)
		}
	}

	sn := cursor[Sectn] + 1
	cursor = counts{Sectn: sn}
	scope = Sectn
	indexSectn(len(document))
}

// Append runes to the document.
// TODO implement insertion instead
func AppendRunes(runes []rune) {
	if initialCap && unicode.IsLower(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}

	document = append(document, []byte(string(runes))...)
	initialCap = false
	osectn = 0
	scope = Char

	if cursor[Sectn] != len(sections) {
		cursor[Sectn] = len(sections)
		cursor[Char] = uniseg.GraphemeClusterCount(string(document[sections[cursor[Sectn]-1].bsectn:]))
	} else {
		cursor[Char] = buffer[cursy].beg_c + uniseg.GraphemeClusterCount(string(document[buffer[cursy].beg_b:]))
	}
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
	i := len(document) - 1
	if scope == Sectn || i < 0 {
		return
	}

	lb := document[i]
	switch scope {
	case Char:
		lr, _ := utf8.DecodeLastRune(document)
		if lb != ' ' && lb != '\n' && lb != '\f' {
			AppendRunes([]rune(" "))
		}
		if unicode.Is(unicode.Sentence_Terminal, lr) {
			scope = Sent
		} else {
			scope = Word
		}
	case Word:
		if lb != ' ' && lb != '\n' && lb != '\f' {
			AppendRunes([]rune(" "))
		}
		if lb == ' ' {
			lr, _ := utf8.DecodeLastRune(document[:i])
			if unicode.In(lr, unicode.L, unicode.N) { // alphanumeric
				document[i] = '.'
				AppendRunes([]rune(" "))
			}
		}
		scope = Sent
	case Sent:
		appendParaBreak()
	default: // Para because Sectn has already been excluded above
		appendSectnBreak()
	}

	initialCap = scope >= Sent
	osectn = 0
}

func Enter() {
	if scope == Sectn || len(document) == 0 {
		return
	}

	if scope <= Sent {
		appendParaBreak()
	} else { // scope == Para
		appendSectnBreak()
	}

	initialCap = true
	osectn = 0
}
