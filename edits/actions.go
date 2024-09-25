package edits

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	ps "git.sericyb.com.au/jotty/permascroll"
	"github.com/rivo/uniseg"
)

// Implements miscellaneous actions

func insertParaBreak() {
	pn := cursor[Para]
	t := before.String()
	i := len(t)
	for i > 0 && t[i-1] == ' ' {
		i-- // Trim spaces before cursor
	}

	if i < len(t) {
		ps.DeleteText(pn, i, len(t))
	}
	ps.SplitParagraph(pn, i)

	pn++
	cursor = counts{0, 0, 0, pn}
	cache = slices.Insert[[]para](cache, pn-1, para{})
	initialCap = true
	ocursor = counts{}
	scope = Para
}

func insert(t string) {
	if len(mark) > 0 {
		cutPrimary()
	} else {
		updateSelections()
	}

	if cursor[Char] >= cache[cursor[Para]-1].chars {
		ps.AppendText(cursor[Para], t)
	} else {
		ps.InsertText(cursor[Para], before.Len(), t)
	}
	cursor[Char] += uniseg.GraphemeClusterCount(t)
	initialCap = false
	ocursor = counts{}
	scope = Char
}

// Insert contents of cut buffer into the document.
func InsertCut() {
	cutBuffer := ps.GetCut()
	if len(cutBuffer) > 0 {
		insert(cutBuffer)
	}
}

// Insert runes into the document.
func InsertRunes(runes []rune) {
	if initialCap && unicode.IsLower(runes[0]) {
		runes[0] = unicode.ToUpper(runes[0])
	}

	insert(string(runes))
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

	updateSelections()
}

func IncScope() {
	if scope == Para {
		scope = Char
		initialCap = false
	} else {
		scope++
	}

	updateSelections()
}

func ClearMarks() { mark = nil; updateSelections() }

func Mark() {
	// Remove existing marks?
	if markPara != cursor[Para] {
		mark = nil

		if markPara > 0 {
			updateSelections()
			drawPara(markPara)
		}
		markPara = cursor[Para]
	} else {
		for i, m := range mark {
			if m == cursor[Char] {
				mark = slices.Delete(mark, i, i+1)
				updateSelections()

				return
			}
		}
	}

	if len(mark) > 3 {
		mark = slices.Delete(mark, 0, 1)
	}
	mark = append(mark, cursor[Char])
	updateSelections()
}

func exchange() {
	if prevSelected {
		ps.ExchangeParagraphs(cursor[Para])
		primary.cend, secondary.cend = cache[cursor[Para]-2].chars, cache[cursor[Para]-1].chars
	} else {
		ps.ExchangeText(cursor[Para], primary.obegin, primary.oend, secondary.obegin, secondary.oend)
		primarySize, secondarySize := primary.cend-primary.cbegin, secondary.cend-secondary.cbegin
		if secondary.cbegin < primary.cbegin {
			end := primary.cbegin + primarySize
			primary.cbegin, primary.cend, secondary.cend = end-secondarySize, end, secondary.cbegin+primarySize
		} else {
			end := secondary.cbegin + secondarySize
			primary.cend, secondary.cbegin, secondary.cend = primary.cbegin+secondarySize, end-primarySize, end
		}
	}
}

func Space() {
	if prevSelected || secondary.cend > secondary.cbegin {
		exchange()
		mark = nil
		refresh()

		return
	}

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
			ps.InsertText(cursor[Para], i, ".")
			cursor[Char]++
			ocursor = counts{}
		}
	}
}

func Enter() {
	if len(mark) > 0 {
		cut()
	} else {
		updateSelections()
		insertParaBreak()
	}
}

func mergePrevPara() {
	pn := cursor[Para]
	if pn == 1 {
		return
	}

	clearCache(pn)
	cache = slices.Delete(cache, pn-1, pn)
	pn--
	cursPara, cursor[Para] = pn, pn
	t := ps.GetText(pn)
	cursor[Char] = uniseg.GraphemeClusterCount(t)
	ps.MergeParagraph(pn)

	if after.Len() > 0 && len(t) > 0 && t[len(t)-1] != ' ' {
		ps.InsertText(pn, len(t), " ")
	}
}

func mergeNextPara() {
	pn := cursor[Para]
	if pn >= ps.Paragraphs() {
		return
	}

	clearCache(pn + 1)
	cache = slices.Delete(cache, pn, pn+1)
	t, s := ps.GetText(pn), ps.GetSize(pn+1)
	ps.MergeParagraph(pn)

	if s > 0 && len(t) > 0 && t[len(t)-1] != ' ' {
		ps.InsertText(pn, len(t), " ")
	}
}

// Extract the first n characters in s.
func getChars(n int, s string) string {
	var r strings.Builder // Result
	state := -1

	for range n {
		var t string
		t, s, _, state = uniseg.FirstGraphemeClusterInString(s, state)
		r.WriteString(t)
	}

	return r.String()
}

// Extract the first n words in s.
func getWords(n int, s string) string {
	var r strings.Builder // Result
	state := -1

	for range n {
		var t string
		t, s, state = uniseg.FirstWordInString(s, state)
		for len(s) > 0 && !isAlphanumeric([]byte(s)) {
			r.WriteString(t)
			t, s, state = uniseg.FirstWordInString(s, state)
		}
		r.WriteString(t)
	}

	return r.String()
}

// Extract the first n sentences in s.
func getSents(n int, s string) string {
	var r strings.Builder // Result
	state := -1

	for range n {
		var t string
		t, s, state = uniseg.FirstSentenceInString(s, state)
		r.WriteString(t)
	}

	return r.String()
}

func Backspace() {
	ClearMarks()

	if cursor[Char] == 0 {
		mergePrevPara()

		return
	}

	if scope == Para {
		ps.DeleteText(cursor[Para], 0, before.Len())
		cursor[Char] = 0
		initialCap = true

		return
	}

	b := before.String()
	n := cursor[scope] - 1 // Number of scope units to keep

	var s string
	switch scope {
	case Char:
		s = getChars(n, b)

	case Word:
		s = getWords(n, b)

	default: // scope == Sent by exclusion
		s = getSents(n, b)
		initialCap = true
	}

	ps.DeleteText(cursor[Para], len(s), before.Len())
	cursor[Char] = uniseg.GraphemeClusterCount(s)
}

// Get the byte offset and size of the scope unit after the cursor.
func scopeSpan() (int, int) {
	var t string

	switch scope {
	case Char:
		_, t, _, _ = uniseg.FirstGraphemeClusterInString(after.String(), -1)
	case Word:
		t = strings.TrimLeft(after.String(), " ")
		_, t, _ = uniseg.FirstWordInString(t, -1)
		t = strings.TrimLeft(t, " ")
	case Sent:
		_, t, _ = uniseg.FirstSentenceInString(after.String(), -1)
	default: // scope == Para by exclusion; keep t as empty string
	}

	return before.Len(), before.Len() + after.Len() - len(t)
}

func Copy() {
	if len(mark) > 0 && primary.oend > primary.obegin {
		ps.CopyText(markPara, primary.obegin, primary.oend)

		return
	}

	updateSelections()
	pos, end := scopeSpan()
	if end > pos {
		ps.CopyText(cursor[Para], pos, end)
	}
}

func cutPrimary() {
	ps.CutText(markPara, primary.obegin, primary.oend)
	cursor = counts{Char: primary.cbegin, Para: markPara}

	if cursPara != markPara {
		drawPara(cursPara) // Erase the old cursor position
	}

	ClearMarks()
	drawPara(markPara)
}

func cut() {
	if len(mark) > 2 {
		ps.DeleteText(markPara, secondary.obegin, secondary.oend)
	}

	cutPrimary()
}

func Delete() {
	if len(mark) > 0 {
		cut()

		return
	}
	updateSelections()

	if cursor[Char] == cache[cursor[Para]-1].chars {
		mergeNextPara()

		return
	}

	pos, end := scopeSpan()
	ps.DeleteText(cursor[Para], pos, end)
}

func Export(path string) {
	var err error
	if len(mark) > 0 {
		err = ps.ExportText(path, markPara, primary.obegin, primary.oend)
	} else {
		updateSelections()
		err = ps.ExportText(path, 0, 0, 0)
	}

	if err != nil {
		SetMode(Error, err.Error())
	}
}

func refresh() {
	cache = nil
	total = counts{0, 0, 0, 1}
	pn, pos := ps.GetPos()
	cursor = counts{uniseg.GraphemeClusterCount(ps.GetText(pn)[:pos]), 0, 0, pn}
}

func Undo() {
	op := ps.Undo()
	if op > 0 {
		refresh()
		ClearMarks()
	}
}

func Redo() {
	op := ps.Redo()
	if op > 0 {
		refresh()
		ClearMarks()
	}
}
