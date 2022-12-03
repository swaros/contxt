package systools

// Just maps the term functions and use StdOut as default
import (
	"os"

	"golang.org/x/term"
)

func IsStdOutTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func GetStdOutTermSize() (width, height int, err error) {
	return term.GetSize(int(os.Stdout.Fd()))
}
