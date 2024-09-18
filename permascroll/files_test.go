package permascroll

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errInvalidArg = errors.New("invalid argument")

type mockFileType struct {
	contents string
	err      error
	errDelay int
}

func (f *mockFileType) Close() (err error)                { return f.err }
func (f *mockFileType) Sync() (err error)                 { return f.err }
func (f *mockFileType) Write(p []byte) (n int, err error) { return f.WriteString(string(p)) }
func (f *mockFileType) WriteString(s string) (n int, err error) {
	f.contents += s
	if f.errDelay > 0 {
		f.errDelay--
	} else {
		err = f.err
	}

	return len(s), err
}

type mockOpenerType struct{ err error }

func (o *mockOpenerType) OpenFile(_ string, _ int, _ fs.FileMode) (f FileInterface, err error) {
	return mockFile, o.err
}

var (
	mockFile   = new(mockFileType)
	mockOpener = new(mockOpenerType)
)

func init() { of = mockOpener }

func TestClosePermascroll(t *testing.T) {
	file = &mockFileType{err: errInvalidArg}
	require.ErrorContains(t, ClosePermascroll(), "failed to close permascroll: invalid argument")

	file = &mockFileType{}
	require.NoError(t, ClosePermascroll())
}

func TestExportText(t *testing.T) {
	assert := assert.New(t)

	assert.PanicsWithError("paragraph '2' out of range", func() { _ = ExportText("", 2, 0, 0) })

	document = []string{"One", "Two"}
	mockOpener.err = errInvalidArg
	require.ErrorContains(t, ExportText("", 1, 0, 1), "failed export: ")
	mockOpener.err = nil

	mockFile = &mockFileType{}
	require.NoError(t, ExportText("", 0, 0, 0))
	assert.Equal("One\n\nTwo\n", mockFile.contents)

	mockFile.contents = ""
	require.NoError(t, ExportText("", 1, 1, 3))
	assert.Equal("ne\n", mockFile.contents)

	mockFile.err = errInvalidArg
	require.ErrorContains(t, ExportText("", 0, 0, 0), "failed export: ")
	require.ErrorContains(t, ExportText("", 1, 0, 1), "failed export: ")

	mockFile.errDelay = 1
	require.ErrorContains(t, ExportText("", 0, 0, 0), "failed export: ")

	mockFile.err = nil
}

func TestOpenPermascroll(t *testing.T) {
	mockOpener.err = errInvalidArg
	require.ErrorContains(t, OpenPermascroll(""), "failed to open permascroll: ")

	mockOpener.err = nil
	mockFile = &mockFileType{err: errInvalidArg}
	require.ErrorContains(t, OpenPermascroll(""), "failed to open permascroll: ")

	const testData = magic + "I1,0:Test\n"
	mockFile = &mockFileType{contents: testData}
	testFile, err := os.CreateTemp("", "jotty")
	if err != nil {
		panic(err)
	}
	name := testFile.Name()
	defer os.Remove(name)
	if _, err = testFile.WriteString(testData); err != nil {
		panic(err)
	}
	document = []string{""}
	require.NoError(t, OpenPermascroll(name))
	assert.Equal(t, []string{"Test"}, document)
	assert.Equal(t, testData, mockFile.contents)
	require.NoError(t, ClosePermascroll())
}

func TestPersist(t *testing.T) {
	docInsert("Test")
	file = &mockFileType{err: errInvalidArg}
	assert.PanicsWithError(t, "persist failed: invalid argument", func() { persist("error") })

	file = &mockFileType{}
	require.NotPanics(t, func() { persist("OK") })
}

func TestSyncPermascroll(t *testing.T) {
	file = &mockFileType{err: errInvalidArg}
	require.ErrorContains(t, SyncPermascroll(), "failed to sync permascroll: invalid argument")

	file = &mockFileType{}
	require.NoError(t, SyncPermascroll())
}
