package document

import (
	"slices"
)

/*
Implements storage and retrieval of the document contents.

The document is UTF-8 encoded text divided into sections and paragraphs.
Sections are numbered starting from 1.  There is no section 0.
Paragraphs are numbered starting from 1 relative to the containing section.
*/

var document = [][][]byte{{{}}} // Initial document has an empty paragraph

// Append to a paragraph
func AppendText(sn, pn int, t string) {
	document[sn-1][pn-1] = append(document[sn-1][pn-1], t...)
}

// Create a new paragraph
func CreateParagraph(sn, pn int) {
	document[sn-1] = slices.Insert[[][]byte](document[sn-1], pn-1, []byte{})
}

// Create a new section
func CreateSection(sn int) {
	document = slices.Insert[[][][]byte](document, sn-1, [][]byte{{}}) // Include empty paragraph
}

// Delete a paragraph
func DeleteParagraph(sn, pn int) {
	document[sn-1] = slices.Delete(document[sn-1], pn-1, pn)
}

// Delete a section
func DeleteSection(sn int) { document = slices.Delete(document, sn-1, sn) }

// Get the text of a paragraph
func GetText(sn, pn int) string { return string(document[sn-1][pn-1]) }

// Move paragraphs starting at (osn, opn) to (nsn, npn)
func MoveParagraphs(osn, opn, nsn, npn int) {
	document[nsn-1] = slices.Insert(document[nsn-1], npn-1, document[osn-1][opn-1:]...)
	document[osn-1] = slices.Delete(document[osn-1], opn-1, len(document[osn-1]))
	if len(document[osn-1]) == 0 {
		document[osn-1] = [][]byte{{}}
	}
}

// Number of paragraphs in a section
func Paragraphs(sn int) int { return len(document[sn-1]) }

// Number of sections in the document
func Sections() int { return len(document) }

// Set the text of a paragraph
func SetText(sn, pn int, t string) { document[sn-1][pn-1] = []byte(t) }
