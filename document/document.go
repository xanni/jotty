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

// Append to a paragraph
func AppendText(pn int, t string) { document[pn-1] = append(document[pn-1], t...) }

// Create a new paragraph
func CreateParagraph(pn int) {
	document = slices.Insert[[][]byte](document, pn-1, []byte{})
}

// Delete a paragraph
func DeleteParagraph(pn int) { document = slices.Delete(document, pn-1, pn) }

// Get the text of a paragraph
func GetText(pn int) string { return string(document[pn-1]) }

// Number of paragraphs in the document
func Paragraphs() int { return len(document) }

// Set the text of a paragraph
func SetText(pn int, t string) { document[pn-1] = []byte(t) }
