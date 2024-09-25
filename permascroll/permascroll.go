package permascroll

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/cespare/xxhash/v2"
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

It also maintains hashes of the document and cut buffer contents and uses them
to detect when a previous state is revisited, thus avoiding persisting the same
state to the permascroll again and instead updating the version history.
*/

const magic = "JottyV0\n"

// Regular expressions for parsing permascroll entries.
var (
	ccRx = regexp.MustCompile(`(\d+),(\d+)([+:])(.+)\n`)                // Copy and Cut arguments
	diRx = regexp.MustCompile(`(\d+),(\d+):(.+)\n`)                     // Delete and Insert arguments
	msRx = regexp.MustCompile(`(\d+),(\d+)\n`)                          // Merge and Split arguments
	opRx = regexp.MustCompile(`(\d*)([CDIMRSX])`)                       // Operation prefix
	reRx = regexp.MustCompile(`(\d+),(\d+):(.+)\t(.+)\n`)               // Replace arguments
	exRx = regexp.MustCompile(`(\d+)(?:,(\d+)\+(\d+)/(\d+)\+(\d+))?\n`) // Exchange arguments
)

type (
	span    struct{ begin, end int }
	version struct{ source, parent, lastChild int }
)

var (
	current     int            // Current version in the history
	cut         string         // Text cut from the document
	deleting    int            // Number of bytes to delete starting from offset
	docHash     []uint64       // Hash of each paragraph
	document    []string       // Text of each paragraph
	histHash    map[uint64]int // Map of hashes to version numbers
	history     []version      // Document history
	mutex       sync.Mutex     // Mutex to ensure safety of Flush()
	offset      int            // Current offset in the paragraph
	paragraph   int            // Current paragraph number
	pending     string         // Text not yet written to the permascroll
	permascroll []byte         // Serialised history of all document versions
)

var (
	errParse = errors.New("parse failed")
	errRange = errors.New("out of range")
)

func init() { Init() }

// Compute the hash of the current version of the document.
func hashDocument() uint64 {
	size := len(docHash) * 8                                          // Each uint64 is 8 bytes
	buf := (*[1 << 32]byte)(unsafe.Pointer(&docHash[0]))[0:size:size] // Get the underlying docHash array

	hash := xxhash.New()
	_, _ = hash.Write(buf)
	_, _ = hash.WriteString(cut)

	return hash.Sum64()
}

func updateHash(pn int) {
	docHash[pn-1] = xxhash.Sum64String(document[pn-1])
}

// Initialise permascroll.
func Init() {
	current, cut, deleting, pending, paragraph, offset = 0, "", 0, "", 1, 0
	document = []string{""} // Start with a single empty paragraph
	docHash = []uint64{xxhash.Sum64String("")}
	history = []version{{}} // Start with a single empty version
	histHash = map[uint64]int{hashDocument(): 0}
	permascroll = []byte(magic)
}

// Append text to a paragraph.
func AppendText(pn int, text string) { InsertText(pn, GetSize(pn), text) }

// Copy text from a paragraph between pos and end.
func CopyText(pn, pos, end int) {
	validateSpan(pn, pos, end)

	Flush()
	cut = document[pn-1][pos:end]
	persist(fmt.Sprintf("C%d,%d+%d", pn, pos, end-pos))
}

// Cut text from a paragraph between pos and end.
func CutText(pn, pos, end int) {
	validateSpan(pn, pos, end)

	Flush()
	cut = document[pn-1][pos:end]
	paragraph, offset = pn, pos
	docDelete(end - pos)
	persist(fmt.Sprintf("C%d,%d:%s", pn, pos, cut))
}

// Delete text from a paragraph between pos and end.
func DeleteText(pn, pos, end int) {
	validateSpan(pn, pos, end)

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
	updateHash(paragraph)
}

func docExchange(first, second span) {
	if first.end == 0 { // Exchange paragraphs
		document[paragraph-1], document[paragraph-2] = document[paragraph-2], document[paragraph-1]
		updateHash(paragraph - 1)
	} else { // Exchange text ranges
		p := document[paragraph-1]
		var t strings.Builder
		t.WriteString(p[:first.begin])
		t.WriteString(p[second.begin:second.end])
		t.WriteString(p[first.end:second.begin])
		t.WriteString(p[first.begin:first.end])
		t.WriteString(p[second.end:])
		document[paragraph-1] = t.String()
	}
	updateHash(paragraph)
	offset = first.begin
}

func docInsert(text string) {
	p := document[paragraph-1]
	document[paragraph-1] = p[:offset] + text + p[offset:]
	updateHash(paragraph)
	offset += len(text)
}

func docReplace(size int, text string) {
	p := document[paragraph-1]
	document[paragraph-1] = p[:offset] + text + p[offset+size:]
	updateHash(paragraph)
	offset += len(text)
}

func docMerge() {
	offset = len(document[paragraph-1])
	document[paragraph-1] += document[paragraph]
	updateHash(paragraph)
	document = slices.Delete(document, paragraph, paragraph+1)
	docHash = slices.Delete(docHash, paragraph, paragraph+1)
}

func docSplit() {
	p := document[paragraph-1]
	document = slices.Insert(document, paragraph, p[offset:])
	docHash = slices.Insert(docHash, paragraph, 0)
	updateHash(paragraph + 1)
	document[paragraph-1] = p[:offset]
	updateHash(paragraph)
	paragraph++
	offset = 0
}

func docRedo(op operation) {
	paragraph, offset = op.pn, op.offset1
	switch op.code {
	case 'C':
		if op.size1 > 0 {
			cut = document[paragraph-1][offset : offset+op.size1]
		} else {
			cut = document[paragraph-1][offset : offset+len(op.text1)]
			docDelete(len(op.text1))
		}
	case 'D':
		docDelete(len(op.text1))
	case 'I':
		docInsert(op.text1)
	case 'M':
		docMerge()
	case 'R':
		docReplace(len(op.text1), op.text2)
	case 'S':
		docSplit()
	default: // 'X'
		docExchange(span{op.offset1, op.offset1 + op.size1}, span{op.offset2, op.offset2 + op.size2})
	}
}

func docUndo() byte {
	source := history[current].source
	current = history[current].parent
	_, op := parseOperation(&source)
	paragraph, offset = op.pn, op.offset1
	switch op.code {
	case 'C':
		if len(op.text1) > 0 {
			docInsert(op.text1)
		}
		cut = ""
	case 'D':
		docInsert(op.text1)
	case 'I':
		docDelete(len(op.text1))
	case 'M':
		docSplit()
	case 'R':
		docReplace(len(op.text2), op.text1)
		offset = op.offset1
	case 'S':
		docMerge()
	default: // 'X'
		begin := op.offset2 + op.size2 - op.size1
		docExchange(span{op.offset1, op.offset1 + op.size2}, span{begin, begin + op.size1})
	}

	return op.code
}

// Exchange two paragraphs.
func ExchangeParagraphs(pn int) {
	validatePn(pn)
	if pn < 2 {
		panic(fmt.Errorf("paragraph '%d' %w", pn, errRange))
	}

	Flush()
	paragraph = pn
	docExchange(span{}, span{})
	persist(fmt.Sprintf("X%d", pn))
}

// Exchange two text spans.
func ExchangeText(pn, b1, e1, b2, e2 int) {
	validateSpan(pn, b1, e1)
	validateSpan(pn, b2, e2)
	if b2 < b1 {
		b1, e1, b2, e2 = b2, e2, b1, e1
	}

	if b2 < e1 {
		panic(fmt.Errorf("overlap '%d-%d/%d-%d' %w", b1, e1, b2, e2, errRange))
	}

	Flush()
	paragraph = pn
	docExchange(span{b1, e1}, span{b2, e2})
	persist(fmt.Sprintf("X%d,%d+%d/%d+%d", pn, b1, e1-b1, b2, e2-b2))
}

// Write pending insertion or deletion to permascroll.
// Safe to use concurrently.
func Flush() {
	mutex.Lock()
	defer mutex.Unlock()
	if deleting > 0 {
		p := document[paragraph-1]
		t := p[offset : offset+deleting]
		docDelete(deleting)
		persist(fmt.Sprintf("D%d,%d:%s", paragraph, offset, t))
		deleting = 0
	} else if len(pending) > 0 {
		o := offset
		docInsert(pending)
		persist(fmt.Sprintf("I%d,%d:%s", paragraph, o, pending))
		pending = ""
	}
}

// Get the text cut from the document.
func GetCut() string { return cut }

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

	if pn == paragraph && deleting == 0 && pos >= offset && pos <= offset+len(pending) {
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
		persist(fmt.Sprintf("M%d,%d", pn, offset))
	}
}

// Add a new version to the history.
func newVersion(source int) int {
	parent := current
	h := hashDocument()
	if v, found := histHash[h]; found {
		current = v

		return -1
	}

	current = len(history)
	history = append(history, version{source, parent, 0})
	history[parent].lastChild, histHash[h] = current, current

	return (current - parent) - 1
}

// Number of paragraphs in the document.
func Paragraphs() int { return len(document) }

/*
NOTE that this package violates the Go convention that panics should not cross
package boundaries, because any invalid arguments indicate errors in the calling
code rather than anything that can be resolved by error handling.  If errors
were returned instead, every caller would still have to panic after checking for
them.
*/

// Parse the arguments of a copy or cut operation from the permascroll.
func parseCopyCut(source *int) (op operation, match [][]byte) {
	op.code = 'C'
	match = ccRx.FindSubmatch(permascroll[*source:])
	if match == nil {
		return op, match
	}

	if match[3][0] == ':' {
		op.text1 = string(match[4])
	} else {
		var err error
		if op.size1, err = strconv.Atoi(string(match[4])); err != nil {
			panic(fmt.Errorf("invalid size for 'C', %w", err))
		}
	}

	return op, match
}

// Parse the arguments of an exchange operation from the permascroll.
func parseExchange(source *int) (op operation, match [][]byte) {
	op.code = 'X'
	match = exRx.FindSubmatch(permascroll[*source:])
	if match == nil {
		return op, match
	}

	if len(match[2]) > 0 {
		op.offset1, _ = strconv.Atoi(string(match[2]))
		op.size1, _ = strconv.Atoi(string(match[3]))
		op.offset2, _ = strconv.Atoi(string(match[4]))
		op.size2, _ = strconv.Atoi(string(match[5]))
	}

	return op, match
}

type operation struct {
	code                               byte
	pn, offset1, size1, offset2, size2 int
	text1, text2                       string
}

// Parse an operation from the permascroll.
func parseOperation(source *int) (delta int, op operation) {
	match := opRx.FindSubmatch(permascroll[*source:])
	if match == nil {
		panic(fmt.Errorf("invalid operation %q, %w", permascroll[*source], errParse))
	}
	*source += len(match[0])

	if len(match[1]) > 0 {
		delta, _ = strconv.Atoi(string(match[1]))
	}

	op.code = match[2][0]
	switch op.code {
	case 'C':
		op, match = parseCopyCut(source)
	case 'D', 'I':
		if match = diRx.FindSubmatch(permascroll[*source:]); match != nil {
			op.text1 = string(match[3])
		}
	case 'R':
		if match = reRx.FindSubmatch(permascroll[*source:]); match != nil {
			op.text1 = string(match[3])
			op.text2 = string(match[4])
		}
	case 'M', 'S':
		match = msRx.FindSubmatch(permascroll[*source:])
	default: // 'X'
		op, match = parseExchange(source)
	}

	if match == nil {
		panic(fmt.Errorf("invalid arguments for %q, %w", op.code, errParse))
	}

	*source += len(match[0])
	op.pn, _ = strconv.Atoi(string(match[1]))

	if op.code != 'X' {
		op.offset1, _ = strconv.Atoi(string(match[2]))
	}

	return delta, op
}

// Parse the entire permascroll.
func parsePermascroll() {
	if len(permascroll) < len(magic) || !bytes.Equal(permascroll[:len(magic)], []byte(magic)) {
		panic(fmt.Errorf("invalid magic, %w", errParse))
	}

	source := len(magic)
	for source < len(permascroll) {
		opSource := source
		delta, op := parseOperation(&source)
		for range delta {
			docUndo()
		}
		docRedo(op)
		newVersion(opSource)
	}
}

func ReplaceText(pn, pos, end int, text string) {
	validateSpan(pn, pos, end)

	Flush()
	paragraph, offset = pn, pos
	d := document[paragraph-1][offset:end]
	docReplace(end-offset, text)
	persist(fmt.Sprintf("R%d,%d:%s\t%s", paragraph, pos, d, text))
}

// Redo the last undone operation, if any.
func Redo() (code byte) {
	child := history[current].lastChild
	if child > 0 && deleting == 0 && len(pending) == 0 {
		current = child
		source := history[current].source
		_, op := parseOperation(&source)
		docRedo(op)
		code = op.code
	}

	return code
}

// Split a paragraph at a specified position.
func SplitParagraph(pn, pos int) {
	validatePos(pn, pos)

	Flush()
	paragraph, offset = pn, pos
	docSplit()
	persist(fmt.Sprintf("S%d,%d", pn, pos))
}

// Undo the immediately preceding operation, if any.
func Undo() (op byte) {
	if current > 0 {
		Flush()
		op = docUndo()
	}

	return op
}

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

func validateSpan(pn, pos, end int) {
	validatePos(pn, pos)
	if end <= pos || end > len(document[pn-1])+len(pending)+1 {
		panic(fmt.Errorf("end '%d,%d-%d' %w", pn, pos, end, errRange))
	}
}
