package permascroll

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	assert := assert.New(t)
	Init()
	assert.Equal([]string{""}, document)
	assert.Equal(magic, string(permascroll))
}

func TestAppendText(t *testing.T) {
	assert := assert.New(t)
	Init()
	AppendText(1, "Test")
	assert.Equal("Test", pending)
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
			Init()
			InsertText(1, 0, "Sample ")
			SplitParagraph(1, 7)
			InsertText(2, 0, "2")
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

	Init()
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

func TestFlushDeleting(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		offset, deleting  int
		para, permascroll string
	}{
		"Beginning": {0, 1, "est", "D1,0:T"},
		"Middle":    {1, 1, "Tst", "D1,1:e"},
		"End":       {3, 1, "Tes", "D1,3:t"},
		"All":       {0, 4, "", "D1,0:Test"},
	}

	for name, test := range tests {
		t.Run(name, func(_ *testing.T) {
			Init()
			document = []string{"Test"}
			offset, deleting = test.offset, test.deleting
			Flush()
			assert.Equal(test.para, document[0])
			assert.Equal(magic+test.permascroll+"\n", string(permascroll))
		})
	}
}

func TestFlushInserting(t *testing.T) {
	assert := assert.New(t)
	Init()

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
			Init()
			document = []string{"Test"}
			offset, pending = test.offset, "New"
			Flush()
			assert.Equal(test.para, document[0])
			assert.Equal(magic+test.permascroll+"\n", string(permascroll))
		})
	}
}

func TestGetSize(t *testing.T) {
	assert := assert.New(t)
	Init()
	AppendText(1, "Test")
	SplitParagraph(1, 4)
	AppendText(2, "Two")
	assert.Equal(4, GetSize(1))
	assert.Equal(3, GetSize(2))
}

func TestGetText(t *testing.T) {
	assert := assert.New(t)
	Init()
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
	Init()
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
	InsertText(1, 14, "Nine")
	assert.Equal("ThreeSixOneTwoNine", GetText(1))
}

func TestMergeParagraph(t *testing.T) {
	assert := assert.New(t)

	Init()
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
			Init()
			document = test.document
			MergeParagraph(1)
			assert.Equal(test.para, document[0])
			assert.Equal(test.offset, offset)
		})
	}
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
			Init()
			document[0] = test.para
			SplitParagraph(1, test.pos)
			assert.Equal(test.document, document)
		})
	}
}

func TestValidatePn(t *testing.T) {
	assert := assert.New(t)
	Init()
	assert.PanicsWithError("paragraph '0' out of range", func() { validatePn(0) })
	assert.NotPanics(func() { validatePn(1) })
	assert.PanicsWithError("paragraph '2' out of range", func() { validatePn(2) })
}

func TestValidatePos(t *testing.T) {
	assert := assert.New(t)
	Init()
	assert.PanicsWithError("pos '1,-1' out of range", func() { validatePos(1, -1) })
	assert.NotPanics(func() { validatePos(1, 0) })
	assert.PanicsWithError("pos '1,1' out of range", func() { validatePos(1, 1) })

	InsertText(1, 0, "Test")
	assert.NotPanics(func() { validatePos(1, 4) })
	assert.PanicsWithError("pos '1,5' out of range", func() { validatePos(1, 5) })
}

func TestValidateRange(t *testing.T) {
	assert := assert.New(t)
	Init()
	assert.PanicsWithError("end '1,0-0' out of range", func() { validateRange(1, 0, 0) })
	assert.NotPanics(func() { validateRange(1, 0, 1) })

	InsertText(1, 0, "Test")
	assert.NotPanics(func() { validateRange(1, 4, 5) })
	assert.PanicsWithError("end '1,4-6' out of range", func() { validateRange(1, 4, 6) })
}
