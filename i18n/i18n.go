package i18n

import (
	"embed"
	"slices"
	"strings"

	"github.com/jeandeaual/go-locale"
	"github.com/rivo/uniseg"
)

//go:embed help.* text.*
var translations embed.FS

var (
	HelpText  []string
	HelpWidth int
	Text      = make(map[string]string)
	TextWidth = make(map[string]int)
)

func init() {
	var err error
	var userLanguage string
	if userLanguage, err = locale.GetLanguage(); err != nil {
		userLanguage = "en"
	}

	var b []byte
	if b, err = translations.ReadFile("help." + userLanguage); err != nil {
		b, _ = translations.ReadFile("help.en")
	}

	HelpText = strings.Split(string(b), "\n")
	if len(HelpText[len(HelpText)-1]) == 0 { // Trim final blank line
		HelpText = slices.Delete(HelpText, len(HelpText)-1, len(HelpText))
	}

	for _, l := range HelpText {
		HelpWidth = max(HelpWidth, uniseg.StringWidth(l))
	}

	if b, err = translations.ReadFile("text." + userLanguage); err != nil {
		b, _ = translations.ReadFile("text.en")
	}

	s := strings.Split(string(b), "\n")
	for _, t := range s {
		k, v, _ := strings.Cut(t, "|")
		v = strings.ReplaceAll(v, `\n`, "\n")
		Text[k] = v
		TextWidth[k] = uniseg.StringWidth(v)
	}
}
