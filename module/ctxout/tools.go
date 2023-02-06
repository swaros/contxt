package ctxout

import (
	"regexp"
	"strings"
	//"github.com/muesli/reflow/ansi"
)

// LenPrintable returns the length of a string, but only counts printable characters
// it ignores ANSI escape codes
// and also ignores the length of the ANSI escape codes
// also igrnores any newlines
func LenPrintable(s string) int {
	//return ansi.PrintableRuneWidth(s)  // investigate why it is different
	return len(StringPure(s))
}

func StringPure(s string) string {
	removeChars := []string{
		"\n",
		"\t",
		"\r",
		"\x08",
	}
	for _, c := range removeChars {
		s = strings.ReplaceAll(s, c, "")
	}
	return StringCleanEscapeCodes(s)
}

func StringCleanEscapeCodes(s string) string {
	match := "[^\x08]\x08"
	match += "|\\x1b\\[[0-9;]*[a-zA-Z]"
	match += "|\\x1b\\[[0-9;]*m"

	re := regexp.MustCompile(match)
	return re.ReplaceAllString(s, "")
}
