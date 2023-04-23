package runner_test

import (
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/swaros/contxt/module/runner"
)

func GetTestScreen(t *testing.T) *runner.CtCell {
	testApp := runner.NewTcell()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Error(err)
	}
	screen.SetSize(1000, 100)
	testApp.SetScreen(screen)
	return testApp
}

// testing stop sign
func TestLoop(t *testing.T) {
	app := GetTestScreen(t)

	go app.Loop()
	//time.Sleep(10 * time.Millisecond)

	event := tcell.NewEventMouse(0, 0, 0, 0)

	app.SendEvent(event)
	time.Sleep(100 * time.Millisecond)
	app.Stop()
	if app.GetLastLoopTime().Microseconds() < 1 {
		t.Errorf("maybe this hardware is just to fast, but expected looptime should be more then 1 Microsecond, got %v", app.GetLastLoopTime().Microseconds())
	}
}

func TestElementHit(t *testing.T) {
	app := GetTestScreen(t)
	app.SetMouse(true)

	// using a waitgroup to wait for the app to be ready
	wg := sync.WaitGroup{}
	wg.Add(1)

	getHit := false
	// adding a active Text Element
	clickableText := app.ActiveText("I am clickable")
	clickableText.SetPos(10, 10).OnSelect = func(selectedt bool) {
		if selectedt {
			getHit = true
			wg.Done()
			app.Stop()
		}
	}
	if id := app.AddElement(clickableText); id != 1 {
		t.Errorf("Expected id to be 1, but got %v", id)
	}

	// running the app in a goroutine
	go app.Loop()

	// send mouse event that simulates the click on the text
	event := tcell.NewEventMouse(10, 10, tcell.Button1, 1)
	app.SendEvent(event)

	// wait for the app to stop
	wg.Wait()
	if !getHit {
		t.Errorf("Expected to get a hit, but didn't")
	}
}
