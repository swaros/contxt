package systools

import "os"

type ExitBehavior struct {
	proceedWithExit bool
}

var (
	Continue  ExitBehavior = ExitBehavior{proceedWithExit: true}
	Interrupt ExitBehavior = ExitBehavior{proceedWithExit: false}
)

// contains all listener they should be executed
// if we want to exit the app, so some clanup can be executed.
var exitListener map[string]func(int) ExitBehavior = make(map[string]func(int) ExitBehavior)

// adds a callback as listener
func AddExitListener(name string, callbk func(int) ExitBehavior) {
	exitListener[name] = callbk
}

// Exit maps the os.Exit but
// executes all callbacks before
func Exit(code int) {
	for _, listener := range exitListener {
		listener(code)
	}
	os.Exit(code)
}
