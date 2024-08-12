package document

import (
	"slices"
)

/*
Implements storage and retrieval of the document contents.

The document is UTF-8 encoded text divided into paragraphs.
Paragraphs are numbered starting from 1.  There is no paragraph 0.
*/

var document = [][]byte{{}} // Initial document has an empty paragraph

// Initialise permascroll.
func Init() {
	document = [][]byte{{}}
}

// Append to a paragraph.
func AppendText(pn int, t string) { document[pn-1] = append(document[pn-1], t...) }

// Delete text from a paragraph between pos and end.
func DeleteText(pn, pos, end int) { document[pn-1] = slices.Delete(document[pn-1], pos, end) }

// Get the size of a paragraph.
func GetSize(pn int) int { return len(document[pn-1]) }

// Get the text of a paragraph.
func GetText(pn int) string { return string(document[pn-1]) }

// Insert text into a paragraph at pos.
func InsertText(pn int, pos int, t string) {
	document[pn-1] = slices.Insert(document[pn-1], pos, []byte(t)...)
}

// Merge two paragraphs.
func MergeParagraph(pn int) {
	if pn < len(document) {
		document[pn-1] = append(document[pn-1], document[pn]...)
		document = slices.Delete(document, pn, pn+1)
	}
}

// Number of paragraphs in the document.
func Paragraphs() int { return len(document) }

// Split a paragraph at a specified position.
func SplitParagraph(pn, pos int) {
	p := document[pn-1]
	document = slices.Insert(document, pn, p[pos:])
	document[pn-1] = p[:pos]
}
