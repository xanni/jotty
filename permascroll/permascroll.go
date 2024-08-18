package permascroll

import (
	"errors"
	"fmt"
	"slices"
)

/*
Implements storage and retrieval of the document contents with a full version
history inspired by Ted Nelson's Xanadu designs.

The document is UTF-8 encoded text divided into paragraphs. Paragraphs are
numbered starting from 1.  There is no paragraph 0.

As the textual contents are divided into paragraphs, newlines never appear in
the primedia and are always and only permascroll record separators.  All
operations contain sufficient information to ensure they are reversible.

This package coalesces adjacent and overlapping text insertions and deletions.
*/

const magic = "JottyV0\n"

var (
	deleting    int      // Number of bytes to delete starting from offset
	document    []string // Text of each paragraph
	offset      int      // Current offset in the paragraph
	paragraph   int      // Current paragraph number
	pending     string   // Text not yet written to the permascroll
	permascroll []byte   // Serialised history of all document versions
)

var errRange = errors.New("out of range")

func init() { Init() }

// Initialise permascroll.
func Init() {
	deleting, pending, paragraph, offset = 0, "", 1, 0
	document = []string{""} // Start with a single empty paragraph
	permascroll = []byte(magic)
}

// Append text to a paragraph.
func AppendText(pn int, text string) { InsertText(pn, GetSize(pn), text) }

// Delete text from a paragraph between pos and end.
func DeleteText(pn, pos, end int) {
	validateRange(pn, pos, end)

	dEnd := offset + deleting
	pEnd := offset + len(pending)
	switch {
	case paragraph != pn || end < offset || pos > max(dEnd, pEnd) || (pos < offset && len(pending) > 0):
		Flush()
		deleting, paragraph, offset = end-pos, pn, pos
	case len(pending) == 0:
		if pos < offset {
			offset = pos
		}
		deleting += end - pos
	default:
		var s string
		if end < pEnd {
			s = pending[end-offset:]
		}
		pending = pending[:pos-offset] + s
		if end > pEnd {
			Flush()
			offset, deleting = pos, end-pEnd
		}
	}
}

// Write pending insertion or deletion to permascroll.
func Flush() {
	if deleting > 0 {
		p := document[paragraph-1]
		t := p[offset : offset+deleting]
		permascroll = append(permascroll, []byte(fmt.Sprintf("D%d,%d:%s\n", paragraph, offset, t))...)
		document[paragraph-1] = p[:offset] + p[offset+deleting:]
		deleting = 0
	} else if len(pending) > 0 {
		permascroll = append(permascroll, []byte(fmt.Sprintf("I%d,%d:%s\n", paragraph, offset, pending))...)
		p := document[paragraph-1]
		document[paragraph-1] = p[:offset] + pending + p[offset:]
		offset += len(pending)
		pending = ""
	}
}

// Get the size of a paragraph.
func GetSize(pn int) (size int) {
	validatePn(pn)

	size = len(document[pn-1])

	if pn == paragraph {
		size += len(pending) - deleting
	}

	return size
}

// Get the text of a paragraph.
func GetText(pn int) (t string) {
	validatePn(pn)

	t = document[pn-1]

	if pn == paragraph {
		if deleting > 0 {
			t = t[:offset] + t[offset+deleting:]
		} else if len(pending) > 0 {
			t = t[:offset] + pending + t[offset:]
		}
	}

	return t
}

// Insert text into a paragraph at pos.
func InsertText(pn int, pos int, text string) {
	validatePos(pn, pos)

	if pn == paragraph && pos >= offset && pos <= offset+len(pending) {
		pending = pending[:pos-offset] + text + pending[pos-offset:]
	} else {
		Flush()
		pending, paragraph, offset = text, pn, pos
	}
}

// Merge two paragraphs.
func MergeParagraph(pn int) {
	validatePn(pn)

	if pn < len(document) {
		Flush()
		paragraph, offset = pn, len(document[pn-1])
		permascroll = append(permascroll, []byte(fmt.Sprintf("M%d,%d\n", pn, offset))...)
		document[pn-1] += document[pn]
		document = slices.Delete(document, pn, pn+1)
	}
}

// Number of paragraphs in the document.
func Paragraphs() int { return len(document) }

// Split a paragraph at a specified position.
func SplitParagraph(pn, pos int) {
	validatePos(pn, pos)

	Flush()
	paragraph, offset = pn, pos
	permascroll = append(permascroll, []byte(fmt.Sprintf("S%d,%d\n", pn, offset))...)
	p := document[pn-1]
	document = slices.Insert(document, pn, p[pos:])
	document[pn-1] = p[:pos]
}

/*
NOTE that this package violates the Go convention that panics should not cross
package boundaries, because any invalid arguments indicate errors in the calling
code rather than anything that can be resolved by error handling.  If errors
were returned instead, every caller would still have to panic after checking for
them.
*/

func validatePn(pn int) {
	if pn < 1 || pn > len(document) {
		panic(fmt.Errorf("paragraph '%d' %w", pn, errRange))
	}
}

func validatePos(pn, pos int) {
	validatePn(pn)
	if pos < 0 || pos > len(document[pn-1])+len(pending) {
		panic(fmt.Errorf("pos '%d,%d' %w", pn, pos, errRange))
	}
}

func validateRange(pn, pos, end int) {
	validatePos(pn, pos)
	if end <= pos || end > len(document[pn-1])+len(pending)+1 {
		panic(fmt.Errorf("end '%d,%d-%d' %w", pn, pos, end, errRange))
	}
}
