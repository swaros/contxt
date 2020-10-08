package systools

import "fmt"

var (
	// Info Color reference
	Info = Teal
	// Warn Color reference
	Warn = Yellow
	// Fata Color reference
	Fata = Red
)

var (
	// Black Color
	Black = Color("\033[1;30m%s\033[0m")
	// Red Color
	Red = Color("\033[1;31m%s\033[0m")
	// Green Color
	Green = Color("\033[1;32m%s\033[0m")
	// Yellow Color
	Yellow = Color("\033[1;33m%s\033[0m")
	// Purple Color
	Purple = Color("\033[1;34m%s\033[0m")
	// Magenta Color
	Magenta = Color("\033[1;35m%s\033[0m")
	// Teal Color
	Teal = Color("\033[1;36m%s\033[0m")
	// White Color
	White = Color("\033[1;37m%s\033[0m")
)

// Color get the colorized console code
func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
