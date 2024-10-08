package permascroll

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	if err := OpenPermascroll(os.DevNull); err != nil {
		panic(err)
	}
}

func TestInit(t *testing.T) {
	assert := assert.New(t)
	Init("")
	assert.Equal([]string{""}, document)
	assert.Equal(magic, string(permascroll))
}

func TestAppendText(t *testing.T) {
	assert := assert.New(t)
	Init("")
	AppendText(1, "Test")
	assert.Equal("Test", pending)
}

func TestCopyText(t *testing.T) {
	assert := assert.New(t)
	Init("I1,0:Test\n")

	assert.Equal(1, CopyText(1, 0, 4))
	assert.Equal(1, CopyText(1, 0, 4)) // Copy repeated
}

func TestCutText(t *testing.T) {
	assert := assert.New(t)
	Init("I1,0:Tested\n")

	assert.Equal(1, CutText(1, 4, 5)) // Cut 'e'
	assert.Equal(1, CutText(1, 1, 2)) // Cut 'e' repeated
}

func TestCutTime(t *testing.T) {
	assert := assert.New(t)
	Init("")

	docCopy("1", epoch.Add(3*time.Millisecond))
	assert.Equal("+3", cutTime())

	docCopy("2", epoch.Add(2*time.Minute+time.Second))
	assert.Equal("@2", cutTime())

	cut[0].ts = time.Time{}
	assert.Equal("@2", cutTime())
}

func TestDeleteText(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		isPending              bool
		pn, pos, end, deleting int
	}{
		"Paragraph":          {false, 2, 0, 1, 1},
		"Before offset":      {false, 1, 0, 1, 1},
		"At end":             {false, 1, 10, 11, 1},
		"Over pending start": {true, 1, 6, 8, 2},
		"Over pending end":   {true, 1, 11, 13, 1},
		"After pending":      {true, 1, 15, 16, 1},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			Init("I1,0:Sample 2\nS1,7\n")
			InsertText(1, 7, "data")
			Flush()
			if test.isPending {
				InsertText(1, 7, "test ")
			}

			DeleteText(test.pn, test.pos, test.end)
			assert.Equal(test.pos, offset)
			assert.Equal(test.deleting, deleting)
		})
	}

	Init("")
	InsertText(1, 0, "Test")
	DeleteText(1, 1, 2)
	assert.Equal("Tst", pending)
	assert.Equal(0, offset)
	assert.Equal(0, deleting)

	Flush()
	DeleteText(1, 1, 2)
	DeleteText(1, 1, 2)
	assert.Equal(1, offset)
	assert.Equal(2, deleting)
}

func TestDocCopy(t *testing.T) {
	assert := assert.New(t)
	Init("")

	assert.Equal(0, docCopy("1", time.Time{}))
	assert.Equal(0, docCopy("2", time.Time{}))
	assert.Equal(1, docCopy("1", time.Time{}))
}

func TestDocDelete(t *testing.T) {
	assert := assert.New(t)
	paragraph = 1
	tests := map[string]struct {
		offset, size int
		expect       string
	}{"Beginning": {0, 1, "est"}, "Middle": {1, 1, "Tst"}, "End": {3, 1, "Tes"}, "All": {0, 4, ""}}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			document = []string{"Test"}
			offset = test.offset
			docDelete(test.size)
			assert.Equal(test.expect, document[0])
		})
	}
}

func TestDocExchange(t *testing.T) {
	assert := assert.New(t)
	paragraph = 2
	tests := map[string]struct {
		begin1, end1, begin2, end2 int
		expect                     []string
	}{
		"Paragraphs": {0, 0, 0, 0, []string{"strings", "Test"}},
		"Middle":     {2, 4, 4, 6, []string{"Test", "stngris"}},
		"All":        {0, 3, 3, 7, []string{"Test", "ingsstr"}},
		"Disjoint":   {1, 3, 4, 6, []string{"Test", "sngitrs"}},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {})
		document = []string{"Test", "strings"}
		docHash = []uint64{0, 0}
		docExchange(span{test.begin1, test.end1}, span{test.begin2, test.end2})
		assert.Equal(test.expect, document)
	}
}

func TestDocReplace(t *testing.T) {
	assert := assert.New(t)
	document = []string{"Test"}
	docHash = []uint64{0}
	paragraph, offset = 1, 1
	docReplace(1, "12")
	assert.Equal([]string{"T12st"}, document)
	assert.Equal(3, offset)
}

func TestExchangeParagraphs(t *testing.T) {
	assert := assert.New(t)
	assert.PanicsWithError("paragraph '1' out of range", func() { ExchangeParagraphs(1) })

	Init("I1,0:OneTwo\nS1,3\n")

	ExchangeParagraphs(2)
	assert.Equal(3, current)
	assert.Equal([]string{"Two", "One"}, document)
	expect := magic + "I1,0:OneTwo\nS1,3\nX2\n"
	assert.Equal(expect, string(permascroll))

	ExchangeParagraphs(2)
	assert.Equal(2, current)
	assert.Equal([]string{"One", "Two"}, document)
	assert.Equal(expect, string(permascroll))

	ExchangeParagraphs(2)
	assert.Equal(3, current)
	assert.Equal([]string{"Two", "One"}, document)
	assert.Equal(expect, string(permascroll))
}

func TestExchangeText(t *testing.T) {
	assert := assert.New(t)
	Init("")

	docInsert("Test")
	assert.PanicsWithError("overlap '1-3/2-4' out of range", func() { ExchangeText(1, 1, 3, 2, 4) })

	ExchangeText(1, 1, 4, 0, 1)
	assert.Equal("estT", document[0])
	expect := magic + "X1,0+1/1+3\n"
	assert.Equal(expect, string(permascroll))

	ExchangeText(1, 1, 2, 3, 4)
	assert.Equal("eTts", document[0])
	expect += "X1,1+1/3+1\n"
	assert.Equal(expect, string(permascroll))

	ExchangeText(1, 1, 2, 3, 4)
	assert.Equal("estT", document[0])
	assert.Equal(expect, string(permascroll))

	ExchangeText(1, 1, 2, 3, 4)
	assert.Equal("eTts", document[0])
	assert.Equal(expect, string(permascroll))
}

func TestFlushDeleting(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		offset, deleting  int
		para, permascroll string
	}{
		"Beginning": {0, 1, "est", "D1,0:T\n"},
		"Middle":    {1, 1, "Tst", "D1,1:e\n"},
		"End":       {3, 1, "Tes", "D1,3:t\n"},
		"All":       {0, 4, "", ""},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			Init("")
			docInsert("Test")
			offset, deleting = test.offset, test.deleting
			Flush()
			assert.Equal(test.para, document[0])
			assert.Equal(magic+test.permascroll, string(permascroll))
		})
	}
}

func TestFlushInserting(t *testing.T) {
	assert := assert.New(t)
	Init("")

	Flush()
	assert.Equal([]string{""}, document)

	pending = "Test"
	Flush()
	assert.Equal([]string{"Test"}, document)
	assert.Equal(magic+"I1,0:Test\n", string(permascroll))

	tests := map[string]struct {
		offset            int
		para, permascroll string
	}{
		"Beginning": {0, "NewTest", "I1,0:New"},
		"Middle":    {2, "TeNewst", "I1,2:New"},
		"End":       {4, "TestNew", "I1,4:New"},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			Init("")
			docInsert("Test")
			offset, pending = test.offset, "New"
			Flush()
			assert.Equal(test.para, document[0])
			assert.Equal(magic+test.permascroll+"\n", string(permascroll))
		})
	}
}

func TestGetSize(t *testing.T) {
	assert := assert.New(t)
	Init("I1,0:TestTwo\nS1,4\n")
	assert.Equal(4, GetSize(1))
	assert.Equal(3, GetSize(2))
}

func TestGetText(t *testing.T) {
	assert := assert.New(t)
	Init("")
	pending = "Two "
	Flush()
	pending, offset = "words", 4
	assert.Equal("Two words", GetText(1))

	Flush()
	assert.Equal("Two words", GetText(1))

	deleting, offset = 1, 4
	assert.Equal("Two ords", GetText(1))
}

func TestInsertText(t *testing.T) {
	assert := assert.New(t)
	Init("")
	InsertText(1, 0, "One")
	InsertText(1, 3, "Two")
	InsertText(1, 0, "Three")
	InsertText(1, 5, "Four")
	assert.Equal("ThreeFourOneTwo", pending)

	SplitParagraph(1, 15)
	AppendText(2, "Five")
	InsertText(1, 9, "Six")
	InsertText(2, 4, "Seven")
	Flush()
	InsertText(2, 4, "Eight")
	assert.Equal([]string{"ThreeFourSixOneTwo", "FiveSeven"}, document)
	assert.Equal(magic+"I1,0:ThreeFourOneTwo\nS1,15\nI2,0:Five\nI1,9:Six\nI2,4:Seven\n", string(permascroll))
	assert.Equal("ThreeFourSixOneTwo", GetText(1))
	assert.Equal("FiveEightSeven", GetText(2))

	DeleteText(1, 5, 9)
	InsertText(1, 5, "Nine")
	InsertText(1, 18, "Ten")
	assert.Equal("ThreeNineSixOneTwoTen", GetText(1))
}

func TestReplaceText(t *testing.T) {
	assert := assert.New(t)
	Init("I1,0:Test\n")
	ReplaceText(1, 2, 3, "12")
	assert.Equal([]string{"Te12t"}, document)
	assert.Equal(magic+"I1,0:Test\nR1,2:s\t12\n", string(permascroll))
}

func TestMergeParagraph(t *testing.T) {
	assert := assert.New(t)

	Init("")
	MergeParagraph(1)
	assert.Equal([]string{""}, document)

	tests := map[string]struct {
		document []string
		para     string
		offset   int
	}{
		"Empty":  {[]string{"", ""}, "", 0},
		"First":  {[]string{"One", ""}, "One", 3},
		"Second": {[]string{"", "Two"}, "Two", 0},
		"Both":   {[]string{"One", "Two"}, "OneTwo", 3},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			Init("")
			document = test.document
			docHash = []uint64{0, 0}
			MergeParagraph(1)
			assert.Equal(test.para, document[0])
			assert.Equal(test.offset, offset)
		})
	}
}

func TestParseCopyCut(t *testing.T) {
	assert := assert.New(t)

	permascroll = []byte(magic + "Cinvalid\n")
	source := len(magic) + 1
	op, match := parseCopyCut(&source)
	assert.Equal(operation{code: 'C'}, op)
	assert.Nil(match)

	permascroll = []byte(magic + "C1,0+x\n")
	source = len(magic) + 1
	assert.PanicsWithError(`invalid size for 'C', strconv.Atoi: parsing "x": invalid syntax`,
		func() { parseCopyCut(&source) })

	tests := map[string]struct {
		arguments string
		op        operation
		pn        string
	}{
		"Copy": {"1,2+3", operation{'C', 0, 0, 3, 0, 0, "", "", time.Time{}}, "1"},
		"Cut":  {"4,5:Test", operation{'C', 0, 0, 0, 0, 0, "Test", "", time.Time{}}, "4"},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			permascroll = []byte(magic + "C" + test.arguments + "\n")
			source := len(magic) + 1
			op, match := parseCopyCut(&source)
			assert.Equal(test.op, op)
			assert.Equal(test.pn, string(match[1]))
		})
	}
}

func TestParseExchange(t *testing.T) {
	assert := assert.New(t)

	permascroll = []byte(magic + "Xinvalid\n")
	source := len(magic) + 1
	op, match := parseExchange(&source)
	assert.Equal(operation{code: 'X'}, op)
	assert.Nil(match)

	tests := map[string]struct {
		arguments string
		op        operation
		pn        string
	}{
		"Paragraph": {"2", operation{code: 'X'}, "2"},
		"Text":      {"1,0+1/2+3", operation{'X', 0, 0, 1, 2, 3, "", "", time.Time{}}, "1"},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			permascroll = []byte(magic + "X" + test.arguments + "\n")
			source := len(magic) + 1
			op, match := parseExchange(&source)
			assert.Equal(test.op, op)
			assert.Equal(test.pn, string(match[1]))
		})
	}
}

func TestParseOperation(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		i            int
		code         byte
		text1, text2 string
	}{
		"Insert":   {1, 'I', "Test", ""},
		"Split":    {2, 'S', "", ""},
		"Exchange": {3, 'X', "", ""},
		"Copy":     {4, 'C', "", ""},
		"Delete":   {5, 'D', "e", ""},
		"Merge":    {6, 'M', "", ""},
		"Replace":  {7, 'R', "t", "en"},
	}

	Init("I1,0:Test\nS1,2\nX2\nC2,1+1\nD2,1:e\nM1,0\nR1,1:t\ten\n")

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			source := history[test.i].source
			delta, op := parseOperation(&source)
			assert.Equal(0, delta)
			assert.Equal(test.code, op.code)
			assert.Equal(test.text1, op.text1)
			assert.Equal(test.text2, op.text2)
		})
	}
}

func TestParsePermascroll(t *testing.T) {
	assert := assert.New(t)

	permascroll = []byte{}
	assert.PanicsWithError(`invalid magic, parse failed`, func() { parsePermascroll() })

	permascroll = []byte("bad magic\n")
	assert.PanicsWithError(`invalid magic, parse failed`, func() { parsePermascroll() })

	permascroll = []byte(magic + "bad\n")
	assert.PanicsWithError(`invalid operation 'b', parse failed`, func() { parsePermascroll() })

	permascroll = []byte(magic + "I:bad\n")
	assert.PanicsWithError(`invalid arguments for 'I', parse failed`, func() { parsePermascroll() })

	permascroll = []byte(magic + "R1,0:bad\n")
	assert.PanicsWithError(`invalid arguments for 'R', parse failed`, func() { parsePermascroll() })

	Init("")
	assert.Equal([]version{{}}, history)

	permascroll = []byte(magic + "S1,0\nI1,0:Test\n2I1,0:Two\n@3C1,0+3\n")
	parsePermascroll()
	assert.Equal(4, current)
	assert.Equal([]cutType{{"Two", epoch.Add(3 * time.Minute)}}, cut)
	assert.Equal([]string{"Two"}, document)
	assert.Equal([]version{{0, 0, 3}, {8, 0, 2}, {13, 1, 0}, {23, 0, 4}, {33, 3, 0}}, history)
}

func TestParseTime(t *testing.T) {
	assert := assert.New(t)
	Init("")

	assert.Equal(time.Time{}, parseTime(""))
	assert.Equal(epoch.Add(time.Millisecond), parseTime("+1"))
	assert.Equal(epoch.Add(time.Minute), parseTime("@1"))

	docCopy("Test", epoch.Add(time.Millisecond))
	assert.Equal(epoch.Add(3*time.Millisecond), parseTime("+2"))
}

func TestSplitParagraph(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		pos      int
		para     string
		document []string
	}{
		"Empty":  {0, "", []string{"", ""}},
		"First":  {3, "One", []string{"One", ""}},
		"Second": {0, "Two", []string{"", "Two"}},
		"Both":   {3, "OneTwo", []string{"One", "Two"}},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			Init("")
			document[0] = test.para
			SplitParagraph(1, test.pos)
			assert.Equal(test.document, document)
		})
	}
}

func TestRedo(t *testing.T) {
	assert := assert.New(t)

	Init("")
	Redo()
	assert.Equal(0, current)

	SplitParagraph(1, 0)
	AppendText(1, "Test")
	Undo()
	Redo()
	assert.Equal(2, current)
	assert.Equal([]string{"Test", ""}, document)

	Redo()
	assert.Equal(2, current)

	Undo()
	deleting = 1
	Redo()
	assert.Equal(1, current)

	deleting, pending = 0, "more"
	Redo()
	assert.Equal(1, current)

	MergeParagraph(1)
	CopyText(1, 2, 3)
	DeleteText(1, 2, 3)
	SplitParagraph(1, 2)
	CutText(1, 1, 2)
	ExchangeParagraphs(2)
	ReplaceText(2, 0, 1, "nd")
	for range 7 {
		Undo()
	}
	Redo()
	assert.Equal(4, current)
	assert.Equal([]string{"more"}, document)

	Redo()
	assert.Equal(5, current)
	assert.Equal("r", cut[0].text)

	Redo()
	assert.Equal(6, current)
	assert.Equal([]string{"moe"}, document)

	Redo()
	assert.Equal(7, current)
	assert.Equal([]string{"mo", "e"}, document)

	Redo()
	assert.Equal(8, current)
	assert.Equal("o", cut[1].text)

	Redo()
	assert.Equal(9, current)
	assert.Equal([]string{"e", "m"}, document)

	Redo()
	assert.Equal(10, current)
	assert.Equal([]string{"e", "nd"}, document)
}

func TestUndo(t *testing.T) {
	assert := assert.New(t)

	Init("")
	Undo()
	assert.Equal(0, current)

	SplitParagraph(1, 0)
	AppendText(1, "Test")
	MergeParagraph(1)
	Undo()
	assert.Equal(2, current)
	assert.Equal([]string{"Test", ""}, document)

	docCopy("x", time.Now().Add(-3*time.Millisecond))
	CopyText(1, 1, 2)
	Undo()
	assert.Equal(2, current)

	DeleteText(1, 1, 2)
	Undo()
	assert.Equal(2, current)
	assert.Equal([]string{"Test", ""}, document)
	expectHist := []version{{0, 0, 1}, {8, 0, 2}, {13, 1, 5}, {23, 2, 0}, {28, 2, 0}, {38, 2, 0}}
	assert.Equal(expectHist, history)
	expect := magic + "S1,0\nI1,0:Test\nM1,4\n1+3C1,1+1\n2D1,1:e\n"
	assert.Equal(expect, string(permascroll))

	Undo()
	assert.Equal(1, current)

	SplitParagraph(1, 0)
	Undo()
	assert.Equal(1, current)
	assert.Equal([]string{"", ""}, document)
	expectHist = append(expectHist, version{46, 1, 0})
	expectHist[1].lastChild = 6
	assert.Equal(expectHist, history)
	expect += "4S1,0\n"
	assert.Equal(expect, string(permascroll))
}

func TestValidatePn(t *testing.T) {
	assert := assert.New(t)
	Init("")
	assert.PanicsWithError("paragraph '0' out of range", func() { validatePn(0) })
	assert.NotPanics(func() { validatePn(1) })
	assert.PanicsWithError("paragraph '2' out of range", func() { validatePn(2) })
}

func TestValidatePos(t *testing.T) {
	assert := assert.New(t)
	Init("")
	assert.PanicsWithError("pos '1,-1' out of range", func() { validatePos(1, -1) })
	assert.NotPanics(func() { validatePos(1, 0) })
	assert.PanicsWithError("pos '1,1' out of range", func() { validatePos(1, 1) })

	InsertText(1, 0, "Test")
	assert.NotPanics(func() { validatePos(1, 4) })
	assert.PanicsWithError("pos '1,5' out of range", func() { validatePos(1, 5) })
}

func TestValidateSpan(t *testing.T) {
	assert := assert.New(t)
	Init("")
	assert.PanicsWithError("end '1,0-0' out of range", func() { validateSpan(1, 0, 0) })
	assert.NotPanics(func() { validateSpan(1, 0, 1) })

	InsertText(1, 0, "Test")
	assert.NotPanics(func() { validateSpan(1, 4, 5) })
	assert.PanicsWithError("end '1,4-6' out of range", func() { validateSpan(1, 4, 6) })
}
