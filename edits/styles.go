package edits

import (
	"os"

	"github.com/muesli/termenv"
	"github.com/xanni/jotty/i18n"
)

const (
	confirmColor   = "11" // Confirmation message: ANSIBrightYellow
	cutColor       = "8"  // Cut text: ANSIBrightBlack
	errorColor     = "9"  // Error message: ANSIBrightRed
	helpColor      = "14" // Help text: ANSIBrightCyan
	markColor      = "11" // Edit mark: ANSIBrightYellow
	primaryColor   = "9"  // Primary selection: ANSIBrightRed
	secondaryColor = "13" // Secondary selection: ANSIBrightMagenta
)

var (
	cursorCapString string           // Cursor indicating initial capital letter
	cursorString    [MaxScope]string // Cursor string for each scope
	output          = termenv.NewOutput(os.Stdout)
)

func init() {
	cursorCapString = cursorStyle(string(cursorCharCap))
	for i := range MaxScope {
		cursorString[i] = cursorStyle(string(cursorChar[i]))
	}
}

func errorString() string {
	return output.String(i18n.Text["error"]).Blink().Foreground(output.Color(errorColor)).String()
}

func markString() string {
	return output.String(string(markChar)).Blink().Foreground(output.Color(markColor)).String()
}

func confirmStyle(s string) string {
	return output.String(s).Blink().Foreground(output.Color(confirmColor)).String()
}

func cursorStyle(s string) string {
	return output.String(s).Reverse().Blink().String()
}

func cutStyle(s string) string {
	return output.String(s).CrossOut().Foreground(output.Color(cutColor)).String()
}

// Currently selected cut.
func cutCurStyle(s string) string {
	return output.String(s).Reverse().String()
}

// Timestamp of unselected cut.
func cutTimeStyle(s string) string {
	return output.String(s).Reverse().Foreground(output.Color(cutColor)).String()
}

// Cut window.
func cutWinStyle(s string) string {
	return output.String(s).Foreground(output.Color(cutColor)).String()
}

func errorStyle(s string) string {
	return output.String(s).Foreground(output.Color(errorColor)).String()
}

func helpStyle(s string) string {
	return output.String(s).Foreground(output.Color(helpColor)).String()
}

func primaryStyle(s string) string {
	return output.String(s).Reverse().Foreground(output.Color(primaryColor)).String()
}

func secondaryStyle(s string) string {
	return output.String(s).Underline().Foreground(output.Color(secondaryColor)).String()
}
