package permascroll

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
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

// Regular expressions for parsing permascroll entries.
var (
	diRx = regexp.MustCompile(`(\d+),(\d+):(.+)\n`) // Delete and Insert arguments
	msRx = regexp.MustCompile(`(\d+),(\d+)\n`)      // Merge and Split arguments
	opRx = regexp.MustCompile(`(\d*)([DIMRS])`)     // Operation prefix
)

type version struct{ source, parent, lastChild int }

var (
	current     int       // Current version in the history
	deleting    int       // Number of bytes to delete starting from offset
	document    []string  // Text of each paragraph
	history     []version // Document history
	offset      int       // Current offset in the paragraph
	paragraph   int       // Current paragraph number
	pending     string    // Text not yet written to the permascroll
	permascroll []byte    // Serialised history of all document versions
)

var errRange = errors.New("out of range")

func init() { Init() }

// Persist a new version to the history.
func newVersion(s string) {
	parent := current
	current = len(history)
	history = append(history, version{len(permascroll), parent, 0})
	history[parent].lastChild = current

	delta := (current - parent) - 1
	if delta > 0 {
		permascroll = append(permascroll, []byte(strconv.Itoa(delta))...)
	}

	permascroll = append(permascroll, []byte(s+"\n")...)
}

// Initialise permascroll.
func Init() {
	current, deleting, pending, paragraph, offset = 0, 0, "", 1, 0
	document = []string{""} // Start with a single empty paragraph
	history = []version{{}} // Start with a single empty version
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

func docDelete(size int) {
	p := document[paragraph-1]
	document[paragraph-1] = p[:offset] + p[offset+size:]
}

func docInsert(text string) {
	p := document[paragraph-1]
	document[paragraph-1] = p[:offset] + text + p[offset:]
	offset += len(text)
}

func docMerge() {
	offset = len(document[paragraph-1])
	document[paragraph-1] += document[paragraph]
	document = slices.Delete(document, paragraph, paragraph+1)
}

func docSplit() {
	p := document[paragraph-1]
	document = slices.Insert(document, paragraph, p[offset:])
	document[paragraph-1] = p[:offset]
	paragraph++
	offset = 0
}

// Write pending insertion or deletion to permascroll.
func Flush() {
	if deleting > 0 {
		p := document[paragraph-1]
		t := p[offset : offset+deleting]
		newVersion(fmt.Sprintf("D%d,%d:%s", paragraph, offset, t))
		docDelete(deleting)
		deleting = 0
	} else if len(pending) > 0 {
		newVersion(fmt.Sprintf("I%d,%d:%s", paragraph, offset, pending))
		docInsert(pending)
		pending = ""
	}
}

// Get the current position in the document.
func GetPos() (int, int) { return paragraph, offset }

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
		paragraph = pn
		docMerge()
		newVersion(fmt.Sprintf("M%d,%d", pn, offset))
	}
}

// Number of paragraphs in the document.
func Paragraphs() int { return len(document) }

// Split a paragraph at a specified position.
func SplitParagraph(pn, pos int) {
	validatePos(pn, pos)

	Flush()
	paragraph, offset = pn, pos
	newVersion(fmt.Sprintf("S%d,%d", pn, offset))
	docSplit()
}

// Parse an operation from the permascroll.
func parseOperation(source int) (delta int, op byte, text string) {
	match := opRx.FindSubmatch(permascroll[source:])
	source += len(match[0])

	if len(match[1]) > 0 {
		delta, _ = strconv.Atoi(string(match[1]))
	}

	op = match[2][0]
	if op == 'D' || op == 'I' {
		match = diRx.FindSubmatch(permascroll[source:])
		text = string(match[3])
	} else {
		match = msRx.FindSubmatch(permascroll[source:])
	}

	paragraph, _ = strconv.Atoi(string(match[1]))
	offset, _ = strconv.Atoi(string(match[2]))

	return
}

// Redo the last undone operation, if any.
func Redo() byte {
	child := history[current].lastChild
	if child == 0 || deleting > 0 || len(pending) > 0 {
		return 0
	}

	current = child
	_, op, text := parseOperation(history[current].source)

	switch op {
	case 'D':
		docDelete(len(text))
	case 'I':
		docInsert(text)
	case 'M':
		docMerge()
	default: // 'R' not implemented yet, must be 'S'
		docSplit()
	}

	return op
}

// Undo the immediately preceding operation, if any.
func Undo() byte {
	if current == 0 {
		return 0
	}

	Flush()
	_, op, text := parseOperation(history[current].source)
	current = history[current].parent

	switch op {
	case 'D':
		docInsert(text)
	case 'I':
		docDelete(len(text))
	case 'M':
		docSplit()
	default: // 'R' not implemented yet, must be 'S'
		docMerge()
	}

	return op
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
