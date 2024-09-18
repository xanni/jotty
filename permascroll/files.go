package permascroll

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
)

type FileInterface interface {
	Sync() error
	io.WriteCloser
	WriteString(string) (int, error)
}

type opener interface {
	OpenFile(name string, flag int, perms fs.FileMode) (FileInterface, error)
}

type defaultOpener struct{}

func (o defaultOpener) OpenFile(name string, flag int, perms fs.FileMode) (f FileInterface, err error) {
	return os.OpenFile(name, flag, perms) // nolint:wrapcheck
}

var (
	of   opener        = defaultOpener{}
	file FileInterface // Permascroll backing storage
)

// Close the permascroll file.
func ClosePermascroll() (err error) {
	if err = file.Close(); err != nil {
		err = fmt.Errorf("failed to close permascroll: %w", err)
	}

	return err
}

// Export entire document or text from a paragraph between pos and end.
func ExportText(path string, pn, pos, end int) (err error) {
	if pn > 0 {
		validateSpan(pn, pos, end)
	}

	var f FileInterface
	if f, err = of.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644); err != nil {
		return fmt.Errorf("failed export: %w", err)
	}
	defer f.Close() // Ignore error; WriteString error takes precedence

	Flush()
	if pn > 0 {
		_, err = f.WriteString(document[pn-1][pos:end] + "\n")
	} else {
		for i, t := range document {
			if i > 0 {
				_, err = f.WriteString("\n")
			}

			if err == nil {
				_, err = f.WriteString(t + "\n")
			}

			if err != nil {
				break
			}
		}
	}

	if err != nil {
		err = fmt.Errorf("failed export: %w", err)
	}

	return err
}

// Open or create a permascroll file.
func OpenPermascroll(path string) (err error) {
	permascroll, err = os.ReadFile(path)
	if err == nil && len(permascroll) > 0 {
		parsePermascroll()
	}

	file, err = of.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil && len(permascroll) == 0 {
		permascroll = []byte(magic)
		if _, err = file.WriteString(magic); err != nil {
			file.Close() // Ignore error; WriteString error takes precedence
		}
	}

	if err != nil {
		err = fmt.Errorf("failed to open permascroll: %w", err)
	}

	return err
}

// Persist an operation to the permascroll.
func persist(s string) {
	delta := newVersion(len(permascroll))
	if delta < 0 {
		return
	}

	if delta > 0 {
		s = strconv.Itoa(delta) + s
	}

	s += "\n"
	permascroll = append(permascroll, []byte(s)...)
	if _, err := file.WriteString(s); err != nil {
		file.Close() // ignore error; Write error takes precedence
		panic(fmt.Errorf("persist failed: %w", err))
	}
}

// Ensure the permascroll backing store is written to stable storage.
func SyncPermascroll() (err error) {
	Flush()
	if err = file.Sync(); err != nil {
		err = fmt.Errorf("failed to sync permascroll: %w", err)
	}

	return err
}
