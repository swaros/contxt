package systools

import "os"

var exitListener map[string]func(int) = make(map[string]func(int))

func AddExitListener(name string, callbk func(int)) {
	exitListener[name] = callbk
}

func Exit(code int) {
	for _, listener := range exitListener {
		listener(code)
	}
	os.Exit(code)
}
