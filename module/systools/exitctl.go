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
// if we want to exit the app, so some cleanup can be executed.
var exitListener map[string]func(int) ExitBehavior = make(map[string]func(int) ExitBehavior)

// adds a callback as listener
func AddExitListener(name string, callbk func(int) ExitBehavior) {
	exitListener[name] = callbk
}

// Exit maps the os.Exit but
// executes all callbacks before
// it the exit was aborted, you will get
// false in return
func Exit(code int) bool {
	for _, listener := range exitListener {
		if behave := listener(code); !behave.proceedWithExit {
			return false
		}
	}
	os.Exit(code)
	return true
}
