package process

import (
	"os"
	"time"
)

type Signal struct {
	Signal   os.Signal
	ThenWait time.Duration
}

var (
	DefaultKillSignal      = Signal{Signal: os.Kill, ThenWait: 20 * time.Millisecond}
	DefaultInterruptSignal = Signal{Signal: os.Interrupt, ThenWait: 800 * time.Millisecond}
)
